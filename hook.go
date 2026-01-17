package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// HooksConfig contains all hook configurations
type HooksConfig struct {
	PostAdd *HookConfig `toml:"post-add"`
}

// HookConfig defines configuration for a single hook
type HookConfig struct {
	Commands []string `toml:"commands"`
	OnError  string   `toml:"on_error"` // "abort" | "continue" | "warn" (default)
}

// ProjectConfig represents project-specific configuration stored in .git/wtm/config.toml
type ProjectConfig struct {
	Hooks HooksConfig `toml:"hooks"`
}

// HookRunner executes hook commands
type HookRunner struct {
	config *HookConfig
}

// NewHookRunner creates a new HookRunner with the given configuration
func NewHookRunner(config *HookConfig) *HookRunner {
	return &HookRunner{config: config}
}

// Run executes all hook commands in the specified working directory
func (h *HookRunner) Run(worktreePath string) error {
	if h.config == nil || len(h.config.Commands) == 0 {
		return nil
	}

	onError := strings.ToLower(strings.TrimSpace(h.config.OnError))
	if onError == "" {
		onError = "warn"
	}

	for _, cmdStr := range h.config.Commands {
		fmt.Printf("  Running: %s\n", cmdStr)

		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = worktreePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			switch onError {
			case "abort":
				return fmt.Errorf("hook command failed: %s: %w", cmdStr, err)
			case "continue":
				// Silently continue
			default: // "warn"
				fmt.Fprintf(os.Stderr, "  Warning: hook command failed: %s: %v\n", cmdStr, err)
			}
		}
	}

	return nil
}

// loadProjectConfig loads project-specific configuration from .git/wtm/config.toml
func loadProjectConfig() (*ProjectConfig, error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(repoRoot, ".git", "wtm", "config.toml")
	data, err := os.ReadFile(configPath) //nolint:gosec // Config file path is derived from git repository root
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var config ProjectConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	return &config, nil
}

// runPostAddHook executes post-add hooks if configured
func runPostAddHook(worktreePath string) error {
	config, err := loadProjectConfig()
	if err != nil {
		return err
	}

	if config == nil || config.Hooks.PostAdd == nil {
		return nil
	}

	fmt.Println("  Running post-add hooks...")
	runner := NewHookRunner(config.Hooks.PostAdd)
	return runner.Run(worktreePath)
}
