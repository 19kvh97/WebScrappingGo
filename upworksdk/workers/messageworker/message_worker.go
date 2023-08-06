package messageworker

import (
	"context"
	"fmt"
	"log"
	"time"

	md "github.com/19kvh97/webscrappinggo/upworksdk/models"
	wk "github.com/19kvh97/webscrappinggo/upworksdk/workers"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type MessageWorker struct {
	wk.Worker
}

func (mw *MessageWorker) GetMode() md.RunningMode {
	return md.SYNC_MESSAGE
}

func (mw *MessageWorker) PrepareTask() func(context.Context) {
	return func(ctx context.Context) {
		cookies := mw.Account.Cookie

		runningmode := mw.GetMode()
		log.Printf("cookies length in PrepareTask %d", len(cookies))
		tasks := chromedp.Tasks{
			chromedp.ActionFunc(func(ctx context.Context) error {
				// create cookie expiration
				expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
				// add cookies to chrome
				for _, cookie := range cookies {
					err := network.SetCookie(cookie.Name, cookie.Value).
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
