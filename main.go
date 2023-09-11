package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	uw "github.com/19kvh97/webscrappinggo/upworksdk"
	"github.com/19kvh97/webscrappinggo/upworksdk/models"

	"github.com/chromedp/chromedp"
)

const (
	googleSignin = "https://accounts.google.com"
)

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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	email := "hung.kv22011997@gmail.com"
	password := "Kimvanhung1997"
	var rawCookie []models.Cookie
	content, err := ioutil.ReadFile("hungkv_cookie.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, &rawCookie)
	if err != nil {
		panic(err)
	}

	validCookie, err := uw.ExtractValidateCookies(rawCookie)
	if err != nil {
		panic(err)
	}

	err = uw.SdkInstance().Run(models.Config{
		Mode: models.LOGIN_AS_CREDENTICAL,
		Account: models.UpworkAccount{
			Email:    email,
			Password: password,
			Cookie:   validCookie,
		},
	})

	if err != nil {
		panic(err)
	}

	// err = uw.SdkInstance().RegisterListener(email, models.LOGIN_AS_CREDENTICAL, DataAvailable)

	// if err != nil {
	// 	panic(err)
	// }

	for {
		time.Sleep(5 * time.Minute)
	}
}

func DataAvailable(email string, parcell models.IParcell) {
	log.Println("received data")
	if job, ok := parcell.(models.Job); ok {
		log.Printf("Job title: %s", job.Title)
	}
}
