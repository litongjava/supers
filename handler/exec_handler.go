package handler

import (
  "context"
  "deploy-server/myutils"
  "github.com/cloudwego/hertz/pkg/app"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "github.com/hertz-contrib/sse"
  "io"
  "log"
  "os"
  "strings"
)

// ExecHandler handles the SSE and executes the command
func ExecHandler(ctx context.Context, reqCtx *app.RequestContext) {
  password := reqCtx.FormValue("p")
  cmd := reqCtx.FormValue("c")
  w := reqCtx.FormValue("w")
  env := reqCtx.FormValue("e")

  if string(password) != myutils.CONFIG.App.Password {
    reqCtx.String(401, "password is not correct")
    return
  }

  cmdStr := string(cmd)
  if cmdStr == "" {
    reqCtx.String(400, "c is empty")
    return
  }

  envVariables := []string{}
  if len(env) > 0 {
    split := strings.Split(string(env), ";")
    for i := range split {
      envVariables = append(envVariables, split[i])
    }
  }

  var workDir = string(w)
  // 检查目录是否存在
  if workDir != "" {
    if _, err := os.Stat(workDir); os.IsNotExist(err) {
      // 如果目录不存在，则创建目录
      err := os.MkdirAll(workDir, os.ModePerm)
      if err != nil {
        msg := "Failed to create work dir:" + err.Error()
        hlog.Error(msg)
        reqCtx.String(500, msg)
        return
      }
    }
  }

  s := sse.NewStream(reqCtx)

  // Create a pipe to capture command output
  reader, writer := io.Pipe()
  // Start a goroutine to read from the pipe and publish SSE events
  go func() {
    buf := make([]byte, 1024)
    for {
      n, err := reader.Read(buf)
      if n > 0 {
        bytes := buf[:n]
        hlog.Info("output:", string(bytes))
        event := &sse.Event{
          Event: "output",
          Data:  bytes,
        }
        _ = s.Publish(event)
      }
      if err != nil {
        if err == io.EOF {
          break
        }
        log.Println("Error reading from pipe:", err)
        break
      }
    }
  }()

  // Execute the command and write the output to the pipe
  err := myutils.ExecuteCommand(cmdStr, workDir, envVariables, writer)
  writer.Close() // Close the writer to signal the end of output

  if err != nil {
    event := &sse.Event{
      Event: "error",
      Data:  []byte(err.Error()),
    }
    _ = s.Publish(event)
  }
}
