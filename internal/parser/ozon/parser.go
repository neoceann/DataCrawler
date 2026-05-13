package ozon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	"crawler/internal/browser"
	"crawler/internal/config"
	"crawler/internal/parser"
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

func (p *OzonParser) GetProductByID(ctx context.Context, productID string) (*parser.BaseProduct, error) {
	url := fmt.Sprintf("https://www.ozon.ru/product/%s/", productID)

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
		return nil, fmt.Errorf("empty data for this product")
	}

	var ozonProduct OzonProduct

	json.Unmarshal([]byte(jsonLD), &ozonProduct)

	return ozonProduct.ToBaseProduct(), nil
}

func (p *OzonParser) GetTopProducts(ctx context.Context, s *config.SearchConfig) ([]*parser.BaseProduct, error) {
	searchURL := fmt.Sprintf("https://www.ozon.ru/search/?text=%s&sorting=%s", url.QueryEscape(s.Query), s.GetSortParamForMarket(p.Name(), s.SortBy))

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

	var wg sync.WaitGroup
	results := make(chan *parser.BaseProduct, len(links))

	sem := make(chan struct{}, 3)

	for _, link := range links {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()

           select {
            case sem <- struct{}{}:
                defer func() { <-sem }()
            case <-ctx.Done():
                return
            }

			select {
            case <-ctx.Done():
                return
            default:
            }

			id := extractIDFromURL(l)

			product, err := p.GetProductByID(ctx, id)
			if err != nil {
				log.Printf("Warning: failed to parse %s: %v", l, err)
				return
			}
			
			select {
            case results <- product:
            case <-ctx.Done():
                return
            }

		}(link)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var products []*parser.BaseProduct
	for p := range results {
		products = append(products, p)
	}

	log.Printf("Received products from %s: %d", p.Name(), len(products))

	return products, nil
}

func extractIDFromURL(url string) string {
	re := regexp.MustCompile(`(\d{8,})`)
	return re.FindString(url)
}
