package config

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/luci/go-render/render"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type AppConfig struct {
	AppPath       string
	AppConfigPath string
	DaemonMode    bool           `yaml:"daemon"`
	AppName       string         `yaml:"app_name"`
	LogFile       string         `yaml:"log_file"`
	LogLevel      string         `yaml:"log_level"`
	Telegram      TelegramConf   `yaml:"telegram"`
	Slack         SlackConf      `yaml:"slack"`
	Servers       []NinJamServer `yaml:"servers"`
}

type NinJamServer struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Anonymous    bool   `yaml:"anonymous"`
	UserName     string `yaml:"user_name"`
	UserPassword string `yaml:"user_password"`
}

type TelegramConf struct {
	Token  string `yaml:"token"`
	ChatID int64  `yaml:"chat_id"`
}

type SlackConf struct {
	BotName string `yaml:"bot_name"`
	Token   string `yaml:"token"`
	Channel string  `yaml:"channel"`
}

var appConfig *AppConfig

func init() {
	appConfig = &AppConfig{}

	workPath, _ := os.Getwd()
	workPath, _ = filepath.Abs(workPath)
	// initialize default configurations
	appConfig.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	strPtr := flag.String("c", "config.yaml", "config path")

	flag.Parse()

	appConfig.AppConfigPath = *strPtr

	if workPath != appConfig.AppPath {
		if FileExists(appConfig.AppConfigPath) {
			os.Chdir(appConfig.AppPath)
		} else {
			appConfig.AppConfigPath = filepath.Join(workPath, "config.yaml")
		}
	}

	appConfig.DaemonMode = false
	appConfig.AppName = "ninjam-chatbot"
	appConfig.LogFile = "stdout"

	content, err := ioutil.ReadFile(appConfig.AppConfigPath)
	if err != nil {
		logrus.Fatalf("Can`t read config file (%s): %v\n", appConfig.AppConfigPath, err)
	}

	err = yaml.Unmarshal(content, appConfig)
	if err != nil {
		logrus.Fatalf("Yaml file %s parsing error: %v", appConfig.AppConfigPath, err)
	}

	setLogger(appConfig.LogLevel, appConfig.LogFile)
	if !appConfig.DaemonMode {
		logrus.Info("Config loaded:", render.Render(appConfig))
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
}

func setLogger(level, dest string) {
	lvl, err := logrus.ParseLevel(level)

	if err != nil {
		logrus.Fatalf("Unable to parse '%v' as a log level", level)
	}

	logrus.SetLevel(lvl)

	if dest != "stdout" {
		absDest, err := filepath.Abs(dest)
		if err != nil {
			logrus.Fatalf("Unable to get absolute file path %s: err: %s", dest, err)
		}

		out, err := os.OpenFile(absDest, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			logrus.Fatalf("Unable to open file %s: err: %s", dest, err)
		}

		logrus.SetOutput(out)
	}

	return
}

// FileExists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func Get() *AppConfig {
	return appConfig
}
