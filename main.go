package main

import (
  "deploy-server/config"
  "deploy-server/router"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "io"
  "net/http"
  "os"
  "strconv"
)

func Init() {
  hlog.SetLevel(hlog.LevelDebug)
  f, err := os.Create("app.log")
  if err != nil {
    panic(err)
  }
  defer f.Close()
  fileWriter := io.MultiWriter(f, os.Stdout)
  hlog.SetOutput(fileWriter)
}
func main() {
  port := strconv.Itoa(config.CONFIG.App.Port)
  for i := 1; i < len(os.Args); i += 2 {
    param := os.Args[i]
    if param == "--port" {
      port = os.Args[i+1]
    }
  }
  hlog.Info("start listen on:", port)
  router.RegisterRoutes()
  err := http.ListenAndServe(":"+port, nil)
  if err != nil {
    hlog.Error(err.Error())
  }
}
