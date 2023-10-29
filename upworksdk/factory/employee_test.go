package factory

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk"
	"github.com/19kvh97/webscrappinggo/upworksdk/models"
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

func TestEmployeeStartWorking(t *testing.T) {
	numGoru := runtime.NumGoroutine()

	validConfig := initValidConfig(t, 30000)

	empl := Employee{
		UpdateJobChan: make(chan models.Config),
		StopWork:      make(chan struct{}),
		ResultChan:    make(chan models.IParcell),
		ErrorChan:     make(chan WorkFailed),
	}

	go empl.StartWorking(validConfig)
	time.Sleep(20 * time.Second)
	require.Equal(t, 13, runtime.NumGoroutine()-numGoru)
}

func TestEmployeeStopWorking(t *testing.T) {
	numGoru := runtime.NumGoroutine()

	validConfig := initValidConfig(t, 300000)

	empl := Employee{
		UpdateJobChan: make(chan models.Config),
		StopWork:      make(chan struct{}),
		ResultChan:    make(chan models.IParcell),
		ErrorChan:     make(chan WorkFailed),
	}

	go empl.StartWorking(validConfig)
	time.Sleep(10 * time.Second)
	require.Equal(t, 13, runtime.NumGoroutine()-numGoru)
	close(empl.StopWork)
	time.Sleep(11 * time.Second)
	require.Equal(t, numGoru, runtime.NumGoroutine())
}
