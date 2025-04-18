package protocol

import (
	"bufio"
	"io"
	"net"
)

type ProtocolType int

const (
	TextProtocol ProtocolType = iota
	RESPProtocol
)

func DetectProtocol(conn net.Conn) (ProtocolType, *bufio.Reader, error) {
	reader := bufio.NewReader(conn)

	// peek at the first byte without consuming it
	b, err := reader.Peek(1)
	if err != nil {
		if err == io.EOF {
			return TextProtocol, reader, nil
		}
		return TextProtocol, reader, err
	}

	// check for resp
	switch b[0] {
	case SimpleString, Error, Integer, BulkString, Array:
		return RESPProtocol, reader, nil
	default:
		return TextProtocol, reader, nil
	}
}
