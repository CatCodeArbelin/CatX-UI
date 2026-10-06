#!/usr/bin/env bash
set -Eeuo pipefail

: "${CATX_HARDENING_CANDIDATE_BINARY:?candidate binary required}"
: "${CATX_HARDENING_CANDIDATE_XRAY_BINARY:?candidate xray required}"
: "${CATX_HARDENING_CANDIDATE_XRAY_ASSET_DIR:?candidate xray assets required}"
: "${CATX_HARDENING_CANDIDATE_SHA:?candidate sha required}"

readonly RUN_DIR="$(mktemp -d "${TMPDIR:-/tmp}/catx-hardening-real.XXXXXX")"
readonly APP_DIR="$RUN_DIR/app"
readonly DB_DIR="$RUN_DIR/db"
readonly LOG_DIR="$RUN_DIR/log"
readonly COOKIE_FILE="$RUN_DIR/cookies"
readonly EVIDENCE_DIR="${CATX_HARDENING_EVIDENCE_DIR:-${RUNNER_TEMP:-/tmp}/catx-hardening-real-evidence}"
readonly PORT=19286
readonly BASE_PATH="/catx-hardening-real/"
readonly BASE_URL="http://127.0.0.1:${PORT}/catx-hardening-real"
readonly USERNAME="catx-hardening-operator"
readonly PASSWORD="catx-hardening-operator-password"
readonly INBOUND_REMARK="CatX hardening disposable inbound"
readonly CLIENT_EMAIL="catx-hardening-client@example.invalid"
readonly UNCONFIGURED_CLIENT_EMAIL="catx-hardening-unconfigured@example.invalid"
readonly POLICY_NAME="CatX hardening policy"
readonly RUN_BROWSER="${CATX_HARDENING_RUN_BROWSER:-1}"
readonly LOGFILE="$RUN_DIR/observer.log"
readonly FEATURE_FLAGS='{"flags":{"analytics.enabled":true,"dns_intelligence.enabled":true,"policies.enabled":true,"traffic_control.enabled":true,"security_anomaly.enabled":true,"audit.enabled":true,"self_service.enabled":true,"fleet_updates.enabled":true,"fleet_updates.mutation.enabled":true,"sponsors.enabled":true}}'
readonly DISABLED_FLAGS='{"flags":{"analytics.enabled":false,"dns_intelligence.enabled":false,"policies.enabled":false,"traffic_control.enabled":false,"security_anomaly.enabled":false,"audit.enabled":false,"self_service.enabled":false,"fleet_updates.enabled":false,"fleet_updates.mutation.enabled":false,"sponsors.enabled":false}}'
readonly FEATURE_KEYS=(
  analytics.enabled
  dns_intelligence.enabled
  policies.enabled
  traffic_control.enabled
  security_anomaly.enabled
  audit.enabled
  self_service.enabled
  fleet_updates.enabled
  fleet_updates.mutation.enabled
  sponsors.enabled
)

PID=""
CSRF=""
: > "$LOGFILE"

log() { printf '%s %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$LOGFILE"; }
fail() { log "FAIL: $*"; exit 1; }

finish() {
  local rc=$?
  set +e
  mkdir -p "$EVIDENCE_DIR"
  sed -E 's/(password|token|cookie|authorization)[=:][^[:space:]]+/\1=<redacted>/Ig' "$LOGFILE" > "$EVIDENCE_DIR/observation.log"
  if [[ -f "$RUN_DIR/panel.log" ]]; then
    sed -E 's/(password|token|cookie|authorization)[=:][^[:space:]]+/\1=<redacted>/Ig' "$RUN_DIR/panel.log" > "$EVIDENCE_DIR/panel.log"
  fi
  if [[ -n "$PID" ]]; then kill "$PID" 2>/dev/null || true; fi
  rm -rf "$RUN_DIR"
  exit "$rc"
}

trap finish EXIT
trap 'rc=$?; log "ERROR line=$LINENO status=$rc"; exit "$rc"' ERR

[[ "$(id -u)" == 0 ]] || fail "must run as root"
for command_name in curl jq grep sed sha256sum pgrep pkill; do
  command -v "$command_name" >/dev/null || fail "missing $command_name"
done
if [[ "$RUN_BROWSER" == 1 ]]; then
  command -v node >/dev/null || fail "missing node for browser qualification"
fi

mkdir -p "$APP_DIR/bin" "$DB_DIR" "$LOG_DIR" "$EVIDENCE_DIR"
cp "$CATX_HARDENING_CANDIDATE_BINARY" "$APP_DIR/x-ui"
cp "$CATX_HARDENING_CANDIDATE_XRAY_BINARY" "$APP_DIR/bin/xray-linux-amd64"
cp "$CATX_HARDENING_CANDIDATE_XRAY_ASSET_DIR/geoip.dat" "$CATX_HARDENING_CANDIDATE_XRAY_ASSET_DIR/geosite.dat" "$APP_DIR/bin/"
chmod +x "$APP_DIR/x-ui" "$APP_DIR/bin/xray-linux-amd64"
[[ "$CATX_HARDENING_CANDIDATE_SHA" =~ ^[0-9a-fA-F]{40}$ ]] || fail "candidate SHA is not a full Git SHA"
log "PASS: exact candidate SHA=$CATX_HARDENING_CANDIDATE_SHA"

export XUI_DB_FOLDER="$DB_DIR" XUI_LOG_FOLDER="$LOG_DIR" XUI_BIN_FOLDER="$APP_DIR/bin" XUI_INIT_WEB_BASE_PATH="$BASE_PATH" XUI_PORT="$PORT"
pushd "$APP_DIR" >/dev/null
./x-ui setting -username "$USERNAME" -password "$PASSWORD" -port "$PORT" -webBasePath "$BASE_PATH" -listenIP 127.0.0.1 > "$RUN_DIR/setting.log" 2>&1
popd >/dev/null

start() {
  pushd "$APP_DIR" >/dev/null
  ./x-ui run >> "$RUN_DIR/panel.log" 2>&1 &
  PID=$!
  popd >/dev/null
}

stop() {
  if [[ -n "$PID" ]]; then kill "$PID" 2>/dev/null || true; fi
  for _ in $(seq 1 30); do
    [[ -z "$PID" ]] || ! kill -0 "$PID" 2>/dev/null && break
    sleep 1
  done
  if [[ -n "$PID" ]]; then kill -9 "$PID" 2>/dev/null || true; fi
  PID=""
  pkill -f "$APP_DIR/bin/xray-linux-amd64" 2>/dev/null || true
}

wait_panel() {
  local code=000
  for _ in $(seq 1 60); do
    code=$(curl --connect-timeout 2 --max-time 5 -so /dev/null -w '%{http_code}' "$BASE_URL/csrf-token" || true)
    [[ "$code" == 200 ]] && return
    sleep 1
  done
  [[ ! -f "$RUN_DIR/panel.log" ]] || tail -n 120 "$RUN_DIR/panel.log"
  fail "panel health HTTP $code"
}

login() {
  rm -f "$COOKIE_FILE"
  CSRF=$(curl --fail -sS -c "$COOKIE_FILE" "$BASE_URL/csrf-token" | jq -er .obj)
  curl --fail -sS -c "$COOKIE_FILE" -b "$COOKIE_FILE" -H "X-CSRF-Token: $CSRF" \
    --data-urlencode "username=$USERNAME" --data-urlencode "password=$PASSWORD" \
    "$BASE_URL/login" | jq -e '.success == true' >/dev/null || fail login
}

get() {
  local path=$1 out=$2 code
  code=$(curl -sS -b "$COOKIE_FILE" -o "$out" -w '%{http_code}' "$BASE_URL/panel/api$path" || true)
  [[ "$code" == 2* ]] || fail "GET $path HTTP $code $(cat "$out")"
  jq -e '.success == true' "$out" >/dev/null || fail "GET $path unsuccessful: $(cat "$out")"
}

expect() {
  local method=$1 path=$2 body=$3 want=$4 out=$5 code
  code=$(curl -sS -b "$COOKIE_FILE" -c "$COOKIE_FILE" -X "$method" \
    -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" --data "$body" \
    -o "$out" -w '%{http_code}' "$BASE_URL/panel/api$path" || true)
  [[ "$code" == $want ]] || fail "$method $path HTTP $code $(cat "$out")"
}

mutate() {
  expect "$1" "$2" "$3" '2*' "$4"
  jq -e '.success == true' "$4" >/dev/null || fail "$1 $2 unsuccessful: $(cat "$4")"
}

health() {
  local label=$1
  get /server/status "$RUN_DIR/status-$label.json"
  jq -e '(.obj.xray.state == "running") or (.obj.xrayState == "running")' "$RUN_DIR/status-$label.json" >/dev/null || fail "Xray not running $label"
}

panel_restart() {
  local label=$1
  mutate POST /setting/restartPanel '{}' "$RUN_DIR/restart-$label.json"
  sleep 4
  wait_panel
  login
  health "$label"
  log "PASS: real restartPanel recovered $label"
}

assert_features() {
  local file=$1 state=$2 enabled=$3 active=$4 restart=$5
  for key in "${FEATURE_KEYS[@]}"; do
    jq -e --arg key "$key" --arg state "$state" --argjson enabled "$enabled" --argjson active "$active" --argjson restart "$restart" \
      'any(.obj.items[]; .key == $key and .state == $state and .enabled == $enabled and .active == $active and .restartRequired == $restart)' "$file" \
      >/dev/null || fail "feature $key state mismatch in $(basename "$file"): $(cat "$file")"
  done
}

start
wait_panel
login
get /fork/settings/features "$RUN_DIR/features-initial.json"
assert_features "$RUN_DIR/features-initial.json" feature_off false false false
get /analytics/status "$RUN_DIR/analytics-off-initial.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true' "$RUN_DIR/analytics-off-initial.json" >/dev/null || fail 'initial analytics state wrong'
get /policies "$RUN_DIR/policies-off-initial.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true and (.obj.items | length) == 0' "$RUN_DIR/policies-off-initial.json" >/dev/null || fail 'initial policy state wrong'
log 'PASS: fresh panel starts with explicit feature-off and empty-safe responses'

# Enable the feature set through the same operator path used by Human Review.
# restartPanel must prepare newly enabled feature-owned schemas before it
# publishes active runtime state; a full process restart is not a prerequisite.
mutate PUT /fork/settings/features "$FEATURE_FLAGS" "$RUN_DIR/features-schema-enabled.json"
get /fork/settings/features "$RUN_DIR/features-schema-enabled-pending.json"
assert_features "$RUN_DIR/features-schema-enabled-pending.json" restart_required true false true
panel_restart runtime-enable
get /fork/settings/features "$RUN_DIR/features-schema-active.json"
assert_features "$RUN_DIR/features-schema-active.json" active true true false
get /policies "$RUN_DIR/policies-runtime-enable.json"
jq -e '.obj.state == "active" and (.obj.items | length) == 0 and .obj.featureDisabled == false' "$RUN_DIR/policies-runtime-enable.json" >/dev/null || fail 'policy schema was not prepared during restartPanel'
get /analytics/traffic "$RUN_DIR/analytics-runtime-enable.json"
jq -e '.obj.state == "active" and (.obj.items | length) == 0 and .obj.featureDisabled == false' "$RUN_DIR/analytics-runtime-enable.json" >/dev/null || fail 'analytics schema was not prepared during restartPanel'
get /fleet-updates/campaigns "$RUN_DIR/fleet-updates-runtime-enable.json"
jq -e '.obj | length == 0' "$RUN_DIR/fleet-updates-runtime-enable.json" >/dev/null || fail 'fleet updates did not expose an empty-safe campaign list'
get /nodes/list "$RUN_DIR/nodes-runtime-enable.json"
jq -e '.obj | length == 0' "$RUN_DIR/nodes-runtime-enable.json" >/dev/null || fail 'fleet inventory did not expose an empty-safe node list'
get '/fork/audit/events?eventType=review.no-such-event&limit=100' "$RUN_DIR/audit-empty.json"
jq -e '.obj.items | length == 0' "$RUN_DIR/audit-empty.json" >/dev/null || fail 'audit zero-event response was not empty-safe'
get '/fork/audit/events?eventType=auth.login&limit=100' "$RUN_DIR/audit-synthetic-login.json"
jq -e 'any(.obj.items[]; .eventType == "auth.login" and .metadata != null)' "$RUN_DIR/audit-synthetic-login.json" >/dev/null || fail 'synthetic admin audit event was not recorded'
log 'PASS: restartPanel prepares enabled schemas, empty audit is HTTP-successful, and admin login is audited'

mutate PUT /fork/settings/features "$DISABLED_FLAGS" "$RUN_DIR/features-off-pending.json"
panel_restart feature-off
get /fork/settings/features "$RUN_DIR/features-off.json"
assert_features "$RUN_DIR/features-off.json" feature_off false false false
get /analytics/status "$RUN_DIR/analytics-off.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true' "$RUN_DIR/analytics-off.json" >/dev/null || fail 'analytics feature-off state wrong'
get /policies "$RUN_DIR/policies-off.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true and (.obj.items | length) == 0' "$RUN_DIR/policies-off.json" >/dev/null || fail 'policy feature-off state wrong'
get /traffic-control/status "$RUN_DIR/traffic-off.json"
jq -e '.obj.state == "disabled" and .obj.platform == "disabled"' "$RUN_DIR/traffic-off.json" >/dev/null || fail 'traffic feature-off state wrong'
expect GET /portal/options '' 409 "$RUN_DIR/portal-off.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true' "$RUN_DIR/portal-off.json" >/dev/null || fail 'portal feature-off state wrong'
log 'PASS: real restartPanel applies feature-off state without generic failures'

mutate PUT /fork/settings/features "$FEATURE_FLAGS" "$RUN_DIR/features-on-pending.json"
get /fork/settings/features "$RUN_DIR/features-on-pending-state.json"
assert_features "$RUN_DIR/features-on-pending-state.json" restart_required true false true
log 'PASS: saved-on versus active-off is reported as restart-required'
panel_restart feature-active
get /fork/settings/features "$RUN_DIR/features-active.json"
assert_features "$RUN_DIR/features-active.json" active true true false
get /analytics/status "$RUN_DIR/analytics-active.json"
jq -e '.obj.state == "active" and .obj.featureDisabled == false' "$RUN_DIR/analytics-active.json" >/dev/null || fail 'analytics active state wrong'
ACTIVITY_TO_MS=$(( $(date +%s) * 1000 ))
ACTIVITY_FROM_MS=$(( ACTIVITY_TO_MS - 604800000 ))
get "/analytics/clients/no-client@example.invalid/activity?from=$ACTIVITY_FROM_MS&to=$ACTIVITY_TO_MS&page=1&pageSize=25" "$RUN_DIR/activity-empty.json"
jq -e '.obj.state == "active" and (.obj.items | length) == 0 and .obj.total == 0' "$RUN_DIR/activity-empty.json" >/dev/null || fail 'analytics empty state wrong'
get /policies/status "$RUN_DIR/policy-status-active.json"
jq -e '.obj.state == "active" and .obj.featureDisabled == false' "$RUN_DIR/policy-status-active.json" >/dev/null || fail 'policy active state wrong'
get /policies "$RUN_DIR/policies-empty.json"
jq -e '.obj.state == "active" and (.obj.items | length) == 0 and .obj.featureDisabled == false' "$RUN_DIR/policies-empty.json" >/dev/null || fail 'policy empty state wrong'
get /traffic-control/status "$RUN_DIR/traffic-active.json"
jq -e '.obj.userAttribution == false' "$RUN_DIR/traffic-active.json" >/dev/null || fail 'traffic attribution truthfulness changed'
get /portal/options "$RUN_DIR/portal-active.json"
jq -e '.obj.state == "active" and .obj.featureDisabled == false' "$RUN_DIR/portal-active.json" >/dev/null || fail 'portal active state wrong'

inbound_body=$(jq -nc --arg remark "$INBOUND_REMARK" '{enable:true,remark:$remark,listen:"127.0.0.1",port:18443,protocol:"vless",expiryTime:0,total:0,settings:{clients:[],decryption:"none",fallbacks:[]},streamSettings:{network:"tcp",security:"none",tcpSettings:{header:{type:"none"}}},sniffing:{enabled:false,destOverride:[]}}')
mutate POST /inbounds/add "$inbound_body" "$RUN_DIR/inbound-add.json"
get /inbounds/options "$RUN_DIR/inbound-options.json"
INBOUND_ID=$(jq -er --arg remark "$INBOUND_REMARK" '.obj[] | select(.remark == $remark) | .id' "$RUN_DIR/inbound-options.json" | head -n 1)
[[ "$INBOUND_ID" =~ ^[0-9]+$ ]] || fail 'disposable inbound id was not returned'
client_body=$(jq -nc --arg email "$CLIENT_EMAIL" --argjson id "$INBOUND_ID" '{client:{email:$email,totalGB:0,expiryTime:0,limitIp:0,limitHwid:0,enable:true},inboundIds:[$id]}')
mutate POST /clients/add "$client_body" "$RUN_DIR/client-add.json"
unconfigured_client_body=$(jq -nc --arg email "$UNCONFIGURED_CLIENT_EMAIL" --argjson id "$INBOUND_ID" '{client:{email:$email,totalGB:0,expiryTime:0,limitIp:0,limitHwid:0,enable:true},inboundIds:[$id]}')
mutate POST /clients/add "$unconfigured_client_body" "$RUN_DIR/client-unconfigured-add.json"
ENCODED_EMAIL=$(jq -nr --arg email "$CLIENT_EMAIL" '$email | @uri')
ENCODED_UNCONFIGURED_EMAIL=$(jq -nr --arg email "$UNCONFIGURED_CLIENT_EMAIL" '$email | @uri')
get "/traffic-control/clients/$ENCODED_EMAIL/policy" "$RUN_DIR/traffic-unconfigured.json"
jq -e '.obj.state == "unconfigured" and .obj.featureDisabled == false' "$RUN_DIR/traffic-unconfigured.json" >/dev/null || fail 'traffic unconfigured state wrong'
get "/traffic-control/clients/$ENCODED_UNCONFIGURED_EMAIL/policy" "$RUN_DIR/traffic-unconfigured-second.json"
jq -e '.obj.state == "unconfigured" and .obj.featureDisabled == false' "$RUN_DIR/traffic-unconfigured-second.json" >/dev/null || fail 'second traffic unconfigured state wrong'
traffic_policy=$(jq -nc '{enabled:true,windowSeconds:3600,quotaBytes:0,activeUploadBps:0,activeDownloadBps:0,throttleUploadBps:0,throttleDownloadBps:0}')
mutate PUT "/traffic-control/clients/$ENCODED_EMAIL/policy" "$traffic_policy" "$RUN_DIR/traffic-policy-save.json"
get "/traffic-control/clients/$ENCODED_EMAIL/policy" "$RUN_DIR/traffic-configured.json"
jq -e '.obj.state == "active" and .obj.featureDisabled == false and .obj.enabled == true' "$RUN_DIR/traffic-configured.json" >/dev/null || fail 'traffic configured state wrong'
panel_restart persistence
get "/traffic-control/clients/$ENCODED_EMAIL/policy" "$RUN_DIR/traffic-persistence.json"
jq -e '.obj.state == "active" and .obj.enabled == true' "$RUN_DIR/traffic-persistence.json" >/dev/null || fail 'traffic policy did not persist across restart'

sponsor_body=$(jq -nc '{id:"catx-hardening-sponsor",enabled:true,name:"CatX hardening sponsor",priority:1,slots:["page"],destinationUrl:"https://example.com/catx-hardening",logoUrl:"",title:{"en-US":"CatX hardening sponsor","ru-RU":"Спонсор проверки CatX"},text:{"en-US":"Disposable qualification record","ru-RU":"Временная запись проверки"}}')
mutate POST /fork/sponsors "$sponsor_body" "$RUN_DIR/sponsor-add.json"
log 'PASS: representative active APIs, empty/unconfigured states, persistence, and sponsor data prepared'

if [[ "$RUN_BROWSER" == 1 ]]; then
  CATX_HARDENING_BASE_URL="$BASE_URL" \
    CATX_HARDENING_USERNAME="$USERNAME" \
    CATX_HARDENING_PASSWORD="$PASSWORD" \
    CATX_HARDENING_CLIENT_EMAIL="$CLIENT_EMAIL" \
    CATX_HARDENING_UNCONFIGURED_CLIENT_EMAIL="$UNCONFIGURED_CLIENT_EMAIL" \
    CATX_HARDENING_POLICY_NAME="$POLICY_NAME" \
    CATX_HARDENING_EVIDENCE_DIR="$EVIDENCE_DIR" \
    node scripts/staging/observe-product-hardening-real-browser.mjs
  log 'PASS: real-panel Playwright browser qualification completed'
else
  log 'SKIP: browser qualification disabled by CATX_HARDENING_RUN_BROWSER'
fi

mutate PUT /fork/settings/features "$DISABLED_FLAGS" "$RUN_DIR/features-final-off-pending.json"
panel_restart final-feature-off
get /fork/settings/features "$RUN_DIR/features-final-off.json"
assert_features "$RUN_DIR/features-final-off.json" feature_off false false false
get "/analytics/clients/$ENCODED_EMAIL/activity?from=$ACTIVITY_FROM_MS&to=$ACTIVITY_TO_MS&page=1&pageSize=25" "$RUN_DIR/activity-final-off.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true and (.obj.items | length) == 0' "$RUN_DIR/activity-final-off.json" >/dev/null || fail 'final analytics feature-off regression failed'
get /policies "$RUN_DIR/policies-final-off.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true and (.obj.items | length) == 0' "$RUN_DIR/policies-final-off.json" >/dev/null || fail 'final policy feature-off regression failed'
expect GET "/traffic-control/clients/$ENCODED_EMAIL/policy" '' 409 "$RUN_DIR/traffic-final-off.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true' "$RUN_DIR/traffic-final-off.json" >/dev/null || fail 'final traffic feature-off regression failed'
expect GET /portal/options '' 409 "$RUN_DIR/portal-final-off.json"
jq -e '.obj.state == "feature_off" and .obj.featureDisabled == true' "$RUN_DIR/portal-final-off.json" >/dev/null || fail 'final portal feature-off regression failed'
log 'PASS: final feature-off regression is explicit and empty-safe'

cp "$RUN_DIR"/*.json "$EVIDENCE_DIR/" 2>/dev/null || true
log 'CATX PRODUCT HARDENING REAL PANEL LIFECYCLE PROVEN'
