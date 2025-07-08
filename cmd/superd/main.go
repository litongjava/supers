package main

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/internal/process"
	"github.com/litongjava/supers/router"
	"github.com/litongjava/supers/utils"
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

	// 1. 加载配置
	port := strconv.Itoa(utils.CONFIG.App.Port)
	// 2. 启动演示进程：sleep
	if err := process.Start("sleep", "60"); err != nil {
		hlog.Errorf("Failed to start demo process: %v", err)
	}

	// 3. 启动 HTTP 服务
	for i := 1; i < len(os.Args); i += 2 {
		if os.Args[i] == "--port" {
			port = os.Args[i+1]
		}
	}
	hlog.Infof("start listen on: %s", port)
	router.RegisterRoutes()
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		hlog.Error(err.Error())
	}
}
