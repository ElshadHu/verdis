package command

import (
	"fmt"

	"github.com/ElshadHu/verdis/internal/protocol"
)

type Result = protocol.RESPValue

type Context struct {
	// TODO: Engine *mvcc.Engine
}

type Handler interface {
	// Execute processes the command and returns a RESP response
	Execute(ctx *Context, cmd *protocol.Command) Result
}

type HandlerFunc func(ctx *Context, cmd *protocol.Command) Result

func (f HandlerFunc) Execute(ctx *Context, cmd *protocol.Command) Result {
	return f(ctx, cmd)
}

// CommandSpec defines a command with Verdis-specific metadata
type CommandSpec struct {
	Name string

	// Handler executes the command
	Handler Handler

	// MinArgs is the minimum number of arguments
	MinArgs int

	// MaxArgs is the maximum number of arguments (-1 for unlimited)
	MaxArgs int

	// Description for documentation/debugging
	Description string

	// ReadOnly is true if this command only reads data
	ReadOnly bool

	// Mutates is true if this command writes data
	Mutates bool
}

// Validate if command argument meet the requirements
func (s *CommandSpec) Validate(cmd *protocol.Command) error {
	argc := len(cmd.Args())

	if argc < s.MinArgs {
		return &WrongArityError{Command: s.Name, Got: argc, Min: s.MinArgs}
	}

	if s.MaxArgs >= 0 && argc > s.MaxArgs {
		return &WrongArityError{Command: s.Name, Got: argc, Max: s.MaxArgs}
	}

	return nil
}

type WrongArityError struct {
	Command string
	Got     int
	Min     int
	Max     int
}

func (e *WrongArityError) Error() string {
	return fmt.Sprintf("ERR wrong number of arguments for %s command. Got %d, expects min: %d, max: %d", e.Command, e.Got, e.Min, e.Max)
}
