package workers

import (
	"context"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

type IWorker interface {
	PrepareTask() (func(context.Context), error)
	GetMode() models.RunningMode
	SendResult(models.IParcell)
}

type Worker struct {
	IWorker
	Account  models.UpworkAccount
	Listener func(models.IParcell)
}

func (w Worker) SendResult(parsell models.IParcell) {
	if w.Listener != nil {
		w.Listener(parsell)
	}

}

func (w Worker) PrepareTask() (func(context.Context), error) {
	return func(ctx context.Context) {}, nil
}

func (w Worker) GetMode() models.RunningMode {
	return 0
}
