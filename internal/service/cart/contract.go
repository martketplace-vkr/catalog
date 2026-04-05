package cart

import (
	"context"

	"github.com/martketplace-vkr/catalog/domain"
)

type (
	repository interface {
		SelectProducts(ctx context.Context, productIDs []int64) (products domain.ProductList, err error)
	}
)
