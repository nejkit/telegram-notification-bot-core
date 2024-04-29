package util

import (
	"errors"
	"time"
)

type WeekOrder int

const (
	WeekOrderNone  WeekOrder = -2
	WeekOrderUpper WeekOrder = 1
	WeekOrderDown  WeekOrder = 2
)

func GetCurrentWeekOrder() WeekOrder {
	actualTime := time.Now()

	actualWeekCount := actualTime.Day() / 7

	if actualWeekCount%2 == 0 {
		return 1
	}
	return 2
}

func ConvertFromHumanReadableOrderWeek(data string) (WeekOrder, error) {
	values := map[string]WeekOrder{
		"Верхній":   WeekOrderUpper,
		"Нижній":    WeekOrderDown,
		"Статичний": WeekOrderNone,
	}
	converted, ok := values[data]

	if !ok {
		return 0, errors.New("InvalidWeekOrder")
	}

	return converted, nil

}

func ConvertFromHumanReadableWeek(data string) (time.Weekday, error) {

	values := map[string]time.Weekday{
		"Понеділок": time.Monday,
		"Вівторок":  time.Tuesday,
		"Середа":    time.Wednesday,
		"Четвер":    time.Thursday,
		"П'ятниця":  time.Friday,
		"Субота":    time.Saturday,
		"Неділя":    time.Sunday,
	}

	converted, ok := values[data]

	if !ok {
		return 0, errors.New("InvalidDayOfWeek")
	}

	return converted, nil

}

func ConvertToHumanReadableWeekOrder(weekOrder WeekOrder) string {
	switch weekOrder {
	case WeekOrderDown:
		return "Нижній"
	case WeekOrderUpper:
		return "Верхній"
	case WeekOrderNone:
		return "Статичний"
	default:
		return ""
	}
}

func ConvertToHumanReadableWeek(weekday time.Weekday) string {
	switch weekday {
	case time.Sunday:
		return "Неділя"
	case time.Monday:
		return "Понеділок"
	case time.Tuesday:
		return "Вівторок"
	case time.Wednesday:
		return "Середа"
	case time.Thursday:
		return "Четвер"
	case time.Friday:
		return "П'ятниця"
	case time.Saturday:
		return "Субота"
	default:
		return ""
	}
}

func GetMidnightTime() time.Time {
	currentTime := time.Now()

	// Get the time of midnight for the current day
	midnight := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())

	return midnight
}
