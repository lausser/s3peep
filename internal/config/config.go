package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ActiveProfile string    `json:"active_profile"`
	Profiles      []Profile `json:"profiles"`
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "s3peep", "config.json")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func GetActiveProfile(cfg *Config) *Profile {
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == cfg.ActiveProfile {
			return &cfg.Profiles[i]
		}
	}
	return nil
}

func FindProfile(cfg *Config, name string) *Profile {
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == name {
			return &cfg.Profiles[i]
		}
	}
	return nil
}
