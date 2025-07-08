package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: supers <list|status|stop|start|reload> [name]")
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
	conn.Write([]byte(req))

	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	fmt.Print(string(buf[:n]))
}
