package utils

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	App    *App          `yaml:"app"`
	Events *EventsConfig `yaml:"events"`
}

type App struct {
	Port     int    `yaml:"port"`
	FilePath string `yaml:"filePath"`
	Password string `yaml:"password"`
}

type EventsConfig struct {
	Webhooks []string `yaml:"webhooks"`
}

var CONFIG *Config

func init() {
	yamlFile, err := ioutil.ReadFile("./config/config.yml")
	if err != nil {
		hlog.Error(err.Error())
		return
	}
	CONFIG = &Config{}
	err = yaml.Unmarshal(yamlFile, CONFIG)
	if err != nil {
		fmt.Println("error parsing config:", err)
	}
}
