#!/usr/bin/env bash
set -u

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
# shellcheck source=catx-update-transaction.sh
source "$repo_root/scripts/catx-update-transaction.sh"

failures=0
assert_eq() {
    local want="$1" got="$2" message="$3"
    if [[ "$want" != "$got" ]]; then
        echo "FAIL: $message: got '$got', want '$want'" >&2
        failures=$((failures + 1))
    fi
}

assert_file() {
    local path="$1" want="$2" message="$3"
    if [[ ! -f "$path" ]]; then
        echo "FAIL: $message: $path is missing" >&2
        failures=$((failures + 1))
        return
    fi
    assert_eq "$want" "$(<"$path")" "$message"
}

arch() { echo amd64; }

archive_root=$(mktemp -d)
mkdir -p "$archive_root/safe/x-ui"
printf 'payload\n' > "$archive_root/safe/x-ui/x-ui"
tar -czf "$archive_root/safe.tar.gz" -C "$archive_root/safe" x-ui
if ! _catx_validate_archive_paths "$archive_root/safe.tar.gz"; then
    echo "FAIL: archive preflight rejected regular files" >&2
    failures=$((failures + 1))
fi
mkdir -p "$archive_root/hardlink/x-ui"
printf 'payload\n' > "$archive_root/hardlink/x-ui/x-ui"
ln "$archive_root/hardlink/x-ui/x-ui" "$archive_root/hardlink/x-ui/hardlink"
tar -czf "$archive_root/hardlink.tar.gz" -C "$archive_root/hardlink" x-ui
if _catx_validate_archive_paths "$archive_root/hardlink.tar.gz"; then
    echo "FAIL: archive preflight accepted a hard link" >&2
    failures=$((failures + 1))
fi
mkdir -p "$archive_root/special/x-ui"
mkfifo "$archive_root/special/x-ui/device-like"
tar -czf "$archive_root/special.tar.gz" -C "$archive_root/special" x-ui
if _catx_validate_archive_paths "$archive_root/special.tar.gz"; then
    echo "FAIL: archive preflight accepted a special file" >&2
    failures=$((failures + 1))
fi
rm -rf "$archive_root"

make_binary() {
    local path="$1" version="$2"
    cat > "$path" <<EOF
#!/usr/bin/env bash
case "\${1:-}" in
    migrate)
        printf 'migrated-by-${version}\n' > "\${XUI_DB_FOLDER}/state"
        [[ "\${CATX_TEST_MIGRATE_FAIL:-0}" == 1 ]] && exit 1
        exit 0
        ;;
    release-info)
        printf 'product=CatX-UI\nrepository=CatCodeArbelin/CatX-UI\nfork_version=0.1.0\nupstream_base_version=3.8.5\nbundled_xray_version=26.9.9\nchannel=stable\nbuild_commit=\n'
        ;;
esac
EOF
    chmod +x "$path"
}

setup_case() {
    CASE_ROOT=$(mktemp -d)
    export XUI_DB_TYPE=sqlite
    export XUI_DB_FOLDER="$CASE_ROOT/etc/x-ui"
    export CATX_CLI_PATH="$CASE_ROOT/usr/bin/x-ui"
    export CATX_ENV_FILE_PATHS="$CASE_ROOT/etc/default/x-ui"
    export CATX_UPDATE_AUDIT_FILE="$CASE_ROOT/var/log/x-ui/update-audit.log"
    export xui_folder="$CASE_ROOT/usr/local/x-ui"
    export xui_service="$CASE_ROOT/etc/systemd/system"
    export release=debian
    export CATX_HEALTHCHECK_ATTEMPTS=1
    unset CATX_TEST_STOP_FAIL CATX_TEST_MIGRATE_FAIL CATX_TEST_START_FAIL CATX_TEST_HEALTH_FAIL CATX_TEST_ROLLBACK_START_FAIL

    mkdir -p "$xui_folder/bin" "$XUI_DB_FOLDER" "$(dirname "$CATX_CLI_PATH")" "$xui_service" "$(dirname "$CATX_ENV_FILE_PATHS")"
    make_binary "$xui_folder/x-ui" old
    printf 'old-install\n' > "$xui_folder/version-state"
    printf '# old menu\n' > "$xui_folder/x-ui.sh"
    printf 'custom\n' > "$xui_folder/bin/geoip_custom.dat"
    printf 'old-db\n' > "$XUI_DB_FOLDER/state"
    printf 'old-cli\n' > "$CATX_CLI_PATH"
    printf 'old-service\n' > "$xui_service/x-ui.service"
    printf 'old-env\n' > "$CATX_ENV_FILE_PATHS"
    catx_transaction_dir=$(mktemp -d "$CASE_ROOT/usr/local/.catx-update.XXXXXX")
    catx_transaction_active=0
    catx_transaction_tag=v0.1.0
    catx_rollback_healthy=0

    CANDIDATE="$catx_transaction_dir/stage/x-ui"
    mkdir -p "$CANDIDATE/bin"
    make_binary "$CANDIDATE/x-ui" new
    printf 'new-install\n' > "$CANDIDATE/version-state"
    printf '# new menu\n' > "$CANDIDATE/x-ui.sh"
    printf 'new-service\n' > "$CANDIDATE/x-ui.service.debian"
}

teardown_case() {
    rm -rf "$CASE_ROOT"
}

_catx_stop_service() {
    [[ "${CATX_TEST_STOP_FAIL:-0}" != 1 ]]
}
_catx_start_service() {
    if [[ -f "$xui_folder/version-state" && "$(<"$xui_folder/version-state")" == new-install && "${CATX_TEST_START_FAIL:-0}" == 1 ]]; then
        return 1
    fi
    if [[ -f "$xui_folder/version-state" && "$(<"$xui_folder/version-state")" == old-install && "${CATX_TEST_ROLLBACK_START_FAIL:-0}" == 1 ]]; then
        return 1
    fi
    return 0
}
_catx_service_healthy() {
    if [[ -f "$xui_folder/version-state" && "$(<"$xui_folder/version-state")" == new-install && "${CATX_TEST_HEALTH_FAIL:-0}" == 1 ]]; then
        return 1
    fi
    return 0
}

assert_rolled_back() {
    local label="$1"
    assert_file "$xui_folder/version-state" old-install "$label restored installation"
    assert_file "$XUI_DB_FOLDER/state" old-db "$label restored database"
    assert_file "$CATX_CLI_PATH" old-cli "$label restored CLI"
    assert_file "$xui_service/x-ui.service" old-service "$label restored service unit"
    assert_file "$CATX_ENV_FILE_PATHS" old-env "$label restored environment"
    assert_eq 1 "$catx_rollback_healthy" "$label rollback health"
}

run_failure_case() {
    local label="$1" failure_var="$2"
    setup_case
    export "$failure_var=1"
    catx_apply_staged_update "$CANDIDATE" v0.1.0
    rc=$?
    assert_eq 2 "$rc" "$label return code"
    assert_rolled_back "$label"
    grep -q 'outcome=rollback ' "$CATX_UPDATE_AUDIT_FILE" || {
        echo "FAIL: $label did not audit rollback" >&2
        failures=$((failures + 1))
    }
    teardown_case
}

run_failure_case migration-failure CATX_TEST_MIGRATE_FAIL
run_failure_case start-failure CATX_TEST_START_FAIL
run_failure_case healthcheck-failure CATX_TEST_HEALTH_FAIL

setup_case
export CATX_TEST_STOP_FAIL=1
catx_apply_staged_update "$CANDIDATE" v0.1.0
rc=$?
assert_eq 1 "$rc" "stop-failure return code"
assert_file "$xui_folder/version-state" old-install "stop-failure preserved installation"
assert_file "$XUI_DB_FOLDER/state" old-db "stop-failure preserved database"
assert_file "$CATX_CLI_PATH" old-cli "stop-failure preserved CLI"
[[ ! -e "$catx_transaction_dir/backup/live" ]] || {
    echo "FAIL: stop-failure moved the live installation" >&2
    failures=$((failures + 1))
}
teardown_case

setup_case
rm -f "$CANDIDATE/x-ui.service.debian"
catx_apply_staged_update "$CANDIDATE" v0.1.0
rc=$?
assert_eq 2 "$rc" "install-failure return code"
assert_rolled_back install-failure
teardown_case

setup_case
catx_apply_staged_update "$CANDIDATE" v0.1.0
rc=$?
assert_eq 0 "$rc" "successful update return code"
assert_file "$xui_folder/version-state" new-install "successful update installed candidate"
assert_file "$XUI_DB_FOLDER/state" migrated-by-new "successful update retained migrated DB"
assert_file "$xui_folder/bin/geoip_custom.dat" custom "successful update retained custom bin asset"
assert_file "$CATX_CLI_PATH" '# new menu' "successful update installed CLI"
[[ ! -e "$catx_transaction_dir" ]] || {
    echo "FAIL: successful update retained transaction directory" >&2
    failures=$((failures + 1))
}
teardown_case

setup_case
export CATX_TEST_MIGRATE_FAIL=1 CATX_TEST_ROLLBACK_START_FAIL=1
catx_apply_staged_update "$CANDIDATE" v0.1.0
rc=$?
assert_eq 3 "$rc" "rollback-health-failure return code"
assert_file "$xui_folder/version-state" old-install "rollback-health-failure restored files before reporting"
assert_file "$XUI_DB_FOLDER/state" old-db "rollback-health-failure restored DB before reporting"
assert_eq 0 "$catx_rollback_healthy" "rollback-health-failure status"
grep -q 'outcome=rollback-failed ' "$CATX_UPDATE_AUDIT_FILE" || {
    echo "FAIL: rollback health failure was not audited" >&2
    failures=$((failures + 1))
}
catx_update_exit_guard 3
assert_file "$xui_folder/version-state" old-install "rollback retry preserved restored installation"
unset CATX_TEST_ROLLBACK_START_FAIL
_catx_rollback
assert_eq 0 "$?" "rollback retry after service recovery"
assert_eq 1 "$catx_rollback_healthy" "rollback retry health"
teardown_case

# PostgreSQL uses a logical custom-format snapshot and a single-transaction
# restore. Fake client tools prove both sides of that recovery contract run
# without exposing the DSN in updater logs.
setup_case
export XUI_DB_TYPE=postgres XUI_DB_DSN='postgres://catx:secret@db.example/catx'
export CATX_PG_LOG="$CASE_ROOT/pg-tools.log"
mock_bin="$CASE_ROOT/mock-bin"
mkdir -p "$mock_bin"
cat > "$mock_bin/pg_dump" <<'EOF'
#!/usr/bin/env bash
for arg in "$@"; do
    case "$arg" in --file=*) dump=${arg#--file=} ;; esac
done
printf 'postgres-snapshot\n' > "$dump"
printf 'dump\n' >> "$CATX_PG_LOG"
EOF
cat > "$mock_bin/pg_restore" <<'EOF'
#!/usr/bin/env bash
printf 'restore %s\n' "$*" >> "$CATX_PG_LOG"
EOF
chmod +x "$mock_bin/pg_dump" "$mock_bin/pg_restore"
old_path=$PATH
export PATH="$mock_bin:$PATH" CATX_TEST_MIGRATE_FAIL=1
catx_apply_staged_update "$CANDIDATE" v0.1.0
rc=$?
export PATH=$old_path
assert_eq 2 "$rc" "postgres migration-failure return code"
assert_file "$xui_folder/version-state" old-install "postgres rollback restored installation"
assert_file "$CATX_CLI_PATH" old-cli "postgres rollback restored CLI"
grep -q '^dump$' "$CATX_PG_LOG" || {
    echo "FAIL: PostgreSQL snapshot was not created" >&2
    failures=$((failures + 1))
}
grep -q '^restore .*--single-transaction' "$CATX_PG_LOG" || {
    echo "FAIL: PostgreSQL rollback did not use a single-transaction restore" >&2
    failures=$((failures + 1))
}
if grep -q 'secret' "$CATX_UPDATE_AUDIT_FILE"; then
    echo "FAIL: PostgreSQL DSN leaked into the update audit" >&2
    failures=$((failures + 1))
fi
teardown_case

if ((failures > 0)); then
    echo "transaction tests: $failures failure(s)" >&2
    exit 1
fi
echo "transaction tests: PASS"
