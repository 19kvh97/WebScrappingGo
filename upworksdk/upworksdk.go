package upworksdk

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/19kvh97/webscrappinggo/upworksdk/common"
	"github.com/19kvh97/webscrappinggo/upworksdk/factory"
	"github.com/19kvh97/webscrappinggo/upworksdk/models"
)

type SdkManager struct {
	factoryManager *factory.Manager
	configs        []models.Config
	mutex          sync.Mutex
}

var instance *SdkManager

func SdkInstance() *SdkManager {
	if instance == nil {
		instance = &SdkManager{}
		instance.init()
	}

	return instance
}

func (sdkM *SdkManager) init() {
	sdkM.factoryManager = factory.NewManager(5)
	sdkM.factoryManager.StartWorking()
}

func (sdkM *SdkManager) ErrorChannel() chan string {
	return sdkM.factoryManager.ErrorChannel
}

func (sdkM *SdkManager) Stop(conf models.Config) error {
	log.Printf("Stop config")

	return fmt.Errorf("stop config %s failed", conf.Id)
}

func (sdkM *SdkManager) GetActiveConfigCount() int {
	return sdkM.factoryManager.ActiveEmployeeCount()
}

func (sdkM *SdkManager) IsConfigActived(email string, mode models.RunningMode) bool {
	log.Printf("IsConfigActived email : %s, mode : %s", email, mode.GetName())
	for _, cf := range sdkM.configs {
		if cf.Account.Email == email && cf.Mode == mode {
			return sdkM.factoryManager.IsActive(cf.Id)
		}
	}

	return false
}

func (sdkM *SdkManager) DeleteConfig(email string, mode models.RunningMode) error {

	return fmt.Errorf("cant find config with email %s and mode %s", email, mode.GetName())
}

func (sdkM *SdkManager) RegisterListener(email string, mode models.RunningMode, listener func(string, models.IParcell)) error {

	return errors.New("failed to register listener. May listener registered")
}

func (sdkM *SdkManager) Run(configs ...models.Config) []string {

	var ids []string

	var updatedIndex []int
	for _, conf := range configs {
		isExisted := false
		sdkM.mutex.Lock()
		for i, existedCf := range sdkM.configs {
			if existedCf.Equal(conf) {
				isExisted = true
				sdkM.configs[i].Update(conf)
				updatedIndex = append(updatedIndex, i)
				break
			}
		}
		sdkM.mutex.Unlock()
		if isExisted {
			continue
		}

		id, err := common.GenerateUniqueId(conf)
		if err != nil {
			panic(err)
		}
		sdkM.mutex.Lock()
		sdkM.configs = append(sdkM.configs, models.Config{
			Id:       id,
			Mode:     conf.Mode,
			Account:  conf.Account,
			Interval: conf.Interval,
		})
		updatedIndex = append(updatedIndex, len(sdkM.configs)-1)
		sdkM.mutex.Unlock()
	}

	log.Printf("new added config length : %d", len(sdkM.configs))
	for _, idx := range updatedIndex {
		sdkM.factoryManager.RunConfig(sdkM.configs[idx])
		ids = append(ids, sdkM.configs[idx].Id)
	}

	return ids
}
