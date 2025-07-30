package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis/command"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/types"
)

var ErrAggregateLength = fmt.Errorf("failed to get length of aggregate type")
var ErrTypeByteCheck = fmt.Errorf("failed type byte check")

type RedisParser interface {
	ParseInput(input []byte) ([]interfaces.Command, error)
	GetAggregateLength(input []byte) (int, int, error)
}

type RedisParserImpl struct{}

func (_ *RedisParserImpl) GetAggregateLength(input []byte) (int, int, error) {
	len := 0
	p := 1

	// fmt.Printf("get legnth, p: %v \n", p)
	for ; input[p] != '\r'; p++ {
		digit, err := strconv.Atoi(string(input[p]))
		if err != nil {
			return -1, -1, fmt.Errorf("could not parse digit: %s", string(input[p]))
		}
		// fmt.Printf("digit: %v \n", digit)
		len = (len * 10) + digit
	}

	// Add 1 for RF
	return len, p + 2, nil
}

func (parser *RedisParserImpl) ParseBulkString(input []byte) (string, int, error) {
	// fmt.Printf("hello 2, %q \n", input)
	// fmt.Printf("parsing bulk string, input: %q \n", input)
	typeByte := input[0]
	err := types.CheckTypeByte(typeByte, types.RespBulkStr)
	if err != nil {
		return "", -1, fmt.Errorf("%w: %w", ErrTypeByteCheck, err)
	}

	// fmt.Println("passed type byte check")
	strLen, p, err := parser.GetAggregateLength(input)
	if err != nil {
		return "", -1, fmt.Errorf("%w: %w", ErrAggregateLength, err)
	}

	// fmt.Printf("str length: %v, p: %v \n", strLen, p)
	strEnd := p + strLen
	str := string(input[p:strEnd])

	// Add 2 for CLRF
	return str, strEnd + 2, nil
}

func (parser *RedisParserImpl) ParseArray(input []byte) ([]string, int, error) {
	// fmt.Println("parsing array")

	typeByte := input[0]
	err := types.CheckTypeByte(typeByte, types.RespArray)
	if err != nil {
		return nil, -1, fmt.Errorf("%w: %w", ErrTypeByteCheck, err)
	}

	// fmt.Println("getting length of array")
	arrayLen, p, err := parser.GetAggregateLength(input)
	if err != nil {
		return nil, -1, fmt.Errorf("%w: %w", ErrAggregateLength, err)
	}

	// fmt.Printf("array length: %v, p: %v \n", arrayLen, p)
	array := []string{}
	for i := 0; i < arrayLen && p < len(input); i++ {
		typeByte = input[p]
		// fmt.Printf("typeByte: %q \n", typeByte)
		respType, err := types.GetRespType(typeByte)
		if err != nil {
			return nil, -1, fmt.Errorf("%w", err)
		}
		if respType != types.RespBulkStr {
			return nil, -1, fmt.Errorf("RESP type (%v) not supported", respType)
		}

		// test := input[p:]
		// fmt.Printf("parsing bulk string, p: %v, test: %q \n", p, test)
		item, end, err := parser.ParseBulkString(input[p:])
		p += end
		if err != nil {
			return nil, -1, fmt.Errorf("failed to parse bulk string: %w", err)
		}

		// fmt.Printf("parsed bulk string: %v, p: %v \n", item, p)
		array = append(array, item)
	}

	if len(array) < arrayLen {
		return nil, -1, fmt.Errorf("could not parse the declared number of array items: got %v, expected %v", len(array), arrayLen)
	}

	return array, p, nil
}

func (_ *RedisParserImpl) ParseCommand(parts []string) (interfaces.Command, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	commandName := strings.ToUpper(parts[0])
	commandArgs := []string{}
	if len(parts) > 1 {
		commandArgs = parts[1:]
	}

	command, err := command.CommandFactory(commandName, commandArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to create command %s: %w", commandName, err)
	}
	return command, nil
}

// Expects full input string -> array consisting only of bulk strings
// E.g. *2\r\n $4\r\n LLEN\r\n $6\r\n mylist\r\n
// See Redis docs on protocol spec: https://redis.io/docs/latest/develop/reference/protocol-spec/
func (parser *RedisParserImpl) ParseInput(input []byte) ([]interfaces.Command, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("input is empty")
	}

	commands := []interfaces.Command{}
	p := 0

	for p < len(input) {
		commandParts, end, err := parser.ParseArray(input[p:])
		if err != nil {
			return nil, fmt.Errorf("failed to parse input array: %w", err)
		}
		p += end

		// fmt.Printf("p: %v, input len: %v, p < input len: %v\n", p, len(input), p < len(input))

		if len(commandParts) == 0 {
			return nil, fmt.Errorf("input array is empty")
		}

		command, err := parser.ParseCommand(commandParts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse command %s: %w", commandParts[0], err)
		}

		commands = append(commands, command)
	}

	return commands, nil
}
