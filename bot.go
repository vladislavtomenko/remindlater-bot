package main

import (
  "log"
  "time"
  "strconv"
  "os"
  "gopkg.in/telegram-bot-api.v4"
)

type Reminder struct {
    Basics *tgbotapi.Message
    Duration int
}

type Config struct {
    Token string
    Duration int
    Debug bool
}

func Notify(reminders chan *Reminder, bot *tgbotapi.BotAPI) {
  for reminder := range reminders {
    msg := tgbotapi.NewMessage(reminder.Basics.Chat.ID, "You asked to remind")
    msg.ReplyToMessageID = reminder.Basics.MessageID
    bot.Send(msg)
  }
}

func ReminderProcessor(reminders *[]*Reminder, remindersChannel chan *Reminder) {
  for {
    for i := len(*reminders); i > 0; i-- {
      reminder := &(*reminders)[i-1]

      if (*reminder).Basics.Date + (*reminder).Duration <= int(time.Now().Unix()) {
        remindersChannel <- *reminder
        *reminders = append((*reminders)[:i-1], (*reminders)[i:]...)
      }
    }

    time.Sleep(1 * time.Second)
  }
}

func GenerateConfig() Config {
  token := os.Getenv("TOKEN")
  duration := 20
  debug := false
  if os.Getenv("DURATION") != "" {
    duration, _ = strconv.Atoi(os.Getenv("DURATION"))
  }
  if os.Getenv("DEBUG") != "" {
    debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
  }

  return Config {
    Token: token,
    Duration: duration,
    Debug: debug,
  }
}


func main() {

  config := GenerateConfig()

  bot, err := tgbotapi.NewBotAPI(config.Token)
  if err != nil {
    log.Panic(err)
  }

  bot.Debug = config.Debug

  log.Printf("Authorized on account %s", bot.Self.UserName)

  u := tgbotapi.NewUpdate(0)
  u.Timeout = 60

  updates, err := bot.GetUpdatesChan(u)

  reminders := make(chan *Reminder)

  go Notify(reminders, bot)

  remindersList := []*Reminder{}
  go ReminderProcessor(&remindersList, reminders)

  for update := range updates {

    if update.Message == nil {
      continue
    }

    log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

    reminder := &Reminder {
      Basics: update.Message,
      Duration: config.Duration,
    }

    remindersList = append(remindersList, reminder)
   
  }
  
}
