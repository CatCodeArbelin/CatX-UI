#!/usr/bin/env bash
set -Eeuo pipefail

# RC-6 observes published RC assets by default, or a locally supplied exact-
# SHA candidate. It does not build, install, update, or publish CatX-UI and
# uses synthetic data only.
readonly RELEASE_REPOSITORY="CatCodeArbelin/CatX-UI"
readonly RELEASE_TAG="${CATX_RC6_PUBLIC_TAG:-v0.1.0-rc.2}"
readonly RELEASE_VERSION="${CATX_RC6_PUBLIC_VERSION:-0.1.0-rc.2}"
readonly RELEASE_COMMIT="${CATX_RC6_PUBLIC_COMMIT:-4d8feae2e62db914d3146504340d9b9f802088b2}"
readonly STABLE_RESTART_REGRESSION_ONLY="${CATX_STABLE_RESTART_REGRESSION_ONLY:-0}"
readonly CANDIDATE_BINARY="${CATX_RC6_CANDIDATE_BINARY:-}"
readonly CANDIDATE_XRAY_BINARY="${CATX_RC6_CANDIDATE_XRAY_BINARY:-}"
readonly CANDIDATE_XRAY_ASSET_DIR="${CATX_RC6_CANDIDATE_XRAY_ASSET_DIR:-}"
readonly CANDIDATE_COMMIT="${CATX_RC6_CANDIDATE_COMMIT:-}"
readonly CANDIDATE_BINARY_SHA256="${CATX_RC6_CANDIDATE_SHA256:-}"
readonly ASSET_PREFIX="catx-ui"
readonly PANEL_PORT="${CATX_RC6_PANEL_PORT:-19185}"
readonly SUB_PORT="${CATX_RC6_SUB_PORT:-2096}"
readonly BASE_PATH="/rc6-observe/"
readonly MANAGED_INTERFACE="rc6dummy0"
readonly ADMIN_INTERFACE="rc6adm0"
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
    ip link show dev "$ADMIN_INTERFACE" >/dev/null 2>&1 && ip link delete "$ADMIN_INTERFACE" 2>/dev/null || true
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

if [[ -n "$CANDIDATE_BINARY" ]]; then
    [[ -x "$CANDIDATE_BINARY" ]] || fail "candidate x-ui binary is unavailable"
    [[ -x "$CANDIDATE_XRAY_BINARY" ]] || fail "candidate Xray binary is unavailable"
    [[ -n "$CANDIDATE_XRAY_ASSET_DIR" ]] || fail "candidate Xray asset directory is unavailable"
    [[ -f "$CANDIDATE_XRAY_ASSET_DIR/geoip.dat" ]] || fail "candidate geoip.dat is unavailable"
    [[ -f "$CANDIDATE_XRAY_ASSET_DIR/geosite.dat" ]] || fail "candidate geosite.dat is unavailable"
    [[ -n "$CANDIDATE_COMMIT" && -n "$CANDIDATE_BINARY_SHA256" ]] || fail "candidate identity inputs are incomplete"
    actual_candidate_sha256=$(sha256sum "$CANDIDATE_BINARY" | awk '{print $1}')
    [[ "$actual_candidate_sha256" == "$CANDIDATE_BINARY_SHA256" ]] || fail "candidate binary checksum did not match the expected exact-SHA artifact"
    mkdir -p "$PAYLOAD_DIR/x-ui/bin"
    cp -f "$CANDIDATE_BINARY" "$PAYLOAD_DIR/x-ui/x-ui"
    cp -f "$CANDIDATE_XRAY_BINARY" "$PAYLOAD_DIR/x-ui/bin/xray-linux-amd64"
    cp -f "$CANDIDATE_XRAY_ASSET_DIR/geoip.dat" "$CANDIDATE_XRAY_ASSET_DIR/geosite.dat" "$PAYLOAD_DIR/x-ui/bin/"
    chmod +x "$PAYLOAD_DIR/x-ui/x-ui" "$PAYLOAD_DIR/x-ui/bin/xray-linux-amd64"
    "$PAYLOAD_DIR/x-ui/x-ui" release-info > "$EVIDENCE_DIR/candidate-release-info.txt"
    grep -Fq "build_commit=$CANDIDATE_COMMIT" "$EVIDENCE_DIR/candidate-release-info.txt" || fail "candidate release identity did not contain the expected exact commit"
    log "PASS: exact hosted candidate identity matched commit=$CANDIDATE_COMMIT binary_sha256=$actual_candidate_sha256"
else
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
        fail "public ${RELEASE_VERSION} metadata identity did not match the qualified release"
    cp -f "$ASSET_DIR/${ASSET_PREFIX}-release-metadata.json" "$EVIDENCE_DIR/release-metadata.json"
    log "PASS: public ${RELEASE_VERSION} metadata identity matched $RELEASE_COMMIT"
    tar -xzf "$ASSET_DIR/${ASSET_PREFIX}-linux-amd64.tar.gz" -C "$PAYLOAD_DIR"
fi
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
    CLIENT_XRAY_PID=""
}

start_client_xray() {
    "$APP_DIR/bin/xray-linux-amd64" run -c "$RUN_DIR/client-xray.json" > "$RUN_DIR/xray-client.log" 2>&1 &
    CLIENT_XRAY_PID=$!
    sleep 2
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
    local path=$1 target=$2 status summary
    status=$(curl --silent --show-error -b "$COOKIE_FILE" -o "$target" -w '%{http_code}' \
        "$BASE_URL/panel/api$path" || true)
    if [[ "$status" != 2* ]]; then
        summary=$(jq -c '{success,msg}' "$target" 2>/dev/null || printf '%s' '<unparseable response>')
        log "API GET $path returned HTTP $status response=$summary"
        fail "GET $path failed"
    fi
    jq -e '.success == true' "$target" >/dev/null || fail "GET $path returned an unsuccessful API envelope"
}

api_mutate() {
    local method=$1 path=$2 body=$3 target=$4 status summary
    status=$(curl --silent --show-error -b "$COOKIE_FILE" -c "$COOKIE_FILE" -X "$method" \
        -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF_TOKEN" \
        --data "$body" -o "$target" -w '%{http_code}' "$BASE_URL/panel/api$path" || true)
    if [[ "$status" != 2* ]]; then
        summary=$(jq -c '{success,msg}' "$target" 2>/dev/null || printf '%s' '<unparseable response>')
        log "API $method $path returned HTTP $status response=$summary"
        fail "$method $path failed"
    fi
    jq -e '.success == true' "$target" >/dev/null || {
        log "API $method $path response was unsuccessful"
        jq -c '{success,msg}' "$target" >&2 || true
        fail "$method $path failed"
    }
}

restart_panel_via_api() {
    local label=$1
    local previous_pid="$PANEL_PID"
    local csrf_code=000

    api_mutate POST "/setting/restartPanel" '{}' "$RUN_DIR/restart-${label}.json"
    cp -f "$RUN_DIR/restart-${label}.json" "$EVIDENCE_DIR/restart-${label}.json"
    log "PASS: POST /setting/restartPanel accepted for ${label}"

    # The endpoint schedules the real SIGHUP path three seconds later. Do not
    # accept the old listener as proof that the restart completed. The release
    # logger records the two completed server restarts, but does not always
    # retain the preceding signal-receipt line.
    sleep 4
    grep -Fq "Web server restarted successfully." "$RUN_DIR/panel.log" ||
        fail "panel log did not record web-server completion for ${label} restart"
    grep -Fq "Sub server restarted successfully." "$RUN_DIR/panel.log" ||
        fail "panel log did not record sub-server completion for ${label} restart"

    for _ in $(seq 1 60); do
        kill -0 "$previous_pid" 2>/dev/null || fail "panel process exited during ${label} restart"
        csrf_code=$(curl --connect-timeout 2 --max-time 5 --silent --output /dev/null \
            --write-out '%{http_code}' "$BASE_URL/csrf-token" || true)
        if [[ "$csrf_code" == 200 ]]; then
            login
            [[ "$PANEL_PID" == "$previous_pid" ]] || fail "${label} restart changed the panel process unexpectedly"
            log "PASS: panel recovered through the same process after real ${label} restart"
            return 0
        fi
        sleep 1
    done
    fail "panel did not recover after real ${label} restart (last HTTP status $csrf_code)"
}

assert_restart_surfaces() {
    local phase=$1 expected=$2
    log "INFO: checking ${phase} restart server health"
    api_get "/server/status" "$RUN_DIR/status-${phase}.json"
    cp -f "$RUN_DIR/status-${phase}.json" "$EVIDENCE_DIR/"
    jq -e '(.obj.xray.state == "running") or (.obj.xrayState == "running")' \
        "$RUN_DIR/status-${phase}.json" >/dev/null || fail "Xray was not healthy after ${phase} restart"

    log "INFO: checking ${phase} persisted CatX feature flags"
    api_get "/fork/settings/features" "$RUN_DIR/features-${phase}.json"
    cp -f "$RUN_DIR/features-${phase}.json" "$EVIDENCE_DIR/"
    if [[ "$expected" == enabled ]]; then
        jq -e '([.obj.items[] | select(.key == "analytics.enabled" or .key == "dns_intelligence.enabled") | .enabled] | sort) == [true, true]' \
            "$RUN_DIR/features-${phase}.json" >/dev/null || fail "CatX enabled flags did not survive ${phase} restart"
    else
        jq -e '([.obj.items[] | select(.key == "analytics.enabled" or .key == "dns_intelligence.enabled") | .enabled] | sort) == [false, false]' \
            "$RUN_DIR/features-${phase}.json" >/dev/null || fail "CatX disabled flags did not survive ${phase} restart"
    fi

    log "INFO: checking ${phase} analytics and DNS Intelligence APIs"
    api_get "/analytics/status" "$RUN_DIR/analytics-status-${phase}.json"
    cp -f "$RUN_DIR/analytics-status-${phase}.json" "$EVIDENCE_DIR/"
    api_get "/analytics/settings" "$RUN_DIR/analytics-settings-${phase}.json"
    cp -f "$RUN_DIR/analytics-settings-${phase}.json" "$EVIDENCE_DIR/"
    api_get "/analytics/clients/${CLIENT_EMAIL}/activity" "$RUN_DIR/activity-${phase}.json"
    cp -f "$RUN_DIR/activity-${phase}.json" "$EVIDENCE_DIR/"
    api_get "/analytics/clients/${CLIENT_EMAIL}/dns" "$RUN_DIR/dns-${phase}.json"
    cp -f "$RUN_DIR/dns-${phase}.json" "$EVIDENCE_DIR/"
    log "INFO: checking ${phase} Activity page"
    activity_page_code=$(curl --connect-timeout 2 --max-time 5 --silent --show-error \
        -b "$COOKIE_FILE" -o "$RUN_DIR/activity-page-${phase}.html" -w '%{http_code}' \
        "$BASE_URL/activity" || true)
    [[ "$activity_page_code" == 200 ]] || fail "Activity page returned HTTP ${activity_page_code} after ${phase} restart"

    if [[ "$expected" == enabled ]]; then
        jq -e '.obj.enabled == true and .obj.dnsIntelligence == true' "$RUN_DIR/analytics-status-${phase}.json" >/dev/null ||
            fail "Analytics/DNS Intelligence status was not enabled after ${phase} restart"
        jq -e '.obj.enabled == true and .obj.dnsIntelligence == true' "$RUN_DIR/analytics-settings-${phase}.json" >/dev/null ||
            fail "Analytics/DNS Intelligence settings were not enabled after ${phase} restart"
        jq -e '.obj.enabled == true' "$RUN_DIR/activity-${phase}.json" >/dev/null ||
            fail "Activity API was not operational after ${phase} restart"
        jq -e '.obj.enabled == true' "$RUN_DIR/dns-${phase}.json" >/dev/null ||
            fail "DNS Intelligence API was not operational after ${phase} restart"
        log "PASS: Activity and DNS Intelligence surfaces were operational after ${phase} restart"
    else
        jq -e '.obj.enabled == false and .obj.dnsIntelligence == false' "$RUN_DIR/analytics-status-${phase}.json" >/dev/null ||
            fail "feature-off analytics status was not explicit after ${phase} restart"
        jq -e '.obj.enabled == false and .obj.dnsIntelligence == false' "$RUN_DIR/analytics-settings-${phase}.json" >/dev/null ||
            fail "feature-off analytics settings were not explicit after ${phase} restart"
        jq -e '.obj.enabled == false' "$RUN_DIR/activity-${phase}.json" >/dev/null ||
            fail "feature-off Activity API was not explicit after ${phase} restart"
        jq -e '.obj.enabled == false' "$RUN_DIR/dns-${phase}.json" >/dev/null ||
            fail "feature-off DNS API was not explicit after ${phase} restart"
        log "PASS: feature-off Activity and DNS Intelligence behavior was explicit and healthy after ${phase} restart"
    fi
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
api_mutate POST "/setting/defaultSettings" '{}' "$RUN_DIR/settings-defaults.json"
SUB_URI=$(jq -er '.obj.subURI' "$RUN_DIR/settings-defaults.json")
SUB_HOST=$(sed -E 's#^(https?://[^/]+).*$#\1#' <<<"$SUB_URI")
SUB_PATH_LENGTH=$((${#SUB_URI} - ${#SUB_HOST}))
log "INFO: resolved subscription endpoint host=$SUB_HOST path_length=$SUB_PATH_LENGTH"
log "PASS: created synthetic VLESS inbound/client and restarted Xray"

subscription_code=000
for _ in $(seq 1 30); do
    subscription_code=$(curl --connect-timeout 2 --max-time 5 --silent --show-error -o "$RUN_DIR/subscription.txt" -w '%{http_code}' "${SUB_URI%/}/$CLIENT_SUB_ID" || true)
    if [[ "$subscription_code" == 200 ]] && [[ -s "$RUN_DIR/subscription.txt" ]]; then
        break
    fi
    sleep 1
done
[[ -s "$RUN_DIR/subscription.txt" ]] || fail "public RC-2 subscription endpoint did not return data (HTTP $subscription_code)"
log "PASS: fetched synthetic client subscription from the public RC-2 panel"

cat > "$RUN_DIR/client-xray.json" <<EOF
{
  "log": {"loglevel": "warning"},
  "inbounds": [{"listen":"127.0.0.1","port":${SOCKS_PORT},"protocol":"socks","settings":{"auth":"noauth","udp":false}}],
  "outbounds": [{"tag":"proxy","protocol":"vless","settings":{"vnext":[{"address":"127.0.0.1","port":${INBOUND_PORT},"users":[{"id":"${CLIENT_UUID}","encryption":"none"}]}]},"streamSettings":{"network":"tcp","security":"none"}}]
}
EOF
start_client_xray
proxy_code=$(curl --fail --silent --show-error --socks5-hostname "127.0.0.1:${SOCKS_PORT}" -o "$RUN_DIR/proxy-response.html" -w '%{http_code}' https://example.com || true)
[[ "$proxy_code" == 200 ]] || { tail -n 80 "$RUN_DIR/xray-client.log" >&2 || true; fail "VLESS proxy probe returned HTTP $proxy_code"; }
log "PASS: synthetic VLESS client reached example.com through Xray (HTTP 200)"

feature_body='{"flags":{"analytics.enabled":true,"dns_intelligence.enabled":true,"policies.enabled":true,"traffic_control.enabled":true,"audit.enabled":true,"self_service.enabled":true,"fleet_updates.enabled":true,"fleet_updates.mutation.enabled":true}}'
api_mutate PUT "/fork/settings/features" "$feature_body" "$RUN_DIR/features-on-save.json"

if [[ "$STABLE_RESTART_REGRESSION_ONLY" == 1 ]]; then
    api_get "/fork/settings/features" "$RUN_DIR/features-on-before-restart.json"
    jq -e '([.obj.items[] | select(.key == "analytics.enabled" or .key == "dns_intelligence.enabled") | .enabled] | sort) == [true, true]' \
        "$RUN_DIR/features-on-before-restart.json" >/dev/null || fail "CatX enabled flags were not persisted before the real restart"
    restart_panel_via_api enabled
    assert_restart_surfaces enabled

    api_mutate PUT "/fork/settings/features" \
        '{"flags":{"analytics.enabled":false,"dns_intelligence.enabled":false,"policies.enabled":false,"traffic_control.enabled":false,"audit.enabled":false,"self_service.enabled":false,"fleet_updates.enabled":false,"fleet_updates.mutation.enabled":false}}' \
        "$RUN_DIR/features-off-save.json"
    api_get "/fork/settings/features" "$RUN_DIR/features-off-before-restart.json"
    jq -e '([.obj.items[] | select(.key == "analytics.enabled" or .key == "dns_intelligence.enabled") | .enabled] | sort) == [false, false]' \
        "$RUN_DIR/features-off-before-restart.json" >/dev/null || fail "CatX disabled flags were not persisted before the real restart"
    restart_panel_via_api disabled
    assert_restart_surfaces disabled
    api_get "/clients/get/${CLIENT_EMAIL}" "$RUN_DIR/client-after-restart.json"
    jq -e --arg email "$CLIENT_EMAIL" '.obj.client.email == $email' "$RUN_DIR/client-after-restart.json" >/dev/null ||
        fail "synthetic database client was not preserved across real panel restarts"
    if grep -Eiq 'panic|fatal|startup loop' "$RUN_DIR/panel.log"; then
        fail "panel log contained panic/fatal/startup-loop evidence during real restart regression"
    fi
    log "PASS: Stable Qualification real restart regression completed on public ${RELEASE_VERSION} assets"
    exit 0
fi

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

log "INFO: preparing owned Linux Traffic Control substrate rule"
substrate_rule=$(jq -nc --arg iface "$MANAGED_INTERFACE" '{rules:[{nodeKey:"rc6-linux-node",clientKey:"rc6-substrate",interface:$iface,mark:6001,uploadRateBps:1000000,downloadRateBps:1000000,selectors:["127.0.0.1/32"]}]}') || fail "could not encode substrate rule"
log "INFO: substrate rule encoded"
log "INFO: applying owned Linux Traffic Control substrate rule"
api_mutate POST "/traffic-control/reconcile" "$substrate_rule" "$RUN_DIR/reconcile-apply-1.json"
log "INFO: first substrate reconcile returned successfully"
api_mutate POST "/traffic-control/reconcile" "$substrate_rule" "$RUN_DIR/reconcile-apply-2.json"
log "INFO: second substrate reconcile returned successfully"
tc qdisc show dev "$MANAGED_INTERFACE" > "$EVIDENCE_DIR/tc-after-apply.txt" || fail "tc qdisc inspection failed after substrate apply"
nft list table inet catx_traffic_control > "$EVIDENCE_DIR/nft-after-apply.txt" || fail "nft inspection failed after substrate apply"
log "INFO: inspecting applied tc/nft ownership succeeded"
api_mutate POST "/traffic-control/reconcile" '{"rules":[]}' "$RUN_DIR/reconcile-remove-1.json"
api_mutate POST "/traffic-control/reconcile" '{"rules":[]}' "$RUN_DIR/reconcile-remove-2.json"
log "INFO: explicit empty reconciles returned successfully"
tc qdisc show dev "$MANAGED_INTERFACE" > "$EVIDENCE_DIR/tc-after-remove.txt" || fail "tc qdisc inspection failed after substrate remove"
if nft list table inet catx_traffic_control > "$EVIDENCE_DIR/nft-after-remove.txt" 2>/dev/null; then
    fail "CatX nft table remained after explicit empty reconcile"
fi
grep -q 'noqueue' "$EVIDENCE_DIR/tc-after-remove.txt" || fail "managed qdisc was not removed"
log "PASS: Linux Traffic Control reconcile/apply/remove was idempotent and cleaned its owned state"

if [[ -n "$CANDIDATE_BINARY" || "$RELEASE_TAG" == "v0.1.0-rc.3" ]]; then
    ip link show dev "$ADMIN_INTERFACE" >/dev/null 2>&1 && ip link delete "$ADMIN_INTERFACE" 2>/dev/null || true
    ip link add "$ADMIN_INTERFACE" type dummy
    ip link set dev "$ADMIN_INTERFACE" up
    tc qdisc replace dev "$ADMIN_INTERFACE" root handle 8000: fq_codel
    tc qdisc show dev "$ADMIN_INTERFACE" > "$EVIDENCE_DIR/admin-qdisc-before.txt"
    export CATX_TRAFFIC_CONTROL_INTERFACES="$MANAGED_INTERFACE,$ADMIN_INTERFACE"
    stop_panel
    start_panel
    wait_panel
    login
    admin_reconcile_status=$(curl --silent --show-error -b "$COOKIE_FILE" -c "$COOKIE_FILE" -X POST \
        -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF_TOKEN" \
        --data "$substrate_rule" -o "$RUN_DIR/admin-reconcile.json" -w '%{http_code}' \
        "$BASE_URL/panel/api/traffic-control/reconcile" || true)
    [[ "$admin_reconcile_status" == 500 ]] || fail "admin-owned qdisc refusal returned HTTP $admin_reconcile_status instead of 500"
    jq -e '.success == false' "$RUN_DIR/admin-reconcile.json" >/dev/null || fail "admin-owned qdisc refusal returned an invalid API envelope"
    tc qdisc show dev "$ADMIN_INTERFACE" > "$EVIDENCE_DIR/admin-qdisc-after.txt"
    cmp -s "$EVIDENCE_DIR/admin-qdisc-before.txt" "$EVIDENCE_DIR/admin-qdisc-after.txt" || fail "admin-owned qdisc changed during refusal"
    if nft list table inet catx_traffic_control > "$EVIDENCE_DIR/admin-nft-after.txt" 2>/dev/null; then
        fail "admin-owned qdisc refusal unexpectedly left CatX nft state"
    fi
    log "PASS: hosted admin-owned qdisc refusal left the separate safe interface unchanged"
    stop_panel
    tc qdisc del dev "$ADMIN_INTERFACE" root
    ip link delete "$ADMIN_INTERFACE"
    export CATX_TRAFFIC_CONTROL_INTERFACES="$MANAGED_INTERFACE"
    start_panel
    wait_panel
    login
fi

group_body=$(jq -nc '{quotaBytes:1073741824,multiplierPpm:1000000,resetPeriod:"never",resetDay:1}')
api_mutate PUT "/clients/groups/quota/${CLIENT_GROUP}" "$group_body" "$RUN_DIR/group-save.json"
api_get "/clients/groups/quota/${CLIENT_GROUP}" "$RUN_DIR/group-view-before.json"
jq -e '.obj.quotaEnabled == true and .obj.quotaBytes == 1073741824' "$RUN_DIR/group-view-before.json" >/dev/null || fail "group quota configuration was not active"

policy_body=$(jq -nc --arg email "$CLIENT_EMAIL" '{clientEmail:$email,enabled:true,windowSeconds:60,quotaBytes:1,activeUploadBps:1000000,activeDownloadBps:1000000,throttleUploadBps:1000,throttleDownloadBps:1000}')
api_mutate PUT "/traffic-control/clients/${CLIENT_EMAIL}/policy" "$policy_body" "$RUN_DIR/policy-save.json"
api_get "/traffic-control/clients/${CLIENT_EMAIL}/policy" "$RUN_DIR/policy-before.json"
jq -e '.obj.enforcement == "unsupported" and (.obj.enforcementNote | contains("generic Xray users"))' "$RUN_DIR/policy-before.json" >/dev/null || fail "generic per-client enforcement was not reported honestly"

[[ -n "$CLIENT_XRAY_PID" ]] || start_client_xray
# The first Xray stats poll after a panel restart establishes baselines. Let
# that poll complete before generating the traffic whose delta this observation
# measures.
sleep 6
api_get "/clients/traffic/${CLIENT_EMAIL}" "$RUN_DIR/traffic-before.json"
traffic_before=$(jq -er '(.obj.up // 0) + (.obj.down // 0)' "$RUN_DIR/traffic-before.json")
traffic_after=$traffic_before
for _ in $(seq 1 18); do
    curl --fail --silent --show-error --socks5-hostname "127.0.0.1:${SOCKS_PORT}" -o /dev/null https://example.com
    sleep 1
    api_get "/clients/traffic/${CLIENT_EMAIL}" "$RUN_DIR/traffic-after.json"
    traffic_after=$(jq -er '(.obj.up // 0) + (.obj.down // 0)' "$RUN_DIR/traffic-after.json")
    if (( traffic_after > traffic_before )); then
        break
    fi
    sleep 4
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
