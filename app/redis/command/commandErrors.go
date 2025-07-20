package command

import "fmt"

var ErrTooManyArgs = fmt.Errorf("too many arguments supplied")
var ErrNotEnoughArgs = fmt.Errorf("not enough arguments supplied")
var ErrUnknownSubCommand = fmt.Errorf("unknown subcommand")
var ErrNotImplemented = func(command string) error {
	return fmt.Errorf("command %s is not implemented", command)
}
