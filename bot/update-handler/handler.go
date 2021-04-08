package update_handler

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gogoscrap/repository"
	botservice "gogoscrap/usecase/bot-service"
	"gogoscrap/usecase/scrapper"
	userconfig "gogoscrap/usecase/user-config"
	"strings"
)

type UpdateHandler struct {
	botService  *botservice.BotService
	userService *userconfig.UserService
}

func MakeUpdateHandler(botService *botservice.BotService, userService *userconfig.UserService) *UpdateHandler {
	return &UpdateHandler{
		botService:  botService,
		userService: userService,
	}
}

func (u *UpdateHandler) HandleUpdate(update *tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	if update.Message.Text == "/start" || strings.Contains(update.Message.Text, "/restart") {
		err := u.botService.SendMessage(update.Message.Chat.ID,
			"Welcome here! Please type in the google query and website you want to track. Please, "+
				"separate with a comma. Currently, I only support Belarus google location.", 0, "")
		return err
	}

	if update.Message.ReplyToMessage != nil && strings.Contains(update.Message.ReplyToMessage.Text, "Please, "+
		"reply to this message") {
		url := strings.Trim(update.Message.Text, " .,")

		u.userService.Save(repository.UserConfig{
			UserID:  update.Message.Chat.ID,
			Website: url,
		})

		err := u.botService.SendMessage(update.Message.Chat.ID,
			"I've started a search for URL. "+
				"Now you can check queries without mentioning your website directly.", 0, "")
		return err
	}

	if strings.Contains(update.Message.Text, "/urlsearch") {
		url, err := parseCommand(update.Message.Text)
		if err != nil {
			err := u.botService.SendMessage(update.Message.Chat.ID, "Please, reply to this message, "+
				"with website URL you want to start search on.",
				update.Message.MessageID, "Markdown")
			return err
		}

		u.userService.Save(repository.UserConfig{
			UserID:  update.Message.Chat.ID,
			Website: url,
		})

		err = u.botService.SendMessage(update.Message.Chat.ID,
			"I've started a search for URL. "+
				"Now you can check queries without mentioning your website directly.", 0, "")
		return err
	}

	if strings.Contains(update.Message.Text, "/stopurlsearch") {
		u.userService.RemoveConfigByUser(update.Message.Chat.ID)
		err := u.botService.SendMessage(update.Message.Chat.ID, "I've stopped a search by URL for you. "+
			"Please type in the google query and website you want to track. You should use comma as a separator.", 0, "")
		return err
	}

	query, website, err := parseMessage(update.Message.Text)

	if err != nil {
		err := u.botService.SendMessage(update.Message.Chat.ID, "I could not parse your message, please, "+
			"use following format to start new search by website: \n**pizza in Minsk, dominos.by**.",
			update.Message.MessageID, "Markdown")
		return err
	}

	if website == "" {
		config, err := u.userService.GetByUser(update.Message.Chat.ID)
		if err != nil {
			err := u.botService.SendMessage(update.Message.Chat.ID, "I could not parse your message, please, "+
				"use following format to start new search by website: \n**pizza in Minsk, dominos.by**.",
				update.Message.MessageID, "Markdown")
			return err
		}
		website = config.Website
	}

	langCode := update.Message.From.LanguageCode

	urlRank, err := scrapper.GetWebsitePositionForQuery(query, langCode, website)

	if err != nil {
		err := u.botService.SendMessage(update.Message.Chat.ID, "I met an error, "+
			"trying to scrap google for you. Please, try again later. ", update.Message.MessageID, "")
		return err
	}

	if urlRank.Rank == 0 {
		err := u.botService.SendMessage(update.Message.Chat.ID, "I haven't met your website in TOP-100 links in a google search. Please, "+
			"check your query and website and try again!", update.Message.MessageID, "")
		return err
	}

	err = u.botService.SendMessage(update.Message.Chat.ID,
		fmt.Sprintf("I met your website on TOP-%v place in Google Search with the URL: %s.", urlRank.Rank,
			urlRank.URL), update.Message.MessageID, "")
	return err
}

func parseMessage(message string) (string, string, error) {
	arr := strings.Split(message, ",")
	if len(arr) > 2 {
		return "", "", errors.New("wrong format")
	} else if len(arr) == 2 {
		return strings.Trim(arr[0], " "), strings.Trim(arr[1], " "), nil
	} else if len(arr) == 1 {
		return strings.Trim(arr[0], " "), "", nil
	}
	return "", "", errors.New("unexpected error")
}

func parseCommand(message string) (string, error) {
	arr := strings.Split(message, " ")
	if len(arr) != 2 {
		return "", errors.New("command value not set")
	}
	return strings.Trim(arr[1], " "), nil
}
