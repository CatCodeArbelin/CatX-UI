#!/usr/bin/env bash
set -euo pipefail

# This is intentionally a Docker harness rather than a Windows approximation.
# The release workflow downloads its own emitted artifacts and runs this file
# on Ubuntu, so install/update behavior is exercised in the same Linux-like
# environment that operators use.

usage() {
    echo "usage: $0 ARTIFACT_DIR LEGACY_DIR [RELEASE_TAG]" >&2
    exit 2
}
[[ ($# -eq 2 || $# -eq 3) && -d "$1" && -d "$2" ]] || usage
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
artifacts=$(cd "$1" && pwd)
legacy=$(cd "$2" && pwd)
release_tag=${3:-dev-latest}

command -v docker >/dev/null 2>&1 || {
    echo "linux release-path qualification requires Docker" >&2
    exit 1
}

docker run --rm \
    -v "$repo_root:/repo:ro" \
    -v "$artifacts:/assets:ro" \
    -v "$legacy:/legacy:ro" \
    -e CATX_STAGING_RELEASE_TAG="$release_tag" \
    -e DEBIAN_FRONTEND=noninteractive \
    debian:bookworm-slim bash -euo pipefail -c '
        apt-get update -qq
        apt-get install -y -qq --no-install-recommends curl jq procps python3 sqlite3 tar > /dev/null || true
        for required in curl jq pgrep python3 sqlite3 tar sha256sum; do
            command -v "$required" >/dev/null || {
                echo "required staging tool is unavailable: $required" >&2
                exit 1
            }
        done

        release_tag=${CATX_STAGING_RELEASE_TAG:-dev-latest}
        mkdir -p "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag"
        cp -a /assets/. "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/"
        test -f "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/catx-ui-install.sh" || { find /assets -maxdepth 2 -type f -print >&2; exit 1; }
        test -f "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/catx-ui-update.sh"
        install_script=$(cat "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/catx-ui-install.sh")
        update_script=$(cat "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/catx-ui-update.sh")
        # The service manager is deliberately a disposable shim.  The panel
        # itself is started and health-probed below; the shim lets the real
        # updater execute its service lifecycle in a container without PID 1
        # systemd, while preserving every transactional boundary.
        cat > /usr/local/bin/systemctl <<"EOF"
#!/usr/bin/env bash
set -euo pipefail

live=/usr/local/x-ui
pid_file=/tmp/catx-x-ui-service.pid
service_log=/tmp/catx-x-ui-service.log
evidence_file=${CATX_STAGING_ROLLBACK_EVIDENCE_FILE:-/tmp/catx-staging-rollback.log}

binary_sha() {
  sha256sum "$live/x-ui" | awk '{print $1}'
}

service_running() {
  [[ -s "$pid_file" ]] || return 1
  local pid
  pid=$(cat "$pid_file")
  kill -0 "$pid" 2>/dev/null
}

stop_service() {
  if service_running; then
    local pid
    pid=$(cat "$pid_file")
    kill "$pid" 2>/dev/null || true
    for _ in $(seq 1 30); do
      kill -0 "$pid" 2>/dev/null || break
      sleep 1
    done
    kill -9 "$pid" 2>/dev/null || true
  fi
  rm -f "$pid_file"
}

start_service() {
  local current_sha
  current_sha=$(binary_sha)
  if [[ "${CATX_STAGING_RUNTIME:-0}" == 1 && "${CATX_STAGING_ROLLBACK_INJECT:-0}" == 1 && \
        "$current_sha" == "${CATX_STAGING_ROLLBACK_CANDIDATE_SHA:-}" && \
        ! -e /tmp/catx-staging-candidate-failure-injected ]]; then
    printf 'candidate-activated=%s failure=service-start\n' "$current_sha" >> "$evidence_file"
    : > /tmp/catx-staging-candidate-failure-injected
    printf 'staging: candidate activated before injected service-start failure sha=%s\n' "$current_sha" >> "$evidence_file"
    return 42
  fi
  stop_service
  (cd "$live" && XUI_PORT=28080 ./x-ui run) >> "$service_log" 2>&1 &
  echo "$!" > "$pid_file"
  for _ in $(seq 1 30); do
    if curl -fsS http://127.0.0.1:28080/staging/ >/dev/null 2>&1; then
      return 0
    fi
    if ! service_running; then
      return 1
    fi
    sleep 1
  done
  return 1
}

case "${1:-}" in
  daemon-reload|enable) exit 0 ;;
  start)
    if [[ "${CATX_STAGING_RUNTIME:-0}" != 1 ]]; then exit 0; fi
    start_service
    ;;
  stop)
    if [[ "${CATX_STAGING_RUNTIME:-0}" != 1 ]]; then exit 0; fi
    stop_service
    ;;
  is-active)
    if [[ "${CATX_STAGING_RUNTIME:-0}" != 1 ]]; then exit 0; fi
    service_running && curl -fsS http://127.0.0.1:28080/staging/ >/dev/null 2>&1
    ;;
  *) exit 0 ;;
esac
EOF
        chmod +x /usr/local/bin/systemctl

        # The real installer is exercised below, but its dependency bootstrap
        # is intentionally bypassed after the harness has installed the tools
        # needed for this rehearsal.  Otherwise the cron mail integration pulls
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
            if [[ -f /tmp/x-ui.log ]]; then
                grep -Ei "error|fatal|panic|listen|failed" /tmp/x-ui.log | tail -n 80 >&2 || true
            fi
            return 1
        }

        export CI=true CATX_TEST_RELEASE_MODE=1 CATX_TEST_RELEASE_BASE_URL="$base"
        export XUI_NONINTERACTIVE=1 XUI_SSL_MODE=none XUI_ENABLE_FAIL2BAN=false
        export XUI_USERNAME=staging-admin XUI_PASSWORD=staging-password
        export XUI_WEB_BASE_PATH=staging XUI_PANEL_PORT=28080 XUI_SERVER_IP=127.0.0.1
        export XUI_MAIN_FOLDER=/usr/local/x-ui XUI_SERVICE=/etc/systemd/system
        export XUI_DB_FOLDER=/etc/x-ui XUI_UPDATE_STATUS_FILE=/etc/x-ui/update-status.json

        # A. Fresh install through the supported installer path.
        printf "%s\n" "$install_script" >/tmp/staging-install.sh
        chmod 700 /tmp/staging-install.sh
        set +e
        bash /tmp/staging-install.sh "$release_tag" < /dev/null >/tmp/install.log 2>&1
        install_rc=$?
        set -e
        echo "staging: installer exit code $install_rc" >&2
        set +e
        if [[ "$install_rc" -ne 0 ]]; then
            cat /tmp/install.log >&2 || true
            exit "$install_rc"
        fi
        echo "staging: install completed" >&2
        echo "staging: binary-check exists=$(test -e /usr/local/x-ui/x-ui && echo yes || echo no) executable=$(test -x /usr/local/x-ui/x-ui && echo yes || echo no)" >&2
        set -e
        if [[ ! -e /usr/local/x-ui/x-ui ]]; then
            echo "staging: binary path absent" >&2
            find /usr/local -maxdepth 4 -type f -print >&2 || true
            exit 1
        fi
        if [[ ! -x /usr/local/x-ui/x-ui ]]; then
            echo "staging: binary path is not executable" >&2
            ls -l /usr/local/x-ui/x-ui >&2 || true
            exit 1
        fi
        echo "staging: panel binary present"
        set +e
        setting_output=$(/usr/local/x-ui/x-ui setting -show 2>&1)
        setting_rc=$?
        set -e
        if [[ "$setting_rc" -ne 0 ]]; then
            echo "initial setting inspection failed with exit code $setting_rc" >&2
            echo "$setting_output" >&2
            exit "$setting_rc"
        fi
        default_credential=$(printf "%s\n" "$setting_output" | grep -F "hasDefaultCredential: " | cut -d " " -f2)
        echo "staging: setting inspection rc=$setting_rc defaultCredential=$default_credential"
        if [[ "$default_credential" != "false" ]]; then
            echo "default credential state was not cleared" >&2
            echo "$setting_output" >&2
            exit 1
        fi
        set +e
        setting_apply_output=$(/usr/local/x-ui/x-ui setting -username staging-admin -password staging-password -port 28080 -listenIP 127.0.0.1 2>&1)
        setting_apply_rc=$?
        set -e
        if [[ "$setting_apply_rc" -ne 0 ]]; then
            echo "initial setting update failed with exit code $setting_apply_rc" >&2
            echo "$setting_apply_output" >&2
            exit "$setting_apply_rc"
        fi
        echo "staging: initial panel settings applied"
        (cd /usr/local/x-ui && XUI_PORT=28080 ./x-ui run) >/tmp/x-ui.log 2>&1 &
        panel_pid=$!
        sleep 2
        if ! kill -0 "$panel_pid" 2>/dev/null; then
            echo "staging: panel process exited before health probe" >&2
            tail -n 80 /tmp/x-ui.log >&2 || true
        fi
        trap "kill $panel_pid 2>/dev/null || true; kill $server_pid 2>/dev/null || true" EXIT
        wait_http http://127.0.0.1:28080/staging/
        curl -fsS http://127.0.0.1:28080/staging/ >/dev/null
        kill "$panel_pid" 2>/dev/null || true
        wait "$panel_pid" 2>/dev/null || true

        # Preserve a populated, restarted installation before each legacy
        # upgrade.  The settings row and panel-generated DB are synthetic and
        # contain no real credentials or addresses.
        cp -a /etc/x-ui /tmp/populated-db
        sqlite3 /tmp/populated-db/x-ui.db "CREATE TABLE IF NOT EXISTS catx_rc2_fixture (key TEXT PRIMARY KEY, value TEXT NOT NULL); INSERT OR REPLACE INTO catx_rc2_fixture VALUES (\"synthetic-client\",\"fixture-value\");"

        for label in production-baseline pre-rc-develop rc1-public; do
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
            legacy_sha=$(sha256sum /usr/local/x-ui/x-ui | cut -d " " -f1)
            candidate_sha=$(tar -xOzf /assets/catx-ui-linux-amd64.tar.gz x-ui/x-ui | sha256sum | cut -d " " -f1)
            test "$legacy_sha" != "$candidate_sha"

            case "$label" in
                production-baseline) run_id=900719925474099312345 ;;
                pre-rc-develop) run_id=900719925474099312346 ;;
                rc1-public) run_id=900719925474099312347 ;;
                *) echo "unexpected legacy label: $label" >&2; exit 1 ;;
            esac
            printf "%s\n" "$update_script" >/tmp/staging-update.sh
            chmod 700 /tmp/staging-update.sh
            set +e
            XUI_UPDATE_RUN_ID="$run_id" XUI_UPDATE_TAG="$release_tag" bash /tmp/staging-update.sh >/tmp/update-$label.log 2>&1 < /dev/null
            update_rc=$?
            set -e
            if [[ "$update_rc" -ne 0 ]]; then
                cat /tmp/update-$label.log >&2
                exit "$update_rc"
            fi
            jq -e --arg run "$run_id" ".runId == \$run and .state == \"success\" and (.runId | type == \"string\") and .exitCode == 0 and .rolledBack == false and .rollbackHealthy == false" /etc/x-ui/update-status.json >/dev/null
            grep -Fq "product=CatX-UI" <(/usr/local/x-ui/x-ui release-info)
            if [[ "$release_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-rc\.[1-9][0-9]*$ ]]; then
                grep -Fxq "channel=rc" <(/usr/local/x-ui/x-ui release-info)
                grep -Fxq "release_version=${release_tag#v}" <(/usr/local/x-ui/x-ui release-info)
            elif [[ "$release_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
                grep -Fxq "channel=stable" <(/usr/local/x-ui/x-ui release-info)
            else
                grep -Fxq "channel=dev" <(/usr/local/x-ui/x-ui release-info)
            fi
            test "$(sqlite3 /etc/x-ui/x-ui.db "SELECT value FROM catx_rc2_fixture WHERE key = \"synthetic-client\";")" = fixture-value
            /usr/local/x-ui/x-ui migrate >/dev/null
            # Two clean restarts are required after migration/update.
            for restart in 1 2; do
                (cd /usr/local/x-ui && XUI_PORT=28080 ./x-ui run) >/tmp/x-ui-$label-$restart.log 2>&1 &
                panel_pid=$!
                wait_http http://127.0.0.1:28080/staging/
                kill "$panel_pid" 2>/dev/null || true
                wait "$panel_pid" 2>/dev/null || true
            done
        done

        # Exercise the real post-activation rollback path.  The candidate is
        # the exact archive downloaded by the release workflow; only the
        # disposable systemctl shim injects a failure after the live swap.
        rollback_archive="/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/catx-ui-linux-amd64.tar.gz"
        rollback_run_id=900719925474099312348
        rollback_evidence=/tmp/catx-staging-rollback.log
        rollback_audit=/tmp/catx-staging-update-audit.log
        rm -f "$rollback_evidence" "$rollback_audit" /tmp/catx-staging-candidate-failure-injected
        pkill -f '[/]usr/local/x-ui/bin/xray-linux-' >/dev/null 2>&1 || true
        rm -rf /usr/local/x-ui /etc/x-ui
        mkdir -p /usr/local/x-ui /etc/x-ui /etc/systemd/system
        cp -a /tmp/populated-db/. /etc/x-ui/
        cp /legacy/rc1-public/x-ui /usr/local/x-ui/x-ui
        cp /assets/catx-ui.sh /usr/local/x-ui/x-ui.sh
        cp /assets/catx-ui-update-lib.sh /usr/local/x-ui/catx-update-lib.sh
        cp /assets/catx-ui-update.sh /usr/local/x-ui/update.sh
        cp /repo/x-ui.service.debian /usr/local/x-ui/x-ui.service.debian
        chmod +x /usr/local/x-ui/x-ui /usr/local/x-ui/x-ui.sh /usr/local/x-ui/update.sh
        mkdir -p /usr/local/x-ui/bin
        for runtime_asset in xray-linux-amd64 geoip.dat geosite.dat geoip_IR.dat geosite_IR.dat geoip_RU.dat geosite_RU.dat; do
            tar -xOzf "$rollback_archive" "x-ui/bin/$runtime_asset" > "/usr/local/x-ui/bin/$runtime_asset"
        done
        chmod +x /usr/local/x-ui/bin/xray-linux-amd64

        # These synthetic external files make restoration observable without
        # introducing credentials or relying on a host service manager.
        printf '#!/usr/bin/env bash\n# synthetic known-good CLI\n' > /usr/bin/x-ui
        chmod 755 /usr/bin/x-ui
        printf 'synthetic-known-good-service\n' > /etc/systemd/system/x-ui.service
        printf 'synthetic-known-good-environment\n' > /etc/default/x-ui

        known_good_sha=$(sha256sum /usr/local/x-ui/x-ui | cut -d " " -f1)
        candidate_sha=$(tar -xOzf "$rollback_archive" x-ui/x-ui | sha256sum | cut -d " " -f1)
        candidate_identity=$(mktemp)
        tar -xOzf "$rollback_archive" x-ui/x-ui > "$candidate_identity"
        chmod +x "$candidate_identity"
        candidate_release_info=$("$candidate_identity" release-info)
        rm -f "$candidate_identity"
        known_good_identity=$(/usr/local/x-ui/x-ui release-info)
        db_marker_before=$(sqlite3 /etc/x-ui/x-ui.db "SELECT value FROM catx_rc2_fixture WHERE key = \"synthetic-client\";")
        cli_before=$(sha256sum /usr/bin/x-ui | cut -d " " -f1)
        service_before=$(sha256sum /etc/systemd/system/x-ui.service | cut -d " " -f1)
        environment_before=$(sha256sum /etc/default/x-ui | cut -d " " -f1)
        echo "staging: rollback known-good binary sha=$known_good_sha"
        echo "staging: rollback candidate binary sha=$candidate_sha"
        echo "staging: rollback candidate identity=$(printf '%s' "$candidate_release_info" | tr '\n' ';')"
        echo "staging: rollback known-good identity=$(printf '%s' "$known_good_identity" | tr '\n' ';')"
        echo "staging: rollback database marker before=$db_marker_before"
        test "$known_good_sha" != "$candidate_sha"

        export CATX_STAGING_RUNTIME=1
        export CATX_STAGING_ROLLBACK_INJECT=1
        export CATX_STAGING_ROLLBACK_CANDIDATE_SHA="$candidate_sha"
        export CATX_STAGING_ROLLBACK_EVIDENCE_FILE="$rollback_evidence"
        set +e
        XUI_UPDATE_RUN_ID="$rollback_run_id" XUI_UPDATE_TAG="$release_tag" CATX_UPDATE_AUDIT_FILE="$rollback_audit" bash /tmp/staging-update.sh >/tmp/update-post-activation-rollback.log 2>&1 < /dev/null
        rollback_rc=$?
        set -e
        echo "staging: post-activation rollback updater exit=$rollback_rc"
        cat "$rollback_evidence" >&2
        grep -Fq "candidate-activated=$candidate_sha" "$rollback_evidence"
        grep -Fq "failure=service-start" "$rollback_evidence"
        test "$rollback_rc" -eq 2
        jq -e --arg run "$rollback_run_id" '.runId == $run and .state == "failed" and .exitCode == 2 and .rolledBack == true and .rollbackHealthy == true' /etc/x-ui/update-status.json >/dev/null
        grep -Fq "outcome=failure" "$rollback_audit"
        grep -Fq "outcome=rollback rollback_healthy=1" "$rollback_audit"

        restored_sha=$(sha256sum /usr/local/x-ui/x-ui | cut -d " " -f1)
        restored_cli=$(sha256sum /usr/bin/x-ui | cut -d " " -f1)
        restored_service=$(sha256sum /etc/systemd/system/x-ui.service | cut -d " " -f1)
        restored_environment=$(sha256sum /etc/default/x-ui | cut -d " " -f1)
        db_marker_after=$(sqlite3 /etc/x-ui/x-ui.db "SELECT value FROM catx_rc2_fixture WHERE key = \"synthetic-client\";")
        test "$restored_sha" = "$known_good_sha"
        test "$restored_sha" != "$candidate_sha"
        test "$restored_cli" = "$cli_before"
        test "$restored_service" = "$service_before"
        test "$restored_environment" = "$environment_before"
        test "$db_marker_after" = "$db_marker_before"
        test "$known_good_identity" = "$(/usr/local/x-ui/x-ui release-info)"
        echo "staging: rollback restored binary sha=$restored_sha"
        echo "staging: rollback database marker after=$db_marker_after"
        test -s /tmp/catx-x-ui-service.pid
        service_pid=$(cat /tmp/catx-x-ui-service.pid)
        kill -0 "$service_pid"
        wait_http http://127.0.0.1:28080/staging/
        curl -fsS http://127.0.0.1:28080/staging/ >/dev/null
        xray_pid=$(pgrep -f '[/]usr/local/x-ui/bin/xray-linux-' | head -n 1 || true)
        test -n "$xray_pid"
        kill -0 "$xray_pid"
        grep -Eiq 'xray.*started' /tmp/catx-x-ui-service.log
        echo "staging: rollback restored panel healthy pid=$service_pid"
        echo "staging: rollback restored xray healthy pid=$xray_pid"
        systemctl stop x-ui
        export CATX_STAGING_RUNTIME=0 CATX_STAGING_ROLLBACK_INJECT=0

        # Corruption is rejected before the known-good installation is
        # replaced.  This is an actual updater invocation, not a unit mock.
        before_version=$(/usr/local/x-ui/x-ui release-info)
        printf corruption >> "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/catx-ui-linux-amd64.tar.gz"
        sha256sum "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/catx-ui-linux-amd64.tar.gz" > "/srv/release/CatCodeArbelin/CatX-UI/releases/download/$release_tag/catx-ui-linux-amd64.tar.gz.sha256"
        set +e
        XUI_UPDATE_RUN_ID=900719925474099312347 XUI_UPDATE_TAG="$release_tag" bash /tmp/staging-update.sh >/tmp/update-corrupt.log 2>&1 < /dev/null
        corrupt_rc=$?
        set -e
        test "$corrupt_rc" -ne 0
        grep -Eq "checksum|archive|failed" /tmp/update-corrupt.log
        test "$before_version" = "$(/usr/local/x-ui/x-ui release-info)"

        echo "linux release-path qualification: PASS"
    '
