package utils

import (
  "deploy-server/model"
  "os/exec"
  "time"
)

func RunComamnd(workDir string, name string, arg ...string) (model.CommandResult, error) {
  start := time.Now().Unix()
  command := exec.Command(name, arg...)
  command.Dir = workDir
  //会自动执行命令
  bytes, err := command.CombinedOutput()
  end := time.Now().Unix()

  cmdResult := model.CommandResult{}
  cmdResult.Time = end - start
  cmdResult.Output = (string(bytes))
  if err != nil {
    cmdResult.Success = false
    return cmdResult, err
  } else {
    cmdResult.Success = true
  }
  return cmdResult, nil
}
