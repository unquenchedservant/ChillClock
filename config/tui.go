package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TuiConfig struct {
	ServerURL    string `json:"server_url"`
	DefaultTimer int    `json:"default_timer"`
}

func DefaultTuiConfig() TuiConfig {
	return TuiConfig{ServerURL: "http://localhost:2420", DefaultTimer: 1}
}

func EnsureTuiConfigExists() error {
	configDir, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultCfg := DefaultConfig()
		return SaveConfig(defaultCfg)
	}
	return nil
}

func LoadTuiConfig() (TuiConfig, error) {
	configDir, err := GetConfigPath()
	if err != nil {
		return TuiConfig{}, err
	}

	configFile := filepath.Join(configDir, "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return TuiConfig{}, err
	}

	var cfg TuiConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return TuiConfig{}, err
	}

	return cfg, nil
}

func SaveTuiConfig(cfg TuiConfig) error {
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
