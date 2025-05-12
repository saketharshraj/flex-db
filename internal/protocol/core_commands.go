package protocol

import "flex-db/internal/resp"

// adds all the core commands to the registry
func (r *CommandRegistry) registerCoreCommands() {
	r.Register("PING", pingCommand)
}

func pingCommand(h *Handler, args []resp.Value) resp.Value {
	if len(args) == 0 {
		return resp.NewSimpleString("PONG")
	}

	return args[0]
}
