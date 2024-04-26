package actions

type UserAction int

const (
	UserActionNone                UserAction = 0
	UserActionInputCourseName     UserAction = 1
	UserActionInputTeacherName    UserAction = 2
	UserActionInputTeacherContact UserAction = 3
	UserActionInputMeetLink       UserAction = 4
	UserActionChooseCourse        UserAction = 6
	UserActionInputWeekday        UserAction = 7
	UserActionInputWeekOrder      UserAction = 8
	UserActionInputOrder          UserAction = 9
	UserActionInputDate           UserAction = 10
)
