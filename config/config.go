package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	GeminiAPIKey  string `toml:"gemini_api_key"`
	ClaudeApproved bool   `toml:"claude_approved"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cura", "config.toml")
}

func Load() (Config, error) {
	var cfg Config
	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil // no config yet, return empty
	}
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}

func Save(cfg Config) error {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
