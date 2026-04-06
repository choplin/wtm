package wtm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const noteFileName = "wtm-notes"

// noteFilePath resolves the path to the note file for a given worktree name.
// For linked worktrees: .git/worktrees/<name>/wtm-notes
// For the main worktree: .git/wtm-notes
func noteFilePath(name string) (string, error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return "", err
	}

	gitDir := filepath.Join(repoRoot, ".git")

	// Check if it's a linked worktree
	wtDir := filepath.Join(gitDir, "worktrees", name)
	if info, err := os.Stat(wtDir); err == nil && info.IsDir() {
		return filepath.Join(wtDir, noteFileName), nil
	}

	// Check if the name matches an existing worktree (could be the main worktree).
	// Use git worktree list directly to avoid circular calls via GetWorktrees.
	output, err := runGitCommand("worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "worktree ") {
			wtPath := strings.TrimPrefix(line, "worktree ")
			if filepath.Base(wtPath) == name {
				return filepath.Join(gitDir, noteFileName), nil
			}
		}
	}

	return "", fmt.Errorf("worktree '%s' not found", name)
}

// readNoteFile reads a note file for a worktree given the repo root.
// This is used internally by GetWorktrees to avoid circular calls.
func readNoteFile(repoRoot, name string, isMain bool) string {
	gitDir := filepath.Join(repoRoot, ".git")

	var path string
	if isMain {
		path = filepath.Join(gitDir, noteFileName)
	} else {
		path = filepath.Join(gitDir, "worktrees", name, noteFileName)
	}

	data, err := os.ReadFile(path) //nolint:gosec // Path is derived from git directory
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(data), "\n")
}

// GetNote reads the note for the given worktree.
// Returns empty string and nil error if the note file does not exist.
func GetNote(name string) (string, error) {
	path, err := noteFilePath(name)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path) //nolint:gosec // Path is derived from git directory
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimRight(string(data), "\n"), nil
}

// AddNote writes a note for the given worktree.
// If force is false and a note already exists, it returns an error.
func AddNote(name, message string, force bool) error {
	path, err := noteFilePath(name)
	if err != nil {
		return err
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("note already exists for worktree '%s' (use -f to overwrite)", name)
		}
	}

	return os.WriteFile(path, []byte(message+"\n"), 0o644)
}

// EditNote opens the note file in the user's $EDITOR.
func EditNote(name string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("EDITOR environment variable is not set")
	}

	path, err := noteFilePath(name)
	if err != nil {
		return err
	}

	// Create the file if it doesn't exist so the editor can open it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
			return fmt.Errorf("failed to create note file: %w", err)
		}
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemoveNote deletes the note for the given worktree.
// Returns an error if no note exists.
func RemoveNote(name string) error {
	path, err := noteFilePath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("no note found for worktree '%s'", name)
	}

	return os.Remove(path)
}
