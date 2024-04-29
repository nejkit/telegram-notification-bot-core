package services

import (
	"errors"
	"sort"
	"telegram-notification-bot-core/abstractions"
	"telegram-notification-bot-core/configuration"
	"telegram-notification-bot-core/dao"
	"telegram-notification-bot-core/dto"
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

	return s.provider.CreateNewSchedule(dao.ScheduleModel{
		Weekday:   request.Weekday,
		WeekOrder: request.WeekOrder,
		CourseId:  request.CourseId,
		Order:     request.Order,
	})
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

func (s ScheduleService) GetCurrentSchedule() (*dto.GetScheduleResponse, error) {
	currentTime := util.GetMidnightTime()
	schedule, err := s.provider.GetScheduleByDate(currentTime)

	if err != nil {
		return nil, err
	}

	var schedules []dto.ScheduleDto

	for _, val := range schedule {
		courseInfo, err := s.courseProvider.GetCourseById(val.CourseId)

		if err != nil {
			continue
		}
		schedules = append(schedules, dto.ScheduleDto{
			CourseInfo: dto.CourseDto{
				Name:           courseInfo.Name,
				Id:             courseInfo.Id,
				TeacherName:    courseInfo.TeacherName,
				TeacherContact: courseInfo.TeacherContact,
				MeetLink:       courseInfo.MeetLink,
			},
			Order:     val.Order,
			WeekOrder: val.WeekOrder,
		})
	}

	sort.Slice(schedules, func(i, j int) bool {
		return schedules[i].Order < schedules[j].Order
	})

	return &dto.GetScheduleResponse{
		CurrentDate:      currentTime,
		CurrentWeekOrder: util.GetCurrentWeekOrder(),
		Schedules:        schedules,
	}, err
}

func (s ScheduleService) GetCommonSchedule() *dto.GetCommonScheduleResponse {
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

			courseDto, _ := s.courseProvider.GetCourseById(v.CourseId)

			orderToSchedules[v.Order] = append(values,
				dto.ScheduleDto{
					CourseInfo: dto.CourseDto{
						Name:           courseDto.Name,
						Id:             courseDto.Id,
						TeacherName:    courseDto.TeacherName,
						TeacherContact: courseDto.TeacherContact,
						MeetLink:       courseDto.MeetLink,
					},
					Order:     v.Order,
					WeekOrder: v.WeekOrder,
				})

		}

		resultDto.Schedules[key] = dto.CommonScheduleDto{OrderToSchedules: orderToSchedules}
	}

	return &resultDto
}
