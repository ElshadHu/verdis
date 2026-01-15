package protocol

import (
	"bufio"
	"fmt"
	"strings"
)

type Command struct {
	name string
	args [][]byte
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Args() [][]byte {
	return c.args
}

type CommandParser struct {
	respParser *RESPParser
}

func NewCommandParser(reader *bufio.Reader) *CommandParser {
	return &CommandParser{
		respParser: NewRESPParser(reader),
	}
}

func (cp *CommandParser) ParseCommand() (*Command, error) {
	b, err := cp.respParser.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("failed to detect command format: %w", err)
	}

	switch rune(b[0]) {
	case TypeArray:
		return cp.parseRESPCommand()
	default:
		return cp.parseInlineCommand()
	}
}

func (cp *CommandParser) parseRESPCommand() (*Command, error) {
	value, err := cp.respParser.ParseValue()
	if err != nil {
		return nil, fmt.Errorf("failed to parse RESP value: %w", err)
	}

	arr, ok := value.(*Array)
	if !ok {
		return nil, fmt.Errorf("expected array command, got %T", value)
	}

	return cp.extractCommandFromArray(arr)
}

func (cp *CommandParser) parseInlineCommand() (*Command, error) {
	line, err := cp.respParser.ReadCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to read inline command: %w", err)
	}

	// TODO: Support quoted arguments in inline format
	// Current: splits on whitespace only
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, fmt.Errorf("no command name in inline format")
	}

	commandName := strings.ToUpper(parts[0])
	args := make([][]byte, len(parts)-1)
	for i, part := range parts[1:] {
		args[i] = []byte(part)
	}

	return &Command{name: commandName, args: args}, nil
}

func (cp *CommandParser) extractCommandFromArray(arr *Array) (*Command, error) {
	if arr.IsNull() {
		return nil, fmt.Errorf("null array cannot be a command")
	}

	elements := arr.Elements()
	if len(elements) == 0 {
		return nil, fmt.Errorf("empty array cannot be a command")
	}

	cmdNameBulk, ok := elements[0].(*BulkString)
	if !ok {
		return nil, fmt.Errorf("command name must be bulk string, got %T", elements[0])
	}

	if cmdNameBulk.IsNull() {
		return nil, fmt.Errorf("command name cannot be null")
	}

	cmdName := strings.ToUpper(string(cmdNameBulk.Data()))
	if cmdName == "" {
		return nil, fmt.Errorf("command name cannot be empty")
	}

	args := make([][]byte, len(elements)-1)
	for i, elem := range elements[1:] {
		bulkArg, ok := elem.(*BulkString)
		if !ok {
			return nil, fmt.Errorf("command argument %d must be bulk string, got %T", i, elem)
		}

		if bulkArg.IsNull() {
			args[i] = nil
			continue
		}

		args[i] = bulkArg.Data()
	}

	return &Command{name: cmdName, args: args}, nil
}
