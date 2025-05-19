package protocol

import (
	"flex-db/internal/resp"
	"fmt"
)

// registerHashCommands registers all hash-related commands in the command registry.
func (r *CommandRegistry) registerHashCommands() {
	r.Register("HSET", hsetCommand)
	r.Register("HGET", hgetCommand)
	r.Register("HDEL", hdelCommand)
	r.Register("HGETALL", hgetallCommand)
	r.Register("HEXISTS", hexistsCommand)
	r.Register("HLEN", hlenCommand)
	r.Register("HKEYS", hkeysCommand)
	r.Register("HVALS", hvalsCommand)
}

// hsetCommand handles the HSET command.
// Syntax: HSET key field value
// Sets the field in the hash stored at key to value.
// Returns 1 if the field is new, 0 if it was updated.
func hsetCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 3 {
		return resp.NewError("ERR wrong number of arguments for 'hset' command")
	}

	key := args[0].Str
	field := args[1].Str
	value := args[2].Str

	created, err := h.DB.HSet(key, field, value)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}

	return resp.NewInteger(int64(created))
}

// hgetCommand handles the HGET command.
// Syntax: HGET key field
// Gets the value of a field in a hash.
// Returns nil if the key or field doesn't exist.
func hgetCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.NewError("ERR wrong number of arguments for 'hget' command")
	}

	key := args[0].Str
	field := args[1].Str

	value, err := h.DB.HGet(key, field)
	if err != nil {
		return resp.NewNullBulkString()
	}

	return resp.NewBulkString(value)
}

// hdelCommand handles the HDEL command.
// Syntax: HDEL key field [field ...]
// Removes fields from a hash.
// Returns the number of fields that were removed.
func hdelCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.NewError("ERR wrong number of arguments for 'hdel' command")
	}

	key := args[0].Str
	fields := make([]string, len(args)-1)
	for i := 1; i < len(args); i++ {
		fields[i-1] = args[i].Str
	}

	deleted, err := h.DB.HDel(key, fields...)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}

	return resp.NewInteger(int64(deleted))
}

// hgetallCommand handles the HGETALL command.
// Syntax: HGETALL key
// Returns all fields and values in a hash.
// Returns an empty array if the key doesn't exist.
func hgetallCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewError("ERR wrong number of arguments for 'hgetall' command")
	}

	key := args[0].Str
	hashMap, err := h.DB.HGetAll(key)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}

	result := resp.Value{
		Type:  resp.Array,
		Array: make([]resp.Value, 0, len(hashMap)*2),
	}

	for field, value := range hashMap {
		result.Array = append(result.Array, resp.NewBulkString(field))
		result.Array = append(result.Array, resp.NewBulkString(value))
	}

	return result
}

// hexistsCommand handles the HEXISTS command.
// Syntax: HEXISTS key field
// Checks if a field exists in a hash.
// Returns 1 if the field exists, 0 otherwise.
func hexistsCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.NewError("ERR wrong number of arguments for 'hexists' command")
	}

	key := args[0].Str
	field := args[1].Str

	exists, err := h.DB.HExists(key, field)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}

	if exists {
		return resp.NewInteger(1)
	}
	return resp.NewInteger(0)
}

// hlenCommand handles the HLEN command.
// Syntax: HLEN key
// Returns the number of fields in a hash.
// Returns 0 if the key doesn't exist.
func hlenCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewError("ERR wrong number of arguments for 'hlen' command")
	}

	key := args[0].Str
	length, err := h.DB.HLen(key)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}

	return resp.NewInteger(int64(length))
}

// hkeysCommand handles the HKEYS command.
// Syntax: HKEYS key
// Returns all fields in a hash.
// Returns an empty array if the key doesn't exist.
func hkeysCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewError("ERR wrong number of arguments for 'hkeys' command")
	}

	key := args[0].Str
	keys, err := h.DB.HKeys(key)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}

	result := resp.Value{
		Type:  resp.Array,
		Array: make([]resp.Value, len(keys)),
	}

	for i, key := range keys {
		result.Array[i] = resp.NewBulkString(key)
	}

	return result
}

// hvalsCommand handles the HVALS command.
// Syntax: HVALS key
// Returns all values in a hash.
// Returns an empty array if the key doesn't exist.
func hvalsCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewError("ERR wrong number of arguments for 'hvals' command")
	}

	key := args[0].Str
	values, err := h.DB.HVals(key)
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err))
	}

	result := resp.Value{
		Type:  resp.Array,
		Array: make([]resp.Value, len(values)),
	}

	for i, value := range values {
		result.Array[i] = resp.NewBulkString(value)
	}

	return result
}
