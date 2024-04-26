package services

import (
	"telegram-notification-bot-core/abstractions"
	"telegram-notification-bot-core/actions"
	"telegram-notification-bot-core/dto"
)

type ActionService struct {
	states   map[int]dto.UserActionDto
	provider abstractions.IUserActionProvider
}

func NewActionService(provider abstractions.IUserActionProvider) *ActionService {
	states, err := provider.RestoreData()

	if err != nil {
		states = make(map[int]dto.UserActionDto)
	}

	return &ActionService{states: states, provider: provider}
}

func (a ActionService) SaveUserCurrentState(id int, action dto.UserActionDto) {
	a.states[id] = action
	go a.provider.StoreData(a.states)
}

func (a ActionService) GetUserCurrentState(ud int) dto.UserActionDto {
	state, ok := a.states[ud]

	if !ok {
		return dto.UserActionDto{
			Action: actions.UserActionNone,
		}
	}

	return state
}
