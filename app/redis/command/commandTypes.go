package command

import (
	"fmt"

	infocommand "github.com/codecrafters-io/redis-starter-go/app/redis/command/infoCommand"
	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	replicationconfig "github.com/codecrafters-io/redis-starter-go/app/redis/command/replicationConfigCommand"
)

// ---------------Command Type----------------
type CommandType int

const (
	CommandUnknown CommandType = iota
	CommandPing
	CommandEcho
	CommandSet
	CommandGet
	CommandConfig
	CommandKeys
	CommandInfo
	CommandReplicationConfig
)

var commandName = map[CommandType]string{
	CommandUnknown:           "UNKNOWN",
	CommandPing:              "PING",
	CommandEcho:              "ECHO",
	CommandSet:               "SET",
	CommandGet:               "GET",
	CommandConfig:            "CONFIG",
	CommandKeys:              "KEYS",
	CommandInfo:              "INFO",
	CommandReplicationConfig: "REPLCONF",
}

func (ct CommandType) String() string {
	return commandName[ct]
}

// --------------Command Constructors-----------------
type CommandConstructor func([]string) (interfaces.Command, error)

var CommandConstructorMap = map[string]CommandConstructor{
	commandName[CommandPing]:              NewPing,
	commandName[CommandEcho]:              NewEcho,
	commandName[CommandSet]:               NewSet,
	commandName[CommandGet]:               NewGet,
	commandName[CommandConfig]:            NewConfig,
	commandName[CommandKeys]:              NewKeysCommand,
	commandName[CommandInfo]:              infocommand.NewInfoCommand,
	commandName[CommandReplicationConfig]: replicationconfig.NewReplicationConfigCommand,
}

// ------------------Command-------------------
func CommandFactory(name string, args []string) (interfaces.Command, error) {
	constructor, ok := CommandConstructorMap[name]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	return constructor(args)
}

// -----------------------------------------

type KeyValuePair struct {
	Key   string
	Value string
}

type CommandArguments interface {
	GetPositionalArgs() []KeyValuePair
	GetNamedArgs() map[string]string
	GetVariadicArgs() []string
}

type CommandArgumentsSpec interface {
	GetPositionalArgKeys() []string
	GetNamedArgs() map[string]string
}

// --------------------------------------
