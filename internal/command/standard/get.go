package standard

import (
	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/protocol"
)

// Get retrieves a key's value.
// Usage: GET [key]
func Get(ctx *command.Context, cmd *protocol.Command) command.Result {
	key := string(cmd.Args()[0])
	value, exists := ctx.Engine.Get(key)
	if !exists {
		return protocol.NewNullBulkString()
	}

	return protocol.NewBulkString(value)
}

func GetSpec() *command.CommandSpec {
	return &command.CommandSpec{
		Name:        "GET",
		Handler:     command.HandlerFunc(Get),
		MinArgs:     1,
		MaxArgs:     1,
		Description: "Get the value of a key.",
		ReadOnly:    true,
		Mutates:     false,
	}
}
