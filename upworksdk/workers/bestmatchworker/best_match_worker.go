package bestmatchwoker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	md "github.com/19kvh97/webscrappinggo/upworksdk/models"
	wk "github.com/19kvh97/webscrappinggo/upworksdk/workers"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type BestMatchWorker struct {
	wk.Worker
}

func (bmw *BestMatchWorker) GetMode() wk.RunningMode {
	return wk.SYNC_BEST_MATCH
}

func (bmw *BestMatchWorker) PrepareTask() func(context.Context) {
	return func(ctx context.Context) {
		cookies := bmw.Account.Cookie

		runningmode := bmw.GetMode()

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
		}
		if err := chromedp.Run(ctx, tasks); err != nil {
			fmt.Println(err)
		}
		var nodes []*cdp.Node
		var job md.Job
		for {
			log.Println("Refresh")
			err := chromedp.Run(ctx,
				chromedp.Navigate(runningmode.GetLink()),
				chromedp.Nodes("section.up-card-section.up-card-list-section.up-card-hover", &nodes, chromedp.ByQueryAll),
				chromedp.Sleep(3*time.Second))
			if err != nil {
				log.Printf("error : %v", err)
			}
			log.Printf("get nodes : %d", len(nodes))
			for i, node := range nodes {
				log.Printf("Node[%d]", i)
				err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
					if err != nil {
						return err
					}
					doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
					if err != nil {
						return err
					}

					doc.Find("section.up-card-section.up-card-list-section.up-card-hover").Each(func(index int, info *goquery.Selection) {
						job.Title = info.Find(".up-n-link").Text()
					})

					return nil
				}))
				if err != nil {
					log.Printf("error : %v", err)
				} else {
					str, _ := json.Marshal(job)
					log.Printf("Job : %s", str)
				}
			}
			time.Sleep(1 * time.Minute)
		}
	}
}
