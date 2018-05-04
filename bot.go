package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

// Reminder is a sent user message and a date when remind about it.
type Reminder struct {
	Message *tgbotapi.Message

	IsSent     bool
	RemindTime int
}

// Check if it's time to send the reminder.
func (reminder Reminder) IsTimeToProcess() bool {
	return reminder.RemindTime <= int(time.Now().Unix())
}

type ReminderResponse struct {
	Reminder *Reminder
	Message  *tgbotapi.Message
}

// Config is the bot configuration.
//
// Token is a telegram bot API token,
// Duration is the default reminder delay in seconds ,
// Debug enables extended output.
type Config struct {
	Token    string
	Duration int
	Debug    bool
}

// CallbackHandler handles the callbacks from inline buttons.
//
// It gets the callback, updates (or remove) a reminder in reminders,
// and send a message to the bot.
func CallbackHandler(callback tgbotapi.CallbackQuery, reminders *[]*Reminder, bot *tgbotapi.BotAPI) {

	var newReplyText string

	var reminderIndex int
	for i, reminder := range *reminders {
		if reminder.Message.MessageID == callback.Message.ReplyToMessage.MessageID {
			reminderIndex = i
		}
	}

	if callback.Data == "Complete" {
		*reminders = append((*reminders)[:reminderIndex], (*reminders)[reminderIndex+1:]...)
		newReplyText = "Completed \u2705"
	} else {
		newRemindTime := int(time.Now().Unix())

		switch callback.Data {
		case "5 sec":
			newRemindTime += 5
		case "20 min":
			newRemindTime += 1200
		case "1 hour":
			newRemindTime += 3600
		case "3 hours":
			newRemindTime += 10800
		case "1 day":
			newRemindTime += 86400
		}

		(*reminders)[reminderIndex].RemindTime = newRemindTime
		(*reminders)[reminderIndex].IsSent = false

		newReplyText = "Snoozed for " + callback.Data + " \u23F3"
	}

	// Hide the inline buttons.
	editMarkup := tgbotapi.NewEditMessageReplyMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
		},
	)
	// Edit the reply text.
	editText := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		newReplyText,
	)
	bot.Send(editMarkup)
	bot.Send(editText)

}

// ReminderHandler handles the reminders queue.
//
// The function sends a reminder if it's time to remind.
// Otherwise it waits for an appropriate reminder.
func ReminderHandler(reminders *[]*Reminder, bot *tgbotapi.BotAPI) {
	for {
		for i := len(*reminders); i > 0; i-- {
			reminder := (*reminders)[i-1]

			if (*reminder).IsTimeToProcess() && (*reminder).IsSent == false {
				msg := tgbotapi.NewMessage((*reminder).Message.Chat.ID, "You asked me to remind you about this message. Snooze it?")
				msg.ReplyToMessageID = (*reminder).Message.MessageID

				// Add the inline buttons to a reply.
				keyboard := tgbotapi.InlineKeyboardMarkup{}
				var row1 []tgbotapi.InlineKeyboardButton
				var row2 []tgbotapi.InlineKeyboardButton
				row1 = append(row1,
					tgbotapi.NewInlineKeyboardButtonData("20 min", "20 min"),
					tgbotapi.NewInlineKeyboardButtonData("1 hour", "1 hour"),
					tgbotapi.NewInlineKeyboardButtonData("3 hours", "3 hours"),
					tgbotapi.NewInlineKeyboardButtonData("1 day", "1 day"),
				)
				row2 = append(row2,
					tgbotapi.NewInlineKeyboardButtonData("\u2705 Mark as Complete", "Complete"),
				)

				// One more button if the debug mode is on.
				if bot.Debug {
					row2 = append(row2,
						tgbotapi.NewInlineKeyboardButtonData("5 sec", "sec5"),
					)
				}

				(*reminder).IsSent = true

				keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row1, row2)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)

			}
		}
		time.Sleep(1 * time.Second)
	}
}

// NewConfig creates a new Config.
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

	go ReminderHandler(&remindersList, bot)

	for update := range updates {

		// Handle a callback
		if update.CallbackQuery != nil {
			CallbackHandler(*update.CallbackQuery, &remindersList, bot)
			continue
		}

		// Do nothing if the message text is empty or callback received
		if update.Message == nil {
			continue
		}

		// Show the welcome message
		if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi! Just write me or forward a message and I'll remind you about it in 20 minutes. \u23F0")
			bot.Send(msg)
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		reminder := Reminder{
			Message:    update.Message,
			RemindTime: int(time.Now().Unix()) + config.Duration,
			IsSent:     false,
		}

		// Put a reminder in the qeue
		remindersList = append(remindersList, &reminder)
	}
}
