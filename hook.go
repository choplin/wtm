package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// HookOperation defines a single hook operation
type HookOperation struct {
	Type    string   `toml:"type"`     // "copy" | "link" | "shell"
	Paths   []string `toml:"paths"`    // for copy/link
	Run     string   `toml:"run"`      // for shell
	OnError string   `toml:"on_error"` // "abort" | "continue" | "warn" (default)
}

// HooksConfig contains all hook configurations
type HooksConfig struct {
	PostAdd []*HookOperation `toml:"post-add"`
}

// ProjectConfig represents project-specific configuration stored in .wtm/config.toml
type ProjectConfig struct {
	Hooks HooksConfig `toml:"hooks"`
}

// HookRunner executes hook operations
type HookRunner struct {
	operations []*HookOperation
	repoRoot   string
}

// NewHookRunner creates a new HookRunner with the given operations
func NewHookRunner(operations []*HookOperation, repoRoot string) *HookRunner {
	return &HookRunner{operations: operations, repoRoot: repoRoot}
}

// Run executes all hook operations in the specified working directory
func (h *HookRunner) Run(worktreePath string) error {
	if len(h.operations) == 0 {
		return nil
	}

	for _, op := range h.operations {
		onError := strings.ToLower(strings.TrimSpace(op.OnError))
		if onError == "" {
			onError = "warn"
		}

		switch strings.ToLower(op.Type) {
		case "copy":
			for _, path := range op.Paths {
				if err := h.copyFile(path, worktreePath); err != nil {
					if handleErr := h.handleError("copy", path, err, onError); handleErr != nil {
						return handleErr
					}
				}
			}
		case "link":
			for _, path := range op.Paths {
				if err := h.linkFile(path, worktreePath); err != nil {
					if handleErr := h.handleError("link", path, err, onError); handleErr != nil {
						return handleErr
					}
				}
			}
		case "shell":
			if op.Run == "" {
				continue
			}
			fmt.Printf("  Running: %s\n", op.Run)

			cmd := exec.Command("sh", "-c", op.Run)
			cmd.Dir = worktreePath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				if handleErr := h.handleError("shell", op.Run, err, onError); handleErr != nil {
					return handleErr
				}
			}
		}
	}

	return nil
}

// handleError handles errors based on the on_error setting
func (h *HookRunner) handleError(operation, target string, err error, onError string) error {
	switch onError {
	case "abort":
		return fmt.Errorf("hook %s failed: %s: %w", operation, target, err)
	case "continue":
		// Silently continue
	default: // "warn"
		fmt.Fprintf(os.Stderr, "  Warning: hook %s failed: %s: %v\n", operation, target, err)
	}
	return nil
}

// copyFile copies a file from repo root to worktree
func (h *HookRunner) copyFile(relPath, worktreePath string) error {
	src := filepath.Join(h.repoRoot, relPath)
	dst := filepath.Join(worktreePath, relPath)

	// Check if source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source not found: %w", err)
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Copy file
	srcFile, err := os.Open(src) //nolint:gosec // Path is derived from config
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	fmt.Printf("  Copied: %s\n", relPath)
	return nil
}

// linkFile creates a symlink from worktree to repo root
func (h *HookRunner) linkFile(relPath, worktreePath string) error {
	src := filepath.Join(h.repoRoot, relPath)
	dst := filepath.Join(worktreePath, relPath)

	// Check if source exists
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source not found: %w", err)
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create symlink (use absolute path for source)
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.Symlink(absSrc, dst); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("  Linked: %s\n", relPath)
	return nil
}

// loadProjectConfig loads project-specific configuration from .wtm/config.toml
func loadProjectConfig() (*ProjectConfig, error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(repoRoot, ".wtm", "config.toml")
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
	repoRoot, err := getRepoRoot()
	if err != nil {
		return err
	}

	config, err := loadProjectConfig()
	if err != nil {
		return err
	}

	if config == nil || len(config.Hooks.PostAdd) == 0 {
		return nil
	}

	fmt.Println("  Running post-add hooks...")
	runner := NewHookRunner(config.Hooks.PostAdd, repoRoot)
	return runner.Run(worktreePath)
}

// projectConfigPath returns the path to the project configuration file
func projectConfigPath() (string, error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(repoRoot, ".wtm", "config.toml"), nil
}

// EditConfig opens the project configuration file in the user's editor
func EditConfig() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("EDITOR environment variable is not set")
	}

	configPath, err := projectConfigPath()
	if err != nil {
		return err
	}

	// Create config directory and empty file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if err := os.WriteFile(configPath, []byte{}, 0o644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
