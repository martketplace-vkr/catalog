package cart

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	cartpb "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/cart"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	service service
	cartpb.UnimplementedCatalogCartServiceServer
}

func New(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetProductList(
	ctx context.Context,
	req *cartpb.GetProductListRequest,
) (*cartpb.GetProductListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	for idx, productID := range req.ProductIds {
		if productID <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "product_ids[%d] must be positive", idx)
		}
	}

	products, err := h.service.GetProductList(ctx, req.ProductIds)
	if err != nil {
		return nil, mapError(err)
	}

	return &cartpb.GetProductListResponse{
		Products: products.ToProto(),
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
		return status.Error(codes.AlreadyExists, "resource already exists")
	case "23514":
		return status.Error(codes.InvalidArgument, "request violates data constraints")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
