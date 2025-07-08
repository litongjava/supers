package process

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"os/exec"
)

// Start launches a new process with given name and args, logs its PID and exit status.
func Start(name string, args ...string) error {
	hlog.Infof("Starting process: %s %v", name, args)
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	// Wait in background
	go func() {
		if err := cmd.Wait(); err != nil {
			hlog.Errorf("Process %s exited with error: %v", name, err)
		} else {
			hlog.Infof("Process %s exited normally", name)
		}
	}()
	hlog.Infof("Process %s started with PID %d", name, cmd.Process.Pid)
	return nil
}
