package wtm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoteFilePath(t *testing.T) {
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

	// Create a linked worktree
	if err := AddWorktree("linked-wt", "", "", "", ""); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	t.Run("linked worktree resolves to .git/worktrees/<name>/wtm-notes", func(t *testing.T) {
		path, err := noteFilePath("linked-wt")
		if err != nil {
			t.Fatalf("noteFilePath failed: %v", err)
		}
		if !strings.HasSuffix(path, filepath.Join(".git", "worktrees", "linked-wt", "wtm-notes")) {
			t.Errorf("expected path ending with .git/worktrees/linked-wt/wtm-notes, got %s", path)
		}
	})

	t.Run("main worktree resolves to .git/wtm-notes", func(t *testing.T) {
		mainName := filepath.Base(repoPath)
		path, err := noteFilePath(mainName)
		if err != nil {
			t.Fatalf("noteFilePath failed: %v", err)
		}
		if !strings.HasSuffix(path, filepath.Join(".git", "wtm-notes")) {
			t.Errorf("expected path ending with .git/wtm-notes, got %s", path)
		}
		// Should NOT end with worktrees path
		if strings.Contains(path, "worktrees") {
			t.Errorf("main worktree path should not contain 'worktrees', got %s", path)
		}
	})

	t.Run("non-existent worktree returns error", func(t *testing.T) {
		_, err := noteFilePath("no-such-worktree")
		if err == nil {
			t.Error("expected error for non-existent worktree, got nil")
		}
	})
}

func TestAddNote(t *testing.T) {
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

	if err := AddWorktree("note-add", "", "", "", ""); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	t.Run("add note successfully", func(t *testing.T) {
		if err := AddNote("note-add", "test note", false); err != nil {
			t.Fatalf("AddNote failed: %v", err)
		}
		note, err := GetNote("note-add")
		if err != nil {
			t.Fatalf("GetNote failed: %v", err)
		}
		if note != "test note" {
			t.Errorf("expected 'test note', got %q", note)
		}
	})

	t.Run("duplicate note without force returns error", func(t *testing.T) {
		err := AddNote("note-add", "another note", false)
		if err == nil {
			t.Error("expected error for duplicate note, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("force overwrites existing note", func(t *testing.T) {
		if err := AddNote("note-add", "overwritten", true); err != nil {
			t.Fatalf("AddNote with force failed: %v", err)
		}
		note, err := GetNote("note-add")
		if err != nil {
			t.Fatalf("GetNote failed: %v", err)
		}
		if note != "overwritten" {
			t.Errorf("expected 'overwritten', got %q", note)
		}
	})
}

func TestGetNote(t *testing.T) {
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

	if err := AddWorktree("note-get", "", "", "", ""); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	t.Run("returns empty string when no note exists", func(t *testing.T) {
		note, err := GetNote("note-get")
		if err != nil {
			t.Fatalf("GetNote failed: %v", err)
		}
		if note != "" {
			t.Errorf("expected empty string, got %q", note)
		}
	})

	t.Run("returns note content when note exists", func(t *testing.T) {
		if err := AddNote("note-get", "hello world", false); err != nil {
			t.Fatalf("AddNote failed: %v", err)
		}
		note, err := GetNote("note-get")
		if err != nil {
			t.Fatalf("GetNote failed: %v", err)
		}
		if note != "hello world" {
			t.Errorf("expected 'hello world', got %q", note)
		}
	})
}

func TestRemoveNote(t *testing.T) {
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

	if err := AddWorktree("note-rm", "", "", "", ""); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	t.Run("remove existing note", func(t *testing.T) {
		if err := AddNote("note-rm", "to be removed", false); err != nil {
			t.Fatalf("AddNote failed: %v", err)
		}
		if err := RemoveNote("note-rm"); err != nil {
			t.Fatalf("RemoveNote failed: %v", err)
		}
		note, err := GetNote("note-rm")
		if err != nil {
			t.Fatalf("GetNote failed: %v", err)
		}
		if note != "" {
			t.Errorf("expected empty string after removal, got %q", note)
		}
	})

	t.Run("remove non-existent note returns error", func(t *testing.T) {
		err := RemoveNote("note-rm")
		if err == nil {
			t.Error("expected error for non-existent note, got nil")
		}
		if !strings.Contains(err.Error(), "no note found") {
			t.Errorf("expected 'no note found' error, got: %v", err)
		}
	})
}

func TestNoteInGetWorktrees(t *testing.T) {
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

	if err := AddWorktree("noted-wt", "", "", "", ""); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	if err := AddNote("noted-wt", "integrated note", false); err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}

	worktrees, err := GetWorktrees()
	if err != nil {
		t.Fatalf("GetWorktrees failed: %v", err)
	}

	var found bool
	for _, wt := range worktrees {
		if wt.Name == "noted-wt" {
			found = true
			if wt.Note != "integrated note" {
				t.Errorf("expected note 'integrated note', got %q", wt.Note)
			}
		}
	}
	if !found {
		t.Error("worktree 'noted-wt' not found")
	}
}

func TestNoteInShowOutput(t *testing.T) {
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

	if err := AddWorktree("show-note", "", "", "", ""); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	if err := AddNote("show-note", "visible note", false); err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}

	t.Run("pretty format includes note", func(t *testing.T) {
		output, err := captureStdout(t, func() error {
			return ShowWorktree("show-note", "pretty", "")
		})
		if err != nil {
			t.Fatalf("ShowWorktree failed: %v", err)
		}
		if !strings.Contains(output, "Note:     visible note") {
			t.Errorf("expected pretty output to contain note, got:\n%s", output)
		}
	})

	t.Run("json format includes note", func(t *testing.T) {
		output, err := captureStdout(t, func() error {
			return ShowWorktree("show-note", "json", "")
		})
		if err != nil {
			t.Fatalf("ShowWorktree failed: %v", err)
		}
		var wt Worktree
		if err := json.Unmarshal([]byte(output), &wt); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
		if wt.Note != "visible note" {
			t.Errorf("expected note 'visible note' in JSON, got %q", wt.Note)
		}
	})

	t.Run("field output returns note", func(t *testing.T) {
		output, err := captureStdout(t, func() error {
			return ShowWorktree("show-note", "", "note")
		})
		if err != nil {
			t.Fatalf("ShowWorktree failed: %v", err)
		}
		if strings.TrimSpace(output) != "visible note" {
			t.Errorf("expected 'visible note', got %q", strings.TrimSpace(output))
		}
	})
}
