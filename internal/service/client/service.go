package client

import (
	"context"
	"errors"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/client/dto"
)

const defaultProductPageSize uint32 = 20

var ErrUSDTExchangeRateRequired = errors.New("USDT exchange rate is required")

type service struct {
	txManager  txManager
	repository repository
}

func New(txManager txManager, repository repository) *service {
	return &service{
		txManager:  txManager,
		repository: repository,
	}
}

func (s *service) GetCategories(ctx context.Context, req dto.GetCategoriesRequest) (categories domain.CategoryList, err error) {
	categories, err = s.repository.SelectCategories(ctx)
	if err != nil {
		return nil, err
	}

	return filterCategories(categories, req), nil
}

func (s *service) ListProducts(
	ctx context.Context,
	req dto.ListProductsRequest,
) (products domain.ProductList, nextPageToken uint64, err error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = defaultProductPageSize
	}

	products, err = s.repository.SelectProducts(ctx, req, pageSize+1)
	if err != nil {
		return nil, 0, err
	}

	if len(products) <= int(pageSize) {
		return products, 0, nil
	}

	nextPageToken = uint64(products[pageSize-1].ID)

	return products[:pageSize], nextPageToken, nil
}

func (s *service) GetProduct(ctx context.Context, productID int64) (product domain.Product, err error) {
	return s.repository.SelectProduct(ctx, productID)
}

func (s *service) GetVendorProducts(ctx context.Context, vendorID int64) (products domain.ProductList, err error) {
	return s.repository.SelectVendorProducts(ctx, vendorID)
}

func (s *service) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (product domain.Product, err error) {
	if err := s.validateExchangeRate(ctx, req.AcceptsCrypto, req.CryptoPricingMode); err != nil {
		return product, err
	}

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		product, err = s.repository.CreateProduct(ctx, req)
		return err
	})
	if err != nil {
		return product, err
	}

	return product, nil
}

func (s *service) UpdateProduct(ctx context.Context, req dto.UpdateProductRequest) (product domain.Product, err error) {
	if err := s.validateExchangeRate(ctx, req.AcceptsCrypto, req.CryptoPricingMode); err != nil {
		return product, err
	}

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		product, err = s.repository.UpdateProduct(ctx, req)
		return err
	})
	if err != nil {
		return product, err
	}

	return product, nil
}

func (s *service) validateExchangeRate(ctx context.Context, acceptsCrypto bool, mode string) error {
	if !acceptsCrypto || mode != "rub_rate" {
		return nil
	}

	exists, err := s.repository.HasUSDTExchangeRate(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUSDTExchangeRateRequired
	}

	return nil
}

func (s *service) DeleteProduct(ctx context.Context, req dto.DeleteProductRequest) (err error) {
	return s.txManager.Do(ctx, func(ctx context.Context) error {
		return s.repository.DeleteProduct(ctx, req)
	})
}

func filterCategories(categories domain.CategoryList, req dto.GetCategoriesRequest) domain.CategoryList {
	filtered := make(domain.CategoryList, 0, len(categories))

	for _, category := range categories {
		if req.FilterByParent && categoryParentID(category.ParentID) != req.ParentID {
			continue
		}

		if req.IncludeChildren && !req.FilterByParent && categoryParentID(category.ParentID) != 0 {
			continue
		}

		if !req.IncludeChildren {
			category.Children = nil
		} else {
			category.Children = buildCategoryChildren(categories, category.ID)
		}

		filtered = append(filtered, category)
	}

	return filtered
}

func buildCategoryChildren(categories domain.CategoryList, parentID int64) domain.CategoryList {
	children := make(domain.CategoryList, 0)

	for _, category := range categories {
		if categoryParentID(category.ParentID) != parentID {
			continue
		}

		category.Children = buildCategoryChildren(categories, category.ID)
		children = append(children, category)
	}

	return children
}

func categoryParentID(parentID *int64) int64 {
	if parentID == nil {
		return 0
	}

	return *parentID
}
