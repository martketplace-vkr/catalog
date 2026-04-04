package admin

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	serviceadmin "github.com/martketplace-vkr/catalog/internal/service/admin"
	"github.com/martketplace-vkr/catalog/internal/service/admin/dto"
	adminpb "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	service service
	adminpb.UnimplementedCatalogAdminServiceServer
}

func New(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateCategory(
	ctx context.Context,
	req *adminpb.CreateCategoryRequest,
) (*adminpb.CreateCategoryResponse, error) {
	if err := validateCreateCategoryRequest(req); err != nil {
		return nil, err
	}

	category, err := h.service.CreateCategory(ctx, dto.CreateCategoryRequestFromProto(req))
	if err != nil {
		return nil, mapError(err)
	}

	return &adminpb.CreateCategoryResponse{
		Category: category.ToProto(),
	}, nil
}

func (h *Handler) UpdateCategory(
	ctx context.Context,
	req *adminpb.UpdateCategoryRequest,
) (*adminpb.UpdateCategoryResponse, error) {
	if err := validateUpdateCategoryRequest(req); err != nil {
		return nil, err
	}

	category, err := h.service.UpdateCategory(ctx, dto.UpdateCategoryRequestFromProto(req))
	if err != nil {
		return nil, mapError(err)
	}

	return &adminpb.UpdateCategoryResponse{
		Category: category.ToProto(),
	}, nil
}

func (h *Handler) DeleteCategory(
	ctx context.Context,
	req *adminpb.DeleteCategoryRequest,
) (*adminpb.DeleteCategoryResponse, error) {
	if err := validateDeleteCategoryRequest(req); err != nil {
		return nil, err
	}

	deletedCategoryID, err := h.service.DeleteCategory(ctx, req.CategoryId)
	if err != nil {
		return nil, mapError(err)
	}

	return &adminpb.DeleteCategoryResponse{
		CategoryId: deletedCategoryID,
	}, nil
}

func validateCreateCategoryRequest(req *adminpb.CreateCategoryRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return status.Error(codes.InvalidArgument, "name must not be blank")
	}

	if req.ParentId != nil && *req.ParentId <= 0 {
		return status.Error(codes.InvalidArgument, "parent_id must be positive")
	}

	return nil
}

func validateUpdateCategoryRequest(req *adminpb.UpdateCategoryRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if req.CategoryId <= 0 {
		return status.Error(codes.InvalidArgument, "category_id must be positive")
	}

	if strings.TrimSpace(req.Name) == "" {
		return status.Error(codes.InvalidArgument, "name must not be blank")
	}

	if req.ParentId != nil {
		if *req.ParentId <= 0 {
			return status.Error(codes.InvalidArgument, "parent_id must be positive")
		}

		if *req.ParentId == req.CategoryId {
			return status.Error(codes.InvalidArgument, "parent_id must not equal category_id")
		}
	}

	return nil
}

func validateDeleteCategoryRequest(req *adminpb.DeleteCategoryRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if req.CategoryId <= 0 {
		return status.Error(codes.InvalidArgument, "category_id must be positive")
	}

	return nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	if errors.Is(err, serviceadmin.ErrCategoryCycle) {
		return status.Error(codes.InvalidArgument, "parent_id cannot reference a descendant category")
	}

	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request was canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case errors.Is(err, sql.ErrNoRows):
		return status.Error(codes.NotFound, "resource not found")
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return mapPGError(pgErr)
		}

		return status.Error(codes.Internal, "internal server error")
	}
}

func mapPGError(err *pgconn.PgError) error {
	switch err.Code {
	case "22003":
		return status.Error(codes.InvalidArgument, "numeric value is out of range")
	case "22P02":
		return status.Error(codes.InvalidArgument, "request contains an invalid value")
	case "23503":
		return status.Error(codes.InvalidArgument, "referenced resource not found")
	case "23505":
		switch err.ConstraintName {
		case "categories_parent_name_uidx":
			return status.Error(codes.AlreadyExists, "category with same parent and name already exists")
		default:
			return status.Error(codes.AlreadyExists, "resource already exists")
		}
	case "23514":
		return status.Error(codes.InvalidArgument, "request violates data constraints")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
