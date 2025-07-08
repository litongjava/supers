package services

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
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
}

// LoadConfigs 读取 dir 下所有 *.service，解析 ExecStart、WorkingDirectory、Restart、RestartSec
func LoadConfigs(dir string) (map[string]ServiceConfig, error) {
	pattern := filepath.Join(dir, "*.service")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	services := make(map[string]ServiceConfig)
	for _, file := range files {
		name := filepath.Base(file)
		name = strings.TrimSuffix(name, ".service")

		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		lines := strings.Split(string(b), "\n")

		var execStart, workDir string
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
			}
		}

		if execStart == "" {
			// 跳过没有 ExecStart 的文件
			continue
		}

		parts := strings.Fields(execStart)
		policy := process.RestartPolicy{
			MaxRetries:    -1,           // 无限重试
			Delay:         restartDelay, // 重启延迟
			RestartOnZero: false,        // 退出码 0 不重启
		}

		services[name] = ServiceConfig{
			Name:             name,
			Cmd:              parts,
			RestartPolicy:    policy,
			WorkingDirectory: workDir,
		}
	}

	return services, nil
}

// LoadConfigFile 仅加载单个 /etc/super/<name>.service，用于 on-demand start
func LoadConfigFile(dir, name string) (ServiceConfig, error) {
	path := filepath.Join(dir, name+".service")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ServiceConfig{}, err
	}
	lines := strings.Split(string(data), "\n")

	var execStart, workDir string
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
		}
	}

	if execStart == "" {
		return ServiceConfig{}, fmt.Errorf("no ExecStart in %s", path)
	}

	parts := strings.Fields(execStart)
	policy := process.RestartPolicy{
		MaxRetries:    -1,
		Delay:         restartDelay,
		RestartOnZero: false,
	}

	hlog.Infof("Loaded service %s: args=%v, workDir=%s", name, parts, workDir)
	return ServiceConfig{
		Name:             name,
		Cmd:              parts,
		RestartPolicy:    policy,
		WorkingDirectory: workDir,
	}, nil
}
