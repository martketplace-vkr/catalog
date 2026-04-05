package cart

import (
	"context"

	"github.com/martketplace-vkr/catalog/domain"
)

type Service struct {
	repository repository
}

func New(repository repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetProductList(ctx context.Context, productIDs []int64) (products domain.ProductList, err error) {
	if len(productIDs) == 0 {
		return domain.ProductList{}, nil
	}

	return s.repository.SelectProducts(ctx, productIDs)
}
