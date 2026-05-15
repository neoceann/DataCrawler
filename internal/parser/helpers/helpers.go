package helpers

import (
	"context"
	"crawler/internal/parser"
	"log"
	"regexp"
	"sync"
)

func ExtractIDFromURL(url string) string {
	re := regexp.MustCompile(`(\d{8,})`)
	return re.FindString(url)
}

func ParseProducts(ctx context.Context, p parser.Parser, buffer int, links []string) <-chan *parser.BaseProduct {
	results := make(chan *parser.BaseProduct, len(links))

	sem := make(chan struct{}, 2)
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

			product, err := p.ParseProductPage(ctx, l)
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

	return results
}
