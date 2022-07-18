package module

import "context"

// Module a plugin that can be initialized
type Module interface {
	Init(context.Context) error
}

// Command represents an executable a command
type Command interface {
	Version()
	PackageName()
	Description()
	Init(context.Context, []string) (context.Context, error)
}

// Commands a plugin that contains one or more command
type Commands interface {
	Module
	Registry() map[string]Command
}
