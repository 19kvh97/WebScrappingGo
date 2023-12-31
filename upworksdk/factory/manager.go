package factory

import (
	"log"
	"sync"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

type Manager struct {
	resultChannel      chan models.IParcell //
	internalErrChannel chan ErrorMessage    //
	ErrorChannel       chan string          //
	employees          map[string]Employee  // string is config id
	subcribers         map[models.PackageType][]*models.Distributor
	mutex              sync.Mutex
	StopWork           chan struct{}
	MaxEmployees       int
}

func NewManager(maxEmployees int) *Manager {
	return &Manager{
		resultChannel:      make(chan models.IParcell),
		internalErrChannel: make(chan ErrorMessage, maxEmployees),
		ErrorChannel:       make(chan string),
		employees:          make(map[string]Employee),
		subcribers:         make(map[models.PackageType][]*models.Distributor),
		StopWork:           make(chan struct{}),
		MaxEmployees:       maxEmployees,
	}
}

func (m *Manager) StartWorking() {
	go func() {
		for {
			select {
			case result := <-m.resultChannel:
				go m.publish(models.Package{
					Type: models.UPWORK_JOB_PACKAGE,
					Data: result,
				})
			case <-m.StopWork:
				log.Println("Manager is stop working!")
				return
			case msg := <-m.internalErrChannel:
				log.Printf("msg : %v", msg)
				m.StopConfig(msg.ConfigID)
			}
		}
	}()
}

func (m *Manager) RunConfig(cf models.Config) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	log.Printf("Run or update config %s", cf.Id)

	if employee, ok := m.employees[cf.Id]; ok {
		employee.UpdateConfig(cf)
	} else {
		employee := Employee{
			StopWork:   make(chan struct{}),
			ResultChan: m.resultChannel,
			ErrorChan:  m.internalErrChannel,
		}
		go employee.StartWorking(cf)
		m.employees[cf.Id] = employee
	}
}

func (m *Manager) StopConfig(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if empl, ok := m.employees[id]; ok {
		close(empl.StopWork)
		delete(m.employees, id)
	} else {
		log.Printf("employe with id %s not existed", id)
	}
}

func (m *Manager) IsActive(configId string) bool {
	_, ok := m.employees[configId]
	return ok
}

func (m *Manager) ActiveEmployeeCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.employees)
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

func (m *Manager) publish(pg models.Package) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if distributors, ok := m.subcribers[pg.Type]; ok {
		for _, dis := range distributors {
			dis.Channel <- pg.Data
		}
	}
}
