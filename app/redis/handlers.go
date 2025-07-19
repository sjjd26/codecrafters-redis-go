package redis

import (
	"fmt"
	"strings"
)

var dataStore = make(map[string]string)

var ErrHandlerNotFound = fmt.Errorf("failed to route command to a handler")

const respOk = "+OK\r\n"
const respNull = "$-1\r\n"

func HandleCommand(command Command) (string, error) {
	handler, ok := routeMapper[command.Type]
	if !ok {
		return "", fmt.Errorf("error handling command: %w", ErrHandlerNotFound)
	}

	resp, err := handler(command)
	if err != nil {
		return "", fmt.Errorf("error handling command: %w", err)
	}
	return resp, nil
}

var routeMapper = map[CommandType]func(Command) (string, error){
	CommandPing: handlePing,
	CommandEcho: handleEcho,
	CommandGet:  handleGet,
	CommandSet:  handleSet,
}

func handlePing(_ Command) (string, error) {
	pong := "+PONG\r\n"
	return pong, nil
}

func handleEcho(command Command) (string, error) {
	bulkStr := createBulkString(strings.Join(command.Args, ""))
	// fmt.Printf("writing: %q, command: %v, args: %q \n", bulkStr, command, command.Args)
	return bulkStr, nil
}

func handleSet(command Command) (string, error) {
	if len(command.Args) < 2 {
		return "", fmt.Errorf("not enough command arguments: expected > 2, got %v", len(command.Args))
	}

	key := command.Args[0]
	value := command.Args[1]
	dataStore[key] = value

	return respOk, nil
}

func handleGet(command Command) (string, error) {
	key := command.Args[0]
	if value, ok := dataStore[key]; ok {
		return createBulkString(value), nil
	}
	return respNull, nil
}

func createBulkString(str string) string {
	return fmt.Sprintf("$%v\r\n%s\r\n", len(str), str)
}
