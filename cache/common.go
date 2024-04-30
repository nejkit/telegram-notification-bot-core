package cache

import (
	"github.com/redis/go-redis"
	"github.com/sirupsen/logrus"
	"telegram-notification-bot-core/exceptions"
	"time"
)

type CommonProvider struct {
	client *redis.Client
}

func NewCommonProvider(conString string) (*CommonProvider, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     conString,
		Password: "",
		DB:       0,
	})

	err := client.Ping().Err()

	if err != nil {
		return nil, err
	}

	return &CommonProvider{client: client}, nil
}

func (c *CommonProvider) saveIntoHashNX(hashKey string, hashValue string, value interface{}) error {
	ex, err := c.client.HSetNX(hashKey, hashValue, value).Result()

	if !ex {
		return exceptions.AlreadyExists
	}

	if err != nil {
		logrus.Errorln("Failed save into hash: ", err.Error())
		return exceptions.InternalError
	}

	return nil
}

func (c *CommonProvider) removeFromHash(hashKey string, hashValue string) error {
	err := c.client.HDel(hashKey, hashValue).Err()

	if err != nil {
		logrus.Errorln("Failed remove from hash: ", err.Error())
		return exceptions.InternalError
	}

	return nil
}

func (c *CommonProvider) getFromHash(hashKey string, hashValue string) (string, error) {
	value, err := c.client.HGet(hashKey, hashValue).Result()

	if err == redis.Nil {
		return "", exceptions.NotFound
	}

	if err != nil {
		logrus.Error("Failed get from hash: ", err.Error())
		return "", exceptions.InternalError
	}

	return value, nil
}

func (c *CommonProvider) saveKeyValue(key string, value interface{}, expire time.Duration) error {
	err := c.client.Set(key, value, expire).Err()

	if err != nil {
		return exceptions.InternalError
	}

	return nil
}

func (c *CommonProvider) getValueByKey(key string) (string, error) {
	value, err := c.client.Get(key).Result()

	if err == redis.Nil {
		return "", exceptions.NotFound
	}

	if err != nil {
		return "", exceptions.InternalError
	}

	return value, nil
}
