# wtm Development Guidelines

## Documentation Updates

When adding or changing features, always update both:

1. **CHANGELOG.md** - Add entry under `[Unreleased]` section
2. **README.md** - Update relevant sections (Usage, Installation, etc.)

This ensures documentation stays in sync with the codebase.

## Go Commands

Always run Go-related commands (`go test`, `go build`, `go run`) without overriding the default Go build cache. Do not set `GOCACHE` or other cache-related environment variables unless explicitly instructed.

## Release Procedure

1. Ensure the working tree is clean and all tests pass (`go test ./...`).
2. Run `make release-prep RELEASE_VERSION=X.Y.Z` to update `.claude-plugin/marketplace.json`.
3. Update `CHANGELOG.md` with the new version section, release date, and comparison links.
4. Commit the release preparation with an appropriate conventional commit message (e.g., `docs(changelog): prepare vX.Y.Z release`).
5. Create the annotated release tag matching the new version (e.g., `git tag vX.Y.Z`).
6. Push the main branch and the new tag (`git push origin main` and `git push origin vX.Y.Z`).
