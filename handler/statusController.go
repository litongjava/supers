package handler

import (
  "fmt"
  "net/http"
)

func RegisterStatusRouter() {
  http.HandleFunc("/deploy/status", handleStatus)
}

func handleStatus(writer http.ResponseWriter, request *http.Request) {
  fmt.Fprintln(writer, "OK")
}
