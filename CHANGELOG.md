# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[0.1.0]: https://github.com/choplin/wtm/releases/tag/v0.1.0

[Unreleased]: https://github.com/choplin/wtm/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/choplin/wtm/compare/v0.1.0...v0.1.1
