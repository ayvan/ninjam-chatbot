package telegram_bot

import (
	"github.com/ayvan/ninjam-chatbot/models"
	"github.com/sirupsen/logrus"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
	"time"
)

type TelegramBot struct {
	sigChan              chan bool
	token                string
	chatID               int64
	messagesToTelegram   chan string
	messagesFromTelegram chan models.Message
	models.Mountser
	disabled bool
}

func NewTelegramBot(token string, chatID int64, mounts models.Mountser) *TelegramBot {
	return &TelegramBot{
		sigChan:              make(chan bool, 1),
		token:                token,
		chatID:               chatID,
		messagesToTelegram:   make(chan string, 1000),
		messagesFromTelegram: make(chan models.Message, 1000),
		Mountser:             mounts,
	}
}

func (t *TelegramBot) Disabled(disabled bool) {
	t.disabled = disabled
}

func (t *TelegramBot) IncomingMessages() <-chan models.Message {
	return t.messagesFromTelegram
}

func (t *TelegramBot) SendMessage(message string) {
	if t.disabled {
		return
	}
	go func() {
		t.messagesToTelegram <- message
	}()
}

func (t *TelegramBot) Connect() {
	if t.disabled {
		return
	}
f:
	for {
		select {
		case s := <-t.sigChan:
			t.sigChan <- s
			break f
		default:
			t.connect()
			// если коннект прервался - запустим таймаут перед реконнектом
			time.Sleep(time.Second * 5)
		}
	}
}

func (t *TelegramBot) Stop() {
	if t.disabled {
		return
	}
	t.sigChan <- true
}

func (t *TelegramBot) connect() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Warnf("Recovered in connect: %s ", r)
		}
	}()

	// подключаемся к боту с помощью токена
	bot, err := tgbotapi.NewBotAPI(t.token)

	if err != nil {
		logrus.Errorf("NewBotAPI error: %s", err)
		return
	}

	bot.Debug = true
	logrus.Infof("Authorized on account %s", bot.Self.UserName)

	// инициализируем канал, куда будут прилетать обновления от API
	var ucfg = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60

	updates, err := bot.GetUpdatesChan(ucfg)

	if err != nil {
		logrus.Errorf("GetUpdatesChan error: %s", err)
		return
	}
	// читаем обновления из канала
	for {
		select {
		case s := <-t.sigChan:
			t.sigChan <- s
			return
		case message := <-t.messagesToTelegram:
			logrus.Infof("Sending message to Telegram: %s", message)
			// Созадаем сообщение
			msg := tgbotapi.NewMessage(t.chatID, message)
			// и отправляем его
			bot.Send(msg)

		case update := <-updates:

			if update.Message == nil {
				continue
			}
			// Пользователь, который написал боту
			var UserName string
			// ID чата/диалога.
			// Может быть идентификатором как чата с пользователем
			// (тогда он равен UserID) так и публичного чата/канала
			var ChatID int64
			// Текст сообщения
			var Text string

			UserName = update.Message.From.UserName
			ChatID = update.Message.Chat.ID
			Text = update.Message.Text

			if len(update.Message.From.UserName) != 0 {
				UserName = update.Message.From.UserName
			} else {
				UserName = update.Message.From.FirstName
			}

			logrus.Infof("Message received: [%s] %d %s", UserName, ChatID, Text)

			i := strings.Index(Text, "@"+bot.Self.UserName)

			if i != -1 {
				Text = Text[:i]
			}
			Text = strings.Trim(Text, "/")
			Text = strings.TrimSpace(Text)

			if Text == "" {
				continue
			}

			mount, ok := t.Mounts()[Text]

			switch true {
			case Text == "start":
				reply := ""
				if len(t.Mounts()) > 0 {

					reply = "Сейчас активны серверы: "
					for k := range t.Mounts() {
						reply = reply + k + " "
					}
					for k, m := range t.Mounts() {
						users := strings.Join(m, ", ")
						if users != "" {
							reply = reply + "\nНа сервере " + k + " играют: " + users
						} else {
							reply = reply + "\nНа сервере " + k + " никого нет. "
						}
					}
				} else {
					reply = "Нет активных серверов!"
				}

				// Созадаем сообщение
				msg := tgbotapi.NewMessage(ChatID, reply)
				// и отправляем его
				bot.Send(msg)

			case ok:
				users := strings.Join(mount, ", ")
				reply := ""
				if users != "" {
					reply = "На сервере " + Text + " играют: " + users
				} else {
					reply = "На сервере " + Text + " никого нет. "
				}
				// Созадаем сообщение
				msg := tgbotapi.NewMessage(ChatID, reply)
				// и отправляем его
				bot.Send(msg)
			case Text == "help":
				reply := "Сайт джем-серверов с информацией об адресах находится по адресу http://guitar-jam.ru\n"
				reply = reply + "Подробнее о джем-серверах, подключении к ним и по остальным вопросам читайте тему http://forum.gitarizm.ru/showthread.php?t=39731 и задавайте вопросы там или в этом чате."

				// Созадаем сообщение
				msg := tgbotapi.NewMessage(ChatID, reply)
				// и отправляем его
				bot.Send(msg)
			default:
				logrus.Infof("Received chat message from %s: %s", UserName, Text)

				m := models.Message{
					Type: models.MSG,
					Name: UserName,
					Text: Text,
				}

				t.messagesFromTelegram <- m
			}
		}

	}
}

type Status struct {
	Mount []Mount `xml:"mount"`
}

type Mount struct {
	Name      string `xml:"name"`
	Listeners string `xml:"listeners"`
	Users     string `xml:"users"`
}
