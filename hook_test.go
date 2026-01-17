package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookRunner_Run(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		config      *HookConfig
		wantErr     bool
		errContains string
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "empty commands",
			config: &HookConfig{
				Commands: []string{},
			},
		},
		{
			name: "successful command",
			config: &HookConfig{
				Commands: []string{"echo hello"},
			},
		},
		{
			name: "multiple commands",
			config: &HookConfig{
				Commands: []string{"echo first", "echo second"},
			},
		},
		{
			name: "failed command with warn (default)",
			config: &HookConfig{
				Commands: []string{"exit 1", "echo after"},
			},
		},
		{
			name: "failed command with continue",
			config: &HookConfig{
				Commands: []string{"exit 1", "echo after"},
				OnError:  "continue",
			},
		},
		{
			name: "failed command with abort",
			config: &HookConfig{
				Commands: []string{"exit 1", "echo after"},
				OnError:  "abort",
			},
			wantErr:     true,
			errContains: "hook command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := NewHookRunner(tt.config)
			err := runner.Run(tmpDir)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestHookRunner_Run_WorkingDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "marker.txt")

	config := &HookConfig{
		Commands: []string{"echo 'created' > marker.txt"},
	}

	runner := NewHookRunner(config)
	if err := runner.Run(tmpDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("hook did not create marker file in working directory")
	}
}

func TestLoadProjectConfig(t *testing.T) {
	// Not parallel because we use os.Chdir

	tests := []struct {
		name       string
		configTOML string
		wantNil    bool
		wantErr    bool
		validate   func(*testing.T, *ProjectConfig)
	}{
		{
			name:    "no config file",
			wantNil: true,
		},
		{
			name: "valid config with post-add hook",
			configTOML: `[hooks.post-add]
commands = ["npm install", "echo done"]
on_error = "abort"
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				if cfg.Hooks.PostAdd == nil {
					t.Fatal("PostAdd hook is nil")
				}
				if len(cfg.Hooks.PostAdd.Commands) != 2 {
					t.Errorf("expected 2 commands, got %d", len(cfg.Hooks.PostAdd.Commands))
				}
				if cfg.Hooks.PostAdd.OnError != "abort" {
					t.Errorf("expected on_error 'abort', got %q", cfg.Hooks.PostAdd.OnError)
				}
			},
		},
		{
			name: "empty hooks config",
			configTOML: `[hooks]
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				if cfg.Hooks.PostAdd != nil {
					t.Error("expected PostAdd to be nil")
				}
			},
		},
		{
			name:       "invalid TOML",
			configTOML: "invalid { toml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Not parallel because we use os.Chdir

			// Create a temporary git repository using git init
			tmpDir := t.TempDir()

			// Initialize git repo with git init
			cmd := exec.Command("git", "init")
			cmd.Dir = tmpDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("failed to initialize git repo: %v: %s", err, out)
			}

			// Create config file if test requires it
			if tt.configTOML != "" {
				wtmDir := filepath.Join(tmpDir, ".git", "wtm")
				if err := os.MkdirAll(wtmDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(wtmDir, "config.toml"), []byte(tt.configTOML), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			// Change to the temp directory for this test
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Chdir(oldDir); err != nil {
					t.Logf("failed to restore directory: %v", err)
				}
			})

			cfg, err := loadProjectConfig()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if cfg != nil {
					t.Error("expected nil config")
				}
				return
			}

			if cfg == nil {
				t.Fatal("expected non-nil config")
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}
