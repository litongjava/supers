package myutils

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/model"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func RunComamnd(name string, arg ...string) model.CommandResult {
	start := time.Now().Unix()
	command := exec.Command(name, arg...)
	//会自动执行命令
	result, err := command.CombinedOutput()
	end := time.Now().Unix()

	cmdResult := model.CommandResult{}
	cmdResult.Time = end - start
	if err != nil {
		cmdResult.Err = err
		cmdResult.Success = false
	} else {
		cmdResult.Success = true
	}
	cmdResult.Output = (string(result))
	return cmdResult
}

// ExecuteCommand executes a command in the specified working directory and writes the output to the provided io.Writer
func ExecuteCommand(commandStr, workDir string, envVariables []string, outputWriter io.Writer) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", commandStr)
	} else {
		script, err := SaveAsScript(commandStr)
		if err != nil {
			msg := "Failed to save script:" + err.Error()
			hlog.Error(msg)
			return err
		}
		cmd = exec.Command("sh", script)
	}

	// Set the working directory
	cmd.Dir = workDir

	// Add environment variables to the command
	currEnv := os.Environ()
	for _, env := range envVariables {
		log.Println("Add env variable:", env)
		currEnv = append(currEnv, env)
	}
	cmd.Env = currEnv

	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter

	return cmd.Run()
}

func SaveAsScript(commandStr string) (string, error) {
	//创建一个文件
	hash := md5.New()
	hash.Write([]byte(commandStr))
	md5String := hex.EncodeToString(hash.Sum(nil))
	movedDir := "script"
	// 创建移动路径
	err := os.MkdirAll(movedDir, 0755)
	if err != nil {
		return "", err

	}
	//移动到这个文件夹
	dstFilePath := movedDir + "/" + md5String
	hlog.Info("dstFilePath:", dstFilePath)

	_, err = os.Stat(dstFilePath)
	if os.IsNotExist(err) {
		//创建文件
		scriptFile, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return "", err
		}
		defer scriptFile.Close()
		_, err = scriptFile.WriteString(commandStr)
		if err != nil {
			return "", err
		}
	} else {
		return "", err
	}

	return dstFilePath, nil
}
