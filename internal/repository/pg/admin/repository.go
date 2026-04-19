package admin

import (
	"context"

	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/jmoiron/sqlx"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/admin/dto"
)

type repository struct {
	ctxGetter *trmsqlx.CtxGetter
	db        *sqlx.DB
}

func New(db *sqlx.DB, ctxGetter *trmsqlx.CtxGetter) *repository {
	return &repository{
		db:        db,
		ctxGetter: ctxGetter,
	}
}

func (r *repository) SelectCategories(ctx context.Context) (categories domain.CategoryList, err error) {
	query := `
		select
			id,
			name,
			parent_id,
			created_at,
			updated_at
		from categories
		order by id
	`

	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &categories, query)
	if err != nil {
		return categories, err
	}

	return categories, nil
}

func (r *repository) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (category domain.Category, err error) {
	query := `
		insert into categories (
			name,
			parent_id
		)
		values ($1, $2)
		returning id
	`

	var categoryID int64
	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &categoryID, query, req.Name, req.ParentID)
	if err != nil {
		return category, err
	}

	return r.selectCategory(ctx, categoryID)
}

func (r *repository) UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (category domain.Category, err error) {
	query := `
		update categories
		set
			name = $2,
			parent_id = $3
		where id = $1
		returning id
	`

	var categoryID int64
	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &categoryID, query, req.CategoryID, req.Name, req.ParentID)
	if err != nil {
		return category, err
	}

	return r.selectCategory(ctx, categoryID)
}

func (r *repository) DeleteCategory(ctx context.Context, categoryID int64) (deletedCategoryID int64, err error) {
	query := `
		with recursive category_tree as (
			select id, parent_id
			from categories
			where id = $1

			union all

			select c.id, c.parent_id
			from categories c
			join category_tree ct on c.parent_id = ct.id
		),
		deleted as (
			delete from categories
			where id in (select id from category_tree)
			returning id
		)
		select id
		from deleted
		where id = $1
	`

	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &deletedCategoryID, query, categoryID)
	if err != nil {
		return 0, err
	}

	return deletedCategoryID, nil
}

func (r *repository) selectCategory(ctx context.Context, categoryID int64) (category domain.Category, err error) {
	query := `
		select
			id,
			name,
			parent_id,
			created_at,
			updated_at
		from categories
		where id = $1
	`

	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &category, query, categoryID)
	if err != nil {
		return category, err
	}

	return category, nil
}
