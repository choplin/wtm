package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool input/output structures

type AddWorktreeInput struct {
	Name     string `json:"name" jsonschema:"required,description=Name of the worktree (used as directory name)"`
	Branch   string `json:"branch,omitempty" jsonschema:"description=Create new branch with this name (default: same as worktree name)"`
	Checkout string `json:"checkout,omitempty" jsonschema:"description=Use existing branch with this name"`
	Base     string `json:"base,omitempty" jsonschema:"description=Base branch for new branch (default: current HEAD)"`
}

type AddWorktreeOutput struct {
	Name   string `json:"name" jsonschema:"description=Created worktree name"`
	Branch string `json:"branch" jsonschema:"description=Branch name"`
	Path   string `json:"path" jsonschema:"description=Absolute path to the worktree"`
}

type ListWorktreesInput struct{}

type ListWorktreesOutput struct {
	Worktrees []Worktree `json:"worktrees" jsonschema:"description=List of all worktrees"`
}

type ShowWorktreeInput struct {
	Name string `json:"name" jsonschema:"required,description=Name of the worktree to show"`
}

type ShowWorktreeOutput struct {
	Worktree Worktree `json:"worktree" jsonschema:"description=Worktree details"`
}

type RemoveWorktreeInput struct {
	Name  string `json:"name" jsonschema:"required,description=Name of the worktree to remove"`
	Force bool   `json:"force,omitempty" jsonschema:"description=Skip confirmation prompt"`
}

type RemoveWorktreeOutput struct {
	Removed bool   `json:"removed" jsonschema:"description=Whether the worktree was removed"`
	Message string `json:"message" jsonschema:"description=Result message"`
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
	err := RemoveWorktree(input.Name, input.Force)
	if err != nil {
		return nil, RemoveWorktreeOutput{
			Removed: false,
			Message: fmt.Sprintf("Failed to remove worktree: %v", err),
		}, nil
	}

	return nil, RemoveWorktreeOutput{
		Removed: true,
		Message: fmt.Sprintf("Removed worktree: %s", input.Name),
	}, nil
}

// StartMCPServer starts the MCP server over stdio transport
func StartMCPServer(ctx context.Context) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "wtm",
		Version: version,
	}, nil)

	// Register tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_add",
		Description: "Create a new git worktree with specified name. The worktree name is used as the directory name and identifier, independent from the branch name.",
	}, handleAddWorktree)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_list",
		Description: "List all git worktrees in the current repository.",
	}, handleListWorktrees)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_show",
		Description: "Show detailed information about a specific worktree by name.",
	}, handleShowWorktree)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "wtm_remove",
		Description: "Remove a git worktree by name. Optionally skip confirmation prompt with force flag.",
	}, handleRemoveWorktree)

	// Run server over stdio transport
	transport := &mcp.StdioTransport{}
	return server.Run(ctx, transport)
}
