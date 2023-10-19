package workers

import (
	"context"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

type IWorker interface {
	PrepareTask() (func(context.Context), error)
	GetMode() models.RunningMode
	SendResult(models.IParcell)
	Stop()
	IsRunning() bool
}

type Worker struct {
	IWorker
	Account  models.UpworkAccount
	Listener func(string, models.IParcell)
	IsActive bool
}

func (w Worker) SendResult(parsell models.IParcell) {
	if w.Listener != nil {
		w.Listener(w.Account.Email, parsell)
	}

}

func (w *Worker) PrepareTask() (func(context.Context), error) {
	return func(ctx context.Context) {}, nil
}

func (w Worker) GetMode() models.RunningMode {
	return 0
}

func (w *Worker) Stop() {
	w.IsActive = false
}

func (w *Worker) IsRunning() bool {
	return w.IsActive
}
