package upworksdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type RunningMode int

const (
	SYNC_BEST_MATCH RunningMode = iota
	SYNC_RECENTLY
	SYNC_MESSAGE
)

func (rm *RunningMode) GetLink() string {
	switch *rm {
	case SYNC_BEST_MATCH:
		return "https://www.upwork.com/nx/find-work/best-matches"
	case SYNC_RECENTLY:
		return "https://www.upwork.com/nx/find-work/most-recent"
	case SYNC_MESSAGE:
		return "https://www.upwork.com/ab/messages"
	default:
		return ""
	}
}

type Worker interface {
	PrepareTask() func(context.Context)
	GetMode() RunningMode
}

type BestMatchWorker struct {
	Worker
}

func (bmw *BestMatchWorker) GetMode() RunningMode {
	return SYNC_BEST_MATCH
}

func (bmw *BestMatchWorker) PrepareTask() func(context.Context) {
	return func(ctx context.Context) {
		cookieFile, err := ioutil.ReadFile("cookie.json")
		if err != nil {
			log.Printf("error : %v", err)
			return
		}
		var cookieMap map[string]string
		err = json.Unmarshal(cookieFile, &cookieMap)
		if err != nil {
			log.Printf("error : %v", err)
			return
		}
		var cookies []string
		for key, value := range cookieMap {
			cookies = append(cookies, key)
			cookies = append(cookies, value)
		}

		runningmode := bmw.GetMode()

		tasks := chromedp.Tasks{
			chromedp.ActionFunc(func(ctx context.Context) error {
				// create cookie expiration
				expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
				// add cookies to chrome
				for i := 0; i < len(cookies); i += 2 {
					err := network.SetCookie(cookies[i], cookies[i+1]).
						WithExpires(&expr).
						WithDomain("www.upwork.com").
						WithHTTPOnly(false).
						Do(ctx)
					if err != nil {
						return err
					}
				}
				return nil
			}),

			// navigate to site
			chromedp.Navigate(runningmode.GetLink()),
			chromedp.Sleep(30 * time.Second),
		}
		if err := chromedp.Run(ctx, tasks); err != nil {
			fmt.Println(err)
		}
	}
}

type RecentlyWorker struct {
	Worker
}

func (rw *RecentlyWorker) GetMode() RunningMode {
	return SYNC_RECENTLY
}

func (rw *RecentlyWorker) PrepareTask() func(context.Context) {
	return func(ctx context.Context) {
		cookieFile, err := ioutil.ReadFile("cookie.json")
		if err != nil {
			log.Printf("error : %v", err)
			return
		}
		var cookieMap map[string]string
		err = json.Unmarshal(cookieFile, &cookieMap)
		if err != nil {
			log.Printf("error : %v", err)
			return
		}
		var cookies []string
		for key, value := range cookieMap {
			cookies = append(cookies, key)
			cookies = append(cookies, value)
		}

		runningmode := rw.GetMode()

		tasks := chromedp.Tasks{
			chromedp.ActionFunc(func(ctx context.Context) error {
				// create cookie expiration
				expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
				// add cookies to chrome
				for i := 0; i < len(cookies); i += 2 {
					err := network.SetCookie(cookies[i], cookies[i+1]).
						WithExpires(&expr).
						WithDomain("www.upwork.com").
						WithHTTPOnly(false).
						Do(ctx)
					if err != nil {
						return err
					}
				}
				return nil
			}),

			// navigate to site
			chromedp.Navigate(runningmode.GetLink()),
			chromedp.Sleep(30 * time.Second),
		}
		if err := chromedp.Run(ctx, tasks); err != nil {
			fmt.Println(err)
		}
	}
}

type Config struct {
	Mode RunningMode
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
		var worker Worker
		switch config.Mode {
		case SYNC_BEST_MATCH:
			worker = &BestMatchWorker{}
		case SYNC_RECENTLY:
			worker = &RecentlyWorker{}
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
