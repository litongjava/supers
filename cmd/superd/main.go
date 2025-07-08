package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/internal/process"
	"github.com/litongjava/supers/router"
	"github.com/litongjava/supers/utils"
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

	// Load configuration
	port := strconv.Itoa(utils.CONFIG.App.Port)

	// Start and monitor demo process with restart policy
	policy := process.RestartPolicy{
		MaxRetries:    -1,
		Delay:         5 * time.Second,
		RestartOnZero: false,
	}
	process.Manage("sleep", []string{"60"}, policy)

	// Start HTTP server
	for i := 1; i < len(os.Args); i += 2 {
		if os.Args[i] == "--port" {
			port = os.Args[i+1]
		}
	}
	hlog.Infof("start listen on: %s", port)
	router.RegisterRoutes()
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		hlog.Error(err.Error())
	}
}
