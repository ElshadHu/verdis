package protocol

import (
	"bytes"
	"fmt"
)

type Serializer struct{}

func NewSerializer() *Serializer {
	return &Serializer{}
}

func (s *Serializer) serializeSimpleString(simpleString *SimpleString) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%c%s%s", TypeSimpleString, simpleString.Value(), CRLF)
	return buf.Bytes()
}

func (s *Serializer) serializeInteger(i *Integer) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%c%d%s", TypeInteger, i.Value(), CRLF)
	return buf.Bytes()
}

func (s *Serializer) serializeError(err *Error) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%c%s%s", TypeError, err.Msg(), CRLF)
	return buf.Bytes()
}

func (s *Serializer) serializeBulkString(bulkString *BulkString) []byte {
	var buf bytes.Buffer

	if bulkString.IsNull() {
		fmt.Fprintf(&buf, "%c-1%s", TypeBulkString, CRLF)
		return buf.Bytes()
	}
	data := bulkString.Data()
	fmt.Fprintf(&buf, "%c%d%s", TypeBulkString, len(data), CRLF)
	buf.Write(data)
	buf.WriteString(CRLF)
	return buf.Bytes()
}

func (s *Serializer) serializeArray(arr *Array) ([]byte, error) {
	var buf bytes.Buffer

	if arr.IsNull() {
		fmt.Fprintf(&buf, "%c-1%s", TypeArray, CRLF)
		return buf.Bytes(), nil
	}

	fmt.Fprintf(&buf, "%c%d%s", TypeArray, len(arr.Elements()), CRLF)
	for _, v := range arr.Elements() {
		serialized, err := s.Serialize(v)
		if err != nil {
			return nil, err
		}
		buf.Write(serialized)
	}

	return buf.Bytes(), nil
}

func (s *Serializer) Serialize(val RESPValue) ([]byte, error) {
	switch v := val.(type) {
	case SimpleString:
		return s.serializeSimpleString(&v), nil
	case Error:
		return s.serializeError(&v), nil
	case Integer:
		return s.serializeInteger(&v), nil
	case BulkString:
		return s.serializeBulkString(&v), nil
	case Array:
		return s.serializeArray(&v)
	default:
		return nil, fmt.Errorf("unsupported RESP type: %T", val)
	}
}
