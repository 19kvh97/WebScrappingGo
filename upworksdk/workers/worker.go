package workers

import (
	"context"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

type RunningMode int

const (
	SYNC_BEST_MATCH RunningMode = iota
	SYNC_RECENTLY
	SYNC_MESSAGE
)

func (rm *RunningMode) GetName() string {
	return []string{
		"SYNC_BEST_MATCH", "SYNC_RECENTLY", "SYNC_MESSAGE",
	}[*rm]
}

func (rm *RunningMode) GetLink() string {
	switch *rm {
	case SYNC_BEST_MATCH:
		return "https://www.upwork.com/nx/find-work/best-matches"
	case SYNC_RECENTLY:
		return "https://www.upwork.com/nx/find-work/most-recent"
	case SYNC_MESSAGE:
		return "https://www.upwork.com/ab/messages"
	default:
		return ""
	}
}

type IWorker interface {
	PrepareTask() func(context.Context)
	GetMode() RunningMode
	RegisterChannel(chan models.IParcell) error
}

type Worker struct {
	IWorker
	Account  models.UpworkAccount
	Channels []chan models.IParcell
}
