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

	if err = r.replaceProductCharacteristics(ctx, productID, req.Characteristics, req.Attributes); err != nil {
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

	if err = r.replaceProductCharacteristics(ctx, productID, req.Characteristics, req.Attributes); err != nil {
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

type productCharacteristicRow struct {
	ProductID int64  `db:"product_id"`
	ID        int64  `db:"id"`
	Title     string `db:"title"`
}

type productCharacteristicAttributeRow struct {
	ProductID        int64  `db:"product_id"`
	CharacteristicID int64  `db:"characteristic_id"`
	ID               int64  `db:"id"`
	Name             string `db:"name"`
	Value            string `db:"value"`
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

	characteristics, err := r.selectProductCharacteristics(ctx, productIDs)
	if err != nil {
		return err
	}

	characteristicIndexByID := make(map[int64]map[int64]int, len(products))
	for _, characteristic := range characteristics {
		productIdx := productIndexByID[characteristic.ProductID]
		products[productIdx].Characteristics = append(products[productIdx].Characteristics, domain.ProductCharacteristic{
			ID:         characteristic.ID,
			Title:      characteristic.Title,
			Attributes: domain.ProductAttributeList{},
		})

		if characteristicIndexByID[characteristic.ProductID] == nil {
			characteristicIndexByID[characteristic.ProductID] = make(map[int64]int)
		}

		characteristicIndexByID[characteristic.ProductID][characteristic.ID] = len(products[productIdx].Characteristics) - 1
	}

	attributes, err := r.selectProductCharacteristicAttributes(ctx, productIDs)
	if err != nil {
		return err
	}

	for _, attribute := range attributes {
		productIdx := productIndexByID[attribute.ProductID]
		characteristicIdx := characteristicIndexByID[attribute.ProductID][attribute.CharacteristicID]
		nextAttribute := domain.ProductAttribute{
			ID:    attribute.ID,
			Name:  attribute.Name,
			Value: attribute.Value,
		}

		products[productIdx].Characteristics[characteristicIdx].Attributes = append(
			products[productIdx].Characteristics[characteristicIdx].Attributes,
			nextAttribute,
		)
		products[productIdx].Attributes = append(products[productIdx].Attributes, nextAttribute)
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

func (r *repository) selectProductCharacteristics(ctx context.Context, productIDs []int64) ([]productCharacteristicRow, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
		select
			id,
			product_id,
			title
		from product_characteristics
		where product_id in (?)
		order by product_id, sort_order, id
	`, productIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var rows []productCharacteristicRow
	err = r.ctxGetter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *repository) selectProductCharacteristicAttributes(ctx context.Context, productIDs []int64) ([]productCharacteristicAttributeRow, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
		select
			pca.id,
			pc.product_id,
			pca.characteristic_id,
			pca.attribute_name as name,
			pca.attribute_value as value
		from product_characteristic_attributes pca
		join product_characteristics pc on pc.id = pca.characteristic_id
		where pc.product_id in (?)
		order by pc.product_id, pca.characteristic_id, pca.sort_order, pca.id
	`, productIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var rows []productCharacteristicAttributeRow
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

func (r *repository) replaceProductCharacteristics(
	ctx context.Context,
	productID int64,
	characteristics []dto.ProductCharacteristicInput,
	legacyAttributes []dto.ProductAttributeInput,
) error {
	db := r.ctxGetter.DefaultTrOrDB(ctx, r.db)

	if _, err := db.ExecContext(ctx, `delete from product_characteristics where product_id = $1`, productID); err != nil {
		return err
	}

	if len(characteristics) == 0 {
		characteristics = legacyCharacteristicsFromAttributes(legacyAttributes)
	}

	if len(characteristics) == 0 {
		return nil
	}

	characteristicQuery := `
		insert into product_characteristics (
			product_id,
			title,
			sort_order
		)
		values ($1, $2, $3)
		returning id
	`

	attributeQuery := `
		insert into product_characteristic_attributes (
			characteristic_id,
			attribute_name,
			attribute_value,
			sort_order
		)
		values ($1, $2, $3, $4)
	`

	for sectionIndex, characteristic := range characteristics {
		var characteristicID int64
		if err := db.GetContext(ctx, &characteristicID, characteristicQuery, productID, characteristic.Title, sectionIndex); err != nil {
			return err
		}

		for attributeIndex, attribute := range characteristic.Attributes {
			if _, err := db.ExecContext(
				ctx,
				attributeQuery,
				characteristicID,
				attribute.Name,
				attribute.Value,
				attributeIndex,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func legacyCharacteristicsFromAttributes(attributes []dto.ProductAttributeInput) []dto.ProductCharacteristicInput {
	if len(attributes) == 0 {
		return nil
	}

	result := []dto.ProductCharacteristicInput{
		{
			Title:      "Основная информация",
			Attributes: make([]dto.ProductAttributeInput, 0, len(attributes)),
		},
	}

	for _, attribute := range attributes {
		result[0].Attributes = append(result[0].Attributes, dto.ProductAttributeInput{
			Name:  attribute.Name,
			Value: attribute.Value,
		})
	}

	return result
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
