package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gogoscrap/repository"
	"gogoscrap/storage"
	"gogoscrap/usecase/user-config"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panicf("Unable to start tgbot, %v", err)
	}

	st := storage.NewStorage()
	service := user_config.NewService(st)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/start" || strings.Contains(update.Message.Text, "/restart") {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Welcome here! Please type in the google query and website you want to track. Please, "+
					"separate with a comma. Currently I only support Belarus google location.")
			sentMessage, err := bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send message to bot: %v", err)
				continue
			}
			log.Printf("Message sent succesfully: %v", sentMessage)
			continue
		}

		if strings.Contains(update.Message.Text, "/start-url-search") {
			url, err := parseCommand(update.Message.Text)
			if err != nil {
				log.Printf("Unable to parse command, err: %v.", err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I could not parse your message, please, "+
					"use following format to start new search by website: \n**/start-url-search dominos.by**.")
				msg.ReplyToMessageID = update.Message.MessageID
				msg.ParseMode = "Markdown"
				sentMessage, err := bot.Send(msg)
				if err != nil {
					log.Printf("Unable to send message to bot: %v.", err)
					continue
				}
				log.Printf("Message sent succesfully: %v.", sentMessage)
				continue
			}

			service.Save(repository.UserConfig{
				UserID:  update.Message.Chat.ID,
				Website: url,
			})

			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"I've started a search for URL. "+
					"Now you can check queries without mentioning your website directly.")
			sentMessage, err := bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send message to bot: %v", err)
				continue
			}

			log.Printf("Message sent succesfully: %v", sentMessage)
			continue
		}

		if strings.Contains(update.Message.Text, "/stop-url-search") {
			service.RemoveConfigByUser(update.Message.Chat.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I've stopped a search by URL for you. "+
				"Please type in the google query and website you want to track. You should use comma as a separator.")
			sentMessage, err := bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send message to bot: %v", err)
				continue
			}

			log.Printf("Message sent succesfully: %v", sentMessage)
			continue
		}

		query, website, err := parseMessage(update.Message.Text)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I could not parse your message, please, "+
				"use following format to track your query: \n**pizza in Minsk, dominos.com**.")
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ParseMode = "Markdown"
			sentMessage, err := bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send message to bot: %v.", err)
				continue
			}
			log.Printf("Message sent succesfully: %v.", sentMessage)
			continue
		}

		if website == "" {
			config, err := service.GetByUser(update.Message.Chat.ID)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I could not parse your message, please, "+
					"use following format to track your query: \n**pizza in Minsk, dominos.com**.")
				msg.ReplyToMessageID = update.Message.MessageID
				msg.ParseMode = "Markdown"
				sentMessage, err := bot.Send(msg)
				if err != nil {
					log.Printf("Unable to send message to bot: %v.", err)
					continue
				}
				log.Printf("Message sent succesfully: %v.", sentMessage)
				continue
			}
			website = config.Website
		}

		langCode := update.Message.From.LanguageCode
		countryCode := "by"

		googleUrl := buildGoogleUrl(query, countryCode, langCode)
		res, err := googleRequest(googleUrl)

		if err != nil {
			log.Printf("Unable to make google request, err: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Something went wrong when trying to make google request. Please, try again.")
			msg.ReplyToMessageID = update.Message.MessageID
			sentMessage, err := bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send message to bot: %v", err)
				continue
			}
			log.Printf("Message sent succesfully: %v", sentMessage)
			continue
		}

		result, err := googleResultParser(res, website)
		if err != nil {
			log.Printf("Unable to parse google response, err: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something went wrong when trying to parse google response. Please, try again.")
			msg.ReplyToMessageID = update.Message.MessageID
			sentMessage, err := bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send message to bot: %v", err)
				continue
			}
			log.Printf("Message sent succesfully: %v", sentMessage)
			continue
		}

		if result.ResultRank == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("I haven't met your website in TOP-100 links in a google search. Please, "+
					"check your query and website and try again!"))
			msg.ReplyToMessageID = update.Message.MessageID
			sentMessage, err := bot.Send(msg)
			if err != nil {
				log.Printf("Unable to send message to bot: %v", err)
				continue
			}
			log.Printf("Message sent succesfully: %v", sentMessage)
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("I met your website on TOP-%v place in Google Search with the URL: %s.", result.ResultRank,
				result.ResultURL))
		msg.ReplyToMessageID = update.Message.MessageID

		sentMessage, err := bot.Send(msg)
		if err != nil {
			log.Printf("Unable to send message to bot: %v", err)
			continue
		}
		log.Printf("Message sent succesfully: %v", sentMessage)
	}
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

var googleDomains = map[string]string{
	"com": "https://www.google.com/search?q=",
	"uk":  "https://www.google.co.uk/search?q=",
	"ru":  "https://www.google.ru/search?q=",
	"fr":  "https://www.google.fr/search?q=",
	"by":  "https://www.google.com/search?q=",
	"":    "https://www.google.com/search?q=",
}

type GoogleResult struct {
	ResultRank int
	ResultURL  string
}

func buildGoogleUrl(searchTerm string, countryCode string, languageCode string) string {
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	if googleBase, found := googleDomains[countryCode]; found {
		return fmt.Sprintf("%s%s&num=100&hl=%s", googleBase, searchTerm, languageCode)
	} else {
		return fmt.Sprintf("%s%s&num=100&hl=%s", googleDomains["com"], searchTerm, languageCode)
	}
}

func googleRequest(searchURL string) (*http.Response, error) {
	baseClient := &http.Client{}

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")

	res, err := baseClient.Do(req)

	return res, err
}

func googleResultParser(response *http.Response, siteName string) (GoogleResult, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return GoogleResult{}, err
	}
	sel := doc.Find("div.g")
	log.Printf(sel.Text())
	rank := 1
	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		link = strings.Trim(link, " ")
		if link != "" && link != "#" {
			result := GoogleResult{
				rank,
				link,
			}
			if strings.Contains(link, siteName) {
				return result, nil
			}
			rank += 1
		}
	}
	return GoogleResult{}, err
}
