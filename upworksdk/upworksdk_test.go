package upworksdk

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk/common"
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
	content, err := ioutil.ReadFile("../valid_cookie1.json")

	log.Printf("%v", err)
	require.Nil(t, err)

	err = json.Unmarshal(content, &rawCookie)
	require.Nil(t, err)

	validCookie, err := common.ExtractValidateCookies(rawCookie)
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
	content, err := ioutil.ReadFile("../expired_cookie.json")

	log.Printf("%v", err)
	require.Nil(t, err)

	err = json.Unmarshal(content, &rawCookie)
	require.Nil(t, err)

	validCookie, err := common.ExtractValidateCookies(rawCookie)
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

//realTest
func TestRunValidConfig(t *testing.T) {
	cf := initValidConfig(t, 30)

	ids := SdkInstance().Run(cf)
	require.Equal(t, len(ids), 1)
	time.Sleep(20 * time.Second)
	isActive := SdkInstance().IsConfigActived(cf.Account.Email, cf.Mode)
	require.Equal(t, true, isActive)
}

func TestRunInvalidConfig(t *testing.T) {

}

func TestRestartValidConfig(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cf := initValidConfig(t, 30)

	numGoroutine := runtime.NumGoroutine()
	log.Printf("NumGoroutine before %d", numGoroutine)
	err := SdkInstance().Run(cf)

	require.Nil(t, err)

	log.Printf("numGorountine after %d", runtime.NumGoroutine())
	isConfigActived := false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, true)

	time.Sleep(5 * time.Second)
	log.Printf("numGorountine after %d", runtime.NumGoroutine())

	require.Equal(t, 11, runtime.NumGoroutine()-numGoroutine)

	//restart
	cf.Interval = 50000
	err = SdkInstance().Run(cf)

	require.Nil(t, err)

	log.Printf("numGorountine after %d", runtime.NumGoroutine())
	isConfigActived = false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, true)

	time.Sleep(10 * time.Second)
	log.Printf("numGorountine after %d", runtime.NumGoroutine())

	require.Equal(t, runtime.NumGoroutine()-numGoroutine, 11)
}

func TestRestartInvalidConfig(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cf := initValidConfig(t, 30)

	numGoroutine := runtime.NumGoroutine()
	log.Printf("NumGoroutine before %d", numGoroutine)
	err := SdkInstance().Run(cf)

	require.Nil(t, err)

	log.Printf("numGorountine after %d", runtime.NumGoroutine())
	isConfigActived := false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, true)

	time.Sleep(5 * time.Second)
	log.Printf("numGorountine after %d", runtime.NumGoroutine())

	require.Equal(t, 11, runtime.NumGoroutine()-numGoroutine)

	//restart
	cf.Interval = 50000
	err = SdkInstance().Run(cf)

	require.Nil(t, err)

	log.Printf("numGorountine after %d", runtime.NumGoroutine())
	isConfigActived = false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, true)

	time.Sleep(10 * time.Second)
	log.Printf("numGorountine after %d", runtime.NumGoroutine())

	require.Equal(t, runtime.NumGoroutine()-numGoroutine, 11)
}

func TestStopValidConfig(t *testing.T) {
	cf := initValidConfig(t, 30)

	numGoroutine := runtime.NumGoroutine()
	log.Printf("NumGoroutine before %d", numGoroutine)
	err := SdkInstance().Run(cf)

	require.Nil(t, err)

	isConfigActived := false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, true)
	time.Sleep(5 * time.Second)

	err = SdkInstance().Stop(cf)

	require.Nil(t, err)

	time.Sleep(5 * time.Second)

	isConfigActived = false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, false)
}

func TestStopInvalidConfig(t *testing.T) {
	cf := initValidConfig(t, 30)

	numGoroutine := runtime.NumGoroutine()
	log.Printf("NumGoroutine before %d", numGoroutine)
	err := SdkInstance().Run(cf)

	require.Nil(t, err)

	isConfigActived := false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, true)
	time.Sleep(5 * time.Second)

	err = SdkInstance().Stop(cf)

	require.Nil(t, err)

	time.Sleep(5 * time.Second)

	isConfigActived = false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, false)
}

func TestRegisterListener(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cf := initValidConfig(t, 30*1000)

	err := SdkInstance().Run(cf)

	require.Nil(t, err)

	isConfigActived := false
	for i := 0; i < 3; i++ {
		if SdkInstance().IsConfigActived(cf.Account.Email, models.SYNC_BEST_MATCH) {
			isConfigActived = true
			break
		}
		time.Sleep(time.Second)
	}

	require.Equal(t, isConfigActived, true)

	passChannel := make(chan bool)
	jobCount := 0
	err = SdkInstance().RegisterListener(cf.Account.Email, models.SYNC_BEST_MATCH, func(email string, parcell models.IParcell) {
		log.Println("received data")
		if job, ok := parcell.(models.Job); ok {
			log.Printf("Job title: %s", job.Title)
			jobCount++
			if jobCount > 50 {
				passChannel <- true
			}
		}
	})

	require.Nil(t, err)

	isPass := <-passChannel
	require.Equal(t, isPass, true)
}
