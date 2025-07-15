package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer l.Close()

	for conn, err := l.Accept(); true; conn, err = l.Accept() {
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handlePing(conn)
	}
}

func handlePing(conn net.Conn) {
	defer conn.Close()

	readBuf := make([]byte, 1024)
	pong := "+PONG\r\n"

	for n, err := conn.Read(readBuf); n != 0; n, err = conn.Read(readBuf) {
		if err != nil {
			return
		}

		count := strings.Count(string(readBuf), "PING")
		conn.Write([]byte(strings.Repeat(pong, count)))
	}
}
