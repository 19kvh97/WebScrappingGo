package bestmatchwoker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

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
		}
		if err := chromedp.Run(ctx, tasks); err != nil {
			fmt.Println(err)
		}
		var nodes []*cdp.Node
		var jobTitle string
		for {
			log.Println("Refresh")
			err = chromedp.Run(ctx,
				chromedp.Navigate(runningmode.GetLink()),
				chromedp.Nodes(".up-card-section", &nodes, chromedp.ByQueryAll),
				chromedp.Sleep(3*time.Second))
			if err != nil {
				log.Printf("error : %v", err)
			}
			log.Printf("get nodes : %d", len(nodes))
			for i, _ := range nodes {
				log.Printf("Node[%d]", i)
				err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					node, err := dom.GetDocument().Do(ctx)
					if err != nil {
						return err
					}
					res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
					if err != nil {
						return err
					}
					doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
					if err != nil {
						return err
					}

					doc.Find("section.up-card-section.up-card-list-section.up-card-hover").Each(func(index int, info *goquery.Selection) {
						text := info.Find("a.up-n-link").Text()
						fmt.Println(text)
					})

					return nil
				}))
				if err != nil {
					log.Printf("error : %v", err)
				} else {
					log.Printf("Jobtitle : %v", jobTitle)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}
}
