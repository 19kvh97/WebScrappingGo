package factory

import (
	"log"
	"sync"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

type Manager struct {
	JobChannel    chan models.Config //Listen new config from outside, one employee just do one job
	ResultChannel chan models.IParcell
	Employees     map[string]Employee
	subcribers    map[models.PackageType][]*models.Distributor
	mutex         sync.Mutex
	StopWork      chan struct{}
}

func (m *Manager) StartWorking() {
	for {
		select {
		case config := <-m.JobChannel:
			log.Printf("Manager received config with email %s and mode %s", config.Account.Email, config.Mode.GetName())
		case result := <-m.ResultChannel:
			go m.Publish(models.Package{
				Type: models.UPWORK_JOB_PACKAGE,
				Data: result,
			})
		case <-m.StopWork:
			log.Println("Manager is stop working!")
			return
		}
	}
}

//Add a distributor with the special package type
func (m *Manager) Subcribe(pgType models.PackageType, distributor *models.Distributor) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.subcribers[pgType] = append(m.subcribers[pgType], distributor)
}

func (m *Manager) Unsubcribe(pgType models.PackageType, distributor *models.Distributor) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if distributors, ok := m.subcribers[pgType]; ok {
		for i, d := range distributors {
			if d == distributor {
				m.subcribers[pgType] = append(distributors[:i], distributors[i+1:]...)
				return
			}
		}
	}
}

func (m *Manager) Publish(pg models.Package) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if distributors, ok := m.subcribers[pg.Type]; ok {
		for _, dis := range distributors {
			dis.Channel <- pg.Data
		}
	}
}
