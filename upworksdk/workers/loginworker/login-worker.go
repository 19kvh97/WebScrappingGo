package loginworker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	md "github.com/19kvh97/webscrappinggo/upworksdk/models"
	wk "github.com/19kvh97/webscrappinggo/upworksdk/workers"
	"github.com/chromedp/cdproto/storage"
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
			chromedp.Sleep(2 * time.Second),
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
			chromedp.Sleep(10 * time.Second),
		}

		// Create a channel to listen for cookies.

		if err := chromedp.Run(ctx, tasks); err != nil {
			fmt.Println(err)
		}

		// Check if the input field exists.
		log.Println("check otp node")
		// var nodes []*cdp.Node
		var elementContent string
		if err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools("document.getElementById('deviceAuthOtp_otp')", &elementContent)); err != nil {
			panic(err)
		}

		log.Printf("content %s", elementContent)

		if len(elementContent) > 0 {
			log.Println("verified 2fa")
			verifiesTask := chromedp.Tasks{
				chromedp.SendKeys(otpXPath, generateOtp(jw.Account.TwoFA)),
				chromedp.Click(nextXPath),
				chromedp.Sleep(10 * time.Second),
			}

			if err := chromedp.Run(ctx, verifiesTask); err != nil {
				panic(err)
			}
		}

		log.Println("run custom action")

		if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("check cookie")
			cookies, err := storage.GetCookies().Do(ctx)

			log.Printf("cookies length : %d", len(cookies))
			var c []string
			for _, v := range cookies {
				aCookie := v.Name + " - " + v.Domain
				c = append(c, aCookie)
			}

			stringSlices := strings.Join(c[:], ",\n")
			fmt.Printf("%v", stringSlices)

			if err != nil {
				return err
			}
			return nil
		})); err != nil {
			panic(err)
		}

	}, nil
}

func generateOtp(_2faStr string) string {
	totp := gotp.NewDefaultTOTP(_2faStr)
	return totp.Now()
}
