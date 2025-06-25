package main

import (
	"github.com/kmilanbanda/gator/internal/config"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading in main(): %w", err)
	}

	err = cfg.SetUser("kmilanbanda")
	if err != nil {
		log.Fatalf("Error Setting User in main(): %w", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("Error reading updated config in main(): %w", err)
	}

	fmt.Printf("Config db_url: %s\nConfig current_user_name: %s\n", cfg.DbUrl, cfg.CurrentUser)
}
