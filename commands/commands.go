package commands

type CommandType string

const (
	CreateCourseCommand             CommandType = "create_course"
	UpdateCourseCommand             CommandType = "update_course"
	GetCoursesCommand               CommandType = "get_courses"
	DeleteCoursesCommand            CommandType = "delete_course"
	CreateScheduleCommand           CommandType = "create_schedule"
	CreateAdditionalScheduleCommand CommandType = "create_aditional_schedule"
	GetScheduleTodayCommand         CommandType = "get_schedule_today"
	GetScheduleCommonCommand        CommandType = "get_schedule_common"
	CancelCommand                   CommandType = "cancel"
	ClearScheduleCommand            CommandType = "clear_schedule"
	LinkOptionalCourseCommand       CommandType = "link_optional_course"
)
