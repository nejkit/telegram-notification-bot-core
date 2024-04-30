package services

import (
	"context"
	"telegram-notification-bot-core/abstractions"
	"telegram-notification-bot-core/configuration"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/util"
	"time"
)

type HandleFunc func(scheduleDto dto.ScheduleDto, startTime time.Time, recipient int64)

type BackgroundService struct {
	scheduleService abstractions.IScheduleService
	chatProvider    abstractions.IChatProvider
	cfg             configuration.Configuration
	handlers        map[int]map[int]<-chan struct{}

	notificationCache abstractions.INotificationCacheProvider
}

func NewBackgroundService(scheduleService abstractions.IScheduleService, chatProvider abstractions.IChatProvider, cfg configuration.Configuration) *BackgroundService {
	return &BackgroundService{scheduleService: scheduleService, chatProvider: chatProvider, cfg: cfg, handlers: map[int]map[int]<-chan struct{}{}}
}

func (b BackgroundService) Run(ctx context.Context, handleFunc HandleFunc) {
	ticker := time.NewTicker(b.cfg.ScheduleSettings.ScheduleRefreshInterval)
	accounts := b.cfg.Security.AllowedAccountIds
	schedules, err := b.scheduleService.PrepareSchedulesListForNotify(accounts)

	if err == nil {
		for _, accountId := range accounts {
			b.handlers[accountId] = map[int]<-chan struct{}{}
			b.doCycle(ctx, schedules, accountId, handleFunc)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			accounts = b.cfg.Security.AllowedAccountIds

			schedules, err = b.scheduleService.PrepareSchedulesListForNotify(accounts)

			if err != nil {
				continue
			}

			for _, accountId := range accounts {

				b.doCycle(ctx, schedules[accountId], accountId, handleFunc)
			}
		}
	}
}

func (b BackgroundService) doCycle(
	ctx context.Context,
	schedules []dto.ScheduleDto,
	accountId int,
	handleFunc HandleFunc) {
	filteredNotifications := b.filterOverdueNotifications(schedules)

	for _, notification := range filteredNotifications {

		notify, err := b.notificationCache.GetInfoAboutNotification(accountId, notification.Order)

		if err == nil {
			continue
		}

		err = b.notificationCache.GetInfoAboutNotificationLock(accountId, notification.Order)

		if err == nil {
			continue
		}

		err = b.notificationCache.SaveInfosByNotification(accountId, dto.NotificationInfoDto{})

		b.handlers[accountId][notification.Order] = b.initHandler(ctx, handleFunc, notification, accountId)

		notification := notification
		go func() {
			for {
				select {
				case <-b.handlers[accountId][notification.Order]:
					delete(b.handlers[accountId], notification.Order)
				default:
					time.Sleep(time.Second)
				}
			}
		}()
	}
}

func (b BackgroundService) filterOverdueNotifications(scheduleListDto []dto.ScheduleDto) []dto.ScheduleDto {
	actualTime := time.Now()

	var filteredSchedules []dto.ScheduleDto

	for _, schedule := range scheduleListDto {
		startTime := util.GetMidnightTime().Add(b.cfg.ScheduleSettings.TimeSlotsConfiguration[schedule.Order].StartTime)

		if actualTime.After(startTime) {
			continue
		}

		filteredSchedules = append(filteredSchedules, schedule)
	}

	return filteredSchedules
}

func (b BackgroundService) initHandler(
	ctx context.Context,
	handleFunc HandleFunc,
	args dto.ScheduleDto,
	accountId int) chan struct{} {
	ticker := time.NewTicker(time.Second)
	startTime := util.GetMidnightTime().Add(b.cfg.ScheduleSettings.TimeSlotsConfiguration[args.Order].StartTime)
	reminderSlice := append(b.cfg.ScheduleSettings.ReminderIntervals, 0)

	cancelChan := make(chan struct{})

	go func() {

		chatId, err := b.chatProvider.GetChatByUserId(accountId)

		if err != nil {
			return
		}

		notificationDto := dto.NotificationInfoDto{
			NotificationDate: startTime,
			ChatId:           chatId,
			Data:             args,
			States:           map[int]bool{},
		}

		for _, interval := range reminderSlice {
			notificationDto.States[interval] = false
		}

		err = b.notificationCache.GetInfoAboutNotificationLock(accountId, args.Order)

		if err != nil {
			return
		}

		err = b.notificationCache.SaveInfosByNotification(accountId, notificationDto)

		if err != nil {
			return
		}

		for {
			select {
			case <-ctx.Done():
				cancelChan <- struct{}{}
				return
			case <-ticker.C:
				actualTime := time.Now()

				notificationInfo := b.notificationCache.GetInfoAboutNotification(accountId)

				for _, rem := range reminderSlice {

					if startTime.Minute()-actualTime.Minute() < rem {
						reminderSlice = reminderSlice[1:]
					}

					if startTime.Minute()-actualTime.Minute() == rem {
						reminderSlice = reminderSlice[1:]

						handleFunc(args, startTime, chatId)
					}
				}
			default:
				if len(reminderSlice) == 0 {
					cancelChan <- struct{}{}
					return
				}
				time.Sleep(time.Millisecond)
			}
		}
	}()

	return cancelChan
}
