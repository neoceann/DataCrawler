package ozon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	"crawler/internal/browser"
	"crawler/internal/config"
	"crawler/internal/parser"
	parserErrors "crawler/internal/parser/errors"
	"crawler/internal/parser/helpers"
	cache "crawler/internal/redis"
)

const (
	BaseURLSearch  = "https://www.ozon.ru/search/"
	BaseURLProduct = "https://www.ozon.ru/product/"
)

type OzonParser struct {
	cfg     *config.Config
	browser *browser.Browser
}

func NewOzonParser(cfg *config.Config) (*OzonParser, error) {
	browserCfg := browser.DefaultConfig()
	browserCfg.UserAgent = cfg.UserAgent

	browser, err := browser.New(browserCfg)

	if err != nil {
		return nil, err
	}

	return &OzonParser{
		browser: browser,
		cfg:     cfg,
	}, nil
}

func (p *OzonParser) Name() string {
	return config.OZON
}

func (p *OzonParser) Close() {
	p.browser.Close()
}

func (p *OzonParser) ParseProductPage(ctx context.Context, pageURL string) (*parser.BaseProduct, error) {
	productID := helpers.ExtractIDFromURL(pageURL)
	url := fmt.Sprintf("%s%s/", BaseURLProduct, productID)

	ctx, cancel := p.browser.NewTab(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var htmlContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),

		chromedp.WaitVisible(`h1`, chromedp.ByQuery),

		chromedp.Sleep(1*time.Second),

		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp error: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("goquery parse error: %w", err)
	}

	var jsonLD string
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		jsonLD = s.Text()
	})

	if jsonLD == "" {
		return nil, parserErrors.ErrEmptyDataForProduct
	}

	var ozonProduct OzonProduct

	json.Unmarshal([]byte(jsonLD), &ozonProduct)

	return ozonProduct.ToBaseProduct(), nil
}

func (p *OzonParser) GetTopProducts(ctx context.Context, s *config.SearchConfig, cache *cache.ProductCache) ([]*parser.BaseProduct, error) {
	searchURL := fmt.Sprintf("%s?text=%s&sorting=%s", BaseURLSearch, url.QueryEscape(s.Query), s.GetSortParamForMarket(p.Name(), s.SortBy))

	ctx, cancel := p.browser.NewTab(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	log.Printf("Getting top %d products from %s...", s.Limit, p.Name())
	var links []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`a[href*="/product/"]`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),

		chromedp.Evaluate(fmt.Sprintf(`
            Array.from(new Set(
                Array.from(document.querySelectorAll('a[href*="/product/"]'))
                    .map(a => a.href)
            )).slice(0, %d)
        `, s.Limit), &links),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp error: %w", err)
	}

	if len(links) == 0 {
		return nil, parserErrors.ErrEmptyLinks
	}

	results := helpers.ParseProducts(ctx, p, 3, links)

	var products []*parser.BaseProduct
	for p := range results {
		products = append(products, p)
	}

	log.Printf("Received products from %s: %d", p.Name(), len(products))

	return products, nil
}
