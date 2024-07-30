package main

import (
  "deploy-server/router"
  "deploy-server/utils"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "io"
  "net/http"
  "os"
  "strconv"
)

func InitLog() (*os.File, error) {
  hlog.SetLevel(hlog.LevelDebug)
  f, err := os.Create("app.log")
  if err != nil {
    return nil, err
  }
  fileWriter := io.MultiWriter(f, os.Stdout)
  hlog.SetOutput(fileWriter)
  return f, nil
}

func main() {
  logFile, err := InitLog()
  if err != nil {
    panic(err)
  }
  defer logFile.Close()

  port := strconv.Itoa(utils.CONFIG.App.Port)
  for i := 1; i < len(os.Args); i += 2 {
    param := os.Args[i]
    if param == "--port" {
      port = os.Args[i+1]
    }
  }
  hlog.Info("start listen on:", port)
  router.RegisterRoutes()
  err = http.ListenAndServe(":"+port, nil)
  if err != nil {
    hlog.Error(err.Error())
  }
}
