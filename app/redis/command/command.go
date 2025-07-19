package command

// ---------------Command Type----------------
type CommandType int

const (
	CommandUnknown CommandType = iota
	CommandPing
	CommandEcho
	CommandSet
	CommandGet
)

var commandName = map[CommandType]string{
	CommandUnknown: "UNKNOWN",
	CommandPing:    "PING",
	CommandEcho:    "ECHO",
	CommandSet:     "SET",
	CommandGet:     "GET",
}

func (ct CommandType) String() string {
	return commandName[ct]
}

var commandTypeMap = map[string]CommandType{
	commandName[CommandPing]: CommandPing,
	commandName[CommandEcho]: CommandEcho,
	commandName[CommandSet]:  CommandSet,
	commandName[CommandGet]:  CommandGet,
}

// --------------Command Spec-----------------
type CommandConstructor func([]string) (Command, error)

type CommandSpec struct {
	Name        string
	Constructor CommandConstructor
	Type        CommandType
	MinArgs     int
	MaxArgs     int
	Variadic    bool
}

var CommandSpecMap = map[string]CommandSpec{
	commandName[CommandPing]: {Name: commandName[CommandPing], Constructor: NewPing, Type: CommandPing, MinArgs: 0, MaxArgs: 1},
	commandName[CommandEcho]: {Name: commandName[CommandEcho], Constructor: NewEcho, Type: CommandEcho, MinArgs: 1, MaxArgs: 1},
	commandName[CommandSet]:  {Name: commandName[CommandSet], Constructor: NewSet, Type: CommandSet, MinArgs: 2, MaxArgs: 2, Variadic: true},
	commandName[CommandGet]:  {Name: commandName[CommandGet], Constructor: NewGet, Type: CommandGet, MinArgs: 1, MaxArgs: 1, Variadic: true},
}

var CommandSpecUnknown = CommandSpec{Name: "UNKNOWN", Type: CommandUnknown}

// ------------------Command-------------------
type Command interface {
	// GetType() CommandType
	Handle() (string, error)
}

// ------------------Common responses-------------
const respOk = "+OK\r\n"
const respNull = "$-1\r\n"
