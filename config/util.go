package config

import (
  "fmt"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "gopkg.in/yaml.v2"
  "io/ioutil"
)

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
