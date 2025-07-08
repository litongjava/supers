package main

import (
	"github.com/litongjava/supers/router"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/internal/process"
	"github.com/litongjava/supers/utils"
)

func InitLog() (*os.File, error) {
	f, err := os.Create("app.log")
	if err != nil {
		return nil, err
	}
	hlog.SetLevel(hlog.LevelDebug)
	hlog.SetOutput(io.MultiWriter(f, os.Stdout))
	return f, nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)
	n, _ := conn.Read(buf)
	fields := strings.Fields(string(buf[:n]))
	if len(fields) == 0 {
		return
	}
	switch fields[0] {
	case "list":
		for name, st := range process.List() {
			conn.Write([]byte(name + ": " + st + "\n"))
		}
	case "status":
		name := ""
		if len(fields) > 1 {
			name = fields[1]
		}
		if name == "" {
			// default to first
			for n := range process.List() {
				name = n
				break
			}
		}
		conn.Write([]byte(name + ": " + process.Status(name) + "\n"))
	case "stop":
		if len(fields) > 1 {
			conn.Write([]byte(process.Stop(fields[1]).Error() + "\n"))
		}
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

	// start demo
	policy := process.RestartPolicy{MaxRetries: -1, Delay: 5 * time.Second}
	process.Manage("sleep", []string{"60"}, policy)

	// socket
	sock := "/var/run/super.sock"
	os.Remove(sock)
	ln, _ := net.Listen("unix", sock)
	go func() {
		for {
			conn, _ := ln.Accept()
			go handleConn(conn)
		}
	}()

	// http
	hlog.Infof("HTTP on %s", port)
	handler := http.DefaultServeMux
	router.RegisterRoutes()
	http.ListenAndServe(":"+port, handler)
}
