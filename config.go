package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	WorktreeRoot string `toml:"worktreeRoot"`
}

var (
	configOnce   sync.Once
	cachedConfig Config
	configErr    error
)

const (
	defaultWorktreeRoot = ".git/wtm/worktrees"
	configFileEnv       = "WTM_CONFIG_FILE"
)

func loadConfig() (Config, error) {
	configOnce.Do(func() {
		path, err := configFilePath()
		if err != nil {
			configErr = err
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return
			}
			configErr = err
			return
		}
		if err := toml.Unmarshal(data, &cachedConfig); err != nil {
			configErr = err
		}
	})
	return cachedConfig, configErr
}

func configFilePath() (string, error) {
	if override := strings.TrimSpace(os.Getenv(configFileEnv)); override != "" {
		return filepath.Clean(override), nil
	}

	cfgDir := os.Getenv("XDG_CONFIG_HOME")
	if cfgDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cfgDir = filepath.Join(home, ".config")
	}
	return filepath.Clean(filepath.Join(cfgDir, "wtm", "config.toml")), nil
}

func resetConfigCache() {
	configOnce = sync.Once{}
	cachedConfig = Config{}
	configErr = nil
}
