package ozon

import (
	"crawler/internal/config"
	"crawler/internal/parser"
	"strconv"
	"strings"
)

type OzonProduct struct {
	ID              string `json:"sku"`
	Name            string `json:"name"`
	Brand           string `json:"brand"`
	AggregateRating struct {
		RatingValue string `json:"ratingValue"`
		ReviewCount string `json:"reviewCount"`
	} `json:"aggregateRating"`
	Offers struct {
		Price        string `json:"price"`
		Availability string `json:"availability"`
	} `json:"offers"`
}

func (p *OzonProduct) ToBaseProduct() *parser.BaseProduct {
	supplier := "unknown"
	supplierRate := 0.0

	productRating, err := strconv.ParseFloat(p.AggregateRating.RatingValue, 64)
	if err != nil {
		productRating = 0
	}
	feedbacks, err := strconv.Atoi(p.AggregateRating.ReviewCount)
	if err != nil {
		feedbacks = 0
	}

	price, err := strconv.Atoi(p.Offers.Price)
	if err != nil {
		price = 0
	}

	inStock := strings.Contains(p.Offers.Availability, "InStock")
	quantity := 1
	if !inStock {
		quantity = 0
	}

	return &parser.BaseProduct{
		Marketplace:    config.OZON,
		ProductID:      p.ID,
		Brand:          p.Brand,
		Name:           p.Name,
		Supplier:       supplier,
		SupplierRating: supplierRate,
		ProductRating:  productRating,
		Feedbacks:      int64(feedbacks),
		PriceBasic:     0,
		PriceActual:    int64(price),
		Quantity:       int64(quantity),
	}
}
