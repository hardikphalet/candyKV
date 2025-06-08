package commands

import (
	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/store"
)

type SetCommand struct {
	Key     string
	Value   string
	Options *options.SetOptions
}

func NewSetCommand(key, value string, opts *options.SetOptions) *SetCommand {
	return &SetCommand{
		Key:     key,
		Value:   value,
		Options: opts,
	}
}

func (c *SetCommand) Execute(store store.Store) (interface{}, error) {
	return store.Set(c.Key, c.Value, c.Options)
}
