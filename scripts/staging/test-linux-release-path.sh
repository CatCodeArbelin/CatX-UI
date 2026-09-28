#!/usr/bin/env bash
set -euo pipefail

# This is intentionally a Docker harness rather than a Windows approximation.
# The release workflow downloads its own emitted artifacts and runs this file
# on Ubuntu, so install/update behavior is exercised in the same Linux-like
# environment that operators use.

usage() {
    echo "usage: $0 ARTIFACT_DIR LEGACY_DIR" >&2
    exit 2
}
[[ $# -eq 2 && -d "$1" && -d "$2" ]] || usage
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
artifacts=$(cd "$1" && pwd)
legacy=$(cd "$2" && pwd)

command -v docker >/dev/null 2>&1 || {
    echo "linux release-path qualification requires Docker" >&2
    exit 1
}

docker run --rm \
    -v "$repo_root:/repo:ro" \
    -v "$artifacts:/assets:ro" \
    -v "$legacy:/legacy:ro" \
    -e DEBIAN_FRONTEND=noninteractive \
    debian:bookworm-slim bash -euo pipefail -c '
        apt-get update -qq
        apt-get install -y -qq --no-install-recommends curl jq python3 sqlite3 tar > /dev/null || true
        for required in curl jq python3 sqlite3 tar sha256sum; do
            command -v "$required" >/dev/null || {
                echo "required staging tool is unavailable: $required" >&2
                exit 1
            }
        done

        mkdir -p /srv/release/CatCodeArbelin/CatX-UI/releases/download/dev-latest
        cp /assets/* /srv/release/CatCodeArbelin/CatX-UI/releases/download/dev-latest/
        for required_file in /repo/install.sh /repo/update.sh /repo/scripts/catx-update-transaction.sh; do
            test -f "$required_file" || { echo "repository fixture is missing: $required_file" >&2; exit 1; }
        done
        test -f /repo/install.sh
        test -f /repo/update.sh

        # The service manager is deliberately a disposable shim.  The panel
        # itself is started and health-probed below; the shim lets the real
        # updater execute its service lifecycle in a container without PID 1
        # systemd, while preserving every transactional boundary.
        cat > /usr/local/bin/systemctl <<"EOF"
#!/usr/bin/env bash
case "${1:-}" in
  is-active) exit 0 ;;
  *) exit 0 ;;
esac
EOF
        chmod +x /usr/local/bin/systemctl

        # The real installer is exercised below, but its dependency bootstrap
        # is intentionally bypassed after the harness has installed the tools
        # needed for this rehearsal.  Otherwise cron's mail integration pulls
        # a full MTA into the minimal container and fails while configuring
        # service state that the harness does not run.
        cat > /usr/local/bin/apt-get <<"EOF"
#!/usr/bin/env bash
case " ${*:-} " in
  *" install "*) exit 0 ;;
  *) exec /usr/bin/apt-get "$@" ;;
esac
EOF
        chmod +x /usr/local/bin/apt-get

        python3 -m http.server 8765 --directory /srv/release >/tmp/release-server.log 2>&1 &
        server_pid=$!
        trap "kill $server_pid 2>/dev/null || true" EXIT
        base=http://127.0.0.1:8765/CatCodeArbelin/CatX-UI

        wait_http() {
            local url=$1
            for _ in $(seq 1 30); do
                code=$(curl -sS -o /dev/null -w "%{http_code}" "$url" || true)
                case "$code" in 200|301|302|307|308) return 0 ;; esac
                sleep 1
            done
            echo "health probe failed for $url" >&2
            cat /tmp/release-server.log >&2 || true
            return 1
        }

        export CI=true CATX_TEST_RELEASE_MODE=1 CATX_TEST_RELEASE_BASE_URL="$base"
        export XUI_NONINTERACTIVE=1 XUI_SSL_MODE=none XUI_ENABLE_FAIL2BAN=false
        export XUI_MAIN_FOLDER=/usr/local/x-ui XUI_SERVICE=/etc/systemd/system
        export XUI_DB_FOLDER=/etc/x-ui XUI_UPDATE_STATUS_FILE=/etc/x-ui/update-status.json

        # A. Fresh install through the supported installer path.
        bash /repo/install.sh dev-latest < /dev/null
        test -x /usr/local/x-ui/x-ui
        /usr/local/x-ui/x-ui setting -show | grep -q "hasDefaultCredential: false"
        /usr/local/x-ui/x-ui setting -username staging-admin -password staging-password -port 28080 -listenIP 127.0.0.1
        XUI_PORT=28080 /usr/local/x-ui/x-ui run >/tmp/x-ui.log 2>&1 &
        panel_pid=$!
        trap "kill $panel_pid 2>/dev/null || true; kill $server_pid 2>/dev/null || true" EXIT
        wait_http http://127.0.0.1:28080/
        curl -fsS http://127.0.0.1:28080/login >/dev/null
        kill "$panel_pid" 2>/dev/null || true
        wait "$panel_pid" 2>/dev/null || true

        # Preserve a populated, restarted installation before each legacy
        # upgrade.  The settings row and panel-generated DB are synthetic and
        # contain no real credentials or addresses.
        cp -a /etc/x-ui /tmp/populated-db
        sqlite3 /tmp/populated-db/x-ui.db "CREATE TABLE IF NOT EXISTS catx_rc2_fixture (key TEXT PRIMARY KEY, value TEXT NOT NULL); INSERT OR REPLACE INTO catx_rc2_fixture VALUES (\"synthetic-client\",\"fixture-value\");"

        for label in production-baseline pre-rc-develop; do
            rm -rf /usr/local/x-ui /etc/x-ui
            mkdir -p /usr/local/x-ui /etc/x-ui /etc/systemd/system
            cp -a /tmp/populated-db/. /etc/x-ui/
            cp /legacy/$label/x-ui /usr/local/x-ui/x-ui
            cp /assets/catx-ui.sh /usr/local/x-ui/x-ui.sh
            cp /assets/catx-ui-update-lib.sh /usr/local/x-ui/catx-update-lib.sh
            # Keep the supported service layout present for the updater.
            cp /assets/catx-ui-update.sh /usr/local/x-ui/update.sh
            cp /repo/x-ui.service.debian /usr/local/x-ui/x-ui.service.debian
            chmod +x /usr/local/x-ui/x-ui /usr/local/x-ui/x-ui.sh /usr/local/x-ui/update.sh
            legacy_sha=$(sha256sum /usr/local/x-ui/x-ui | awk '{print $1}')
            candidate_sha=$(tar -xOzf /assets/catx-ui-linux-amd64.tar.gz x-ui/x-ui | sha256sum | awk '{print $1}')
            test "$legacy_sha" != "$candidate_sha"

            case "$label" in
                production-baseline) run_id=900719925474099312345 ;;
                pre-rc-develop) run_id=900719925474099312346 ;;
                *) echo "unexpected legacy label: $label" >&2; exit 1 ;;
            esac
            set +e
            XUI_UPDATE_RUN_ID="$run_id" XUI_UPDATE_TAG=dev-latest bash /repo/update.sh >/tmp/update-$label.log 2>&1
            update_rc=$?
            set -e
            if [[ "$update_rc" -ne 0 ]]; then
                cat /tmp/update-$label.log >&2
                exit "$update_rc"
            fi
            jq -e --arg run "$run_id" ".runId == \$run and .state == \"success\" and (.runId | type == \"string\") and .exitCode == 0 and .rolledBack == false and .rollbackHealthy == false" /etc/x-ui/update-status.json >/dev/null
            grep -Fq "product=CatX-UI" <(/usr/local/x-ui/x-ui release-info)
            test "$(sqlite3 /etc/x-ui/x-ui.db "SELECT value FROM catx_rc2_fixture WHERE key = \"synthetic-client\";")" = fixture-value
            /usr/local/x-ui/x-ui migrate >/dev/null
            # Two clean restarts are required after migration/update.
            for restart in 1 2; do
                XUI_PORT=28080 /usr/local/x-ui/x-ui run >/tmp/x-ui-$label-$restart.log 2>&1 &
                panel_pid=$!
                wait_http http://127.0.0.1:28080/
                kill "$panel_pid" 2>/dev/null || true
                wait "$panel_pid" 2>/dev/null || true
            done
        done

        # Corruption is rejected before the known-good installation is
        # replaced.  This is an actual updater invocation, not a unit mock.
        before_version=$(/usr/local/x-ui/x-ui release-info)
        printf corruption >> /srv/release/CatCodeArbelin/CatX-UI/releases/download/dev-latest/catx-ui-linux-amd64.tar.gz
        sha256sum /srv/release/CatCodeArbelin/CatX-UI/releases/download/dev-latest/catx-ui-linux-amd64.tar.gz > /srv/release/CatCodeArbelin/CatX-UI/releases/download/dev-latest/catx-ui-linux-amd64.tar.gz.sha256
        set +e
        XUI_UPDATE_RUN_ID=900719925474099312347 XUI_UPDATE_TAG=dev-latest bash /repo/update.sh >/tmp/update-corrupt.log 2>&1
        corrupt_rc=$?
        set -e
        test "$corrupt_rc" -ne 0
        grep -Eq "checksum|archive|failed" /tmp/update-corrupt.log
        test "$before_version" = "$(/usr/local/x-ui/x-ui release-info)"

        echo "linux release-path qualification: PASS"
    '
