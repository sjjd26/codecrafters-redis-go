package redis

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
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

	inputQueue         chan *ConnectionInput
	replicationDetails *redisConfig.ReplicationDetails
}

type ConnectionInput struct {
	Conn           net.Conn
	Input          []byte
	HshakeStep     interfaces.HandshakeStep
	ResponseChan   chan []byte
	IsWriteCommand bool
}

func NewRedisInstance(selfDetails, masterDetails *redisConfig.HostDetails) (*RedisInstance, error) {
	config := redisConfig.NewRedisConfig()
	store := store.NewRedisStore()
	parser := &parser.RedisParserImpl{}
	inputQueue := make(chan *ConnectionInput, 10)

	replicationDetails, err := redisConfig.NewReplicationDetails(selfDetails, masterDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to create replication details: %w", err)
	}
	config.SetReplicationDetails(replicationDetails)

	instance := &RedisInstance{
		Config:             config,
		Store:              store,
		Parser:             parser,
		inputQueue:         inputQueue,
		replicationDetails: replicationDetails,
	}

	return instance, nil
}

func (inst *RedisInstance) ListenAndRun() {
	go inst.Listen()

	if inst.replicationDetails.Role == redisConfig.RoleSlave {
		conn, err := inst.Handshake()
		if err != nil {
			panic(err)
		}
		go func() {
			err := inst.handleConnection(conn)
			if err != nil {
				panic(fmt.Errorf("Error handling connection: %w", err))
			}
		}()
	}

	inst.RunMainEventLoop()
}

func (inst *RedisInstance) RunMainEventLoop() {
	for {
		// fmt.Println("event loop waiting for input")
		inputConn := <-inst.inputQueue
		response, err := inst.handleInput(inputConn)
		if err != nil {
			panic(fmt.Errorf("Error handling input: %w", err))
		}
		inputConn.ResponseChan <- response
	}
}

func (inst *RedisInstance) Listen() {
	port := inst.replicationDetails.SelfDetails.Port
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
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
				panic(fmt.Errorf("Error handling connection: %w", err))
			}
		}()
	}
}

func (inst *RedisInstance) handleInput(connInput *ConnectionInput) ([]byte, error) {
	// fmt.Println("handling new input... " + string(connInput.Input))

	command, err := inst.Parser.ParseInput(connInput.Input)
	if err != nil {
		return nil, fmt.Errorf("could not parse input: %q, %w", string(connInput.Input), err)
	}

	resp, err := command.Handle()
	if err != nil {
		return nil, fmt.Errorf("command %s failed: %w", command, err)
	}

	isMaster := inst.replicationDetails.Role == redisConfig.RoleMaster
	if wc, ok := command.(interfaces.WriteCommand); isMaster && ok && wc.IsWriteCommand() {
		err := inst.PropagateInput(connInput.Input)
		if err != nil {
			// Should return error here? Need better handling
			return nil, fmt.Errorf("Error propagating input: %q, %s\n", string(connInput.Input), err.Error())
		}
		connInput.IsWriteCommand = true
	}

	if hc, ok := command.(interfaces.HandshakeCommand); isMaster && ok && hc.IsHandshakeCommand() {
		inst.handleHandshakeStep(connInput, hc.GetHandshakeStep())
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
	// inputChan := make(chan []byte)

	connInput := &ConnectionInput{
		Conn:           conn,
		Input:          []byte{},
		HshakeStep:     interfaces.HandshakeStepNone,
		ResponseChan:   make(chan []byte),
		IsWriteCommand: false,
	}

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
		connInput.IsWriteCommand = false
		connInput.Input = readBuf[:n]
		inst.inputQueue <- connInput
		// inputChan <- readBuf[:n]

		resp := <-connInput.ResponseChan

		// Replica slaves don't send responses for propogated writes back to the master
		skipResponse := inst.replicationDetails.Role == redisConfig.RoleSlave &&
			connInput.IsWriteCommand
		// fmt.Println("skip? ", skipResponse)
		if skipResponse {
			continue
		}

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

func (inst *RedisInstance) Handshake() (net.Conn, error) {
	fmt.Println("Initiating handshake with master node...")

	if inst.replicationDetails.Role == redisConfig.RoleMaster {
		return nil, fmt.Errorf("master node does not send handshake")
	}
	masterDetails := inst.replicationDetails.MasterDetails
	if masterDetails == nil {
		return nil, fmt.Errorf("master details not set for slave node")
	}

	address := fmt.Sprintf("%s:%d", masterDetails.Host, masterDetails.Port)
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master %s: %w", address, err)
	}

	// fmt.Println("Sending PING...")
	if err := inst.sendPing(conn); err != nil {
		return nil, err
	}

	// fmt.Println("Sending 1st REPLCONF...")
	if err := inst.sendReplConfListeningPort(conn); err != nil {
		return nil, err
	}

	// fmt.Println("Sending 2nd REPLCONF...")
	if err := inst.sendReplConfCapaPysync2(conn); err != nil {
		return nil, err
	}

	// fmt.Println("Sending PSYNC...")
	if err := inst.sendPsync(conn); err != nil {
		return nil, err
	}

	return conn, nil
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

func (inst *RedisInstance) handleHandshakeStep(connInput *ConnectionInput, newStep interfaces.HandshakeStep) {
	if connInput.HshakeStep == interfaces.HandshakeStepPsync {
		return // already finished the handshake so nothing to do
	}

	isNextStep := newStep-connInput.HshakeStep == 1
	if isNextStep {
		connInput.HshakeStep = newStep
		if newStep == interfaces.HandshakeStepPsync {
			inst.replicationDetails.AddSlaveConn(connInput.Conn)
		}
	} else {
		connInput.HshakeStep = interfaces.HandshakeStepNone
	}
}

func (inst *RedisInstance) PropagateInput(input []byte) error {
	// fmt.Printf("Propagating input: %q \n", input)
	if inst.replicationDetails.Role != redisConfig.RoleMaster {
		return fmt.Errorf("only master nodes can propagate inputs")
	}

	for _, slaveConn := range inst.replicationDetails.SlaveConnections {
		if _, err := slaveConn.Write(input); err != nil {
			return fmt.Errorf("failed to write input to slave connection: %w", err)
		}
	}

	return nil
}
