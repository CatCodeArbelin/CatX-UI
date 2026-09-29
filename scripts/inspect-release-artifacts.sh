#!/usr/bin/env bash
set -euo pipefail

# Inspect the exact files emitted by .github/workflows/release.yml.  The
# release job is the source of truth for the matrix; this checker deliberately
# accepts no "close enough" names so a renamed or missing asset fails closed.

usage() {
    echo "usage: $0 ARTIFACT_DIR [EXPECTED_COMMIT]" >&2
    exit 2
}

artifact_dir=${1:-}
expected_commit=${2:-}
[[ -n "$artifact_dir" && -d "$artifact_dir" ]] || usage
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
source "$repo_root/internal/forkrelease/identity.env"
fork_version=$(tr -d '[:space:]' < "$repo_root/internal/forkrelease/fork_version")
rc_version=$(tr -d '[:space:]' < "$repo_root/internal/forkrelease/rc_version")
upstream_version=$(tr -d '[:space:]' < "$repo_root/internal/forkrelease/upstream_version")
if command -v python3 >/dev/null 2>&1; then
    python_cmd=python3
elif command -v python >/dev/null 2>&1; then
    python_cmd=python
else
    python_cmd=
fi

fail() {
    echo "artifact qualification: FAIL: $*" >&2
    exit 1
}

require_file() {
    local name=$1
    local path="$artifact_dir/$name"
    [[ -f "$path" ]] || fail "missing asset $name"
    [[ -s "$path" ]] || fail "empty asset $name"
}

check_checksum() {
    local name=$1
    local sum="$name.sha256"
    require_file "$name"
    require_file "$sum"
    (cd "$artifact_dir" && sha256sum -c "$sum" >/dev/null) || fail "checksum rejected for $name"
}

check_metadata() {
    local name="${CATX_ASSET_PREFIX}-release-metadata.json"
    require_file "$name"
    local path="$artifact_dir/$name"
    [[ -n "$python_cmd" ]] || fail "python3 or python is required for metadata validation"
    "$python_cmd" - "$path" "$CATX_PRODUCT_NAME" "$CATX_RELEASE_OWNER/$CATX_RELEASE_REPOSITORY" \
        "$fork_version" "$rc_version" "$upstream_version" "$CATX_XRAY_VERSION" "$expected_commit" <<'PY' || fail "release metadata is invalid"
import json
import sys

path, product, repository, fork, rc, upstream, xray, expected_commit = sys.argv[1:]
with open(path, encoding="utf-8") as handle:
    data = json.load(handle)
assert data["product"] == product
assert data["repository"] == repository
assert data["forkVersion"] == fork
assert data["releaseVersion"]
assert data["releaseTag"]
assert data["upstreamBaseVersion"] == upstream
assert data["bundledXrayVersion"] == xray
assert isinstance(data.get("buildCommit"), str) and data["buildCommit"]
assert isinstance(data.get("releaseApiUrl"), str)
assert isinstance(data.get("releaseHtmlUrl"), str)
assert data["releaseApiUrl"].startswith("https://api.github.com/repos/" + repository + "/releases/")
assert data["releaseHtmlUrl"] == "https://github.com/" + repository + "/releases/tag/" + data["releaseTag"]
assert data["releaseHtmlUrl"].startswith("https://github.com/" + repository + "/releases/")
channel = data["channel"]
if channel == "stable":
    assert data["releaseVersion"] == fork
    assert data["releaseTag"] == "v" + fork
    assert data["prerelease"] is False
    assert data["latest"] is True
elif channel == "rc":
    assert data["releaseVersion"] == rc
    assert data["releaseTag"] == "v" + rc
    assert data["prerelease"] is True
    assert data["latest"] is False
elif channel == "dev":
    assert data["releaseTag"] == "dev-latest"
    assert data["releaseVersion"].startswith("dev+")
    assert data["prerelease"] is True
    assert data["latest"] is False
else:
    raise AssertionError("unknown release channel")
if expected_commit:
    assert data["buildCommit"] == expected_commit
PY
}

reject_sensitive_or_foreign_text() {
    local path=$1
    if grep -aEiq 'MHSanaei/3x-ui|mhsanaei/3x-ui|C:\\Users\\|/home/runner/work/' "$path"; then
        fail "foreign repository or local runner path found in $(basename "$path")"
    fi
}

check_tar_archive() {
    local name=$1 arch=$2
    check_checksum "$name"
    local archive="$artifact_dir/$name"
    local listing
    listing=$(tar -tzf "$archive") || fail "cannot list $name"
    while IFS= read -r entry; do
        [[ -z "$entry" ]] && continue
        [[ "$entry" == x-ui || "$entry" == x-ui/* ]] || fail "$name contains unsafe path $entry"
        [[ "$entry" != /* && "$entry" != *"../"* && "$entry" != *"/.."* ]] || fail "$name contains traversal path $entry"
    done <<< "$listing"
    local root
    root=$(mktemp -d)
    trap 'rm -rf "$root"' RETURN
    # Preserve executable bits for the binary check, but never restore archive
    # ownership into the qualification workspace.
    tar -xzf "$archive" --no-same-owner -C "$root" || fail "cannot extract $name"
    [[ -s "$root/x-ui/x-ui" ]] || fail "$name has no panel binary"
    [[ -x "$root/x-ui/x-ui" || "${CATX_ARTIFACT_TEST_MODE:-0}" == 1 ]] || fail "$name panel binary is not executable"
    [[ -s "$root/x-ui/x-ui.sh" ]] || fail "$name has no menu script"
    [[ -s "$root/x-ui/x-ui.service.debian" && -s "$root/x-ui/x-ui.service.arch" && -s "$root/x-ui/x-ui.service.rhel" ]] \
        || fail "$name is missing one or more service units"
    [[ -s "$root/x-ui/bin/xray-linux-$arch" ]] || fail "$name has no xray-linux-$arch"
    if [[ "${CATX_ARTIFACT_TEST_MODE:-0}" != 1 ]]; then
        command -v file >/dev/null 2>&1 || fail "file is required for binary architecture validation"
        local file_type
        file_type=$(file -b "$root/x-ui/x-ui")
        case "$arch" in
            amd64) grep -Eqi 'x86-64|x86_64' <<< "$file_type" || fail "$name panel binary is not amd64" ;;
            arm64) grep -Eqi 'aarch64|arm64' <<< "$file_type" || fail "$name panel binary is not arm64" ;;
            armv7|armv6|armv5) grep -Eqi 'ARM' <<< "$file_type" || fail "$name panel binary is not ARM" ;;
            386) grep -Eqi '80386|i386' <<< "$file_type" || fail "$name panel binary is not 386" ;;
            s390x) grep -Eqi 's/390|s390' <<< "$file_type" || fail "$name panel binary is not s390x" ;;
        esac
    fi
    rm -rf "$root"
    trap - RETURN
}

check_zip_archive() {
    local name=$1
    check_checksum "$name"
    local archive="$artifact_dir/$name"
    local listing
    if command -v unzip >/dev/null 2>&1; then
        listing=$(unzip -Z1 "$archive") || fail "cannot list $name"
    else
        [[ -n "$python_cmd" ]] || fail "unzip or python is required for $name"
        listing=$("$python_cmd" - "$archive" <<'PY'
import sys
import zipfile
with zipfile.ZipFile(sys.argv[1]) as archive:
    print("\n".join(archive.namelist()))
PY
        ) || fail "cannot list $name"
    fi
    while IFS= read -r entry; do
        [[ -z "$entry" ]] && continue
        [[ "$entry" == x-ui || "$entry" == x-ui/* ]] || fail "$name contains unsafe path $entry"
        [[ "$entry" != /* && "$entry" != *"../"* && "$entry" != *"/.."* ]] || fail "$name contains traversal path $entry"
    done <<< "$listing"
    local root
    root=$(mktemp -d)
    if command -v unzip >/dev/null 2>&1; then
        unzip -q "$archive" -d "$root" || fail "cannot extract $name"
    else
        "$python_cmd" - "$archive" "$root" <<'PY' || fail "cannot extract $name"
import sys
import zipfile
with zipfile.ZipFile(sys.argv[1]) as archive:
    archive.extractall(sys.argv[2])
PY
    fi
    [[ -s "$root/x-ui/x-ui.exe" ]] || fail "$name has no Windows panel binary"
    [[ -s "$root/x-ui/bin/xray-windows-amd64.exe" ]] || fail "$name has no Windows Xray binary"
    rm -rf "$root"
}

for arch in amd64 arm64 armv7 armv6 386 armv5 s390x; do
    check_tar_archive "${CATX_ASSET_PREFIX}-linux-${arch}.tar.gz" "$arch"
done
check_zip_archive "${CATX_ASSET_PREFIX}-windows-amd64.zip"

for name in \
    "${CATX_ASSET_PREFIX}-update.sh" \
    "${CATX_ASSET_PREFIX}-install.sh" \
    "${CATX_ASSET_PREFIX}.sh" \
    "${CATX_ASSET_PREFIX}-update-lib.sh" \
    "${CATX_ASSET_PREFIX}-release-metadata.json"; do
    check_checksum "$name"
    reject_sensitive_or_foreign_text "$artifact_dir/$name"
done
check_checksum "${CATX_ASSET_PREFIX}-changelog.txt"

for name in "${CATX_ASSET_PREFIX}-update.sh" "${CATX_ASSET_PREFIX}-install.sh" "${CATX_ASSET_PREFIX}.sh" "${CATX_ASSET_PREFIX}-update-lib.sh"; do
    path="$artifact_dir/$name"
    grep -Fq "CATX_RC_VERSION=\"${rc_version}\"" "$path" || fail "$name has wrong RC version"
    grep -Fq "CATX_RELEASE_OWNER=\"$CATX_RELEASE_OWNER\"" "$path" || fail "$name has wrong release owner"
    grep -Fq "CATX_RELEASE_REPOSITORY=\"$CATX_RELEASE_REPOSITORY\"" "$path" || fail "$name has wrong release repository"
    grep -Fq "CATX_ASSET_PREFIX=\"$CATX_ASSET_PREFIX\"" "$path" || fail "$name has wrong asset prefix"
    grep -Fq "CATX_DEV_RELEASE_TAG=\"$CATX_DEV_RELEASE_TAG\"" "$path" || fail "$name has wrong dev tag"
done
check_metadata

echo "artifact qualification: PASS (${CATX_ASSET_PREFIX}, fork ${fork_version}, upstream ${upstream_version})"
