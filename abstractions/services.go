package abstractions

import "telegram-notification-bot-core/dto"

type ICourseService interface {
	CreateNewCourse(request dto.CreateNewCourseRequest) (string, error)
	UpdateCourse(request dto.UpdateCourseInfoRequest) error
	DeleteCourse(request dto.ArchiveCourseRequest) error
	GetCourses() (*dto.GetCoursesResponse, error)
	GetOptionalCourses() (*dto.GetCoursesResponse, error)
	GetCourseById(id string) (*dto.CourseDto, error)
}

type IScheduleService interface {
	CreateNewSchedule(request dto.CreateNewScheduleRequest) error
	ClearSchedule() error
	InsertAdditionalSchedule(request dto.CreateNewAdditionalScheduleRequest) error
	GetCurrentSchedule(userId int) (*dto.GetScheduleResponse, error)
	GetCommonSchedule(userId int) (*dto.GetCommonScheduleResponse, error)
	PrepareSchedulesListForNotify(userIds []int) (map[int][]dto.ScheduleDto, error)
	LinkOptionalCourseToUser(request dto.LinkOptionalCourseToUserRequest) error
}

type IBackgroundService interface {
	Run()
}

type IActionService interface {
	SaveUserCurrentState(id int, action dto.UserActionDto)
	GetUserCurrentState(ud int) dto.UserActionDto
}
