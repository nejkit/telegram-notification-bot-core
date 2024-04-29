package abstractions

import (
	"telegram-notification-bot-core/dao"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/util"
	"time"
)

type ICourseProvider interface {
	CreateNewCourse(model dao.CourseModel) (string, error)
	UpdateCourse(model dao.CourseModel) error
	ArchiveCourse(id string) error
	GetCourseByParams(name string) (*dao.CourseModel, error)
	GetCourseById(id string) (*dao.CourseModel, error)
	GetCourses() ([]dao.CourseModel, error)
}

type IScheduleProvider interface {
	CreateNewSchedule(model dao.ScheduleModel) error
	GetScheduleByDate(time time.Time) ([]dao.ScheduleModel, error)
	CreateNewAdditionalSchedule(model dao.AdditionalScheduleModel) error
	ValidateAddScheduleCreation(date time.Time, order int) (bool, error)
	ValidateScheduleCreation(weekday time.Weekday, order int, weekOrder util.WeekOrder) (bool, error)
	DropAllSchedules() error
	GetCommonSchedule() map[time.Weekday][]dao.ScheduleModel
	LinkCourseToUser(userId int, courseId string) error
}

type IUserActionProvider interface {
	StoreData(map[int]dto.UserActionDto) error
	RestoreData() (map[int]dto.UserActionDto, error)
}

type IChatProvider interface {
	GetChatByUserId(userId int) (int64, error)
	SaveChatForUser(userId int, chatId int64) error
}
