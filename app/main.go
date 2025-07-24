package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis/parser"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
)

var inputChannelQueue = make(chan chan []byte, 10)

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	dir := flag.String("dir", "/data", "Directory to store data files")
	dbfilename := flag.String("dbfilename", "dump.rdb", "Filename for the database dump")
	replicaOf := flag.String("replicaof", "", "'<Host> <port>' of the master node to replicate from (optional)")
	flag.Parse()

	initConfig(dir, dbfilename, replicaOf, *port)
	go listen(*port)

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

func initConfig(dir, dbfilename, replicaOf *string, port int) {
	config := redisConfig.NewRedisConfig()
	config.Set(redisConfig.ConfigDir, *dir)
	config.Set(redisConfig.ConfigDbFilename, *dbfilename)

	replicationDetails, err := redisConfig.NewReplicationDetails(redisConfig.RoleMaster, port)
	if err != nil {
		panic(err)
	}
	if *replicaOf != "" {
		masterHost := strings.Split(*replicaOf, " ")
		if len(masterHost) != 2 {
			fmt.Println("Invalid replicaOf format. Use '<host> <port>'")
			os.Exit(1)
		}
		masterPort, err := strconv.Atoi(masterHost[1])
		if err != nil {
			fmt.Println("Invalid port number:", masterHost[1])
			os.Exit(1)
		}
		replicationDetails.MasterDetails = &redisConfig.HostDetails{
			Host: masterHost[0],
			Port: masterPort,
		}
		replicationDetails.Role = redisConfig.RoleSlave
	}
	config.SetReplicationDetails(replicationDetails)

	redisStore := store.NewRedisStore()
	err = redisStore.RdbRestore()
	if err != nil {
		fmt.Println(err)
	}
}

func listen(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
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
		if err != nil && err.Error() == "EOF" {
			// fmt.Println("Client closed connection")
			return
		} else if err != nil {
			fmt.Println("Error reading from connection: ", err)
			return
		}
		if n == 0 {
			// Client closed connection gracefully
			// fmt.Println("Finished reading from connection")
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
