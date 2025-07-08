package router

import "github.com/litongjava/supers/controller"

func RegisterRoutes() {
	controller.RegisterWebRouter()
	controller.RegisterFileRouter()
	controller.RegisterUnzipRouter()
	controller.RegisterStatusRouter()
}
