package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

type Reminder struct {
	Message *tgbotapi.Message

	IsSent     bool
	RemindTime int
}

type ReminderResponse struct {
	Reminder *Reminder
	Message  *tgbotapi.Message
}

type Config struct {
	Token    string
	Duration int
	Debug    bool
}

func SendMessage(message tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) (*tgbotapi.Message, error) {

	sentMessage, err := bot.Send(message)
	return &sentMessage, err

}

func CallbackProcessor(callback tgbotapi.CallbackQuery, reminders *[]*Reminder, bot *tgbotapi.BotAPI) {
	edit := tgbotapi.NewEditMessageReplyMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
		},
	)

	bot.Send(edit)

	if callback.Data == "Complete" {
		return
	}

	newRemindTime := int(time.Now().Unix())
	switch callback.Data {
	case "sec5":
		newRemindTime += 5
	case "min20":
		newRemindTime += 1200
	case "hour1":
		newRemindTime += 3600
	case "hour3":
		newRemindTime += 3600
	case "day1":
		newRemindTime += 86400
	}

	for _, reminder := range *reminders {
		if reminder.Message.MessageID == callback.Message.ReplyToMessage.MessageID {
			reminder.RemindTime = newRemindTime
			reminder.IsSent = false
		}
	}

}

func ReminderProcessor(reminders *[]*Reminder, bot *tgbotapi.BotAPI) {
	for {
		for i := len(*reminders); i > 0; i-- {
			reminder := (*reminders)[i-1]

			if (*reminder).RemindTime <= int(time.Now().Unix()) && (*reminder).IsSent == false {
				msg := tgbotapi.NewMessage((*reminder).Message.Chat.ID, "You asked me to remind you about this message. Snooze it?")
				msg.ReplyToMessageID = (*reminder).Message.MessageID

				keyboard := tgbotapi.InlineKeyboardMarkup{}
				var row1 []tgbotapi.InlineKeyboardButton
				var row2 []tgbotapi.InlineKeyboardButton

				row1 = append(row1,
					tgbotapi.NewInlineKeyboardButtonData("20 min", "min20"),
					tgbotapi.NewInlineKeyboardButtonData("1 hour", "hour1"),
					tgbotapi.NewInlineKeyboardButtonData("3 hours", "hour1"),
					tgbotapi.NewInlineKeyboardButtonData("1 day", "day1"),
				)

				row2 = append(row2,
					tgbotapi.NewInlineKeyboardButtonData("\u2705 Mark as Complete", "Complete"))
				keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row1, row2)

				msg.ReplyMarkup = keyboard

				(*reminder).IsSent = true
				bot.Send(msg)

			}
		}
		time.Sleep(1 * time.Second)
	}

}

func NewConfig() Config {
	token := os.Getenv("TOKEN")
	duration := 20
	debug := false
	if os.Getenv("DURATION") != "" {
		duration, _ = strconv.Atoi(os.Getenv("DURATION"))
	}
	if os.Getenv("DEBUG") != "" {
		debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	}

	return Config{
		Token:    token,
		Duration: duration,
		Debug:    debug,
	}
}

func main() {

	config := NewConfig()

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = config.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	remindersList := []*Reminder{}

	go ReminderProcessor(&remindersList, bot)

	for update := range updates {

		if update.CallbackQuery != nil {
			CallbackProcessor(*(update).CallbackQuery, &remindersList, bot)
		}

		if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi! Just write me or forward a message and I'll remind you about it in 20 minutes. \u23F0")
			bot.Send(msg)
			continue
		}

		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		reminder := Reminder{
			Message:    update.Message,
			RemindTime: int(time.Now().Unix()) + config.Duration,
			IsSent:     false,
		}

		remindersList = append(remindersList, &reminder)

	}

}
