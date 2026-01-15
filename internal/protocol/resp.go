package protocol

import (
	"bufio"
	"fmt"
)

type RESPConnection struct {
	// reader from TCP socket
	reader *bufio.Reader
	// writer from TCP socket
	writer *bufio.Writer

	// serializer marshals responses
	serializer *Serializer
	// cmdParser extracts commands
	cmdParser *CommandParser
}

func NewRESPConnection(reader *bufio.Reader, writer *bufio.Writer) *RESPConnection {
	return &RESPConnection{
		reader:     reader,
		writer:     writer,
		serializer: NewSerializer(),
		cmdParser:  NewCommandParser(reader),
	}
}

func (rc *RESPConnection) ReadCommand() (*Command, error) {
	return rc.cmdParser.ParseCommand()
}

func (rc *RESPConnection) WriteResponse(resp *RESPValue) error {
	serialized, err := rc.serializer.Serialize(*resp)
	if err != nil {
		return fmt.Errorf("failed to serialize response: %w", err)
	}

	// write bytes to buffered writer
	_, err = rc.writer.Write(serialized)
	if err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	// flush to send to client
	err = rc.writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush response: %w", err)
	}
	return nil
}

func (rc *RESPConnection) Close() error {
	if err := rc.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	return nil
}
