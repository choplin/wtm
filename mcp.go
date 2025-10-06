package main

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
	// Force skips the confirmation prompt before removing the worktree
	Force bool `json:"force,omitempty" jsonschema:"skip confirmation prompt"`
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

func handleAddWorktree(ctx context.Context, req *mcp.CallToolRequest, input AddWorktreeInput) (*mcp.CallToolResult, AddWorktreeOutput, error) {
	err := AddWorktree(input.Name, input.Branch, input.Checkout, input.Base)
	if err != nil {
		return nil, AddWorktreeOutput{}, fmt.Errorf("failed to add worktree: %w", err)
	}

	// Get the created worktree info
	worktrees, err := getWorktrees()
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

func handleListWorktrees(ctx context.Context, req *mcp.CallToolRequest, input ListWorktreesInput) (*mcp.CallToolResult, ListWorktreesOutput, error) {
	worktrees, err := getWorktrees()
	if err != nil {
		return nil, ListWorktreesOutput{}, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return nil, ListWorktreesOutput{Worktrees: worktrees}, nil
}

func handleShowWorktree(ctx context.Context, req *mcp.CallToolRequest, input ShowWorktreeInput) (*mcp.CallToolResult, ShowWorktreeOutput, error) {
	worktrees, err := getWorktrees()
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

func handleRemoveWorktree(ctx context.Context, req *mcp.CallToolRequest, input RemoveWorktreeInput) (*mcp.CallToolResult, RemoveWorktreeOutput, error) {
	if input.DeleteBranch && input.DeleteBranchForce {
		return nil, RemoveWorktreeOutput{
			Removed: false,
			Message: "Cannot combine deleteBranch and deleteBranchForce options",
		}, nil
	}

	opts := RemoveOptions{Force: input.Force}
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

// StartMCPServer starts the MCP server over stdio transport
func StartMCPServer(ctx context.Context) error {
	server := newMCPServer()

	// Run server over stdio transport
	transport := &mcp.StdioTransport{}
	return server.Run(ctx, transport)
}

func newMCPServer() *mcp.Server {
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

	return server
}
