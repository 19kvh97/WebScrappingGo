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
	jw "github.com/19kvh97/webscrappinggo/upworksdk/workers/jobworker"
	mw "github.com/19kvh97/webscrappinggo/upworksdk/workers/messageworker"
	"github.com/chromedp/chromedp"
)

type SdkManager struct {
	configs []models.Config
	wg      sync.WaitGroup
	Workers map[string]wk.IWorker
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

func (sdkM *SdkManager) RegisterListener(email string, mode models.RunningMode, listener func(models.IParcell)) error {
	log.Printf("worker leng : %d", len(sdkM.Workers))
	for em, worker := range sdkM.Workers {
		log.Printf("em : %s , email : %s", em, email)
		if em == email {
			log.Printf("runtime type %T", worker)
			if bmWorker, ok := worker.(*jw.JobWorker); ok {
				if bmWorker.GetMode() == mode && bmWorker.Listener == nil {
					bmWorker.Listener = listener
					return nil
				}
			} else if msWorker, ok := worker.(*mw.MessageWorker); ok {
				if msWorker.GetMode() == mode && msWorker.Listener == nil {
					msWorker.Listener = listener
					return nil
				}
			}
		}
	}

	return errors.New("failed to register listener. May listener registered")
}

func (sdkM *SdkManager) Run(configs ...models.Config) error {
	if sdkM.configs == nil {
		sdkM.configs = []models.Config{}
	}
	if sdkM.Workers == nil {
		sdkM.Workers = make(map[string]wk.IWorker)
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

	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		if len(sdkM.configs) == len(sdkM.Workers) {
			for key := range sdkM.Workers {
				log.Printf("email : %s", key)
			}
			return nil
		}
	}

	return errors.New("run failed")
}

func (sdkM *SdkManager) newSession(config models.Config) {
	var worker wk.IWorker
	switch config.Mode {
	case models.SYNC_BEST_MATCH:
		worker = &jw.JobWorker{
			Mode: models.SYNC_BEST_MATCH,
			Worker: wk.Worker{
				Account: config.Account,
			},
		}
	case models.SYNC_RECENTLY:
		worker = &jw.JobWorker{
			Mode: models.SYNC_RECENTLY,
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
	sdkM.Workers[config.Account.Email] = worker

	runner, err := worker.PrepareTask()
	if err != nil {
		panic(err)
	}

	_, cancel := newChromedp(runner)
	defer cancel()
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
