package router

import (
  "deploy-server/handler"
  "github.com/cloudwego/hertz/pkg/app/server"
)

//func RegisterRoutes() {
//  handler.RegisterWebRouter()
//  handler.RegisterFileRouter()
//  handler.RegisterUnzipRouter()
//  handler.RegisterStatusRouter()
//}

func RegisterHadlder(h *server.Hertz) {
  h.GET("/deploy/ping", handler.PingHandler)
  h.POST("/deploy/exec", handler.ExecHandler)
}
