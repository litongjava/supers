package router

import "deploy-server/controller"

func RegisterRoutes() {
  controller.RegisterWebRouter()
  controller.RegisterFileRouter()
  controller.RegisterUnzipRouter()
  controller.RegisterStatusRouter()
}
