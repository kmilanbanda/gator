package main

import _ "github.com/lib/pq"

import (
	"github.com/kmilanbanda/gator/internal/config"
	"github.com/kmilanbanda/gator/internal/database"
	"github.com/google/uuid"
	"fmt"
	"log"
	"os"
	"database/sql"
	"context"
	"time"
)

type state struct {
	db	*database.Queries
	cfgPtr	*config.Config
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
		return fmt.Errorf("Error occured when running command: %w", err)
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f	
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Error: the login command expects a single argument.")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == sql.ErrNoRows {
		return fmt.Errorf("User does not exist.")
	} else if err != nil {
		return fmt.Errorf("User not found.")
	}

	err = s.cfgPtr.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error during handler login: %w", err)
	}
	fmt.Printf("User has been set to %s\n", cmd.args[0])

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Error: the register command expects a single argument.")
	}
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		return fmt.Errorf("Error: user already exists")
	} else if err == sql.ErrNoRows {
		
	} else {
		return fmt.Errorf("Error: %w", err)
	}

	params := database.CreateUserParams{
		ID:		uuid.New(),
		CreatedAt:	time.Now(),
		UpdatedAt:	time.Now(),
		Name:		cmd.args[0],
	}

	user, err = s.db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("Error during user creation: %w", err)
	}

	err = s.cfgPtr.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error setting user: %w", err)
	}
	fmt.Printf("Created user: %+v\n", user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("Error: the reset command expects no arguments.")
	}

	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error resetting users: %w", err)
	}

	return nil
}

func handlerUsers(s *state, cmd command) error {
	
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error resetting users: %w", err)
	}

	for _, user := range users {
		fmt.Printf(" * %v", user)
		if s.cfgPtr.CurrentUser == user {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}

	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading in main(): %v", err)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	dbQueries := database.New(db)

	var s state
	s.db = dbQueries
	s.cfgPtr = &cfg
	
	var cmds commands
	cmds.commandMap = make(map[string]func(*state, command) error)

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)

	if len(os.Args) < 2 {
		log.Fatalf("Error: no arguments")
	}
	cmd := command{
		name: 	os.Args[1],
		args: 	os.Args[2:],
	}

	err = cmds.run(&s, cmd)
	if err != nil {
		log.Fatalf("Error running command:\n--- %v", err)
	}


	fmt.Printf("Config db_url: %s\nConfig current_user_name: %s\n", cfg.DbUrl, cfg.CurrentUser)
}
