#!/usr/bin/env bash
set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

base_version=$(tr -d '[:space:]' < internal/forkrelease/upstream_version)
base_tag="v${base_version#v}"
compare_tag=""
if [[ "${1:-}" == "--compare" ]]; then
    compare_tag="${2:?usage: $0 [--compare vX.Y.Z]}"
fi

require_ref() {
    git rev-parse --verify --quiet "$1^{commit}" >/dev/null || {
        echo "missing git ref: $1" >&2
        exit 2
    }
}

[[ "$base_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || {
    echo "recorded upstream base is not a stable tag: $base_tag" >&2
    exit 2
}
if [[ -n "$compare_tag" && ! "$compare_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "comparison ref is not a stable tag: $compare_tag" >&2
    exit 2
fi

require_ref "$base_tag"
if [[ -n "$compare_tag" ]]; then require_ref "$compare_tag"; fi

echo "CatX upstream maintenance report"
echo "base: ${base_tag} ($(git rev-parse --short "$base_tag^{commit}"))"
echo "head: $(git branch --show-current) ($(git rev-parse --short HEAD))"
echo "origin: $(git remote get-url origin)"
echo "upstream: $(git remote get-url upstream)"

echo
echo "stable tags available locally:"
git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-version:refname | head -n 10

if [[ -n "$compare_tag" ]]; then
    echo
    echo "upstream delta: ${base_tag}..${compare_tag}"
    mapfile -t files < <(git diff --name-only "$base_tag..$compare_tag")
    sensitive_re='^(main\.go|internal/(database|web/runtime|web/service|sub|xray)/|frontend/src/(routes\.tsx|layouts/|components/command-palette/|forkext/)|internal/fork(ext|recovery|release)/|((update|install)\.sh|x-ui\.sh)|scripts/catx-update-transaction\.sh|\.github/workflows/(release|smoke|fork-verify)\.yml)'
    sensitive=()
    for path in "${files[@]}"; do
        [[ "$path" =~ $sensitive_re ]] && sensitive+=("$path")
    done
    echo "changed paths: ${#files[@]}"
    echo "sensitive paths: ${#sensitive[@]}"
    printf '%s\n' "${sensitive[@]:-}" | sed '/^$/d' | sort -u
    echo
    echo "touchpoint overlap (known CatX boundaries):"
    touch_re='^(main\.go|internal/(database/db\.go|web/web\.go|web/runtime/|web/service/xray\.go|web/controller/api\.go|sub/|xray/)|frontend/src/(routes\.tsx|layouts/AppSidebar\.tsx|components/command-palette/|pages/api-docs/endpoints\.ts)|internal/fork(ext|recovery|release)/|((update|install)\.sh|x-ui\.sh)|scripts/catx-update-transaction\.sh)'
    for path in "${files[@]}"; do
        [[ "$path" =~ $touch_re ]] && echo "$path"
    done | sort -u
fi
