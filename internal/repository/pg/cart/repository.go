package cart

import (
	"context"

	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/jmoiron/sqlx"
	"github.com/martketplace-vkr/catalog/domain"
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

func (r *repository) SelectProducts(
	ctx context.Context,
	productIds []int64,
) (products domain.ProductList, err error) {
	query := `
		select
			id,
			vendor_id,
			category_id,
			name,
			description,
			price::text as price,
			stock_count,
			created_at,
			updated_at
		from products
		where id = any($1)
		order by id
	`

	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).SelectContext(
		ctx,
		&products,
		query,
		productIds,
	)
	if err != nil {
		return products, err
	}

	if len(products) == 0 {
		return products, nil
	}

	if err = r.loadProductsRelations(ctx, products); err != nil {
		return nil, err
	}

	return products, nil
}

type productAttributeRow struct {
	ProductID int64  `db:"product_id"`
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	Value     string `db:"value"`
}

type productImageRow struct {
	ProductID int64  `db:"product_id"`
	ID        int64  `db:"id"`
	URL       string `db:"url"`
	IsMain    bool   `db:"is_main"`
}

func (r *repository) loadProductsRelations(ctx context.Context, products domain.ProductList) error {
	productIDs := make([]int64, 0, len(products))
	productIndexByID := make(map[int64]int, len(products))

	for idx, product := range products {
		productIDs = append(productIDs, product.ID)
		productIndexByID[product.ID] = idx
	}

	attributes, err := r.selectProductAttributes(ctx, productIDs)
	if err != nil {
		return err
	}

	for _, attribute := range attributes {
		productIdx := productIndexByID[attribute.ProductID]
		products[productIdx].Attributes = append(products[productIdx].Attributes, domain.ProductAttribute{
			ID:    attribute.ID,
			Name:  attribute.Name,
			Value: attribute.Value,
		})
	}

	images, err := r.selectProductImages(ctx, productIDs)
	if err != nil {
		return err
	}

	for _, image := range images {
		productIdx := productIndexByID[image.ProductID]
		products[productIdx].Images = append(products[productIdx].Images, domain.ProductImage{
			ID:     image.ID,
			URL:    image.URL,
			IsMain: image.IsMain,
		})
	}

	return nil
}

func (r *repository) selectProductAttributes(ctx context.Context, productIDs []int64) ([]productAttributeRow, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
		select
			id,
			product_id,
			attribute_name as name,
			attribute_value as value
		from product_attributes
		where product_id in (?)
		order by id
	`, productIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var rows []productAttributeRow
	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *repository) selectProductImages(ctx context.Context, productIDs []int64) ([]productImageRow, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
		select
			id,
			product_id,
			url,
			is_main
		from product_images
		where product_id in (?)
		order by is_main desc, id
	`, productIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var rows []productImageRow
	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
