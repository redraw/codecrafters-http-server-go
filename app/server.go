package main

import (
	"fmt"
	"net"
	"os"
)

const (
	PORT = 4221
)

func main() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Listening on port:", PORT)

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}
