package cache

import (
	"context"
	m "crawler/internal/parser/parser_model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProductCache struct {
	client *redis.Client
	Ttl    time.Duration
}

func NewProductCache(addr string, ttl time.Duration) (*ProductCache, error) {
	c := redis.NewClient(&redis.Options{Addr: addr})

	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &ProductCache{client: c, Ttl: ttl}, nil
}

func (c *ProductCache) Close() error {
	return c.client.Close()
}

func (c *ProductCache) Get(ctx context.Context, market, productID string) (*m.BaseProduct, error) {
	key := c.keyFormation(market, productID)

	data, err := c.client.Get(ctx, key).Bytes()

	if err == redis.Nil {
		return nil, fmt.Errorf("key not found in redis")
	}

	if err != nil {
		return nil, err
	}

	var product m.BaseProduct

	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (c *ProductCache) Set(ctx context.Context, product *m.BaseProduct) error {
	key := c.keyFormation(product.Marketplace, product.ProductID)

	data, err := json.Marshal(product)

	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.Ttl).Err()

}

func (c *ProductCache) keyFormation(market, productID string) string {
	return fmt.Sprintf("product:%s:%s", market, productID)
}
