package config

import (
	"os"
	"fmt"
	"encoding/json"
)



type Config struct {
	DbUrl			string	`json:"db_url"`
	CurrentUser	string	`json:"current_user_name"`
}

func Read() (Config, error) {
	var cfg Config

	filepath, err := getConfigFilePath()
	if err != nil {
		return cfg, err
	}

	var bytes []byte
	bytes, err = os.ReadFile(filepath)
	if err != nil {
		return cfg, fmt.Errorf("Error reading file at %s: %w", filepath, err)
	}

	if err = json.Unmarshal(bytes, &cfg); err != nil {	
		return cfg, fmt.Errorf("Error unmarshalling JSON from %s: %w", filepath, err)
	}

	return cfg, nil
}

func (cfg Config) SetUser(username string) error {
	cfg.CurrentUser = username
	return write(cfg)
}

func getConfigFilePath() (string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return homeDirectory, fmt.Errorf("Error getting home directory: %w", err)
	}

	return homeDirectory + "/" + configFileName, nil
}

func write(cfg Config) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("Error marshalling JSON data: %w", err)
	}

	filepath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("Error writing to file: %w", err)
	}

	return nil	
}

const configFileName = ".gatorconfig.json"
