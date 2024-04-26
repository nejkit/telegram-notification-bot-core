package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"telegram-notification-bot-core/configuration"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/util"
	"time"
)

type Api struct {
	client      *tgbotapi.BotAPI
	cfg         configuration.Configuration
	updatesChan tgbotapi.UpdatesChannel
}

func NewApi(cfg configuration.Configuration) (*Api, error) {
	client, err := tgbotapi.NewBotAPI(cfg.TelegramTokenBot)

	if err != nil {
		return nil, err
	}

	return &Api{client: client, cfg: cfg}, nil
}

func (a *Api) BatchSend(scheduleDto dto.ScheduleDto, startTime time.Time, recipients []int64) {
	for _, rec := range recipients {
		msg := tgbotapi.NewMessage(rec, fmt.Sprintf(
			"Пара № %d, тиждень: %s, %s \n Вчитель: %s \n Контакт: %s \n Посилання на зустріч: %s \n Час зустрічі: %s",
			scheduleDto.Order,
			util.ConvertToHumanReadableWeekOrder(scheduleDto.WeekOrder),
			scheduleDto.CourseInfo.Name,
			scheduleDto.CourseInfo.TeacherName,
			scheduleDto.CourseInfo.TeacherContact,
			scheduleDto.CourseInfo.MeetLink,
			startTime.Format(time.DateTime)))
		go a.executeMessage(msg)
	}

}

func (a *Api) StartServe() {
	upd, err := a.client.GetUpdatesChan(tgbotapi.NewUpdate(0))

	if err != nil {
		return
	}

	a.updatesChan = upd
}

func (a *Api) executeCallback(config tgbotapi.CallbackConfig) {
	a.client.AnswerCallbackQuery(config)
}

func (a *Api) executeMessage(config tgbotapi.MessageConfig) {
	a.client.Send(config)
}
