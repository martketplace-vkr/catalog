package admin

import (
	"context"
	"errors"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/admin/dto"
)

var ErrCategoryCycle = errors.New("category cycle detected")

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

func (s *service) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (category domain.Category, err error) {
	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		category, err = s.repository.CreateCategory(ctx, req)
		return err
	})
	if err != nil {
		return category, err
	}

	return category, nil
}

func (s *service) UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (category domain.Category, err error) {
	if req.ParentID != nil {
		categories, err := s.repository.SelectCategories(ctx)
		if err != nil {
			return category, err
		}

		if categoryCreatesCycle(categories, req.CategoryID, *req.ParentID) {
			return category, ErrCategoryCycle
		}
	}

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		category, err = s.repository.UpdateCategory(ctx, req)
		return err
	})
	if err != nil {
		return category, err
	}

	return category, nil
}

func (s *service) DeleteCategory(ctx context.Context, categoryID int64) (deletedCategoryID int64, err error) {
	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		deletedCategoryID, err = s.repository.DeleteCategory(ctx, categoryID)
		return err
	})
	if err != nil {
		return 0, err
	}

	return deletedCategoryID, nil
}

func (s *service) GetUSDTExchangeRate(ctx context.Context) (rate domain.ExchangeRate, err error) {
	return s.repository.GetUSDTExchangeRate(ctx)
}

func (s *service) UpdateUSDTExchangeRate(ctx context.Context, req dto.UpdateExchangeRateRequest) (rate domain.ExchangeRate, err error) {
	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		rate, err = s.repository.UpdateUSDTExchangeRate(ctx, req)
		return err
	})
	if err != nil {
		return rate, err
	}

	return rate, nil
}

func categoryCreatesCycle(categories domain.CategoryList, categoryID int64, parentID int64) bool {
	parentByID := make(map[int64]*int64, len(categories))

	for _, category := range categories {
		parentByID[category.ID] = category.ParentID
	}

	currentID := &parentID
	visited := make(map[int64]struct{}, len(categories))

	for currentID != nil {
		if *currentID == categoryID {
			return true
		}

		if _, ok := visited[*currentID]; ok {
			return true
		}

		visited[*currentID] = struct{}{}
		currentID = parentByID[*currentID]
	}

	return false
}
