package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	googleSignin = "https://accounts.google.com"
)

func newChromedp() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", true),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Login google
	googleTask(ctx)

	return ctx, cancel
}

func googleTask(ctx context.Context) {
	email := "//*[@id='identifierId']"
	password := "//*[@id='password']/div[1]/div/div[1]/input"
	buttonEmailNext := "//*[@id='identifierNext']/div/button"
	buttonPasswordNext := "//*[@id='passwordNext']/div/button/span"

	task := chromedp.Tasks{
		chromedp.Navigate(googleSignin),
		chromedp.SendKeys(email, "email"),
		chromedp.Sleep(2 * time.Second),

		chromedp.Click(buttonEmailNext),
		chromedp.Sleep(2 * time.Second),

		chromedp.SendKeys(password, "pw"),
		chromedp.Sleep(2 * time.Second),

		chromedp.Click(buttonPasswordNext),
		chromedp.Sleep(2 * time.Second),
	}

	if err := chromedp.Run(ctx, task); err != nil {
		fmt.Println(err)
	}
}

func main() {
	_, cancel := newChromedp()
	defer cancel()
}
