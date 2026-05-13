package wb

import (
	"crawler/internal/config"
	"crawler/internal/parser"
	"fmt"
)

type WBTopProducts struct {
	Products []*WBProduct `json:"products"`
}

type WBProduct struct {
	ID             int64   `json:"id"`
	Brand          string  `json:"brand"`
	Name           string  `json:"name"`
	Supplier       string  `json:"supplier"`
	SupplierRating float64 `json:"supplierRating"`
	ProductRating  float64 `json:"reviewRating"`
	FeedbackCount  int64   `json:"feedbacks"`
	Sizes          []struct {
		Price struct {
			BasicPrice  int64 `json:"basic"`
			ActualPrice int64 `json:"product"`
		} `json:"price"`
	} `json:"sizes"`
	TotalQuantity int64 `json:"totalQuantity"`
}

func (p *WBProduct) ToBaseProduct() *parser.BaseProduct {
	return &parser.BaseProduct{
		Marketplace:    config.WB,
		ProductID:      fmt.Sprintf("%d", p.ID),
		Brand:          p.Brand,
		Name:           p.Name,
		Supplier:       p.Supplier,
		SupplierRating: p.SupplierRating,
		ProductRating:  p.ProductRating,
		Feedbacks:      p.FeedbackCount,
		PriceBasic:     p.Sizes[0].Price.BasicPrice / 100,
		PriceActual:    p.Sizes[0].Price.ActualPrice / 100,
		Quantity:       p.TotalQuantity,
	}
}

func (r *WBTopProducts) ToBaseProducts() []*parser.BaseProduct {
	var products []*parser.BaseProduct
	for _, product := range r.Products {
		products = append(products, product.ToBaseProduct())
	}

	return products
}
