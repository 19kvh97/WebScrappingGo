package workers

import (
	"context"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

type IWorker interface {
	PrepareTask() func(context.Context)
	GetMode() models.RunningMode
	RegisterChannel(chan models.IParcell) error
}

type Worker struct {
	IWorker
	Account  models.UpworkAccount
	Channels []chan models.IParcell
}
