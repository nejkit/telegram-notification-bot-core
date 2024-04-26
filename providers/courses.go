package providers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"sync"
	"telegram-notification-bot-core/dao"
	"telegram-notification-bot-core/exceptions"
)

type CourseProvider struct {
	common *CommonProvider
	cache  map[string]dao.CourseModel
	mutex  *sync.RWMutex
}

func NewCourseProvider() *CourseProvider {
	common := newCommonProvider("courses")

	data, err := common.getAllDataFromStorage()

	cache := make(map[string]dao.CourseModel)

	if err == nil {
		err = json.Unmarshal(data, &cache)

		if err != nil {
			cache = make(map[string]dao.CourseModel)
		}
	}

	return &CourseProvider{common: common, cache: cache, mutex: &sync.RWMutex{}}
}

func (c *CourseProvider) CreateNewCourse(model dao.CourseModel) (str string, err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	id := uuid.NewString()

	model.Id = id

	c.cache[id] = model

	defer func() {
		if err != nil {
			delete(c.cache, id)
		}
	}()

	data, err := json.Marshal(c.cache)

	if err != nil {

		return "", err
	}

	err = c.common.saveAllDataToStorage(data)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *CourseProvider) UpdateCourse(model dao.CourseModel) (err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	backup := c.cache[model.Id]

	c.cache[model.Id] = model
	defer func() {
		if err != nil {
			c.cache[backup.Id] = backup
		}
	}()

	data, err := json.Marshal(c.cache)

	if err != nil {
		return err
	}

	err = c.common.saveAllDataToStorage(data)

	if err != nil {
		return err
	}

	return nil
}

func (c *CourseProvider) ArchiveCourse(id string) (err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	backup := c.cache[id]

	delete(c.cache, id)

	defer func() {
		if err != nil {
			c.cache[id] = backup
		}
	}()

	data, err := json.Marshal(c.cache)

	if err != nil {
		return err
	}

	err = c.common.saveAllDataToStorage(data)

	if err != nil {
		return err
	}

	return nil
}

func (c *CourseProvider) GetCourseByParams(name string) (*dao.CourseModel, error) {
	for _, val := range c.cache {
		if val.Name == name {
			return &val, nil
		}
	}

	return nil, exceptions.NotFound
}

func (c *CourseProvider) GetCourses() ([]dao.CourseModel, error) {
	var result []dao.CourseModel

	for _, val := range c.cache {
		result = append(result, val)
	}

	return result, nil
}

func (c *CourseProvider) GetCourseById(id string) (*dao.CourseModel, error) {
	data, ok := c.cache[id]

	if !ok {
		return nil, errors.New("")
	}

	return &data, nil
}
