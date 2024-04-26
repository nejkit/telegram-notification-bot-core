package dao

import "time"

type AdditionalScheduleModel struct {
	Id             string
	AdditionalTime time.Time
	Order          int
	CourseId       string
}
