#!/usr/bin/env bash
set -Eeuo pipefail

# RC-6 observes the already-published RC-2 payload. It does not build,
# install, update, or publish CatX-UI and uses synthetic data only.
readonly RELEASE_REPOSITORY="CatCodeArbelin/CatX-UI"
readonly RELEASE_TAG="v0.1.0-rc.2"
readonly RELEASE_VERSION="0.1.0-rc.2"
readonly RELEASE_COMMIT="4d8feae2e62db914d3146504340d9b9f802088b2"
readonly ASSET_PREFIX="catx-ui"
readonly PANEL_PORT="${CATX_RC6_PANEL_PORT:-19185}"
readonly SUB_PORT="${CATX_RC6_SUB_PORT:-2096}"
readonly BASE_PATH="/rc6-observe/"
readonly MANAGED_INTERFACE="catxrc6dummy0"
readonly ADMIN_USER="rc6-observer"
readonly ADMIN_PASSWORD="rc6-observer-password"
readonly CLIENT_EMAIL="rc6-linux-client@example.invalid"
readonly CLIENT_GROUP="rc6-linux-group"
readonly INBOUND_PORT="24443"
readonly SOCKS_PORT="11080"

readonly RUN_DIR="$(mktemp -d "${TMPDIR:-/tmp}/catx-rc6-linux.XXXXXX")"
readonly EVIDENCE_DIR="${CATX_RC6_EVIDENCE_DIR:-${RUNNER_TEMP:-/tmp}/catx-rc6-linux-evidence}"
readonly ASSET_DIR="$RUN_DIR/assets"
readonly PAYLOAD_DIR="$RUN_DIR/payload"
readonly DB_DIR="$RUN_DIR/db"
readonly LOG_DIR="$RUN_DIR/log"
readonly COOKIE_FILE="$RUN_DIR/session.cookies"
readonly RUN_LOG="$RUN_DIR/observation.log"
readonly BASE_URL="http://127.0.0.1:${PANEL_PORT}${BASE_PATH%/}"
readonly SUB_URL="http://127.0.0.1:${SUB_PORT}"

PANEL_PID=""
CLIENT_XRAY_PID=""
CSRF_TOKEN=""
INBOUND_ID=""
CLIENT_UUID=""
CLIENT_SUB_ID=""

mkdir -p "$ASSET_DIR" "$PAYLOAD_DIR" "$DB_DIR" "$LOG_DIR" "$EVIDENCE_DIR"

log() { printf '%s %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$RUN_LOG"; }
fail() { log "FAIL: $*"; exit 1; }

require_command() {
    command -v "$1" >/dev/null 2>&1 || fail "required command is unavailable: $1"
}

sanitize_file() {
    local source=$1 target=$2
    [[ -f "$source" ]] || return 0
    sed -E \
        -e 's/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}/<redacted-uuid>/g' \
        -e 's/(password|passwd|token|cookie|authorization)[=:][^[:space:]]+/\1=<redacted>/Ig' \
        "$source" > "$target"
}

finish() {
    local exit_code=$?
    set +e
    sanitize_file "$RUN_LOG" "$EVIDENCE_DIR/observation.log"
    sanitize_file "$RUN_DIR/panel.log" "$EVIDENCE_DIR/panel.log"
    sanitize_file "$RUN_DIR/xray-client.log" "$EVIDENCE_DIR/xray-client.log"
    [[ -z "$CLIENT_XRAY_PID" ]] || kill "$CLIENT_XRAY_PID" 2>/dev/null || true
    [[ -z "$PANEL_PID" ]] || kill "$PANEL_PID" 2>/dev/null || true
    pkill -f "$RUN_DIR/payload/x-ui/bin/xray-linux-amd64" 2>/dev/null || true
    ip link show dev "$MANAGED_INTERFACE" >/dev/null 2>&1 && ip link delete "$MANAGED_INTERFACE" 2>/dev/null || true
    rm -f "$COOKIE_FILE"
    rm -rf "$RUN_DIR"
    exit "$exit_code"
}
trap finish EXIT
trap 'rc=$?; printf "%s ERROR: command failed at shell line %s (status %s)\n" "$(date -u +%FT%TZ)" "$LINENO" "$rc" | tee -a "$RUN_LOG"; exit "$rc"' ERR

[[ "$(id -u)" == 0 ]] || fail "Linux observation must run as root for CAP_NET_ADMIN"
for command in curl jq ip tc nft sha256sum tar pgrep pkill sed grep; do require_command "$command"; done

download_verified_asset() {
    local name=$1
    local url="https://github.com/${RELEASE_REPOSITORY}/releases/download/${RELEASE_TAG}/${name}"
    curl --fail --location --retry 3 --silent --show-error -o "$ASSET_DIR/$name" "$url"
    curl --fail --location --retry 3 --silent --show-error -o "$ASSET_DIR/$name.sha256" "$url.sha256"
    (cd "$ASSET_DIR" && sha256sum -c "$name.sha256") >> "$EVIDENCE_DIR/checksums.txt"
    log "PASS: verified public asset $name"
}

download_verified_asset "${ASSET_PREFIX}-linux-amd64.tar.gz"
download_verified_asset "${ASSET_PREFIX}-release-metadata.json"
jq -e \
    --arg product "CatX-UI" --arg repository "$RELEASE_REPOSITORY" \
    --arg fork "0.1.0" --arg upstream "3.8.5" --arg xray "26.9.9" \
    --arg channel "rc" --arg version "$RELEASE_VERSION" --arg tag "$RELEASE_TAG" \
    --arg commit "$RELEASE_COMMIT" \
    '.product == $product and .repository == $repository and .forkVersion == $fork and
     .upstreamBaseVersion == $upstream and .bundledXrayVersion == $xray and
     .channel == $channel and .releaseVersion == $version and .releaseTag == $tag and
     .prerelease == true and .latest == false and .buildCommit == $commit' \
    "$ASSET_DIR/${ASSET_PREFIX}-release-metadata.json" >/dev/null ||
    fail "public RC-2 metadata identity did not match the qualified release"
cp -f "$ASSET_DIR/${ASSET_PREFIX}-release-metadata.json" "$EVIDENCE_DIR/release-metadata.json"
log "PASS: public RC-2 metadata identity matched $RELEASE_COMMIT"

tar -xzf "$ASSET_DIR/${ASSET_PREFIX}-linux-amd64.tar.gz" -C "$PAYLOAD_DIR"
readonly APP_DIR="$PAYLOAD_DIR/x-ui"
[[ -x "$APP_DIR/x-ui" ]] || fail "Linux archive did not contain executable x-ui"
[[ -x "$APP_DIR/bin/xray-linux-amd64" ]] || fail "Linux archive did not contain Xray 26.9.9"

export XUI_DB_FOLDER="$DB_DIR"
export XUI_LOG_FOLDER="$LOG_DIR"
export XUI_BIN_FOLDER="$APP_DIR/bin"
export XUI_INIT_WEB_BASE_PATH="$BASE_PATH"
export CATX_TRAFFIC_CONTROL_INTERFACES="$MANAGED_INTERFACE"
export XUI_PORT="$PANEL_PORT"

ip link show dev "$MANAGED_INTERFACE" >/dev/null 2>&1 && ip link delete "$MANAGED_INTERFACE"
ip link add "$MANAGED_INTERFACE" type dummy
ip link set dev "$MANAGED_INTERFACE" up
log "PASS: created disposable managed dummy interface $MANAGED_INTERFACE; primary NIC untouched"

pushd "$APP_DIR" >/dev/null
./x-ui setting -username "$ADMIN_USER" -password "$ADMIN_PASSWORD" \
    -port "$PANEL_PORT" -webBasePath "$BASE_PATH" -listenIP 127.0.0.1 > "$RUN_DIR/setting.log" 2>&1
popd >/dev/null

start_panel() {
    pushd "$APP_DIR" >/dev/null
    ./x-ui run >> "$RUN_DIR/panel.log" 2>&1 &
    PANEL_PID=$!
    popd >/dev/null
}

stop_panel() {
    if [[ -n "$PANEL_PID" ]] && kill -0 "$PANEL_PID" 2>/dev/null; then
        kill "$PANEL_PID" 2>/dev/null || true
        for _ in $(seq 1 30); do kill -0 "$PANEL_PID" 2>/dev/null || break; sleep 1; done
        kill -9 "$PANEL_PID" 2>/dev/null || true
    fi
    PANEL_PID=""
    pkill -f "$APP_DIR/bin/xray-linux-amd64" 2>/dev/null || true
}

wait_panel() {
    local code=000
    for _ in $(seq 1 60); do
        code=$(curl --connect-timeout 2 --max-time 5 --silent --output /dev/null --write-out '%{http_code}' "$BASE_URL/csrf-token" || true)
        [[ "$code" == 200 ]] && return 0
        sleep 1
    done
    tail -n 100 "$RUN_DIR/panel.log" >&2 || true
    fail "panel did not become ready (last HTTP status $code)"
}

login() {
    rm -f "$COOKIE_FILE"
    local csrf_response login_response
    csrf_response=$(curl --fail --silent --show-error -c "$COOKIE_FILE" -b "$COOKIE_FILE" "$BASE_URL/csrf-token")
    CSRF_TOKEN=$(jq -er '.obj' <<<"$csrf_response")
    login_response=$(curl --fail --silent --show-error -c "$COOKIE_FILE" -b "$COOKIE_FILE" \
        -H "X-CSRF-Token: $CSRF_TOKEN" --data-urlencode "username=$ADMIN_USER" \
        --data-urlencode "password=$ADMIN_PASSWORD" "$BASE_URL/login")
    jq -e '.success == true' <<<"$login_response" >/dev/null || fail "panel login failed"
}

api_get() {
    local path=$1 target=$2
    curl --fail --silent --show-error -b "$COOKIE_FILE" "$BASE_URL/panel/api$path" > "$target"
    jq -e '.success == true' "$target" >/dev/null || fail "GET $path returned an unsuccessful API envelope"
}

api_mutate() {
    local method=$1 path=$2 body=$3 target=$4
    curl --fail --silent --show-error -b "$COOKIE_FILE" -c "$COOKIE_FILE" -X "$method" \
        -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF_TOKEN" \
        --data "$body" "$BASE_URL/panel/api$path" > "$target"
    jq -e '.success == true' "$target" >/dev/null || {
        log "API $method $path response was unsuccessful"
        jq -c '{success,msg}' "$target" >&2 || true
        fail "$method $path failed"
    }
}

start_panel
wait_panel
login

api_get "/server/status" "$RUN_DIR/status-off.json"
api_get "/fork/settings/features" "$RUN_DIR/features-off.json"
api_get "/traffic-control/capabilities" "$RUN_DIR/capabilities-off.json"
jq -e 'all(.obj.items[]; .enabled == false)' "$RUN_DIR/features-off.json" >/dev/null || fail "feature-off baseline was not default-off"
jq -e '(.obj.xray.state == "running") or (.obj.xrayState == "running")' "$RUN_DIR/status-off.json" >/dev/null || fail "feature-off Xray was not running"
jq -e '.obj.state == "disabled"' "$RUN_DIR/capabilities-off.json" >/dev/null || fail "feature-off Traffic Control was not disabled"
cp -f "$RUN_DIR/features-off.json" "$EVIDENCE_DIR/features-off.json"
cp -f "$RUN_DIR/capabilities-off.json" "$EVIDENCE_DIR/capabilities-off.json"
log "PASS: public RC-2 feature-off Linux baseline reached panel, SQLite, login, and Xray"

inbound_body=$(jq -nc --argjson port "$INBOUND_PORT" '{remark:"rc6-linux-vless",enable:true,listen:"127.0.0.1",port:$port,protocol:"vless",settings:{clients:[],decryption:"none",fallbacks:[]},streamSettings:{network:"tcp",security:"none"},sniffing:{enabled:true,destOverride:["http","tls"]},total:0,expiryTime:0}')
api_mutate POST "/inbounds/add" "$inbound_body" "$RUN_DIR/inbound-add.json"
INBOUND_ID=$(jq -er '.obj.id' "$RUN_DIR/inbound-add.json")

client_body=$(jq -nc --arg email "$CLIENT_EMAIL" --arg group "$CLIENT_GROUP" --argjson inbound "$INBOUND_ID" '{client:{email:$email,enable:true,comment:"RC-6 synthetic Linux observer",group:$group,totalGB:0},inboundIds:[$inbound]}')
api_mutate POST "/clients/add" "$client_body" "$RUN_DIR/client-add.json"
api_get "/clients/get/$CLIENT_EMAIL" "$RUN_DIR/client.json"
CLIENT_UUID=$(jq -er '.obj.client.uuid' "$RUN_DIR/client.json")
CLIENT_SUB_ID=$(jq -er '.obj.client.subId' "$RUN_DIR/client.json")
api_get "/inbounds/get/$INBOUND_ID" "$RUN_DIR/inbound.json"
api_mutate POST "/server/restartXrayService" '{}' "$RUN_DIR/xray-restart.json"
log "PASS: created synthetic VLESS inbound/client and restarted Xray"

for _ in $(seq 1 30); do
    if curl --fail --silent --show-error -o "$RUN_DIR/subscription.txt" "$SUB_URL/sub/$CLIENT_SUB_ID"; then
        [[ -s "$RUN_DIR/subscription.txt" ]] && break
    fi
    sleep 1
done
[[ -s "$RUN_DIR/subscription.txt" ]] || fail "public RC-2 subscription endpoint did not return data"
log "PASS: fetched synthetic client subscription from the public RC-2 panel"

cat > "$RUN_DIR/client-xray.json" <<EOF
{
  "log": {"loglevel": "warning"},
  "inbounds": [{"listen":"127.0.0.1","port":${SOCKS_PORT},"protocol":"socks","settings":{"auth":"noauth","udp":false}}],
  "outbounds": [{"tag":"proxy","protocol":"vless","settings":{"vnext":[{"address":"127.0.0.1","port":${INBOUND_PORT},"users":[{"id":"${CLIENT_UUID}","encryption":"none"}]}]},"streamSettings":{"network":"tcp","security":"none"}}]
}
EOF
"$APP_DIR/bin/xray-linux-amd64" run -c "$RUN_DIR/client-xray.json" > "$RUN_DIR/xray-client.log" 2>&1 &
CLIENT_XRAY_PID=$!
sleep 2
proxy_code=$(curl --fail --silent --show-error --socks5-hostname "127.0.0.1:${SOCKS_PORT}" -o "$RUN_DIR/proxy-response.html" -w '%{http_code}' https://example.com || true)
[[ "$proxy_code" == 200 ]] || { tail -n 80 "$RUN_DIR/xray-client.log" >&2 || true; fail "VLESS proxy probe returned HTTP $proxy_code"; }
log "PASS: synthetic VLESS client reached example.com through Xray (HTTP 200)"

feature_body='{"flags":{"analytics.enabled":true,"dns_intelligence.enabled":true,"policies.enabled":true,"traffic_control.enabled":true,"audit.enabled":true,"self_service.enabled":true,"fleet_updates.enabled":true,"fleet_updates.mutation.enabled":true}}'
api_mutate PUT "/fork/settings/features" "$feature_body" "$RUN_DIR/features-on-save.json"
stop_panel
start_panel
wait_panel
login
api_get "/fork/settings/features" "$RUN_DIR/features-on.json"
jq -e 'all(.obj.items[]; if (.key == "analytics.enabled" or .key == "dns_intelligence.enabled" or .key == "policies.enabled" or .key == "traffic_control.enabled" or .key == "audit.enabled" or .key == "self_service.enabled" or .key == "fleet_updates.enabled" or .key == "fleet_updates.mutation.enabled") then .enabled else true end)' "$RUN_DIR/features-on.json" >/dev/null || fail "enabled feature flags did not survive panel restart"
api_get "/traffic-control/capabilities" "$RUN_DIR/capabilities-on.json"
jq -e '.obj.state == "ready" and .obj.platform == "linux" and .obj.tc == true and .obj.nftables == true and .obj.netAdmin == true and .obj.userAttribution == false' "$RUN_DIR/capabilities-on.json" >/dev/null || {
    cp -f "$RUN_DIR/capabilities-on.json" "$EVIDENCE_DIR/capabilities-on.json"
    fail "Linux Traffic Control capability contract was not ready/unsupported-honest"
}
cp -f "$RUN_DIR/capabilities-on.json" "$EVIDENCE_DIR/capabilities-on.json"
log "PASS: public RC-2 Linux capabilities ready; generic user attribution remained false"

substrate_rule=$(jq -nc --arg iface "$MANAGED_INTERFACE" '{rules:[{nodeKey:"rc6-linux-node",clientKey:"rc6-substrate",interface:$iface,mark:60001,uploadRateBps:1000000,downloadRateBps:1000000,selectors:["127.0.0.1/32"]}]}')
api_mutate POST "/traffic-control/reconcile" "$substrate_rule" "$RUN_DIR/reconcile-apply-1.json"
api_mutate POST "/traffic-control/reconcile" "$substrate_rule" "$RUN_DIR/reconcile-apply-2.json"
tc qdisc show dev "$MANAGED_INTERFACE" > "$EVIDENCE_DIR/tc-after-apply.txt"
nft list table inet catx_traffic_control > "$EVIDENCE_DIR/nft-after-apply.txt"
api_mutate POST "/traffic-control/reconcile" '{"rules":[]}' "$RUN_DIR/reconcile-remove-1.json"
api_mutate POST "/traffic-control/reconcile" '{"rules":[]}' "$RUN_DIR/reconcile-remove-2.json"
tc qdisc show dev "$MANAGED_INTERFACE" > "$EVIDENCE_DIR/tc-after-remove.txt"
if nft list table inet catx_traffic_control > "$EVIDENCE_DIR/nft-after-remove.txt" 2>/dev/null; then
    fail "CatX nft table remained after explicit empty reconcile"
fi
grep -q 'noqueue' "$EVIDENCE_DIR/tc-after-remove.txt" || fail "managed qdisc was not removed"
log "PASS: Linux Traffic Control reconcile/apply/remove was idempotent and cleaned its owned state"

group_body=$(jq -nc '{quotaBytes:1073741824,multiplierPpm:1000000,resetPeriod:"never",resetDay:1}')
api_mutate PUT "/clients/groups/quota/${CLIENT_GROUP}" "$group_body" "$RUN_DIR/group-save.json"
api_get "/clients/groups/quota/${CLIENT_GROUP}" "$RUN_DIR/group-view-before.json"
jq -e '.obj.quotaEnabled == true and .obj.quotaBytes == 1073741824' "$RUN_DIR/group-view-before.json" >/dev/null || fail "group quota configuration was not active"

policy_body=$(jq -nc --arg email "$CLIENT_EMAIL" '{clientEmail:$email,enabled:true,windowSeconds:60,quotaBytes:1,activeUploadBps:1000000,activeDownloadBps:1000000,throttleUploadBps:1000,throttleDownloadBps:1000}')
api_mutate PUT "/traffic-control/clients/${CLIENT_EMAIL}/policy" "$policy_body" "$RUN_DIR/policy-save.json"
api_get "/traffic-control/clients/${CLIENT_EMAIL}/policy" "$RUN_DIR/policy-before.json"
jq -e '.obj.enforcement == "unsupported" and (.obj.enforcementNote | contains("generic Xray users"))' "$RUN_DIR/policy-before.json" >/dev/null || fail "generic per-client enforcement was not reported honestly"

api_get "/clients/traffic/${CLIENT_EMAIL}" "$RUN_DIR/traffic-before.json"
traffic_before=$(jq -er '(.obj.up // 0) + (.obj.down // 0)' "$RUN_DIR/traffic-before.json")
for _ in $(seq 1 3); do
    curl --fail --silent --show-error --socks5-hostname "127.0.0.1:${SOCKS_PORT}" -o /dev/null https://example.com
done
traffic_after=0
for _ in $(seq 1 90); do
    api_get "/clients/traffic/${CLIENT_EMAIL}" "$RUN_DIR/traffic-after.json"
    traffic_after=$(jq -er '(.obj.up // 0) + (.obj.down // 0)' "$RUN_DIR/traffic-after.json")
    if (( traffic_after > traffic_before )); then
        break
    fi
    sleep 1
done
(( traffic_after > traffic_before )) || fail "public RC-2 traffic counters did not advance after VLESS traffic"
api_get "/traffic-control/clients/${CLIENT_EMAIL}/policy" "$RUN_DIR/policy-after-traffic.json"
jq -e '.obj.enforcement == "unsupported" and (.obj.lifecycle == "active" or .obj.lifecycle == "throttled")' "$RUN_DIR/policy-after-traffic.json" >/dev/null || fail "traffic policy lifecycle/enforcement state was inconsistent"

small_group_body=$(jq -nc '{quotaBytes:1,multiplierPpm:1000000,resetPeriod:"never",resetDay:1}')
api_mutate PUT "/clients/groups/quota/${CLIENT_GROUP}" "$small_group_body" "$RUN_DIR/group-deplete.json"
api_get "/clients/groups/quota/${CLIENT_GROUP}" "$RUN_DIR/group-view-after.json"
jq -e '.obj.depleted == true and .obj.usedBytes >= 1' "$RUN_DIR/group-view-after.json" >/dev/null || fail "group quota did not observe authoritative traffic usage"
cp -f "$RUN_DIR/group-view-after.json" "$EVIDENCE_DIR/group-view.json"
cp -f "$RUN_DIR/policy-after-traffic.json" "$EVIDENCE_DIR/policy-view.json"
log "PASS: group quota and fixed-window lifecycle observed authoritative counters; kernel enforcement stayed unsupported without attribution"

# Reapply an explicitly owned substrate rule immediately before the settings
# disable/restart. The following assertion is the lifecycle safety gate.
api_mutate POST "/traffic-control/reconcile" "$substrate_rule" "$RUN_DIR/reconcile-before-disable.json"
api_mutate PUT "/fork/settings/features" '{"flags":{"analytics.enabled":false,"dns_intelligence.enabled":false,"policies.enabled":false,"traffic_control.enabled":false,"audit.enabled":false,"self_service.enabled":false,"fleet_updates.enabled":false,"fleet_updates.mutation.enabled":false}}' "$RUN_DIR/features-off-save.json"
stop_panel
start_panel
wait_panel
login
api_get "/fork/settings/features" "$RUN_DIR/features-final.json"
api_get "/traffic-control/capabilities" "$RUN_DIR/capabilities-final.json"
jq -e 'all(.obj.items[]; .enabled == false)' "$RUN_DIR/features-final.json" >/dev/null || fail "feature disablement did not survive restart"
jq -e '.obj.state == "disabled"' "$RUN_DIR/capabilities-final.json" >/dev/null || fail "Traffic Control did not report disabled after restart"
if nft list table inet catx_traffic_control > "$EVIDENCE_DIR/nft-after-disable.txt" 2>/dev/null; then
    fail "Traffic Control disable left CatX-owned nft state behind"
fi
tc qdisc show dev "$MANAGED_INTERFACE" > "$EVIDENCE_DIR/tc-after-disable.txt"
grep -q 'noqueue' "$EVIDENCE_DIR/tc-after-disable.txt" || fail "Traffic Control disable left the managed qdisc behind"
cp -f "$RUN_DIR/features-final.json" "$EVIDENCE_DIR/features-final.json"
cp -f "$RUN_DIR/capabilities-final.json" "$EVIDENCE_DIR/capabilities-final.json"
log "PASS: feature disablement and restart removed CatX-owned Linux Traffic Control state"
log "PASS: RC-6 public RC-2 Linux observation completed using synthetic disposable data"
