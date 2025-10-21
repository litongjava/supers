package process

import (
  "fmt"
  "os"
  "os/exec"
  "strings"
  "sync"
  "time"

  "github.com/cloudwego/hertz/pkg/common/hlog"
  "github.com/litongjava/supers/internal/events"
  "github.com/litongjava/supers/internal/logger"
)

type RestartPolicy struct {
  MaxRetries    int
  Delay         time.Duration
  RestartOnZero bool
}

// ---- concurrency-safe registry ----

type registry struct {
  mu          sync.RWMutex
  procs       map[string]*exec.Cmd
  manualStop  map[string]bool
  workingDirs map[string]string
  startTimes  map[string]time.Time
  commands    map[string]string
  policies    map[string]RestartPolicy // ⭐ 新增:保存重启策略
  cmdLines    map[string][]string      // ⭐ 新增:保存命令行
  envs        map[string][]string      // ⭐ 新增:保存环境变量
}

func newRegistry() *registry {
  return &registry{
    procs:       make(map[string]*exec.Cmd),
    manualStop:  make(map[string]bool),
    workingDirs: make(map[string]string),
    startTimes:  make(map[string]time.Time),
    commands:    make(map[string]string),
    policies:    make(map[string]RestartPolicy),
    cmdLines:    make(map[string][]string),
    envs:        make(map[string][]string),
  }
}

var reg = newRegistry()

// helpers
func (r *registry) setProc(name string, cmd *exec.Cmd) {
  r.mu.Lock()
  r.procs[name] = cmd
  r.mu.Unlock()
}
func (r *registry) getProc(name string) (*exec.Cmd, bool) {
  r.mu.RLock()
  cmd, ok := r.procs[name]
  r.mu.RUnlock()
  return cmd, ok
}
func (r *registry) setManualStop(name string, v bool) {
  r.mu.Lock()
  r.manualStop[name] = v
  r.mu.Unlock()
}
func (r *registry) isManualStop(name string) bool {
  r.mu.RLock()
  v := r.manualStop[name]
  r.mu.RUnlock()
  return v
}
func (r *registry) setWorkingDir(name, dir string) {
  r.mu.Lock()
  r.workingDirs[name] = dir
  r.mu.Unlock()
}
func (r *registry) getWorkingDir(name string) (string, bool) {
  r.mu.RLock()
  d, ok := r.workingDirs[name]
  r.mu.RUnlock()
  return d, ok
}
func (r *registry) setStartTime(name string, t time.Time) {
  r.mu.Lock()
  r.startTimes[name] = t
  r.mu.Unlock()
}
func (r *registry) getStartTime(name string) (time.Time, bool) {
  r.mu.RLock()
  t, ok := r.startTimes[name]
  r.mu.RUnlock()
  return t, ok
}
func (r *registry) setCommand(name, cmd string) {
  r.mu.Lock()
  r.commands[name] = cmd
  r.mu.Unlock()
}
func (r *registry) getCommand(name string) (string, bool) {
  r.mu.RLock()
  c, ok := r.commands[name]
  r.mu.RUnlock()
  return c, ok
}
func (r *registry) snapshotProcNames() []string {
  r.mu.RLock()
  names := make([]string, 0, len(r.procs))
  for n := range r.procs {
    names = append(names, n)
  }
  r.mu.RUnlock()
  return names
}
func (r *registry) setMetadata(name string, cmdLine []string, workDir string, policy RestartPolicy, env []string) {
  r.mu.Lock()
  r.cmdLines[name] = cmdLine
  r.workingDirs[name] = workDir
  r.policies[name] = policy
  r.envs[name] = env
  r.mu.Unlock()
}
func (r *registry) getMetadata(name string) ([]string, string, RestartPolicy, []string, bool) {
  r.mu.RLock()
  cmd := r.cmdLines[name]
  dir := r.workingDirs[name]
  policy := r.policies[name]
  env := r.envs[name]
  ok := len(cmd) > 0
  r.mu.RUnlock()
  return cmd, dir, policy, env, ok
}

// ---- process manager ----

// Manage 同步启动进程,返回 PID 或错误
func Manage(name string, cmd []string, WorkingDirectory string, policy RestartPolicy, env []string) (int, error) {
  // 保存元数据供重启使用
  reg.setMetadata(name, cmd, WorkingDirectory, policy, env)

  // 清除手动停止标志(如果是重启)
  reg.setManualStop(name, false)

  stdoutW, stderrW, err := logger.SetupLog(name)
  if err != nil {
    hlog.Errorf("logger setup failed for %s: %v", name, err)
  }

  program := name
  cmdArgs := cmd
  if len(cmd) > 0 && (strings.HasPrefix(cmd[0], "/") || strings.Contains(cmd[0], "/")) {
    program = cmd[0]
    cmdArgs = cmd[1:]
  }

  // record metadata
  reg.setStartTime(name, time.Now())
  fullCmd := program
  if len(cmdArgs) > 0 {
    fullCmd += " " + strings.Join(cmdArgs, " ")
  }
  reg.setCommand(name, fullCmd)

  hlog.Infof("Starting %s %v", name, cmd)
  c := exec.Command(program, cmdArgs...)
  if len(env) > 0 {
    c.Env = append(os.Environ(), env...)
  }
  if WorkingDirectory != "" {
    c.Dir = WorkingDirectory
  }

  c.Stdout = stdoutW
  c.Stderr = stderrW

  // ⭐ 同步启动
  if err := c.Start(); err != nil {
    errMsg := "failed: " + name + " err=" + err.Error()
    hlog.Error(errMsg)
    events.Emit(events.Event{
      Name:  name,
      Type:  events.EventProcessStartFailed,
      Error: err.Error(),
    })
    return 0, err
  }

  pid := c.Process.Pid
  events.Emit(events.Event{
    Name: name,
    Type: events.EventProcessStarted,
    PID:  pid,
  })
  reg.setProc(name, c)
  hlog.Infof("%s PID=%d", name, pid)

  // ⭐ 异步监控进程
  go monitorProcess(name, c, 0)

  return pid, nil
}

// monitorProcess 监控进程退出并处理重启
func monitorProcess(name string, c *exec.Cmd, retries int) {
  err := c.Wait()
  exitCode := c.ProcessState.ExitCode()
  events.Emit(events.Event{Name: name, ExitCode: exitCode, Type: events.EventProcessExited})

  msg := fmt.Sprintf("%s exited code %d", name, exitCode)
  if err != nil {
    hlog.Errorf(msg)
  } else {
    hlog.Infof(msg)
  }

  if reg.isManualStop(name) {
    hlog.Infof("Process %s was manually stopped; skipping restart", name)
    return
  }

  // 获取保存的元数据
  cmd, workDir, policy, env, ok := reg.getMetadata(name)
  if !ok {
    hlog.Errorf("Cannot restart %s: metadata not found", name)
    return
  }

  if !shouldRestart(exitCode, retries, policy) {
    return
  }

  events.Emit(events.Event{Name: name, ExitCode: exitCode, Type: events.EventProcessRestarted})
  retries++
  hlog.Infof("restart %s in %s (retry %d)", name, policy.Delay, retries)
  time.Sleep(policy.Delay)

  // 重启
  _, err = Manage(name, cmd, workDir, policy, env)
  if err != nil {
    hlog.Errorf("restart %s failed: %v", name, err)
  }
}

func Stop(name string) error {
  cmd, ok := reg.getProc(name)
  if !ok || cmd.Process == nil {
    return fmt.Errorf("no process: %s", name)
  }

  // 订阅退出事件
  exitedCh := events.SubscribeOnce(name, events.EventProcessExited)

  // 设置手动停止标志
  reg.setManualStop(name, true)

  // 发送 SIGKILL
  if err := cmd.Process.Kill(); err != nil {
    return err
  }

  // 等待退出事件(最多 5 秒)
  select {
  case <-exitedCh:
    hlog.Infof("Process %s exited", name)
    return nil
  case <-time.After(5 * time.Second):
    return fmt.Errorf("timeout waiting for %s to exit", name)
  }
}

func Status(name string) string {
  if reg.isManualStop(name) {
    return "stopped"
  }
  cmd, ok := reg.getProc(name)
  if !ok {
    return "not found"
  }
  if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
    return "exited"
  }
  return "running"
}

func List() map[string]string {
  statuses := make(map[string]string)
  for _, name := range reg.snapshotProcNames() {
    statuses[name] = Status(name)
  }
  return statuses
}

func Uptime(name string) string {
  start, ok := reg.getStartTime(name)
  if !ok {
    return ""
  }
  d := time.Since(start)
  h := int(d.Hours())
  if h > 0 {
    return fmt.Sprintf("Up %d hours", h)
  }
  m := int(d.Minutes())
  return fmt.Sprintf("Up %d minutes", m)
}

func Command(name string) string {
  cmd, ok := reg.getCommand(name)
  if !ok {
    return ""
  }
  const maxLen = 20
  if len(cmd) <= maxLen {
    return cmd
  }
  return cmd[:maxLen] + "…"
}

func WorkingDir(name string) string {
  d, ok := reg.getWorkingDir(name)
  if !ok {
    return ""
  }
  const maxLen = 20
  if len(d) <= maxLen {
    return d
  }
  return d[:maxLen] + "…"
}

func shouldRestart(exitCode, retries int, policy RestartPolicy) bool {
  if exitCode == 0 && !policy.RestartOnZero {
    return false
  }
  if policy.MaxRetries >= 0 && retries >= policy.MaxRetries {
    return false
  }
  return true
}
