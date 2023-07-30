package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const (
	googleSignin = "https://accounts.google.com"
)

func newChromedp() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("remote-debugging-port", "9222"),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Login google
	// googleTask(ctx)
	fbTasks(ctx)

	return ctx, cancel
}

func googleTask(ctx context.Context) {
	email := "//*[@id='identifierId']"
	password := "//*[@id='password']/div[1]/div/div[1]/input"
	buttonEmailNext := "//*[@id='identifierNext']/div/button"
	buttonPasswordNext := "//*[@id='passwordNext']/div/button/span"

	task := chromedp.Tasks{
		chromedp.Navigate(googleSignin),
		chromedp.SendKeys(email, "19kvh97"),
		chromedp.Sleep(2 * time.Second),

		chromedp.Click(buttonEmailNext),
		chromedp.Sleep(2 * time.Second),

		chromedp.SendKeys(password, "kimvanhung1997"),
		chromedp.Sleep(2 * time.Second),

		chromedp.Click(buttonPasswordNext),
		chromedp.Sleep(2 * time.Second),
	}

	if err := chromedp.Run(ctx, task); err != nil {
		fmt.Println(err)
	}
}

func fbTasks(ctx context.Context) {
	cookies := []string{"xs", "2%3AhnJ5IsiLRCyVzw%3A2%3A1681221329%3A-1%3A6371%3A%3AAcVHopKQidud0vuFk1KSsU30Gmxd-zyguYporL1n8MTi", "c_user", "100004036156703"}
	tasks := chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// create cookie expiration
			expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			// add cookies to chrome
			for i := 0; i < len(cookies); i += 2 {
				err := network.SetCookie(cookies[i], cookies[i+1]).
					WithExpires(&expr).
					WithDomain("www.facebook.com").
					WithHTTPOnly(false).
					Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
		// navigate to site
		chromedp.Navigate("https://www.facebook.com"),
		chromedp.Sleep(30 * time.Second),
	}
	if err := chromedp.Run(ctx, tasks); err != nil {
		fmt.Println(err)
	}
}

func upworkTask(ctx context.Context) {

}

func main() {
	_, cancel := newChromedp()
	defer cancel()
}
