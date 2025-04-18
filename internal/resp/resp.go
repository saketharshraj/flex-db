package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// RESP data types
const (
	SimpleString = '+'
	Error        = '-'
	Integer      = ':'
	BulkString   = '$'
	Array        = '*'
)

type Value struct {
	Type  byte
	Str   string
	Int   int64
	Array []Value
	Null  bool
}

// Common RESP errors
var (
	ErrInvalidSyntax = errors.New("invalid RESP syntax")
	ErrNotRESP       = errors.New("not a RESP message")
)

// util func to convert a value to its RESP wire format
func Marshal(v Value) []byte {
	switch v.Type {
	case SimpleString:
		return []byte(fmt.Sprintf("+%s\r\n", v.Str))
	case Error:
		return []byte(fmt.Sprintf("+%s\r\n", v.Str))
	case Integer:
		return []byte(fmt.Sprintf("+%d\r\n", v.Int))
	case BulkString:
		if v.Null {
			return []byte("$-1\r\n")
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v.Str), v.Str))
	case Array:
		if v.Null {
			return []byte("*-1\r\n")
		}
		result := []byte(fmt.Sprintf("*%d\r\n", len(v.Array)))
		return result
	default:
		return []byte{}
	}
}

// NewSimpleString creates a new RESP simple string
func NewSimpleString(str string) Value {
	return Value{Type: SimpleString, Str: str}
}

// NewError creates a new RESP error
func NewError(str string) Value {
	return Value{Type: Error, Str: str}
}

// NewInteger creates a new RESP integer
func NewInteger(val int64) Value {
	return Value{Type: Integer, Int: val}
}

// NewBulkString creates a new RESP bulk string
func NewBulkString(str string) Value {
	return Value{Type: BulkString, Str: str}
}

// NewNullBulkString creates a new null RESP bulk string
func NewNullBulkString() Value {
	return Value{Type: BulkString, Null: true}
}

// NewArray creates a new RESP array
func NewArray(items []Value) Value {
	return Value{Type: Array, Array: items}
}

// NewNullArray creates a new null RESP array
func NewNullArray() Value {
	return Value{Type: Array, Null: true}
}


func Parse(reader *bufio.Reader) (Value, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch b {
	case SimpleString:
		return parseSimpleString(reader)
	case Error:
		return parseError(reader)
	case Integer:
		return parseInteger(reader)
	case BulkString:
		return parseBulkString(reader)
	case Array:
		return parseArray(reader)
	default:
		reader.UnreadByte()
		return parseSimpleString(reader)
	}
}

func parseSimpleString(reader *bufio.Reader) (Value, error) {
	line, err := readLine(reader)
	if err != nil {
		return Value{}, err
	}
	return Value{Type: SimpleString, Str: line}, nil
}

func parseError(reader *bufio.Reader) (Value, error) {
	line, err := readLine(reader)
	if err != nil {
		return Value{}, err
	}

	return Value{Type: Error, Str: line}, nil
}

func parseInteger(reader *bufio.Reader) (Value, error) {
	line, err := readLine(reader)
	if err != nil {
		return Value{}, err
	}

	n, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return Value{}, ErrInvalidSyntax
	}

	return Value{Type: Integer, Int: n}, nil
}

func parseBulkString(reader *bufio.Reader) (Value, error) {
	line, err := readLine(reader)
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, ErrInvalidSyntax
	}

	if length == -1 {
		return Value{Type: BulkString, Null: true}, nil
	}

	if length < 0 {
		return Value{}, ErrInvalidSyntax
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return Value{}, err
	}

	_, err = reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	_, err = reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	return Value{Type: BulkString, Str: string(buf)}, nil
}

func parseArray(reader *bufio.Reader) (Value, error) {
	line, err := readLine(reader)
	if err != nil {
		return Value{}, err
	}

	count, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, ErrInvalidSyntax
	}

	if count == -1 {
		return Value{Type: Array, Null: true}, nil
	}

	if count < 0 {
		return Value{}, ErrInvalidSyntax
	}

	items := make([]Value, 0, count)
	for i := 0; i < count; i++ {
		item, err := Parse(reader)
		if err != nil {
			return Value{}, err
		}
		items = append(items, item)
	}

	return Value{Type: Array, Array: items}, nil
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", ErrInvalidSyntax
	}

	return line[:len(line)-2], nil
}
