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
	out := &lumberjack.Logger{
		Filename:   filepath.Join(dir, "stdout.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     7, // days
		Compress:   true,
	}
	errOut := &lumberjack.Logger{
		Filename:   filepath.Join(dir, "stderr.log"),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     7,
		Compress:   true,
	}
	return io.MultiWriter(os.Stdout, out), io.MultiWriter(os.Stderr, errOut), nil
}
