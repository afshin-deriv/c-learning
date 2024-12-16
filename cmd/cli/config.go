package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Config struct {
	UserID     string `json:"user_id"`
	LastLesson int32  `json:"last_lesson"`
	WorkingDir string `json:"working_dir"`
	CurrentDir string `json:"current_dir"` // Added to track current lesson directory
}

// homeDir returns the user's home directory or current directory as fallback
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get home directory: %v, using current directory", err)
		homeDir = "."
	}
	return filepath.Join(homeDir, ".c-learning", "config.json")
}

// generateUserID creates a unique user ID
func generateUserID() string {
	return uuid.New().String()
}

// loadConfig loads or creates the configuration file
func loadConfig() Config {
	configPath := getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Create default config if file doesn't exist
		config := Config{
			UserID:     generateUserID(),
			LastLesson: 1,
			WorkingDir: filepath.Join(homeDir(), "c-learning"),
			CurrentDir: "",
		}

		// Save the new config
		if err := saveConfig(config); err != nil {
			log.Fatalf("Failed to save initial config: %v", err)
		}

		return config
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	return config
}

// saveConfig saves the configuration to file
func saveConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	configPath := getConfigPath()
	return os.WriteFile(configPath, data, 0644)
}
