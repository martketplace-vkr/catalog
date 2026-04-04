package vendor

import (
	"context"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/client/dto"
)

type (
	service interface {
		GetVendorProducts(ctx context.Context, vendorID int64) (products domain.ProductList, err error)
		CreateProduct(ctx context.Context, req dto.CreateProductRequest) (product domain.Product, err error)
		UpdateProduct(ctx context.Context, req dto.UpdateProductRequest) (product domain.Product, err error)
		DeleteProduct(ctx context.Context, req dto.DeleteProductRequest) error
	}
)
