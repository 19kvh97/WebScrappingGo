package upworksdk

import (
	"context"
	"log"

	wk "github.com/19kvh97/webscrappinggo/upworksdk/workers"
	bmw "github.com/19kvh97/webscrappinggo/upworksdk/workers/bestmatchworker"
	rw "github.com/19kvh97/webscrappinggo/upworksdk/workers/recentlyworker"
	"github.com/chromedp/chromedp"
)

type Config struct {
	Mode wk.RunningMode
}

type SdkManager struct {
}

var instance *SdkManager

func SdkInstance() *SdkManager {
	if instance == nil {
		instance = &SdkManager{}
	}
	return instance
}

func (sdkM *SdkManager) NewSession(config Config) error {
	go func() {
		var worker wk.Worker
		switch config.Mode {
		case wk.SYNC_BEST_MATCH:
			worker = &bmw.BestMatchWorker{}
		case wk.SYNC_RECENTLY:
			worker = &rw.RecentlyWorker{}
		default:
			break
		}
		_, cancel := newChromedp(worker.PrepareTask())
		defer cancel()
	}()
	return nil
}

func newChromedp(worker func(context.Context)) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("remote-debugging-port", "9222"),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Login google
	// googleTask(ctx)
	// fbTasks(ctx)
	worker(ctx)

	return ctx, cancel
}
