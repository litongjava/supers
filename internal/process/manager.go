package process

import (
	"fmt"
	"github.com/litongjava/supers/internal/events"
	"github.com/litongjava/supers/internal/logger"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
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
)
var manualStop = make(map[string]bool)

// Manage starts and monitors a named process with policy.
func Manage(name string, args []string, policy RestartPolicy) {
	go func() {
		retries := 0
		for {
			// prepare logs directory
			// assume logger.SetupLog creates dirs and returns writers
			stdoutW, stderrW, err := logger.SetupLog(name)
			if err != nil {
				hlog.Errorf("logger setup failed for %s: %v", name, err)
			}

			// start
			hlog.Infof("Starting %s %v", name, args)
			cmd := exec.Command(name, args...)
			cmd.Stdout = stdoutW
			cmd.Stderr = stderrW
			if err := cmd.Start(); err != nil {
				hlog.Errorf("start %s failed: %v", name, err)
				return
			}
			// register
			procs[name] = cmd
			pid := cmd.Process.Pid
			hlog.Infof("%s PID=%d", name, pid)

			err = cmd.Wait()
			exitCode := cmd.ProcessState.ExitCode()
			e := events.Event{Name: name, ExitCode: exitCode, Type: events.EventProcessExited}
			events.Emit(e)

			msg := strings.Join([]string{name, "exited", "code", string(exitCode)}, " ")
			if err != nil {
				hlog.Errorf(msg)
			} else {
				hlog.Infof(msg)
			}

			if shouldRestart(exitCode, retries, policy) {
				re := events.Event{Name: name, ExitCode: exitCode, Type: events.EventProcessRestarted}
				events.Emit(re)
				retries++
				time.Sleep(policy.Delay)
				continue
			}
			if manualStop[name] {
				hlog.Infof("Process %s was manually stopped; skipping restart", name)
				return
			}

			// decide restart
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
