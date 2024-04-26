package bot

import (
	"context"
	"fmt"
	"github.com/dipsycat/calendar-telegram-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"telegram-notification-bot-core/abstractions"
	"telegram-notification-bot-core/actions"
	"telegram-notification-bot-core/commands"
	"telegram-notification-bot-core/configuration"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/util"
	"time"
)

type Handler struct {
	course   abstractions.ICourseService
	schedule abstractions.IScheduleService
	actions  abstractions.IActionService
	chats    abstractions.IChatProvider
	cfg      configuration.Configuration

	createCourseRequests      map[int]dto.CreateNewCourseRequest
	createScheduleRequests    map[int]dto.CreateNewScheduleRequest
	createAddScheduleRequests map[int]dto.CreateNewAdditionalScheduleRequest
	updateCourseRequests      map[int]dto.UpdateCourseInfoRequest
	calendarPosition          map[int]dto.CalendarPositionDto

	api *Api
}

func NewHandler(
	course abstractions.ICourseService,
	actions abstractions.IActionService,
	schedules abstractions.IScheduleService,
	chats abstractions.IChatProvider,
	cfg configuration.Configuration, api *Api) *Handler {

	return &Handler{
		actions:                   actions,
		cfg:                       cfg,
		course:                    course,
		schedule:                  schedules,
		chats:                     chats,
		api:                       api,
		createCourseRequests:      map[int]dto.CreateNewCourseRequest{},
		updateCourseRequests:      map[int]dto.UpdateCourseInfoRequest{},
		calendarPosition:          map[int]dto.CalendarPositionDto{},
		createScheduleRequests:    map[int]dto.CreateNewScheduleRequest{},
		createAddScheduleRequests: map[int]dto.CreateNewAdditionalScheduleRequest{},
	}

}

func (h *Handler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-h.api.updatesChan:
			if !ok {
				return
			}

			if update.Message != nil {
				go func() {
					answer := h.handleMsg(update)
					for _, a := range answer {
						h.api.executeMessage(a)
					}
				}()
			}

			if update.CallbackQuery != nil {
				go func() {
					if handled := h.handleCalendarButtons(update); handled {
						return
					}

					answer := h.handleCallback(update)
					go h.api.executeCallback(answer)

					h.handlePreparingActionsAfterCallback(update)
				}()
			}
		default:
			time.Sleep(time.Second)
		}
	}
}

func (h *Handler) handleCallback(query tgbotapi.Update) tgbotapi.CallbackConfig {
	userId := query.CallbackQuery.From.ID

	action := h.actions.GetUserCurrentState(userId)

	switch action.Action {
	case actions.UserActionInputDate:
		h.handleInputDate(query, userId)
	case actions.UserActionChooseCourse:

		switch action.Command {
		case commands.CreateAdditionalScheduleCommand:
			h.handleChooseCourseForAdditionalSchedule(query)
		case commands.CreateScheduleCommand:
			h.handleChooseCourseForCreateSchedule(query)
		case commands.UpdateCourseCommand:
			h.handleChooseCourseForUpdate(query)
		case commands.DeleteCoursesCommand:
			h.handleChooseCourseForDelete(query)
		}
	}
	return tgbotapi.CallbackConfig{}
}

func (h *Handler) handleMsg(upd tgbotapi.Update) []tgbotapi.MessageConfig {

	h.chats.SaveChatForUser(upd.Message.From.ID, upd.Message.Chat.ID)
	userId := upd.Message.From.ID

	if authenticated := h.baseAuth(userId); !authenticated {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Ви не маєте прав, зверніться до власника")}
	}

	action := h.actions.GetUserCurrentState(userId)

	if upd.Message.IsCommand() && upd.Message.Command() == string(commands.CancelCommand) {
		return h.handleCommand(userId, upd)
	}

	if upd.Message.IsCommand() && action.Action != actions.UserActionNone {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Ви не закінчили минулу операцію")}
	}

	if upd.Message.IsCommand() && action.Action == actions.UserActionNone {
		return h.handleCommand(userId, upd)
	}

	return h.handleAction(action, userId, upd)
}

func (h *Handler) handleCalendarButtons(update tgbotapi.Update) bool {

	if update.CallbackQuery.Data == ">" {
		curSettings := h.calendarPosition[update.CallbackQuery.From.ID]
		calendarr, year, newMonth := calendar.HandlerNextButton(curSettings.Year, curSettings.Month)
		h.calendarPosition[update.CallbackQuery.From.ID] = dto.CalendarPositionDto{
			Month: newMonth,
			Year:  year,
		}
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введіть дату заміни")
		msg.ReplyMarkup = calendarr
		h.api.executeMessage(msg)
		return true
	}

	if update.CallbackQuery.Data == "<" {
		curSettings := h.calendarPosition[update.CallbackQuery.From.ID]
		calendarr, year, newMonth := calendar.HandlerPrevButton(curSettings.Year, curSettings.Month)
		h.calendarPosition[update.CallbackQuery.From.ID] = dto.CalendarPositionDto{
			Month: newMonth,
			Year:  year,
		}
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введіть дату заміни")
		msg.ReplyMarkup = calendarr
		h.api.executeMessage(msg)
		return true
	}

	return false
}

func (h *Handler) handlePreparingActionsAfterCallback(update tgbotapi.Update) {
	action := h.actions.GetUserCurrentState(update.CallbackQuery.From.ID)

	if action.Action == actions.UserActionInputCourseName && action.Command == commands.UpdateCourseCommand {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Введіть ім'я курсу")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Без змін")))
		go h.api.executeMessage(msg)
	}

	if action.Action == actions.UserActionInputWeekday && action.Command == commands.CreateScheduleCommand {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Виберіть день тижня")
		markup := tgbotapi.NewReplyKeyboard()

		for _, days := range []time.Weekday{0, 1, 2, 3, 4, 5, 6} {
			markup.Keyboard = append(markup.Keyboard,
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(util.ConvertToHumanReadableWeek(days))))
		}

		msg.ReplyMarkup = markup
		go h.api.executeMessage(msg)
	}

	if action.Action == actions.UserActionInputOrder && action.Command == commands.CreateAdditionalScheduleCommand {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Виберіть, коли буде пара")
		markup := tgbotapi.NewReplyKeyboard()

		for orders, _ := range h.cfg.ScheduleSettings.TimeSlotsConfiguration {
			markup.Keyboard = append(markup.Keyboard,
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("%d", orders))))
		}

		msg.ReplyMarkup = markup
		go h.api.executeMessage(msg)
	}
}

func (h *Handler) handleInputDate(query tgbotapi.Update, userId int) tgbotapi.CallbackConfig {
	req := h.createAddScheduleRequests[userId]

	date, err := time.Parse("2006.01.02", query.CallbackQuery.Data)

	if err != nil {
		return tgbotapi.CallbackConfig{
			CallbackQueryID: query.CallbackQuery.ID,
			Text:            "Помилка серверу",
		}
	}

	req.Date = date

	delete(h.createAddScheduleRequests, userId)

	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{Action: actions.UserActionNone})

	err = h.schedule.InsertAdditionalSchedule(req)

	if err != nil {
		return tgbotapi.CallbackConfig{
			CallbackQueryID: query.CallbackQuery.ID,
			Text:            "Помилка під час виконання запиту",
		}
	}

	return tgbotapi.CallbackConfig{
		CallbackQueryID: query.CallbackQuery.ID,
		Text:            "Заміну успішно додано",
	}
}

func (h *Handler) handleChooseCourseForAdditionalSchedule(query tgbotapi.Update) tgbotapi.CallbackConfig {
	req := dto.CreateNewAdditionalScheduleRequest{
		CreateNewScheduleRequest: dto.CreateNewScheduleRequest{
			CourseId: query.CallbackQuery.Data,
		},
	}
	h.createAddScheduleRequests[query.CallbackQuery.From.ID] = req
	h.actions.SaveUserCurrentState(query.CallbackQuery.From.ID, dto.UserActionDto{
		Command: commands.CreateAdditionalScheduleCommand,
		Action:  actions.UserActionInputOrder,
	})

	return tgbotapi.CallbackConfig{
		CallbackQueryID: query.CallbackQuery.ID,
		Text:            "Предмет обрано",
	}
}

func (h *Handler) handleChooseCourseForCreateSchedule(query tgbotapi.Update) tgbotapi.CallbackConfig {
	req := dto.CreateNewScheduleRequest{
		CourseId: query.CallbackQuery.Data,
	}
	h.createScheduleRequests[query.CallbackQuery.From.ID] = req
	h.actions.SaveUserCurrentState(query.CallbackQuery.From.ID, dto.UserActionDto{
		Command: commands.CreateScheduleCommand,
		Action:  actions.UserActionInputWeekday,
	})

	return tgbotapi.CallbackConfig{
		CallbackQueryID: query.CallbackQuery.ID,
		Text:            "Курс вибрано!",
	}
}

func (h *Handler) handleChooseCourseForUpdate(query tgbotapi.Update) tgbotapi.CallbackConfig {
	req := h.updateCourseRequests[query.CallbackQuery.From.ID]
	req.Id = query.CallbackQuery.Data

	info, _ := h.course.GetCourseById(req.Id)
	req.Name = info.Name
	req.MeetLink = info.MeetLink
	req.TeacherName = info.TeacherName
	req.TeacherContact = info.TeacherContact

	h.updateCourseRequests[query.CallbackQuery.From.ID] = req

	h.actions.SaveUserCurrentState(query.CallbackQuery.From.ID, dto.UserActionDto{
		Command: commands.UpdateCourseCommand,
		Action:  actions.UserActionInputCourseName,
	})

	return tgbotapi.CallbackConfig{
		CallbackQueryID: query.CallbackQuery.ID,
		Text:            "Курс для оновлення обрано!",
	}
}

func (h *Handler) handleChooseCourseForDelete(query tgbotapi.Update) tgbotapi.CallbackConfig {
	h.actions.SaveUserCurrentState(query.CallbackQuery.From.ID, dto.UserActionDto{
		Action: actions.UserActionNone,
	})

	h.course.DeleteCourse(dto.ArchiveCourseRequest{
		CourseId: query.CallbackQuery.Data,
	})

	return tgbotapi.CallbackConfig{
		CallbackQueryID: query.CallbackQuery.ID,
		Text:            "Курс видалено",
	}
}

func (h *Handler) baseAuth(userId int) bool {
	for _, users := range h.cfg.Security.AllowedAccountIds {
		if users == userId {
			return true
		}
	}
	return false
}

func (h *Handler) adminAuth(userId int) bool {
	for _, users := range h.cfg.Security.TrustedAccountIds {
		if users == userId {
			return true
		}
	}
	return false
}

func (h *Handler) handleCommandCreateAdditionalSchedule(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {

	if authenticated := h.adminAuth(userId); !authenticated {
		chat, _ := h.chats.GetChatByUserId(userId)
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(chat, "Ви не маєте прав, зверніться до власника")}
	}

	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Command: commands.CreateAdditionalScheduleCommand,
		Action:  actions.UserActionChooseCourse,
	})

	h.createAddScheduleRequests[userId] = dto.CreateNewAdditionalScheduleRequest{}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Оберіть предмет")
	keys := tgbotapi.NewInlineKeyboardMarkup()
	courses, _ := h.course.GetCourses()

	for _, course := range courses.Courses {
		keys.InlineKeyboard = append(keys.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(course.Name, course.Id)))
	}

	msg.ReplyMarkup = keys
	return []tgbotapi.MessageConfig{msg}
}

func (h *Handler) handleGetCommonSchedules(upd tgbotapi.Update) []tgbotapi.MessageConfig {
	schedules := h.schedule.GetCommonSchedule()

	var res []tgbotapi.MessageConfig
	text := "Розклад"
	res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
	text = ""
	for week, val := range schedules.Schedules {
		patchedTxt := util.ConvertToHumanReadableWeek(week) + "\n"

		for order, value := range val.OrderToSchedules {
			for _, v := range value {
				patchedTxt += fmt.Sprintf("№ %d. %s \n Вчитель: %s \n Контакт: %s \n Тиждень: %s \n Посилання на зустріч: %s \n",
					order, v.CourseInfo.Name, v.CourseInfo.TeacherName, v.CourseInfo.TeacherContact, util.ConvertToHumanReadableWeekOrder(v.WeekOrder), v.CourseInfo.MeetLink)
			}

			patchedTxt += "\n"
		}

		if len(text)+len(patchedTxt) > 4096 {
			res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
			text = ""
		}
		text += patchedTxt

		if text != "" {
			res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
			text = ""
		}
	}

	return res
}

func (h *Handler) handleGetScheduleAtToday(upd tgbotapi.Update) []tgbotapi.MessageConfig {
	schedules, _ := h.schedule.GetCurrentSchedule()
	var res []tgbotapi.MessageConfig
	text := "Розклад. Дата: " + schedules.CurrentDate.Format("2006-01-02") + " Тиждень: " + util.ConvertToHumanReadableWeekOrder(schedules.CurrentWeekOrder)
	res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
	text = ""
	for _, val := range schedules.Schedules {
		patchedTxt := fmt.Sprintf("№ %d. %s \n Вчитель: %s \n Контакт: %s \n Посилання на зустріч: %s \n",
			val.Order, val.CourseInfo.Name, val.CourseInfo.TeacherName, val.CourseInfo.TeacherContact, val.CourseInfo.MeetLink)
		if len(text)+len(patchedTxt) > 4096 {
			res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
			text = ""
		}
		text += patchedTxt
	}

	if text != "" {
		res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
	}

	return res
}

func (h *Handler) handleCommandCreateCourse(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {

	if authenticated := h.adminAuth(userId); !authenticated {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Ви не маєте прав, зверніться до власника")}
	}

	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Command: commands.CreateCourseCommand,
		Action:  actions.UserActionInputCourseName,
	})
	h.createCourseRequests[userId] = dto.CreateNewCourseRequest{}

	return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть назву предмета")}

}

func (h *Handler) handleCommandUpdateCourse(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Command: commands.UpdateCourseCommand,
		Action:  actions.UserActionChooseCourse,
	})
	h.updateCourseRequests[userId] = dto.UpdateCourseInfoRequest{}
	infos, _ := h.course.GetCourses()

	keys := tgbotapi.NewInlineKeyboardMarkup()

	for _, val := range infos.Courses {
		keys.InlineKeyboard = append(keys.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(val.Name, val.Id)))
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Оберіть предмет")
	msg.ReplyMarkup = keys
	return []tgbotapi.MessageConfig{msg}
}

func (h *Handler) handleGetCoursesCommand(upd tgbotapi.Update) []tgbotapi.MessageConfig {
	courses, _ := h.course.GetCourses()

	var res []tgbotapi.MessageConfig
	text := "Список наявних курсів"
	res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
	text = ""
	for k, val := range courses.Courses {
		patchedTxt := fmt.Sprintf("\n %d. %s. \n Вчитель: %s \n його контакт: %s \n Посилання на зустріч: %s \n",
			k, val.Name, val.TeacherName, val.TeacherContact, val.MeetLink)
		if len(text)+len(patchedTxt) > 4096 {
			res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
			text = ""
		}
		text += patchedTxt
	}

	if text != "" {
		res = append(res, tgbotapi.NewMessage(upd.Message.Chat.ID, text))
	}

	return res
}

func (h *Handler) handleCommandDeleteCourse(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {

	if authenticated := h.adminAuth(userId); !authenticated {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Ви не маєте прав, зверніться до власника")}
	}

	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Command: commands.DeleteCoursesCommand,
		Action:  actions.UserActionChooseCourse,
	})

	infos, _ := h.course.GetCourses()

	keys := tgbotapi.NewInlineKeyboardMarkup()

	for _, val := range infos.Courses {
		keys.InlineKeyboard = append(keys.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(val.Name, val.Id)))
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Оберіть предмет")
	msg.ReplyMarkup = keys
	return []tgbotapi.MessageConfig{msg}
}

func (h *Handler) handleCommandCreateSchedule(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {

	if authenticated := h.adminAuth(userId); !authenticated {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Ви не маєте прав, зверніться до власника")}
	}

	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Action:  actions.UserActionChooseCourse,
		Command: commands.CreateScheduleCommand,
	})

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Оберіть предмет")
	keys := tgbotapi.NewInlineKeyboardMarkup()
	courses, _ := h.course.GetCourses()

	for _, course := range courses.Courses {
		keys.InlineKeyboard = append(keys.InlineKeyboard,
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(course.Name, course.Id)))
	}

	msg.ReplyMarkup = keys
	return []tgbotapi.MessageConfig{msg}
}

func (h *Handler) handleCancelCommand(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {

	delete(h.createCourseRequests, userId)
	delete(h.updateCourseRequests, userId)
	delete(h.createScheduleRequests, userId)
	delete(h.createAddScheduleRequests, userId)
	delete(h.calendarPosition, userId)

	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Action: actions.UserActionNone,
	})

	return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Усі дії скасовано")}
}

func (h *Handler) handleCommand(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	switch upd.Message.Command() {
	case string(commands.CreateAdditionalScheduleCommand):
		return h.handleCommandCreateAdditionalSchedule(userId, upd)
	case string(commands.GetScheduleCommonCommand):
		return h.handleGetCommonSchedules(upd)
	case string(commands.GetScheduleTodayCommand):
		return h.handleGetScheduleAtToday(upd)
	case string(commands.CreateCourseCommand):
		return h.handleCommandCreateCourse(userId, upd)
	case string(commands.UpdateCourseCommand):
		return h.handleCommandUpdateCourse(userId, upd)
	case string(commands.GetCoursesCommand):
		return h.handleGetCoursesCommand(upd)
	case string(commands.DeleteCoursesCommand):
		return h.handleCommandDeleteCourse(userId, upd)
	case string(commands.CreateScheduleCommand):
		return h.handleCommandCreateSchedule(userId, upd)
	case string(commands.CancelCommand):
		return h.handleCancelCommand(userId, upd)
	default:
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Невідома команда")}
	}
}

func (h *Handler) handleActionInputCourseName(action dto.UserActionDto, userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	if action.Command == commands.CreateCourseCommand {
		req, _ := h.createCourseRequests[userId]
		req.Name = upd.Message.Text
		h.createCourseRequests[userId] = req
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть ім'я вчителя")}
	}
	if action.Command == commands.UpdateCourseCommand {
		if upd.Message.Text != "Без змін" {
			req, _ := h.updateCourseRequests[userId]
			req.Name = upd.Message.Text
			h.updateCourseRequests[userId] = req
		}
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть ім'я вчителя")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Без змін")))
		return []tgbotapi.MessageConfig{msg}
	}
	panic("")
}

func (h *Handler) handleActionInputTeacherName(action dto.UserActionDto, userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	if action.Command == commands.CreateCourseCommand {
		req, _ := h.createCourseRequests[userId]
		req.TeacherName = upd.Message.Text
		h.createCourseRequests[userId] = req

		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть контакт вчителя")}
	}
	if action.Command == commands.UpdateCourseCommand {
		if upd.Message.Text != "Без змін" {
			req, _ := h.updateCourseRequests[userId]
			req.TeacherName = upd.Message.Text
			h.updateCourseRequests[userId] = req
		}
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть контакт вчителя")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Без змін")))
		return []tgbotapi.MessageConfig{msg}
	}
	panic("")
}

func (h *Handler) handleActionInputTeacherContact(action dto.UserActionDto, userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	if action.Command == commands.CreateCourseCommand {
		req, _ := h.createCourseRequests[userId]
		req.TeacherContact = upd.Message.Text
		h.createCourseRequests[userId] = req

		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть посилання на зустріч")}
	}
	if action.Command == commands.UpdateCourseCommand {
		if upd.Message.Text != "Без змін" {
			req, _ := h.updateCourseRequests[userId]
			req.TeacherContact = upd.Message.Text
			h.updateCourseRequests[userId] = req
		}
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть посилання на зустріч")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Без змін")))
		return []tgbotapi.MessageConfig{msg}
	}
	panic("")
}

func (h *Handler) handleActionInputMeetLink(action dto.UserActionDto, userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	if action.Command == commands.CreateCourseCommand {
		req, _ := h.createCourseRequests[userId]
		req.MeetLink = upd.Message.Text
		delete(h.createCourseRequests, userId)
		_, err := h.course.CreateNewCourse(req)

		if err != nil {
			return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Під час виконання запиту трапилась помилка "+err.Error())}
		}

		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Курс було створено")}
	}
	if action.Command == commands.UpdateCourseCommand {
		req, _ := h.updateCourseRequests[userId]
		if upd.Message.Text != "Без змін" {
			req.MeetLink = upd.Message.Text
		}
		h.course.UpdateCourse(req)
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Курс було оновлено")
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
		return []tgbotapi.MessageConfig{msg}
	}
	panic("")
}

func (h *Handler) handleActionInputWeekDay(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	req, _ := h.createScheduleRequests[userId]
	weekday, err := util.ConvertFromHumanReadableWeek(upd.Message.Text)

	if err != nil {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Невірні дані, спробуйте ще раз")}
	}

	req.Weekday = weekday
	h.createScheduleRequests[userId] = req
	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Command: commands.CreateScheduleCommand,
		Action:  actions.UserActionInputWeekOrder,
	})
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть, на якому тижні буде заняття")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Верхній"), tgbotapi.NewKeyboardButton("Нижній"), tgbotapi.NewKeyboardButton("Статичний")))

	return []tgbotapi.MessageConfig{msg}
}

func (h *Handler) handleActionInputWeekOrder(userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	req, _ := h.createScheduleRequests[userId]
	weekOrder, err := util.ConvertFromHumanReadableOrderWeek(upd.Message.Text)

	if err != nil {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Невірні дані, спробуйте ще раз")}
	}

	req.WeekOrder = weekOrder
	h.createScheduleRequests[userId] = req
	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Command: commands.CreateScheduleCommand,
		Action:  actions.UserActionInputOrder,
	})
	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть, на якій парі буде заняття")
	markup := tgbotapi.NewReplyKeyboard()

	for orders, _ := range h.cfg.ScheduleSettings.TimeSlotsConfiguration {

		markup.Keyboard = append(markup.Keyboard,
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("%d", orders))))
	}
	msg.ReplyMarkup = markup
	return []tgbotapi.MessageConfig{msg}
}

func (h *Handler) handleActionInputOrder(action dto.UserActionDto, userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {

	if valid := h.validateOrderInput(upd.Message.Text); !valid {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Невірні дані, повторіть спробу")}
	}

	converted, err := strconv.ParseInt(upd.Message.Text, 10, 32)

	if err != nil {
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Невірні дані, повторіть спробу")}
	}

	if action.Command == commands.CreateAdditionalScheduleCommand {
		req, _ := h.createAddScheduleRequests[userId]
		req.Order = int(converted)

		h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
			Command: commands.CreateAdditionalScheduleCommand,
			Action:  actions.UserActionInputDate,
		})

		h.createAddScheduleRequests[userId] = req

		cleanMarkup := tgbotapi.NewMessage(upd.Message.Chat.ID, "Дата заміни")
		cleanMarkup.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}

		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Введіть час, коли відбудється заміна")
		h.calendarPosition[userId] = dto.CalendarPositionDto{Month: time.Now().Month(), Year: time.Now().Year()}
		markup := calendar.GenerateCalendar(time.Now().Year(), time.Now().Month())

		msg.ReplyMarkup = markup
		return []tgbotapi.MessageConfig{cleanMarkup, msg}
	}

	req, _ := h.createScheduleRequests[userId]

	req.Order = int(converted)

	delete(h.createScheduleRequests, userId)
	h.actions.SaveUserCurrentState(userId, dto.UserActionDto{
		Action: actions.UserActionNone,
	})

	err = h.schedule.CreateNewSchedule(req)

	if err != nil {
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Виникла помилка під час збереження")
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
		return []tgbotapi.MessageConfig{msg}
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Пару було збережено")
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
	return []tgbotapi.MessageConfig{msg}
}

func (h *Handler) validateOrderInput(data string) bool {
	for orders, _ := range h.cfg.ScheduleSettings.TimeSlotsConfiguration {
		if fmt.Sprintf("%d", orders) == data {
			return true
		}
	}
	return false
}

func (h *Handler) handleAction(action dto.UserActionDto, userId int, upd tgbotapi.Update) []tgbotapi.MessageConfig {
	switch action.Action {
	case actions.UserActionInputCourseName:
		defer h.actions.SaveUserCurrentState(userId, dto.UserActionDto{Command: action.Command, Action: actions.UserActionInputTeacherName})
		return h.handleActionInputCourseName(action, userId, upd)
	case actions.UserActionInputTeacherName:
		defer h.actions.SaveUserCurrentState(userId, dto.UserActionDto{Command: action.Command, Action: actions.UserActionInputTeacherContact})
		return h.handleActionInputTeacherName(action, userId, upd)
	case actions.UserActionInputTeacherContact:
		defer h.actions.SaveUserCurrentState(userId, dto.UserActionDto{Command: action.Command, Action: actions.UserActionInputMeetLink})
		return h.handleActionInputTeacherContact(action, userId, upd)
	case actions.UserActionInputMeetLink:
		defer h.actions.SaveUserCurrentState(userId, dto.UserActionDto{Action: actions.UserActionNone})
		return h.handleActionInputMeetLink(action, userId, upd)
	case actions.UserActionInputWeekday:
		return h.handleActionInputWeekDay(userId, upd)
	case actions.UserActionInputWeekOrder:
		return h.handleActionInputWeekOrder(userId, upd)
	case actions.UserActionInputOrder:
		return h.handleActionInputOrder(action, userId, upd)
	default:
		return []tgbotapi.MessageConfig{tgbotapi.NewMessage(upd.Message.Chat.ID, "Виникла помилка, повторіть спробу пізніше")}
	}
}
