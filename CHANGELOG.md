---
created: 2025-10-02
updated: 2026-03-04
---

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `wtm notes` command group for attaching freeform text notes to worktrees (`add`, `show`, `edit`, `remove`)
- `-m/--message` flag on `wtm add` to attach a note at creation time
- MCP tools for worktree notes: `wtm_notes_add`, `wtm_notes_show`, `wtm_notes_remove`
- Notes are displayed in `wtm show` output and accessible via `wtm show <name> -f note`

## [0.8.0] - 2026-03-04

### Added

- Claude Code plugin support with worktree management skill

## [0.7.0] - 2026-02-26

### Added

- Global hooks support: define hooks in `~/.config/wtm/config.toml` to apply across all repositories
- Automatic `.gitignore` creation in `.wtm/` directory to exclude worktrees from git tracking

### Changed

- Moved worktree storage from `.git/wtm/worktrees/` to `.wtm/` for simpler path structure
- Moved project configuration from `.git/wtm/config.toml` to `.wtm/config.toml`

## [0.6.1] - 2026-01-18

### Fixed

- Homebrew formula now includes `git-wtm` symlink for `git wtm` subcommand support

## [0.6.0] - 2026-01-18

### Added

- `wtm config edit` command to open project configuration in `$EDITOR`
- Hook file operations: `copy` and `link` options for copying files or creating symlinks from repo root to worktree
- Post-add hook support via `.wtm/config.toml` for running commands after worktree creation
- Shell completion command (`wtm completion [bash|zsh|fish|powershell]`)
- Dynamic worktree name completion for `show` and `remove` commands
- Git subcommand support (`git wtm`) via `make install-git`

## [0.5.0] - 2025-11-12

### Added

- Allow omitting worktree name when using `-B` option to checkout existing branch. The branch name is used as worktree name with slashes replaced by hyphens.

## [0.4.0] - 2025-10-09

### Added

- Added `ls` alias for `wtm list` to mirror common UNIX tooling.
- Added `rm` alias for `wtm remove` so deletion flows have a short form.

### Fixed

- Removed the MCP `force` option and always run worktree removal non-interactively to avoid prompts during AI execution.

## [0.3.0] - 2025-10-09

### Changed

- Added a `(primary)` suffix to the main worktree in the `wtm list` NAME column so it stands out from other worktrees.

## [0.1.1] - 2025-10-06

### Added

- In-memory MCP server end-to-end test verifying tool registration and schema metadata.

### Changed

- Simplified MCP tool schema descriptions and refactored server setup for reuse in tests.

## [0.1.0] - 2025-10-02

### Added

- Core worktree management with 4 commands: `add`, `list`, `show`, `remove`
- Multiple output formats: table, plain, JSON
- Worktree name and branch name separation for flexible naming
- MCP (Model Context Protocol) server for AI tool integration
- GitHub Actions workflows for automated testing (Ubuntu/macOS)
- GoReleaser configuration for cross-platform releases
- Homebrew tap support (`brew install choplin/tap/wtm`)
- Shell integration examples (wtm-cd, fzf integration)
- MIT License

[Unreleased]: https://github.com/choplin/wtm/compare/v0.8.0...HEAD
[0.8.0]: https://github.com/choplin/wtm/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/choplin/wtm/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/choplin/wtm/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/choplin/wtm/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/choplin/wtm/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/choplin/wtm/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/choplin/wtm/compare/v0.1.1...v0.3.0
[0.1.1]: https://github.com/choplin/wtm/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/choplin/wtm/releases/tag/v0.1.0
