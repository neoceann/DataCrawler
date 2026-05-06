// internal/ozon/parser.go
package ozon

import (
	"context"
	"encoding/json"
	"fmt"

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
        chromedp.Flag("disable-blink-features", "AutomationControlled"), // скрываем автоматизацию
        chromedp.Flag("disable-automation", true),                       // убираем флаги автоматизации
        chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36"),
        // chromedp.WindowSize(1920, 1080),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )
    allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
    defer cancel()

    ctx1, cancel := chromedp.NewContext(allocCtx)
    defer cancel()

    ctx2, cancel := context.WithTimeout(ctx1, 45*time.Second)
    defer cancel()

    var htmlContent string

    err := chromedp.Run(ctx2,
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

// fullHTML, _ := doc.Html()
// err = os.WriteFile("ozon_page.html", []byte(fullHTML), 0644)
// if err != nil {
//     log.Fatal(err)
// }
// log.Print("HTML сохранён в ozon_page.html")


    var jsonLD string
    doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
        jsonLD = s.Text()
    })

    if jsonLD == "" {
        return nil, fmt.Errorf("empty data for this product")
    }

    var ozonProduct OzonProduct

    json.Unmarshal([]byte(jsonLD), &ozonProduct)

    return ozonProduct.ToBaseProduct()
}

func (p *OzonParser) SaveProductToDB(ctx context.Context, product *parser.BaseProduct) error {
	return p.queries.CreateProduct(ctx, *product.ToDbParams())
}

func (p *OzonParser) GetTopProducts(ctx context.Context, s *config.SearchConfig) ([]*parser.BaseProduct, error) {
	return nil, nil
}