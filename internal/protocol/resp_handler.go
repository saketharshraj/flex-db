package protocol

import (
	"bufio"
	"flex-db/internal/db"
	"flex-db/internal/resp"
	"fmt"
	"net"
	"strings"
)

type RESPHandler struct {
	DB *db.FlexDB
}

func (h *Handler) HandleRESPConnection(conn net.Conn, reader *bufio.Reader) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	fmt.Printf("[+] RESP client connected: %s\n", addr)
	defer fmt.Printf("[-] RESP client disconnted: %s\n", addr)

	writer := bufio.NewWriter(conn)

	for {
		// parse the RESP command
		value, err := resp.Parse(reader)
		if err != nil {
			return
		}

		// command should be a arry of bulk strings
		if value.Type != resp.Array || value.Null {
			writeRESPError(writer, "ERR invalid command format")
			continue
		}

		if len(value.Array) == 0 {
			writeRESPError(writer, "ERR empty command")
		}

		if value.Array[0].Type != resp.BulkString {
			writeRESPError(writer, "ERR command must be a bulk string")
		}

		cmd := value.Array[0].Str
		args := value.Array[1:]

		result := h.executeCommand(cmd, args)
		writer.Write(resp.Marshal(result))
		writer.Flush()
	}
}

// command executor and returns a RESP value
func (h *Handler) executeCommand(cmd string, args []resp.Value) resp.Value {
	cmd = strings.ToUpper(cmd)

	handler, exists := h.registry.Get(cmd)
	if !exists {
		return resp.NewError(fmt.Sprintf("ERR unknown command '%s'", cmd))
	}
	return handler(h, args)

}

func writeRESPError(writer *bufio.Writer, msg string) {
	writer.Write(resp.Marshal(resp.NewError(msg)))
	writer.Flush()
}
