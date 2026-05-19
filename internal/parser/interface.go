package parser

import (
	"context"
	"crawler/internal/config"
	m "crawler/internal/parser/parser_model"
	cache "crawler/internal/redis"
)

type BaseProduct = m.BaseProduct

type Parser interface {
	Name() string

	GetTopProducts(ctx context.Context, s *config.SearchConfig, cache *cache.ProductCache) ([]*BaseProduct, error)

	ParseProductPage(ctx context.Context, pageURL string) (*BaseProduct, error)

	Close()
}
