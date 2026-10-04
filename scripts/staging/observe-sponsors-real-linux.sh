#!/usr/bin/env bash
set -Eeuo pipefail

: "${CATX_SPONSORS_CANDIDATE_BINARY:?candidate binary required}"
: "${CATX_SPONSORS_CANDIDATE_XRAY_BINARY:?candidate xray required}"
: "${CATX_SPONSORS_CANDIDATE_XRAY_ASSET_DIR:?candidate xray assets required}"
: "${CATX_SPONSORS_CANDIDATE_SHA:?candidate sha required}"
readonly RUN_DIR="$(mktemp -d "${TMPDIR:-/tmp}/catx-sponsors-real.XXXXXX")"
readonly APP_DIR="$RUN_DIR/app"
readonly EVIDENCE_DIR="${CATX_SPONSORS_EVIDENCE_DIR:-${RUNNER_TEMP:-/tmp}/catx-sponsors-real-evidence}"
readonly DB_DIR="$RUN_DIR/db"
readonly LOG_DIR="$RUN_DIR/log"
readonly COOKIE_FILE="$RUN_DIR/cookies"
readonly PORT=19285
readonly BASE_PATH="/sponsors-real/"
readonly BASE_URL="http://127.0.0.1:${PORT}/sponsors-real"
readonly USERNAME="catx-sponsors-operator" PASSWORD="catx-sponsors-operator-password"
readonly ID="catx-real-smoke" DEST="https://example.com/catx-synthetic-destination"
readonly LOGO="https://example.com/catx-synthetic-logo.png"
readonly REMOTE="https://example.com/catx-synthetic-sponsors.json"
readonly LOGFILE="$RUN_DIR/observer.log"
PID="" CSRF=""
: > "$LOGFILE"

log() { printf '%s %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$LOGFILE"; }
fail() { log "FAIL: $*"; exit 1; }
finish() { local rc=$?; set +e; mkdir -p "$EVIDENCE_DIR"; sed -E 's/(password|token|cookie|authorization)[=:][^[:space:]]+/\1=<redacted>/Ig' "$LOGFILE" > "$EVIDENCE_DIR/observation.log"; [[ ! -f "$RUN_DIR/panel.log" ]] || sed -E 's/(password|token|cookie|authorization)[=:][^[:space:]]+/\1=<redacted>/Ig' "$RUN_DIR/panel.log" > "$EVIDENCE_DIR/panel.log"; [[ -z "$PID" ]] || kill "$PID" 2>/dev/null || true; rm -rf "$RUN_DIR"; exit "$rc"; }
trap finish EXIT
trap 'rc=$?; log "ERROR line=$LINENO status=$rc"; exit "$rc"' ERR

[[ "$(id -u)" == 0 ]] || fail "must run as root"
for c in curl jq grep sed sha256sum pgrep pkill; do command -v "$c" >/dev/null || fail "missing $c"; done
mkdir -p "$APP_DIR/bin" "$DB_DIR" "$LOG_DIR" "$EVIDENCE_DIR"
cp "$CATX_SPONSORS_CANDIDATE_BINARY" "$APP_DIR/x-ui"
cp "$CATX_SPONSORS_CANDIDATE_XRAY_BINARY" "$APP_DIR/bin/xray-linux-amd64"
cp "$CATX_SPONSORS_CANDIDATE_XRAY_ASSET_DIR/geoip.dat" "$CATX_SPONSORS_CANDIDATE_XRAY_ASSET_DIR/geosite.dat" "$APP_DIR/bin/"
chmod +x "$APP_DIR/x-ui" "$APP_DIR/bin/xray-linux-amd64"
log "PASS: exact candidate SHA=$CATX_SPONSORS_CANDIDATE_SHA"
export XUI_DB_FOLDER="$DB_DIR" XUI_LOG_FOLDER="$LOG_DIR" XUI_BIN_FOLDER="$APP_DIR/bin" XUI_INIT_WEB_BASE_PATH="$BASE_PATH" XUI_PORT="$PORT"
pushd "$APP_DIR" >/dev/null
./x-ui setting -username "$USERNAME" -password "$PASSWORD" -port "$PORT" -webBasePath "$BASE_PATH" -listenIP 127.0.0.1 > "$RUN_DIR/setting.log" 2>&1
popd >/dev/null

start() { pushd "$APP_DIR" >/dev/null; ./x-ui run >> "$RUN_DIR/panel.log" 2>&1 & PID=$!; popd >/dev/null; }
stop() { [[ -z "$PID" ]] || kill "$PID" 2>/dev/null || true; for _ in $(seq 1 30); do [[ -z "$PID" ]] || ! kill -0 "$PID" 2>/dev/null && break; sleep 1; done; [[ -z "$PID" ]] || kill -9 "$PID" 2>/dev/null || true; PID=""; pkill -f "$APP_DIR/bin/xray-linux-amd64" 2>/dev/null || true; }
wait_panel() { local code=000; for _ in $(seq 1 60); do code=$(curl --connect-timeout 2 --max-time 5 -so /dev/null -w '%{http_code}' "$BASE_URL/csrf-token" || true); [[ "$code" == 200 ]] && return; sleep 1; done; tail -n 100 "$RUN_DIR/panel.log"; fail "panel health HTTP $code"; }
login() { rm -f "$COOKIE_FILE"; CSRF=$(curl --fail -sS -c "$COOKIE_FILE" "$BASE_URL/csrf-token" | jq -er .obj); curl --fail -sS -c "$COOKIE_FILE" -b "$COOKIE_FILE" -H "X-CSRF-Token: $CSRF" --data-urlencode "username=$USERNAME" --data-urlencode "password=$PASSWORD" "$BASE_URL/login" | jq -e '.success == true' >/dev/null || fail login; }
get() { local path=$1 out=$2 code; code=$(curl -sS -b "$COOKIE_FILE" -o "$out" -w '%{http_code}' "$BASE_URL/panel/api$path" || true); [[ "$code" == 2* ]] || fail "GET $path HTTP $code $(cat "$out")"; jq -e '.success == true' "$out" >/dev/null || fail "GET $path unsuccessful"; }
expect() { local method=$1 path=$2 body=$3 want=$4 out=$5 code; code=$(curl -sS -b "$COOKIE_FILE" -c "$COOKIE_FILE" -X "$method" -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" --data "$body" -o "$out" -w '%{http_code}' "$BASE_URL/panel/api$path" || true); [[ "$code" == $want ]] || fail "$method $path HTTP $code $(cat "$out")"; }
mutate() { expect "$1" "$2" "$3" '2*' "$4"; jq -e '.success == true' "$4" >/dev/null || fail "$1 $2 unsuccessful"; }
health() { local label=$1; local check_sponsors=${2:-true}; get /server/status "$RUN_DIR/status-$label.json"; jq -e '(.obj.xray.state == "running") or (.obj.xrayState == "running")' "$RUN_DIR/status-$label.json" >/dev/null || fail "Xray not running $label"; if [[ "$check_sponsors" == true ]]; then curl --fail -sS -b "$COOKIE_FILE" "$BASE_URL/panel/catx/sponsors" -o "$RUN_DIR/page-$label.html"; fi; get /fork/sponsors/status "$RUN_DIR/sponsors-status-$label.json"; ! grep -Fq sponsors.sanaei.dev "$RUN_DIR/panel.log" || fail "Sanaei request marker"; }
panel_restart() { local label=$1; local check_sponsors=${2:-true}; mutate POST /setting/restartPanel '{}' "$RUN_DIR/restart-$label.json"; sleep 4; grep -Fq 'Web server restarted successfully.' "$RUN_DIR/panel.log" || fail "web restart $label"; grep -Fq 'Sub server restarted successfully.' "$RUN_DIR/panel.log" || fail "sub restart $label"; wait_panel; login; health "$label" "$check_sponsors"; log "PASS: real restartPanel recovered $label"; }
startup_restart() { stop; start; wait_panel; login; health audit-schema-startup false; log 'PASS: setup-only startup prepared Audit schema'; }

start; wait_panel; login
get /fork/settings/features "$RUN_DIR/features-initial.json"
jq -e 'any(.obj.items[]; .key == "sponsors.enabled" and .enabled == false)' "$RUN_DIR/features-initial.json" >/dev/null || fail 'Sponsors not initially disabled'
get /fork/sponsors/status "$RUN_DIR/status-initial.json"
jq -e '.obj.enabled == false and .obj.providerMode == "local"' "$RUN_DIR/status-initial.json" >/dev/null || fail 'initial Sponsors state wrong'
log 'PASS: initial Sponsors disabled/LOCAL'

# Audit must be enabled before startup migration; this process restart is only
# schema setup. All lifecycle transitions below use the real restartPanel API.
mutate PUT /fork/settings/features '{"flags":{"audit.enabled":true}}' "$RUN_DIR/audit-enabled.json"
startup_restart
mutate PUT /fork/settings/features '{"flags":{"sponsors.enabled":true}}' "$RUN_DIR/sponsors-enabled.json"
mutate PUT /fork/sponsors/settings '{"providerMode":"local","sourceUrl":"","contactUrl":""}' "$RUN_DIR/provider-local.json"
panel_restart sponsors-enabled
jq -e '.obj.enabled == true and .obj.providerMode == "local"' "$RUN_DIR/sponsors-status-sponsors-enabled.json" >/dev/null || fail 'Sponsors did not enable LOCAL'

body=$(jq -nc --arg id "$ID" --arg d "$DEST" --arg l "$LOGO" '{id:$id,enabled:true,name:"CatX real smoke",priority:7,slots:["dashboard","sidebar","page"],destinationUrl:$d,logoUrl:$l,title:{"en-US":"CatX smoke title"},text:{"en-US":"Synthetic sponsor"}}')
mutate POST /fork/sponsors "$body" "$RUN_DIR/create.json"
get /fork/sponsors/manage "$RUN_DIR/manage-create.json"; get "/fork/sponsors/$ID" "$RUN_DIR/read.json"; get /fork/sponsors "$RUN_DIR/public-create.json"
jq -e --arg id "$ID" 'any(.obj[]?; .id == $id)' "$RUN_DIR/manage-create.json" >/dev/null || fail 'created Sponsor missing from management list'
jq -e --arg id "$ID" 'any(.obj.sponsors[]?; .id == $id)' "$RUN_DIR/public-create.json" >/dev/null || fail 'created Sponsor missing from public list'
updated=$(jq '.name="CatX real smoke updated" | .priority=3' <<< "$body")
mutate PUT "/fork/sponsors/$ID" "$updated" "$RUN_DIR/update.json"
jq -e '.obj.name == "CatX real smoke updated" and .obj.priority == 3' "$RUN_DIR/update.json" >/dev/null || fail 'UPDATE did not persist'
disabled=$(jq '.enabled=false' <<< "$updated")
mutate PUT "/fork/sponsors/$ID" "$disabled" "$RUN_DIR/disable.json"; get /fork/sponsors "$RUN_DIR/public-disabled-record.json"
jq -e --arg id "$ID" 'all(.obj.sponsors[]?; .id != $id)' "$RUN_DIR/public-disabled-record.json" >/dev/null || fail 'disabled Sponsor rendered'
mutate PUT "/fork/sponsors/$ID" "$updated" "$RUN_DIR/enable.json"; get /fork/sponsors "$RUN_DIR/public-enabled-record.json"
jq -e --arg id "$ID" 'any(.obj.sponsors[]?; .id == $id)' "$RUN_DIR/public-enabled-record.json" >/dev/null || fail 're-enabled Sponsor missing'
log 'PASS: authenticated CRUD, enable/disable, slots, and rendering'

panel_restart local-persistence; get /fork/sponsors/status "$RUN_DIR/status-local.json"
jq -e '.obj.enabled == true and .obj.providerMode == "local" and .obj.localSponsorCount == 1' "$RUN_DIR/status-local.json" >/dev/null || fail 'local persistence failed'
get /fork/sponsors "$RUN_DIR/public-local.json"; jq -e --arg id "$ID" 'any(.obj.sponsors[]?; .id == $id)' "$RUN_DIR/public-local.json" >/dev/null || fail 'local Sponsor lost after restart'
mutate PUT /fork/settings/features '{"flags":{"sponsors.enabled":false}}' "$RUN_DIR/sponsors-disabled.json"
panel_restart feature-off false; get /fork/sponsors/status "$RUN_DIR/status-off.json"
jq -e '.obj.enabled == false and .obj.providerMode == "local" and .obj.localSponsorCount == 1 and .obj.cacheState == "disabled"' "$RUN_DIR/status-off.json" >/dev/null || fail 'feature-off status wrong'
expect GET /fork/sponsors '' 200 "$RUN_DIR/public-off.json"; expect GET "/fork/sponsors/logo/$ID" '' 404 "$RUN_DIR/logo-off.json"; expect GET /fork/sponsors/manage '' 503 "$RUN_DIR/manage-off.json"
jq -e '.success == true and (.obj.sponsors | length == 0)' "$RUN_DIR/public-off.json" >/dev/null || fail 'feature-off public list was not empty'
log 'PASS: feature-off retained storage and disabled metadata/logo/rendering'
mutate PUT /fork/settings/features '{"flags":{"sponsors.enabled":true}}' "$RUN_DIR/sponsors-reenabled.json"
panel_restart feature-on-again; get /fork/sponsors "$RUN_DIR/public-reenabled.json"; jq -e --arg id "$ID" 'any(.obj.sponsors[]?; .id == $id)' "$RUN_DIR/public-reenabled.json" >/dev/null || fail 're-enabled Sponsor missing'

mutate PUT /fork/sponsors/settings "$(jq -nc --arg s "$REMOTE" '{providerMode:"remote",sourceUrl:$s,contactUrl:""}')" "$RUN_DIR/provider-remote.json"
panel_restart remote-provider false; get /fork/sponsors/status "$RUN_DIR/status-remote.json"
jq -e '.obj.enabled == true and .obj.providerMode == "remote" and .obj.remoteProviderConfigured == true' "$RUN_DIR/status-remote.json" >/dev/null || fail 'REMOTE provider did not persist'
expect GET /fork/sponsors '' 503 "$RUN_DIR/public-remote.json"; ! grep -Fq "$ID" "$RUN_DIR/public-remote.json" || fail 'REMOTE returned mixed local data'
mutate PUT /fork/sponsors/settings '{"providerMode":"local","sourceUrl":"","contactUrl":""}' "$RUN_DIR/provider-local-again.json"
panel_restart local-provider-again; get /fork/sponsors "$RUN_DIR/public-local-again.json"; jq -e --arg id "$ID" 'any(.obj.sponsors[]?; .id == $id)' "$RUN_DIR/public-local-again.json" >/dev/null || fail 'LOCAL Sponsor missing after REMOTE'
log 'PASS: LOCAL -> REMOTE -> LOCAL authority/persistence'

mutate DELETE "/fork/sponsors/$ID" '' "$RUN_DIR/delete.json"; get /fork/sponsors/manage "$RUN_DIR/manage-delete.json"
jq -e --arg id "$ID" 'all(.obj[]?; .id != $id)' "$RUN_DIR/manage-delete.json" >/dev/null || fail 'DELETE did not remove Sponsor'
get '/fork/audit/events?limit=100' "$RUN_DIR/audit.json"
jq -e '([.obj.items[].eventType] | index("sponsors.create") and index("sponsors.update") and index("sponsors.disable") and index("sponsors.enable") and index("sponsors.delete")) != null' "$RUN_DIR/audit.json" >/dev/null || fail 'audit mutation events missing'
! grep -Fq sponsors.sanaei.dev "$RUN_DIR/panel.log" || fail 'sponsors.sanaei.dev request marker'
cp "$RUN_DIR"/*.json "$EVIDENCE_DIR/" 2>/dev/null || true
log 'PASS: audit events and no sponsors.sanaei.dev request'
log 'SPONSORS REAL PANEL LIFECYCLE PROVEN'
