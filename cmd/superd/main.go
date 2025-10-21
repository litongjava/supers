package main

import (
  "fmt"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "github.com/litongjava/supers/internal/events"
  "github.com/litongjava/supers/internal/process"
  "github.com/litongjava/supers/internal/services"
  "github.com/litongjava/supers/router"
  "github.com/litongjava/supers/utils"
  "io"
  "net"
  "net/http"
  "os"
  "path/filepath"
  "strconv"
  "strings"
  "sync"
  "time"
)

var (
  serviceConfigs = make(map[string]services.ServiceConfig)
  configMutex    sync.Mutex
)

const dir = "/etc/super"
const sock = "/var/run/super.sock"

func ensureDir(path string, perm os.FileMode) error {
  info, err := os.Stat(path)
  if err == nil {
    if !info.IsDir() {
      return fmt.Errorf("%s exists but is not a directory", path)
    }
    return nil
  }
  if os.IsNotExist(err) {
    return os.MkdirAll(path, perm)
  }
  return err
}

func ensureParentDir(filePath string, perm os.FileMode) error {
  // 确保用于 unix socket 的父目录存在（例如 /var/run）
  parent := filepath.Dir(filePath)
  return ensureDir(parent, perm)
}

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
      process.Manage(name, cfg.Cmd, cfg.WorkingDirectory, cfg.RestartPolicy, cfg.Env)
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
    for svc := range serviceConfigs {
      status := process.Status(svc)
      uptime := process.Uptime(svc)
      cmdSummary := process.Command(svc)
      workDir := process.WorkingDir(svc)
      line := fmt.Sprintf("%s %s %s %s %s\n",
        svc, status, uptime, workDir, cmdSummary)
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
    if name == "" {
      conn.Write([]byte("error: no service name\n"))
      return
    }

    // Stop() 会阻塞直到进程真正退出
    if err := process.Stop(name); err != nil {
      conn.Write([]byte("error: stop failed: " + err.Error() + "\n"))
      return
    }
    conn.Write([]byte("stopped: " + name + "\n"))

    // 此时可以安全地启动新进程
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
    fmt.Fprintln(conn, "error: no service name")
    return
  }

  // 加载/确认 cfg
  configMutex.Lock()
  cfg, exists := serviceConfigs[name]
  configMutex.Unlock()

  if !exists {
    c, err := services.LoadConfigFile("/etc/super", name)
    if err != nil {
      fmt.Fprintf(conn, "error: load config failed: %v\n", err)
      return
    }
    configMutex.Lock()
    serviceConfigs[name] = c
    configMutex.Unlock()
    cfg = c
  }

  // 订阅一次性事件（本次调用专属）
  startedCh := events.SubscribeOnce(name, events.EventProcessStarted)
  failedCh := events.SubscribeOnce(name, events.EventProcessStartFailed)

  // 触发启动
  process.Manage(name, cfg.Cmd, cfg.WorkingDirectory, cfg.RestartPolicy, cfg.Env)

  // 等待结果/超时
  select {
  case ev := <-startedCh:
    hlog.Infof("started: %s PID=%d", ev.Name, ev.PID)
    fmt.Fprintf(conn, "started: %s PID=%d\n", ev.Name, ev.PID)
  case ev := <-failedCh:
    hlog.Errorf("failed: %s error=%s", ev.Name, ev.Error)
    fmt.Fprintf(conn, "failed: %s error=%s\n", ev.Name, ev.Error)
  case <-time.After(5 * time.Second):
    hlog.Infof("start pending: %s", name)
    fmt.Fprintf(conn, "start pending: %s\n", name)
  }
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

  // 1) 确保 /etc/super 目录存在（缺失则创建）
  if err := ensureDir(dir, 0o755); err != nil {
    hlog.Fatalf("ensure dir %s failed: %v", dir, err)
  }
  // 初始加载
  if err := loadAndManageAll(); err != nil {
    hlog.Errorf("initial load failed: %v", err)
  }

  // unix sock 服务
  // 2) 确保 socket 的父目录存在（通常 /var/run）
  if err := ensureParentDir(sock, 0o755); err != nil {
    hlog.Fatalf("ensure parent dir for socket failed: %v", err)
  }
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
