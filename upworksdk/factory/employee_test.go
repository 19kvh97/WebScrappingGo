package factory

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk"
	"github.com/19kvh97/webscrappinggo/upworksdk/models"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func initValidConfig(t *testing.T, interval int) models.Config {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	email := "hung.kv22011997@gmail.com"
	password := "Kimvanhung@1997"

	// log.Printf("test opt %s", totp.Now())

	var rawCookie []models.Cookie
	// content, err := ioutil.ReadFile("../nothing_cookie.json")
	content, err := ioutil.ReadFile("../../valid_cookie1.json")

	log.Printf("%v", err)
	require.Nil(t, err)

	err = json.Unmarshal(content, &rawCookie)
	require.Nil(t, err)

	validCookie, err := upworksdk.ExtractValidateCookies(rawCookie)
	require.Nil(t, err)
	return models.Config{
		Mode: models.SYNC_BEST_MATCH,
		Account: models.UpworkAccount{
			Email:    email,
			Password: password,
			Cookie:   validCookie,
		},
		Interval: interval, // 30 second
	}
}

func initInvalidConfig(t *testing.T, interval int) models.Config {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	email := "hung.kv22011997@gmail.com"
	password := "Kimvanhung@1997"

	// log.Printf("test opt %s", totp.Now())

	var rawCookie []models.Cookie
	// content, err := ioutil.ReadFile("../nothing_cookie.json")
	content, err := ioutil.ReadFile("../../expired_cookie.json")

	log.Printf("%v", err)
	require.Nil(t, err)

	err = json.Unmarshal(content, &rawCookie)
	require.Nil(t, err)

	validCookie, err := upworksdk.ExtractValidateCookies(rawCookie)
	require.Nil(t, err)
	return models.Config{
		Mode: models.SYNC_BEST_MATCH,
		Account: models.UpworkAccount{
			Email:    email,
			Password: password,
			Cookie:   validCookie,
		},
		Interval: interval, // 30 second
	}
}

func TestSetUpChromeInstance(t *testing.T) {
	validConfig := initValidConfig(t, 30)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	defer cancel()

	taskChan := make(chan Task)
	errChan := make(chan ErrorMessage)
	resultChan := make(chan models.IParcell)

	go setUpChromeInstance(ctx, taskChan, errChan, resultChan)
	time.Sleep(2 * time.Second)
	taskChan <- func(ctx context.Context) (*Result, error) {
		cookies := validConfig.Account.Cookie

		runningmode := validConfig.Mode

		log.Printf("cookies length in PrepareTask %d", len(cookies))
		var nodes []*cdp.Node
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
			chromedp.Nodes("section.up-card-section.up-card-list-section.up-card-hover", &nodes, chromedp.ByQueryAll, chromedp.AtLeast(0)),
		}
		if err := chromedp.Run(ctx, tasks); err != nil {
			return nil, err
		}
		if len(nodes) == 0 {
			return nil, errors.New("err : Can't find any node")
		}
		return nil, nil
	}

	select {
	case msg := <-errChan:
		log.Printf("errorChan %s", msg.Message)
	case <-resultChan:
		log.Println("resultChan")
	}
}

func TestEmployeeStartWorking(t *testing.T) {
	numGoru := runtime.NumGoroutine()

	validConfig := initValidConfig(t, 30000)

	empl := Employee{
		StopWork:   make(chan struct{}),
		ResultChan: make(chan models.IParcell),
		ErrorChan:  make(chan ErrorMessage),
	}

	go empl.StartWorking(validConfig)
	for {
		log.Printf("state : %v", empl.State)
		if empl.State == CREATE_JOB_STATE {
			break
		}
		time.Sleep(time.Second)
	}
	require.Equal(t, CREATE_JOB_STATE, empl.State)
	require.Equal(t, 10, runtime.NumGoroutine()-numGoru)
}

func TestEmployeeStopWorking(t *testing.T) {
	numGoru := runtime.NumGoroutine()

	validConfig := initValidConfig(t, 300000)

	empl := Employee{
		StopWork:   make(chan struct{}),
		ResultChan: make(chan models.IParcell),
		ErrorChan:  make(chan ErrorMessage),
	}

	go empl.StartWorking(validConfig)
	time.Sleep(10 * time.Second)
	require.Equal(t, 10, runtime.NumGoroutine()-numGoru)
	close(empl.StopWork)
	time.Sleep(20 * time.Second)
	require.Equal(t, numGoru, runtime.NumGoroutine())
}

func TestGetResult(t *testing.T) {
	validConfig := initValidConfig(t, 30000)

	resultChan := make(chan models.IParcell)

	empl := Employee{
		StopWork:   make(chan struct{}),
		ResultChan: resultChan,
		ErrorChan:  make(chan ErrorMessage),
	}

	go empl.StartWorking(validConfig)

	select {
	case <-time.After(30 * time.Second):
		require.FailNow(t, "timeout")
	case rs := <-resultChan:
		require.NotNil(t, rs)
		require.Equal(t, "HelloWorld", rs.(*Result).Data)
	}
}

func TestErrorMessage(t *testing.T) {
	invalidConfig := initInvalidConfig(t, 30000)

	errChan := make(chan ErrorMessage)

	empl := Employee{
		StopWork:   make(chan struct{}),
		ResultChan: make(chan models.IParcell),
		ErrorChan:  errChan,
	}

	go empl.StartWorking(invalidConfig)

	select {
	case <-time.After(30 * time.Second):
		require.FailNow(t, "timeout")
	case err := <-errChan:
		require.Greater(t, len(err.Message), 0)
	}
}

func TestChangeInterval(t *testing.T) {
	validConfig := initValidConfig(t, 10000)

	resultChan := make(chan models.IParcell)

	empl := Employee{
		StopWork:   make(chan struct{}),
		ResultChan: resultChan,
		ErrorChan:  make(chan ErrorMessage),
	}

	go empl.StartWorking(validConfig)

	firstResultTime := time.Now()
	select {
	case <-time.After(30 * time.Second):
		require.FailNow(t, "timeout")
	case rs := <-resultChan:
		require.NotNil(t, rs)
		require.Equal(t, "HelloWorld", rs.(*Result).Data)
		firstResultTime = time.Now()
	}

	select {
	case <-time.After(30 * time.Second):
		require.FailNow(t, "timeout")
	case rs := <-resultChan:
		require.NotNil(t, rs)
		require.Equal(t, "HelloWorld", rs.(*Result).Data)
		require.LessOrEqual(t, time.Duration(16), time.Since(firstResultTime)/time.Second) // actual time is time+6s
	}

	validConfig.Interval = 15000

	empl.UpdateConfig(validConfig)

	select {
	case <-time.After(60 * time.Second):
		require.FailNow(t, "timeout")
	case rs := <-resultChan:
		require.NotNil(t, rs)
		require.Equal(t, "HelloWorld", rs.(*Result).Data)
		firstResultTime = time.Now()
	}

	select {
	case <-time.After(60 * time.Second):
		require.FailNow(t, "timeout")
	case rs := <-resultChan:
		require.NotNil(t, rs)
		require.Equal(t, "HelloWorld", rs.(*Result).Data)
		require.LessOrEqual(t, time.Duration(21), time.Since(firstResultTime)/time.Second) // actual time is time+6s
	}
}
