---
created: 2025-10-02
updated: 2026-01-18
---

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.0] - 2026-01-18

### Added

- `wtm config edit` command to open project configuration in `$EDITOR`
- Hook file operations: `copy` and `link` options for copying files or creating symlinks from repo root to worktree
- Post-add hook support via `.git/wtm/config.toml` for running commands after worktree creation
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

[Unreleased]: https://github.com/choplin/wtm/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/choplin/wtm/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/choplin/wtm/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/choplin/wtm/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/choplin/wtm/compare/v0.1.1...v0.3.0
[0.1.1]: https://github.com/choplin/wtm/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/choplin/wtm/releases/tag/v0.1.0
