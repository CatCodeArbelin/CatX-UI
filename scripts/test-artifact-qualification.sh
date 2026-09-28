#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
checker="$repo_root/scripts/inspect-release-artifacts.sh"
source "$repo_root/internal/forkrelease/identity.env"
if command -v python3 >/dev/null 2>&1; then
    python_cmd=python3
else
    python_cmd=python
fi
fixture=$(mktemp -d)
trap 'rm -rf "$fixture"' EXIT

make_control() {
    local name=$1
    case "$name" in
        "${CATX_ASSET_PREFIX}-release-metadata.json")
            cat > "$fixture/$name" <<EOF
{"product":"$CATX_PRODUCT_NAME","repository":"$CATX_RELEASE_OWNER/$CATX_RELEASE_REPOSITORY","forkVersion":"$(tr -d '[:space:]' < "$repo_root/internal/forkrelease/fork_version")","releaseVersion":"dev+fixture-commit","releaseTag":"dev-latest","prerelease":true,"latest":false,"upstreamBaseVersion":"$(tr -d '[:space:]' < "$repo_root/internal/forkrelease/upstream_version")","bundledXrayVersion":"$CATX_XRAY_VERSION","channel":"dev","buildCommit":"fixture-commit","releaseApiUrl":"https://api.github.com/repos/$CATX_RELEASE_OWNER/$CATX_RELEASE_REPOSITORY/releases/1","releaseHtmlUrl":"https://github.com/$CATX_RELEASE_OWNER/$CATX_RELEASE_REPOSITORY/releases/tag/dev-latest"}
EOF
            ;;
        "${CATX_ASSET_PREFIX}-changelog.txt") printf 'fixture-commit release qualification (upstream issue MHSanaei/3x-ui#1)\n' > "$fixture/$name" ;;
        *)
            cat > "$fixture/$name" <<EOF
#!/usr/bin/env bash
readonly CATX_RELEASE_OWNER="$CATX_RELEASE_OWNER"
readonly CATX_RELEASE_REPOSITORY="$CATX_RELEASE_REPOSITORY"
readonly CATX_ASSET_PREFIX="$CATX_ASSET_PREFIX"
readonly CATX_DEV_RELEASE_TAG="$CATX_DEV_RELEASE_TAG"
readonly CATX_RC_VERSION="0.1.0-rc.1"
EOF
            chmod +x "$fixture/$name"
            ;;
    esac
}

for name in \
    "${CATX_ASSET_PREFIX}-update.sh" "${CATX_ASSET_PREFIX}-install.sh" "${CATX_ASSET_PREFIX}.sh" \
    "${CATX_ASSET_PREFIX}-update-lib.sh" "${CATX_ASSET_PREFIX}-release-metadata.json" "${CATX_ASSET_PREFIX}-changelog.txt"; do
    make_control "$name"
done

for arch in amd64 arm64 armv7 armv6 386 armv5 s390x; do
    root=$(mktemp -d)
    mkdir -p "$root/x-ui/bin"
    printf 'CatX-UI fixture binary\n' > "$root/x-ui/x-ui"
    printf '#!/usr/bin/env bash\n' > "$root/x-ui/x-ui.sh"
    for service in debian arch rhel; do printf '[Unit]\nDescription=CatX fixture\n' > "$root/x-ui/x-ui.service.$service"; done
    printf 'xray fixture\n' > "$root/x-ui/bin/xray-linux-$arch"
    tar -czf "$fixture/${CATX_ASSET_PREFIX}-linux-$arch.tar.gz" -C "$root" x-ui
    rm -rf "$root"
done
root=$(mktemp -d)
mkdir -p "$root/x-ui/bin"
printf 'CatX-UI Windows fixture\n' > "$root/x-ui/x-ui.exe"
printf 'Xray Windows fixture\n' > "$root/x-ui/bin/xray-windows-amd64.exe"
if command -v zip >/dev/null 2>&1; then
    (cd "$root" && zip -q -r "$fixture/${CATX_ASSET_PREFIX}-windows-amd64.zip" x-ui)
else
    "$python_cmd" - "$root" "$fixture/${CATX_ASSET_PREFIX}-windows-amd64.zip" <<'PY'
import sys
import zipfile
from pathlib import Path

root, output = sys.argv[1:]
with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED) as archive:
    for path in Path(root, "x-ui").rglob("*"):
        if path.is_file():
            archive.write(path, path.relative_to(root).as_posix())
PY
fi
rm -rf "$root"

for path in "$fixture"/*; do
    case "$path" in *.sha256) continue ;; esac
    sha256sum "$path" > "$path.sha256"
done

CATX_ARTIFACT_TEST_MODE=1 "$checker" "$fixture" fixture-commit

sed -i 's/"releaseVersion":"dev+fixture-commit"/"releaseVersion":"0.1.0-rc.1"/; s/"releaseTag":"dev-latest"/"releaseTag":"v0.1.0-rc.1"/; s/"channel":"dev"/"channel":"rc"/; s#releases/tag/dev-latest#releases/tag/v0.1.0-rc.1#' "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json"
sha256sum "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json" > "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json.sha256"
CATX_ARTIFACT_TEST_MODE=1 "$checker" "$fixture" fixture-commit

expect_failure() {
    local label=$1
    shift
    if "$@" >/dev/null 2>&1; then
        echo "FAIL: negative artifact case passed: $label" >&2
        exit 1
    fi
}

cp "$fixture/${CATX_ASSET_PREFIX}-linux-amd64.tar.gz" "$fixture/original-amd64.tar.gz"
printf 'corruption' >> "$fixture/${CATX_ASSET_PREFIX}-linux-amd64.tar.gz"
expect_failure "modified archive" env CATX_ARTIFACT_TEST_MODE=1 "$checker" "$fixture" fixture-commit
mv "$fixture/original-amd64.tar.gz" "$fixture/${CATX_ASSET_PREFIX}-linux-amd64.tar.gz"

printf '0%.0s' {1..64} > "$fixture/${CATX_ASSET_PREFIX}-linux-amd64.tar.gz.sha256"
printf '  %s-linux-amd64.tar.gz\n' "$CATX_ASSET_PREFIX" >> "$fixture/${CATX_ASSET_PREFIX}-linux-amd64.tar.gz.sha256"
expect_failure "modified checksum" env CATX_ARTIFACT_TEST_MODE=1 "$checker" "$fixture" fixture-commit
sha256sum "$fixture/${CATX_ASSET_PREFIX}-linux-amd64.tar.gz" > "$fixture/${CATX_ASSET_PREFIX}-linux-amd64.tar.gz.sha256"

rm -f "$fixture/${CATX_ASSET_PREFIX}-windows-amd64.zip.sha256"
expect_failure "missing checksum" env CATX_ARTIFACT_TEST_MODE=1 "$checker" "$fixture" fixture-commit
sha256sum "$fixture/${CATX_ASSET_PREFIX}-windows-amd64.zip" > "$fixture/${CATX_ASSET_PREFIX}-windows-amd64.zip.sha256"

mv "$fixture/${CATX_ASSET_PREFIX}-linux-arm64.tar.gz" "$fixture/wrong-name.tar.gz"
expect_failure "wrong asset name" env CATX_ARTIFACT_TEST_MODE=1 "$checker" "$fixture" fixture-commit
mv "$fixture/wrong-name.tar.gz" "$fixture/${CATX_ASSET_PREFIX}-linux-arm64.tar.gz"

sed -i "s#CatCodeArbelin/CatX-UI#MHSanaei/3x-ui#" "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json"
sha256sum "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json" > "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json.sha256"
expect_failure "wrong repository owner" env CATX_ARTIFACT_TEST_MODE=1 "$checker" "$fixture" fixture-commit
sed -i "s#MHSanaei/3x-ui#CatCodeArbelin/CatX-UI#" "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json"
sha256sum "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json" > "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json.sha256"

sed -i 's#https://api.github.com/repos/CatCodeArbelin/CatX-UI/releases/1#https://api.github.com/repos/other-owner/other-repo/releases/1#' "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json"
sha256sum "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json" > "$fixture/${CATX_ASSET_PREFIX}-release-metadata.json.sha256"
expect_failure "wrong release URL" env CATX_ARTIFACT_TEST_MODE=1 "$checker" "$fixture" fixture-commit

echo "artifact negative qualification: PASS"
