package main

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
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

		fmt.Println("Handling connection")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	readBuf := make([]byte, 1024)
	// writeBuf := make([]byte, 1024)
	for n, err := conn.Read(readBuf); n != 0; n, err = conn.Read(readBuf) {
		if err != nil {
			return
		}

		commands, err := redis.ParseInput(readBuf[:n])
		if err != nil {
			fmt.Printf("Could not parse input: %q, %s \n", string(readBuf[:n]), err.Error())
			return
		}

		fmt.Printf("commands: %v\n", commands)
		for _, command := range commands {
			resp, err := redis.HandleCommand(command)
			if err != nil {
				fmt.Printf("command %v failed: %s \n", command, err.Error())
				return
			}

			fmt.Printf("writing: %q \n", resp)
			conn.Write([]byte(resp))
			// writeBuf = append(writeBuf, []byte(resp)...)
		}
	}

	// conn.Write(writeBuf)
	// fmt.Println("Finished with connection")
}
