#!/usr/bin/env bash
# Install the wtm binary from GitHub Releases.
# Intended to be called by the Claude Code plugin SessionStart hook.
#
# Environment variables:
#   CLAUDE_PLUGIN_ROOT  – plugin installation directory (set by Claude Code)
#   CLAUDE_PLUGIN_DATA  – persistent data directory (set by Claude Code)
#   WTM_SKIP_INSTALL    – set to "1" to skip installation (e.g. when using Homebrew)

set -euo pipefail

# Skip if the user opted out.
if [[ "${WTM_SKIP_INSTALL:-}" == "1" ]]; then
  exit 0
fi

# Skip if wtm is already on PATH outside the plugin bin/ directory.
if command -v wtm >/dev/null 2>&1; then
  existing="$(command -v wtm)"
  case "$existing" in
    "${CLAUDE_PLUGIN_ROOT}"/bin/*|"${CLAUDE_PLUGIN_DATA}"/bin/*)
      # This is the plugin-managed copy; continue to check version.
      ;;
    *)
      # System-installed (e.g. Homebrew) — nothing to do.
      exit 0
      ;;
  esac
fi

# ---------------------------------------------------------------------------
# Determine the expected version from marketplace.json
# ---------------------------------------------------------------------------
MARKETPLACE="${CLAUDE_PLUGIN_ROOT}/.claude-plugin/marketplace.json"
if [[ ! -f "$MARKETPLACE" ]]; then
  echo "[wtm] marketplace.json not found; skipping binary install." >&2
  exit 0
fi

EXPECTED_VERSION="$(python3 -c "
import json, sys
with open(sys.argv[1]) as f:
    print(json.load(f)['plugins'][0]['version'])
" "$MARKETPLACE" 2>/dev/null || true)"

if [[ -z "$EXPECTED_VERSION" ]]; then
  echo "[wtm] Could not read version from marketplace.json; skipping." >&2
  exit 0
fi

# ---------------------------------------------------------------------------
# Check if the plugin-managed binary already matches the expected version
# ---------------------------------------------------------------------------
BIN_DIR="${CLAUDE_PLUGIN_DATA}/bin"
WTM_BIN="${BIN_DIR}/wtm"

if [[ -x "$WTM_BIN" ]]; then
  CURRENT_VERSION="$("$WTM_BIN" version 2>/dev/null | awk '{print $NF}' || true)"
  if [[ "$CURRENT_VERSION" == "$EXPECTED_VERSION" ]]; then
    exit 0
  fi
fi

# ---------------------------------------------------------------------------
# Resolve platform and architecture
# ---------------------------------------------------------------------------
OS="$(uname -s)"   # Darwin, Linux, …
ARCH="$(uname -m)" # x86_64, arm64, aarch64, …

case "$ARCH" in
  x86_64|amd64) ARCH="x86_64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "[wtm] Unsupported architecture: $ARCH" >&2
    exit 0
    ;;
esac

case "$OS" in
  Darwin|Linux) EXT="tar.gz" ;;
  MINGW*|MSYS*|CYGWIN*|Windows*) OS="Windows"; EXT="zip" ;;
  *)
    echo "[wtm] Unsupported OS: $OS" >&2
    exit 0
    ;;
esac

ASSET_NAME="wtm_${EXPECTED_VERSION}_${OS}_${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/choplin/wtm/releases/download/v${EXPECTED_VERSION}/${ASSET_NAME}"

# ---------------------------------------------------------------------------
# Download and extract
# ---------------------------------------------------------------------------
TMPDIR_DL="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_DL"' EXIT

echo "[wtm] Downloading wtm v${EXPECTED_VERSION} for ${OS}/${ARCH}..." >&2

if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "${TMPDIR_DL}/${ASSET_NAME}" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "${TMPDIR_DL}/${ASSET_NAME}" "$DOWNLOAD_URL"
else
  echo "[wtm] Neither curl nor wget found; cannot download binary." >&2
  exit 0
fi

mkdir -p "$BIN_DIR"

case "$EXT" in
  tar.gz)
    tar -xzf "${TMPDIR_DL}/${ASSET_NAME}" -C "$TMPDIR_DL"
    ;;
  zip)
    unzip -qo "${TMPDIR_DL}/${ASSET_NAME}" -d "$TMPDIR_DL"
    ;;
esac

install -m 755 "${TMPDIR_DL}/wtm" "$WTM_BIN"

echo "[wtm] Installed wtm v${EXPECTED_VERSION} to ${WTM_BIN}" >&2
