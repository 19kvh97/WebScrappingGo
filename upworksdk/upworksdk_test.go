package upworksdk

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
	"github.com/stretchr/testify/require"
)

func TestWorkerProcess(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var rawCookie []models.Cookie
	content, err := ioutil.ReadFile("../hungkv_cookie.json")
	require.Nil(t, err)
	err = json.Unmarshal(content, &rawCookie)
	require.Nil(t, err)

	validCookie, err := ExtractValidateCookies(rawCookie)
	require.Nil(t, err)

	testMail := "hung.kv22011997@gmail.com"
	testPass := "testPass"

	testcase := []struct {
		cookies             []models.Cookie
		expectedResultCount int
		expectedErr         error
	}{
		{
			cookies:             validCookie,
			expectedResultCount: 1,
			expectedErr:         nil,
		},
		{
			cookies:             validCookie,
			expectedResultCount: 1,
			expectedErr:         nil,
		},
	}

	for _, test := range testcase {
		err = SdkInstance().Run(models.Config{
			Mode: models.SYNC_BEST_MATCH,
			Account: models.UpworkAccount{
				Email:    testMail,
				Password: testPass,
				Cookie:   test.cookies,
			},
		})

		require.Equal(t, test.expectedErr, err)
		time.Sleep(20 * time.Second)
	}

	time.Sleep(3 * time.Minute)
}

func TestMultipleMode(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var rawCookie []models.Cookie
	content, err := ioutil.ReadFile("../valid_cookie1.json")
	require.Nil(t, err)
	err = json.Unmarshal(content, &rawCookie)
	require.Nil(t, err)

	validCookie, err := ExtractValidateCookies(rawCookie)
	require.Nil(t, err)

	testMail := "hung.kv22011997@gmail.com"
	testPass := "testPass"

	err = SdkInstance().Run(models.Config{
		Mode: models.SYNC_BEST_MATCH,
		Account: models.UpworkAccount{
			Email:    testMail,
			Password: testPass,
			Cookie:   validCookie,
		},
	})

	require.Equal(t, nil, err)
	time.Sleep(5 * time.Second)

	err = SdkInstance().Run(models.Config{
		Mode: models.SYNC_RECENTLY,
		Account: models.UpworkAccount{
			Email:    testMail,
			Password: testPass,
			Cookie:   validCookie,
		},
	})

	require.Equal(t, nil, err)
	time.Sleep(5 * time.Minute)
}

func TestChannel(t *testing.T) {
	configChanged := make(chan string)

	configs := []models.Config{
		{
			Id: "11",
		},
		{
			Id: "22",
		},
		{
			Id: "33",
		},
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Run init")
		for {
			id := <-configChanged
			log.Printf("id changed %s", id)
			for i, cf := range configs {
				if cf.Id == id {
					switch cf.State {
					case models.NEW_STATE:
						configs[i].State = models.ACTIVE_STATE
						log.Printf("Change state of %s to ACTIVE_STATE", cf.Id)
					case models.INACTIVE_STATE:
						configs = append(configs[:i], configs[i+1:]...)
						log.Printf("Remove config with id : %s", cf.Id)
					default:
						log.Printf("Receive state %d, do nothing", cf.State)
					}
					break
				}
			}
		}
	}()

	time.Sleep(time.Second)

	for _, cf := range configs {
		wg.Add(1)
		go func(cf models.Config) {
			defer wg.Done()
			log.Printf("Run config %s", cf.Id)
			configChanged <- cf.Id
			rd := rand.Intn(20) + 5
			log.Printf("config %s sleep %ds", cf.Id, rd)
			time.Sleep(time.Duration(rd * int(time.Second)))
			log.Printf("config %s is finished", cf.Id)
			for i, conf := range configs {
				if conf.Id == cf.Id {
					configs[i].State = models.INACTIVE_STATE
					break
				}
			}
			configChanged <- cf.Id
		}(cf)
	}

	wg.Wait()
	log.Println("Finish test")
}
