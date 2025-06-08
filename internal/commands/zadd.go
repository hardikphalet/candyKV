package commands

import (
	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/store"
	"github.com/hardikphalet/go-redis/internal/types"
)

type ZAddCommand struct {
	Key     string
	Members []types.ScoreMember
	Options *options.ZAddOptions
}

func (c *ZAddCommand) Execute(store store.Store) (interface{}, error) {
	return store.ZAdd(c.Key, c.Members, c.Options)
}
