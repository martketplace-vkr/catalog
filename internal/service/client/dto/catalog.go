package dto

import (
	"strings"

	clientpb "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/client"
	domainpb "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/domain"
	vendorpb "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/vendor"
)

type GetCategoriesRequest struct {
	ParentID        int64
	FilterByParent  bool
	IncludeChildren bool
}

func GetCategoriesRequestFromProto(req *clientpb.GetCategoriesRequest) GetCategoriesRequest {
	if req == nil {
		return GetCategoriesRequest{}
	}

	return GetCategoriesRequest{
		ParentID:        req.ParentId,
		FilterByParent:  req.FilterByParent,
		IncludeChildren: req.IncludeChildren,
	}
}

type ListProductsRequest struct {
	CategoryID int64
	VendorID   int64
	PageSize   uint32
	PageToken  uint64
}

func ListProductsRequestFromProto(req *clientpb.ListProductsRequest) ListProductsRequest {
	if req == nil {
		return ListProductsRequest{}
	}

	return ListProductsRequest{
		CategoryID: req.CategoryId,
		VendorID:   req.VendorId,
		PageSize:   req.PageSize,
		PageToken:  req.PageToken,
	}
}

type ProductAttributeInput struct {
	Name  string
	Value string
}

type ProductImageInput struct {
	URL    string
	IsMain bool
}

type CreateProductRequest struct {
	VendorID    int64
	CategoryID  int64
	Name        string
	Description string
	Price       string
	StockCount  uint32
	Attributes  []ProductAttributeInput
	Images      []ProductImageInput
}

func CreateProductRequestFromProto(req *vendorpb.CreateProductRequest) CreateProductRequest {
	if req == nil {
		return CreateProductRequest{}
	}

	return CreateProductRequest{
		VendorID:    req.VendorId,
		CategoryID:  req.CategoryId,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Price:       strings.TrimSpace(req.Price),
		StockCount:  req.StockCount,
		Attributes:  productAttributeInputsFromProto(req.Attributes),
		Images:      productImageInputsFromProto(req.Images),
	}
}

type UpdateProductRequest struct {
	ProductID   int64
	VendorID    int64
	CategoryID  int64
	Name        string
	Description string
	Price       string
	StockCount  uint32
	Attributes  []ProductAttributeInput
	Images      []ProductImageInput
}

type DeleteProductRequest struct {
	VendorID  int64
	ProductID int64
}

func UpdateProductRequestFromProto(req *vendorpb.UpdateProductRequest) UpdateProductRequest {
	if req == nil {
		return UpdateProductRequest{}
	}

	return UpdateProductRequest{
		ProductID:   req.ProductId,
		VendorID:    req.VendorId,
		CategoryID:  req.CategoryId,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Price:       strings.TrimSpace(req.Price),
		StockCount:  req.StockCount,
		Attributes:  productAttributeInputsFromProto(req.Attributes),
		Images:      productImageInputsFromProto(req.Images),
	}
}

func DeleteProductRequestFromProto(req *vendorpb.DeleteProductRequest) DeleteProductRequest {
	if req == nil {
		return DeleteProductRequest{}
	}

	return DeleteProductRequest{
		VendorID:  req.VendorId,
		ProductID: req.ProductId,
	}
}

func productAttributeInputsFromProto(attributes []*domainpb.ProductAttributeInput) []ProductAttributeInput {
	if len(attributes) == 0 {
		return nil
	}

	result := make([]ProductAttributeInput, 0, len(attributes))

	for _, attribute := range attributes {
		if attribute == nil {
			result = append(result, ProductAttributeInput{})
			continue
		}

		result = append(result, ProductAttributeInput{
			Name:  strings.TrimSpace(attribute.Name),
			Value: strings.TrimSpace(attribute.Value),
		})
	}

	return result
}

func productImageInputsFromProto(images []*domainpb.ProductImageInput) []ProductImageInput {
	if len(images) == 0 {
		return nil
	}

	result := make([]ProductImageInput, 0, len(images))

	for _, image := range images {
		if image == nil {
			result = append(result, ProductImageInput{})
			continue
		}

		result = append(result, ProductImageInput{
			URL:    strings.TrimSpace(image.Url),
			IsMain: image.IsMain,
		})
	}

	return result
}
