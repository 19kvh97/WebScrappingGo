package factory

import (
	"context"
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func TestRunChromeTask(t *testing.T) {
	chromTaskChan := make(chan ChromeTask)
	numGo := runtime.NumGoroutine()
	go func(taskChan chan<- ChromeTask) {
		for i := 0; i < 10; i++ {
			taskChan <- func(ctx context.Context) {
				tasks := chromedp.Tasks{
					chromedp.Navigate("https://www.randomnumberapi.com/api/v1.0/randomnumber"),
				}

				if err := chromedp.Run(ctx, tasks); err != nil {
					log.Printf("err : %s", err.Error())
				}
			}
			time.Sleep(4 * time.Second)
		}
	}(chromTaskChan)

	go NewChromeInstance(chromTaskChan)
	time.Sleep(10 * time.Second)
	require.Equal(t, 10, runtime.NumGoroutine()-numGo)
}
