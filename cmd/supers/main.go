package main

import (
  "fmt"
  "io"
  "net"
  "os"
)

func main() {
  if len(os.Args) < 2 {
    fmt.Println("Usage: supers <cmd> [name]")
    return
  }
  cmd := os.Args[1]
  name := ""
  if len(os.Args) > 2 {
    name = os.Args[2]
  }

  conn, err := net.Dial("unix", "/var/run/super.sock")
  if err != nil {
    fmt.Println("connect error:", err)
    return
  }
  defer conn.Close()

  req := cmd
  if name != "" {
    req += " " + name
  }
  if _, err := conn.Write([]byte(req)); err != nil {
    fmt.Println("write error:", err)
    return
  }

  // 方案 A：直接拷贝到 stdout（一直读到 EOF）
  if _, err := io.Copy(os.Stdout, conn); err != nil && err != io.EOF {
    fmt.Println("read error:", err)
  }
}
