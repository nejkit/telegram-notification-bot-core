package dto

import "time"

type NotificationInfoDto struct {
	ChatId           int64
	Data             ScheduleDto
	States           map[int]bool
	NotificationDate time.Time
}
