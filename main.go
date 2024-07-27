package main

import (
  "deploy-server/router"
  "deploy-server/utils"
  "github.com/cloudwego/hertz/pkg/app/server"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "io"
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
  ports := server.WithHostPorts("0.0.0.0:" + port)
  h := server.Default(ports)
  router.RegisterHadlder(h)
  h.Spin()
}
