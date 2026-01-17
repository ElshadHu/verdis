package standard

import (
	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/protocol"
)

// Exists check if one or more keys exist.
// Usage: EXISTS [key ...]
func Exists(ctx *command.Context, cmd *protocol.Command) command.Result {
	args := cmd.Args()
	count := int64(0)
	for _, arg := range args {
		key := string(arg)
		if ctx.Engine.Exists(key) {
			count++
		}
	}
	return protocol.NewInteger(count)
}

func ExistsSpec() *command.CommandSpec {
	return &command.CommandSpec{
		Name:        "EXISTS",
		Handler:     command.HandlerFunc(Exists),
		MinArgs:     1,
		MaxArgs:     -1,
		Description: "Check if keys exists",
		ReadOnly:    true,
		Mutates:     false,
	}
}
