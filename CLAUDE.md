# wtm Project Instructions

## Architecture

```
main.go         → CLI entry point, command parsing
worktree.go     → Core business logic (Add/List/Show/Remove)
mcp.go          → MCP server (thin wrappers around worktree.go)
worktree_test.go → Integration tests with real git repos
```

All operations flow: CLI → worktree.go → git command → parse output

## Design Philosophy

1. **Do One Thing Well**: Focus exclusively on git worktree management
2. **Stateless Operation**: Git is the single source of truth, no metadata files
3. **Minimal Dependencies**: Standard library + essential tools only
4. **Worktree/Branch Separation**: Worktree names independent from branch names

## Core Principles

### Key Technical Decisions

- **Worktree name**: Extracted from path (last segment), not from branch name
- **Git integration**: All operations via `git worktree` porcelain format
- **Output formats**: table (human), plain (scripting), json (programmatic)

### What NOT to Do

- ❌ Add session/task management
- ❌ Create configuration files
- ❌ Store metadata (stay stateless)
- ❌ Add TUI interface
- ❌ Implement hooks/plugins

### Lessons from amux

This project was born from lessons learned developing amux:
- Avoid feature creep (amux had 20+ commands, wtm has 4)
- Stay stateless (amux had metadata, wtm has none)
- Keep it simple (amux was 1000s of LOC, wtm is <1000)

Must justify any addition against core philosophy.
