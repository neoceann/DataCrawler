package yandex

import (
	"context"
	"crawler/internal/browser"
	"crawler/internal/config"
	"crawler/internal/parser"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"

	"sync"

	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
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

func (p *YandexParser) GetTopProducts(ctx context.Context, s *config.SearchConfig) ([]*parser.BaseProduct, error) {
	links, err := p.getProductLinks(ctx, s)
	if err != nil {
		return nil, err
	}

	log.Printf("Getting top %d products from %s...", s.Limit, p.Name())

	sem := make(chan struct{}, 2)

	results := make(chan *parser.BaseProduct, len(links))
	var wg sync.WaitGroup

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

			product, err := p.parseProductPage(ctx, l)
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
	for product := range results {
		products = append(products, product)
		if len(products) >= s.Limit {
			break
		}
	}

	log.Printf("Received products from %s: %d", p.Name(), len(products))
	return products, nil
}

func (p *YandexParser) parseProductPage(ctx context.Context, pageURL string) (*parser.BaseProduct, error) {
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
		return nil, fmt.Errorf("empty data for this product")
	}

	var ozonProduct YandexProduct

	json.Unmarshal([]byte(jsonLD), &ozonProduct)

	ozonProduct.ID = extractIDFromURL(pageURL)

	return ozonProduct.ToBaseProduct(), nil
}

func (p *YandexParser) getProductLinks(ctx context.Context, s *config.SearchConfig) ([]string, error) {
	ctx, cancel := p.browser.NewTab(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("https://market.yandex.ru/search?text=%s&how=%s", url.QueryEscape(s.Query), s.GetSortParamForMarket(p.Name(), s.SortBy))

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
		return nil, fmt.Errorf("empty links")
	}

	return links, err
}

func (p *YandexParser) GetProductByID(ctx context.Context, productID string) (*parser.BaseProduct, error) {
	return nil, nil
}

func extractIDFromURL(url string) string {
	re := regexp.MustCompile(`(\d{8,})`)
	return re.FindString(url)
}
