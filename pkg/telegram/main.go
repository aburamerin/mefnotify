package telegram

import (
	"log"

	"mefnotify/pkg/posts"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func New(apiKey string, chatID int64) *Client {
	bot, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		log.Panic(err)
	}

	return &Client{
		bot:    bot,
		chatID: chatID,
	}
}

func (c *Client) SendMessage(p posts.Post) error {
	buttonURL := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Читать на сайте", p.Info.URL),
		),
	)
	msg := tgbotapi.NewMessage(c.chatID, p.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = buttonURL
	_, err := c.bot.Send(msg)
	return err
}

func (c *Client) GetUpdate() {
	c.bot.Debug = true

	log.Printf("Authorized on account %s", c.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			resp, err := c.bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %S", err)
			}

			log.Printf("resp: %v", resp)
		}
	}
}
