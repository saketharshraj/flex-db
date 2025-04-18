package protocol

import (
	"bufio"
	"errors"
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
