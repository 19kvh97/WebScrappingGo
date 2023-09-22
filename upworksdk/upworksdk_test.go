package upworksdk

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
