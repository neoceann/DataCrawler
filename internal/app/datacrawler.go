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
	Parsers      []parser.Parser
	Pool         *db.Pool
	Queries      *db.Queries
	Config       *config.Config
	SearchConfig *config.SearchConfig
}

func New(ctx context.Context) (*DataCrawler, error) {
	var c config.Config
	err := config.ReadEnv(&c)

	if err != nil {
		log.Fatal("Read config from env failed:", err.Error())
	}

	sc, err := config.NewSearchConfig()

	if err != nil {
		log.Fatal(err.Error())
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
		Parsers:      parsers,
		Pool:         pool,
		Queries:      queries,
		Config:       &c,
		SearchConfig: sc,
	}, nil
}

func (d *DataCrawler) Close() {
	d.Pool.Close()
}

func (d *DataCrawler) GetProductByID(ctx context.Context, marketplace, productID string) (*parser.BaseProduct, error) {
	p, err := d.findParserForMarketplace(marketplace)

	if err != nil {
		return nil, err
	}

	processCtx, processCancel := context.WithTimeout(ctx, 10*time.Second)
	defer processCancel()

	return p.GetProductByID(processCtx, productID)
}

func (d *DataCrawler) GetTopProducts(ctx context.Context) ([]*parser.BaseProduct, error) {
	var products []*parser.BaseProduct

	for _, market := range d.SearchConfig.Marketplaces {
		p, err := d.findParserForMarketplace(market)

		if err != nil {
			return nil, err
		}

		processCtx, processCancel := context.WithTimeout(ctx, 10*time.Second)
		defer processCancel()

		product, err := p.GetTopProducts(processCtx, d.SearchConfig)

		if err != nil {
			return nil, err
		}

		products = append(products, product...)
	}

	return products, nil
}

func (d *DataCrawler) SaveProductToDB(ctx context.Context, product *parser.BaseProduct) error {
	processCtx, processCancel := context.WithTimeout(ctx, 10*time.Second)
	defer processCancel()

	p, err := d.findParserForMarketplace(product.Marketplace)

	if err != nil {
		return err
	}

	return p.SaveProductToDB(processCtx, product)
}

func (d *DataCrawler) findParserForMarketplace(name string) (parser.Parser, error) {
	var p parser.Parser
	for _, parser := range d.Parsers {
		if parser.Name() == name {
			p = parser
			break
		}
	}

	if p == nil {
		return nil, fmt.Errorf("parser for \"%s\" failed", name)
	}

	return p, nil
}
