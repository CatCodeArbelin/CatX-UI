#!/usr/bin/env bash
set -euo pipefail

usage() {
    echo "usage: $0 OUTPUT_DIR COMMIT LABEL [COMMIT LABEL ...]" >&2
    exit 2
}
[[ $# -ge 3 && $((($# - 1) % 2)) -eq 0 ]] || usage
out_dir=$1
shift
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
mkdir -p "$out_dir"

while [[ $# -gt 0 ]]; do
    commit=$1
    label=$2
    shift 2
    work=$(mktemp -d)
    trap 'rm -rf "$work"' RETURN
    mkdir -p "$out_dir/$label"
    git -C "$repo_root" archive "$commit" | tar -x -C "$work"
    # Older CatX snapshots predate the ignored embed directory.  A stub is
    # sufficient for the legacy binary used only as the known-good updater
    # starting point; the current release artifact supplies the real bundle.
    mkdir -p "$work/internal/web/dist"
    touch "$work/internal/web/dist/.gitkeep"
    (cd "$work" && CGO_ENABLED=1 go build -buildvcs=false -o "$out_dir/$label/x-ui" .)
    chmod +x "$out_dir/$label/x-ui"
    rm -rf "$work"
    trap - RETURN
done
