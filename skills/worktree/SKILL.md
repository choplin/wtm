---
name: worktree
description: >
  Manage Git worktrees using the wtm CLI. Triggers on: create worktree, add worktree,
  list worktrees, show worktree, remove worktree, delete worktree, switch worktree,
  worktree management, parallel branches, isolated workspace.
---

# Worktree Management with wtm

You manage Git worktrees using the `wtm` CLI tool.

## Prerequisites

- `wtm` must be available on PATH
- Must be run inside a Git repository

## Commands

### wtm add - Create a new worktree

```bash
wtm add <name> [flags]
```

Creates a worktree under `.wtm/<name>`.

**Arguments:**
- `<name>` (optional if `-B` is used): Worktree directory name. When omitted with `-B`, the name is derived from the branch name with slashes replaced by hyphens.

**Flags:**
- `-b, --branch <name>`: Create a new branch with the specified name
- `-B, --checkout <name>`: Use an existing branch
- `--base <branch>`: Set the base branch for a new branch (defaults to current HEAD)

**Behavior:**
- By default, creates both a worktree and a branch with the same `<name>`
- `-b` and `-B` cannot be combined
- Post-add hooks run automatically after creation

**Examples:**
```bash
wtm add feature-auth                              # New worktree and branch "feature-auth"
wtm add api -b feature/api-refactoring --base main # New branch from main, worktree "api"
wtm add review-456 -B origin/feature/complex       # Existing branch, worktree "review-456"
wtm add -B feature/long-name                       # Worktree auto-named "feature-long-name"
```

### wtm list - List all worktrees

```bash
wtm list [--format <format>]
```

Alias: `wtm ls`

**Formats:**
- `table` (default): Human-readable table with NAME, BRANCH, CREATED columns
- `plain`: Script-friendly output: `<name> <branch> <path>` per line
- `json`: Machine-readable JSON array

**Examples:**
```bash
wtm list                 # Table format
wtm list --format plain  # Script-friendly
wtm list --format json   # JSON output
```

### wtm show - Show worktree details

```bash
wtm show <name> [flags]
```

**Flags:**
- `--format <format>`: Output format (`pretty` default, `json`)
- `-f, --field <field>`: Output only a specific field

**Available fields:** `name`, `branch`, `path`, `head`, `created`

**Examples:**
```bash
wtm show api                # Pretty format
wtm show api --format json  # JSON
wtm show api -f path        # Only the absolute path
wtm show api -f branch      # Only the branch name
```

### wtm remove - Remove a worktree

```bash
wtm remove <name> [flags]
```

Alias: `wtm rm`

**Flags:**
- `-f, --force`: Skip interactive confirmation
- `-d, --delete-branch`: Delete associated branch (safe deletion)
- `-D, --delete-branch-force`: Force delete associated branch

**Behavior:**
- Without `--force`, prompts for confirmation
- `-d` and `-D` cannot be combined
- Even if branch deletion fails, the worktree is still removed

**Examples:**
```bash
wtm remove feature-auth           # Interactive confirmation
wtm remove feature-auth --force   # Skip confirmation
wtm remove feature-auth -d        # Remove worktree and safely delete branch
wtm remove feature-auth -D        # Remove worktree and force delete branch
```

### wtm config edit - Edit project configuration

```bash
wtm config edit
```

Opens `.wtm/config.toml` in `$EDITOR`. Creates the file if it doesn't exist.

### wtm version - Show version

```bash
wtm version
```

## .wtm/ Directory Structure

```
<repo-root>/
└── .wtm/
    ├── .gitignore          # Auto-created, excludes worktrees from git
    ├── config.toml         # Project-specific configuration and hooks
    ├── <worktree-name>/    # Git worktree checkout
    └── <worktree-name>/    # Another worktree
```

`wtm` is stateless. Git stores all worktree metadata. The `.wtm/` directory only contains the actual worktree checkouts and configuration.

## Hooks

Hooks run automatically after `wtm add`. Global hooks run first, then project hooks.

### Configuration locations

- **Global:** `~/.config/wtm/config.toml` (or `$XDG_CONFIG_HOME/wtm/config.toml`)
- **Project:** `.wtm/config.toml`

### Hook types

| Type    | Fields  | Description                                  |
|---------|---------|----------------------------------------------|
| `copy`  | `paths` | Copy files from repo root to worktree        |
| `link`  | `paths` | Create symlinks from worktree to repo root   |
| `shell` | `run`   | Execute a shell command in worktree directory |

### Error handling

Each hook can set `on_error`:
- `"warn"` (default): Print warning and continue
- `"continue"`: Silently skip on error
- `"abort"`: Stop execution (worktree remains created)

### Example configuration

```toml
# .wtm/config.toml

[[hooks.post-add]]
type = "copy"
paths = [".env", "config/local.json"]

[[hooks.post-add]]
type = "link"
paths = ["node_modules"]
on_error = "continue"

[[hooks.post-add]]
type = "shell"
run = "npm install"
on_error = "abort"
```

## Guidelines

- Always use `--force` flag when removing worktrees non-interactively
- Use `wtm list --format json` or `wtm show <name> --format json` when you need to parse output programmatically
- Use `wtm show <name> -f path` to get the absolute path for `cd` or file operations
- Worktree names must be unique within a repository
