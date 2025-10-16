# Worktree Manager (wtm) 🪵

![Go Version](https://img.shields.io/badge/go-1.24.4-00ADD8?logo=go) ![CI](https://github.com/choplin/wtm/actions/workflows/test.yml/badge.svg) ![License](https://img.shields.io/badge/license-MIT-blue)

> Seamless Git worktree orchestration for humans and AI agents alike.

`wtm` keeps every active branch in a tidy `.wtm/` playground with one-line commands, so you can spin up review branches, isolate experiments, and hand off context to AI copilots without juggling checkout paths or branch names.

## ✨ Highlights

- ⚡ Instant setup: one command spins up or removes worktrees—zero path or branch wrangling.
- 🗂️ Hassle-free flow: wtm tracks everything so you never have to manage folder paths or branch names by hand.
- 🤖 AI-ready workflows: the MCP server gives assistants safe access without ceding control.
- 🛠️ Automation ready: structured output drops straight into your scripts and CI checks.

## 🚀 Quick Start

1. Install `wtm` (see Installation for available options).
2. Create a worktree: `wtm add feature-auth`.
3. Jump in and ship: `cd .wtm/feature-auth` or use the `wcd` helper below.

### Example session

```bash
wtm add feature-login
wtm list
wtm show feature-login
wtm remove feature-login --force
```

## 📦 Installation

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
make build   # builds with version embedding
make run     # runs the CLI with embedded version information
make test    # runs go test ./...
```

The `make build` target automatically discovers the version using `git describe` and falls back to `dev` when that metadata is unavailable.

## 🧭 Usage Cheatsheet

### Create a worktree

```bash
wtm add feature-auth
wtm add api -b feature/api-refactoring --base main
wtm add review-pr-456 -B origin/feature/complex-branch-name
```

By default, `wtm add <name>` creates a new branch and worktree that both use `<name>` so you can start working immediately without extra flags.

Options:

- `-b, --branch <name>`: Create a new branch with the provided name.
- `-B, --checkout <name>`: Use an existing branch.
- `--base <branch>`: Set the base branch for a new branch (defaults to current HEAD).

### List worktrees

```bash
wtm list                # table (default)
wtm list --format plain # script-friendly
wtm list --format json  # machine-readable
```

### Show worktree details

```bash
wtm show api
wtm show api --format json
wtm show api --field path
wtm show api -f branch
```

Available fields: `name`, `branch`, `path`, `head`, `created`.

### Remove a worktree

```bash
wtm remove feature-auth
wtm remove feature-auth --force
```

### Version information

```bash
wtm version
```

## 🤖 MCP Server (AI integration)

Launch the MCP (Model Context Protocol) server to let AI agents manage worktrees:

```bash
wtm mcp
```

The server exposes these tools over stdio:

- `wtm_add`: Create a new worktree.
- `wtm_list`: List all worktrees.
- `wtm_show`: Show worktree details.
- `wtm_remove`: Remove a worktree.

### Claude Code example

```json
{
  "mcpServers": {
    "wtm": {
      "command": "wtm",
      "args": ["mcp"]
    }
  }
}
```

## 🗂️ Worktree Layout (`.wtm/`)

By default, `wtm` creates real Git worktrees under `.wtm/<worktree-name>`—whether you run the CLI directly or via the MCP server. Each directory is a standard Git worktree, so you can open it in an editor, run tests, or remove it with `wtm remove`. `wtm` itself remains stateless—Git stores all metadata—while the `.wtm/` folder simply keeps the worktree directories grouped in one place.

## 🧠 Design Principles

### Do One Thing Well

`wtm` focuses exclusively on Git worktree management—no task runners, no extra scaffolding.

### Stateless by Design

Git is the single source of truth. `wtm` reads the repository state directly instead of maintaining custom metadata. The `.wtm/` directory only contains actual Git worktree checkouts.

### Worktree Names vs. Branch Names

`wtm add` defaults to matching the worktree and branch names so you can move fast, yet you can split them whenever a workflow demands it—even running multiple worktrees off the same branch for parallel experiments.

## 📚 Tips & Tricks

### Shell helpers

```bash
wtm-cd() {
    local dir=$(wtm show "$1" -f path)
    if [ -d "$dir" ]; then
        cd "$dir"
    else
        echo "worktree not found: $1" >&2
        return 1
    fi
}
alias wcd=wtm-cd

# Usage
wcd api
```

### fzf integration

```bash
wtm-select() {
    local selection=$(
        wtm list \
            | fzf --header-lines=1 --preview 'wtm show {1} --format pretty' \
            | awk '{print $1}'
    )
    if [ -n "$selection" ]; then
        wtm-cd "$selection"
    fi
}
alias wsel=wtm-select

# Usage
wsel
```

### Scripting helpers

```bash
# Compare with main branch
git diff main..$(wtm show api -f branch)

# Check status of all worktrees
for name in $(wtm list --format plain | awk '{print $1}'); do
    echo "$name: $(git -C $(wtm show "$name" -f path) status --short)"
done
```

### Exploratory workflows

```bash
# Try different solutions in parallel
wtm add fix-123-approach1 -b bugfix/issue-123
wtm add fix-123-approach2 -b bugfix/issue-123-alternative

# Compare and choose the best one
git diff $(wtm show fix-123-approach1 -f branch)..$(wtm show fix-123-approach2 -f branch)
```

## 🧑‍💻 Development

```bash
go test -v
go build -o wtm
```

## 🤝 Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

## 📄 License

MIT License. See `LICENSE` for details.
