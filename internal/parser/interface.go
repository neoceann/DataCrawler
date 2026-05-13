package parser

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/repository/db"
)

type Parser interface {
	Name() string

	GetTopProducts(ctx context.Context, s *config.SearchConfig) ([]*BaseProduct, error)

	GetProductByID(ctx context.Context, productID string) (*BaseProduct, error)

	Close()
}

type BaseProduct struct {
	Marketplace    string  `json:"marketplace"`
	ProductID      string  `json:"product_id"`
	Brand          string  `json:"brand"`
	Name           string  `json:"name"`
	Supplier       string  `json:"supplier"`
	SupplierRating float64 `json:"supplier_rating"`
	ProductRating  float64 `json:"product_rating"`
	Feedbacks      int64   `json:"feedbacks"`
	PriceBasic     int64   `json:"price_basic"`
	PriceActual    int64   `json:"price_actual"`
	Quantity       int64   `json:"quantity"`
}

func (p *BaseProduct) ToDbParams() *db.CreateProductParams {
	return &db.CreateProductParams{
		Marketplace:    p.Marketplace,
		ProductID:      p.ProductID,
		Brand:          p.Brand,
		Name:           p.Name,
		Supplier:       p.Supplier,
		SupplierRating: p.SupplierRating,
		ProductRating:  p.ProductRating,
		Feedbacks:      p.Feedbacks,
		PriceBasic:     p.PriceBasic,
		PriceActual:    p.PriceActual,
		Quantity:       p.Quantity,
	}
}
