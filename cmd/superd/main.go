package main

import (
	"io"
	"net"
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
	hlog.SetOutput(io.MultiWriter(f, os.Stdout))
	return f, nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	req := string(buf[:n])
	switch req {
	case "list":
		for name, st := range process.List() {
			conn.Write([]byte(name + ": " + st + "\n"))
		}
	case "status":
		conn.Write([]byte(process.Status("sleep") + "\n"))
		// add parse for start/stop etc...
	}
}

func main() {
	logFile, err := InitLog()
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	// load config
	port := strconv.Itoa(utils.CONFIG.App.Port)

	// manage demo process
	policy := process.RestartPolicy{MaxRetries: -1, Delay: 5 * time.Second}
	process.Manage("sleep", []string{"60"}, policy)

	// unix sock listener
	sock := "/var/run/super.sock"
	os.Remove(sock)
	ln, _ := net.Listen("unix", sock)
	go func() {
		for {
			conn, _ := ln.Accept()
			go handleConn(conn)
		}
	}()

	// http API
	hlog.Infof("start http on %s", port)
	router.RegisterRoutes()
	http.ListenAndServe(":"+port, nil)
}
