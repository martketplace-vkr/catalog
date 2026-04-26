package vendor

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/martketplace-vkr/catalog/internal/service/client/dto"
	domainpb "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/domain"
	vendorpb "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/vendor"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var pricePattern = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

type Handler struct {
	service service
	vendorpb.UnimplementedCatalogVendorServiceServer
}

func New(service service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetVendorProduct(
	ctx context.Context,
	req *vendorpb.GetVendorProductRequest,
) (*vendorpb.GetVendorProductResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.VendorID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "vendor_id must be positive")
	}

	products, err := h.service.GetVendorProducts(ctx, req.VendorID)
	if err != nil {
		return nil, mapError(err)
	}

	return &vendorpb.GetVendorProductResponse{
		Products: products.ToProto(),
	}, nil
}

func (h *Handler) CreateProduct(
	ctx context.Context,
	req *vendorpb.CreateProductRequest,
) (*vendorpb.CreateProductResponse, error) {
	if err := validateCreateProductRequest(req); err != nil {
		return nil, err
	}

	product, err := h.service.CreateProduct(ctx, dto.CreateProductRequestFromProto(req))
	if err != nil {
		return nil, mapError(err)
	}

	return &vendorpb.CreateProductResponse{
		Product: product.ToProto(),
	}, nil
}

func (h *Handler) UpdateProduct(
	ctx context.Context,
	req *vendorpb.UpdateProductRequest,
) (*vendorpb.UpdateProductResponse, error) {
	if err := validateUpdateProductRequest(req); err != nil {
		return nil, err
	}

	product, err := h.service.UpdateProduct(ctx, dto.UpdateProductRequestFromProto(req))
	if err != nil {
		return nil, mapError(err)
	}

	return &vendorpb.UpdateProductResponse{
		Product: product.ToProto(),
	}, nil
}

func (h *Handler) DeleteProduct(
	ctx context.Context,
	req *vendorpb.DeleteProductRequest,
) (*vendorpb.DeleteProductResponse, error) {
	if err := validateDeleteProductRequest(req); err != nil {
		return nil, err
	}

	if err := h.service.DeleteProduct(ctx, dto.DeleteProductRequestFromProto(req)); err != nil {
		return nil, mapError(err)
	}

	return &vendorpb.DeleteProductResponse{}, nil
}

func validateCreateProductRequest(req *vendorpb.CreateProductRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	return validateProductMutation(req.VendorId, req.CategoryId, req.Name, req.Price, req.StockCount, req.Attributes, req.Characteristics, req.Images)
}

func validateUpdateProductRequest(req *vendorpb.UpdateProductRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if req.ProductId <= 0 {
		return status.Error(codes.InvalidArgument, "product_id must be positive")
	}

	return validateProductMutation(req.VendorId, req.CategoryId, req.Name, req.Price, req.StockCount, req.Attributes, req.Characteristics, req.Images)
}

func validateDeleteProductRequest(req *vendorpb.DeleteProductRequest) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	if req.VendorId <= 0 {
		return status.Error(codes.InvalidArgument, "vendor_id must be positive")
	}

	if req.ProductId <= 0 {
		return status.Error(codes.InvalidArgument, "product_id must be positive")
	}

	return nil
}

func validateProductMutation(
	vendorID int64,
	categoryID int64,
	name string,
	price string,
	stockCount uint32,
	attributes []*domainpb.ProductAttributeInput,
	characteristics []*domainpb.ProductCharacteristicInput,
	images []*domainpb.ProductImageInput,
) error {
	if vendorID <= 0 {
		return status.Error(codes.InvalidArgument, "vendor_id must be positive")
	}

	if categoryID <= 0 {
		return status.Error(codes.InvalidArgument, "category_id must be positive")
	}

	if stockCount > math.MaxInt32 {
		return status.Error(codes.InvalidArgument, "stock_count is out of range")
	}

	if strings.TrimSpace(name) == "" {
		return status.Error(codes.InvalidArgument, "name must not be blank")
	}

	trimmedPrice := strings.TrimSpace(price)
	if trimmedPrice == "" {
		return status.Error(codes.InvalidArgument, "price must not be blank")
	}

	if !pricePattern.MatchString(trimmedPrice) {
		return status.Error(codes.InvalidArgument, "price must be a non-negative decimal with up to 2 fractional digits")
	}

	if err := validateProductCharacteristics(characteristics, attributes); err != nil {
		return err
	}

	if err := validateProductImages(images); err != nil {
		return err
	}

	return nil
}

func validateProductCharacteristics(characteristics []*domainpb.ProductCharacteristicInput, legacyAttributes []*domainpb.ProductAttributeInput) error {
	if len(characteristics) == 0 {
		return validateLegacyProductAttributes(legacyAttributes)
	}

	for sectionIdx, characteristic := range characteristics {
		if characteristic == nil {
			return status.Errorf(codes.InvalidArgument, "characteristics[%d] is required", sectionIdx)
		}

		if strings.TrimSpace(characteristic.Title) == "" {
			return status.Errorf(codes.InvalidArgument, "characteristics[%d].title must not be blank", sectionIdx)
		}

		if len(characteristic.Attributes) == 0 {
			return status.Errorf(codes.InvalidArgument, "characteristics[%d].attributes must not be empty", sectionIdx)
		}

		seenNames := make(map[string]struct{}, len(characteristic.Attributes))
		for attrIdx, attribute := range characteristic.Attributes {
			if attribute == nil {
				return status.Errorf(codes.InvalidArgument, "characteristics[%d].attributes[%d] is required", sectionIdx, attrIdx)
			}

			name := strings.TrimSpace(attribute.Name)
			if name == "" {
				return status.Errorf(codes.InvalidArgument, "characteristics[%d].attributes[%d].name must not be blank", sectionIdx, attrIdx)
			}

			if strings.TrimSpace(attribute.Value) == "" {
				return status.Errorf(codes.InvalidArgument, "characteristics[%d].attributes[%d].value must not be blank", sectionIdx, attrIdx)
			}

			if _, exists := seenNames[name]; exists {
				return status.Errorf(codes.InvalidArgument, "characteristics[%d].attributes[%d].name must be unique within section", sectionIdx, attrIdx)
			}

			seenNames[name] = struct{}{}
		}
	}

	return nil
}

func validateLegacyProductAttributes(attributes []*domainpb.ProductAttributeInput) error {
	if len(attributes) == 0 {
		return nil
	}

	for idx, attribute := range attributes {
		if attribute == nil {
			return status.Errorf(codes.InvalidArgument, "attributes[%d] is required", idx)
		}

		name := strings.TrimSpace(attribute.Name)
		if name == "" {
			return status.Errorf(codes.InvalidArgument, "attributes[%d].name must not be blank", idx)
		}

		if strings.TrimSpace(attribute.Value) == "" {
			return status.Errorf(codes.InvalidArgument, "attributes[%d].value must not be blank", idx)
		}
	}

	return nil
}

func validateProductImages(images []*domainpb.ProductImageInput) error {
	if len(images) == 0 {
		return nil
	}

	mainImagesCount := 0

	for idx, image := range images {
		if image == nil {
			return status.Errorf(codes.InvalidArgument, "images[%d] is required", idx)
		}

		if strings.TrimSpace(image.Url) == "" {
			return status.Errorf(codes.InvalidArgument, "images[%d].url must not be blank", idx)
		}

		if image.IsMain {
			mainImagesCount++
			if mainImagesCount > 1 {
				return status.Error(codes.InvalidArgument, "only one image can be marked as main")
			}
		}
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
