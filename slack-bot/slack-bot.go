package slack_bot

import (
	"encoding/json"
	"github.com/ayvan/ninjam-chatbot/models"
	"github.com/sirupsen/logrus"
	"github.com/nlopes/slack"
	"strings"
	"time"
)

type SlackBot struct {
	sigChan           chan bool
	botName           string
	token             string
	channel           string
	channelID         string
	messagesToSlack   chan string
	messagesFromSlack chan models.Message
	models.Mountser
	disabled bool
}

func NewSlackBot(token, channel, botName string, mounts models.Mountser) *SlackBot {
	return &SlackBot{
		sigChan:           make(chan bool, 1),
		botName:           botName,
		token:             token,
		channel:           channel,
		messagesToSlack:   make(chan string, 1000),
		messagesFromSlack: make(chan models.Message, 1000),
		Mountser:          mounts,
	}
}

func (sb *SlackBot) Disabled(disabled bool) {
	sb.disabled = disabled
}

func (sb *SlackBot) IncomingMessages() <-chan models.Message {
	return sb.messagesFromSlack
}

func (sb *SlackBot) SendMessage(message string) {
	if sb.disabled {
		return
	}
	go func() {
		sb.messagesToSlack <- message
	}()
}

func (sb *SlackBot) Connect() {
	if sb.disabled {
		return
	}
f:
	for {
		select {
		case s := <-sb.sigChan:
			sb.sigChan <- s
			break f
		default:
			sb.connect()
			// если коннект прервался - запустим таймаут перед реконнектом
			time.Sleep(time.Second * 5)
		}
	}
}

func (sb *SlackBot) Stop() {
	if sb.disabled {
		return
	}
	sb.sigChan <- true
}

func (sb *SlackBot) connect() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Warnf("Recovered in connect(): %s ", r)
		}
	}()

	api := slack.New(sb.token)

	//logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	//slack.SetLogger(logger)
	//api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	cnls, err := rtm.GetChannels(true)

	if err != nil {
		logrus.Errorf("Slack GetChannels error: %s", err)
	}

	for _, c := range cnls {
		if c.Name == sb.channel {
			sb.channelID = c.ID
		}
	}

	for {
		select {
		case s := <-sb.sigChan:
			sb.sigChan <- s
			return
		case msg := <-sb.messagesToSlack:
			logrus.Infof("Sending message to Slack: %s", msg)
			// Созадаем сообщение
			message := rtm.NewOutgoingMessage(msg, sb.channelID)
			// и отправляем его
			rtm.SendMessage(message)
		case msg := <-rtm.IncomingEvents:
			msgJSON, _ := json.Marshal(msg)
			logrus.Infof("Slack event received: %T %s", msg, string(msgJSON))
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				// Пользователь, который написал боту
				var userName string
				// ID чата/диалога.
				// Может быть идентификатором как чата с пользователем
				// (тогда он равен UserID) так и публичного чата/канала
				var channel string
				// Текст сообщения
				var text string

				//userName = ev.User
				channel = ev.Channel
				text = ev.Text

				u, err := rtm.GetUserInfo(ev.User)
				userName = u.Profile.DisplayName

				if err != nil {
					logrus.Error("GetUserInfo error:", err)
				}

				if userName == sb.botName || u.Name == sb.botName {
					continue
				}

				logrus.Infof("Message received: [%s] %s %s", userName, channel, text)

				text = strings.Trim(text, "/")
				text = strings.TrimSpace(text)

				if text == "" {
					continue
				}

				mountName := ""

				i := strings.Index(text, sb.botName)

				if i == 0 {
					mountName = text[len(sb.botName):]
					mountName = strings.TrimSpace(mountName)
				}

				mount, ok := sb.Mounts()[mountName]

				switch true {
				case text == sb.botName+" info":
					reply := ""
					if len(sb.Mounts()) > 0 {

						reply = "Сейчас активны серверы: "
						for k := range sb.Mounts() {
							reply = reply + k + " "
						}
						for k, m := range sb.Mounts() {
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
					message := rtm.NewOutgoingMessage(reply, channel)
					// и отправляем его
					rtm.SendMessage(message)
				case ok:
					users := strings.Join(mount, ", ")
					reply := ""
					if users != "" {
						reply = "На сервере " + text + " играют: " + users
					} else {
						reply = "На сервере " + text + " никого нет. "
					}
					// Созадаем сообщение
					message := rtm.NewOutgoingMessage(reply, channel)
					// и отправляем его
					rtm.SendMessage(message)
				case text == sb.botName+" help":
					reply := "Сайт джем-серверов с информацией об адресах находится по адресу http://guitar-jam.ru\n"
					reply = reply + "Подробнее о джем-серверах, подключении к ним и по остальным вопросам читайте тему http://forum.gitarizm.ru/showthread.php?t=39731 и задавайте вопросы там или в этом чате.\n"
					reply = reply + "Команды бота:\n"
					reply = reply + sb.botName + " info\n"
					reply = reply + sb.botName + " help\n"
					reply = reply + sb.botName + " SERVER_PORT (например \"" + sb.botName + " 2050\")\n"

					// Созадаем сообщение
					message := rtm.NewOutgoingMessage(reply, channel)
					// и отправляем его
					rtm.SendMessage(message)
				default:
					logrus.Infof("Received chat message from %s: %s", userName, text)

					m := models.Message{
						Type: models.MSG,
						Name: userName,
						Text: text,
					}

					sb.messagesFromSlack <- m
				}
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
