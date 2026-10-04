# CatX v0.3 Product Hardening

## Status

`CURRENT — source hardening and qualification harness recorded; hosted behavioral qualification pending`

This document records facts and evidence for the focused hardening package. It
does not authorize a release, tag, Docker publication, merge into `main`, or
automatic upstream synchronization.

## Baseline and ancestry

| Item | Value |
| --- | --- |
| Base branch | `chore/sponsors-real-restart-smoke` |
| Base SHA | `f2152a90992abbdfe402fd5694aa41cceeaddfe6` |
| Hardening branch | `feature/v0.3-product-hardening` |
| Preserved lifecycle harness | `scripts/staging/observe-sponsors-real-linux.sh` |
| Preserved lifecycle workflow | `.github/workflows/sponsors-real-restart-smoke.yml` |
| Stable release boundary | `main` and `v0.2.0` unchanged |

`git fetch origin --prune --tags` reached the downstream remote but refused to
clobber the pre-existing local `dev-latest` tag. No branch ref was rewritten.
The exact tag collision is a qualification item, not a reason to delete or
move a tag during this package.

## Initial product-contract matrix

The following is an implementation inventory from the checked-out source. It
is deliberately not a claim that the behavior has passed real-panel
qualification. The final column must be replaced with observed evidence after
the disposable panel flow is run.

| Feature | Flag | UI route | API/backend boundary | Persistence/runtime | Initial qualification state |
| --- | --- | --- | --- | --- | --- |
| Analytics / Activity | `analytics.enabled` | `/activity`, client Activity tab | `/panel/api/analytics/*`; `internal/analytics` | analytics event/session/aggregate tables; `analytics.Configure`, collector start | pending real restart and empty-state proof |
| DNS Intelligence | `dns_intelligence.enabled` (requires Analytics) | Activity DNS tab | `/panel/api/analytics/clients/:email/dns`; `internal/analytics` | DNS observation tables; evidence runtime | pending capability/disabled/empty proof |
| Policy | `policies.enabled` | `/policies` | `/panel/api/policies*`; `internal/policy` | policy, assignment, override, schedule tables; Xray decorator | pending CRUD and zero-policy proof |
| Simulator / Explain | `policies.enabled` | Policy Simulator tab | `/panel/api/policies/simulate`, `/explain`, `/explain-route` | production decision/compiler path | pending consistency proof |
| Schedules / Temporary Overrides | `policies.enabled` | Policy Schedules/Overrides tabs | `/panel/api/policies/schedules*`, `/temporary-overrides*` | policy schedule/temporary tables; read-time expiry | pending boundary and restart persistence proof |
| Traffic History | `analytics.enabled` | Activity traffic tab | `/panel/api/analytics/clients/:email/traffic` | upstream counters plus analytics snapshots/aggregates | pending empty vs error proof |
| Traffic Control / QoS | `traffic_control.enabled` | client edit and client information surfaces | `/panel/api/traffic-control/*`; `trafficpolicy`, `trafficcontrol` | traffic policy/state + upstream client traffic; capability adapter | mutable controls moved to edit tabs; information view is read-only; real-panel qualification pending |
| Group Quota | `traffic_control.enabled` | existing client/group workflows | `/panel/api/clients/groups/quota/*`; `groupquota` | group quota state/membership; upstream traffic authoritative | pending integration qualification |
| Risk | `security_anomaly.enabled` (requires Analytics) | client Risk panel | `/panel/api/risk/*`; `risk` | risk history/events/scores/suppressions | pending truthful degraded proof |
| Audit | `audit.enabled` | `/audit` | `/panel/api/fork/audit/*`; `audit` | audit events/webhooks/retention | pending authorization/empty proof |
| Webhooks | `audit.enabled` | `/webhooks` | `/panel/api/fork/audit/webhooks*`; `audit` | webhook endpoints/deliveries | pending disabled/degraded proof |
| Metrics | `metrics.enabled` | API/operations surface | `/panel/api/fork/metrics`; audit/metrics boundary | process/runtime metrics; no separate UI route | pending capability proof |
| Portal | `self_service.enabled` | `/portal-access`, client portal `/portal` | `/panel/api/portal/*`, `/portal/*`; `portal` | portal credentials/host grants and selector adapter | selector UX hardening implemented; real-panel qualification pending |
| Fleet | `fleet_updates.enabled` | `/fleet` | fleet/node API plus `fleetupdate` | existing node models and fleet read service | pending representative flow |
| Fleet Update | `fleet_updates.enabled`, mutation sub-flag | `/fleet-updates` | `/panel/api/fleet-updates/*`; `fleetupdate` | campaign/target tables; executor/runtime node boundary | pending safety/degraded proof |
| Sponsors | `sponsors.enabled` | `/catx/sponsors`, `/catx/sponsors/manage` | `/panel/api/fork/sponsors*`; `forkext/sponsors` | `fork_sponsors` plus settings; lifecycle proof preserved | lifecycle evidence exists; requalify in aggregate matrix |

## Human-review findings mapped to source

| Finding | Current evidence | Planned hardening boundary |
| --- | --- | --- |
| “analytics unavailable” / “policy data unavailable” | Existing APIs mixed healthy empty, feature-off, and failure states; real restart proof was missing | typed state envelopes, focused API tests, real restart matrix, and shared UI state component; hosted result remains pending |
| Typed feature-off response lost in the browser adapter | HTTP failures discarded the response `obj`, so a `409` state envelope could become a generic page error | the shared HTTP boundary now preserves error objects and nested `featureDisabled`; regression coverage is in `frontend/src/test/httpUtil.test.ts` |
| Traffic/QoS editing in client information | `TrafficPolicyPanel` was mounted as the primary editor in `ClientInfoModal` | the same panel is mounted as an editable tab in `ClientFormModal`; information view passes `readOnly` |
| duplicate generic “Speed” labels | `TrafficPolicyPanel` used `pages.clients.speed` twice | semantic upload/download/throttle labels and RU/EN keys |
| raw JSON primary Policy editor | `PolicyPage` made `policyJson` the main editing field | structured Definition editor is primary; Advanced JSON preserves unknown fields |
| WP/low-level implementation terminology | RU/EN policy capability text included `WP-4B`; Portal labels exposed technical IDs | operator copy is capability-oriented; technical JSON remains under Advanced |
| raw Portal IDs / mixed validation | Portal admin used numeric inputs and `client/group` wording | searchable upstream entity selectors and localized human labels; IDs remain API-side |
| feature toggle implies active runtime | settings page showed a blanket restart alert without active/pending comparison | saved and active state, pending restart, and runtime error are returned per feature and rendered per row |

## Behavioral evidence table

| Feature | Flag saved | Restart required | Restart completed | Backend route | HTTP result | Runtime initialized | UI result | Classification |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Analytics | pending | pending | pending | pending | pending | pending | pending | pending |
| DNS Intelligence | pending | pending | pending | pending | pending | pending | pending | pending |
| Policy | pending | pending | pending | pending | pending | pending | pending | pending |
| Traffic Control | pending | pending | pending | pending | pending | pending | pending | pending |
| Portal | pending | pending | pending | pending | pending | pending | pending | pending |
| Sponsors | preserved lifecycle evidence; aggregate rerun pending | pending | pending | pending | pending | pending | pending | pending |

## Source-level hardening evidence

The following is verified from focused tests and compilation on this branch;
it is not a substitute for the disposable Linux panel run.

| Contract | Change | Focused evidence |
| --- | --- | --- |
| Saved versus active runtime | `RuntimeState` snapshots distinguish `feature_off`, `active`, `initializing`, `restart_required`, and `error`. | `internal/forkext/runtime_state_test.go`, `internal/forkext/feature_settings_test.go` |
| Healthy empty analytics/policy responses | Active analytics pages and zero-policy lists return successful typed envelopes instead of generic unavailable text. | `internal/analytics/activity_api_test.go`, `internal/policy/api_test.go` |
| Disabled and unsupported states | Analytics, Policy, Simulator, Traffic Policy, and settings expose stable state fields; generic Xray attribution remains unsupported. | focused API tests and traffic-policy tests |
| Non-2xx state envelopes | Feature-off response metadata survives the shared frontend HTTP adapter instead of becoming a generic error. | `frontend/src/test/httpUtil.test.ts` |
| Traffic unconfigured state | A client without a traffic policy returns `200` with `state=unconfigured`; the UI renders the shared state component instead of a red error. | `TestUnconfiguredPolicyRouteExposesTypedFeatureState` (SQLite-backed; local execution requires CGO) |
| Structured Policy editing | Supported Definition fields round-trip while unknown fields survive edits. | `frontend/src/test/policy-definition.test.ts` |
| English/Russian CatX contract | EN/RU key trees and placeholders are checked; other locales use the existing English fallback strategy. | `frontend/src/test/catx-i18n-contract.test.ts` |

## Required qualification and recovery

The real-panel flow is:

```text
start disabled
→ enable and save
→ observe pending restart
→ POST /panel/api/setting/restartPanel
→ wait for panel recovery and authenticate again if needed
→ inspect state, API, runtime, and UI
→ disable and repeat
```

Any Xray, traffic, DNS, updater, or migration-affecting correction must retain
the repository recovery sequence:

```text
snapshot → validate → apply → healthcheck → commit
FAIL → restore known-good state → reload/restart → healthcheck → audit
```

No new sensitive data is collected. Analytics remains metadata-only and does
not inspect TLS/HTTPS bodies, cookies, credentials, Authorization headers, or
passwords.

## Qualification artifacts

The hosted qualification entry point is
`scripts/staging/observe-product-hardening-real-linux.sh`. It uses a fresh
temporary database and application directory, exercises saved-off → restart →
active → restart persistence → feature-off transitions, seeds only synthetic
client/policy/sponsor records, and invokes
`scripts/staging/observe-product-hardening-real-browser.mjs` against the same
live panel. The browser flow covers login, settings pending/active state,
Analytics, structured Policy create/edit, client edit Traffic/QoS, read-only
client information, Portal token issue, Sponsors, EN/RU, 1920×1080, and
1366×768. Logs are redacted and the temporary runtime is removed on exit.

The workflow is `.github/workflows/catx-product-hardening-real.yml`. It is
branch-scoped, does not publish an artifact or release, and preserves the
existing Sponsors lifecycle workflow.

## Local qualification boundary

The Windows workspace passed frontend typecheck and formatting, changed Go
package compilation, focused no-database Go API/runtime tests, the structured
policy contract, the EN/RU translation contract, OpenAPI regeneration through
the equivalent TypeScript runner, browser-script syntax, and `git diff
--check`. The canonical `npm run gen:api` command cannot load `.ts` imports
under the installed Node `20.12.0`; the repository declares Node `>=26`, so
the generated OpenAPI artifact was produced with the temporary TypeScript
runner and is stable.

The full SQLite-backed Go suite and race detector require CGO; this workspace
has `CGO_ENABLED=0` and no C compiler. `make` is not installed, so `make
verify` and `make verify-fork` were not executable locally. Frontend unit tests
and lint stop before test execution because the installed Node 20 dependency
tree is missing the native Rolldown/Oxlint optional bindings. The disposable
Linux real-panel/browser workflow remains the required hosted qualification and
has not been claimed as passed.

## Safe Windows maintainer review setup

Replace only `$CandidateSha` with the exact candidate SHA from the final report.
The block refuses to reuse an existing review resource, checks the candidate
before creating a worktree, uses a normal bridge network and loopback-only
published port, disables Fail2ban for this disposable review, and never runs a
Docker-wide cleanup command.

```powershell
$ErrorActionPreference = 'Stop'
$CandidateSha = '<exact-candidate-sha-from-final-report>'
$Repo = (Resolve-Path 'F:\3xUiPatch\CatX-UI').Path
if ($CandidateSha -notmatch '^[0-9a-fA-F]{40}$') { throw 'Set the exact 40-character candidate SHA first.' }
$Resolved = (git -C $Repo rev-parse "$CandidateSha^{commit}").Trim()
if ($Resolved -ne $CandidateSha) { throw "Candidate SHA mismatch: $Resolved" }

$ShortSha = $CandidateSha.Substring(0, 12).ToLowerInvariant()
$Parent = (Split-Path $Repo -Parent)
$ReviewRoot = Join-Path $Parent "CatX-UI-review-$ShortSha"
$ReviewBranch = "review/v0.3-hardening-$ShortSha"
$Image = "catx-ui-review:$ShortSha"
$Container = "catx-ui-review-$ShortSha"
$Volume = "catx-ui-review-db-$ShortSha"
$Network = "catx-ui-review-net-$ShortSha"
$HostPort = 29286
$ReviewBasePath = '/catx-review/'

if (Test-Path -LiteralPath $ReviewRoot) { throw "Review worktree already exists: $ReviewRoot" }
if (Get-NetTCPConnection -LocalPort $HostPort -State Listen -ErrorAction SilentlyContinue) { throw "Port $HostPort is already listening." }
if (@(docker ps -a --format '{{.Names}}') -contains $Container) { throw "Container already exists: $Container" }
if (@(docker volume ls -q) -contains $Volume) { throw "Volume already exists: $Volume" }
if (@(docker network ls --format '{{.Name}}') -contains $Network) { throw "Network already exists: $Network" }

git -C $Repo worktree add --detach $ReviewRoot $CandidateSha
git -C $ReviewRoot switch -c $ReviewBranch
docker build --label "org.catx.candidate-sha=$CandidateSha" --tag $Image $ReviewRoot
docker volume create --label "org.catx.candidate-sha=$CandidateSha" $Volume | Out-Null
docker network create --driver bridge --label "org.catx.candidate-sha=$CandidateSha" $Network | Out-Null
docker run --detach --name $Container --network $Network `
  --publish "127.0.0.1:${HostPort}:2053" `
  --mount "type=volume,src=$Volume,dst=/etc/x-ui" `
  --env XUI_ENABLE_FAIL2BAN=false `
  --env XUI_DB_FOLDER=/etc/x-ui `
  --env XUI_PORT=2053 `
  --env XUI_INIT_WEB_BASE_PATH=$ReviewBasePath `
  $Image | Out-Null

$ImageSha = (docker image inspect $Image --format '{{ index .Config.Labels "org.catx.candidate-sha" }}').Trim()
if ($ImageSha -ne $CandidateSha) { throw 'Built image provenance label does not match the candidate SHA.' }
Write-Host "Review panel: http://127.0.0.1:$HostPort${ReviewBasePath}"
Write-Host "Container: $Container  Volume: $Volume  Network: $Network"
function Review-Logs { docker logs --tail 200 $Container }
function Review-Stop { docker stop $Container }
function Review-Start { docker start $Container }
Review-Logs
```

The setup block leaves `Review-Logs`, `Review-Stop`, and `Review-Start` in the
PowerShell session for safe inspection and restart. It does not use the
production database, config, host-wide Docker networking, Fail2ban, or
`docker system prune`.

## Safe Windows maintainer review cleanup

Use the same exact SHA value used for setup. Cleanup refuses to remove a
resource whose provenance label or exact name does not match.

```powershell
$ErrorActionPreference = 'Stop'
$CandidateSha = '<exact-candidate-sha-from-final-report>'
$Repo = (Resolve-Path 'F:\3xUiPatch\CatX-UI').Path
if ($CandidateSha -notmatch '^[0-9a-fA-F]{40}$') { throw 'Set the exact 40-character candidate SHA first.' }
$ShortSha = $CandidateSha.Substring(0, 12).ToLowerInvariant()
$Parent = (Split-Path $Repo -Parent)
$ReviewRoot = Join-Path $Parent "CatX-UI-review-$ShortSha"
if (-not $ReviewRoot.StartsWith($Parent, [System.StringComparison]::OrdinalIgnoreCase)) { throw 'Unexpected worktree path.' }
$ReviewBranch = "review/v0.3-hardening-$ShortSha"
$Image = "catx-ui-review:$ShortSha"
$Container = "catx-ui-review-$ShortSha"
$Volume = "catx-ui-review-db-$ShortSha"
$Network = "catx-ui-review-net-$ShortSha"

$actualContainerSha = (docker inspect $Container --format '{{ index .Config.Labels "org.catx.candidate-sha" }}').Trim()
if ($actualContainerSha -ne $CandidateSha) { throw 'Container provenance mismatch; nothing removed.' }
$actualImageSha = (docker image inspect $Image --format '{{ index .Config.Labels "org.catx.candidate-sha" }}').Trim()
if ($actualImageSha -ne $CandidateSha) { throw 'Image provenance mismatch; nothing removed.' }
$actualVolumeSha = (docker volume inspect $Volume --format '{{ index .Labels "org.catx.candidate-sha" }}').Trim()
if ($actualVolumeSha -ne $CandidateSha) { throw 'Volume provenance mismatch; nothing removed.' }
$actualNetworkSha = (docker network inspect $Network --format '{{ index .Labels "org.catx.candidate-sha" }}').Trim()
if ($actualNetworkSha -ne $CandidateSha) { throw 'Network provenance mismatch; nothing removed.' }

docker stop $Container 2>$null | Out-Null
docker rm $Container | Out-Null
docker image rm $Image | Out-Null
docker volume rm $Volume | Out-Null
docker network rm $Network | Out-Null
if (Test-Path -LiteralPath $ReviewRoot) { git -C $Repo worktree remove --force $ReviewRoot }
if (Test-Path -LiteralPath $ReviewRoot) { throw "Worktree was not removed: $ReviewRoot" }
if (@(git -C $Repo branch --list $ReviewBranch).Count -gt 0) { git -C $Repo branch -D $ReviewBranch }
Write-Host 'CatX hardening review resources removed; production resources were not touched.'
```

## Maintainer Human Review checklist

Use the exact candidate and report defects with a screenshot, route, viewport,
locale, and whether the issue survives a refresh.

- Dashboard and normal/collapsed sidebar navigation.
- Feature Settings: saved value, pending restart, active runtime, error and dependency wording.
- Analytics / Activity, DNS Intelligence, Traffic History, and legitimate empty states.
- Policy creation/editing with structured controls, Advanced JSON, Simulator, Explain, schedules, and temporary overrides.
- Client edit → Traffic/QoS mutable controls; client information → read-only effective Traffic state.
- Traffic quota/window, upload/download limits, used/remaining, enforcement capability, and unsupported attribution wording.
- Risk, Audit, Webhooks, Metrics, Fleet, and Fleet Update capability/degraded states.
- Portal token issuance, client/subject/host selectors, status, last use, and empty host visibility.
- Sponsors and Sponsors Management, including feature-off behavior.
- Russian and English pages: no raw CatX keys, arbitrary English in RU, or mixed validation text.
- Empty, unconfigured, unsupported, degraded, restart-required, and real-error states.
- 1920×1080 and 1366×768: modal overflow, tables, forms, Russian text expansion, and clipped controls.
- Browser console: no fatal errors.

## Final report placeholder

The final candidate report must record the exact candidate SHA, all required
verification results, known limitations, Draft PR/compare URL, safe Windows
review setup and cleanup blocks, and the maintainer checklist. It must confirm
that `main`, stable tags, `v0.2.0`, and release artifacts remain unchanged.
