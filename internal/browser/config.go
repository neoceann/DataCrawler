package browser

import (
	"github.com/chromedp/chromedp"
)

func DefaultOptions() []chromedp.ExecAllocatorOption {
	return append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-automation", true),
		chromedp.Flag("no-sandbox", true),
		//chromedp.Flag("disable-dev-shm-usage", true),
		//chromedp.Flag("disable-gpu", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.Flag("disable-images", true),
		chromedp.Flag("disable-javascript", false),
		chromedp.Flag("disable-css", true),
		chromedp.Flag("disable-fonts", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-default-apps", true),
	)
}

type BrowserConfig struct {
	UserAgent string
}

func DefaultConfig() *BrowserConfig {
	return &BrowserConfig{}
}
