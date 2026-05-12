// internal/ozon/parser.go
package ozon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sync"

	//"os"

	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	"crawler/internal/config"
	"crawler/internal/parser"
)

type OzonParser struct {
	cfg     *config.Config
    allocCtx context.Context
    cancel   context.CancelFunc
}

func NewOzonParser(cfg *config.Config) *OzonParser {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("disable-blink-features", "AutomationControlled"),
        chromedp.Flag("disable-automation", true),
        chromedp.UserAgent(cfg.UserAgent),
        chromedp.WindowSize(1920, 1080),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )
    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    return &OzonParser{
        allocCtx: allocCtx,
        cancel:   cancel,
        cfg:     cfg,
    }
}

func (p *OzonParser) Name() string {
	return config.OZON
}

func (p *OzonParser) GetProductByID(ctx context.Context, productID string) (*parser.BaseProduct, error) {
    url := fmt.Sprintf("https://www.ozon.ru/product/%s/", productID)

    ctx, cancel := chromedp.NewContext(p.allocCtx)
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

    ctx, cancel := chromedp.NewContext(p.allocCtx)
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
            sem <- struct{}{}
            defer func() { <-sem }()
            
            id, err := p.getIDFromLink(l)
            if err != nil {
                log.Printf("failed to get product id from %s with link: %s", p.Name(), link)
                return
            }
            product, err := p.GetProductByID(ctx, id)
            if err == nil {
                results <- product
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