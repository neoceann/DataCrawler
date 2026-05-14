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

func (b *Browser) NewTab(parentCtx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := chromedp.NewContext(b.allocCtx)

	go func() {
		select {
		case <-parentCtx.Done():
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}

func (b *Browser) Close() {
	if b.cancel != nil {
		b.cancel()
	}
}
