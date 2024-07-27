package handler

import (
  "deploy-server/myutils"
  "deploy-server/services"
  "encoding/json"
  "fmt"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "net/http"
  "os"
)

func RegisterUnzipRouter() {
  http.HandleFunc("/deploy/file/upload-unzip/", handleUploadUnzip)
  http.HandleFunc("/deploy/file/upload-run/", handleUploadRun)
}

// 上传文件,放到指定目录,并运行脚本
func handleUploadRun(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }
  //验证密码
  var password = r.FormValue("p")
  if password != myutils.CONFIG.App.Password {
    http.Error(w, "passowrd is not correct", http.StatusBadRequest)
    return
  }
  file, header, err := r.FormFile("file")
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }
  defer file.Close()

  //将压缩包移动到的文件夹
  movedDir := r.FormValue("m")
  // 获取解压路径
  targetDir := r.FormValue("d")
  // 获取命令
  workDir := r.FormValue("w")

  // 检查目录是否存在
  if _, err := os.Stat(workDir); os.IsNotExist(err) {
    // 如果目录不存在，则创建目录
    err := os.MkdirAll(workDir, os.ModePerm)
    if err != nil {
      fmt.Println("Failed to create dir:", err)
    }
  }
  cmd1 := r.FormValue("c1")
  cmd2 := r.FormValue("c2")
  cmd3 := r.FormValue("c3")
  cmd := r.FormValue("c")

  if movedDir == "" {
    hlog.Info("Not find m from request parameters")
  } else {
    b, err := myutils.MoveFile(file, movedDir, header.Filename)
    if b {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
  }

  if targetDir == "" {
    hlog.Info("Not find d from request parameters")
  } else {
    length := r.ContentLength
    b, err := myutils.ExtractFile(file, targetDir, length)
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
    result := services.RunWrapperCommand(cmd1)
    if result.Error != nil {
      message := "cmd2 err:" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd2 == "" {
    hlog.Info("Not find c1 from request parameters")
  } else {
    hlog.Info("cmd2:", cmd2)
    result := services.RunWrapperCommand(cmd2)
    if result.Error != nil {
      message := "cmd2 err:" + result.Error.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd3 == "" {
    hlog.Info("Not find c1 from request parameters")
  } else {
    hlog.Info("cmd3:", cmd3)
    result := services.RunWrapperCommand(cmd3)
    if result.Error != nil {
      message := "cmd3 err" + result.Error.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
  }

  if cmd == "" {
    hlog.Info("Not find c from request parameters")
  } else {
    hlog.Info("cmd:", cmd)
    result := services.RunWrapperCommand(cmd)
    if result.Error != nil {
      message := "cmd err" + result.Error.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
    jsonBytes, err := json.Marshal(result)
    if err != nil {
      message := "json marshal err" + err.Error()
      hlog.Info(message)
      http.Error(w, message, http.StatusInternalServerError)
      return
    }
    _, err = fmt.Fprintln(w, string(jsonBytes))
    if err != nil {
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
  if password != myutils.CONFIG.App.Password {
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
  b, err := myutils.ExtractFile(file, targetDir, length)
  if b {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusOK)
  return
}
