package browser

import (
	"context"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	allocCtx context.Context
	cancel   context.CancelFunc
}

func New(cfg *BrowserConfig) (*Browser, error) {
	opts := DefaultOptions()

	opts = append(opts, chromedp.UserAgent(cfg.UserAgent))

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	return &Browser{
		allocCtx: allocCtx,
		cancel:   cancel,
	}, nil
}

func (b *Browser) NewTab(ctx context.Context) (context.Context, context.CancelFunc) {
	return chromedp.NewContext(b.allocCtx)
}

func (b *Browser) Close() error {
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}
