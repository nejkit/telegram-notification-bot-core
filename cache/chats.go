package cache

import (
	"fmt"
	"strconv"
	"telegram-notification-bot-core/exceptions"
)

type ChatCacheProvider struct {
	common *CommonProvider
}

const (
	userToChatKeysPattern = "user-chat:%d"
)

func NewChatCacheProvider(common *CommonProvider) *ChatCacheProvider {
	return &ChatCacheProvider{common: common}
}

func (c *ChatCacheProvider) SaveChatIdForUserId(userId int, chatId int64) error {
	return c.common.saveKeyValue(fmt.Sprintf(userToChatKeysPattern, userId), chatId, 0)
}

func (c *ChatCacheProvider) GetChatIdByUserId(userId int) (int64, error) {
	value, err := c.common.getValueByKey(fmt.Sprintf(userToChatKeysPattern, userId))

	if err != nil {
		return 0, err
	}

	converted, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return 0, exceptions.InternalError
	}

	return converted, nil
}
