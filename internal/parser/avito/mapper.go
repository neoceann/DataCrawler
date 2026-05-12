package avito

import (
	"crawler/internal/config"
	"crawler/internal/parser"
	"strconv"
)

type AvitoProduct struct {
	ID string
	Name string
	Price string
	Supplier string
	SupplierRating string
}

func (p *AvitoProduct) ToBaseProduct() *parser.BaseProduct {
	supplierRate, err := strconv.ParseFloat(p.SupplierRating, 64)

	if err != nil {
		supplierRate = 0
	}

	price, err := strconv.Atoi(p.Price)

	if err != nil {
		price = 0
	}

	return &parser.BaseProduct{
		Marketplace: config.AVITO,
		ProductID: p.ID,
		Name: p.Name,
		PriceActual: int64(price),
		Supplier: p.Supplier,
		SupplierRating: supplierRate,
	}
}

// type BaseProduct struct {
// 	Marketplace    string  `json:"marketplace"`
// 	ProductID      string  `json:"product_id"`
// 	Brand          string  `json:"brand"`
// 	Name           string  `json:"name"`
// 	Supplier       string  `json:"supplier"`
// 	SupplierRating float64 `json:"supplier_rating"`
// 	ProductRating  float64 `json:"product_rating"`
// 	Feedbacks      int64   `json:"feedbacks"`
// 	PriceBasic     int64   `json:"price_basic"`
// 	PriceActual    int64   `json:"price_actual"`
// 	Quantity       int64   `json:"quantity"`
// }