package standard

import (
	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/protocol"
)

// Del deletes one or more keys.
// Usage: DEL key [key ...]
func Del(ctx *command.Context, cmd *protocol.Command) command.Result {
	keys := cmd.Args()
	deleted := int64(0)
	for _, key := range keys {
		existed := ctx.Engine.Del(string(key))
		if existed {
			deleted++
		}
	}
	return protocol.NewInteger(deleted)
}

func DelSpec() *command.CommandSpec {
	return &command.CommandSpec{
		Name:        "DEL",
		Handler:     command.HandlerFunc(Del),
		MinArgs:     1,
		MaxArgs:     -1, // unlimited
		Description: "Deletes one or more keys.",
		ReadOnly:    false,
		Mutates:     true,
	}
}
