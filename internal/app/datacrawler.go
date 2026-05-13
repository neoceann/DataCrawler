package app

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/parser"
	"crawler/internal/parser/avito"
	"crawler/internal/parser/ozon"
	"crawler/internal/parser/wb"
	"crawler/internal/parser/yandex"
	"crawler/internal/repository/db"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
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

	var parsers []parser.Parser

	wb := wb.NewWBParser(&c)

	ozon, err := ozon.NewOzonParser(&c)
	if err != nil {
		return nil, err
	}

	avito, err := avito.NewAvitoParser(&c)
	if err != nil {
		return nil, err
	}

	yandex, err := yandex.NewYandexParser(&c)
	if err != nil {
		return nil, err
	}

	parsers = append(parsers, wb, ozon, avito, yandex)

	return &DataCrawler{
		Parsers:      parsers,
		Pool:         pool,
		Queries:      queries,
		Config:       &c,
		SearchConfig: sc,
	}, nil
}

func (d *DataCrawler) Close() {
	for _, p := range d.Parsers {
		if err := p.Close(); err != nil {
			log.Printf("Error closing parser %s: %v", p.Name(), err)
		}
	}

	d.Pool.Close()
}

func (d *DataCrawler) GetProductByID(ctx context.Context, marketplace, productID string) (*parser.BaseProduct, error) {
	p, err := d.findParserForMarketplace(marketplace)

	if err != nil {
		return nil, err
	}

	processCtx, processCancel := context.WithTimeout(ctx, 15*time.Second)
	defer processCancel()

	return p.GetProductByID(processCtx, productID)
}

func (d *DataCrawler) GetTopProducts(ctx context.Context) ([]*parser.BaseProduct, error) {
	var mu sync.Mutex
	var products []*parser.BaseProduct
	var wg sync.WaitGroup
	errChan := make(chan error, len(d.SearchConfig.Marketplaces))

	for _, market := range d.SearchConfig.Marketplaces {
		wg.Add(1)

		go func(m string) {
			defer wg.Done()

			processCtx, processCancel := context.WithTimeout(ctx, 60*time.Second)
			defer processCancel()

			p, err := d.findParserForMarketplace(m)
			if err != nil {
				errChan <- fmt.Errorf("%s: %w", m, err)
				return
			}

			product, err := p.GetTopProducts(processCtx, d.SearchConfig)
			if err != nil {
				errChan <- fmt.Errorf("%s: %w", m, err)
				return
			}

			mu.Lock()
			products = append(products, product...)
			mu.Unlock()
		}(market)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return products, nil
}

func (d *DataCrawler) SaveProductsToDB(ctx context.Context, products []*parser.BaseProduct) error {
	ctx, processCancel := context.WithTimeout(ctx, 10*time.Second)
	defer processCancel()

	for _, product := range products {
		err := d.Queries.CreateProduct(ctx, *product.ToDbParams())
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DataCrawler) SaveProductsToFile(products []*parser.BaseProduct) {
	p, _ := json.MarshalIndent(products, "", "  ")
	os.WriteFile("products.json", p, 0644)
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
