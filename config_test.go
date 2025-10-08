package main

import (
	"os"
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
	if rel != filepath.Clean(".git/wtm/worktrees") {
		t.Fatalf("expected relative path '.git/wtm/worktrees', got %s", rel)
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
	if err := os.WriteFile(configFile, []byte("worktreeRoot = \"custom/worktrees\"\n"), 0o644); err != nil {
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
