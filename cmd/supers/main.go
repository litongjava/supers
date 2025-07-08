package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: supers <list|start|stop|status> [name]")
		return
	}
	cmd := os.Args[1]
	name := ""
	if len(os.Args) > 2 {
		name = os.Args[2]
	}
	conn, err := net.Dial("unix", "/var/run/super.sock")
	if err != nil {
		fmt.Println("Failed to connect to superd:", err)
		return
	}
	defer conn.Close()
	request := cmd
	if name != "" {
		request += " " + name
	}
	conn.Write([]byte(request))
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	fmt.Print(string(buf[:n]))
}
