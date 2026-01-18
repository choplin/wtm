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

	tests := []struct {
		name        string
		operations  []*HookOperation
		wantErr     bool
		errContains string
	}{
		{
			name:       "nil operations",
			operations: nil,
		},
		{
			name:       "empty operations",
			operations: []*HookOperation{},
		},
		{
			name: "successful shell command",
			operations: []*HookOperation{
				{Type: "shell", Run: "echo hello"},
			},
		},
		{
			name: "multiple shell commands",
			operations: []*HookOperation{
				{Type: "shell", Run: "echo first"},
				{Type: "shell", Run: "echo second"},
			},
		},
		{
			name: "failed shell with warn (default)",
			operations: []*HookOperation{
				{Type: "shell", Run: "exit 1"},
				{Type: "shell", Run: "echo after"},
			},
		},
		{
			name: "failed shell with continue",
			operations: []*HookOperation{
				{Type: "shell", Run: "exit 1", OnError: "continue"},
				{Type: "shell", Run: "echo after"},
			},
		},
		{
			name: "failed shell with abort",
			operations: []*HookOperation{
				{Type: "shell", Run: "exit 1", OnError: "abort"},
				{Type: "shell", Run: "echo after"},
			},
			wantErr:     true,
			errContains: "hook shell failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			runner := NewHookRunner(tt.operations, tmpDir)
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

	operations := []*HookOperation{
		{Type: "shell", Run: "echo 'created' > marker.txt"},
	}

	runner := NewHookRunner(operations, tmpDir)
	if err := runner.Run(tmpDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("hook did not create marker file in working directory")
	}
}

func TestHookRunner_Copy(t *testing.T) {
	t.Parallel()

	// Create repo root with source file
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, ".env"), []byte("SECRET=xxx"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create nested source file
	if err := os.MkdirAll(filepath.Join(repoRoot, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "config", "local.json"), []byte(`{"key":"value"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create worktree directory
	worktree := t.TempDir()

	operations := []*HookOperation{
		{Type: "copy", Paths: []string{".env", "config/local.json"}},
	}

	runner := NewHookRunner(operations, repoRoot)
	if err := runner.Run(worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check .env was copied
	content, err := os.ReadFile(filepath.Join(worktree, ".env"))
	if err != nil {
		t.Fatalf("failed to read copied .env: %v", err)
	}
	if string(content) != "SECRET=xxx" {
		t.Errorf("expected content 'SECRET=xxx', got %q", string(content))
	}

	// Check config/local.json was copied
	content, err = os.ReadFile(filepath.Join(worktree, "config", "local.json"))
	if err != nil {
		t.Fatalf("failed to read copied config/local.json: %v", err)
	}
	if string(content) != `{"key":"value"}` {
		t.Errorf("expected content '{\"key\":\"value\"}', got %q", string(content))
	}
}

func TestHookRunner_Link(t *testing.T) {
	t.Parallel()

	// Create repo root with source file
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "shared.txt"), []byte("shared content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create nested source directory
	if err := os.MkdirAll(filepath.Join(repoRoot, "node_modules", "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "node_modules", "pkg", "index.js"), []byte("module.exports = {}"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create worktree directory
	worktree := t.TempDir()

	operations := []*HookOperation{
		{Type: "link", Paths: []string{"shared.txt", "node_modules"}},
	}

	runner := NewHookRunner(operations, repoRoot)
	if err := runner.Run(worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check shared.txt is a symlink
	linkPath := filepath.Join(worktree, "shared.txt")
	fi, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("failed to stat linked file: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("expected shared.txt to be a symlink")
	}

	// Check symlink target
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}
	expectedTarget := filepath.Join(repoRoot, "shared.txt")
	if target != expectedTarget {
		t.Errorf("expected symlink target %q, got %q", expectedTarget, target)
	}

	// Check node_modules is a symlink
	modulesLink := filepath.Join(worktree, "node_modules")
	fi, err = os.Lstat(modulesLink)
	if err != nil {
		t.Fatalf("failed to stat linked directory: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("expected node_modules to be a symlink")
	}

	// Verify we can read through the symlink
	content, err := os.ReadFile(filepath.Join(worktree, "node_modules", "pkg", "index.js"))
	if err != nil {
		t.Fatalf("failed to read through symlink: %v", err)
	}
	if string(content) != "module.exports = {}" {
		t.Errorf("unexpected content through symlink: %q", string(content))
	}
}

func TestHookRunner_CopyMissingFile(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	worktree := t.TempDir()

	operations := []*HookOperation{
		{Type: "copy", Paths: []string{"nonexistent.txt"}, OnError: "abort"},
	}

	runner := NewHookRunner(operations, repoRoot)
	err := runner.Run(worktree)
	if err == nil {
		t.Error("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "source not found") {
		t.Errorf("expected 'source not found' error, got: %v", err)
	}
}

func TestHookRunner_LinkMissingFile(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	worktree := t.TempDir()

	operations := []*HookOperation{
		{Type: "link", Paths: []string{"nonexistent.txt"}, OnError: "abort"},
	}

	runner := NewHookRunner(operations, repoRoot)
	err := runner.Run(worktree)
	if err == nil {
		t.Error("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "source not found") {
		t.Errorf("expected 'source not found' error, got: %v", err)
	}
}

func TestHookRunner_OperationOrder(t *testing.T) {
	t.Parallel()

	// Create repo root with source files
	repoRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoRoot, "copy.txt"), []byte("copied"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "link.txt"), []byte("linked"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create worktree directory
	worktree := t.TempDir()

	// Operations in explicit order
	operations := []*HookOperation{
		{Type: "copy", Paths: []string{"copy.txt"}},
		{Type: "link", Paths: []string{"link.txt"}},
		{Type: "shell", Run: "echo 'command ran' > command.txt"},
	}

	runner := NewHookRunner(operations, repoRoot)
	if err := runner.Run(worktree); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all operations completed
	if _, err := os.Stat(filepath.Join(worktree, "copy.txt")); os.IsNotExist(err) {
		t.Error("copy.txt not created")
	}
	if _, err := os.Stat(filepath.Join(worktree, "link.txt")); os.IsNotExist(err) {
		t.Error("link.txt not created")
	}
	if _, err := os.Stat(filepath.Join(worktree, "command.txt")); os.IsNotExist(err) {
		t.Error("command.txt not created")
	}
}

func TestEditConfig_NoEditor(t *testing.T) {
	t.Parallel()

	// Save and clear EDITOR
	oldEditor := os.Getenv("EDITOR")
	os.Unsetenv("EDITOR")
	t.Cleanup(func() {
		if oldEditor != "" {
			os.Setenv("EDITOR", oldEditor)
		}
	})

	err := EditConfig()
	if err == nil {
		t.Error("expected error when EDITOR is not set")
	}
	if !strings.Contains(err.Error(), "EDITOR environment variable is not set") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestProjectConfigPath(t *testing.T) {
	// Not parallel because we use os.Chdir

	// Create a temporary git repository
	tmpDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to initialize git repo: %v: %s", err, out)
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

	path, err := projectConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Resolve symlinks for comparison (macOS /tmp -> /private/tmp)
	resolvedTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to resolve tmpDir symlinks: %v", err)
	}

	expected := filepath.Join(resolvedTmpDir, ".git", "wtm", "config.toml")
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
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
			name: "valid config with post-add hooks",
			configTOML: `[[hooks.post-add]]
type = "shell"
run = "npm install"
on_error = "abort"

[[hooks.post-add]]
type = "shell"
run = "echo done"
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				if len(cfg.Hooks.PostAdd) != 2 {
					t.Fatalf("expected 2 operations, got %d", len(cfg.Hooks.PostAdd))
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
			name: "empty hooks config",
			configTOML: `[hooks]
`,
			validate: func(t *testing.T, cfg *ProjectConfig) {
				t.Helper()
				if len(cfg.Hooks.PostAdd) != 0 {
					t.Errorf("expected empty PostAdd, got %d operations", len(cfg.Hooks.PostAdd))
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
