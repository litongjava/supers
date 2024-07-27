package myutils

import (
  "fmt"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "gopkg.in/yaml.v2"
  "io/ioutil"
)

type Config struct {
  App *App `yaml:"app"`
}

type App struct {
  Port     int    `yaml:"port"`
  FilePath string `yaml:"filePath"`
  Password string `yaml:"password"`
}

var CONFIG *Config

func init() {
  yamlFile, err := ioutil.ReadFile("./config/config.yml")
  //yamlFile, err := ioutil.ReadFile(filename)
  if err != nil {
    hlog.Error(err.Error())
  }

  err = yaml.Unmarshal(yamlFile, &CONFIG)
  if err != nil {
    fmt.Println("error", err.Error())
  }
}
