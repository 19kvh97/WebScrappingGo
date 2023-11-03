package factory

import (
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
	"github.com/stretchr/testify/require"
)

func TestStart2Mode(t *testing.T) {
	cf1 := initValidConfig(t, 30000)
	cf1.Id = "cf1"
	cf2 := initValidConfig(t, 50000)
	cf2.Id = "cf2"

	numGo := runtime.NumGoroutine()

	manager := Manager{
		ErrorChannel: make(chan string),
		StopWork:     make(chan struct{}),
	}

	manager.StartWorking()

	require.Equal(t, 1, runtime.NumGoroutine()-numGo)
	manager.RunConfig(cf1)
	require.Equal(t, 2, runtime.NumGoroutine()-numGo)
	manager.RunConfig(cf2)
	require.Equal(t, 3, runtime.NumGoroutine()-numGo)

	require.Equal(t, 2, manager.ActiveEmployeeCount())
}

func TestStop2Mode(t *testing.T) {
	cf1 := initValidConfig(t, 30000)
	cf1.Id = "cf1"
	cf2 := initValidConfig(t, 50000)
	cf2.Id = "cf2"

	manager := Manager{
		ErrorChannel: make(chan string),
		StopWork:     make(chan struct{}),
	}

	manager.StartWorking()

	manager.RunConfig(cf1)
	manager.RunConfig(cf2)

	distri1 := models.Distributor{
		ID:      1,
		Channel: make(chan models.IParcell),
	}

	manager.Subcribe(models.UPWORK_JOB_PACKAGE, &distri1)

	select {
	case <-time.After(30 * time.Second):
		require.FailNow(t, "timeout")
	case <-distri1.Channel:
		log.Println("receive")
	}

	manager.StopConfig(cf1.Id)
	require.Equal(t, 1, manager.ActiveEmployeeCount())
}

func TestSubcribeUnsubcribe(t *testing.T) {
	cf1 := initValidConfig(t, 10000)
	cf1.Id = "cf1"

	manager := Manager{
		ErrorChannel: make(chan string),
		StopWork:     make(chan struct{}),
	}

	manager.StartWorking()
	manager.RunConfig(cf1)

	distri1 := models.Distributor{
		ID:      1,
		Channel: make(chan models.IParcell),
	}

	manager.Subcribe(models.UPWORK_JOB_PACKAGE, &distri1)

	select {
	case <-time.After(30 * time.Second):
		require.FailNow(t, "timeout")
	case job := <-distri1.Channel:
		require.Greater(t, len(job.(*Result).Data), 0)
	}

	manager.Unsubcribe(models.UPWORK_JOB_PACKAGE, &distri1)
	select {
	case <-time.After(20 * time.Second):
	case <-distri1.Channel:
		require.FailNow(t, "Unscribe failed")
	}
}
