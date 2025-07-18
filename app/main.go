package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
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

		fmt.Println("Handling connection")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	readBuf := make([]byte, 1024)
	for n, err := conn.Read(readBuf); n != 0; n, err = conn.Read(readBuf) {
		if err != nil {
			return
		}

		commands, err := parseRedisInput(readBuf)
		if err != nil {
			fmt.Printf("Could not parse input: %q, %s \n", string(readBuf), err.Error())
			return
		}

		// fmt.Printf("commands: %v\n", commands)
		for _, command := range commands {
			handler, err := routeCommand(command)
			if err != nil {
				fmt.Println("Could not route command: ", err.Error())
				return
			}
			fmt.Printf("got handler for command: %v \n", command)
			handler(conn, command)
		}
	}

	fmt.Println("Finished with connection")
}

func routeCommand(command RedisCommand) (func(conn net.Conn, command RedisCommand), error) {
	if command.Type == CommandPing {
		return handlePing, nil
	} else if command.Type == CommandEcho {
		return handleEcho, nil
	}
	return nil, fmt.Errorf("failed to route command: %s", command.Type)
}

func handlePing(conn net.Conn, _ RedisCommand) {
	pong := "+PONG\r\n"
	conn.Write([]byte(pong))
}

func handleEcho(conn net.Conn, command RedisCommand) {
	bulkStr := createBulkString(strings.Join(command.Args, ""))
	fmt.Printf("writing: %q, command: %v, args: %q \n", bulkStr, command, command.Args)
	conn.Write([]byte(bulkStr))
}

func createBulkString(str string) string {
	return fmt.Sprintf("$%v\r\n%s\r\n", len(str), str)
}

type CommandType int

const (
	CommandUnknown CommandType = iota
	CommandPing
	CommandEcho
)

var commandName = map[CommandType]string{
	CommandUnknown: "UNKNOWN",
	CommandPing:    "PING",
	CommandEcho:    "ECHO",
}

func (ct CommandType) String() string {
	return commandName[ct]
}

var commandTypeMap = map[string]CommandType{
	"PING": CommandPing,
	"ECHO": CommandEcho,
}

type CommandSpec struct {
	Name     string
	Type     CommandType
	ArgCount int
	Variadic bool
}

var commandSpecMap = map[string]CommandSpec{
	"PING": {Name: "PING", Type: CommandPing, ArgCount: 0},
	"ECHO": {Name: "ECHO", Type: CommandEcho, ArgCount: 1, Variadic: true},
}

type RedisCommand struct {
	Type CommandType
	Args []string
}

func NewRedisCommand(commandType CommandType, args []string) *RedisCommand {
	command := RedisCommand{Type: commandType, Args: args}
	return &command
}

func parseCommandInput(input []byte, p int) (CommandType, int) {
	if val, ok := commandTypeMap[strings.ToUpper(string(input))]; ok {
		return val, p
	}
	return CommandUnknown, p
}

// Expects full input string -> array consisting only of bulk strings
// E.g. *2\r\n $4\r\n LLEN\r\n $6\r\n mylist\r\n
// See Redis docs on protocol spec: https://redis.io/docs/latest/develop/reference/protocol-spec/
func parseRedisInput(input []byte) ([]RedisCommand, error) {
	commands := make([]RedisCommand, 0, 10)
	p := 0
	arrayLength, p, err := getLengthOfAggregate(input, p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse array length: %s", err.Error())
	}

	// fmt.Printf("arrLen: %v, p: %v \n", arrayLength, p)
	for i := 0; i < arrayLength && p < len(input); i++ {
		bulkStringLength, p, err := getLengthOfAggregate(input, p)
		bulkStrEnd := p + bulkStringLength
		if err != nil {
			return nil, fmt.Errorf("failed to parse bulk string command length: %s", err.Error())
		}

		// fmt.Printf("p: %v, end: %v, len: %v \n", p, bulkStrEnd, bulkStringLength)
		commandString := strings.ToUpper(string(input[p:bulkStrEnd]))
		// p = next

		commandSpec, ok := commandSpecMap[commandString]
		// fmt.Printf("%v, %v\n", strings.Compare(commandString, "ECHO"), commandString == "ECHO")
		if !ok {
			return nil, fmt.Errorf("unknown command: %s, %v", commandString, bulkStringLength)
		}

		p = bulkStrEnd + 2 // CLRF
		args := []string{}
		for j := 0; j < commandSpec.ArgCount && p < len(input); j++ {
			argLength, p, err := getLengthOfAggregate(input, p)
			argEnd := argLength + p
			if err != nil {
				return nil, fmt.Errorf("failed to parse bulk string argument length: %s", err.Error())
			}

			arg := string(input[p:argEnd])
			fmt.Printf("got arg: %q \n", arg)
			args = append(args, arg)
			p = argEnd + 2 // CLRF
			i++
		}

		if !commandSpec.Variadic && len(args) != commandSpec.ArgCount {
			return nil, fmt.Errorf("invalid number of args for %s", commandSpec.Name)
		}

		command := NewRedisCommand(commandSpec.Type, args)

		commands = append(commands, *command)
		// fmt.Printf("command: %v commands: %v \n", command, commands)
	}

	return commands, nil
}

// Expects the byte slice to include the aggregate type character at the beginning
// Returns length of the aggregate as well as the index of the CRLF delimeter
func getLengthOfAggregate(input []byte, start int) (int, int, error) {
	len := 0
	p := start + 1

	for ; input[p] != '\r'; p++ {
		digit, err := strconv.Atoi(string(input[p]))
		if err != nil {
			return -1, -1, fmt.Errorf("could not parse digit: %s", string(input[p]))
		}
		// fmt.Printf("digit: %v \n", digit)
		len += (len * 10) + digit
	}

	return len, p + 2, nil
}
