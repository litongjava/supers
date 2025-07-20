// services/config.go
package services

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/internal/process"
)

// ServiceConfig holds one .service 文件解析后的信息
type ServiceConfig struct {
	Name             string
	Cmd              []string
	RestartPolicy    process.RestartPolicy
	WorkingDirectory string
	Env              []string
}

// 解析一行 Environment="FOO=bar" VAR2=baz
var envLineRe = regexp.MustCompile(`"([^"]+)"|(\S+)`)

// LoadConfigs 读取 dir 下所有 *.service
func LoadConfigs(dir string) (map[string]ServiceConfig, error) {
	pattern := filepath.Join(dir, "*.service")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	services := make(map[string]ServiceConfig)
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".service")
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		lines := strings.Split(string(b), "\n")

		var execStart, workDir string
		var envs []string
		restartDelay := 5 * time.Second

		for _, l := range lines {
			l = strings.TrimSpace(l)
			switch {
			case strings.HasPrefix(l, "ExecStart="):
				execStart = strings.TrimPrefix(l, "ExecStart=")
			case strings.HasPrefix(l, "WorkingDirectory="):
				workDir = strings.TrimPrefix(l, "WorkingDirectory=")
			case strings.HasPrefix(l, "RestartSec="):
				if d, err := time.ParseDuration(strings.TrimPrefix(l, "RestartSec=")); err == nil {
					restartDelay = d
				}
			case strings.HasPrefix(l, "Environment="):
				raw := strings.TrimPrefix(l, "Environment=")
				for _, m := range envLineRe.FindAllStringSubmatch(raw, -1) {
					kv := m[1]
					if kv == "" {
						kv = m[2]
					}
					envs = append(envs, kv)
				}
			}
		}

		if execStart == "" {
			continue
		}

		parts := strings.Fields(execStart)
		policy := process.RestartPolicy{
			MaxRetries:    -1,
			Delay:         restartDelay,
			RestartOnZero: false,
		}

		services[name] = ServiceConfig{
			Name:             name,
			Cmd:              parts,
			RestartPolicy:    policy,
			WorkingDirectory: workDir,
			Env:              envs,
		}
	}

	return services, nil
}

func LoadConfigFile(dir, name string) (ServiceConfig, error) {
	path := filepath.Join(dir, name+".service")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ServiceConfig{}, err
	}
	lines := strings.Split(string(data), "\n")

	var execStart, workDir string
	var envs []string
	restartDelay := 5 * time.Second

	for _, l := range lines {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, "ExecStart="):
			execStart = strings.TrimPrefix(l, "ExecStart=")
		case strings.HasPrefix(l, "WorkingDirectory="):
			workDir = strings.TrimPrefix(l, "WorkingDirectory=")
		case strings.HasPrefix(l, "RestartSec="):
			if d, err := time.ParseDuration(strings.TrimPrefix(l, "RestartSec=")); err == nil {
				restartDelay = d
			}
		case strings.HasPrefix(l, "Environment="):
			raw := strings.TrimPrefix(l, "Environment=")
			for _, m := range envLineRe.FindAllStringSubmatch(raw, -1) {
				kv := m[1]
				if kv == "" {
					kv = m[2]
				}
				envs = append(envs, kv)
			}
		}
	}

	if execStart == "" {
		return ServiceConfig{}, fmt.Errorf("no ExecStart in %s", path)
	}
	parts := strings.Fields(execStart)
	policy := process.RestartPolicy{-1, restartDelay, false}

	hlog.Infof("Loaded service %s: args=%v, workDir=%s, env=%v", name, parts, workDir, envs)
	return ServiceConfig{
		Name:             name,
		Cmd:              parts,
		RestartPolicy:    policy,
		WorkingDirectory: workDir,
		Env:              envs,
	}, nil
}
