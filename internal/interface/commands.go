module interface

import (
	"os"
	"fmt"
	"github.com/kmilanbanda/gator/internal/config"
)

type state struct {
	cfgPtr	*Config
}

type command struct {
	name	string
	args	[]string
}

type commands struct {
	commandMap	map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	err := c.commandMap[cmd.name](s, cmd)
	if err != nil {
		fmt.Errorf("Error occured when running command: %w", err)
	}
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f	
}

func handlerLogin(s *state, cmd command) error {
	if len(args) != 1 {
		return fmt.Errorf("Error: the login command expects a single argument.")
	}

	err := s.cfgPtr.SetUser(args[0])
	if err != nil {
		return fmt.Errorf("Error during handler login: %w", err)
	}
	fmt.Printf("User has been set to %s", args[0])

	return nil
}
