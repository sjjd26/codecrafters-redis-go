package main

import (
	"errors"
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

func handleConnection(conn net.Conn) {
	defer conn.Close()

	readBuf := make([]byte, 1024)
	for n, err := conn.Read(readBuf); n != 0; n, err = conn.Read(readBuf) {
		if err != nil {
			return
		}

		commands, err := parseRedisInput(readBuf)
		if err != nil {
			fmt.Println("Error handling connection: ", err.Error())
		}

		for _, command := range commands {
			handler := routeCommand(command)
			handler(conn, command)
		}
	}
}

func routeCommand(command RedisCommand) func(conn net.Conn, command RedisCommand) {
	if command.Type == CommandPing {
		return handlePing
	} else if command.Type == CommandEcho {
		return handleEcho
	}
}

func handlePing(conn net.Conn, _ RedisCommand) {
	pong := "+PONG\r\n"
	conn.Write([]byte(pong))
}

func handleEcho(conn net.Conn, command RedisCommand) {

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
	arrayLength, p := getLengthOfAggregate(input, p)

	for ; p < arrayLength; p++ {
		bulkStringLength, p := getLengthOfAggregate(input, p)
		commandString := input[p:bulkStringLength]

		if commandType == CommandUnknown {
			return []RedisCommand{}, errors.New("Could not parse: found unknown redis command")
		}

		// Extend to determine the number of arguments
		command := NewRedisCommand(commandType, []string{})
		commands = append(commands, *command)
	}

	return commands, nil
}

// Expects the byte slice to include the aggregate type character at the beginning
// Returns length of the aggregate as well as the index of the CRLF delimeter
func getLengthOfAggregate(input []byte, start int) (int, int) {
	len := 0
	var p int

	for p = start; input[p] != '\r'; p++ {
		len += (len * 10) + (int(input[p]))
	}

	return len, p + 2
}
