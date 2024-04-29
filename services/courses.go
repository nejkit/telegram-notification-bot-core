package services

import (
	"errors"
	"telegram-notification-bot-core/abstractions"
	"telegram-notification-bot-core/dao"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/exceptions"
)

type CourseService struct {
	provider abstractions.ICourseProvider
}

func NewCourseService(provider abstractions.ICourseProvider) *CourseService {
	return &CourseService{provider: provider}
}

func (c CourseService) CreateNewCourse(request dto.CreateNewCourseRequest) (string, error) {
	_, err := c.provider.GetCourseByParams(request.Name)

	if err == exceptions.NotFound {
		id, err := c.provider.CreateNewCourse(dao.CourseModel{
			Name:           request.Name,
			TeacherName:    request.TeacherName,
			TeacherContact: request.TeacherContact,
			MeetLink:       request.MeetLink,
			IsOptional:     request.IsOptional,
		})

		if err != nil {
			return "", err
		}

		return id, nil
	}

	if err == nil {
		return "", errors.New("AlreadyExists")
	}

	return "", err
}

func (c CourseService) UpdateCourse(request dto.UpdateCourseInfoRequest) error {
	return c.provider.UpdateCourse(dao.CourseModel{
		Id:             request.Id,
		Name:           request.Name,
		TeacherName:    request.TeacherName,
		TeacherContact: request.TeacherContact,
		MeetLink:       request.MeetLink,
		IsOptional:     request.IsOptional,
	})
}

func (c CourseService) DeleteCourse(request dto.ArchiveCourseRequest) error {
	return c.provider.ArchiveCourse(request.CourseId)
}

func (c CourseService) GetOptionalCourses() (*dto.GetCoursesResponse, error) {
	courses, err := c.provider.GetCourses()

	if err != nil {
		return nil, err
	}

	var coursesDto []dto.CourseDto

	for _, course := range courses {
		if !course.IsOptional {
			continue
		}

		coursesDto = append(coursesDto, dto.CourseDto{
			Name:           course.Name,
			Id:             course.Id,
			TeacherName:    course.TeacherName,
			TeacherContact: course.TeacherContact,
			MeetLink:       course.MeetLink,
		})
	}

	return &dto.GetCoursesResponse{
		Courses: coursesDto,
	}, nil
}

func (c CourseService) GetCourses() (*dto.GetCoursesResponse, error) {
	courses, err := c.provider.GetCourses()

	if err != nil {
		return nil, err
	}

	var coursesDto []dto.CourseDto

	for _, course := range courses {
		coursesDto = append(coursesDto, dto.CourseDto{
			Name:           course.Name,
			Id:             course.Id,
			TeacherName:    course.TeacherName,
			TeacherContact: course.TeacherContact,
			MeetLink:       course.MeetLink,
		})
	}

	return &dto.GetCoursesResponse{
		Courses: coursesDto,
	}, nil
}

func (c CourseService) GetCourseById(id string) (*dto.CourseDto, error) {

	course, err := c.provider.GetCourseById(id)

	if err != nil {
		return nil, err
	}

	return &dto.CourseDto{
		Name:           course.Name,
		Id:             course.Id,
		TeacherName:    course.TeacherName,
		TeacherContact: course.TeacherContact,
		MeetLink:       course.MeetLink,
	}, nil

}
