package cache

import (
	"encoding/json"
	"fmt"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/exceptions"
)

type ActionCacheProvider struct {
	common *CommonProvider
}

const (
	userActionHashKey = "user:action"
)

func NewActionCacheProvider(common *CommonProvider) *ActionCacheProvider {
	return &ActionCacheProvider{common: common}
}

func (a *ActionCacheProvider) AddActionForAccount(userId int, actionDto dto.UserActionDto) error {
	data, err := json.Marshal(actionDto)

	if err != nil {
		return exceptions.InternalError
	}

	return a.common.saveIntoHash(userActionHashKey, fmt.Sprint(userId), data)
}

func (a *ActionCacheProvider) GetActionByAccount(userId int) (*dto.UserActionDto, error) {
	data, err := a.common.getFromHash(userActionHashKey, fmt.Sprint(userId))

	if err != nil {
		return nil, err
	}

	var actionDto dto.UserActionDto

	err = json.Unmarshal([]byte(data), &actionDto)

	if err != nil {
		return nil, exceptions.InternalError
	}

	return &actionDto, nil
}
