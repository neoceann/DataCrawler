package app

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/parser"
	"crawler/internal/parser/wb"
	"crawler/internal/repository/db"
	"fmt"
	"log"
	"time"
)

type DataCrawler struct {
	Parsers []parser.Parser
	Pool    *db.Pool
	Queries *db.Queries
	Config  *config.Config
}

func New(ctx context.Context) (*DataCrawler, error) {
	var c config.Config
	err := config.ReadEnv(&c)

	if err != nil {
		log.Fatal("Read config from env failed:", err.Error())
	}

	dbCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
	pool, err := db.NewPool(dbCtx, &c)
	dbCancel()

	if err != nil {
		log.Fatal(err)
	}

	queries := db.New(pool.Pool)

	parsers := []parser.Parser{
		wb.NewWBParser(queries, &c),
	}

	return &DataCrawler{
		Parsers: parsers,
		Pool:    pool,
		Queries: queries,
		Config:  &c,
	}, nil
}

func (d *DataCrawler) Close() {
	d.Pool.Close()
}

func (d *DataCrawler) GetProductByID(ctx context.Context, marketplace, productID string) (*parser.BaseProduct, error) {
	var p parser.Parser
	for _, parser := range d.Parsers {
		if parser.Name() == marketplace {
			p = parser
			break
		}
	}

	if p == nil {
		return nil, fmt.Errorf("invalid marketplace:%s", marketplace)
	}

	reqCtx, reqCancel := context.WithTimeout(ctx, 10*time.Second)
	defer reqCancel()

	return p.GetProductByID(reqCtx, productID)
}
