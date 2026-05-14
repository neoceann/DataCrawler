package avito

import (
	"crawler/internal/config"
	"crawler/internal/parser"
	"strconv"
	"strings"
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

	priceTemp := strings.Fields(p.Price)
	priceTemp = priceTemp[:len(priceTemp)-1]

	price := ""
	for _, pricePart := range priceTemp {
		price += pricePart
	}
	priceInt, err := strconv.Atoi(price)

	if err != nil {
		priceInt = 0
	}

	return &parser.BaseProduct{
		Marketplace:    config.AVITO,
		ProductID:      p.ID,
		Name:           p.Name,
		PriceActual:    int64(priceInt),
		Supplier:       p.Supplier,
		SupplierRating: supplierRate,
	}
}
