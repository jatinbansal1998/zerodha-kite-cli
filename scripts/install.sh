#!/bin/sh
set -eu

DEFAULT_REPO="jatinbansal1998/zerodha-kite-cli"
REPO="${ZERODHA_REPO:-$DEFAULT_REPO}"
VERSION="${ZERODHA_VERSION:-latest}"
INSTALL_DIR="${ZERODHA_INSTALL_DIR:-}"

die() {
  printf '%s\n' "error: $*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Install zerodha-kite-cli from GitHub Releases.

Usage:
  sh install.sh [--version <tag>] [--install-dir <path>] [--repo <owner/name>]

Options:
  --version      Release tag like v1.2.3 (default: latest)
  --install-dir  Target directory for the binary
  --repo         GitHub repo in owner/name format (default: jatinbansal1998/zerodha-kite-cli)
  -h, --help     Show help

Environment variables:
  ZERODHA_VERSION
  ZERODHA_INSTALL_DIR
  ZERODHA_REPO
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || die "--version requires a value"
      VERSION="$2"
      shift 2
      ;;
    --install-dir)
      [ "$#" -ge 2 ] || die "--install-dir requires a value"
      INSTALL_DIR="$2"
      shift 2
      ;;
    --repo)
      [ "$#" -ge 2 ] || die "--repo requires a value"
      REPO="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

uname_s="$(uname -s 2>/dev/null | tr '[:upper:]' '[:lower:]')"
case "$uname_s" in
  linux*)
    target_os="linux"
    ;;
  darwin*)
    target_os="darwin"
    ;;
  *)
    die "unsupported OS: $uname_s (this installer supports Linux/macOS only)"
    ;;
esac

uname_m="$(uname -m 2>/dev/null | tr '[:upper:]' '[:lower:]')"
case "$uname_m" in
  x86_64|amd64)
    target_arch="amd64"
    ;;
  aarch64|arm64)
    target_arch="arm64"
    ;;
  *)
    die "unsupported CPU architecture: $uname_m"
    ;;
esac

if [ -z "$INSTALL_DIR" ]; then
  if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
fi

asset_name="zerodha_${target_os}_${target_arch}"
if [ "$VERSION" = "latest" ]; then
  download_url="https://github.com/${REPO}/releases/latest/download/${asset_name}"
else
  case "$VERSION" in
    v*)
      ;;
    *)
      VERSION="v${VERSION}"
      ;;
  esac
  download_url="https://github.com/${REPO}/releases/download/${VERSION}/${asset_name}"
fi

if command -v curl >/dev/null 2>&1; then
  downloader="curl"
elif command -v wget >/dev/null 2>&1; then
  downloader="wget"
else
  die "neither curl nor wget found; please install one of them"
fi

tmp_dir="$(mktemp -d 2>/dev/null || mktemp -d -t zerodha-install)"
tmp_file="${tmp_dir}/zerodha"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

mkdir -p "$INSTALL_DIR"

printf '%s\n' "Downloading ${download_url}"
if [ "$downloader" = "curl" ]; then
  curl -fL --retry 3 --retry-delay 1 "$download_url" -o "$tmp_file"
else
  wget -O "$tmp_file" "$download_url"
fi

chmod 0755 "$tmp_file"
dest="${INSTALL_DIR}/zerodha"

if mv "$tmp_file" "$dest" 2>/dev/null; then
  :
else
  cp "$tmp_file" "$dest" || die "failed to write ${dest} (permission denied?)"
  rm -f "$tmp_file"
fi

printf '%s\n' "Installed zerodha to ${dest}"

case ":$PATH:" in
  *":${INSTALL_DIR}:"*)
    ;;
  *)
    printf '%s\n' "Add ${INSTALL_DIR} to PATH to use 'zerodha' from any shell."
    ;;
esac

printf '%s\n' "Verify with: zerodha version"
