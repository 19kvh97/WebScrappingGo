package upworksdk

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
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
	Workers map[string]*wk.IWorker
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

func (sdkM *SdkManager) RegisterListener(email string, mode models.RunningMode, listener func(models.IParcell)) {
	for em, worker := range sdkM.Workers {
		if em == email && worker.GetMode() == mode {
			worker.Listener = listener
			break
		}
	}
}

func (sdkM *SdkManager) Run(configs ...models.Config) error {
	if sdkM.configs == nil {
		sdkM.configs = []models.Config{}
	}
	var addIdConfigs []models.Config
	for _, conf := range configs {
		id, err := GenerateUniqueId(conf)
		if err != nil {
			log.Printf("err : %s", err.Error())
			continue
		}
		addIdConfigs = append(addIdConfigs, models.Config{
			Id:      id,
			Mode:    conf.Mode,
			Account: conf.Account,
		})
	}
	sdkM.configs = append(sdkM.configs, addIdConfigs...)
	sdkM.wg.Add(len(configs))
	for i := 0; i < len(configs); i++ {
		go func(config models.Config) {
			log.Printf("Running %s", config.Mode.GetName())
			sdkM.newSession(config)
			defer sdkM.wg.Done()
		}(configs[i])
	}

	return nil
}

func (sdkM *SdkManager) newSession(config models.Config) {
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
	sdkM.Workers[config.Account.Email] = &worker
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

func GenerateUniqueId(config models.Config) (string, error) {
	if config.Account.Email == "" {
		return "", fmt.Errorf("account is invalid")
	}
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s_%s_%s", config.Account.Email, config.Mode.GetName(), time.Now().String()))), nil
}
