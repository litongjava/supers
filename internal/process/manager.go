package process

import (
	"fmt"
	"os"
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

	// startTimes 记录每个服务的最近一次启动时间
	startTimes = make(map[string]time.Time)
	// commands 记录每个服务的完整命令行，用于摘要展示
	commands = make(map[string]string)
)

// Manage starts and monitors a named process with policy.
func Manage(name string, cmd []string, WorkingDirectory string, policy RestartPolicy, env []string) {
	go func() {
		retries := 0
		for {
			// 日志准备
			stdoutW, stderrW, err := logger.SetupLog(name)
			if err != nil {
				hlog.Errorf("logger setup failed for %s: %v", name, err)
			}

			// —— 自动识别可执行程序 ——
			program := name
			cmdArgs := cmd
			if len(cmd) > 0 && (strings.HasPrefix(cmd[0], "/") || strings.Contains(cmd[0], "/")) {
				program = cmd[0]
				cmdArgs = cmd[1:]
			}

			// 记录启动时间和命令行
			startTimes[name] = time.Now()
			fullCmd := program
			if len(cmdArgs) > 0 {
				fullCmd += " " + strings.Join(cmdArgs, " ")
			}
			commands[name] = fullCmd

			hlog.Infof("Starting %s %v", name, cmd)
			cmd := exec.Command(program, cmdArgs...)
			if len(env) > 0 {
				cmd.Env = append(os.Environ(), env...)
			}
			if WorkingDirectory != "" {
				cmd.Dir = WorkingDirectory
				workingDirs[name] = WorkingDirectory
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

			// 重启逻辑
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
	manualStop[name] = true
	return nil
}

// Status returns "running", "exited", "stopped" or "not found".
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

// Uptime returns a formatted uptime like "Up 9 hours".
func Uptime(name string) string {
	start, ok := startTimes[name]
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

// Command returns a truncated command summary, e.g. "java -jar target/ai…"
func Command(name string) string {
	cmd, ok := commands[name]
	if !ok {
		return ""
	}
	const maxLen = 20
	if len(cmd) <= maxLen {
		return cmd
	}
	return cmd[:maxLen] + "…"
}

// Command returns a truncated workingDir summary
func WorkingDir(name string) string {
	cmd, ok := workingDirs[name]
	if !ok {
		return ""
	}
	const maxLen = 20
	if len(cmd) <= maxLen {
		return cmd
	}
	return cmd[:maxLen] + "…"
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
