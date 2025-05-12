package protocol

import (
	"bufio"
	"flex-db/internal/db"
	"flex-db/internal/resp"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
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
	
	switch cmd {
	case "PING":
		if len(args) == 0 {
			return resp.NewSimpleString("PONG")
		}
		return args[0]
	
	case "SET":
		if len(args) < 2 {
			return resp.NewError("ERR wrong number of arguments")
		}

		key := args[0].Str
		value := args[1].Str

		var expiry *time.Time

		// Now check for any expiry time argument
		i := 2
		for i < len(args) {
			if i+1 >= len(args) {
				break
			}

			option := args[i].Str
			if option == "EX" {
				seconds, err := strconv.ParseInt(args[i+1].Str, 10, 64)
				if err != nil {
					return resp.NewError("ERR invalid expire time in 'set' command")
				}
				t := time.Now().Add(time.Duration(seconds) * time.Second)
				expiry = &t
				i += 2
			} else if option == "PX" {
				millis, err := strconv.ParseInt(args[i+1].Str, 10, 64)
				if err != nil {
					return resp.NewError("ERR invlaid expire time in 'set' command")
				}
				t := time.Now().Add(time.Duration(millis) * time.Millisecond)
				expiry = &t
				i += 2
			} else {
				break
			}
		}

		h.DB.Set(key, value, expiry)
		return resp.NewSimpleString("OK")
	
	case "GET":
		if len(args) != 1 {
			return resp.NewError("ERR wrong number of arguments for 'get' command")
		}

		key := args[0].Str
		value, err := h.DB.Get(key)
		if err != nil {
			return resp.NewNullBulkString()
		}

		return resp.NewBulkString(fmt.Sprintf("%v", value))
	
	case "DEL":
		if len(args) < 1 {
			return resp.NewError("ERR wrong number of arguments for 'del' command")
		}

		for _, arg := range args {
			h.DB.Delete(arg.Str)
		}

		return resp.NewSimpleString("OK")
	
	case "EXPIRE":
		if len(args)!= 2 {
			return resp.NewError("ERR wrong number of a rguments for 'expire' command")
		}

		key := args[0].Str
		duration, err := strconv.ParseInt(args[1].Str, 10, 64)
		if err != nil {
			return resp.NewError("ERR invalid duration format")
		}

		h.DB.Expire(key, time.Duration(duration)*time.Second)
		return resp.NewSimpleString("OK")

	case "TTL":
		if len(args) != 1 {
			return resp.NewError("ERR wrong number of arguments for 'ttl' command")
		}
		key := args[0].Str
		duration, err := h.DB.TTL(key)
		if err !=  nil {
			return resp.NewInteger(-1)
		}

		return resp.NewInteger(int64(duration.Seconds()))
	
	case "ALL":
		all := h.DB.All()
		result := resp.Value{
			Type: resp.Array,
			Array: make([]resp.Value, 0, len(all)*2),
		}

		for k, v := range all {
			formattedString := fmt.Sprintf("%s : %v", k, v)
			result.Array = append(result.Array, resp.NewBulkString(formattedString))
		}

		return result

	case "FLUSH":
		h.DB.Flush()
		return resp.NewSimpleString("OK")
	
	case "BGREWRITE":
		go func() {
			if err := h.DB.RewriteAOF(); err != nil {
				fmt.Printf("Error rewriting AOF: %v\n", err)
			}
		}()
		return resp.NewSimpleString("Background Rewrite started")
	
	case "HELP":
		helpArray := resp.Value{
			Type: resp.Array,
			Array: make([]resp.Value, len(AVAILABLE_COMMANDS)),
		}

		for i, helperText := range AVAILABLE_COMMANDS {
			helpArray.Array[i] = resp.NewBulkString(helperText)
		}

		return helpArray

	default:
		return resp.NewError(fmt.Sprintf("ERR unknow command %s", cmd))
	}

}


func writeRESPError(writer *bufio.Writer, msg string) {
	writer.Write(resp.Marshal(resp.NewError(msg)))
	writer.Flush()
}