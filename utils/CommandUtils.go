package utils

import (
  "deploy-server/model"
  "os/exec"
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
    cmdResult.Error = err
    cmdResult.Success = false
  } else {
    cmdResult.Success = true
  }
  cmdResult.Output = (string(result))
  return cmdResult
}
