package avito

import (
	"crawler/internal/config"
	"crawler/internal/parser"
	"strconv"
)

type AvitoProduct struct {
	ID             string
	Name           string
	Price          string
	Supplier       string
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
		Marketplace:    config.AVITO,
		ProductID:      p.ID,
		Name:           p.Name,
		PriceActual:    int64(price),
		Supplier:       p.Supplier,
		SupplierRating: supplierRate,
	}
}
