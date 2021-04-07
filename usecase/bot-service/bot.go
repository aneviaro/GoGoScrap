package bot_service

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type BotService struct {
	tg *tgbotapi.BotAPI
}

func MakeBotService(bot *tgbotapi.BotAPI) *BotService {
	return &BotService{tg: bot}
}

func (bot *BotService) SendMessage(chatID int64, message string, replyTo int, parseMode string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	if replyTo != 0 {
		msg.ReplyToMessageID = replyTo
	}
	msg.ParseMode = parseMode
	_, err := bot.tg.Send(msg)
	return err
}
