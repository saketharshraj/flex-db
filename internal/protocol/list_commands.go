package protocol

import (
	"flex-db/internal/resp"
	"fmt"
	"strconv"
)

// registerListCommands adds all list commands to the registry
func (r *CommandRegistry) registerListCommands() {
	r.Register("LPUSH", lpushCommand)
	r.Register("RPUSH", rpushCommand)
	r.Register("LPOP", lpopCommand)
	r.Register("RPOP", rpopCommand)
	r.Register("LRANGE", lrangeCommand)
	r.Register("LLEN", llenCommand)
	r.Register("LINDEX", lindexCommand)
	r.Register("LSET", lsetCommand)
	r.Register("LREM", lremCommand)
	r.Register("LTRIM", ltrimCommand)
}

func lpushCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.NewError("ERR wrong number of arguments for 'lpush' command")
	}
	
	key := args[0].Str
	values := make([]string, len(args)-1)
	for i := 1; i < len(args); i++ {
		values[i-1] = args[i].Str
	}
	
	length, err := h.DB.LPush(key, values...)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}
	
	return resp.NewInteger(int64(length))
}

func rpushCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.NewError("ERR wrong number of arguments for 'rpush' command")
	}
	
	key := args[0].Str
	values := make([]string, len(args)-1)
	for i := 1; i < len(args); i++ {
		values[i-1] = args[i].Str
	}
	
	length, err := h.DB.RPush(key, values...)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}
	
	return resp.NewInteger(int64(length))
}

func lpopCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewError("ERR wrong number of arguments for 'lpop' command")
	}
	
	key := args[0].Str
	value, err := h.DB.LPop(key)
	if err != nil {
		return resp.NewNullBulkString()
	}
	
	return resp.NewBulkString(value)
}

func rpopCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewError("ERR wrong number of arguments for 'rpop' command")
	}
	
	key := args[0].Str
	value, err := h.DB.RPop(key)
	if err != nil {
		return resp.NewNullBulkString()
	}
	
	return resp.NewBulkString(value)
}

func lrangeCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.NewError("ERR wrong number of arguments for 'lrange' command")
	}
	
	key := args[0].Str
	start, err := strconv.Atoi(args[1].Str)
	if err != nil {
		return resp.NewError("ERR value is not an integer or out of range")
	}
	
	stop, err := strconv.Atoi(args[2].Str)
	if err != nil {
		return resp.NewError("ERR value is not an integer or out of range")
	}
	
	values, err := h.DB.LRange(key, start, stop)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}
	
	result := resp.Value{
		Type:  resp.Array,
		Array: make([]resp.Value, len(values)),
	}
	
	for i, val := range values {
		result.Array[i] = resp.NewBulkString(val)
	}
	
	return result
}

func llenCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewError("ERR wrong number of arguments for 'llen' command")
	}
	
	key := args[0].Str
	length, err := h.DB.LLen(key)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}
	
	return resp.NewInteger(int64(length))
}

func lindexCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.NewError("ERR wrong number of arguments for 'lindex' command")
	}
	
	key := args[0].Str
	index, err := strconv.Atoi(args[1].Str)
	if err != nil {
		return resp.NewError("ERR value is not an integer or out of range")
	}
	
	value, err := h.DB.LIndex(key, index)
	if err != nil {
		return resp.NewNullBulkString()
	}
	
	return resp.NewBulkString(value)
}

func lsetCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.NewError("ERR wrong number of arguments for 'lset' command")
	}
	
	key := args[0].Str
	index, err := strconv.Atoi(args[1].Str)
	if err != nil {
		return resp.NewError("ERR value is not an integer or out of range")
	}
	
	value := args[2].Str
	
	err = h.DB.LSet(key, index, value)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}
	
	return resp.NewSimpleString("OK")
}

func lremCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.NewError("ERR wrong number of arguments for 'lrem' command")
	}
	
	key := args[0].Str
	count, err := strconv.Atoi(args[1].Str)
	if err != nil {
		return resp.NewError("ERR value is not an integer or out of range")
	}
	
	value := args[2].Str
	
	removed, err := h.DB.LRem(key, count, value)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}
	
	return resp.NewInteger(int64(removed))
}

func ltrimCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.NewError("ERR wrong number of arguments for 'ltrim' command")
	}
	
	key := args[0].Str
	start, err := strconv.Atoi(args[1].Str)
	if err != nil {
		return resp.NewError("ERR value is not an integer or out of range")
	}
	
	stop, err := strconv.Atoi(args[2].Str)
	if err != nil {
		return resp.NewError("ERR value is not an integer or out of range")
	}
	
	err = h.DB.LTrim(key, start, stop)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}
	
	return resp.NewSimpleString("OK")
}