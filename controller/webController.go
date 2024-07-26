package controller

import (
  "deploy-server/services"
  "deploy-server/utils"
  "encoding/base64"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "regexp"
)

func RegisterWebRouter() {
  http.HandleFunc("/deploy/web/", handleWeb)
}

func handleWeb(writer http.ResponseWriter, request *http.Request) {
  //log.Println(request.URL.Path)
  //验证密码
  var password = request.FormValue("p")
  if password != utils.CONFIG.App.Password {
    http.Error(writer, "passowrd is not correct", http.StatusBadRequest)
    return
  }
  pattern, _ := regexp.Compile(`/web/(.+)`)
  matches := pattern.FindStringSubmatch(request.URL.Path)
  if len(matches) > 0 {
    bytes, _ := base64.StdEncoding.DecodeString(matches[1])
    result := services.RunWrapperCommand(string(bytes))
    //encoder := jsonString.NewEncoder(writer)
    //err := encoder.Encode(result)
    //对返回内容使用base64加密
    jsonString, err := json.Marshal(result)
    if err != nil {
      log.Println("err", err.Error())
      _, err = fmt.Fprintln(writer, err.Error())
      if err != nil {
        return
      }
      return
    }
    base64String := base64.StdEncoding.EncodeToString([]byte(jsonString))
    _, err = fmt.Fprintln(writer, base64String)
    if err != nil {
      return
    }
  } else {
    log.Println("not find base64 string")
    writer.WriteHeader(http.StatusNotFound)
  }
}
