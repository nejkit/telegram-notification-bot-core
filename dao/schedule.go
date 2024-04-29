package dao

import (
	"telegram-notification-bot-core/util"
	"time"
)

type ScheduleModel struct {
	Id              string
	Weekday         time.Weekday
	WeekOrder       util.WeekOrder
	CourseId        string
	Order           int
	IsOptional      bool
	OptCourseParams OptionalCourseSettings
}

type OptionalCourseSettings struct {
	UserIdToCourseId map[int]string
}
