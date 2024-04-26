package dto

import (
	"telegram-notification-bot-core/util"
	"time"
)

type CreateNewScheduleRequest struct {
	CourseId  string
	Weekday   time.Weekday
	WeekOrder util.WeekOrder
	Order     int
}

type CreateNewAdditionalScheduleRequest struct {
	CreateNewScheduleRequest
	Date time.Time
}

type GetScheduleResponse struct {
	CurrentDate      time.Time
	CurrentWeekOrder util.WeekOrder
	Schedules        []ScheduleDto
}

type GetCommonScheduleResponse struct {
	Schedules map[time.Weekday]CommonScheduleDto
}

type CommonScheduleDto struct {
	OrderToSchedules map[int][]ScheduleDto
}

type ScheduleDto struct {
	CourseInfo CourseDto
	Order      int
	WeekOrder  util.WeekOrder
}
