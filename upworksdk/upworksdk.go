package upworksdk

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
	wk "github.com/19kvh97/webscrappinggo/upworksdk/workers"
	bmw "github.com/19kvh97/webscrappinggo/upworksdk/workers/bestmatchworker"
	mw "github.com/19kvh97/webscrappinggo/upworksdk/workers/messageworker"
	rw "github.com/19kvh97/webscrappinggo/upworksdk/workers/recentlyworker"
	"github.com/chromedp/chromedp"
)

type SdkManager struct {
	configs []models.Config
	wg      sync.WaitGroup
}

var instance *SdkManager

func SdkInstance() *SdkManager {
	if instance == nil {
		instance = &SdkManager{}
	}
	go func() {
		err := instance.init()
		if err != nil {
			log.Printf("error on get instance upworkSDK : %s", err.Error())
		}
	}()
	return instance
}

func (sdkM *SdkManager) init() error {
	sdkM.wg = sync.WaitGroup{}
	sdkM.wg.Add(1)
	go func() {
		for {
			// log.Printf("Current goroutine count %d", len(sdkM.configs)+1)
			time.Sleep(10 * time.Second)
		}
	}()
	sdkM.wg.Wait()
	log.Println("Run finished")
	return nil
}

func (sdkM *SdkManager) RegisterListener() {

}

func (sdkM *SdkManager) Run(configs ...models.Config) error {
	if sdkM.configs == nil {
		sdkM.configs = []models.Config{}
	}
	sdkM.configs = append(sdkM.configs, configs...)
	sdkM.wg.Add(len(configs))
	for i := 0; i < len(configs); i++ {
		go func(config models.Config) {
			log.Printf("Running %s", config.Mode.GetName())
			err := sdkM.newSession(config)
			if err != nil {
				log.Printf("Error from routine %s: %v", config.Mode.GetName(), err)
			}
			defer sdkM.wg.Done()
		}(configs[i])
	}

	return nil
}

func (sdkM *SdkManager) newSession(config models.Config) error {
	var worker wk.IWorker
	switch config.Mode {
	case models.SYNC_BEST_MATCH:
		worker = &bmw.BestMatchWorker{
			Worker: wk.Worker{
				Account: config.Account,
			},
		}
	case models.SYNC_RECENTLY:
		worker = &rw.RecentlyWorker{
			Worker: wk.Worker{
				Account: config.Account,
			},
		}
	case models.SYNC_MESSAGE:
		worker = &mw.MessageWorker{
			Worker: wk.Worker{
				Account: config.Account,
			},
		}
	default:
		break
	}
	_, cancel := newChromedp(worker.PrepareTask())
	defer cancel()
	return nil
}

func newChromedp(worker func(context.Context)) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Login google
	// googleTask(ctx)
	// fbTasks(ctx)
	worker(ctx)

	return ctx, cancel
}

func ExtractValidateCookies(cookies []models.Cookie) ([]models.Cookie, error) {
	var validCookies []models.Cookie
	for _, cookie := range cookies {
		if strings.Contains(cookie.Domain, "upwork") && cookie.Secure {
			validCookies = append(validCookies, cookie)
		}
	}
	if len(validCookies) == 0 {
		return nil, errors.New("have no cookie valid")
	}
	log.Printf("cookie leng %d", len(validCookies))
	return validCookies, nil
}
