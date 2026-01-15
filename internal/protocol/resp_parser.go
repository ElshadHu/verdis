package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	MAX_ARRAY_SIZE  = 1_000_000
	MAX_BULK_STRING = 512 * 1024 * 1024 // 512 MB
)

type RESPParser struct {
	reader *bufio.Reader
}

func NewRESPParser(reader *bufio.Reader) *RESPParser {
	return &RESPParser{reader: reader}
}

func (p *RESPParser) ReadCommand() (string, error) {
	return p.readLine()
}

func (p *RESPParser) Peek(n int) ([]byte, error) {
	b, err := p.reader.Peek(n)
	if err != nil {
		return nil, fmt.Errorf("failed to peek")
	}
	return b, nil
}

func (p *RESPParser) ParseValue() (RESPValue, error) {
	typeByte, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch rune(typeByte) {
	case TypeArray:
		return p.parseArray()
	case TypeBulkString:
		return p.parseBulkString()
	case TypeSimpleString:
		return p.parseSimpleString()
	case TypeInteger:
		return p.parseInteger()
	case TypeError:
		return p.parseError()
	default:
		// TODO: Create error types for consistent error messaging
		return nil, fmt.Errorf("invalid RESP type marker: %c", typeByte)
	}
}

func (p *RESPParser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err // Let caller wrap with context
	}

	// Validate CRLF
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", fmt.Errorf("invalid protocol: expected CRLF, got %q", line[len(line)-2:])
	}

	// Return line WITHOUT \r\n
	return line[:len(line)-2], nil
}

func (p *RESPParser) parseArray() (*Array, error) {
	line, err := p.readLine()
	count, err := strconv.Atoi(line)
	if err != nil {
		return nil, fmt.Errorf("invalid array count: %w", err)
	}

	if count == -1 {
		return &Array{isNull: true}, nil
	}
	if count < -1 {
		return nil, fmt.Errorf("invalid array length: %d", count)
	}
	if count > MAX_ARRAY_SIZE {
		return nil, fmt.Errorf("array too large: %d exceeds max %d", count, MAX_ARRAY_SIZE)
	}
	if count == 0 {
		return &Array{elements: []RESPValue{}}, nil
	}

	elements := make([]RESPValue, 0, count)

	for i := 0; i < count; i++ {
		elem, err := p.ParseValue()
		if err != nil {
			return nil, fmt.Errorf("failed to parse array element %d: %w", i, err)
		}
		elements = append(elements, elem)
	}

	return &Array{elements: elements}, nil
}

func (p *RESPParser) parseBulkString() (*BulkString, error) {
	line, err := p.readLine()
	length, err := strconv.Atoi(line)
	if err != nil {
		return nil, fmt.Errorf("invalid bulk string length: %w", err)
	}

	if length == -1 {
		return &BulkString{isNull: true}, nil
	}

	if length < -1 {
		return nil, fmt.Errorf("invalid bulk string length: %d", length)
	}
	if length > MAX_BULK_STRING {
		return nil, fmt.Errorf("bulk string too large: %d bytes", length)
	}

	if length == 0 {
		if err := p.discardCRLF(); err != nil {
			return nil, err
		}
		return &BulkString{data: []byte{}}, nil
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(p.reader, data); err != nil {
		return nil, fmt.Errorf("failed to read bulk string data: %w", err)
	}
	if err := p.discardCRLF(); err != nil {
		return nil, err
	}

	return &BulkString{data: data}, nil
}

func (p *RESPParser) parseInteger() (*Integer, error) {
	line, err := p.readLine()
	num, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer format: %w", err)
	}

	return &Integer{value: num}, nil
}

func (p *RESPParser) parseSimpleString() (*SimpleString, error) {
	line, err := p.readLine()
	return &SimpleString{value: line}, err
}

func (p *RESPParser) parseError() (*Error, error) {
	line, err := p.readLine()
	return &Error{msg: line}, err
}

func (p *RESPParser) discardCRLF() error {
	cr, err := p.reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read CR: %w", err)
	}
	if cr != '\r' {
		return fmt.Errorf("expected CR, got: %c", cr)
	}

	lf, err := p.reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read LF: %w", err)
	}
	if lf != '\n' {
		return fmt.Errorf("expected LF, got: %c", lf)
	}

	return nil
}
