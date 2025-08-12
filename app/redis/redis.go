package redis

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/redis/command"
	"github.com/codecrafters-io/redis-starter-go/app/redis/parser"
	"github.com/codecrafters-io/redis-starter-go/app/redis/rdbRestorer"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

type RedisInstance interface {
	ListenAndRun() error
	// RestoreFromRdb() error
	// HandleInput(connInput *ConnectionInput) ([]byte, error)
	ProcessCommandResponse(connInput *ConnectionInput, commandInput []byte, command command.Command, response string) (string, error)

	GetInputQueue() chan *ConnectionInput
	GetReplicationDetails() *redisConfig.ReplicationDetails
	GetParser() parser.RedisParser
	GetConfig() redisConfig.RedisConfig
	GetStore() *store.RedisStore
}

type RedisCore struct {
	Config redisConfig.RedisConfig
	Store  store.RedisStore
	Parser parser.RedisParser

	inputQueue         chan *ConnectionInput
	replicationDetails *redisConfig.ReplicationDetails
}

type RedisMaster struct {
	Config redisConfig.RedisConfig
	Store  *store.RedisStore
	Parser parser.RedisParser

	inputQueue         chan *ConnectionInput
	replicationDetails *redisConfig.ReplicationDetails

	replicationInputQueue  chan *ReplicaResponse
	replicationOutputQueue chan []byte
}

type RedisReplica struct {
	Config redisConfig.RedisConfig
	Store  *store.RedisStore
	Parser parser.RedisParser

	inputQueue         chan *ConnectionInput
	replicationDetails *redisConfig.ReplicationDetails
}

type ConnectionInput struct {
	Conn         net.Conn
	Input        []byte
	HshakeStep   command.HandshakeStep
	ResponseChan chan []byte
}

type ReplicaResponse struct {
	Conn     net.Conn
	Response []byte
	Error    error
}

func NewRedisInstance(selfDetails, masterDetails *redisConfig.HostDetails) (RedisInstance, error) {
	config := redisConfig.NewRedisConfig()
	store := store.NewRedisStore()
	parser := &parser.RedisParserImpl{}
	inputQueue := make(chan *ConnectionInput, 20)

	replicationDetails, err := redisConfig.NewReplicationDetails(selfDetails, masterDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to create replication details: %w", err)
	}
	config.SetReplicationDetails(replicationDetails)

	if masterDetails == nil {
		fmt.Println("Creating master")
		return &RedisMaster{
			Config:                 config,
			Store:                  store,
			Parser:                 parser,
			inputQueue:             inputQueue,
			replicationDetails:     replicationDetails,
			replicationInputQueue:  make(chan *ReplicaResponse, 20),
			replicationOutputQueue: make(chan []byte, 20),
		}, nil
	} else {
		fmt.Println("Creating replica")
		return &RedisReplica{
			Config:             config,
			Store:              store,
			Parser:             parser,
			inputQueue:         inputQueue,
			replicationDetails: replicationDetails,
		}, nil
	}
}

// --------------------- Redis Master --------------------------

func (m *RedisMaster) GetInputQueue() chan *ConnectionInput {
	return m.inputQueue
}

func (m *RedisMaster) GetReplicationDetails() *redisConfig.ReplicationDetails {
	return m.replicationDetails
}

func (m *RedisMaster) GetParser() parser.RedisParser {
	return m.Parser
}

func (m *RedisMaster) GetConfig() redisConfig.RedisConfig {
	return m.Config
}

func (m *RedisMaster) GetStore() *store.RedisStore {
	return m.Store
}

func (m *RedisMaster) ListenAndRun() error {
	go Listen(m)

	go m.ReplicationInputEventLoop()
	go m.ReplicationOutputEventLoop()

	err := MainEventLoop(m)
	if err != nil {
		return fmt.Errorf("main event loop error: %w", err)
	}
	return nil
}

func (m *RedisMaster) ProcessCommandResponse(connInput *ConnectionInput, commandInput []byte, cmd command.Command, response string) (string, error) {
	if cmd == nil {
		return "", fmt.Errorf("command cannot be nil")
	}

	if wc, ok := cmd.(command.WriteCommand); ok && wc.IsWriteCommand() {
		m.replicationOutputQueue <- commandInput
	}
	if hc, ok := cmd.(command.HandshakeCommand); ok && hc.IsHandshakeCommand() {
		m.HandleHandshakeStep(connInput, hc.GetHandshakeStep())
	}

	return response, nil
}

func (m *RedisMaster) HandleHandshakeStep(connInput *ConnectionInput, newStep command.HandshakeStep) {
	if connInput.HshakeStep == command.HandshakeStepPsync {
		return // already finished the handshake so nothing to do
	}

	isNextStep := newStep-connInput.HshakeStep == 1
	if isNextStep {
		connInput.HshakeStep = newStep
		if newStep == command.HandshakeStepPsync {
			m.replicationDetails.AddSlaveConn(connInput.Conn)
			fmt.Printf("Added slave connection: %s\n", connInput.Conn.RemoteAddr().String())
			// fmt.Printf("Slave connections: %d\n", len(m.replicationDetails.SlaveConnections))
		}
	} else {
		connInput.HshakeStep = command.HandshakeStepNone
	}
}

func (m *RedisMaster) ReplicationInputEventLoop() {
	for response := range m.replicationInputQueue {
		if response.Error != nil {
			fmt.Printf("Error from slave %s: %v\n", response.Conn.RemoteAddr().String(), response.Error)
			continue
		}
		m.HandleReplicaResponse(response.Conn, response.Response)
	}
}

func (m *RedisMaster) HandleReplicaResponse(conn net.Conn, response []byte) error {
	commandParts, _, err := m.Parser.ParseInput(response)
	if err != nil {
		fmt.Printf("Failed to parse input from replica %s: %v\n", conn.RemoteAddr().String(), err)
		return fmt.Errorf("failed to parse input from replica %s: %w", conn.RemoteAddr().String(), err)
	}

	ctx := &command.CommandContext{
		Store:              m.Store,
		Config:             m.Config,
		ReplicationDetails: m.replicationDetails,
		Conn:               conn,
	}
	commandResp, _, err := HandleCommand(m, commandParts, ctx)
	if err != nil {
		return fmt.Errorf("failed to handle command from replica %s: %w", conn.RemoteAddr().String(), err)
	}

	// for now we just log the command response, do not propagate it further
	fmt.Printf("Received command response from replica %s: %q\n", conn.RemoteAddr().String(), commandResp)
	return nil
}

func (m *RedisMaster) ReplicationOutputEventLoop() {
	for command := range m.replicationOutputQueue {
		m.BroadcastToReplicas(command)
	}
}

func (m *RedisMaster) BroadcastToReplicas(command []byte) {
	m.replicationDetails.ReplicaOffset += len(command)

	// fmt.Printf("Propagating input: %q \n", input)
	for conn := range m.replicationDetails.SlaveConnections {
		go func(slaveConn net.Conn) {
			_, err := slaveConn.Write(command)
			if err != nil {
				m.replicationInputQueue <- &ReplicaResponse{
					Conn:  slaveConn,
					Error: err,
				}
				panic(err)
			}
		}(conn)
	}
}

// --------------------- Redis Replica --------------------------

func (r *RedisReplica) GetInputQueue() chan *ConnectionInput {
	return r.inputQueue
}

func (r *RedisReplica) GetReplicationDetails() *redisConfig.ReplicationDetails {
	return r.replicationDetails
}

func (r *RedisReplica) GetParser() parser.RedisParser {
	return r.Parser
}

func (r *RedisReplica) GetConfig() redisConfig.RedisConfig {
	return r.Config
}

func (r *RedisReplica) GetStore() *store.RedisStore {
	return r.Store
}

func (r *RedisReplica) ListenAndRun() error {
	go Listen(r)

	conn, err := r.Handshake()
	if err != nil {
		return fmt.Errorf("handshake error: %w", err)
	}
	go func() {
		// fmt.Printf("Handling connection to master node: %s\n", conn.RemoteAddr().String())
		err := HandleConnection(r, conn)
		if err != nil {
			panic(fmt.Errorf("Error handling connection: %w", err))
		}
	}()

	err = MainEventLoop(r)
	if err != nil {
		return fmt.Errorf("main event loop: %w", err)
	}
	return nil
}

func (r *RedisReplica) Handshake() (net.Conn, error) {
	// fmt.Println("Initiating handshake with master node...")

	if r.replicationDetails.Role == redisConfig.RoleMaster {
		return nil, fmt.Errorf("master node does not send handshake")
	}
	masterDetails := r.replicationDetails.MasterDetails
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
	r.replicationDetails.MasterConn = conn

	// fmt.Println("Sending PING...")
	if err := r.SendPing(conn); err != nil {
		return nil, err
	}

	// fmt.Println("Sending 1st REPLCONF...")
	if err := r.SendReplConfListeningPort(conn); err != nil {
		return nil, err
	}

	// fmt.Println("Sending 2nd REPLCONF...")
	if err := r.SendReplConfCapaPysync2(conn); err != nil {
		return nil, err
	}

	// fmt.Println("Sending PSYNC...")
	input, err := r.SendPsync(conn)
	if err != nil {
		return nil, err
	}

	// handle any remaining input
	if input != nil && len(input) > 0 {
		fmt.Printf("Received input after PSYNC: %q\n", input)
		connInput := &ConnectionInput{
			Conn:         conn,
			Input:        input,
			HshakeStep:   command.HandshakeStepPsync,
			ResponseChan: make(chan []byte),
		}
		resp, err := HandleInput(r, connInput)
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

func (r *RedisReplica) SendPing(conn net.Conn) error {
	command := types.CreateBulkStringArray([]string{"PING"})
	expectedResponse := "+PONG\r\n"

	if err := r.SendHandshakeCommand(conn, command, expectedResponse); err != nil {
		return fmt.Errorf("failed to send PING command: %w", err)
	}

	return nil
}

func (r *RedisReplica) SendReplConfListeningPort(conn net.Conn) error {
	port := strconv.Itoa(r.replicationDetails.SelfDetails.Port)
	commandParts := []string{"REPLCONF", "listening-port", port}
	command := types.CreateBulkStringArray(commandParts)

	if err := r.SendHandshakeCommand(conn, command, types.OkString); err != nil {
		return fmt.Errorf("failed to send initial REPLCONF command: %w", err)
	}

	return nil
}

func (r *RedisReplica) SendReplConfCapaPysync2(conn net.Conn) error {
	commandParts := []string{"REPLCONF", "capa", "pysync2"}
	command := types.CreateBulkStringArray(commandParts)

	if err := r.SendHandshakeCommand(conn, command, types.OkString); err != nil {
		return fmt.Errorf("failed to send second REPLCONF command: %w", err)
	}

	return nil
}

// Returns any remaining input from the master after FULLRESYNC + RDB response
func (r *RedisReplica) SendPsync(conn net.Conn) ([]byte, error) {
	commandParts := []string{"PSYNC", "?", "-1"}
	command := types.CreateBulkStringArray(commandParts)

	var err error
	if _, err = conn.Write([]byte(command)); err != nil {
		return nil, fmt.Errorf("failed to write command %q: %w", command, err)
	}

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
		rdbLen, p, err := r.Parser.GetAggregateLength(rdbResp)
		if err != nil {
			return nil, fmt.Errorf("failed to get length of RDB response: %w", err)
		}
		if p+rdbLen > len(rdbResp) {
			return nil, fmt.Errorf("RDB response length exceeds available data: %d > %d", p+rdbLen, len(rdbResp))
		}
		if p+rdbLen == len(rdbResp) {
			fmt.Println("RDB response is complete, no additional input received")
			return nil, nil
		}
		return rdbResp[p+rdbLen:], nil
	}

	// RDB response may be included in this response or may be sent separately
	if rdbResp != "" {
		fmt.Printf("received RDB response from master with FULLRESYNC: %q\n", rdbResp)
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

	fmt.Printf("received RDB response from master after FULLRESYNC: %q\n", response[:n])
	// again just ignore rdb response
	return getRemainingInput(response[:n])
}

func (r *RedisReplica) SendHandshakeCommand(conn net.Conn, command, expectedResp string) error {
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

func (r *RedisReplica) ProcessCommandResponse(connInput *ConnectionInput, commandInput []byte, cmd command.Command, response string) (string, error) {
	isMasterConn := connInput.Conn == r.GetReplicationDetails().MasterConn
	fmt.Printf("Processing command response for connection (master: %v): %q\n", isMasterConn, response)
	if !isMasterConn {
		return response, nil
	}

	r.replicationDetails.ReplicaOffset += len(commandInput)
	fmt.Printf("Replica offset updated to %d\n", r.replicationDetails.ReplicaOffset)
	if mc, ok := cmd.(command.MasterResponseCommand); ok && mc.IsMasterResponseCommand() {
		return response, nil
	}
	return "", nil
}

// --------------------- Helper functions --------------------------

func MainEventLoop(inst RedisInstance) error {
	// fmt.Println("Starting main event loop...")
	for inputConn := range inst.GetInputQueue() {
		// fmt.Println("event loop received input")
		response, err := HandleInput(inst, inputConn)
		if err != nil {
			return fmt.Errorf("Error handling input: %w", err)
		}
		inputConn.ResponseChan <- response
	}
	return nil
}

func Listen(inst RedisInstance) {
	port := inst.GetReplicationDetails().SelfDetails.Port
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
			err := HandleConnection(inst, conn)
			if err != nil {
				panic(fmt.Errorf("Error handling connection: %w", err))
			}
		}()
	}
}

func HandleConnection(inst RedisInstance, conn net.Conn) error {
	isMasterConn := conn == inst.GetReplicationDetails().MasterConn
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
		HshakeStep:   command.HandshakeStepNone,
		ResponseChan: make(chan []byte),
	}

	for {
		n, err := conn.Read(readBuf)
		if err != nil && err.Error() == "EOF" {
			fmt.Println("Client closed connection")
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

		if m, ok := inst.(*RedisMaster); ok {
			_, isReplicaConn := m.GetReplicationDetails().GetSlaveConnDetails(conn)
			if isReplicaConn {
				replicaResp := &ReplicaResponse{
					Conn:     conn,
					Response: readBuf[:n],
				}
				fmt.Printf("Received input from replica %s: %q\n", conn.RemoteAddr().String(), replicaResp.Response)
				m.replicationInputQueue <- replicaResp
				continue // Skip further processing for replica connections
			}
		}

		// fmt.Printf("received input %q \n", readBuf[:n])
		connInput.Input = readBuf[:n]
		inst.GetInputQueue() <- connInput

		resp := <-connInput.ResponseChan
		// fmt.Printf("response received for connection %s: %q\n", conn.RemoteAddr().String(), resp)

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

func HandleInput(inst RedisInstance, connInput *ConnectionInput) ([]byte, error) {
	fmt.Printf("handling new input: %q\n", connInput.Input)
	var resp string
	ctx := &command.CommandContext{
		Store:              inst.GetStore(),
		Config:             inst.GetConfig(),
		ReplicationDetails: inst.GetReplicationDetails(),
		Conn:               connInput.Conn,
	}
	currentInput := connInput.Input

	for currentInput != nil && len(currentInput) > 0 {
		commandParts, inputLen, err := inst.GetParser().ParseInput(currentInput)
		if err != nil {
			return nil, fmt.Errorf("failed to parse input: %q, %w", currentInput, err)
		}

		commandResp, command, err := HandleCommand(inst, commandParts, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to handle command %s: %w", commandParts[0], err)
		}

		// fmt.Printf("command executed, response before processing: %q\n", commandResp)
		commandResp, err = inst.ProcessCommandResponse(connInput, currentInput[:inputLen], command, commandResp)
		resp += commandResp
		currentInput = currentInput[inputLen:]
	}

	return []byte(resp), nil
}

func HandleCommand(inst RedisInstance, commandParts []string, ctx *command.CommandContext) (string, command.Command, error) {
	command, err := inst.GetParser().ParseCommand(commandParts, ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse command %s: %w", commandParts[0], err)
	}

	commandResp, err := command.Execute()
	if err != nil {
		return "", nil, fmt.Errorf("command %s failed: %w", command, err)
	}

	return commandResp, command, nil
}

func RestoreFromRdb(inst RedisInstance) error {
	dir, ok := inst.GetConfig().Get(redisConfig.ConfigDir)
	if !ok {
		return fmt.Errorf("dir config value cannot be nil")
	}
	dbfilename, ok := inst.GetConfig().Get(redisConfig.ConfigDbFilename)
	if !ok {
		return fmt.Errorf("dbfilename config value cannot be nil")
	}

	filepath := fmt.Sprintf("%s/%s", dir, dbfilename)
	restorer, err := rdbRestorer.NewRdbRestorer(inst.GetStore())
	if err != nil {
		return fmt.Errorf("failed to create restorer: %w", err)
	}
	err = restorer.RestoreFromRdb(filepath)
	if err != nil {
		return fmt.Errorf("restoration from rdb failed: %w", err)
	}

	return nil
}
