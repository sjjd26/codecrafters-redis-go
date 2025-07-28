package redis

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/redis/parser"
	"github.com/codecrafters-io/redis-starter-go/app/redis/rdbRestorer"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type RedisInstance struct {
	Config redisConfig.RedisConfig
	Store  store.RedisStore
	Parser parser.RedisParser

	inputChannelQueue  chan chan []byte
	replicationDetails *redisConfig.ReplicationDetails
}

func NewRedisInstance(selfDetails, masterDetails *redisConfig.HostDetails) (*RedisInstance, error) {
	config := redisConfig.NewRedisConfig()
	store := store.NewRedisStore()
	parser := &parser.RedisParserImpl{}
	channelQueue := make(chan chan []byte, 10)

	replicationDetails, err := redisConfig.NewReplicationDetails(selfDetails, masterDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to create replication details: %w", err)
	}
	config.SetReplicationDetails(replicationDetails)

	instance := &RedisInstance{
		Config:             config,
		Store:              store,
		Parser:             parser,
		inputChannelQueue:  channelQueue,
		replicationDetails: replicationDetails,
	}

	if masterDetails != nil {
		err := instance.Handshake()
		if err != nil {
			return nil, fmt.Errorf("handshake failed: %w", err)
		}
	}

	return instance, nil
}

func (inst *RedisInstance) ListenAndRun() {
	go inst.Listen()
	inst.RunMainEventLoop()
}

func (inst *RedisInstance) RunMainEventLoop() {
	for {
		// fmt.Println("event loop waiting for input")
		inputChan := <-inst.inputChannelQueue
		// fmt.Println("got new input channel")
		input := <-inputChan
		// fmt.Println("got input from channel")
		response, err := inst.handleInput(input)
		if err != nil {
			panic(fmt.Errorf("Error handling input: %w", err))
		}
		inputChan <- response
	}
}

func (inst *RedisInstance) Listen() {
	port := inst.replicationDetails.SelfDetails.Port
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
		panic(err)
	}

	defer l.Close()

	for {
		// fmt.Println("listener waiting for connection")
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go func() {
			err := inst.handleConnection(conn)
			if err != nil {
				fmt.Println("Error handling connection: ", err.Error())
			}
		}()
	}
}

func (inst *RedisInstance) handleInput(input []byte) ([]byte, error) {
	// fmt.Println("handling new input...")

	command, err := inst.Parser.ParseInput(input)
	if err != nil {
		return nil, fmt.Errorf("could not parse input: %q, %w", string(input), err)
	}

	resp, err := command.Handle()
	if err != nil {
		return nil, fmt.Errorf("command %s failed: %w", command, err)
	}

	return []byte(resp), nil
}

func (inst *RedisInstance) handleConnection(conn net.Conn) error {
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
			return nil
		} else if err != nil {
			fmt.Println("Error reading from connection: ", err)
			return err
		}
		if n == 0 {
			// Client closed connection gracefully
			// fmt.Println("Finished reading from connection")
			return nil
		}

		// fmt.Printf("read %v bytes \n", n)
		inst.inputChannelQueue <- inputChan
		inputChan <- readBuf[:n]
		resp := <-inputChan

		fmt.Printf("writing: %q \n", resp)
		_, writeErr := conn.Write([]byte(resp))
		if writeErr != nil {
			fmt.Println("Error writing to connection:", writeErr)
			return writeErr
		}
	}
}

func (inst *RedisInstance) RestoreFromRdb() error {
	dir, ok := inst.Config.Get(redisConfig.ConfigDir)
	if !ok {
		return fmt.Errorf("dir config value cannot be nil")
	}
	dbfilename, ok := inst.Config.Get(redisConfig.ConfigDbFilename)
	if !ok {
		return fmt.Errorf("dbfilename config value cannot be nil")
	}

	filepath := fmt.Sprintf("%s/%s", dir, dbfilename)
	restorer := rdbRestorer.NewRdbRestorer(inst.Store)
	err := restorer.RestoreFromRdb(filepath)
	if err != nil {
		return fmt.Errorf("restoration from rdb failed: %w", err)
	}

	return nil
}

func (inst *RedisInstance) Handshake() error {
	fmt.Println("Initiating handshake with master node...")

	if inst.replicationDetails.Role == redisConfig.RoleMaster {
		return fmt.Errorf("master node does not send handshake")
	}
	masterDetails := inst.replicationDetails.MasterDetails
	if masterDetails == nil {
		return fmt.Errorf("master details not set for slave node")
	}

	address := fmt.Sprintf("%s:%d", masterDetails.Host, masterDetails.Port)
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to master %s: %w", address, err)
	}
	defer conn.Close()

	// fmt.Println("Sending PING...")
	if err := inst.sendPing(conn); err != nil {
		return err
	}

	// fmt.Println("Sending 1st REPLCONF...")
	if err := inst.sendReplConfListeningPort(conn); err != nil {
		return err
	}

	// fmt.Println("Sending 2nd REPLCONF...")
	if err := inst.sendReplConfCapaPysync2(conn); err != nil {
		return err
	}

	// fmt.Println("Sending PSYNC...")
	if err := inst.sendPsync(conn); err != nil {
		return err
	}

	return nil
}

func (inst *RedisInstance) sendPing(conn net.Conn) error {
	command := types.CreateBulkStringArray([]string{"PING"})
	expectedResponse := "+PONG\r\n"

	if err := inst.sendHandshakeCommand(conn, command, expectedResponse); err != nil {
		return fmt.Errorf("failed to send PING command: %w", err)
	}

	return nil
}

func (inst *RedisInstance) sendReplConfListeningPort(conn net.Conn) error {
	port := strconv.Itoa(inst.replicationDetails.SelfDetails.Port)
	commandParts := []string{"REPLCONF", "listening-port", port}
	command := types.CreateBulkStringArray(commandParts)

	if err := inst.sendHandshakeCommand(conn, command, types.OkString); err != nil {
		return fmt.Errorf("failed to send initial REPLCONF command: %w", err)
	}

	return nil
}

func (inst *RedisInstance) sendReplConfCapaPysync2(conn net.Conn) error {
	commandParts := []string{"REPLCONF", "capa", "pysync2"}
	command := types.CreateBulkStringArray(commandParts)

	if err := inst.sendHandshakeCommand(conn, command, types.OkString); err != nil {
		return fmt.Errorf("failed to send second REPLCONF command: %w", err)
	}

	return nil
}

func (inst *RedisInstance) sendPsync(conn net.Conn) error {
	commandParts := []string{"PSYNC", "?", "-1"}
	command := types.CreateBulkStringArray(commandParts)

	if err := inst.sendHandshakeCommand(conn, command, ""); err != nil {
		return fmt.Errorf("failed to send PSYNC command: %w", err)
	}

	return nil
}

func (inst *RedisInstance) sendHandshakeCommand(conn net.Conn, command, expectedResp string) error {
	var err error
	if _, err = conn.Write([]byte(command)); err != nil {
		return fmt.Errorf("failed to write command %q: %w", command, err)
	}

	// Read the response from the master
	response := make([]byte, 1024)
	var n int = 0
	if n, err = conn.Read(response); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if expectedResp != "" && string(response[:n]) != expectedResp {
		return fmt.Errorf("unexpected response: %q", response[:n])
	}

	return nil
}
