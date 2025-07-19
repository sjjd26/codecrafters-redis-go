package command

import "fmt"

var ErrTooManyArgs = fmt.Errorf("too many arguments supplied")
var ErrNotEnoughArgs = fmt.Errorf("not enough arguments supplied")
