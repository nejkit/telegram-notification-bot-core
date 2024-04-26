package dto

type CreateNewCourseRequest struct {
	Name           string
	TeacherName    string
	TeacherContact string
	MeetLink       string
}

type UpdateCourseInfoRequest struct {
	Id             string
	Name           string
	TeacherName    string
	TeacherContact string
	MeetLink       string
}

type ArchiveCourseRequest struct {
	CourseId string
}

type GetCoursesResponse struct {
	Courses []CourseDto
}

type CourseDto struct {
	Name           string
	Id             string
	TeacherName    string
	TeacherContact string
	MeetLink       string
}
