package standard

import (
	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/protocol"
)

// Ping responds with PONG or echoes the given argument.
// Usage: PING [message]
func Ping(ctx *command.Context, cmd *protocol.Command) command.Result {
	args := cmd.Args()
	if len(args) == 0 {
		return protocol.NewSimpleString("PONG")
	}
	return protocol.NewBulkString(args[0])
}

func PingSpec() *command.CommandSpec {
	return &command.CommandSpec{
		Name:        "PING",
		Handler:     command.HandlerFunc(Ping),
		MinArgs:     0,
		MaxArgs:     1,
		Description: "Simple health check. Returns PONG or echoes the argument.",
		ReadOnly:    true,
		Mutates:     false,
	}
}
