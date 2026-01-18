package version

import (
	"strconv"

	"github.com/ElshadHu/verdis/internal/command"
	"github.com/ElshadHu/verdis/internal/protocol"
)

// History returns version history for a key
// Usage: HISTORY [key]
func History(ctx *command.Context, cmd *protocol.Command) command.Result {
	args := cmd.Args()
	key := string(args[0])

	maxVersion := 0
	if len(args) > 1 {
		count, err := strconv.Atoi(string(args[1]))
		if err != nil || count < 0 {
			return protocol.NewError("ERR invalid count")
		}
		maxVersion = count
	}

	history, err := ctx.Engine.History(key, maxVersion)
	if err != nil {
		return protocol.NewNullBulkString()
	}
	// Return as array of arrays: [[version, timestamp, deleted, size], ...]
	result := make([]protocol.RESPValue, len(history))
	for i, info := range history {
		entry := []protocol.RESPValue{
			protocol.NewInteger(int64(info.Version)),
			protocol.NewInteger(info.Timestamp),
			protocol.NewInteger(boolToInt(info.Deleted)),
			protocol.NewInteger(int64(info.Size)),
		}
		result[i] = protocol.NewArray(entry)
	}
	return protocol.NewArray(result)
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func HistorySpec() *command.CommandSpec {
	return &command.CommandSpec{
		Name:        "HISTORY",
		Handler:     command.HandlerFunc(History),
		MinArgs:     1,
		MaxArgs:     2,
		Description: "Get version historyL History key [count]",
		ReadOnly:    true,
		Mutates:     false,
	}
}
