package process

import (
	"os/exec"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// RestartPolicy defines how child processes should be restarted.
type RestartPolicy struct {
	MaxRetries    int           // -1 for unlimited retries
	Delay         time.Duration // delay before restarting
	RestartOnZero bool          // restart even if exit code is zero
}

// Manage starts and monitors a process according to the given policy.
func Manage(name string, args []string, policy RestartPolicy) {
	go func() {
		retries := 0
		for {
			hlog.Infof("Starting process: %s %v", name, args)
			cmd := exec.Command(name, args...)
			if err := cmd.Start(); err != nil {
				hlog.Errorf("Failed to start process %s: %v", name, err)
				return
			}
			pid := cmd.Process.Pid
			hlog.Infof("Process %s started with PID %d", name, pid)

			err := cmd.Wait()
			exitCode := cmd.ProcessState.ExitCode()
			if err != nil {
				hlog.Errorf("Process %s exited with error: %v (exit code %d)", name, err, exitCode)
			} else {
				hlog.Infof("Process %s exited normally with code %d", name, exitCode)
			}

			// decide whether to restart
			if exitCode == 0 && !policy.RestartOnZero {
				hlog.Infof("Process %s exited with zero and RestartOnZero=false; not restarting", name)
				return
			}
			if policy.MaxRetries >= 0 && retries >= policy.MaxRetries {
				hlog.Infof("Max retries reached for %s; not restarting", name)
				return
			}

			retries++
			hlog.Infof("Restarting process %s in %s (retry %d)", name, policy.Delay, retries)
			time.Sleep(policy.Delay)
		}
	}()
}
