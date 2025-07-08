package process

import (
	"os/exec"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/internal/logger"
)

// RestartPolicy defines how child processes should be restarted.
type RestartPolicy struct {
	MaxRetries    int           // -1 for unlimited retries
	Delay         time.Duration // delay before restarting
	RestartOnZero bool          // restart even if exit code is zero
}

// Manage starts, logs and restarts a process according to the given policy.
func Manage(name string, args []string, policy RestartPolicy) {
	go func() {
		retries := 0
		for {
			// setup logs
			stdoutW, stderrW, err := logger.SetupLog(name)
			if err != nil {
				hlog.Errorf("logger setup failed for %s: %v", name, err)
			}

			// start process
			hlog.Infof("Starting process: %s %v", name, args)
			cmd := exec.Command(name, args...)
			cmd.Stdout = stdoutW
			cmd.Stderr = stderrW
			if err := cmd.Start(); err != nil {
				hlog.Errorf("Failed to start process %s: %v", name, err)
				return
			}
			// register for control
			Register(name, cmd)
			pid := cmd.Process.Pid
			hlog.Infof("Process %s started with PID %d", name, pid)

			err = cmd.Wait()
			exitCode := cmd.ProcessState.ExitCode()
			if err != nil {
				hlog.Errorf("Process %s exited with error: %v (code %d)", name, err, exitCode)
			} else {
				hlog.Infof("Process %s exited normally (code %d)", name, exitCode)
			}

			// decide restart
			if exitCode == 0 && !policy.RestartOnZero {
				return
			}
			if policy.MaxRetries >= 0 && retries >= policy.MaxRetries {
				return
			}
			retries++
			hlog.Infof("Restarting %s in %s (retry %d)", name, policy.Delay, retries)
			time.Sleep(policy.Delay)
		}
	}()
}
