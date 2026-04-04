package client

import (
	"context"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/client/dto"
)

type (
	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	}
	repository interface {
		SelectCategories(ctx context.Context) (categories domain.CategoryList, err error)
		SelectVendorProducts(ctx context.Context, vendorID int64) (products domain.ProductList, err error)
		SelectProducts(
			ctx context.Context,
			req dto.ListProductsRequest,
			limit uint32,
		) (products domain.ProductList, err error)
		SelectProduct(ctx context.Context, productID int64) (product domain.Product, err error)
		CreateProduct(ctx context.Context, req dto.CreateProductRequest) (product domain.Product, err error)
		UpdateProduct(ctx context.Context, req dto.UpdateProductRequest) (product domain.Product, err error)
		DeleteProduct(ctx context.Context, req dto.DeleteProductRequest) error
	}
)
