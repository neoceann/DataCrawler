package avito

import (
	"context"
	"crawler/internal/browser"
	"crawler/internal/config"
	"crawler/internal/parser"
	parserErrors "crawler/internal/parser/errors"
	"crawler/internal/parser/helpers"
	cache "crawler/internal/redis"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

const (
	BaseURLSearch = "https://www.avito.ru/kazan"
)

type AvitoParser struct {
	cfg     *config.Config
	browser *browser.Browser
}

func NewAvitoParser(cfg *config.Config) (*AvitoParser, error) {
	browserCfg := browser.DefaultConfig()
	browserCfg.UserAgent = cfg.UserAgent

	browser, err := browser.New(browserCfg)

	if err != nil {
		return nil, err
	}

	return &AvitoParser{
		browser: browser,
		cfg:     cfg,
	}, nil
}

func (p *AvitoParser) Name() string {
	return config.AVITO
}

func (p *AvitoParser) Close() {
	p.browser.Close()
}

func (p *AvitoParser) GetTopProducts(ctx context.Context, s *config.SearchConfig, cache *cache.ProductCache) ([]*parser.BaseProduct, error) {
	links, err := p.getProductLinks(ctx, s.Query, s.Limit)
	if err != nil {
		return nil, err
	}

	log.Printf("Getting top %d products from %s...", s.Limit, p.Name())

	results := helpers.ParseProducts(ctx, p, 1, links)

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

func (p *AvitoParser) GetProductByID(ctx context.Context, productID string) (*parser.BaseProduct, error) {
	return nil, nil
}

func (p *AvitoParser) ParseProductPage(ctx context.Context, pageURL string) (*parser.BaseProduct, error) {
	ctx, cancel := p.browser.NewTab(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(pageURL),
		//chromedp.WaitVisible(`h1`, chromedp.ByQuery),
		chromedp.WaitReady(`[data-marker="item-view/item-price"]`, chromedp.ByQuery),
		//chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp page error: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// fullHTML, _ := doc.Html()
	// err = os.WriteFile("_page.html", []byte(fullHTML), 0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Print("HTML сохранён в _page.html")

	product := &AvitoProduct{}

	product.ID = helpers.ExtractIDFromURL(pageURL)

	product.Name = strings.TrimSpace(doc.Find(`h1[data-marker="item-view/title"]`).Text())
	if product.Name == "" {
		product.Name = strings.TrimSpace(doc.Find(`h1`).First().Text())
	}

	product.Price = (doc.Find(`[data-marker="item-view/item-price"]`).First().Text())
	product.Supplier = doc.Find(`[data-marker="seller-info/name"]`).First().Text()
	product.SupplierRating, _ = doc.Find(`[data-marker="sellerRate"] meta[itemprop="ratingValue"]`).Attr("content")

	return product.ToBaseProduct(), nil
}

func (p *AvitoParser) getProductLinks(ctx context.Context, query string, limit int) ([]string, error) {
	ctx, cancel := p.browser.NewTab(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("%s?q=%s казань", BaseURLSearch, url.QueryEscape(query))

	var links []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`[data-marker="item-title"]`, chromedp.ByQuery),
		//chromedp.WaitVisible(`[data-marker="item"]`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(fmt.Sprintf(`
            (function() {
                const items = document.querySelectorAll('[data-marker="item"]');
                const links = [];
                for (const item of items) {
                    const linkElem = item.querySelector('a[data-marker="item-title"]');
                    if (linkElem && linkElem.href) {
                        links.push(linkElem.href.split('?')[0]);
                    }
                    if (links.length >= %d) break;
                }
                return links;
            })()
        `, limit), &links),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp search error: %w", err)
	}

	if len(links) == 0 {
		return nil, parserErrors.ErrEmptyLinks
	}

	return links, nil
}
