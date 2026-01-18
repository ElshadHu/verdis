package version

import (
	"strconv"

	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/protocol"
)

// GetVersion retrieves a key's value at a specific version.
// Usage: GETV [key] [version]
func GetVersion(ctx *command.Context, cmd *protocol.Command) command.Result {
	args := cmd.Args()
	key := string(args[0])
	versionStr := string(args[1])

	version, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		return protocol.NewError("ERR invalid version number: " + versionStr)
	}

	value, err := ctx.Engine.GetAtVersion(key, version)
	if err != nil {
		// return nil for not found
		return protocol.NewNullBulkString()
	}

	return protocol.NewBulkString(value)
}

func GetVersionSpec() *command.CommandSpec {
	return &command.CommandSpec{
		Name:        "GETV",
		Handler:     command.HandlerFunc(GetVersion),
		MinArgs:     2,
		MaxArgs:     2,
		Description: "Get value at a specific version.",
		ReadOnly:    true,
		Mutates:     false,
	}
}
