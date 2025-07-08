package services

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/litongjava/supers/internal/process"
)

// ServiceConfig holds one .service 文件解析后的信息
type ServiceConfig struct {
	Name             string
	Args             []string
	RestartPolicy    process.RestartPolicy
	WorkingDirectory string
}

// LoadConfigs 读取 dir 下所有 *.service，解析 ExecStart、Restart、RestartSec
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

		var execStart string
		// 默认重启延迟 5s
		restartDelay := 5 * time.Second

		for _, l := range lines {
			l = strings.TrimSpace(l)
			if strings.HasPrefix(l, "ExecStart=") {
				execStart = strings.TrimPrefix(l, "ExecStart=")
			} else if strings.HasPrefix(l, "Restart=") {
				if strings.Contains(l, "on-failure") {
					_ = true
				}
			} else if strings.HasPrefix(l, "RestartSec=") {
				val := strings.TrimPrefix(l, "RestartSec=")
				if d, err := time.ParseDuration(val); err == nil {
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
			Name:          name,
			Args:          parts[1:], // parts[0] 是可执行程序名
			RestartPolicy: policy,
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

	var execStart string
	restartDelay := 5 * time.Second

	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "ExecStart=") {
			execStart = strings.TrimPrefix(l, "ExecStart=")
		} else if strings.HasPrefix(l, "Restart=") {
		} else if strings.HasPrefix(l, "RestartSec=") {
			val := strings.TrimPrefix(l, "RestartSec=")
			if d, err := time.ParseDuration(val); err == nil {
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

	return ServiceConfig{
		Name:          name,
		Args:          parts[1:],
		RestartPolicy: policy,
	}, nil
}
