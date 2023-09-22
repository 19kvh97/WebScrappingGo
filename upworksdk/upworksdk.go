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
	lw "github.com/19kvh97/webscrappinggo/upworksdk/workers/loginworker"
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
		go func() {
			err := instance.init()
			if err != nil {
				log.Printf("error on get instance upworkSDK : %s", err.Error())
			}
		}()
	}

	return instance
}

func (sdkM *SdkManager) init() error {
	sdkM.wg.Add(1)
	go func() {
		defer sdkM.wg.Done()
		for {
			// log.Printf("Current goroutine count %d", len(sdkM.configs)+1)
			time.Sleep(10 * time.Second)
		}
	}()
	sdkM.wg.Wait()
	log.Println("Run finished")
	return nil
}

func (sdkM *SdkManager) RegisterListener(email string, mode models.RunningMode, listener func(string, models.IParcell)) error {
	log.Printf("worker leng : %d", len(sdkM.Workers))
	configId := ""
	for _, cf := range sdkM.configs {
		if cf.Account.Email == email && cf.Mode == mode {
			configId = cf.Id
			break
		}
	}

	if configId == "" {
		return fmt.Errorf("can't find config with email %s and mode %s", email, mode.GetName())
	}

	for cfId, worker := range sdkM.Workers {
		log.Printf("cfId : %s , target Id : %s", cfId, configId)
		if cfId == configId {
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
			panic(err)
		}
		addIdConfigs = append(addIdConfigs, models.Config{
			Id:      id,
			Mode:    conf.Mode,
			Account: conf.Account,
		})
	}

	if len(sdkM.configs) == 0 {
		sdkM.configs = append(sdkM.configs, addIdConfigs...)
	} else {
		for _, newCf := range addIdConfigs {
			for i, oldCf := range sdkM.configs {
				if newCf.Account.Email == oldCf.Account.Email && newCf.Mode == oldCf.Mode {
					sdkM.configs[i] = newCf
					break
				}
				if i == len(sdkM.configs)-1 {
					sdkM.configs = append(sdkM.configs, newCf)
					break
				}
			}
		}
	}
	log.Printf("config length : %d", len(configs))
	sdkM.wg.Add(len(configs))
	for i := 0; i < updateCount; i++ {
		go func(config models.Config) {
			log.Printf("Running %s", config.Mode.GetName())
			sdkM.newSession(config)
			defer sdkM.wg.Done()
			log.Printf("finished config %s", config.Mode.GetName())
		}(configs[i])
	}

	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		log.Printf("length config = %d, workers = %d", len(sdkM.configs), len(sdkM.Workers))
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
	case models.LOGIN_AS_CREDENTICAL, models.LOGIN_AS_GOOGLE:
		worker = &lw.LoginWorker{
			Worker: wk.Worker{
				Account: config.Account,
			},
			Mode: config.Mode,
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
	defer delete(sdkM.Workers, config.Account.Email)
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
		return "", fmt.Errorf("account is invalid, email must not be empty")
	}
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s_%s_%s", config.Account.Email, config.Mode.GetName(), time.Now().String()))), nil
}
