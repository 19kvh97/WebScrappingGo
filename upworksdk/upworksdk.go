package upworksdk

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
	wk "github.com/19kvh97/webscrappinggo/upworksdk/workers"
	bmw "github.com/19kvh97/webscrappinggo/upworksdk/workers/bestmatchworker"
	mw "github.com/19kvh97/webscrappinggo/upworksdk/workers/messageworker"
	rw "github.com/19kvh97/webscrappinggo/upworksdk/workers/recentlyworker"
	"github.com/chromedp/chromedp"
)

type Config struct {
	Mode    wk.RunningMode
	Account models.UpworkAccount
}

type SdkManager struct {
	configs []Config
}

var instance *SdkManager

func SdkInstance() *SdkManager {
	if instance == nil {
		instance = &SdkManager{}
	}
	return instance
}

func (sdkM *SdkManager) Run(configs ...Config) error {
	sdkM.configs = configs
	wg := sync.WaitGroup{}
	numRoutine := len(configs)
	wg.Add(numRoutine)
	for i := 0; i < numRoutine; i++ {
		go func(config Config) {
			log.Printf("Running %s", config.Mode.GetName())
			err := sdkM.newSession(config)
			if err != nil {
				log.Printf("Error from routine %s: %v", config.Mode.GetName(), err)
			}
			defer wg.Done()
		}(sdkM.configs[i])
	}
	wg.Wait()
	log.Println("Run finished")
	return nil
}

func (sdkM *SdkManager) newSession(config Config) error {
	var worker wk.IWorker
	switch config.Mode {
	case wk.SYNC_BEST_MATCH:
		worker = &bmw.BestMatchWorker{
			Worker: wk.Worker{
				Account: config.Account,
			},
		}
	case wk.SYNC_RECENTLY:
		worker = &rw.RecentlyWorker{
			Worker: wk.Worker{
				Account: config.Account,
			},
		}
	case wk.SYNC_MESSAGE:
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
		if strings.Contains(cookie.Domain, "upwork") {
			validCookies = append(validCookies, cookie)
		}
	}
	if len(validCookies) == 0 {
		return nil, errors.New("have no cookie valid")
	}
	log.Printf("cookie leng %d", len(validCookies))
	return validCookies, nil
}
