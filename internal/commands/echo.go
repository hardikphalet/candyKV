package commands

import (
	"fmt"

	"github.com/hardikphalet/go-redis/internal/store"
)

type EchoCommand struct {
	message string
}

func NewEchoCommand(args []string) (*EchoCommand, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("ECHO command requires exactly 1 argument")
	}
	return &EchoCommand{message: args[1]}, nil
}

func (c *EchoCommand) Execute(store store.Store) (interface{}, error) {
	return c.message, nil
}
