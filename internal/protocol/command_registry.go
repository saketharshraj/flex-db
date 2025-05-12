package protocol

import "flex-db/internal/resp"


type CommandHandler func(h *Handler, args []resp.Value) resp.Value

type CommandRegistry struct {
	commands map[string]CommandHandler
}

func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]CommandHandler),
	}

	// register all commands
	registry.registerCoreCommands()

	return registry
}

// register adds a command to the registry
func (r *CommandRegistry) Register(name string, handler CommandHandler) {
	r.commands[name] = handler
}

// returns a command handler if exitsts
func (r *CommandRegistry) Get(name string) (CommandHandler, bool) {
	handler, exists := r.commands[name]
	return handler, exists
}
