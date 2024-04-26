package providers

import (
	"encoding/json"
	"sync"
	"telegram-notification-bot-core/dto"
)

type ActionProvider struct {
	common *CommonProvider
	mutex  *sync.RWMutex
}

func NewActionProvider() *ActionProvider {
	common := newCommonProvider("actions")
	return &ActionProvider{common: common, mutex: &sync.RWMutex{}}
}

func (a ActionProvider) StoreData(m map[int]dto.UserActionDto) error {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	data, err := json.Marshal(m)

	if err != nil {
		return err
	}

	return a.common.saveAllDataToStorage(data)
}

func (a ActionProvider) RestoreData() (map[int]dto.UserActionDto, error) {
	data, err := a.common.getAllDataFromStorage()

	if err != nil {
		return nil, err
	}

	var m map[int]dto.UserActionDto

	err = json.Unmarshal(data, &m)

	if err != nil {
		return nil, err
	}

	return m, nil
}
