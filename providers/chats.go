package providers

import (
	"encoding/json"
	"errors"
)

type ChatProvider struct {
	common *CommonProvider
	cache  map[int]int64
}

func NewChatProvider() *ChatProvider {
	common := newCommonProvider("chats")
	data, err := common.getAllDataFromStorage()

	cache := make(map[int]int64)

	if err == nil {
		err = json.Unmarshal(data, &cache)

		if err != nil {
			cache = make(map[int]int64)
		}
	}

	return &ChatProvider{common: common, cache: cache}
}

func (c *ChatProvider) GetChatByUserId(userId int) (int64, error) {

	data, ok := c.cache[userId]

	if !ok {
		return 0, errors.New("NotFound")
	}

	return data, nil
}

func (c *ChatProvider) SaveChatForUser(userId int, chatId int64) error {

	c.cache[userId] = chatId

	data, err := json.Marshal(c.cache)

	if err != nil {
		return err
	}

	return c.common.saveAllDataToStorage(data)
}
