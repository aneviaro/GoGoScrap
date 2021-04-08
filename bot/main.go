package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	updatehandler "gogoscrap/bot/update-handler"
	"gogoscrap/repository/storage"
	"gogoscrap/usecase/bot-service"
	"gogoscrap/usecase/user-config"
	"log"
	"os"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panicf("Unable to start tgbot, %v", err)
	}

	botService := bot_service.MakeBotService(bot)

	repo := storage.NewRepo()
	userService := user_config.NewService(repo)

	handler := updatehandler.MakeUpdateHandler(botService, userService)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		err := handler.HandleUpdate(&update)
		if err != nil {
			log.Printf("Unable to handle update, err: %v.", err)
		}
	}
}
