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

const productSelectFields = `
	p.id,
	p.vendor_id,
	p.category_id,
	p.name,
	p.description,
	p.price::text as price,
	p.accepts_crypto,
	p.crypto_pricing_mode,
	coalesce(p.crypto_price_usdt::text, '') as crypto_price_usdt,
	coalesce(case
		when p.accepts_crypto and p.crypto_pricing_mode = 'fixed_usdt' then p.crypto_price_usdt::text
		when p.accepts_crypto and p.crypto_pricing_mode = 'rub_rate' and er.rub_per_usdt is not null
			then round((p.price / er.rub_per_usdt)::numeric, 8)::text
		else ''
	end, '') as effective_usdt_price,
	coalesce(er.rub_per_usdt::text, '') as rub_per_usdt,
	p.stock_count,
	p.created_at,
	p.updated_at
`

func (r *repository) SelectProducts(
	ctx context.Context,
	productIds []int64,
) (products domain.ProductList, err error) {
	query := `
		select ` + productSelectFields + `
		from products p
		left join platform_exchange_rates er on er.currency_pair = 'RUB_USDT'
		where p.id = any($1)
		order by p.id
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
