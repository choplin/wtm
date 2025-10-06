package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPToolsListInMemory(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server := newMCPServer()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := serverSession.Wait(); err != nil && ctx.Err() == nil {
			t.Errorf("server wait: %v", err)
		}
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "wtm-test-client", Version: "0.0.1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}

	defer func() {
		_ = clientSession.Close()
		wg.Wait()
	}()

	res, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}

	expectedDescriptions := map[string]string{
		"wtm_add":    "Create a new git worktree. Worktree name is used as directory identifier, independent from branch name.",
		"wtm_list":   "List all git worktrees in the current repository with their details.",
		"wtm_remove": "Remove a git worktree by name. Use force flag to skip confirmation. Optionally delete the associated branch.",
		"wtm_show":   "Show detailed information about a specific worktree by name.",
	}

	if len(res.Tools) != len(expectedDescriptions) {
		t.Fatalf("expected %d tools, got %d", len(expectedDescriptions), len(res.Tools))
	}

	for _, tool := range res.Tools {
		want, ok := expectedDescriptions[tool.Name]
		if !ok {
			t.Fatalf("unexpected tool %q", tool.Name)
		}
		if tool.Description != want {
			t.Fatalf("tool %s description mismatch\nwant: %s\ngot:  %s", tool.Name, want, tool.Description)
		}

		switch tool.Name {
		case "wtm_add":
			assertSchemaPropertyDescription(t, tool.InputSchema, "name", "name of the worktree (used as directory name)")
			assertSchemaPropertyDescription(t, tool.InputSchema, "branch", "create new branch with this name (default: same as worktree name)")
			assertSchemaPropertyDescription(t, tool.InputSchema, "checkout", "use existing branch with this name")
			assertSchemaPropertyDescription(t, tool.InputSchema, "base", "base branch for new branch (default: current HEAD)")
			assertSchemaPropertyDescription(t, tool.OutputSchema, "name", "created worktree name")
			assertSchemaPropertyDescription(t, tool.OutputSchema, "branch", "branch name")
			assertSchemaPropertyDescription(t, tool.OutputSchema, "path", "absolute path to the worktree")
		case "wtm_list":
			assertSchemaPropertyDescription(t, tool.OutputSchema, "worktrees", "list of all worktrees")
		case "wtm_remove":
			assertSchemaPropertyDescription(t, tool.InputSchema, "name", "name of the worktree to remove")
			assertSchemaPropertyDescription(t, tool.InputSchema, "force", "skip confirmation prompt")
			assertSchemaPropertyDescription(t, tool.InputSchema, "deleteBranch", "delete associated branch using git branch -d")
			assertSchemaPropertyDescription(t, tool.InputSchema, "deleteBranchForce", "force delete associated branch using git branch -D")
			assertSchemaPropertyDescription(t, tool.OutputSchema, "removed", "whether the worktree was removed")
			assertSchemaPropertyDescription(t, tool.OutputSchema, "message", "result message")
		case "wtm_show":
			assertSchemaPropertyDescription(t, tool.InputSchema, "name", "name of the worktree to show")
			assertSchemaPropertyDescription(t, tool.OutputSchema, "worktree", "worktree details")
		}
	}
}

func assertSchemaPropertyDescription(t *testing.T, schema any, key, want string) {
	t.Helper()
	obj, ok := schema.(map[string]any)
	if !ok {
		t.Fatalf("schema for property %q is not an object", key)
	}
	props, ok := obj["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema missing properties map for %q", key)
	}
	node, ok := props[key].(map[string]any)
	if !ok {
		t.Fatalf("schema missing property %q", key)
	}
	desc, _ := node["description"].(string)
	if desc != want {
		t.Fatalf("property %s description mismatch\nwant: %s\ngot:  %s", key, want, desc)
	}
}
