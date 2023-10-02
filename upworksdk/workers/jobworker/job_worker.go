package jobworker

import (
	"context"
	"errors"
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

type JobWorker struct {
	wk.Worker
	Mode     md.RunningMode
	Interval int
}

func (jw *JobWorker) GetMode() md.RunningMode {
	return jw.Mode
}

func (jw *JobWorker) PrepareTask() (func(context.Context), error) {
	if jw.Mode == md.UNKNOWN {
		return nil, errors.New("invalid RunningMode")
	}
	if jw.Interval <= 0 {
		return nil, fmt.Errorf("invalid interval %d", jw.Interval)
	}

	return func(ctx context.Context) {
		cookies := jw.Account.Cookie

		runningmode := jw.GetMode()

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
			log.Printf("err : %s", err.Error())
			return
		}

		// var check string
		// if err := chromedp.Run(ctx,
		// 	chromedp.Sleep(5*time.Second),
		// 	chromedp.EvaluateAsDevTools("document.getElementById('fwh-sidebar-profile')", &check)); err != nil {
		// 	log.Printf("err : %v", err)
		// 	return
		// }

		// if len(check) == 0 {
		// 	log.Println("Can't find profile section")
		// 	return
		// }

		var nodes []*cdp.Node
		var job md.Job
		for {
			err := chromedp.Run(ctx,
				chromedp.Nodes("section.up-card-section.up-card-list-section.up-card-hover", &nodes, chromedp.ByQueryAll, chromedp.AtLeast(0)),
				chromedp.Sleep(3*time.Second))

			if err != nil {
				log.Printf("error : %v", err)
			}
			if len(nodes) == 0 {
				log.Println("err : Can't find any node")
				return
			}

			for _, node := range nodes {
				err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
					if err != nil {
						return err
					}
					doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
					if err != nil {
						return err
					}

					doc.Find("section.up-card-section.up-card-list-section.up-card-hover").Each(func(index int, info *goquery.Selection) {
						job.ImportData(info)
						jw.SendResult(job)
					})
					return nil
				}))
				if err != nil {
					log.Printf("error : %v", err)
				}
			}

			time.Sleep(time.Duration(jw.Interval * int(time.Millisecond)))
			log.Println("Refresh")
			err = chromedp.Run(ctx,
				chromedp.Navigate(runningmode.GetLink()))
			if err != nil {
				log.Printf("error : %v", err)
			}
		}
	}, nil
}
