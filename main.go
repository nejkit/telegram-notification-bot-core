package main

import (
	"context"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"telegram-notification-bot-core/bot"
	"telegram-notification-bot-core/configuration"
	"telegram-notification-bot-core/providers"
	"telegram-notification-bot-core/services"
)

func main() {

	filename, _ := filepath.Abs("./configs/config.yml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var config configuration.Configuration

	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	childCtx, cancel := context.WithCancel(ctx)
	actionsProvider := providers.NewActionProvider()
	coursesProvider := providers.NewCourseProvider()
	schedulesProvider := providers.NewScheduleProvider()
	chatProvider := providers.NewChatProvider()

	actionsService := services.NewActionService(actionsProvider)
	coursesService := services.NewCourseService(coursesProvider)
	scheduleService := services.NewScheduleService(config, schedulesProvider, coursesProvider)
	backgroundService := services.NewBackgroundService(scheduleService, chatProvider, config)

	api, err := bot.NewApi(config)

	if err != nil {
		panic(err)
	}
	go backgroundService.Run(childCtx, api.SendNotification)
	handler := bot.NewHandler(coursesService, actionsService, scheduleService, chatProvider, config, api)
	go api.StartServe()
	go handler.Run(childCtx)

	exit := make(chan os.Signal, 1)
	for {
		select {
		case <-exit:
			{
				cancel()

			}
		}
	}
}
