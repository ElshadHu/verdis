package standard

import (
	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/protocol"
)

// Set stores a value under the given key
// Usage: SET [key] [value]
func Set(ctx *command.Context, cmd *protocol.Command) command.Result {
	key := string(cmd.Args()[0])
	value := cmd.Args()[1]

	ctx.Engine.Set(key, value)
	// Elshad: we can come back here boi
	// Dan: alright G
	return protocol.NewSimpleString("OK")
}

func SetSpec() *command.CommandSpec {
	return &command.CommandSpec{
		Name:        "SET",
		Handler:     command.HandlerFunc(Set),
		MinArgs:     2,
		MaxArgs:     2,
		Description: "Set key to value.",
		ReadOnly:    false,
		Mutates:     true,
	}
}
