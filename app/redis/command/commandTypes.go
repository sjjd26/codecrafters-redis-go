package command

import "fmt"

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
)

var commandName = map[CommandType]string{
	CommandUnknown: "UNKNOWN",
	CommandPing:    "PING",
	CommandEcho:    "ECHO",
	CommandSet:     "SET",
	CommandGet:     "GET",
	CommandConfig:  "CONFIG",
	CommandKeys:    "KEYS",
}

func (ct CommandType) String() string {
	return commandName[ct]
}

// --------------Command Constructors-----------------
type CommandConstructor func([]string) (Command, error)

var CommandConstructorMap = map[string]CommandConstructor{
	commandName[CommandPing]:   NewPing,
	commandName[CommandEcho]:   NewEcho,
	commandName[CommandSet]:    NewSet,
	commandName[CommandGet]:    NewGet,
	commandName[CommandConfig]: NewConfig,
	commandName[CommandKeys]:   NewKeysCommand,
}

// ------------------Command-------------------
type Command interface {
	// GetType() CommandType
	Handle() (string, error)
}

func CommandFactory(name string, args []string) (Command, error) {
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
