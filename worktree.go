package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Worktree represents a git worktree
type Worktree struct {
	Name    string    `json:"name"`
	Branch  string    `json:"branch"`
	Path    string    `json:"path"`
	HEAD    string    `json:"head"`
	Created time.Time `json:"created"`
}

// runGitCommand executes a git command and returns the output
func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(output))
	}
	return string(output), nil
}

// AddWorktree creates a new worktree
func AddWorktree(name, branch, checkout, base string) error {
	// Validate we're in a git repository
	if _, err := runGitCommand("rev-parse", "--git-dir"); err != nil {
		return fmt.Errorf("not in a git repository")
	}

	// Check if worktree already exists
	worktrees, err := getWorktrees()
	if err != nil {
		return err
	}
	for _, wt := range worktrees {
		if wt.Name == name {
			return fmt.Errorf("worktree '%s' already exists", name)
		}
	}

	// Determine the path for the worktree
	gitDir, err := runGitCommand("rev-parse", "--git-dir")
	if err != nil {
		return err
	}
	gitDir = strings.TrimSpace(gitDir)
	worktreePath := filepath.Join(gitDir, "worktrees", name)

	// Build git worktree add command
	var args []string

	if checkout != "" && branch != "" {
		return fmt.Errorf("cannot use both -b and -B options")
	}

	if branch != "" {
		// Create new branch
		args = []string{"worktree", "add", worktreePath, "-b", branch}
		if base != "" {
			args = append(args, base)
		}
	} else if checkout != "" {
		// Checkout existing branch
		args = []string{"worktree", "add", worktreePath, checkout}
	} else {
		// Default: create branch with same name as worktree
		args = []string{"worktree", "add", worktreePath, "-b", name}
		if base != "" {
			args = append(args, base)
		}
	}

	// Execute git worktree add
	if _, err := runGitCommand(args...); err != nil {
		return err
	}

	// Get the created worktree info for success message
	worktrees, err = getWorktrees()
	if err != nil {
		return err
	}

	for _, wt := range worktrees {
		if wt.Name == name {
			fmt.Printf("✓ Created worktree: %s\n", wt.Name)
			fmt.Printf("  Branch: %s\n", wt.Branch)
			fmt.Printf("  Path: %s\n", wt.Path)
			return nil
		}
	}

	return nil
}

// ListWorktrees lists all worktrees
func ListWorktrees(format string) error {
	worktrees, err := getWorktrees()
	if err != nil {
		return err
	}

	switch format {
	case "table":
		printTableFormat(worktrees)
	case "plain":
		printPlainFormat(worktrees)
	case "json":
		printJSONFormat(worktrees)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	return nil
}

// ShowWorktree shows detailed information about a worktree
func ShowWorktree(name, format, field string) error {
	worktrees, err := getWorktrees()
	if err != nil {
		return err
	}

	var target *Worktree
	for _, wt := range worktrees {
		if wt.Name == name {
			target = &wt
			break
		}
	}

	if target == nil {
		return fmt.Errorf("worktree '%s' not found", name)
	}

	if field != "" {
		return printField(target, field)
	}

	switch format {
	case "pretty":
		printPrettyFormat(target)
	case "json":
		data, err := json.MarshalIndent(target, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	return nil
}

// RemoveWorktree removes a worktree
func RemoveWorktree(name string, force bool) error {
	worktrees, err := getWorktrees()
	if err != nil {
		return err
	}

	var target *Worktree
	for _, wt := range worktrees {
		if wt.Name == name {
			target = &wt
			break
		}
	}

	if target == nil {
		return fmt.Errorf("worktree '%s' not found", name)
	}

	// Confirm unless force flag is set
	if !force {
		fmt.Printf("Remove worktree '%s' (branch: %s)? [y/N]: ", target.Name, target.Branch)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted")
			return nil
		}
	}

	// Remove worktree
	if _, err := runGitCommand("worktree", "remove", "--force", target.Path); err != nil {
		return err
	}

	fmt.Printf("✓ Removed worktree: %s\n", target.Name)
	return nil
}

// getWorktrees retrieves all worktrees from git
func getWorktrees() ([]Worktree, error) {
	output, err := runGitCommand("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	var current Worktree

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "worktree":
			current.Path = value
			// Extract name from path (last segment)
			current.Name = filepath.Base(value)
		case "HEAD":
			current.HEAD = value
		case "branch":
			// Extract branch name from refs/heads/branch-name
			if strings.HasPrefix(value, "refs/heads/") {
				current.Branch = strings.TrimPrefix(value, "refs/heads/")
			} else {
				current.Branch = value
			}
		}
	}

	// Add last worktree if exists
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	// Get creation time for each worktree
	for i := range worktrees {
		info, err := os.Stat(worktrees[i].Path)
		if err == nil {
			worktrees[i].Created = info.ModTime()
		}
	}

	return worktrees, nil
}

// printTableFormat prints worktrees in table format
func printTableFormat(worktrees []Worktree) {
	if len(worktrees) == 0 {
		return
	}

	fmt.Printf("%-20s %-30s %-15s\n", "NAME", "BRANCH", "CREATED")
	for _, wt := range worktrees {
		created := formatTimeAgo(wt.Created)
		fmt.Printf("%-20s %-30s %-15s\n", wt.Name, wt.Branch, created)
	}
}

// printPlainFormat prints worktrees in plain format
func printPlainFormat(worktrees []Worktree) {
	for _, wt := range worktrees {
		fmt.Printf("%s %s %s\n", wt.Name, wt.Branch, wt.Path)
	}
}

// printJSONFormat prints worktrees in JSON format
func printJSONFormat(worktrees []Worktree) {
	data, err := json.MarshalIndent(worktrees, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// printPrettyFormat prints a single worktree in pretty format
func printPrettyFormat(wt *Worktree) {
	fmt.Printf("Name:     %s\n", wt.Name)
	fmt.Printf("Branch:   %s\n", wt.Branch)
	fmt.Printf("Path:     %s\n", wt.Path)
	fmt.Printf("HEAD:     %s\n", wt.HEAD)
	fmt.Printf("Created:  %s\n", wt.Created.Format("2006-01-02 15:04:05"))
}

// printField prints a specific field of a worktree
func printField(wt *Worktree, field string) error {
	switch field {
	case "name":
		fmt.Println(wt.Name)
	case "branch":
		fmt.Println(wt.Branch)
	case "path":
		fmt.Println(wt.Path)
	case "head":
		fmt.Println(wt.HEAD)
	case "created":
		fmt.Println(wt.Created.Format(time.RFC3339))
	default:
		return fmt.Errorf("unknown field: %s", field)
	}
	return nil
}

// formatTimeAgo formats a time as a relative time string
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	diff := time.Since(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		return t.Format("2006-01-02")
	}
}