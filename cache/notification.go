package cache

import (
	"encoding/json"
	"fmt"
	"telegram-notification-bot-core/dto"
	"telegram-notification-bot-core/exceptions"
	"time"
)

type NotificationCacheProvider struct {
	common *CommonProvider
}

const (
	userNotificationHash         = "user:%d:notifications"
	lockedNotificationKeyPattern = "user:%d:notification:%d:lock"
)

func NewNotificationCacheProvider(common *CommonProvider) *NotificationCacheProvider {
	return &NotificationCacheProvider{common: common}
}

func (n *NotificationCacheProvider) SaveInfosByNotification(userId int, object dto.NotificationInfoDto) error {
	data, err := json.Marshal(object)

	if err != nil {
		return err
	}

	return n.common.saveIntoHashNX(
		fmt.Sprintf(userNotificationHash, userId),
		fmt.Sprint(object.Data.Order),
		string(data))
}

func (n *NotificationCacheProvider) GetInfoAboutNotification(userId int, order int) (*dto.NotificationInfoDto, error) {
	data, err := n.common.getFromHash(fmt.Sprintf(userNotificationHash, userId), fmt.Sprint(order))

	if err != nil {
		return nil, err
	}

	var result dto.NotificationInfoDto

	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return nil, exceptions.InternalError
	}

	return &result, nil
}

func (n *NotificationCacheProvider) DeleteNotification(userId int, order int) error {
	return n.common.removeFromHash(
		fmt.Sprintf(userNotificationHash, userId), fmt.Sprint(order))
}

func (n *NotificationCacheProvider) SaveLockForCompletedNotification(userId int, order int) error {
	return n.common.saveKeyValue(
		fmt.Sprintf(lockedNotificationKeyPattern, userId, order),
		"bebra",
		time.Minute*2)
}

func (n *NotificationCacheProvider) GetInfoAboutNotificationLock(userId int, order int) error {
	_, err := n.common.getValueByKey(fmt.Sprintf(lockedNotificationKeyPattern, userId, order))

	if err == exceptions.NotFound {
		return nil
	}

	return err
}
