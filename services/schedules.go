package services

import (
	"errors"
	"sort"
	"telegram-notification-bot-core/abstractions"
	"telegram-notification-bot-core/configuration"
	"telegram-notification-bot-core/dao"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/exceptions"
	"telegram-notification-bot-core/util"
	"time"
)

type ScheduleService struct {
	config         configuration.Configuration
	provider       abstractions.IScheduleProvider
	courseProvider abstractions.ICourseProvider
}

func NewScheduleService(config configuration.Configuration, provider abstractions.IScheduleProvider, courseProvider abstractions.ICourseProvider) *ScheduleService {
	return &ScheduleService{config: config, provider: provider, courseProvider: courseProvider}
}

func (s ScheduleService) CreateNewSchedule(request dto.CreateNewScheduleRequest) error {
	_, ok := s.config.ScheduleSettings.TimeSlotsConfiguration[request.Order]

	if !ok {
		return errors.New("InvalidOrder")
	}

	ok, err := s.provider.ValidateScheduleCreation(request.Weekday, request.Order, request.WeekOrder)

	if err != nil {
		return err
	}

	if !ok {
		return err
	}

	daoModel := dao.ScheduleModel{
		Weekday:    request.Weekday,
		WeekOrder:  request.WeekOrder,
		Order:      request.Order,
		IsOptional: request.IsOptional,
	}

	if request.IsOptional {
		daoModel.OptCourseParams = dao.OptionalCourseSettings{UserIdToCourseId: map[int]string{}}
	} else {
		daoModel.CourseId = request.CourseId
	}

	return s.provider.CreateNewSchedule(daoModel)
}

func (s ScheduleService) ClearSchedule() error {
	return s.provider.DropAllSchedules()
}

func (s ScheduleService) InsertAdditionalSchedule(request dto.CreateNewAdditionalScheduleRequest) error {
	_, ok := s.config.ScheduleSettings.TimeSlotsConfiguration[request.Order]

	if !ok {
		return errors.New("InvalidOrder")
	}

	ok, err := s.provider.ValidateAddScheduleCreation(request.Date, request.Order)

	if err != nil {
		return err
	}

	if !ok {
		return err
	}

	daoModel := dao.AdditionalScheduleModel{
		AdditionalTime: request.Date,
		Order:          request.Order,
		IsEmpty:        request.IsEmpty,
	}

	if !request.IsEmpty {
		daoModel.CourseId = request.CourseId
	}

	return s.provider.CreateNewAdditionalSchedule(daoModel)
}

func (s ScheduleService) GetCurrentSchedule(userId int) (*dto.GetScheduleResponse, error) {
	currentTime := util.GetMidnightTime()
	schedule, err := s.provider.GetScheduleByDate(currentTime)

	if err != nil {
		return nil, err
	}

	result := s.enrichScheduleInfoByUserId(schedule, userId)

	return &result, nil
}

func (s ScheduleService) GetCommonSchedule(userId int) (*dto.GetCommonScheduleResponse, error) {
	result := s.provider.GetCommonSchedule()

	resultDto := dto.GetCommonScheduleResponse{
		Schedules: map[time.Weekday]dto.CommonScheduleDto{},
	}

	for key, val := range result {
		orderToSchedules := map[int][]dto.ScheduleDto{}

		sort.Slice(val, func(i, j int) bool {
			return val[i].Order < val[j].Order
		})

		for _, v := range val {
			values := orderToSchedules[v.Order]

			if values == nil {
				values = []dto.ScheduleDto{}
			}

			var courseInfo *dao.CourseModel

			if v.IsOptional {
				courseId, exists := v.OptCourseParams.UserIdToCourseId[userId]

				if !exists {
					return nil, exceptions.OptionalCourseNotSelected
				}

				courseInfo, _ = s.courseProvider.GetCourseById(courseId)
			} else {
				courseInfo, _ = s.courseProvider.GetCourseById(v.CourseId)
			}

			orderToSchedules[v.Order] = append(values,
				dto.ScheduleDto{
					CourseInfo: dto.CourseDto{
						Name:           courseInfo.Name,
						Id:             courseInfo.Id,
						TeacherName:    courseInfo.TeacherName,
						TeacherContact: courseInfo.TeacherContact,
						MeetLink:       courseInfo.MeetLink,
					},
					Order:     v.Order,
					WeekOrder: v.WeekOrder,
				})

		}

		resultDto.Schedules[key] = dto.CommonScheduleDto{OrderToSchedules: orderToSchedules}
	}

	return &resultDto, nil
}

func (s ScheduleService) LinkOptionalCourseToUser(request dto.LinkOptionalCourseToUserRequest) error {
	return s.provider.LinkCourseToUser(request.UserId, request.CourseId)
}

func (s ScheduleService) PrepareSchedulesListForNotify(userIds []int) (map[int][]dto.ScheduleDto, error) {
	currentTime := util.GetMidnightTime()
	schedule, err := s.provider.GetScheduleByDate(currentTime)

	if err != nil {
		return nil, err
	}

	resultMap := map[int][]dto.ScheduleDto{}

	for _, userId := range userIds {

		resultMap[userId] = s.enrichScheduleInfoByUserId(schedule, userId).Schedules
	}

	return resultMap, nil
}

func (s ScheduleService) enrichScheduleInfoByUserId(schedule []dao.ScheduleModel, userId int) dto.GetScheduleResponse {

	var schedules []dto.ScheduleDto

	for _, val := range schedule {

		var courseInfo *dao.CourseModel

		if !val.IsOptional {
			courseInfo, _ = s.courseProvider.GetCourseById(val.CourseId)
		} else {
			courseId, ex := val.OptCourseParams.UserIdToCourseId[userId]
			if !ex {
				courseInfo = &dao.CourseModel{
					Id:             "",
					Name:           "Не обрано опціональний курс",
					TeacherName:    "",
					TeacherContact: "",
					MeetLink:       "",
				}
			} else {
				courseInfo, _ = s.courseProvider.GetCourseById(courseId)
			}
		}

		scheduleDto := dto.ScheduleDto{
			Order:     val.Order,
			WeekOrder: val.WeekOrder,
			CourseInfo: dto.CourseDto{
				Name:           courseInfo.Name,
				Id:             courseInfo.Id,
				TeacherName:    courseInfo.TeacherName,
				TeacherContact: courseInfo.TeacherContact,
				MeetLink:       courseInfo.MeetLink,
			},
		}

		schedules = append(schedules, scheduleDto)
	}

	sort.Slice(schedules, func(i, j int) bool {
		return schedules[i].Order < schedules[j].Order
	})

	return dto.GetScheduleResponse{
		CurrentDate:      util.GetMidnightTime(),
		CurrentWeekOrder: util.GetCurrentWeekOrder(),
		Schedules:        schedules,
	}
}

func (s ScheduleService) GetSchedulesIdsWithOptionalCourse() ([]string, error) {
	schedules := s.provider.GetCommonSchedule()

	var result []string

	for _, schedule := range schedules {
		for _, v := range schedule {
			if v.IsOptional {
				result = append(result, v.Id)
			}
		}
	}

	if result == nil {
		return nil, exceptions.NotFound
	}

	if result != nil && len(result) == 0 {
		return nil, exceptions.NotFound
	}

	return result, nil
}
