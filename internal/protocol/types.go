package protocol

import "fmt"

// RESP type markers
const (
	TypeSimpleString rune = '+'
	TypeError        rune = '-'
	TypeInteger      rune = ':'
	TypeBulkString   rune = '$'
	TypeArray        rune = '*'
)

// Line terminator (RESP protocol requirement)
const CRLF = "\r\n"

// RESPValue is the interface all RESP types implement
// [edit: from any to interface] exactly we cannot rely on any in runtime
type RESPValue interface {
	Serialize() []byte
}

type SimpleString struct {
	value string
}

func NewSimpleString(val string) *SimpleString {
	return &SimpleString{value: val}
}

func (s *SimpleString) Value() string {
	return s.value
}

func (s *SimpleString) Serialize() []byte {
	// lil bit less GC pressure
	return fmt.Appendf(nil, "%c%s%s", TypeSimpleString, s.value, CRLF)
}

type Error struct {
	msg string
}

func NewError(val string) *Error {
	return &Error{msg: val}
}

func (e *Error) Msg() string {
	return e.msg
}

func (e *Error) Serialize() []byte {
	return fmt.Appendf(nil, "%c%s%s", TypeError, e.msg, CRLF)
}

type Integer struct {
	value int64
}

func NewInteger(val int64) *Integer {
	return &Integer{val}
}

func (i *Integer) Value() int64 {
	return i.value
}

func (i *Integer) Serialize() []byte {
	return fmt.Appendf(nil, "%c%d%s", TypeInteger, i.value, CRLF)
}

type BulkString struct {
	data   []byte
	isNull bool
}

func NewBulkString(data []byte) *BulkString {
	return &BulkString{data: data, isNull: false}
}

func NewNullBulkString() *BulkString {
	return &BulkString{data: nil, isNull: true}
}

func (b *BulkString) Data() []byte {
	return b.data
}

func (b *BulkString) IsNull() bool {
	return b.isNull
}

func (b *BulkString) Serialize() []byte {
	if b.isNull {
		return fmt.Appendf(nil, "%c-1%s", TypeBulkString, CRLF)
	}
	buf := make([]byte, 0, len(b.data)+20)
	buf = fmt.Appendf(buf, "%c%d%s", TypeBulkString, len(b.data), CRLF)
	buf = append(buf, b.data...)
	buf = append(buf, '\r', '\n')
	return buf
}

type Array struct {
	elements []RESPValue
	isNull   bool
}

func NewArray(elements []RESPValue) *Array {
	return &Array{elements: elements, isNull: false}
}

func NewNullArray() *Array {
	return &Array{elements: nil, isNull: true}
}

func (a *Array) Elements() []RESPValue {
	return a.elements
}

func (a *Array) IsNull() bool {
	return a.isNull
}

func (a *Array) Serialize() []byte {
	if a.isNull {
		return fmt.Appendf(nil, "%c-1%s", TypeArray, CRLF)
	}
	buf := fmt.Appendf(nil, "%c%d%s", TypeArray, len(a.elements), CRLF)
	for _, elem := range a.elements {
		buf = append(buf, elem.Serialize()...)
	}
	return buf
}
