package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	Timer TimerConfig `json:"timer"`
}

// TimerConfig holds timer-specific configuration
type TimerConfig struct {
	Phase1Duration int `json:"phase1_duration_minutes"` // in minutes
	Phase2Duration int `json:"phase2_duration_minutes"` // in minutes
	Phase3Duration int `json:"phase3_duration_minutes"` // in minutes
	Phase1Temp     int `json:"phase1_temp"`             // e.g., "225"
	Phase2Temp     int `json:"phase2_temp"`             // e.g., "200"
	Phase3Temp     int `json:"phase3_temp"`             // e.g., "175"
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Timer: TimerConfig{
			Phase1Duration: 4,
			Phase2Duration: 4,
			Phase3Duration: 2,
			Phase1Temp:     350,
			Phase2Temp:     375,
			Phase3Temp:     400,
		},
	}
}

// GetConfigPath returns the path to the config directory
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "ChillClock"), nil
}

// EnsureConfigExists creates the config directory and file if they don't exist
func EnsureConfigExists() error {
	configDir, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.json")

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config
		defaultCfg := DefaultConfig()
		return SaveConfig(defaultCfg)
	}

	return nil
}

// LoadConfig loads the configuration from disk
func LoadConfig() (Config, error) {
	configDir, err := GetConfigPath()
	if err != nil {
		return Config{}, err
	}

	configFile := filepath.Join(configDir, "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(cfg Config) error {
	configDir, err := GetConfigPath()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
