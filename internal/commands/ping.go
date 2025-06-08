package commands

import (
	"github.com/hardikphalet/go-redis/internal/store"
	"github.com/hardikphalet/go-redis/internal/types"
)

type PingCommand struct{}

func (c *PingCommand) Execute(store store.Store) (interface{}, error) {
	return types.SimpleString("PONG"), nil
}
