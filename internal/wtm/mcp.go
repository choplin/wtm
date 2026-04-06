package wtm

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool input/output structures

type AddWorktreeInput struct {
	Name     string `json:"name" jsonschema:"name of the worktree (used as directory name)"`
	Branch   string `json:"branch,omitempty" jsonschema:"create new branch with this name (default: same as worktree name)"`
	Checkout string `json:"checkout,omitempty" jsonschema:"use existing branch with this name"`
	Base     string `json:"base,omitempty" jsonschema:"base branch for new branch (default: current HEAD)"`
	Note     string `json:"note,omitempty" jsonschema:"attach a note to the worktree"`
}

type AddWorktreeOutput struct {
	Name   string `json:"name" jsonschema:"created worktree name"`
	Branch string `json:"branch" jsonschema:"branch name"`
	Path   string `json:"path" jsonschema:"absolute path to the worktree"`
}

type ListWorktreesInput struct{}

type ListWorktreesOutput struct {
	Worktrees []Worktree `json:"worktrees" jsonschema:"list of all worktrees"`
}

type ShowWorktreeInput struct {
	Name string `json:"name" jsonschema:"name of the worktree to show"`
}

type ShowWorktreeOutput struct {
	Worktree Worktree `json:"worktree" jsonschema:"worktree details"`
}

// RemoveWorktreeInput mirrors CLI options for removing a worktree
type RemoveWorktreeInput struct {
	Name string `json:"name" jsonschema:"name of the worktree to remove"`
	// DeleteBranch requests safe branch deletion (git branch -d) after removal
	DeleteBranch bool `json:"deleteBranch,omitempty" jsonschema:"delete associated branch using git branch -d"`
	// DeleteBranchForce requests forceful branch deletion (git branch -D) after removal
	DeleteBranchForce bool `json:"deleteBranchForce,omitempty" jsonschema:"force delete associated branch using git branch -D"`
}

type RemoveWorktreeOutput struct {
	Removed bool   `json:"removed" jsonschema:"whether the worktree was removed"`
	Message string `json:"message" jsonschema:"result message"`
}

// Tool handlers

func handleAddWorktree(_ context.Context, _ *mcp.CallToolRequest, input AddWorktreeInput) (*mcp.CallToolResult, AddWorktreeOutput, error) {
	err := AddWorktree(input.Name, input.Branch, input.Checkout, input.Base, input.Note)
	if err != nil {
		return nil, AddWorktreeOutput{}, fmt.Errorf("failed to add worktree: %w", err)
	}

	// Get the created worktree info
	worktrees, err := GetWorktrees()
	if err != nil {
		return nil, AddWorktreeOutput{}, fmt.Errorf("failed to get worktree info: %w", err)
	}

	for _, wt := range worktrees {
		if wt.Name == input.Name {
			return nil, AddWorktreeOutput{
				Name:   wt.Name,
				Branch: wt.Branch,
				Path:   wt.Path,
			}, nil
		}
	}

	return nil, AddWorktreeOutput{}, fmt.Errorf("worktree created but not found")
}

func handleListWorktrees(_ context.Context, _ *mcp.CallToolRequest, _ ListWorktreesInput) (*mcp.CallToolResult, ListWorktreesOutput, error) {
	worktrees, err := GetWorktrees()
	if err != nil {
		return nil, ListWorktreesOutput{}, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return nil, ListWorktreesOutput{Worktrees: worktrees}, nil
}

func handleShowWorktree(_ context.Context, _ *mcp.CallToolRequest, input ShowWorktreeInput) (*mcp.CallToolResult, ShowWorktreeOutput, error) {
	worktrees, err := GetWorktrees()
	if err != nil {
		return nil, ShowWorktreeOutput{}, fmt.Errorf("failed to get worktrees: %w", err)
	}

	for _, wt := range worktrees {
		if wt.Name == input.Name {
			return nil, ShowWorktreeOutput{Worktree: wt}, nil
		}
	}

	return nil, ShowWorktreeOutput{}, fmt.Errorf("worktree '%s' not found", input.Name)
}

func handleRemoveWorktree(_ context.Context, _ *mcp.CallToolRequest, input RemoveWorktreeInput) (*mcp.CallToolResult, RemoveWorktreeOutput, error) {
	if input.DeleteBranch && input.DeleteBranchForce {
		return nil, RemoveWorktreeOutput{
			Removed: false,
			Message: "Cannot combine deleteBranch and deleteBranchForce options",
		}, nil
	}

	// MCP runs non-interactively, so we always force removal
	opts := RemoveOptions{Force: true}
	switch {
	case input.DeleteBranch:
		opts.BranchDelete = BranchDeleteSafe // safe deletion mirrors git branch -d
	case input.DeleteBranchForce:
		opts.BranchDelete = BranchDeleteForce // force deletion mirrors git branch -D
	}

	err := RemoveWorktree(input.Name, opts)
	if err != nil {
		return nil, RemoveWorktreeOutput{
			Removed: false,
			Message: fmt.Sprintf("Failed to remove worktree: %v", err),
		}, nil
	}

	message := fmt.Sprintf("Removed worktree: %s", input.Name)
	if opts.BranchDelete != BranchDeleteNone {
		message = fmt.Sprintf("%s (branch deleted)", message)
	}

	return nil, RemoveWorktreeOutput{
		Removed: true,
		Message: message,
	}, nil
}

// Notes MCP tool structures

type NotesAddInput struct {
	Name    string `json:"name" jsonschema:"name of the worktree"`
	Message string `json:"message" jsonschema:"note message"`
	Force   bool   `json:"force,omitempty" jsonschema:"overwrite existing note"`
}

type NotesAddOutput struct {
	Message string `json:"message" jsonschema:"result message"`
}

type NotesShowInput struct {
	Name string `json:"name" jsonschema:"name of the worktree"`
}

type NotesShowOutput struct {
	Note string `json:"note" jsonschema:"note content"`
}

type NotesRemoveInput struct {
	Name string `json:"name" jsonschema:"name of the worktree"`
}

type NotesRemoveOutput struct {
	Message string `json:"message" jsonschema:"result message"`
}

func handleNotesAdd(_ context.Context, _ *mcp.CallToolRequest, input NotesAddInput) (*mcp.CallToolResult, NotesAddOutput, error) {
	if err := AddNote(input.Name, input.Message, input.Force); err != nil {
		return nil, NotesAddOutput{}, fmt.Errorf("failed to add note: %w", err)
	}
	return nil, NotesAddOutput{
		Message: fmt.Sprintf("Added note to worktree '%s'", input.Name),
	}, nil
}

func handleNotesShow(_ context.Context, _ *mcp.CallToolRequest, input NotesShowInput) (*mcp.CallToolResult, NotesShowOutput, error) {
	note, err := GetNote(input.Name)
	if err != nil {
		return nil, NotesShowOutput{}, fmt.Errorf("failed to get note: %w", err)
	}
	if note == "" {
		return nil, NotesShowOutput{}, fmt.Errorf("no note found for worktree '%s'", input.Name)
	}
	return nil, NotesShowOutput{Note: note}, nil
}

func handleNotesRemove(_ context.Context, _ *mcp.CallToolRequest, input NotesRemoveInput) (*mcp.CallToolResult, NotesRemoveOutput, error) {
	if err := RemoveNote(input.Name); err != nil {
		return nil, NotesRemoveOutput{}, fmt.Errorf("failed to remove note: %w", err)
	}
	return nil, NotesRemoveOutput{
		Message: fmt.Sprintf("Removed note from worktree '%s'", input.Name),
	}, nil
}

// StartMCPServer starts the MCP server over stdio transport
func StartMCPServer(ctx context.Context, version string) error {
	server := newMCPServer(version)

	// Run server over stdio transport
	transport := &mcp.StdioTransport{}
	return server.Run(ctx, transport)
}

func newMCPServer(version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "wtm",
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_add",
		Description: "Create a new git worktree. Worktree name is used as directory identifier, independent from branch name.",
	}, handleAddWorktree)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_list",
		Description: "List all git worktrees in the current repository with their details.",
	}, handleListWorktrees)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_show",
		Description: "Show detailed information about a specific worktree by name.",
	}, handleShowWorktree)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_remove",
		Description: "Remove a git worktree by name. Use force flag to skip confirmation. Optionally delete the associated branch.",
	}, handleRemoveWorktree)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_notes_add",
		Description: "Add a note to a worktree. Use force to overwrite an existing note.",
	}, handleNotesAdd)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_notes_show",
		Description: "Show the note attached to a worktree.",
	}, handleNotesShow)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_notes_remove",
		Description: "Remove the note from a worktree.",
	}, handleNotesRemove)

	return server
}
