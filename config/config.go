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
	Phase1Duration_Timer1 int `json:"phase1_timer1_duration_minutes"`
	Phase2Duration_Timer1 int `json:"phase2_timer1_duration_minutes"`
	Phase3Duration_Timer1 int `json:"phase3_timer1_duration_minutes"`
	Phase1Temp_Timer1     int `json:"phase1_timer1_temp"`
	Phase2Temp_Timer1     int `json:"phase2_timer1_temp"`
	Phase3Temp_Timer1     int `json:"phase3_timer1_temp"`
	Phase1Duration_Timer2 int `json:"phase1_timer2_duration_minutes"`
	Phase2Duration_Timer2 int `json:"phase2_timer2_duration_minutes"`
	Phase3Duration_Timer2 int `json:"phase3_timer2_duration_minutes"`
	Phase1Temp_Timer2     int `json:"phase1_timer2_temp"`
	Phase2Temp_Timer2     int `json:"phase2_timer2_temp"`
	Phase3Temp_Timer2     int `json:"phase3_timer2_temp"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Timer: TimerConfig{
			Phase1Duration_Timer1: 4,
			Phase2Duration_Timer1: 4,
			Phase3Duration_Timer1: 2,
			Phase1Temp_Timer1:     350,
			Phase2Temp_Timer1:     375,
			Phase3Temp_Timer1:     400,
			Phase1Duration_Timer2: 4,
			Phase2Duration_Timer2: 6,
			Phase3Duration_Timer2: 5,
			Phase1Temp_Timer2:     350,
			Phase2Temp_Timer2:     375,
			Phase3Temp_Timer2:     400,
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
