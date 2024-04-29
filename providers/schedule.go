package providers

import (
	"encoding/json"
	"github.com/google/uuid"
	"sync"
	"telegram-notification-bot-core/dao"
	"telegram-notification-bot-core/util"
	"time"
)

type ScheduleProvider struct {
	scheduleCommon   *CommonProvider
	additionalCommon *CommonProvider
	scheduleCache    map[time.Weekday][]dao.ScheduleModel
	additionalCache  map[string][]dao.AdditionalScheduleModel
	mutex            *sync.RWMutex
}

func (s *ScheduleProvider) DropAllSchedules() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.scheduleCache = map[time.Weekday][]dao.ScheduleModel{
		time.Monday:    {},
		time.Tuesday:   {},
		time.Wednesday: {},
		time.Thursday:  {},
		time.Friday:    {},
		time.Saturday:  {},
		time.Sunday:    {},
	}
	s.additionalCache = map[string][]dao.AdditionalScheduleModel{}

	data, err := json.Marshal(s.scheduleCache)

	if err != nil {
		return err
	}

	err = s.scheduleCommon.saveAllDataToStorage(data)

	if err != nil {
		return err
	}

	data, err = json.Marshal(s.additionalCache)

	if err != nil {
		return err
	}

	err = s.additionalCommon.saveAllDataToStorage(data)

	if err != nil {
		return err
	}

	return nil
}

func (s *ScheduleProvider) CreateNewSchedule(model dao.ScheduleModel) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	id := uuid.NewString()
	model.Id = id

	courses := s.scheduleCache[model.Weekday]

	if courses == nil {
		courses = []dao.ScheduleModel{}
	}

	s.scheduleCache[model.Weekday] = append(courses, model)

	data, err := json.Marshal(s.scheduleCache)

	if err != nil {
		return err
	}

	err = s.scheduleCommon.saveAllDataToStorage(data)

	if err != nil {
		return err
	}

	return nil
}

func (s *ScheduleProvider) GetScheduleByDate(date time.Time) ([]dao.ScheduleModel, error) {

	var schedules []dao.ScheduleModel
	curWeekOrder := util.GetCurrentWeekOrder()
	curWeekday := date.Weekday()
	additional := s.additionalCache[date.Format("2006-01-02")]
	excludedOrders := map[int]struct{}{}

	if additional != nil && len(additional) > 0 {

		for _, val := range additional {
			excludedOrders[val.Order] = struct{}{}

			// we exclude this schedule order, by not add info about additional
			if val.IsEmpty {
				continue
			}

			schedules = append(schedules, dao.ScheduleModel{
				Id:        val.Id,
				Weekday:   val.AdditionalTime.Weekday(),
				WeekOrder: curWeekOrder,
				CourseId:  val.CourseId,
				Order:     val.Order,
			})
		}

	}

	usualities := s.scheduleCache[curWeekday]

	if usualities == nil {
		return schedules, nil
	}

	for _, val := range usualities {
		if val.WeekOrder != curWeekOrder && val.WeekOrder > 0 {
			continue
		}

		_, exclude := excludedOrders[val.Order]

		if exclude {
			continue
		}

		schedules = append(schedules, val)
	}

	return schedules, nil
}

func (s *ScheduleProvider) CreateNewAdditionalSchedule(model dao.AdditionalScheduleModel) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	id := uuid.NewString()

	model.Id = id

	courses := s.additionalCache[model.AdditionalTime.Format("2006-01-02")]

	if courses == nil {
		courses = []dao.AdditionalScheduleModel{}
	}

	s.additionalCache[model.AdditionalTime.Format("2006-01-02")] = append(courses, model)

	data, err := json.Marshal(s.additionalCache)

	if err != nil {
		return err
	}

	return s.additionalCommon.saveAllDataToStorage(data)
}

func NewScheduleProvider() *ScheduleProvider {
	common := newCommonProvider("schedules")
	addCommon := newCommonProvider("additionals")

	scheduleCache := map[time.Weekday][]dao.ScheduleModel{
		time.Monday:    {},
		time.Tuesday:   {},
		time.Wednesday: {},
		time.Thursday:  {},
		time.Friday:    {},
		time.Saturday:  {},
		time.Sunday:    {},
	}
	addCache := make(map[string][]dao.AdditionalScheduleModel)

	data, err := common.getAllDataFromStorage()

	if err == nil {
		err = json.Unmarshal(data, &scheduleCache)

		if err != nil {
			scheduleCache = map[time.Weekday][]dao.ScheduleModel{
				time.Monday:    {},
				time.Tuesday:   {},
				time.Wednesday: {},
				time.Thursday:  {},
				time.Friday:    {},
				time.Saturday:  {},
				time.Sunday:    {},
			}
		}
	}

	data, err = addCommon.getAllDataFromStorage()

	if err == nil {
		err = json.Unmarshal(data, &addCache)

		if err != nil {
			addCache = make(map[string][]dao.AdditionalScheduleModel)
		}
	}

	return &ScheduleProvider{
		scheduleCommon:   common,
		additionalCommon: addCommon,
		scheduleCache:    scheduleCache,
		additionalCache:  addCache,
		mutex:            &sync.RWMutex{}}
}

func (s *ScheduleProvider) ValidateScheduleCreation(weekday time.Weekday, order int, weekOrder util.WeekOrder) (bool, error) {
	schedule, ok := s.scheduleCache[weekday]

	if !ok {
		return true, nil
	}

	if schedule == nil {
		return true, nil
	}

	for _, val := range schedule {
		//if weekorder diff
		if weekOrder > 0 && val.WeekOrder > 0 && weekOrder != val.WeekOrder {
			continue
		}
		//if weekorder none and special and order equals
		if weekOrder+val.WeekOrder <= 0 && order == val.Order {
			return false, nil
		}
		// if weekorder equals
		if weekOrder == val.WeekOrder && val.Order == order {
			return false, nil
		}
	}

	return true, nil
}

func (s *ScheduleProvider) ValidateAddScheduleCreation(date time.Time, order int) (bool, error) {
	schedule, ok := s.additionalCache[date.Format("2006-01-02")]

	if !ok {
		return true, nil
	}

	if schedule == nil {
		return true, nil
	}

	for _, val := range schedule {

		// if order equals
		if val.Order == order {
			return false, nil
		}
	}

	return true, nil
}

func (s *ScheduleProvider) GetCommonSchedule() map[time.Weekday][]dao.ScheduleModel {
	return s.scheduleCache
}
