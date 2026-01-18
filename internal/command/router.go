package command

import (
	"fmt"
	"strings"
	"sync"

	verr "github.com/ElshadHu/verdis/internal/errors"
	"github.com/ElshadHu/verdis/internal/protocol"
)

// Router dispatches commands to handlers
type Router struct {
	mu       sync.RWMutex            // protects commands map
	commands map[string]*CommandSpec // command name -> spec
	ctx      *Context                // shared context
}

func NewRouter() *Router {
	return &Router{
		commands: make(map[string]*CommandSpec),
		ctx:      &Context{},
	}
}

// SetContext injects dependencies (MVCC engine) after initialization
func (r *Router) SetContext(ctx *Context) {
	r.mu.Lock()
	r.ctx = ctx
	r.mu.Unlock()
}

// Register adds a command spec (panics on duplicates)
func (r *Router) Register(spec *CommandSpec) {
	if spec == nil || spec.Name == "" || spec.Handler == nil {
		panic("command: invalid spec")
	}

	name := strings.ToUpper(spec.Name)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[name]; exists {
		panic(fmt.Sprintf("command: duplicate registered - %s", name))
	}
	r.commands[name] = spec
}

// Execute routes a command to its handlers and returns the response
func (r *Router) Execute(cmd *protocol.Command) protocol.RESPValue {
	if cmd == nil {
		return protocol.NewError(verr.ErrEmptyCommand().Error())
	}

	name := strings.ToUpper(cmd.Name())

	r.mu.RLock()
	spec, exists := r.commands[name]
	ctx := r.ctx
	r.mu.RUnlock()

	if !exists {
		return protocol.NewError(verr.ErrUnknownCommand(cmd.Name()).Error())
	}
	if err := spec.Validate(cmd); err != nil {
		return protocol.NewError(err.Error())
	}
	return spec.Handler.Execute(ctx, cmd)
}
