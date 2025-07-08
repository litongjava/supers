package main

import (
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/internal/process"
	"github.com/litongjava/supers/router"
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
	cmd := fields[0]
	name := ""
	if len(fields) > 1 {
		name = fields[1]
	}
	switch cmd {
	case "list":
		for svc, st := range process.List() {
			conn.Write([]byte(svc + ": " + st + "\n"))
		}
	case "status":
		if name == "" {
			for svc := range process.List() {
				name = svc
				break
			}
		}
		conn.Write([]byte(name + ": " + process.Status(name) + "\n"))
	case "stop":
		if name != "" {
			if err := process.Stop(name); err != nil {
				conn.Write([]byte("error: " + err.Error() + "\n"))
			} else {
				conn.Write([]byte("stopped: " + name + "\n"))
			}
		} else {
			conn.Write([]byte("error: no service name\n"))
		}
	default:
		conn.Write([]byte("error: unknown command\n"))
	}
}

func main() {
	logFile, err := InitLog()
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	port := strconv.Itoa(utils.CONFIG.App.Port)

	policy := process.RestartPolicy{MaxRetries: -1, Delay: 5 * time.Second}
	process.Manage("sleep", []string{"60"}, policy)

	sock := "/var/run/super.sock"
	os.Remove(sock)
	ln, _ := net.Listen("unix", sock)
	go func() {
		for {
			conn, _ := ln.Accept()
			go handleConn(conn)
		}
	}()

	hlog.Infof("HTTP on %s", port)
	router.RegisterRoutes()
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		hlog.Error(err.Error())
	}
}
