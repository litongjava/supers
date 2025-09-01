package process

import (
  "fmt"
  "net"
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
}

func newRegistry() *registry {
  return &registry{
    procs:       make(map[string]*exec.Cmd),
    manualStop:  make(map[string]bool),
    workingDirs: make(map[string]string),
    startTimes:  make(map[string]time.Time),
    commands:    make(map[string]string),
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

// ---- process manager ----

func Manage(conn net.Conn, name string, cmd []string, WorkingDirectory string, policy RestartPolicy, env []string) {
  go func() {
    retries := 0
    for {
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

      // record metadata (locked)
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
        reg.setWorkingDir(name, WorkingDirectory)
      }

      c.Stdout = stdoutW
      c.Stderr = stderrW
      if err := c.Start(); err != nil {
        if conn != nil {
          errMsg := "failed: " + name + " err=" + err.Error()
          hlog.Error(errMsg)
          _, _ = conn.Write([]byte(errMsg + "\n"))
        }
        return
      }

      if conn != nil {
        _, _ = conn.Write([]byte("started: " + name + "\n"))
      }

      reg.setProc(name, c)
      hlog.Infof("%s PID=%d", name, c.Process.Pid)

      err = c.Wait()
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

      if shouldRestart(exitCode, retries, policy) {
        events.Emit(events.Event{Name: name, ExitCode: exitCode, Type: events.EventProcessRestarted})
        retries++
        time.Sleep(policy.Delay)
        continue
      }
      if exitCode == 0 && !policy.RestartOnZero {
        return
      }
      if policy.MaxRetries >= 0 && retries >= policy.MaxRetries {
        return
      }

      retries++
      hlog.Infof("restart %s in %s (retry %d)", name, policy.Delay, retries)
      time.Sleep(policy.Delay)
    }
  }()
}

func Stop(name string) error {
  cmd, ok := reg.getProc(name)
  if !ok || cmd.Process == nil {
    return fmt.Errorf("no process: %s", name)
  }
  if err := cmd.Process.Kill(); err != nil {
    return err
  }
  reg.setManualStop(name, true)
  return nil
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
    statuses[name] = Status(name) // Status has its own locks
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
