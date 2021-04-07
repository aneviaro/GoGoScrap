package bot_service

type Sender interface {
	SendMessage(chatID int64, message string, replyTo int) error
}
