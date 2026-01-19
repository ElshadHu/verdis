package protocol

import (
	"bufio"
	"bytes"
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"testing"
)

// FuzzRESPParser is the main fuzz target
// The main goal: parser must never panic on any input

func FuzzRESPParser(f *testing.F) {
	// seed with valid RESP to guide the fuzzer
	seeds := [][]byte{
		[]byte("+OK\r\n"),
		[]byte("-ERR\r\n"),
		[]byte(":123\r\n"),
		[]byte("$5\r\nhello\r\n"),
		[]byte("$-1\r\n"),
		[]byte("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"),
		[]byte("*-1\r\n"),
		[]byte("*0\r\n"),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		reader := bufio.NewReader(bytes.NewReader(data))
		parser := NewRESPParser(reader)
		// must not panic. Errors are expected for invalid input
		_, _ = parser.ParseValue()
	})
}

// FuzzBulStringBoundary targets parseBulkString length handling
func FuzzBulkStringBoundary(f *testing.F) {
	f.Add(int64(-2))
	f.Add(int64(-1))
	f.Add(int64(0))
	f.Add(int64(5))
	f.Add(int64(MAX_BULK_STRING))
	f.Add(int64(MAX_BULK_STRING + 1))
	f.Fuzz(func(t *testing.T, length int64) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "$%d\r\n", length)
		// Add matching data for valid lengths
		if length > 0 && length <= 1000 {
			buf.Write(bytes.Repeat([]byte("x"), int(length)))
			buf.WriteString("\r\n")
		}
		reader := bufio.NewReader(&buf)
		parser := NewRESPParser(reader)
		_, _ = parser.ParseValue()
	})
}

func FuzzArrayBoundary(f *testing.F) {
	f.Add(int64(-2))
	f.Add(int64(-1))
	f.Add(int64(0))
	f.Add(int64(10))
	f.Add(int64(MAX_ARRAY_SIZE))
	f.Add(int64(MAX_ARRAY_SIZE + 1))
	f.Fuzz(func(t *testing.T, count int64) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "*%d\r\n", count)

		if count > 0 && count <= 50 {
			for i := int64(0); i < count; i++ {
				buf.WriteString(":0\r\n")
			}
		}
		reader := bufio.NewReader(&buf)
		parser := NewRESPParser(reader)
		_, _ = parser.ParseValue()

	})
}

// Serialize -> Parse -> Serialize must equal original

func TestSerializeParseRoundTrip(t *testing.T) {
	cases := []RESPValue{
		NewSimpleString("OK"),
		NewSimpleString(""),
		NewError("ERR unknown"),
		NewInteger(0),
		NewInteger(-9223372036854775808),
		NewInteger(9223372036854775807),
		NewBulkString([]byte("hello")),
		NewBulkString([]byte{}),
		NewBulkString([]byte("line1\r\nline2")),
		NewNullBulkString(),
		NewArray([]RESPValue{NewInteger(1), NewInteger(2)}),
		NewArray([]RESPValue{}),
		NewNullArray(),
	}

	// what if generate 1000 random buk strings with binary data , let's see

	for i := 0; i < 1000; i++ {
		size := rand.Intn(500)
		data := make([]byte, size)
		crand.Read(data)
		cases = append(cases, NewBulkString(data))
	}
	// Generate 100 random arrays
	for i := 0; i < 100; i++ {
		elemCount := rand.Intn(10)
		elems := make([]RESPValue, elemCount)
		for j := 0; j < elemCount; j++ {
			elems[j] = NewInteger(rand.Int63())
		}
		cases = append(cases, NewArray(elems))
	}

	for i, original := range cases {
		serialized := original.Serialize()

		reader := bufio.NewReader(bytes.NewReader(serialized))
		parser := NewRESPParser(reader)
		parsed, err := parser.ParseValue()
		if err != nil {
			t.Errorf("case %d: parse failed %v", i, err)
			continue
		}

		reserialized := parsed.Serialize()
		if !bytes.Equal(serialized, reserialized) {
			t.Errorf("case %d: round-trip mismatch\n got %q\n want: %q", i, reserialized, serialized)
		}
	}
}

// TestMalformedInput targets specific error paths

func TestMalformedInput(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
	}{
		// Unknown type marker
		{"unknown_type", []byte("^hello\r\n")},
		{"empty_input", []byte{}},
		{"just_crlf", []byte("\r\n")},
		{"null_byte_as_type", []byte("\x00hello\r\n")},
		{"all_control_chars", []byte("\x01\x02\x03\r\n")},

		// Integer overflow
		{"int_overflow_positive", []byte(":99999999999999999999\r\n")},
		{"int_overflow_negative", []byte(":-99999999999999999999\r\n")},

		// Bulk string edge cases
		{"bulk_negative_overflow", []byte("$-9999999999999999999\r\n")},
		{"bulk_length_with_space", []byte("$ 5\r\nhello\r\n")},
		{"bulk_data_too_long", []byte("$3\r\nhello\r\n")}, // says 3, gives 5

		// Array edge cases
		{"array_count_overflow", []byte("*99999999999999999999\r\n")},
		{"array_with_non_resp_element", []byte("*1\r\nNOTRESP\r\n")},

		// Nested bombs
		{"deeply_nested_incomplete", []byte("*1\r\n*1\r\n*1\r\n*1\r\n")},

		// CRLF variations
		{"lf_only_terminator", []byte("+OK\n")},
		{"cr_only_terminator", []byte("+OK\r")},
		{"reversed_crlf", []byte("+OK\n\r")},
		// Missing CRLF
		{"no_crlf", []byte("+hello")},
		{"only_lf", []byte("+hello\n")},
		{"only_cr", []byte("+hello\r")},
		// Integer parsing
		{"int_empty", []byte(":\r\n")},
		{"int_letters", []byte(":abc\r\n")},
		// Bulk string length
		{"bulk_no_length", []byte("$\r\n")},
		{"bulk_invalid_negative", []byte("$-2\r\n")},
		{"bulk_too_large", []byte(fmt.Sprintf("$%d\r\n", MAX_BULK_STRING+1))},
		{"bulk_short_data", []byte("$10\r\nhello\r\n")},
		// Array count
		{"array_no_count", []byte("*\r\n")},
		{"array_invalid_negative", []byte("*-2\r\n")},
		{"array_too_large", []byte(fmt.Sprintf("*%d\r\n", MAX_ARRAY_SIZE+1))},
		{"array_incomplete", []byte("*3\r\n:1\r\n:2\r\n")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewReader(tc.input))
			parser := NewRESPParser(reader)
			_, err := parser.ParseValue()
			if err == nil {
				t.Errorf("expected error for %s", tc.name)
			}
		})
	}
}

func TestValidInput(t *testing.T) {
	cases := [][]byte{
		[]byte("+OK\r\n"),
		[]byte("+\r\n"),
		[]byte("-ERR unknown\r\n"),
		[]byte(":0\r\n"),
		[]byte(":-123\r\n"),
		[]byte(":9223372036854775807\r\n"),
		[]byte("$-1\r\n"),
		[]byte("$0\r\n\r\n"),
		[]byte("$5\r\nhello\r\n"),
		[]byte("*-1\r\n"),
		[]byte("*0\r\n"),
		[]byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"),
	}

	for i, input := range cases {
		reader := bufio.NewReader(bytes.NewReader(input))
		parser := NewRESPParser(reader)
		_, err := parser.ParseValue()
		if err != nil {
			t.Errorf("case %d: unexpected error %v\ninput: %q", i, err, input)
		}
	}
}

// TestChunkedDelivery: simulates TCP fragmentation

func TestChunkedDelivery(t *testing.T) {
	// build pipeline of commands
	var pipeline bytes.Buffer
	cmdCount := 10000
	for i := 0; i < cmdCount; i++ {
		argCount := rand.Intn(5) + 1
		pipeline.WriteString(fmt.Sprintf("*%d\r\n", argCount))

		for j := 0; j < argCount; j++ {
			// random binary arg 0-50 bytes
			argLen := rand.Intn(50)
			arg := make([]byte, argLen)
			crand.Read(arg)
			pipeline.WriteString(fmt.Sprintf("$%d\r\n", argLen))
			pipeline.Write(arg)
			pipeline.WriteString("\r\n")

		}
	}
	data := pipeline.Bytes()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		pos := 0
		for pos < len(data) {
			chunkSize := rand.Intn(5) + 1
			if pos+chunkSize > len(data) {
				chunkSize = len(data) - pos
			}
			pw.Write(data[pos : pos+chunkSize])
			pos += chunkSize
		}
	}()

	reader := bufio.NewReader(pr)
	parser := NewRESPParser(reader)
	parsed := 0
	for {
		_, err := parser.ParseValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		parsed++
	}
	if parsed < cmdCount-1 {
		t.Errorf("expected %d commands, got %d", cmdCount, parsed)
	}
}

// TestBinaryPayload: Values can contain any byte including \x00, \r, \n

func TestBinaryPayload(t *testing.T) {
	// Test all 256 byte values
	allBytes := make([]byte, 256)
	for i := 0; i < 256; i++ {
		allBytes[i] = byte(i)
	}
	testData := [][]byte{
		{0x00, 0x01, 0x02},
		{'\r', '\n'},
		{0xFF, 0xFE, 0xFD},
		bytes.Repeat([]byte{0x00}, 100),
		allBytes,                       // all 256 bytes
		bytes.Repeat([]byte{'\r'}, 50), // 50 CRs
		bytes.Repeat([]byte{'\n'}, 50), // 50 LFs
		[]byte("\r\n\r\n\r\n"),         // CRLF pattern
	}
	// 1000 random binary payloads
	for i := 0; i < 1000; i++ {
		size := rand.Intn(200)
		data := make([]byte, size)
		crand.Read(data)
		testData = append(testData, data)
	}
	for i, data := range testData {
		bulk := NewBulkString(data)
		serialized := bulk.Serialize()
		reader := bufio.NewReader(bytes.NewReader(serialized))
		parser := NewRESPParser(reader)
		parsed, err := parser.ParseValue()
		if err != nil {
			t.Errorf("case %d: parse error: %v", i, err)
			continue
		}
		b, ok := parsed.(*BulkString)
		if !ok {
			t.Errorf("case %d: expected *BulkString", i)
			continue
		}
		if !bytes.Equal(data, b.Data()) {
			t.Errorf("case %d: data mismatch", i)
		}
	}
}

func TestInlineCommands(t *testing.T) {
	cases := []struct {
		input    string
		wantName string
		wantArgs int
	}{
		{"PING\r\n", "PING", 0},
		{"GET key\r\n", "GET", 1},
		{"SET key value\r\n", "SET", 2},
		{"ping\r\n", "PING", 0}, // lowercase
		{"  SET  key  value  \r\n", "SET", 2},
	}
	for i, tc := range cases {
		reader := bufio.NewReader(bytes.NewReader([]byte(tc.input)))
		parser := NewCommandParser(reader)
		cmd, err := parser.ParseCommand()
		if err != nil {
			t.Errorf("case %d: %v", i, err)
			continue
		}
		if cmd.Name() != tc.wantName {
			t.Errorf("case %d: name = %q, want %q", i, cmd.Name(), tc.wantName)
		}
		if len(cmd.Args()) != tc.wantArgs {
			t.Errorf("case %d: args = %d, want %d", i, len(cmd.Args()), tc.wantArgs)
		}
	}
}
