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

		tasks := chromedp.Tasks{
			// navigate to site
			chromedp.Navigate(runningmode.GetLink()),
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
