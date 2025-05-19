package protocol

import (
	"flex-db/internal/resp"
	"fmt"
	"strconv"
)

// registerListCommands registers all list-related commands in the command registry.
// This includes LPUSH, RPUSH, LPOP, RPOP, LRANGE, LLEN, LINDEX, LSET, LREM, and LTRIM.
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

// lpushCommand handles the LPUSH command.
// Syntax: LPUSH key value [value ...]
// Inserts values at the beginning of a list.
// Returns the length of the list after the operation.
// Example: LPUSH mylist "world" "hello"
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

// rpushCommand handles the RPUSH command.
// Syntax: RPUSH key value [value ...]
// Appends values to the end of a list.
// Returns the length of the list after the operation.
// Example: RPUSH mylist "hello" "world"
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

// lpopCommand handles the LPOP command.
// Syntax: LPOP key
// Removes and returns the first element of a list.
// Returns nil if the key doesn't exist or the list is empty.
// Example: LPOP mylist
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

// rpopCommand handles the RPOP command.
// Syntax: RPOP key
// Removes and returns the last element of a list.
// Returns nil if the key doesn't exist or the list is empty.
// Example: RPOP mylist
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

// lrangeCommand handles the LRANGE command.
// Syntax: LRANGE key start stop
// Returns a range of elements from a list.
// Start and stop are zero-based indices. Negative indices count from the end.
// Example: LRANGE mylist 0 -1
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

// llenCommand handles the LLEN command.
// Syntax: LLEN key
// Returns the length of a list.
// Returns 0 if the key doesn't exist.
// Example: LLEN mylist
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

// lindexCommand handles the LINDEX command.
// Syntax: LINDEX key index
// Returns the element at the specified index in a list.
// Returns nil if the key doesn't exist or the index is out of range.
// Example: LINDEX mylist 0
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

// lsetCommand handles the LSET command.
// Syntax: LSET key index value
// Sets the value of an element at the specified index in a list.
// Returns an error if the key doesn't exist or the index is out of range.
// Example: LSET mylist 0 "new"
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

// lremCommand handles the LREM command.
// Syntax: LREM key count value
// Removes elements from a list based on the count and value.
// Returns the number of elements removed.
// Example: LREM mylist 1 "hello"
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

// ltrimCommand handles the LTRIM command.
// Syntax: LTRIM key start stop
// Trims a list to the specified range.
// Returns OK if successful.
// Example: LTRIM mylist 0 0
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
