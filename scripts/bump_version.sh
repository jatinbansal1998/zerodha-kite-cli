#!/usr/bin/env bash

set -euo pipefail

mode="${1:-patch}"
if [[ "$mode" != "patch" ]]; then
  echo "unsupported bump mode: $mode (expected: patch)" >&2
  exit 1
fi

version_file="internal/buildinfo/version.go"
if [[ ! -f "$version_file" ]]; then
  echo "version file not found: $version_file" >&2
  exit 1
fi

current_version="$(sed -nE 's/^var Version = "(v[0-9]+\.[0-9]+\.[0-9]+)"$/\1/p' "$version_file")"
if [[ -z "$current_version" ]]; then
  echo "failed to parse version from $version_file" >&2
  exit 1
fi

if [[ ! "$current_version" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "invalid semantic version: $current_version" >&2
  exit 1
fi

major="${BASH_REMATCH[1]}"
minor="${BASH_REMATCH[2]}"
patch="${BASH_REMATCH[3]}"

next_patch="$((patch + 1))"
next_version="v${major}.${minor}.${next_patch}"

tmp_file="$(mktemp)"
sed -E "s/^var Version = \"v[0-9]+\.[0-9]+\.[0-9]+\"$/var Version = \"${next_version}\"/" "$version_file" >"$tmp_file"
mv "$tmp_file" "$version_file"

echo "$next_version"
