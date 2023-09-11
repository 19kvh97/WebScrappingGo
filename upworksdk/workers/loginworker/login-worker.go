package loginworker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	md "github.com/19kvh97/webscrappinggo/upworksdk/models"
	wk "github.com/19kvh97/webscrappinggo/upworksdk/workers"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type LoginWorker struct {
	wk.Worker
	Mode md.RunningMode
}

func (jw *LoginWorker) GetMode() md.RunningMode {
	return jw.Mode
}

func (jw *LoginWorker) PrepareTask() (func(context.Context), error) {
	if jw.Mode != md.LOGIN_AS_CREDENTICAL && jw.Mode != md.LOGIN_AS_GOOGLE {
		return nil, errors.New("invalid RunningMode")
	}
	return func(ctx context.Context) {
		emailXPath := "//*[@id='login_username']"
		continueBtnXPath := "//*[@id='login_password_continue']"
		passwordXPath := "//*[@id='login_password']"
		loginXPath := "//*[@id='login_control_continue']"

		runningmode := jw.GetMode()

		// if err := chromedp.Run(ctx, chromedp.Navigate(runningmode.GetLink())); err != nil {
		// 	fmt.Println(err)
		// }

		// isElementEnabled := func(ctx context.Context) (bool, error) {
		// 	var isDisabled bool
		// 	var nodes []*cdp.Node
		// 	err := chromedp.Run(ctx, chromedp.Nodes(`#login_username:[disabled]`, &nodes))
		// 	if err != nil {
		// 		return false, err
		// 	}
		// 	if len(nodes) > 0 {
		// 		isDisabled = true
		// 	}
		// 	return !isDisabled, nil
		// }

		// for i := 0; i < 10; i++ {
		// 	enabled, err := isElementEnabled(ctx)
		// 	if err != nil {
		// 		log.Printf("error : %s", err.Error())
		// 	}
		// 	if enabled {
		// 		break
		// 	}
		// 	time.Sleep(2 * time.Second)

		// 	if i == 9 {
		// 		return
		// 	}
		// }

		tasks := chromedp.Tasks{
			chromedp.Navigate(runningmode.GetLink()),
			chromedp.WaitEnabled("login_username", chromedp.ByID),
			chromedp.SendKeys(emailXPath, jw.Account.Email),
			chromedp.Sleep(2 * time.Second),
			chromedp.Click(continueBtnXPath),
			chromedp.Sleep(2 * time.Second),
			chromedp.SendKeys(passwordXPath, jw.Account.Password),
			chromedp.Sleep(2 * time.Second),
			chromedp.Click(loginXPath),
			chromedp.Sleep(2 * time.Second),
		}

		if err := chromedp.Run(ctx, tasks); err != nil {
			fmt.Println(err)
		}

		log.Printf("%v", network.GetCookies())

	}, nil
}
