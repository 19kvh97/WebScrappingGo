package recentlyworker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	wk "github.com/19kvh97/webscrappinggo/upworksdk/workers"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type RecentlyWorker struct {
	wk.Worker
}

func (rw *RecentlyWorker) GetMode() wk.RunningMode {
	return wk.SYNC_RECENTLY
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
