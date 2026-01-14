package protocol

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

type RESPValue any

type SimpleString struct {
	value string
}

func NewSimpleString(val string) *SimpleString {
	return &SimpleString{value: val}
}

func (s *SimpleString) Value() string {
	return s.value
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

type Integer struct {
	value int64
}

func NewInteger(val int64) *Integer {
	return &Integer{val}
}

func (i *Integer) Value() int64 {
	return i.value
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
