package redis

import (
	"fmt"
	"strconv"
	"strings"
)

var ErrAggregateLength = fmt.Errorf("failed to get length of aggregate type")
var ErrTypeByteCheck = fmt.Errorf("failed type byte check")

func getAggregateLength(input []byte, start int) (int, int, error) {
	len := 0
	p := start + 1

	fmt.Printf("get legnth, p: %v \n", p)
	for ; input[p] != '\r'; p++ {
		digit, err := strconv.Atoi(string(input[p]))
		if err != nil {
			return -1, -1, fmt.Errorf("could not parse digit: %s", string(input[p]))
		}
		fmt.Printf("digit: %v \n", digit)
		len += (len * 10) + digit
	}

	// Add 1 for RF
	return len, p + 2, nil
}

func parseBulkString(input []byte, start int) (string, int, error) {
	// fmt.Printf("hello 2, %q \n", input)
	fmt.Printf("parsing bulk string, input: %q, start: %v \n", input, start)
	typeByte := input[start]
	err := CheckTypeByte(typeByte, RespBulkStr)
	if err != nil {
		return "", -1, fmt.Errorf("%w: %w", ErrTypeByteCheck, err)
	}

	fmt.Println("passed type byte check")
	strLen, p, err := getAggregateLength(input, start)
	if err != nil {
		return "", -1, fmt.Errorf("%w: %w", ErrAggregateLength, err)
	}

	fmt.Printf("str length: %v, p: %v \n", strLen, p)
	strEnd := p + strLen
	str := string(input[p:strEnd])

	// Add 2 for CLRF
	return str, strEnd + 2, nil
}

func parseArray(input []byte, start int) ([]string, int, error) {
	fmt.Println("parsing array")

	typeByte := input[start]
	err := CheckTypeByte(typeByte, RespArray)
	if err != nil {
		return nil, -1, fmt.Errorf("%w: %w", ErrTypeByteCheck, err)
	}

	fmt.Println("getting length of array")
	arrayLen, p, err := getAggregateLength(input, start)
	if err != nil {
		return nil, -1, fmt.Errorf("%w: %w", ErrAggregateLength, err)
	}

	fmt.Printf("array length: %v, p: %v \n", arrayLen, p)
	array := []string{}
	for i := 0; i < arrayLen && p < len(input); i++ {
		typeByte = input[p]
		respType, err := GetRespType(typeByte)
		if err != nil {
			return nil, -1, fmt.Errorf("%w", err)
		}
		if respType != RespBulkStr {
			return nil, -1, fmt.Errorf("RESP type (%v) not supported", respType)
		}

		test := input[p:]
		fmt.Printf("parsing bulk string, p: %v, test: %q \n", p, test)
		var item string
		fmt.Println("hello")
		item, p, err = parseBulkString(input, p)
		if err != nil {
			return nil, -1, fmt.Errorf("failed to parse bulk string: %w", err)
		}

		array = append(array, item)
	}

	if len(array) < arrayLen {
		return nil, -1, fmt.Errorf("could not parse the declared number of array items: got %v, expected %v", len(array), arrayLen)
	}

	return array, p, nil
}

func parseCommandString(commandStr string) (CommandSpec, error) {
	if commandSpec, ok := CommandSpecMap[strings.ToUpper(commandStr)]; ok {
		return commandSpec, nil
	}
	return CommandSpecUnknown, fmt.Errorf("unknown command: %s", commandStr)
}

// Expects full input string -> array consisting only of bulk strings
// E.g. *2\r\n $4\r\n LLEN\r\n $6\r\n mylist\r\n
// See Redis docs on protocol spec: https://redis.io/docs/latest/develop/reference/protocol-spec/
func ParseInput(input []byte) ([]Command, error) {
	bulkStrArray, p, err := parseArray(input, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	fmt.Printf("p: %v, input len: %v", p, len(input))
	if p > len(input)+1 {
		return nil, fmt.Errorf("input has remaining data after parsing, extra %v bytes than expected %v", p-len(input), len(input))
	}

	commands := []Command{}
	for i := 0; i < len(bulkStrArray); i++ {
		commandSpec, err := parseCommandString(bulkStrArray[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse command string: %w", err)
		}
		i++

		args := []string{}
		for j := 0; j < commandSpec.ArgCount && i < len(bulkStrArray); j++ {
			args = append(args, bulkStrArray[i])
		}

		if commandSpec.ArgCount != len(args) {
			return nil, fmt.Errorf("not enough arguments for command %s, got %v, expected %v", commandSpec.Type, commandSpec.ArgCount, len(args))
		}

		command := NewCommand(commandSpec.Type, args)
		commands = append(commands, *command)
	}

	fmt.Printf("commands: %v \n", commands)
	return commands, nil
}
