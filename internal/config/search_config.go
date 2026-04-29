package config

import (
	"flag"
	"fmt"
	"strings"
)

const (
	WB     = "wb"
	OZON   = "ozon"
	AVITO  = "avito"
	YANDEX = "yandex"
)

var availableMarkets = map[string]struct{}{WB: {}, OZON: {}, AVITO: {}, YANDEX: {}}

var (
	marketplaces = flag.String("markets", WB, marketsNameToString(availableMarkets))
	query        = flag.String("query", "коврик для мышки", "Search query")
	sort         = flag.String("sort", "popular", "Sort by \"popular\" or \"rating\", etc")
	limit        = flag.Int("limit", 10, "Top N results")
	page         = flag.Int("page", 1, "Page number")
	id           = flag.String("id", "200135094", "Product ID (product info on the individual marketplace)")
)

type SearchConfig struct {
	Marketplaces []string
	Query        string
	SortBy       string
	Page         int
	Limit        int
	ProductID    string //for search one product
}

func NewSearchConfig() (*SearchConfig, error) {
	flag.Parse()
	marketList := strings.Split(*marketplaces, ",")

	for i, market := range marketList {
		marketS := strings.TrimSpace(market)

		if _, ok := availableMarkets[marketS]; !ok {
			return nil, fmt.Errorf("invalid marketplace: %s", market)
		}

		marketList[i] = marketS
	}

	return &SearchConfig{
		Marketplaces: marketList,
		Query:        *query,
		SortBy:       *sort,
		Page:         *page,
		Limit:        *limit,
		ProductID:    *id,
	}, nil
}

func marketsNameToString(m map[string]struct{}) string {
	var list []string

	for market := range m {
		list = append(list, market)
	}

	return strings.Join(list, ", ")

}
