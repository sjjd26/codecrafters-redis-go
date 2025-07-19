package redis

// Command Type
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
	commandName[CommandPing]: CommandPing,
	commandName[CommandEcho]: CommandEcho,
}

// Command Spec
type CommandSpec struct {
	Name     string
	Type     CommandType
	ArgCount int
	Variadic bool
}

var CommandSpecMap = map[string]CommandSpec{
	commandName[CommandPing]: {Name: commandName[CommandPing], Type: CommandPing, ArgCount: 0},
	commandName[CommandEcho]: {Name: commandName[CommandEcho], Type: CommandEcho, ArgCount: 1, Variadic: true},
}

var CommandSpecUnknown = CommandSpec{Name: "UNKNOWN", Type: CommandUnknown}

// Command
type Command struct {
	Type CommandType
	Args []string
}

func NewCommand(commandType CommandType, args []string) *Command {
	command := Command{Type: commandType, Args: args}
	return &command
}
