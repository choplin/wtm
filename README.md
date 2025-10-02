# Worktree Manager (wtm)

Minimal git worktree management tool that does one thing well.

## Why wtm?

Managing multiple git worktrees manually is tedious. Native `git worktree` commands are verbose and require managing paths and branch names separately. wtm simplifies this by:

- Providing simple, memorable commands for common worktree operations
- Separating worktree names from branch names for flexible naming strategies
- Offering multiple output formats (table/plain/json) for shell scripting
- Remaining stateless - no configuration files or metadata to manage

## Installation

### Homebrew (macOS/Linux)

```bash
brew install choplin/tap/wtm
```

### Go install

```bash
go install github.com/choplin/wtm@latest
```

### Build from source

```bash
git clone https://github.com/choplin/wtm.git
cd wtm
go build
```

## Usage

### Create a worktree

```bash
# Simple: worktree name = branch name
wtm add feature-auth

# Different worktree and branch names
wtm add api -b feature/api-refactoring --base main

# Checkout existing branch with simple worktree name
wtm add review-pr-456 -B origin/feature/complex-branch-name
```

**Options:**

- `-b, --branch <name>`: Create new branch with specified name
- `-B, --checkout <name>`: Use existing branch
- `--base <branch>`: Base branch for new branch (default: current HEAD)

### List worktrees

```bash
# Table format (default)
wtm list

# Plain format (for scripting)
wtm list --format plain

# JSON format
wtm list --format json
```

### Show worktree details

```bash
# Pretty format (default)
wtm show api

# JSON format
wtm show api --format json

# Get specific field
wtm show api --field path
wtm show api -f branch
```

**Available fields:** `name`, `branch`, `path`, `head`, `created`

### Remove a worktree

```bash
# With confirmation prompt
wtm remove feature-auth

# Skip confirmation
wtm remove feature-auth --force
```

### Version information

```bash
wtm version
```

### MCP Server (for AI integration)

Start an MCP (Model Context Protocol) server to enable AI tools like Claude Code to manage worktrees:

```bash
wtm mcp
```

The MCP server exposes four tools over stdio:

- `wtm_add`: Create a new worktree
- `wtm_list`: List all worktrees
- `wtm_show`: Show worktree details
- `wtm_remove`: Remove a worktree

**Example MCP configuration for Claude Code:**

```json
{
  "mcpServers": {
    "wtm": {
      "command": "/path/to/wtm",
      "args": ["mcp"]
    }
  }
}
```

## Shell Integration

### Quick navigation

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
# Navigate to worktree by name
wtm-cd() {
    local path=$(wtm show "$1" --field path 2>/dev/null)
    if [ -n "$path" ]; then
        cd "$path"
    else
        echo "Worktree not found: $1" >&2
        return 1
    fi
}
alias wcd=wtm-cd

# Usage
wcd api
```

### fzf integration

```bash
# Interactive worktree selector
wtm-select() {
    local selection=$(wtm list --format plain | fzf --preview 'wtm show {1} --format pretty' | awk '{print $1}')
    if [ -n "$selection" ]; then
        wtm-cd "$selection"
    fi
}
alias wsel=wtm-select

# Usage
wsel  # Opens fzf selector
```

### Get branch name for scripting

```bash
# Compare with main branch
git diff main..$(wtm show api -f branch)

# Check status of all worktrees
for name in $(wtm list --format plain | awk '{print $1}'); do
    echo "$name: $(git -C $(wtm show $name -f path) status --short)"
done
```

## Examples

### Multiple approaches for the same issue

```bash
# Try different solutions in parallel
wtm add fix-123-approach1 -b bugfix/issue-123
wtm add fix-123-approach2 -b bugfix/issue-123-alternative

# Compare and choose the best one
git diff $(wtm show fix-123-approach1 -f branch)..$(wtm show fix-123-approach2 -f branch)
```

### PR review workflow

```bash
# Create worktree for PR review
wtm add review-pr-456 -B origin/feature/complex-branch-name

# Navigate and review
wcd review-pr-456
git log -p

# Clean up after review
wtm remove review-pr-456 -f
```

## Design Philosophy

### Do One Thing Well

`wtm` focuses exclusively on git worktree management. No session management, no task execution, no unnecessary complexity.

### Stateless Operation

Git is the single source of truth. No metadata files, no `.wtm` directory. All information is derived directly from git commands.

### Worktree Names vs Branch Names

Unlike other tools, `wtm` separates worktree names from branch names. This allows:

- Simple, memorable worktree names
- Complex branch names (e.g., `feature/long/nested/name`)
- Multiple worktrees for the same branch

## Comparison with Alternatives

### vs. `git worktree` (native)

**git worktree:**

```bash
git worktree add ../project-feature-x -b feature-x origin/main
cd ../project-feature-x
```

**wtm:**

```bash
wtm add feature-x --base origin/main
wcd feature-x
```

### vs. amux

`wtm` is inspired by lessons learned from [amux](https://github.com/choplin/amux):

- Simpler: 4 commands instead of 20+
- Stateless: No metadata files
- Focused: No session/task management
- Lighter: < 500 LOC vs. thousands

## Development

### Run tests

```bash
go test -v
```

### Build

```bash
go build -o wtm
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see LICENSE file for details
