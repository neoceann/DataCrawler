// internal/ozon/parser.go
package ozon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"

	//"os"

	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	"crawler/internal/config"
	"crawler/internal/parser"
	"crawler/internal/repository/db"
)

type OzonParser struct {
	queries *db.Queries
	cfg     *config.Config
}

func NewOzonParser(q *db.Queries, cfg *config.Config) *OzonParser {
	return &OzonParser{
		queries: q,
		cfg:     cfg,
	}
}

func (p *OzonParser) Name() string {
	return config.OZON
}

func (p *OzonParser) GetProductByID(ctx context.Context, productID string) (*parser.BaseProduct, error) {
    url := fmt.Sprintf("https://www.ozon.ru/product/%s/", productID)

    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("disable-blink-features", "AutomationControlled"),
        chromedp.Flag("disable-automation", true),
        chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36"),
        chromedp.WindowSize(1920, 1080),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )
    ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
    defer cancel()

    ctx, cancel = chromedp.NewContext(ctx)
    defer cancel()

    ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
    defer cancel()

    var htmlContent string

    err := chromedp.Run(ctx,
        chromedp.Navigate(url),
        
        chromedp.WaitVisible(`h1`, chromedp.ByQuery),
        
        chromedp.Sleep(3*time.Second),
        
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

func (p *OzonParser) SaveProductToDB(ctx context.Context, product *parser.BaseProduct) error {
	return p.queries.CreateProduct(ctx, *product.ToDbParams())
}

func (p *OzonParser) GetTopProducts(ctx context.Context, s *config.SearchConfig) ([]*parser.BaseProduct, error) {
        searchURL := fmt.Sprintf("https://www.ozon.ru/search/?text=%s&sorting=%s", url.QueryEscape(s.Query), s.SortBy)
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("disable-blink-features", "AutomationControlled"),
        chromedp.Flag("disable-automation", true),
        chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36"),
        chromedp.WindowSize(1920, 1080),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )
    ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
    defer cancel()

    ctx, cancel = chromedp.NewContext(ctx)
    defer cancel()

    ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
    defer cancel()

    log.Printf("Getting top %d products from %s...", s.Limit, p.Name())
    var links []string
    err := chromedp.Run(ctx,
        chromedp.Navigate(searchURL),
        chromedp.WaitVisible(`body`, chromedp.ByQuery),
        chromedp.Sleep(2*time.Second),
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

    productsID := make([]string, len(links))

    for i := range links {
        productsID[i], err = p.getIDFromLink(links[i])
        if err != nil {
            return nil, err
        }
    }

    products := make([]*parser.BaseProduct, len(productsID))

    for i := range products {
        products[i], err = p.GetProductByID(ctx, productsID[i])

        if err != nil {
            return nil, fmt.Errorf("failed to get top products from %s:%s", p.Name(), err.Error())
        }
    }

	return products, nil
}

func (p *OzonParser) getIDFromLink(link string) (string, error) {
    id := ""
    re := regexp.MustCompile(`\b(\d{8,})\b`)
   
    matches := re.FindStringSubmatch(link)

    if len(matches) > 1 {
        id = matches[1]
    }

    if id == "" {
        return id, fmt.Errorf("failed to get product id (%s)", p.Name())
    }

    return id, nil
}