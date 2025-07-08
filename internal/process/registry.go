package process

import (
	"fmt"
	"os/exec"
	"sync"
)

var (
	mu    sync.RWMutex
	procs = make(map[string]*exec.Cmd)
)

func Register(name string, cmd *exec.Cmd) {
	mu.Lock()
	defer mu.Unlock()
	procs[name] = cmd
}

func Stop(name string) error {
	mu.RLock()
	cmd, ok := procs[name]
	mu.RUnlock()
	if !ok || cmd.Process == nil {
		return fmt.Errorf("no such process: %s", name)
	}
	return cmd.Process.Kill()
}

func Status(name string) string {
	mu.RLock()
	cmd, ok := procs[name]
	mu.RUnlock()
	if !ok {
		return "not running"
	}
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return "exited"
	}
	return "running"
}

func List() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	status := make(map[string]string)
	for name := range procs {
		status[name] = Status(name)
	}
	return status
}
