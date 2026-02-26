package wtm

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveWorktreeBaseDefault(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer cleanupTestRepo(t, repoPath)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to test repo: %v", err)
	}

	t.Setenv("WTM_CONFIG_FILE", "")
	tempConfigDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfigDir)
	resetConfigCache()
	defer resetConfigCache()

	base, err := resolveWorktreeBase()
	if err != nil {
		t.Fatalf("resolveWorktreeBase failed: %v", err)
	}

	rel := relativeToRepoRoot(t, base)
	if rel != filepath.Clean(".wtm") {
		t.Fatalf("expected relative path '.wtm', got %s", rel)
	}
}

func TestResolveWorktreeBaseWithConfigFile(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer cleanupTestRepo(t, repoPath)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to test repo: %v", err)
	}

	configFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configFile, []byte("worktreeRoot = \"custom/worktrees\"\n"), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	t.Setenv("WTM_CONFIG_FILE", configFile)
	resetConfigCache()
	defer resetConfigCache()

	base, err := resolveWorktreeBase()
	if err != nil {
		t.Fatalf("resolveWorktreeBase failed: %v", err)
	}

	rel := relativeToRepoRoot(t, base)
	if rel != filepath.Clean("custom/worktrees") {
		t.Fatalf("expected relative path 'custom/worktrees', got %s", rel)
	}
}

func TestLoadEffectiveProjectConfig(t *testing.T) {
	// Not parallel because we use os.Chdir

	tests := []struct {
		name             string
		globalConfigTOML string
		projectConfigTOML string
		wantNil          bool
		validate         func(*testing.T, *ProjectConfig)
	}{
		{
			name:    "no hooks at either level",
			wantNil: true,
		},
		{
			name: "global hooks only",
			globalConfigTOML: `[[hooks.post-add]]
type = "link"
paths = ["node_modules"]
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				if len(cfg.Hooks.PostAdd) != 1 {
					t.Fatalf("expected 1 operation, got %d", len(cfg.Hooks.PostAdd))
				}
				if cfg.Hooks.PostAdd[0].Type != "link" {
					t.Errorf("expected type 'link', got %q", cfg.Hooks.PostAdd[0].Type)
				}
				// on_error is empty; runtime defaults to "warn"
				if cfg.Hooks.PostAdd[0].OnError != "" {
					t.Errorf("expected on_error '' (runtime default 'warn'), got %q", cfg.Hooks.PostAdd[0].OnError)
				}
			},
		},
		{
			name: "project hooks only",
			projectConfigTOML: `[[hooks.post-add]]
type = "shell"
run = "npm install"
on_error = "abort"
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				if len(cfg.Hooks.PostAdd) != 1 {
					t.Fatalf("expected 1 operation, got %d", len(cfg.Hooks.PostAdd))
				}
				if cfg.Hooks.PostAdd[0].Run != "npm install" {
					t.Errorf("expected run 'npm install', got %q", cfg.Hooks.PostAdd[0].Run)
				}
				if cfg.Hooks.PostAdd[0].OnError != "abort" {
					t.Errorf("expected on_error 'abort', got %q", cfg.Hooks.PostAdd[0].OnError)
				}
			},
		},
		{
			name: "both global and project hooks merged in order",
			globalConfigTOML: `[[hooks.post-add]]
type = "link"
paths = ["node_modules"]
`,
			projectConfigTOML: `[[hooks.post-add]]
type = "shell"
run = "npm install"
on_error = "abort"
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				if len(cfg.Hooks.PostAdd) != 2 {
					t.Fatalf("expected 2 operations, got %d", len(cfg.Hooks.PostAdd))
				}
				// Global hook first
				if cfg.Hooks.PostAdd[0].Type != "link" {
					t.Errorf("expected first op type 'link', got %q", cfg.Hooks.PostAdd[0].Type)
				}
				// Project hook second
				if cfg.Hooks.PostAdd[1].Type != "shell" {
					t.Errorf("expected second op type 'shell', got %q", cfg.Hooks.PostAdd[1].Type)
				}
			},
		},
		{
			name: "global on_error unset uses runtime default",
			globalConfigTOML: `[[hooks.post-add]]
type = "copy"
paths = [".env"]

[[hooks.post-add]]
type = "shell"
run = "echo hello"
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				// Both copy and shell keep on_error empty; runtime defaults to "warn"
				if cfg.Hooks.PostAdd[0].OnError != "" {
					t.Errorf("expected on_error '' for global copy, got %q", cfg.Hooks.PostAdd[0].OnError)
				}
				if cfg.Hooks.PostAdd[1].OnError != "" {
					t.Errorf("expected on_error '' for global shell, got %q", cfg.Hooks.PostAdd[1].OnError)
				}
			},
		},
		{
			name: "global explicit on_error is respected",
			globalConfigTOML: `[[hooks.post-add]]
type = "link"
paths = ["node_modules"]
on_error = "abort"
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				if cfg.Hooks.PostAdd[0].OnError != "abort" {
					t.Errorf("expected on_error 'abort' (explicitly set), got %q", cfg.Hooks.PostAdd[0].OnError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary git repository
			tmpDir := t.TempDir()
			cmd := exec.Command("git", "init")
			cmd.Dir = tmpDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("failed to initialize git repo: %v: %s", err, out)
			}

			// Set up global config
			globalConfigFile := filepath.Join(t.TempDir(), "config.toml")
			if tt.globalConfigTOML != "" {
				if err := os.WriteFile(globalConfigFile, []byte(tt.globalConfigTOML), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv(configFileEnv, globalConfigFile)

			// Set up project config
			if tt.projectConfigTOML != "" {
				wtmDir := filepath.Join(tmpDir, ".wtm")
				if err := os.MkdirAll(wtmDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(wtmDir, "config.toml"), []byte(tt.projectConfigTOML), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			// Change to the temp directory
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

			resetConfigCache()
			t.Cleanup(resetConfigCache)

			cfg, err := loadEffectiveProjectConfig()
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

func relativeToRepoRoot(t *testing.T, path string) string {
	commonDir, err := runGitCommand("rev-parse", "--git-common-dir")
	if err != nil {
		t.Fatalf("Failed to get git common dir: %v", err)
	}
	commonDir = strings.TrimSpace(commonDir)
	if !filepath.IsAbs(commonDir) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get cwd: %v", err)
		}
		commonDir = filepath.Join(cwd, commonDir)
	}
	repoRoot := filepath.Clean(filepath.Join(commonDir, ".."))

	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		t.Fatalf("Failed to compute relative path: %v", err)
	}
	return filepath.Clean(rel)
}
