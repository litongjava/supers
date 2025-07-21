package logger

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
)

// SetupLog prepares rotating log writers for a given service name.
func SetupLog(name string) (stdout io.Writer, stderr io.Writer, err error) {
	baseDir := "/etc/super/logs"
	dir := filepath.Join(baseDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, err
	}
	// 用一个 lumberjack.Logger 既写 stdout 又写 stderr
	combined := &lumberjack.Logger{
		Filename:   filepath.Join(dir, "combined.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     7, // days
		Compress:   true,
	}
	// 如果还想同时在控制台看到 stderr，可以这样：
	// stderr = io.MultiWriter(os.Stderr, combined)
	// stdout = combined
	stdout = combined
	stderr = combined
	return
}
