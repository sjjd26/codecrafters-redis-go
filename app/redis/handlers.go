package redis

import (
	"fmt"
	"strings"
)

var ErrHandlerNotFound = fmt.Errorf("failed to route command to a handler")

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

func createBulkString(str string) string {
	return fmt.Sprintf("$%v\r\n%s\r\n", len(str), str)
}
