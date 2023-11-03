package factory

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

type ChromeTask func(context.Context)

func NewChromeInstance(taskChan <-chan ChromeTask) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	for {
		task := <-taskChan
		log.Println("Receive new task")
		task(ctx)
	}
}
