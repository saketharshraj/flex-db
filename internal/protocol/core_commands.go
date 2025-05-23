package protocol

import (
	"flex-db/internal/resp"
	"fmt"
	"strconv"
	"time"
)

// adds all the core commands to the registry
func (r *CommandRegistry) registerCoreCommands() {
	r.Register("PING", pingCommand)
	r.Register("SET", setCommand)
	r.Register("GET", getCommand)
	r.Register("DEL", deleteCommand)
	r.Register("EXPIRE", expireCommand)
	r.Register("TTL", ttlCommand)
	r.Register("ALL", allCommand)
	r.Register("FLUSH", flushCommand)
	r.Register("BGREWRITEAOF", bgrewriteCommand)
	r.Register("HELP", helpCommand)
}

func pingCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) == 0 {
		return resp.NewSimpleString("PONG")
	}

	return args[0]
}

func setCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.NewError("ERR wrong number of arguments")
	}

	key := args[0].Str
	value := args[1].Str

	var expiry *time.Time

	// now check for expiry argument
	i := 2
	for i < len(args) {
		if i+1 >= len(args) {
			break
		}

		option := args[2].Str
		if option == "EX" {
			seconds, err := strconv.ParseInt(args[i+1].Str, 10, 64)
			if err != nil {
				return resp.NewError("ERR in valid expire time in 'set' command")
			}
			t :=  time.Now().Add(time.Duration(seconds) * time.Second)
			expiry = &t
			i += 2
		} else if option == "PX" {
			millis, err := strconv.ParseInt(args[i+1].Str, 10, 64)
			if err != nil {
				return resp.NewError("ERR invalid expire time in 'set' command")
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
}


func getCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) < 1 {
		return resp.NewError("ERR key is required to access the data")
	}

	key := args[0].Str

	val, err := h.DB.Get(key)
	if err != nil {
		return resp.NewError(err.Error())
	}

	return resp.NewBulkString(fmt.Sprintf("%v", val))

}

func deleteCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) < 1 {
		return resp.NewError("ERR key is required to delete the data")
	}

	key := args[0].Str

	err := h.DB.Delete(key)
	if err != nil {
		return resp.NewError(err.Error())
	}

	return resp.NewSimpleString("OK")
}

func expireCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.NewError("ERR key and expire time is required")
	}

	key := args[0].Str
	duration, err := strconv.ParseInt(args[1].Str, 10, 64)

	if err != nil {
		return resp.NewError("ERR invalid duration format")
	}

	h.DB.Expire(key, time.Duration(duration) * time.Second)
	return resp.NewSimpleString("OK")
}

func ttlCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) < 1 {
		return resp.NewError("ERR specify key to fetch its TTL")
	}

	key := args[0].Str

	duration, err :=  h.DB.TTL(key)
	if err != nil {
		return resp.NewSimpleString(err.Error())
	}

	return resp.NewInteger(int64(duration.Seconds()))
}

func allCommand(h *Handler, args []resp.Value) resp.Value {
	all := h.DB.All()

	result := resp.Value{
		Type: resp.Array,
		Array: make([]resp.Value, 0, len(all)*2),
	}

	for k,v := range all {
		formattedString := fmt.Sprintf("%s : %v", k, v)
		result.Array = append(result.Array, resp.NewBulkString(formattedString))
	}

	return result
}

func flushCommand(h *Handler, args []resp.Value) resp.Value {
	h.DB.Flush()
	return resp.NewSimpleString("OK")
}

func bgrewriteCommand(h *Handler, args []resp.Value) resp.Value {
	go func () {
		if err := h.DB.RewriteAOF(); err != nil {
			fmt.Printf("Error rewriting AOF: %v\n", err)
		}
	}()
	return resp.NewSimpleString("Background Rewrite started")
}

func helpCommand(h *Handler, args []resp.Value) resp.Value {
	helpArray := resp.Value{
		Type: resp.Array,
		Array: make([]resp.Value, len(AVAILABLE_COMMANDS)),
	}

	for i, helperText := range AVAILABLE_COMMANDS {
		helpArray.Array[i] = resp.NewBulkString(helperText)
	}

	return helpArray
}
