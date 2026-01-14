package protocol

import (
	"fmt"
	"strings"
)

type Serializer struct{}

func NewSerializer() *Serializer {
	return &Serializer{}
}

func (s *Serializer) serializeSimpleString(simpleString *SimpleString) string {
	return fmt.Sprintf("%c%s%s", TypeSimpleString, simpleString.Value(), CRLF)
}

func (s *Serializer) serializeInteger(i *Integer) string {
	return fmt.Sprintf("%c%d%s", TypeInteger, i.Value(), CRLF)
}

func (s *Serializer) serializeError(err *Error) string {
	return fmt.Sprintf("%c%s%s", TypeError, err.Msg(), CRLF)
}

func (s *Serializer) serializeBulkString(bulkString *BulkString) string {
	if bulkString.IsNull() {
		return fmt.Sprintf("%c-1%s", TypeBulkString, CRLF)
	}
	data := bulkString.Data()
	return fmt.Sprintf("%c%d%s%s%s", TypeBulkString, len(data), CRLF, string(data), CRLF)
}

func (s *Serializer) serializeArray(arr *Array) (string, error) {
	if arr.IsNull() {
		return fmt.Sprintf("%c-1%s", TypeArray, CRLF), nil
	}

	var result strings.Builder
	fmt.Fprintf(&result, "%c%d%s", TypeArray, len(arr.Elements()), CRLF)
	for _, v := range arr.Elements() {
		serialized, err := s.Serialize(v)
		if err != nil {
			return "", err
		}
		result.WriteString(serialized)
	}

	return result.String(), nil
}

func (s *Serializer) Serialize(val RESPValue) (string, error) {
	switch v := val.(type) {
	case *SimpleString:
		return s.serializeSimpleString(v), nil
	case *Error:
		return s.serializeError(v), nil
	case *Integer:
		return s.serializeInteger(v), nil
	case *BulkString:
		return s.serializeBulkString(v), nil
	case *Array:
		return s.serializeArray(v)
	default:
		return "", fmt.Errorf("unsupported RESP type: %T", val)
	}
}
