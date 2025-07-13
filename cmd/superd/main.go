package main

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/internal/process"
	"github.com/litongjava/supers/internal/services"
	"github.com/litongjava/supers/router"
	"github.com/litongjava/supers/utils"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	serviceConfigs = make(map[string]services.ServiceConfig)
	configMutex    sync.Mutex
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

// loadAndManageAll 从 /etc/super/*.service 重新加载所有配置，
// 对比差异：新增 -> 启动；删除 -> 停止
func loadAndManageAll() error {
	const dir = "/etc/super"
	newConfigs, err := services.LoadConfigs(dir)
	if err != nil {
		return err
	}

	configMutex.Lock()
	defer configMutex.Unlock()

	// 停止已删除的
	for name := range serviceConfigs {
		if _, ok := newConfigs[name]; !ok {
			process.Stop(name)
		}
	}

	// 启动新增的
	for name, cfg := range newConfigs {
		if _, ok := serviceConfigs[name]; !ok {
			process.Manage(name, cfg.Cmd, cfg.WorkingDirectory, cfg.RestartPolicy)
		}
	}

	serviceConfigs = newConfigs
	return nil
}

// handleConn 增加 reload 和 start 命令
func handleConn(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			hlog.Error(err)
		}
	}(conn)
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
		for svc := range process.List() {
			status := process.Status(svc)
			uptime := process.Uptime(svc)
			cmdSummary := process.Command(svc)
			workingDirSummary := process.WorkingDir(svc)
			line := fmt.Sprintf("%s %s %s %s %s\n", svc, status, uptime, workingDirSummary, cmdSummary)
			conn.Write([]byte(line))
		}
	case "status":
		if name == "" {
			for svc := range process.List() {
				name = svc
				break
			}
		}
		status := process.Status(name)
		uptime := process.Uptime(name)
		cmdSummary := process.Command(name)
		workingDirSummary := process.WorkingDir(name)
		hlog.Infof("%s %s %s %s %s\n", name, status, uptime, workingDirSummary, cmdSummary)
		line := fmt.Sprintf("%s %s %s %s %s\n", name, status, uptime, workingDirSummary, cmdSummary)
		conn.Write([]byte(line))
	case "stop":
		stop(conn, name)
	case "start":
		start(conn, name)

	case "restart":
		stop(conn, name)
		start(conn, name)
	case "reload":
		if err := loadAndManageAll(); err != nil {
			conn.Write([]byte("error: reload failed: " + err.Error() + "\n"))
		} else {
			conn.Write([]byte("reloaded\n"))
		}
	default:
		conn.Write([]byte("error: unknown command\n"))
	}
}

func start(conn net.Conn, name string) {
	if name == "" {
		conn.Write([]byte("error: no service name\n"))
		return
	}
	configMutex.Lock()
	_, exists := serviceConfigs[name]
	configMutex.Unlock()
	// on-demand 加载新配置
	if !exists {
		cfg, err := services.LoadConfigFile("/etc/super", name)
		if err != nil {
			conn.Write([]byte("error: load config failed: " + err.Error() + "\n"))
			return
		}
		configMutex.Lock()
		serviceConfigs[name] = cfg
		configMutex.Unlock()
		process.Manage(name, cfg.Cmd, cfg.WorkingDirectory, cfg.RestartPolicy)
		conn.Write([]byte("started: " + name + "\n"))
		return
	}

	// 已有的直接启动
	cfg := serviceConfigs[name]
	process.Manage(name, cfg.Cmd, cfg.WorkingDirectory, cfg.RestartPolicy)
	conn.Write([]byte("started: " + name + "\n"))
	return
}

func stop(conn net.Conn, name string) {
	if name == "" {
		conn.Write([]byte("error: no service name\n"))
	} else if err := process.Stop(name); err != nil {
		conn.Write([]byte("error: " + err.Error() + "\n"))
	} else {
		conn.Write([]byte("stopped: " + name + "\n"))
	}
}

func main() {
	logFile, err := InitLog()
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	// 初始加载
	if err := loadAndManageAll(); err != nil {
		hlog.Errorf("initial load failed: %v", err)
	}

	// unix sock 服务
	sock := "/var/run/super.sock"
	os.Remove(sock)
	ln, _ := net.Listen("unix", sock)
	go func() {
		for {
			conn, _ := ln.Accept()
			go handleConn(conn)
		}
	}()

	// HTTP 控制接口
	port := strconv.Itoa(utils.CONFIG.App.Port)
	hlog.Infof("HTTP on %s", port)
	router.RegisterRoutes()
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		hlog.Error(err.Error())
	}
}
