package avito

import (
	"context"
	"crawler/internal/config"
	"crawler/internal/parser"
	"fmt"
	"log"
	//"os"
	"regexp"
	"sync"

	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

type AvitoParser struct {
	cfg     *config.Config
    allocCtx context.Context
    cancel   context.CancelFunc
}

func NewAvitoParser(cfg *config.Config) *AvitoParser {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("disable-blink-features", "AutomationControlled"),
        chromedp.Flag("disable-automation", true),
        chromedp.UserAgent(cfg.UserAgent),
        chromedp.WindowSize(1920, 1080),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-dev-shm-usage", true),
    )
    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    return &AvitoParser{
        allocCtx: allocCtx,
        cancel:   cancel,
        cfg:     cfg,
    }
}

func (p *AvitoParser) Name() string {
	return config.AVITO
}

func (p *AvitoParser) GetTopProducts(ctx context.Context, s *config.SearchConfig) ([]*parser.BaseProduct, error) {
    links, err := p.getProductLinks(ctx, s.Query, s.Limit)
    if err != nil {
        return nil, err
    }

    log.Printf("Getting top %d products from %s (found %d links)...", s.Limit, p.Name(), len(links))

    sem := make(chan struct{}, 2)
    
    results := make(chan *parser.BaseProduct, len(links))
    var wg sync.WaitGroup

    for _, link := range links {
        wg.Add(1)
        go func(l string) {
            defer wg.Done()
            
            sem <- struct{}{}
            defer func() { <-sem }()
            
            product, err := p.parseProductPage(ctx, l)
            if err != nil {
                log.Printf("Warning: failed to parse %s: %v", l, err)
                return
            }
            results <- product
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

func (p *AvitoParser) GetProductByID(ctx context.Context, productID string) (*parser.BaseProduct, error) {
	return nil, nil
}

func (p *AvitoParser) parseProductPage(ctx context.Context, pageURL string) (*parser.BaseProduct, error) {
	ctx, cancel := chromedp.NewContext(p.allocCtx)
    defer cancel()

    ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    var htmlContent string
    err := chromedp.Run(ctx,
        chromedp.Navigate(pageURL),
        chromedp.WaitVisible(`h1`, chromedp.ByQuery),
        chromedp.Sleep(3*time.Second),
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
//     log.Fatal(err)
// }
// log.Print("HTML сохранён в _page.html")

    product := &AvitoProduct{
    }

    product.ID = extractIDFromURL(pageURL)

    product.Name = strings.TrimSpace(doc.Find(`h1[data-marker="item-view/title"]`).Text())
    if product.Name == "" {
        product.Name = strings.TrimSpace(doc.Find(`h1`).First().Text())
    }

	product.Price = strings.Fields(doc.Find(`[data-marker="item-view/item-price"]`).First().Text())[0]
	product.Supplier = doc.Find(`[data-marker="seller-info/name"]`).First().Text()
	product.SupplierRating, _ = doc.Find(`[data-marker="sellerRate"] meta[itemprop="ratingValue"]`).Attr("content")

	log.Print(product.SupplierRating)

    return product.ToBaseProduct(), nil
}

func (p *AvitoParser) getProductLinks(ctx context.Context, query string, limit int) ([]string, error) {
    ctx, cancel := chromedp.NewContext(p.allocCtx)
    defer cancel()

    ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
    defer cancel()

    searchURL := fmt.Sprintf("https://www.avito.ru/kazan?q=%s казань", query)
    
    var links []string
    err := chromedp.Run(ctx,
        chromedp.Navigate(searchURL),
        chromedp.WaitVisible(`[data-marker="item"]`, chromedp.ByQuery),
        chromedp.Sleep(3*time.Second),
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
        return nil, fmt.Errorf("no product links found for query: %s", query)
    }

    return links, nil
}

func extractIDFromURL(url string) string {
    re := regexp.MustCompile(`(\d{8,})`)
    return re.FindString(url)
}