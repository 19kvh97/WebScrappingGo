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
	"github.com/xlzd/gotp"
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
		keepSessionXPath := "//*[@id='login']/div/div/div[1]/div[5]/div/div[1]/div/label/span"
		otpXPath := "//*[@id='deviceAuthOtp_otp']"
		nextXPath := "//*[@id='next_continue']"

		runningmode := jw.GetMode()

		tasks := chromedp.Tasks{
			chromedp.Navigate(runningmode.GetLink()),
			chromedp.WaitEnabled("login_username", chromedp.ByID),
			chromedp.SendKeys(emailXPath, jw.Account.Email),
			chromedp.Sleep(2 * time.Second),
			chromedp.Click(continueBtnXPath),
			chromedp.Sleep(2 * time.Second),
			chromedp.SendKeys(passwordXPath, jw.Account.Password),
			chromedp.Sleep(time.Second),
			chromedp.Click(keepSessionXPath),
			chromedp.Sleep(time.Second),
			chromedp.Click(loginXPath),
			chromedp.Sleep(5 * time.Second),
			chromedp.WaitEnabled("deviceAuthOtp_otp", chromedp.ByID),
			chromedp.SendKeys(otpXPath, generateOtp(jw.Account.TwoFA)),
			chromedp.Click(nextXPath),
			chromedp.Sleep(10 * time.Second),
			chromedp.ActionFunc(func(ctx context.Context) error {
				cookies, err := network.GetCookies().Do(ctx)
				if err != nil {
					return err
				}
				for i, cookie := range cookies {
					log.Printf("chrome cookie %d: %+v", i, cookie.Name)
				}
				return nil
			}),
			chromedp.Sleep(20 * time.Second),
		}

		if err := chromedp.Run(ctx, tasks); err != nil {
			fmt.Println(err)
		}

	}, nil
}

func generateOtp(_2faStr string) string {
	totp := gotp.NewDefaultTOTP(_2faStr)
	return totp.Now()
}
