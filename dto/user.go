package dto

import (
	"telegram-notification-bot-core/actions"
	"telegram-notification-bot-core/commands"
	"time"
)

type UserActionDto struct {
	Action  actions.UserAction
	Command commands.CommandType
}

type CalendarPositionDto struct {
	Month time.Month
	Year  int
}
