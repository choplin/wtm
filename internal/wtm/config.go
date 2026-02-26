// Package wtm implements worktree manager, a CLI tool for managing Git worktrees.
package wtm

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	WorktreeRoot string      `toml:"worktreeRoot"`
	Hooks        HooksConfig `toml:"hooks"`
}

var (
	configOnce   sync.Once
	cachedConfig Config
	configErr    error
)

const (
	defaultWorktreeRoot = ".wtm"
	configFileEnv       = "WTM_CONFIG_FILE"
)

func loadConfig() (Config, error) {
	configOnce.Do(func() {
		path, err := configFilePath()
		if err != nil {
			configErr = err
			return
		}
		data, err := os.ReadFile(path) //nolint:gosec // Config file path is from user's home directory
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

// loadEffectiveProjectConfig loads and merges global + project hook configs.
// Global hooks run first, followed by project hooks.
// Returns nil if no hooks are configured at either level.
func loadEffectiveProjectConfig() (*ProjectConfig, error) {
	globalCfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	projectCfg, err := loadProjectConfig()
	if err != nil {
		return nil, err
	}

	merged := &ProjectConfig{}

	// Global hooks first
	merged.Hooks.PostAdd = append(merged.Hooks.PostAdd, globalCfg.Hooks.PostAdd...)

	// Project hooks second
	if projectCfg != nil {
		merged.Hooks.PostAdd = append(merged.Hooks.PostAdd, projectCfg.Hooks.PostAdd...)
	}

	if len(merged.Hooks.PostAdd) == 0 {
		return nil, nil
	}

	return merged, nil
}
