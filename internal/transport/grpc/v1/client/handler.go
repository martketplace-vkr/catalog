package client

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/martketplace-vkr/catalog/internal/service/client/dto"
	"github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	service service
	client.UnimplementedCatalogClientServiceServer
}

func New(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetCategories(
	ctx context.Context,
	req *client.GetCategoriesRequest,
) (resp *client.GetCategoriesResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.FilterByParent && req.ParentId < 0 {
		return nil, status.Error(codes.InvalidArgument, "parent_id must be non-negative")
	}

	categories, err := h.service.GetCategories(ctx, dto.GetCategoriesRequestFromProto(req))
	if err != nil {
		return nil, mapError(err)
	}

	return &client.GetCategoriesResponse{
		Categories: categories.ToProto(),
	}, nil
}

func (h *Handler) ListProducts(
	ctx context.Context,
	req *client.ListProductsRequest,
) (resp *client.ListProductsResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.CategoryId < 0 {
		return nil, status.Error(codes.InvalidArgument, "category_id must be non-negative")
	}

	if req.VendorId < 0 {
		return nil, status.Error(codes.InvalidArgument, "vendor_id must be non-negative")
	}

	if req.PageToken > math.MaxInt64 {
		return nil, status.Error(codes.InvalidArgument, "page_token is out of range")
	}

	products, nextPageToken, err := h.service.ListProducts(ctx, dto.ListProductsRequestFromProto(req))
	if err != nil {
		return nil, mapError(err)
	}

	return &client.ListProductsResponse{
		Products:      products.ToProto(),
		NextPageToken: nextPageToken,
	}, nil
}

func (h *Handler) GetProduct(
	ctx context.Context,
	req *client.GetProductRequest,
) (resp *client.GetProductResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ProductId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id must be positive")
	}

	product, err := h.service.GetProduct(ctx, req.ProductId)
	if err != nil {
		return nil, mapError(err)
	}

	return &client.GetProductResponse{
		Product: product.ToProto(),
	}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
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
		case "product_attributes_product_name_unique":
			return status.Error(codes.InvalidArgument, "attribute names must be unique within product")
		case "product_images_one_main_per_product_uidx":
			return status.Error(codes.InvalidArgument, "only one image can be marked as main")
		default:
			return status.Error(codes.AlreadyExists, "resource already exists")
		}
	case "23514":
		return status.Error(codes.InvalidArgument, "request violates data constraints")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
