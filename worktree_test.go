package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "wtm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to config git user.name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to config git user.email: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo\n"), 0o644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to add test file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	return tmpDir
}

// cleanupTestRepo removes the temporary test repository
func cleanupTestRepo(t *testing.T, repoPath string) {
	t.Helper()
	if err := os.RemoveAll(repoPath); err != nil {
		t.Errorf("Failed to cleanup test repo: %v", err)
	}
}

func TestAddWorktree(t *testing.T) {
	repoPath := setupTestRepo(t)
	defer cleanupTestRepo(t, repoPath)

	// Change to test repo directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		t.Fatalf("Failed to change to test repo: %v", err)
	}

	t.Run("add worktree with default branch name", func(t *testing.T) {
		err := AddWorktree("feature-1", "", "", "")
		if err != nil {
			t.Errorf("AddWorktree failed: %v", err)
		}

		// Verify worktree was created
		worktrees, err := getWorktrees()
		if err != nil {
			t.Errorf("getWorktrees failed: %v", err)
		}

		found := false
		for _, wt := range worktrees {
			if wt.Name == "feature-1" {
				found = true
				if wt.Branch != "feature-1" {
					t.Errorf("Expected branch 'feature-1', got '%s'", wt.Branch)
				}
			}
		}

		if !found {
			t.Error("Worktree 'feature-1' was not created")
		}
	})

	t.Run("add worktree with custom branch name", func(t *testing.T) {
		err := AddWorktree("api", "feature/api-refactoring", "", "")
		if err != nil {
			t.Errorf("AddWorktree failed: %v", err)
		}

		worktrees, err := getWorktrees()
		if err != nil {
			t.Errorf("getWorktrees failed: %v", err)
		}

		found := false
		for _, wt := range worktrees {
			if wt.Name == "api" {
				found = true
				if wt.Branch != "feature/api-refactoring" {
					t.Errorf("Expected branch 'feature/api-refactoring', got '%s'", wt.Branch)
				}
			}
		}

		if !found {
			t.Error("Worktree 'api' was not created")
		}
	})

	t.Run("add duplicate worktree should fail", func(t *testing.T) {
		err := AddWorktree("feature-1", "", "", "")
		if err == nil {
			t.Error("Expected error when adding duplicate worktree, got nil")
		}
	})
}

func TestListWorktrees(t *testing.T) {
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

	// Create test worktrees
	AddWorktree("test-1", "", "", "")
	AddWorktree("test-2", "", "", "")

	t.Run("list in table format", func(t *testing.T) {
		err := ListWorktrees("table")
		if err != nil {
			t.Errorf("ListWorktrees failed: %v", err)
		}
	})

	t.Run("list in plain format", func(t *testing.T) {
		err := ListWorktrees("plain")
		if err != nil {
			t.Errorf("ListWorktrees failed: %v", err)
		}
	})

	t.Run("list in json format", func(t *testing.T) {
		err := ListWorktrees("json")
		if err != nil {
			t.Errorf("ListWorktrees failed: %v", err)
		}
	})

	t.Run("unknown format should fail", func(t *testing.T) {
		err := ListWorktrees("unknown")
		if err == nil {
			t.Error("Expected error for unknown format, got nil")
		}
	})
}

func TestShowWorktree(t *testing.T) {
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

	// Create test worktree
	AddWorktree("show-test", "", "", "")

	t.Run("show in pretty format", func(t *testing.T) {
		err := ShowWorktree("show-test", "pretty", "")
		if err != nil {
			t.Errorf("ShowWorktree failed: %v", err)
		}
	})

	t.Run("show in json format", func(t *testing.T) {
		err := ShowWorktree("show-test", "json", "")
		if err != nil {
			t.Errorf("ShowWorktree failed: %v", err)
		}
	})

	t.Run("show specific field", func(t *testing.T) {
		fields := []string{"name", "branch", "path", "head"}
		for _, field := range fields {
			err := ShowWorktree("show-test", "", field)
			if err != nil {
				t.Errorf("ShowWorktree with field '%s' failed: %v", field, err)
			}
		}
	})

	t.Run("show non-existent worktree should fail", func(t *testing.T) {
		err := ShowWorktree("non-existent", "pretty", "")
		if err == nil {
			t.Error("Expected error for non-existent worktree, got nil")
		}
	})
}

func TestRemoveWorktree(t *testing.T) {
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

	t.Run("remove worktree with force flag", func(t *testing.T) {
		if err := AddWorktree("remove-test", "", "", ""); err != nil {
			t.Fatalf("AddWorktree failed: %v", err)
		}

		err := RemoveWorktree("remove-test", RemoveOptions{Force: true})
		if err != nil {
			t.Errorf("RemoveWorktree failed: %v", err)
		}

		// Verify worktree was removed
		worktrees, err := getWorktrees()
		if err != nil {
			t.Errorf("getWorktrees failed: %v", err)
		}

		for _, wt := range worktrees {
			if wt.Name == "remove-test" {
				t.Error("Worktree 'remove-test' was not removed")
			}
		}
	})

	t.Run("remove worktree and delete branch safely", func(t *testing.T) {
		const name = "remove-branch-safe"
		if err := AddWorktree(name, "", "", ""); err != nil {
			t.Fatalf("AddWorktree failed: %v", err)
		}

		if err := RemoveWorktree(name, RemoveOptions{Force: true, BranchDelete: BranchDeleteSafe}); err != nil {
			t.Fatalf("RemoveWorktree with branch delete failed: %v", err)
		}

		cmd := exec.Command("git", "branch", "--list", name)
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if strings.TrimSpace(string(output)) != "" {
			t.Errorf("expected branch %q to be deleted, got %q", name, strings.TrimSpace(string(output)))
		}
	})

	t.Run("remove worktree with force branch deletion", func(t *testing.T) {
		const name = "remove-branch-force"
		if err := AddWorktree(name, "", "", ""); err != nil {
			t.Fatalf("AddWorktree failed: %v", err)
		}

		worktrees, err := getWorktrees()
		if err != nil {
			t.Fatalf("getWorktrees failed: %v", err)
		}

		var worktreePath string
		for _, wt := range worktrees {
			if wt.Name == name {
				worktreePath = wt.Path
				break
			}
		}
		if worktreePath == "" {
			t.Fatalf("worktree path for %s not found", name)
		}

		filePath := filepath.Join(worktreePath, "unmerged.txt")
		if err := os.WriteFile(filePath, []byte("unmerged change"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		cmd := exec.Command("git", "add", "unmerged.txt")
		cmd.Dir = worktreePath
		if err := cmd.Run(); err != nil {
			t.Fatalf("git add failed: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "unmerged change")
		cmd.Dir = worktreePath
		if err := cmd.Run(); err != nil {
			t.Fatalf("git commit failed: %v", err)
		}

		if err := RemoveWorktree(name, RemoveOptions{Force: true, BranchDelete: BranchDeleteForce}); err != nil {
			t.Fatalf("RemoveWorktree with force branch delete failed: %v", err)
		}

		cmd = exec.Command("git", "branch", "--list", name)
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if strings.TrimSpace(string(output)) != "" {
			t.Errorf("expected branch %q to be deleted, got %q", name, strings.TrimSpace(string(output)))
		}
	})

	t.Run("remove worktree safe branch deletion fails on unmerged branch", func(t *testing.T) {
		const name = "remove-branch-safe-fail"
		if err := AddWorktree(name, "", "", ""); err != nil {
			t.Fatalf("AddWorktree failed: %v", err)
		}

		worktrees, err := getWorktrees()
		if err != nil {
			t.Fatalf("getWorktrees failed: %v", err)
		}

		var worktreePath string
		for _, wt := range worktrees {
			if wt.Name == name {
				worktreePath = wt.Path
				break
			}
		}
		if worktreePath == "" {
			t.Fatalf("worktree path for %s not found", name)
		}

		filePath := filepath.Join(worktreePath, "pending.txt")
		if err := os.WriteFile(filePath, []byte("pending change"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		cmd := exec.Command("git", "add", "pending.txt")
		cmd.Dir = worktreePath
		if err := cmd.Run(); err != nil {
			t.Fatalf("git add failed: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "pending change")
		cmd.Dir = worktreePath
		if err := cmd.Run(); err != nil {
			t.Fatalf("git commit failed: %v", err)
		}

		err = RemoveWorktree(name, RemoveOptions{Force: true, BranchDelete: BranchDeleteSafe})
		if err == nil {
			t.Fatal("expected error when deleting branch with unmerged commits")
		}
		if !strings.Contains(err.Error(), "failed to delete branch") {
			t.Errorf("unexpected error: %v", err)
		}

		cmd = exec.Command("git", "branch", "--list", name)
		cmd.Dir = repoPath
		output, listErr := cmd.Output()
		if listErr != nil {
			t.Fatalf("git branch --list failed: %v", listErr)
		}
		if !strings.Contains(strings.TrimSpace(string(output)), name) {
			t.Errorf("expected branch %q to remain after failed deletion", name)
		}

		cleanup := exec.Command("git", "branch", "-D", name)
		cleanup.Dir = repoPath
		if err := cleanup.Run(); err != nil {
			t.Fatalf("cleanup branch failed: %v", err)
		}
	})

	t.Run("remove non-existent worktree should fail", func(t *testing.T) {
		err := RemoveWorktree("non-existent", RemoveOptions{Force: true})
		if err == nil {
			t.Error("Expected error for non-existent worktree, got nil")
		}
	})
}

func TestGetWorktrees(t *testing.T) {
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

	t.Run("get worktrees from empty repo", func(t *testing.T) {
		worktrees, err := getWorktrees()
		if err != nil {
			t.Errorf("getWorktrees failed: %v", err)
		}

		// Should have at least the main worktree
		if len(worktrees) == 0 {
			t.Error("Expected at least one worktree (main)")
		}
	})

	t.Run("get worktrees after adding some", func(t *testing.T) {
		AddWorktree("wt1", "", "", "")
		AddWorktree("wt2", "", "", "")

		worktrees, err := getWorktrees()
		if err != nil {
			t.Errorf("getWorktrees failed: %v", err)
		}

		// Should have main + 2 added worktrees
		if len(worktrees) < 3 {
			t.Errorf("Expected at least 3 worktrees, got %d", len(worktrees))
		}
	})
}
