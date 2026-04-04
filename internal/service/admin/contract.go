package admin

import (
	"context"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/admin/dto"
)

type (
	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}
	repository interface {
		SelectCategories(ctx context.Context) (categories domain.CategoryList, err error)
		CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (category domain.Category, err error)
		UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (category domain.Category, err error)
		DeleteCategory(ctx context.Context, categoryID int64) (deletedCategoryID int64, err error)
	}
)
