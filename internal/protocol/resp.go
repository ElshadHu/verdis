package protocol

import (
	"bufio"
	"fmt"
)

type RESPConnection struct {
	reader    *bufio.Reader
	writer    *bufio.Writer
	cmdParser *CommandParser
}

func NewRESPConnection(reader *bufio.Reader, writer *bufio.Writer) *RESPConnection {
	return &RESPConnection{
		reader:    reader,
		writer:    writer,
		cmdParser: NewCommandParser(reader),
	}
}

func (rc *RESPConnection) ReadCommand() (*Command, error) {
	return rc.cmdParser.ParseCommand()
}

func (rc *RESPConnection) WriteResponse(resp RESPValue) error {
	_, err := rc.writer.Write(resp.Serialize())
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return rc.writer.Flush()
}

func (rc *RESPConnection) Close() error {
	return rc.writer.Flush()
}
