package redis

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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
	Conn         net.Conn
	Input        []byte
	HshakeStep   interfaces.HandshakeStep
	ResponseChan chan []byte
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
		inst.replicationDetails.MasterConn = conn
		go func() {
			// fmt.Printf("Handling connection to master node: %s\n", conn.RemoteAddr().String())
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
		fmt.Println("event loop waiting for input")
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
		fmt.Println("listener waiting for connection")
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		fmt.Printf("New connection accepted from %s\n", conn.RemoteAddr().String())
		go func() {
			err := inst.handleConnection(conn)
			if err != nil {
				panic(fmt.Errorf("Error handling connection: %w", err))
			}
		}()
	}
}

func (inst *RedisInstance) handleConnection(conn net.Conn) error {
	isMasterConn := conn == inst.replicationDetails.MasterConn
	fmt.Printf("Handling new connection (master: %v): %s\n", isMasterConn, conn.RemoteAddr().String())
	defer func() {
		fmt.Printf("Finished with connection: %s\n", conn.RemoteAddr().String())
		conn.Close()
	}()

	bufSize := 1024
	readBuf := make([]byte, bufSize)

	connInput := &ConnectionInput{
		Conn:         conn,
		Input:        []byte{},
		HshakeStep:   interfaces.HandshakeStepNone,
		ResponseChan: make(chan []byte),
	}

	for {
		n, err := conn.Read(readBuf)
		if err != nil && err.Error() == "EOF" {
			// fmt.Println("Client closed connection")
			return nil
		} else if err != nil {
			fmt.Printf("Error reading from connection %s: %v\n", conn.RemoteAddr().String(), err)
			return err
		}
		if n == 0 {
			// Client closed connection gracefully
			// fmt.Println("Finished reading from connection")
			return nil
		}

		// fmt.Printf("read %v bytes \n", n)
		connInput.Input = readBuf[:n]
		inst.inputQueue <- connInput

		resp := <-connInput.ResponseChan

		if len(resp) == 0 {
			fmt.Printf("no response for connection (master: %v) %s input\n", isMasterConn, conn.RemoteAddr().String())
			continue
		}

		fmt.Printf("writing to connection %s: %q \n", conn.RemoteAddr().String(), resp)
		_, writeErr := conn.Write([]byte(resp))
		if writeErr != nil {
			fmt.Printf("Error writing to connection %s: %v\n", conn.RemoteAddr().String(), writeErr)
			return writeErr
		}
	}
}

func (inst *RedisInstance) handleInput(connInput *ConnectionInput) ([]byte, error) {
	// fmt.Printf("handling new input: %q\n", connInput.Input)

	isMasterConn := connInput.Conn == inst.replicationDetails.MasterConn
	var resp string
	currentInput := connInput.Input

	for currentInput != nil && len(currentInput) > 0 {
		command, inputLen, err := inst.Parser.ParseInput(currentInput)
		if err != nil {
			return nil, fmt.Errorf("could not parse input: %q, %w", currentInput, err)
		}

		commandResp, err := command.Handle()
		if err != nil {
			return nil, fmt.Errorf("command %s failed: %w", command, err)
		}

		inst.replicationDetails.ReplicaOffset += inputLen
		// fmt.Printf("replica offset updated to %d\n", inst.replicationDetails.ReplicaOffset)

		// Refactor with slave/master post command processing
		if isMasterConn {
			if mc, ok := command.(interfaces.MasterResponseCommand); ok && mc.IsMasterResponseCommand() {
				resp += commandResp
			}
		} else {
			resp += commandResp
			if wc, ok := command.(interfaces.WriteCommand); ok && wc.IsWriteCommand() {
				err := inst.PropagateInput(currentInput)
				if err != nil {
					// Should return error here? Need better handling
					return nil, fmt.Errorf("Error propagating input: %q, %s\n", currentInput, err.Error())
				}
			}

			if hc, ok := command.(interfaces.HandshakeCommand); ok && hc.IsHandshakeCommand() {
				inst.handleHandshakeStep(connInput, hc.GetHandshakeStep())
			}
		}

		currentInput = currentInput[inputLen:]
	}

	return []byte(resp), nil
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
	// fmt.Println("Initiating handshake with master node...")

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
	input, err := inst.sendPsync(conn)
	if err != nil {
		return nil, err
	}

	// handle any remaining input
	if input != nil && len(input) > 0 {
		fmt.Printf("Received input after PSYNC: %q\n", input)
		connInput := &ConnectionInput{
			Conn:         conn,
			Input:        input,
			HshakeStep:   interfaces.HandshakeStepPsync,
			ResponseChan: make(chan []byte),
		}
		resp, err := inst.handleInput(connInput)
		if err != nil {
			return nil, fmt.Errorf("failed to handle input after PSYNC: %w", err)
		}
		if len(resp) > 0 {
			fmt.Printf("Response after PSYNC: %q\n", resp)
			if _, err := conn.Write(resp); err != nil {
				return nil, fmt.Errorf("failed to write response after PSYNC: %w", err)
			}
		}
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

// Returns any remaining input from the master after FULLRESYNC + RDB response
func (inst *RedisInstance) sendPsync(conn net.Conn) ([]byte, error) {
	commandParts := []string{"PSYNC", "?", "-1"}
	command := types.CreateBulkStringArray(commandParts)

	var err error
	if _, err = conn.Write([]byte(command)); err != nil {
		return nil, fmt.Errorf("failed to write command %q: %w", command, err)
	}

	// Read the response from the master
	response := make([]byte, 1024)
	var n int = 0
	if n, err = conn.Read(response); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fullResyncResp, rdbResp, ok := strings.Cut(string(response[:n]), "\r\n")
	if fullResyncResp == "" || !ok {
		return nil, fmt.Errorf("none or invalid FULLRESYNC response received from master for PSYNC command: %q" + string(response[:n]))
	}

	getRemainingInput := func(rdbResp []byte) ([]byte, error) {
		rdbLen, p, err := inst.Parser.GetAggregateLength(rdbResp)
		if err != nil {
			return nil, fmt.Errorf("failed to get length of RDB response: %w", err)
		}
		if p+rdbLen > len(rdbResp) {
			return nil, fmt.Errorf("RDB response length exceeds available data: %d > %d", p+rdbLen, len(rdbResp))
		}
		if p+rdbLen == len(rdbResp) {
			// fmt.Println("RDB response is complete, no additional input received")
			return nil, nil
		}
		return rdbResp[p+rdbLen:], nil
	}

	// RDB response may be included in this response or may be sent separately
	if rdbResp != "" {
		// fmt.Printf("received RDB response from master with FULLRESYNC: %q\n", rdbResp)
		// for now just ignore the rdb response
		return getRemainingInput([]byte(rdbResp))
	}

	// RDB response not included directly in the PSYNC (FULLRESYNC) response
	if n, err = conn.Read(response); err != nil {
		return nil, fmt.Errorf("failed to read RDB response: %w", err)
	}
	if n == 0 {
		return nil, fmt.Errorf("no RDB response received from master after PSYNC command")
	}

	// fmt.Printf("received RDB response from master after FULLRESYNC: %q\n", response[:n])
	// again just ignore rdb response
	return getRemainingInput(response[:n])
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
	if expectedResp != "" && strings.ToUpper(string(response[:n])) != strings.ToUpper(expectedResp) {
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
			fmt.Printf("Added slave connection: %s\n", connInput.Conn.RemoteAddr().String())
			fmt.Printf("Slave connections: %d\n", len(inst.replicationDetails.SlaveConnections))
		}
	} else {
		connInput.HshakeStep = interfaces.HandshakeStepNone
	}
}

// SHOULD USE A REPLICATION STREAM -> SEND INPUT TO REPLICATION STREAM WHICH IS THEN SENT TO ALL SLAVES VIA BACKGROUND PROCESS?
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
