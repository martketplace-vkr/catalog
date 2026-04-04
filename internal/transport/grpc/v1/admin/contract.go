package admin

import (
	"context"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/admin/dto"
)

type (
	service interface {
		CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (category domain.Category, err error)
		UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (category domain.Category, err error)
		DeleteCategory(ctx context.Context, categoryID int64) (deletedCategoryID int64, err error)
	}
)
