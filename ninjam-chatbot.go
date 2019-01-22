package main

import (
	"fmt"
	"github.com/Ayvan/ninjam-chatbot/config"
	"github.com/Ayvan/ninjam-chatbot/models"
	"github.com/Ayvan/ninjam-chatbot/ninjam-bot"
	"github.com/Ayvan/ninjam-chatbot/slack-bot"
	"github.com/Ayvan/ninjam-chatbot/telegram-bot"
	"github.com/VividCortex/godaemon"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type Mounts struct {
	mounts map[string]models.Userser
}

func (m *Mounts) Mounts() map[string][]string {
	ms := map[string][]string{}

	for k, mount := range m.mounts {
		ms[k] = mount.Users()
	}

	return ms
}

func main() {
	if config.Get().DaemonMode {
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	}

	pid := fmt.Sprintf("%d", os.Getpid())

	pidFile := config.Get().AppPath + "/app.pid"

	err := ioutil.WriteFile(pidFile, []byte(pid), 0644)

	if err != nil {
		logrus.Fatal("Error when writing pidfile:", err)
	}

	defer func() {
		os.Remove(pidFile)
	}()

	sChan := make(chan os.Signal, 1)
	// ловим команды на завершение от ОС и корректно завершаем приложение с помощью sync.WaitGroup
	signal.Notify(sChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	mounts := &Mounts{
		mounts: make(map[string]models.Userser),
	}

	bots := make([]*ninjam_bot.NinJamBot, 0)

	for _, server := range config.Get().Servers {
		bot := ninjam_bot.NewNinJamBot(server.Host, server.Port, server.UserName, server.UserPassword, server.Anonymous)
		mounts.mounts[server.Port] = bot
		bots = append(bots, bot)
	}

	tbot := telegram_bot.NewTelegramBot(config.Get().Telegram.Token, config.Get().Telegram.ChatID, mounts)
	if config.Get().Telegram.Disabled {
		tbot.Disabled(true)
	}

	sbot := slack_bot.NewSlackBot(config.Get().Slack.Token, config.Get().Slack.Channel, config.Get().Slack.BotName, mounts)
	if config.Get().Slack.Disabled {
		sbot.Disabled(true)
	}

	// инициализируем глобальный канал завершения горутин
	sigChan := make(chan bool, 1)

	go func() {
		// ловим сигнал завершения, выводим информацию в лог, а затем отправляем его в глобальный канал
		s := <-sChan
		logrus.Info("os.Signal ", s, " received, finishing application...")
		for _, bot := range bots {
			bot.Stop()
		}

		tbot.Stop()
		sbot.Stop()
		sigChan <- true
		return
	}()

	wg := &sync.WaitGroup{}

	logrus.Info("Application ", config.Get().AppName, " started")

	type BotMessage struct {
		Bot     *ninjam_bot.NinJamBot
		Message models.Message
	}

	botChan := make(chan BotMessage, 1000)

	for _, bot := range bots {
		wg.Add(1)
		go func(bot *ninjam_bot.NinJamBot) {
			defer wg.Done()
			bot.Connect()
		}(bot)

		go func(bot *ninjam_bot.NinJamBot) {
			for {
				select {
				case msg := <-bot.IncomingMessages():
					bm := BotMessage{
						Bot:     bot,
						Message: msg,
					}
					botChan <- bm
				case <-sigChan:
					sigChan <- true
					return
				}
			}
		}(bot)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		tbot.Connect()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sbot.Connect()
	}()

	tbotChan := tbot.IncomingMessages()
	sbotChan := sbot.IncomingMessages()

f:
	for {
		select {
		case s := <-sigChan:
			sigChan <- s
			break f
			// messages routers <->
		case msg := <-botChan:
			if strings.HasPrefix(msg.Message.Name, msg.Bot.UserName()) {
				continue
			}

			for _, userName := range config.Get().IgnoreUsers {
				if strings.HasPrefix(msg.Message.Name, userName) {
					continue f
				}
			}

			for _, botName := range config.Get().IgnorePrefix {
				if strings.HasPrefix(msg.Message.Text, botName) {
					continue f
				}
			}

			message := fmt.Sprintf("%s@%s:%s: %s", msg.Message.Name, msg.Bot.Host(), msg.Bot.Port(), msg.Message.Text)

			switch msg.Message.Type {
			case models.MSG:
				logrus.Infof("Sendind to Telegram: %s", message)
				tbot.SendMessage(message)
				logrus.Infof("Sendind to Slack: %s", message)
				sbot.SendMessage(message)
				for _, bot := range bots {
					logrus.Error(bot.Host(), msg.Bot.Host(), bot.Port(), msg.Bot.Port())
					if bot.Host() != msg.Bot.Host() || bot.Port() != msg.Bot.Port() {
						logrus.Infof("Sendind to Ninjam %s:%s", bot.Host(), bot.Port())
						bot.SendMessage(message)
					}
				}
			case models.JOIN:
				message := fmt.Sprintf("%s зашёл на джем-сервер %s:%s ", msg.Message.Name, msg.Bot.Host(), msg.Bot.Port())
				logrus.Infof("Sendind to Telegram: %s", message)
				tbot.SendMessage(message)
				logrus.Infof("Sendind to Slack: %s", message)
				sbot.SendMessage(message)
			case models.PART:
				message := fmt.Sprintf("%s покинул джем-сервер %s:%s ", msg.Message.Name, msg.Bot.Host(), msg.Bot.Port())
				logrus.Infof("Sendind to Telegram: %s", message)
				tbot.SendMessage(message)
				logrus.Infof("Sendind to Slack: %s", message)
				sbot.SendMessage(message)
			}

		case msg := <-tbotChan:
			message := fmt.Sprintf("%s@telegram: %s", msg.Name, msg.Text)
			logrus.Infof("Sendind to NinJam: %s", message)
			for _, bot := range bots {
				bot.SendMessage(message)
			}
			logrus.Infof("Sendind to Slack: %s", message)
			sbot.SendMessage(message)
		case msg := <-sbotChan:
			message := fmt.Sprintf("%s@slack: %s", msg.Name, msg.Text)
			logrus.Infof("Sendind to NinJam: %s", message)
			for _, bot := range bots {
				bot.SendMessage(message)
			}
			logrus.Infof("Sendind to Telegram: %s", message)
			tbot.SendMessage(message)
		}

	}

	wg.Wait()

	// t.SendMessage(fmt.Sprint("Новые участники на джем-сервере ", n.Name, ": ", liUsers))
	// t.SendMessage(fmt.Sprint("Джем-сервер ", n.Name, " покинули: ", loUsers))

	logrus.Info("Application ", config.Get().AppName, " finished")
}
