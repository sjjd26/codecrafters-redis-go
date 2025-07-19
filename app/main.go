package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/redis/parser"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
)

var inputChannelQueue = make(chan chan []byte, 10)

func main() {
	initConfig()

	go listen()

	// main event loop
	for {
		// fmt.Println("event loop waiting for input")
		inputChan := <-inputChannelQueue
		// fmt.Println("got new input channel")
		input := <-inputChan
		// fmt.Println("got input from channel")
		response, err := handleInput(input)
		if err != nil {
			fmt.Println("Error handling input: ", err.Error())
			os.Exit(1)
		}
		inputChan <- response
	}
}

func initConfig() {
	dir := flag.String("dir", "/data", "Directory to store Redis data")
	dbfilename := flag.String("dbfilename", "dump.rdb", "Filename for Redis database")
	flag.Parse()

	store.AddConfig("dir", *dir)
	store.AddConfig("dbfilename", *dbfilename)
}

func listen() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	defer l.Close()

	for {
		// fmt.Println("listener waiting for connection")
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// fmt.Println("Handling new connection")
	defer func() {
		fmt.Printf("Finished with connection \n\n")
		conn.Close()
	}()

	bufSize := 1024
	readBuf := make([]byte, bufSize)
	inputChan := make(chan []byte)

	for {
		n, err := conn.Read(readBuf)
		if err != nil {
			fmt.Println("Error reading from connection: ", err)
			return
		}
		if n == 0 {
			// Client closed connection gracefully
			// fmt.Println("Client closed connection")
			return
		}

		// fmt.Printf("read %v bytes \n", n)
		inputChannelQueue <- inputChan
		inputChan <- readBuf[:n]
		resp := <-inputChan

		fmt.Printf("writing: %q \n", resp)
		_, writeErr := conn.Write([]byte(resp))
		if writeErr != nil {
			fmt.Println("Error writing to connection:", writeErr)
			return
		}
	}
}

func handleInput(input []byte) ([]byte, error) {
	// fmt.Println("handling new input...")

	command, err := parser.ParseInput(input)
	if err != nil {
		return nil, fmt.Errorf("could not parse input: %q, %w", string(input), err)
	}

	resp, err := command.Handle()
	if err != nil {
		return nil, fmt.Errorf("command %v failed: %w", command, err)
	}

	return []byte(resp), nil
}
