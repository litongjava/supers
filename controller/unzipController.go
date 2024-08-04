package controller

import (
  "deploy-server/services"
  "deploy-server/utils"
  "encoding/json"
  "fmt"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "net/http"
  "os"
  "time"
)

func RegisterUnzipRouter() {
  http.HandleFunc("/deploy/file/upload-unzip/", handleUploadUnzip)
  http.HandleFunc("/deploy/file/upload-run/", handleUploadRun)
}

// 上传文件,放到指定目录,并运行脚本
func handleUploadRun(w http.ResponseWriter, r *http.Request) {
  startTime := time.Now().Unix()
  if r.Method != "POST" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }
  //验证密码
  var password = r.FormValue("p")
  if password != utils.CONFIG.App.Password {
    http.Error(w, "passowrd is not correct", http.StatusBadRequest)
    return
  }
  file, header, err := r.FormFile("file")
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }
  defer file.Close()

  //work dir
  workDir := r.FormValue("w")
  //将压缩包移动到的文件夹
  movedDir := r.FormValue("m")
  // 获取解压路径
  targetDir := r.FormValue("d")
  // 获取命令
  cmd1 := r.FormValue("c1")
  cmd2 := r.FormValue("c2")
  cmd3 := r.FormValue("c3")
  cmd4 := r.FormValue("c4")
  cmd5 := r.FormValue("c5")
  cmd6 := r.FormValue("c6")
  cmd7 := r.FormValue("c7")
  cmd8 := r.FormValue("c8")
  cmd9 := r.FormValue("c9")
  cmd := r.FormValue("c")
  if workDir != "" {
    _, err = os.Stat(workDir)
    if os.IsNotExist(err) {
      err := os.MkdirAll(workDir, os.ModePerm)
      if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
      }
    }
  }

  if movedDir == "" {
    hlog.Info("Not find m from request parameters")
  } else {
    b, err := utils.MoveFile(file, movedDir, header.Filename)
    if b {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
  }

  if targetDir == "" {
    hlog.Info("Not find d from request parameters")
  } else {
    length := r.ContentLength
    b, err := utils.ExtractFile(file, targetDir, length)
    if b {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
  }
  //执行命令
  if cmd1 == "" {
    hlog.Info("Not find c1 from request parameters")
  } else {
    hlog.Info("cmd1:", cmd1)
    result, err := services.RunWrapperCommand(workDir, cmd1)
    if err != nil {
      message := "cmd1 " + cmd1 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd2 == "" {
    hlog.Info("Not find c2 from request parameters")
  } else {
    hlog.Info("cmd2:", cmd2)
    result, err := services.RunWrapperCommand(workDir, cmd2)
    if err != nil {
      message := "cmd2 " + cmd2 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd3 == "" {
    hlog.Info("Not find c3 from request parameters")
  } else {
    hlog.Info("cmd3:", cmd3)
    result, err := services.RunWrapperCommand(workDir, cmd3)
    if err != nil {
      message := "cmd3 " + cmd3 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd4 == "" {
    hlog.Info("Not find c4 from request parameters")
  } else {
    hlog.Info("cmd4:", cmd4)
    result, err := services.RunWrapperCommand(workDir, cmd4)
    if err != nil {
      message := "cmd4 " + cmd4 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd5 == "" {
    hlog.Info("Not find c5 from request parameters")
  } else {
    hlog.Info("cmd5:", cmd5)
    result, err := services.RunWrapperCommand(workDir, cmd5)
    if err != nil {
      message := "cmd5 " + cmd5 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd6 == "" {
    hlog.Info("Not find c6 from request parameters")
  } else {
    hlog.Info("cmd6:", cmd6)
    result, err := services.RunWrapperCommand(workDir, cmd6)
    if err != nil {
      message := "cmd6 " + cmd6 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd7 == "" {
    hlog.Info("Not find c7 from request parameters")
  } else {
    hlog.Info("cmd7:", cmd7)
    result, err := services.RunWrapperCommand(workDir, cmd7)
    if err != nil {
      message := "cmd7 " + cmd7 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }
  if cmd8 == "" {
    hlog.Info("Not find c8 from request parameters")
  } else {
    hlog.Info("cmd8:", cmd8)
    result, err := services.RunWrapperCommand(workDir, cmd8)
    if err != nil {
      message := "cmd8 " + cmd8 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd9 == "" {
    hlog.Info("Not find c9 from request parameters")
  } else {
    hlog.Info("cmd9:", cmd9)
    result, err := services.RunWrapperCommand(workDir, cmd9)
    if err != nil {
      message := "cmd9 " + cmd9 + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd == "" {
    hlog.Info("Not find c from request parameters")
  } else {
    hlog.Info("cmd:", cmd)
    result, err := services.RunWrapperCommand(workDir, cmd)
    if err != nil {
      message := "cmd " + cmd + " output:" + result.Output + " err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
    endTime := time.Now().Unix()
    result.Time = endTime - startTime
    jsonBytes, err := json.Marshal(result)
    if err != nil {
      message := "json marshal err" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
    _, err = fmt.Fprintln(w, string(jsonBytes))
    if err != nil {
      message := "Failed to output response" + err.Error()
      hlog.Info(message)
      return
    }
  }

}

// 上传文件并解压
func handleUploadUnzip(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }
  //验证密码
  var password = r.FormValue("p")
  if password != utils.CONFIG.App.Password {
    http.Error(w, "passowrd is not correct", http.StatusBadRequest)
    return
  }
  file, _, err := r.FormFile("file")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer file.Close()

  // 获取解压路径
  targetDir := r.FormValue("d")
  if targetDir == "" {
    http.Error(w, "targetDir is required", http.StatusBadRequest)
    return
  }

  length := r.ContentLength
  b, err := utils.ExtractFile(file, targetDir, length)
  if b {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusOK)
  return
}
