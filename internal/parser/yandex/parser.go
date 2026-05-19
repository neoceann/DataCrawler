package yandex

import (
	"context"
	"crawler/internal/browser"
	"crawler/internal/config"
	"crawler/internal/parser"
	parserErrors "crawler/internal/parser/errors"
	"crawler/internal/parser/helpers"
	cache "crawler/internal/redis"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

const (
	BaseURLSearch = "https://market.yandex.ru/search"
)

type YandexParser struct {
	cfg     *config.Config
	browser *browser.Browser
}

func NewYandexParser(cfg *config.Config) (*YandexParser, error) {
	browserCfg := browser.DefaultConfig()
	browserCfg.UserAgent = cfg.UserAgent

	browser, err := browser.New(browserCfg)

	if err != nil {
		return nil, err
	}

	return &YandexParser{
		browser: browser,
		cfg:     cfg,
	}, nil
}

func (p *YandexParser) Name() string {
	return config.YANDEX
}

func (p *YandexParser) Close() {
	p.browser.Close()
}

func (p *YandexParser) GetTopProducts(ctx context.Context, s *config.SearchConfig, cache *cache.ProductCache) ([]*parser.BaseProduct, error) {
	links, err := p.getProductLinks(ctx, s)
	if err != nil {
		return nil, err
	}

	log.Printf("Getting top %d products from %s...", s.Limit, p.Name())

	results := helpers.ParseProducts(ctx, p, 3, links, cache)

	var products []*parser.BaseProduct
	for product := range results {
		products = append(products, product)
		if len(products) >= s.Limit {
			break
		}
	}

	log.Printf("Received products from %s: %d", p.Name(), len(products))
	return products, nil
}

func (p *YandexParser) ParseProductPage(ctx context.Context, pageURL string, cache *cache.ProductCache) (*parser.BaseProduct, error) {
	productID := helpers.ExtractIDFromURL(pageURL)

	if cache != nil {
		product, err := cache.Get(ctx, p.Name(), productID)

		if err != nil {
			log.Printf("Cache error for: %s:%s:%v", p.Name(), productID, err)
		} else if product != nil {
			log.Printf("Product from cache: %s:%s", p.Name(), productID)
			return product, nil
		}
	}

	ctx, cancel := p.browser.NewTab(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(pageURL),
		chromedp.WaitVisible(`[itemprop="name"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp error: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("goquery parse error: %w", err)
	}

	// fullHTML, _ := doc.Html()
	// err = os.WriteFile("_page.html", []byte(fullHTML), 0644)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// log.Print("HTML сохранён в _page.html")

	var jsonLD string
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		jsonLD = s.Text()
	})

	if jsonLD == "" {
		return nil, parserErrors.ErrEmptyDataForProduct
	}

	var yandexProduct YandexProduct

	json.Unmarshal([]byte(jsonLD), &yandexProduct)

	yandexProduct.ID = productID

	baseProduct := yandexProduct.ToBaseProduct()
	if cache != nil {
		if err := cache.Set(ctx, baseProduct); err != nil {
			log.Printf("Failed to save data in cache: %v", err)
		}
	}

	return baseProduct, nil
}

func (p *YandexParser) getProductLinks(ctx context.Context, s *config.SearchConfig) ([]string, error) {
	ctx, cancel := p.browser.NewTab(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("%s?text=%s&how=%s", BaseURLSearch, url.QueryEscape(s.Query), s.GetSortParamForMarket(p.Name(), s.SortBy))

	var links []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`[data-auto="snippet-link"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(fmt.Sprintf(`
			Array.from(new Set(
				Array.from(document.querySelectorAll('[data-auto="snippet-link"]'))
					.map(a => a.href.split('?')[0])
			)).slice(0, %d)
		`, s.Limit), &links),
	)

	if len(links) == 0 {
		return nil, parserErrors.ErrEmptyLinks
	}

	return links, err
}
