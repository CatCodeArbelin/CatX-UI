#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_root"

# shellcheck source=/dev/null
source internal/forkrelease/identity.env
expected_slug="${CATX_RELEASE_OWNER}/${CATX_RELEASE_REPOSITORY}"

for file in update.sh install.sh x-ui.sh scripts/catx-update-transaction.sh; do
    grep -Fq "CATX_RELEASE_OWNER=\"${CATX_RELEASE_OWNER}\"" "$file"
    grep -Fq "CATX_RELEASE_REPOSITORY=\"${CATX_RELEASE_REPOSITORY}\"" "$file"
    grep -Fq "CATX_ASSET_PREFIX=\"${CATX_ASSET_PREFIX}\"" "$file"
    grep -Fq "CATX_DEV_RELEASE_TAG=\"${CATX_DEV_RELEASE_TAG}\"" "$file"
done

if grep -En 'https?://(api\.github\.com/repos/|github\.com/|raw\.githubusercontent\.com/)[Mm][Hh][Ss]anaei/3x-ui' \
    update.sh install.sh x-ui.sh deploy/cloud-init/cloud-init.yaml frontend/src/layouts/AppSidebar.tsx \
    internal/web/service/panel/panel.go .github/workflows/release.yml; then
    echo "release/update path still targets the official upstream repository" >&2
    exit 1
fi

for asset_expr in \
    '${CATX_ASSET_PREFIX}-update.sh' \
    '${CATX_ASSET_PREFIX}-install.sh' \
    '${CATX_ASSET_PREFIX}.sh' \
    '${CATX_ASSET_PREFIX}-update-lib.sh'; do
    grep -Fq "$asset_expr" .github/workflows/release.yml || {
        echo "release workflow does not publish $asset_expr" >&2
        exit 1
    }
done

grep -Fq 'repository=${CATX_RELEASE_SLUG}' scripts/catx-update-transaction.sh
grep -Fq 'catx_transactional_update "$@"' update.sh
grep -Fq 'requires a same-release checksum sidecar' install.sh
grep -Fq 'rollback-failed' scripts/catx-update-transaction.sh
grep -Fq 'validate-release-identity:' .github/workflows/release.yml
grep -Fq 'needs: validate-release-identity' .github/workflows/release.yml
grep -Fq '"$recorded_name" == "$asset"' deploy/cloud-init/cloud-init.yaml

echo "release identity tests: PASS"
