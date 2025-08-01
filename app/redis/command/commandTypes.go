package command

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/redis/command/interfaces"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
)

type CommandContext struct {
	Store              store.RedisStore
	Config             redisConfig.RedisConfig
	ReplicationDetails *redisConfig.ReplicationDetails
	Conn               net.Conn
}

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
	CommandPsync
	CommandWait
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
	CommandPsync:             "PSYNC",
	CommandWait:              "WAIT",
}

func (ct CommandType) String() string {
	return commandName[ct]
}

// --------------Command Constructors-----------------
type CommandConstructor func([]string, *CommandContext) (interfaces.Command, error)

var CommandConstructorMap = map[string]CommandConstructor{
	commandName[CommandPing]:              NewPingCommand,
	commandName[CommandEcho]:              NewEchoCommand,
	commandName[CommandSet]:               NewSetCommand,
	commandName[CommandGet]:               NewGetCommand,
	commandName[CommandConfig]:            NewConfigCommand,
	commandName[CommandKeys]:              NewKeysCommand,
	commandName[CommandInfo]:              NewInfoCommand,
	commandName[CommandReplicationConfig]: NewReplicationConfigCommand,
	commandName[CommandPsync]:             NewPsyncCommand,
	commandName[CommandWait]:              NewWaitCommand,
}

// ------------------Command-------------------
func CommandFactory(name string, args []string, ctx *CommandContext) (interfaces.Command, error) {
	constructor, ok := CommandConstructorMap[name]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	return constructor(args, ctx)
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
