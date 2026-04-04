package domain

import (
	"time"

	pbdomain "github.com/martketplace-vkr/catalog/pkg/api/grpc/v1/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Category struct {
	ID        int64        `db:"id"`
	Name      string       `db:"name"`
	ParentID  *int64       `db:"parent_id"`
	Children  CategoryList `db:"-"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
}

type CategoryList []Category

func (c Category) ToProto() *pbdomain.Category {
	return &pbdomain.Category{
		Id:        c.ID,
		Name:      c.Name,
		ParentId:  nullableInt64ToProto(c.ParentID),
		Children:  c.Children.ToProto(),
		CreatedAt: timeToProto(c.CreatedAt),
		UpdatedAt: timeToProto(c.UpdatedAt),
	}
}

func (l CategoryList) ToProto() []*pbdomain.Category {
	result := make([]*pbdomain.Category, 0, len(l))

	for _, category := range l {
		result = append(result, category.ToProto())
	}

	return result
}

type Product struct {
	ID          int64                `db:"id"`
	VendorID    int64                `db:"vendor_id"`
	CategoryID  int64                `db:"category_id"`
	Name        string               `db:"name"`
	Description string               `db:"description"`
	Price       string               `db:"price"`
	StockCount  int64                `db:"stock_count"`
	Attributes  ProductAttributeList `db:"-"`
	Images      ProductImageList     `db:"-"`
	CreatedAt   time.Time            `db:"created_at"`
	UpdatedAt   time.Time            `db:"updated_at"`
}

type ProductList []Product

func (p Product) ToProto() *pbdomain.Product {
	return &pbdomain.Product{
		Id:          p.ID,
		VendorId:    p.VendorID,
		CategoryId:  p.CategoryID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		StockCount:  uint32(p.StockCount),
		Attributes:  p.Attributes.ToProto(),
		Images:      p.Images.ToProto(),
		CreatedAt:   timeToProto(p.CreatedAt),
		UpdatedAt:   timeToProto(p.UpdatedAt),
	}
}

func (l ProductList) ToProto() []*pbdomain.Product {
	result := make([]*pbdomain.Product, 0, len(l))

	for _, product := range l {
		result = append(result, product.ToProto())
	}

	return result
}

type ProductAttribute struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Value string `db:"value"`
}

type ProductAttributeList []ProductAttribute

func (l ProductAttributeList) ToProto() []*pbdomain.ProductAttribute {
	result := make([]*pbdomain.ProductAttribute, 0, len(l))

	for _, attribute := range l {
		result = append(result, &pbdomain.ProductAttribute{
			Id:    attribute.ID,
			Name:  attribute.Name,
			Value: attribute.Value,
		})
	}

	return result
}

type ProductImage struct {
	ID     int64  `db:"id"`
	URL    string `db:"url"`
	IsMain bool   `db:"is_main"`
}

type ProductImageList []ProductImage

func (l ProductImageList) ToProto() []*pbdomain.ProductImage {
	result := make([]*pbdomain.ProductImage, 0, len(l))

	for _, image := range l {
		result = append(result, &pbdomain.ProductImage{
			Id:     image.ID,
			Url:    image.URL,
			IsMain: image.IsMain,
		})
	}

	return result
}

func nullableInt64ToProto(value *int64) int64 {
	if value == nil {
		return 0
	}

	return *value
}

func timeToProto(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}

	return timestamppb.New(value)
}
