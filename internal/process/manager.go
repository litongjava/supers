package process

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/internal/events"
	"github.com/litongjava/supers/internal/logger"
)

// RestartPolicy defines restart behavior for a process.
type RestartPolicy struct {
	MaxRetries    int           // -1 for unlimited
	Delay         time.Duration // delay before restart
	RestartOnZero bool          // restart on exit code 0
}

var (
	// procs holds active commands by name
	procs = make(map[string]*exec.Cmd)
	// manualStop 标记手动停止过的服务
	manualStop = make(map[string]bool)
	// workingDirs 保存每个服务的工作目录
	workingDirs = make(map[string]string)
)

// SetWorkingDir 为某个服务设置启动时的工作目录
func SetWorkingDir(name, dir string) {
	workingDirs[name] = dir
}

// Manage starts and monitors a named process with policy.
func Manage(name string, args []string, WorkingDirectory string, policy RestartPolicy) {
	go func() {
		retries := 0
		for {
			// 日志准备
			stdoutW, stderrW, err := logger.SetupLog(name)
			if err != nil {
				hlog.Errorf("logger setup failed for %s: %v", name, err)
			}

			// —— 新增：从 args[0] 自动识别可执行程序 ——
			program := name
			cmdArgs := args
			if len(args) > 0 && (strings.HasPrefix(args[0], "/") || strings.Contains(args[0], "/")) {
				program = args[0]
				cmdArgs = args[1:]
			}
			hlog.Infof("Starting %s %v", name, args)
			cmd := exec.Command(program, cmdArgs...)

			if WorkingDirectory != "" {
				cmd.Dir = WorkingDirectory
			}

			cmd.Stdout = stdoutW
			cmd.Stderr = stderrW
			if err := cmd.Start(); err != nil {
				hlog.Errorf("start %s failed: %v", name, err)
				return
			}

			// 注册并记录 PID
			procs[name] = cmd
			hlog.Infof("%s PID=%d", name, cmd.Process.Pid)

			// 等待退出
			err = cmd.Wait()
			exitCode := cmd.ProcessState.ExitCode()
			events.Emit(events.Event{Name: name, ExitCode: exitCode, Type: events.EventProcessExited})

			msg := fmt.Sprintf("%s exited code %d", name, exitCode)
			if err != nil {
				hlog.Errorf(msg)
			} else {
				hlog.Infof(msg)
			}

			if manualStop[name] {
				hlog.Infof("Process %s was manually stopped; skipping restart", name)
				return
			}

			// 重启与退出逻辑保持不变……
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

// Stop kills the named process.
func Stop(name string) error {
	cmd, ok := procs[name]
	if !ok || cmd.Process == nil {
		return fmt.Errorf("no process: %s", name)
	}
	if err := cmd.Process.Kill(); err != nil {
		return err
	}
	// 标记一下，下次 Status 就能识别
	manualStop[name] = true
	return nil
}

// Status returns "running" or "exited" or "not found".
func Status(name string) string {
	if manualStop[name] {
		return "stopped"
	}
	cmd, ok := procs[name]
	if !ok {
		return "not found"
	}
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return "exited"
	}
	return "running"
}

// List returns map of names to statuses.
func List() map[string]string {
	statuses := make(map[string]string)
	for name := range procs {
		statuses[name] = Status(name)
	}
	return statuses
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
