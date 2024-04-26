package services

import (
	"context"
	"errors"
	"telegram-notification-bot-core/abstractions"
	"telegram-notification-bot-core/configuration"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/util"
	"time"
)

type HandleFunc func(scheduleDto dto.ScheduleDto, startTime time.Time, recipients []int64)

type BackgroundService struct {
	scheduleService abstractions.IScheduleService
	chatProvider    abstractions.IChatProvider
	cfg             configuration.Configuration
	handlers        map[int]<-chan struct{}
}

func NewBackgroundService(scheduleService abstractions.IScheduleService, chatProvider abstractions.IChatProvider, cfg configuration.Configuration) *BackgroundService {
	return &BackgroundService{scheduleService: scheduleService, chatProvider: chatProvider, cfg: cfg, handlers: map[int]<-chan struct{}{}}
}

func (b BackgroundService) Run(ctx context.Context, handleFunc HandleFunc) {
	ticker := time.NewTicker(b.cfg.ScheduleSettings.ScheduleRefreshInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			schedules, err := b.collectDataForCreationHandlers()

			if err != nil {
				continue
			}

			for _, schedule := range schedules {
				_, ok := b.handlers[schedule.Order]

				if ok {
					continue
				}

				b.handlers[schedule.Order] = b.initHandler(ctx, handleFunc, schedule)

				schedule1 := schedule
				go func() {

					for {
						select {
						case <-b.handlers[schedule1.Order]:
							delete(b.handlers, schedule1.Order)
						default:
							time.Sleep(time.Second)
						}

					}

				}()
			}
		}
	}
}

func (b BackgroundService) collectDataForCreationHandlers() ([]dto.ScheduleDto, error) {
	schedules, err := b.scheduleService.GetCurrentSchedule()

	var filteredSchedules []dto.ScheduleDto

	if err != nil {
		return nil, err
	}

	actualTime := time.Now()

	for _, schedule := range schedules.Schedules {
		startTime := util.GetMidnightTime().Add(b.cfg.ScheduleSettings.TimeSlotsConfiguration[schedule.Order].StartTime)

		if actualTime.After(startTime) {
			continue
		}

		filteredSchedules = append(filteredSchedules, schedule)
	}

	if len(filteredSchedules) == 0 {
		return nil, errors.New("NotFound")
	}

	return filteredSchedules, nil
}

func (b BackgroundService) initHandler(
	ctx context.Context,
	handleFunc HandleFunc,
	args dto.ScheduleDto) chan struct{} {
	ticker := time.NewTicker(time.Second * 30)

	startTime := util.GetMidnightTime().Add(b.cfg.ScheduleSettings.TimeSlotsConfiguration[args.Order].StartTime)

	reminderSlice := append(b.cfg.ScheduleSettings.ReminderIntervals, 0)

	cancelChan := make(chan struct{})

	go func() {

		var recipients []int64

		for _, accId := range b.cfg.Security.AllowedAccountIds {
			chatId, err := b.chatProvider.GetChatByUserId(accId)

			if err != nil {
				continue
			}

			recipients = append(recipients, chatId)
		}

		for {
			select {
			case <-ctx.Done():
				cancelChan <- struct{}{}
				return
			case <-ticker.C:
				actualTime := time.Now()

				for _, rem := range reminderSlice {

					if int(startTime.Sub(actualTime).Minutes()) < rem {
						reminderSlice = reminderSlice[1:]
					}

					if int(startTime.Sub(actualTime).Minutes()) == rem {
						reminderSlice = reminderSlice[1:]

						handleFunc(args, startTime, recipients)
					}
				}
			default:
				if len(reminderSlice) == 0 {
					cancelChan <- struct{}{}
					return
				}
			}
		}
	}()

	return cancelChan
}
