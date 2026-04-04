package client

import (
	"context"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/client/dto"
)

type (
	service interface {
		GetCategories(ctx context.Context, req dto.GetCategoriesRequest) (categories domain.CategoryList, err error)
		ListProducts(
			ctx context.Context,
			req dto.ListProductsRequest,
		) (products domain.ProductList, nextPageToken uint64, err error)
		GetProduct(ctx context.Context, productID int64) (product domain.Product, err error)
	}
)
