package main

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

func createBrowser() (context.Context, context.CancelFunc) {
	var cancelFuncs []context.CancelFunc
	// browserURL := "ws://0.0.0.0:9222/devtools/browser/27ac1994-a333-4241-90d6-74ed95b29723"
	// ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), browserURL)
	// cancelFuncs = append(cancelFuncs, cancel)

	// define the proxy settings
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", true),
		chromedp.WindowSize(1920, 1080),
	)

	// specify a new context set up for NewContext
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	cancelFuncs = append(cancelFuncs, cancel)

	ctx, cancel = chromedp.NewContext(ctx)
	if err := chromedp.Run(ctx); err != nil {
		log.Fatal("couldn't create browser context: %w", err)
	}
	cancelFuncs = append(cancelFuncs, cancel)

	cancel = func() {
		for _, c := range cancelFuncs {
			c()
		}
	}

	return ctx, cancel
}

func createTab(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := chromedp.NewContext(ctx)

	return ctx, cancel
}
