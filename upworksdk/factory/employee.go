package factory

import (
	"log"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

type IEmployee interface {
	StartWorking()
}

type Employee struct {
	IEmployee
	CommandChannel chan Command
	ResultChannel  chan models.IParcell
}

func (e *Employee) StartWorking() {
	for cmd := range e.CommandChannel {
		log.Printf("Command : %s with data %v", &cmd.Name(), cmd.Data)
	}
}
