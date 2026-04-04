package client

import (
	"context"

	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	"github.com/jmoiron/sqlx"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/client/dto"
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

func (r *repository) SelectVendorProducts(ctx context.Context, vendorID int64) (products domain.ProductList, err error) {
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
		where vendor_id = $1
		order by id
	`

	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &products, query, vendorID)
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

func (r *repository) SelectProducts(
	ctx context.Context,
	req dto.ListProductsRequest,
	limit uint32,
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
		where ($1 = 0 or category_id = $1)
			and ($2 = 0 or vendor_id = $2)
			and id > $3
		order by id
		limit $4
	`

	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).SelectContext(
		ctx,
		&products,
		query,
		req.CategoryID,
		req.VendorID,
		int64(req.PageToken),
		int64(limit),
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

func (r *repository) SelectProduct(ctx context.Context, productID int64) (product domain.Product, err error) {
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
		where id = $1
	`

	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &product, query, productID)
	if err != nil {
		return product, err
	}

	products := domain.ProductList{product}

	if err = r.loadProductsRelations(ctx, products); err != nil {
		return product, err
	}

	return products[0], nil
}

func (r *repository) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (product domain.Product, err error) {
	query := `
		insert into products (
			vendor_id,
			category_id,
			name,
			description,
			price,
			stock_count
		)
		values ($1, $2, $3, $4, $5, $6)
		returning id
	`

	var productID int64
	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).GetContext(
		ctx,
		&productID,
		query,
		req.VendorID,
		req.CategoryID,
		req.Name,
		req.Description,
		req.Price,
		int64(req.StockCount),
	)
	if err != nil {
		return product, err
	}

	if err = r.replaceProductAttributes(ctx, productID, req.Attributes); err != nil {
		return product, err
	}

	if err = r.replaceProductImages(ctx, productID, req.Images); err != nil {
		return product, err
	}

	return r.SelectProduct(ctx, productID)
}

func (r *repository) UpdateProduct(ctx context.Context, req dto.UpdateProductRequest) (product domain.Product, err error) {
	query := `
		update products
		set
			category_id = $3,
			name = $4,
			description = $5,
			price = $6,
			stock_count = $7
		where id = $1
			and vendor_id = $2
		returning id
	`

	var productID int64
	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).GetContext(
		ctx,
		&productID,
		query,
		req.ProductID,
		req.VendorID,
		req.CategoryID,
		req.Name,
		req.Description,
		req.Price,
		int64(req.StockCount),
	)
	if err != nil {
		return product, err
	}

	if err = r.replaceProductAttributes(ctx, productID, req.Attributes); err != nil {
		return product, err
	}

	if err = r.replaceProductImages(ctx, productID, req.Images); err != nil {
		return product, err
	}

	return r.SelectProduct(ctx, productID)
}

func (r *repository) DeleteProduct(ctx context.Context, req dto.DeleteProductRequest) error {
	query := `
		delete from products
		where id = $1
			and vendor_id = $2
		returning id
	`

	var productID int64
	return r.ctxGetter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &productID, query, req.ProductID, req.VendorID)
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

func (r *repository) replaceProductAttributes(
	ctx context.Context,
	productID int64,
	attributes []dto.ProductAttributeInput,
) error {
	db := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	if _, err := db.ExecContext(ctx, `delete from product_attributes where product_id = $1`, productID); err != nil {
		return err
	}

	if len(attributes) == 0 {
		return nil
	}

	query := `
		insert into product_attributes (
			product_id,
			attribute_name,
			attribute_value
		)
		values ($1, $2, $3)
	`

	for _, attribute := range attributes {
		if _, err := db.ExecContext(ctx, query, productID, attribute.Name, attribute.Value); err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) replaceProductImages(
	ctx context.Context,
	productID int64,
	images []dto.ProductImageInput,
) error {
	db := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	if _, err := db.ExecContext(ctx, `delete from product_images where product_id = $1`, productID); err != nil {
		return err
	}

	if len(images) == 0 {
		return nil
	}

	query := `
		insert into product_images (
			product_id,
			url,
			is_main
		)
		values ($1, $2, $3)
	`

	for _, image := range images {
		if _, err := db.ExecContext(ctx, query, productID, image.URL, image.IsMain); err != nil {
			return err
		}
	}

	return nil
}
