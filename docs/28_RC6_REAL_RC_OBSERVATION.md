# RC-6 Real RC Observation Record

This record is for the observation of the already-published public release
`v0.1.0-rc.2`. It is not a product roadmap and it is not permission to
publish `v0.1.0`.

## Release under observation

- Release: `v0.1.0-rc.2`
- Immutable commit: `4d8feae2e62db914d3146504340d9b9f802088b2`
- Public release: <https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.1.0-rc.2>
- Test data: synthetic identities and disposable storage only
- Hosts: isolated disposable Windows observation environment and GitHub-hosted disposable Ubuntu runner; no production node or data

## Evidence rules

Record the UTC start and end time, exact asset or image digest, host/runtime,
and the command, panel path, or screenshot/log used for each result. Use
`PASS`, `FAIL`, `BLOCKED`, or `NOT RUN`; do not convert `NOT RUN` into a pass.
Redact credentials, tokens, cookies, and authorization headers from retained
evidence.

## Observation checklist

### Baseline with CatX features disabled

- [x] Fresh public RC-2 Windows asset start and first login
- [x] Panel, SQLite database, and Xray 26.9.9 start normally
- [x] Existing upstream API/config routes remain available
- [x] No fork startup failure was present in the runtime log
- [x] Panel/process restart preserves login, inbound, client, and subscription data

### Upstream representative flows

- [x] Create an inbound and synthetic client; edit/disable/remove were not run
- [x] Generate and fetch the client subscription
- [x] Read the running Xray configuration and verify Xray remains healthy
- [x] Exercise a local VLESS client connection through Xray to `example.com`
- [x] Observe non-zero inbound and client traffic counters; counter reset was not run
- [x] Restart Xray and the panel process independently
- [x] Restore a downloaded SQLite backup in the disposable panel and verify data remains
- [ ] Host reboot persistence on Linux

### CatX observation flows

- [x] Activity/DNS/history endpoints return enabled, metadata-only empty views; service/Xray evidence is healthy
- [x] Feature settings are default-off; enabled flags survive restart
- [x] Policy create, assign, deny/allow simulation, schedule, and managed-DNS explanation paths
- [ ] Traffic Control quota, rolling window, and speed/throttle paths — Linux capability detection passed, but public RC-2 substrate apply is source-blocked
- [x] Audit events, retention, bounded metrics, portal token/profile/traffic, and fleet RC dry-run paths
- [ ] Feature disablement after enablement and restart — not reached after the Linux substrate blocker
- [x] Retained evidence contains synthetic identities and sanitized metadata only

### Actual observation result

- Runtime: public `v0.1.0-rc.2` Windows amd64 payload, panel `0.1.0-rc.2`, Xray `26.9.9`
- Feature-off baseline: all managed CatX flags reported `false` before enablement
- Feature-on exercise: analytics, DNS intelligence, policies, traffic control, audit, self-service, and fleet update flags enabled through the panel API and restarted
- Proxy exercise: 30/30 status-plus-VLESS samples passed; each sample received HTTP 200 through the local VLESS client
- Stability window: 2026-09-30 13:24:38Z–13:39:43Z, 15m 05s
- Observation log: `C:\Users\user\AppData\Local\Temp\catx-ui-rc6-observation-window-d572de2a07a54e67af94f61b771fee37.log`
- Linux observation workflow: `RC-6 Linux public RC observation`, run [36742297152](https://github.com/CatCodeArbelin/CatX-UI/actions/runs/36742297152), commit `96396546c567a7dcf0c94ad4622cbbd02e231c9e`
- Linux public payload and metadata sidecars: PASS; exact RC-2 identity and commit `4d8feae2e62db914d3146504340d9b9f802088b2` matched
- Linux feature-off baseline: PASS; panel, SQLite, login, Xray, synthetic VLESS inbound/client, computed subscription endpoint, and VLESS proxy to `example.com` (HTTP 200)
- Linux namespace capability probe: PASS; disposable netns test ran with `CATX_TC_NETNS_TEST=1`
- Linux Traffic Control capabilities: PASS; `tc`, nftables, managed dummy interface, and `CAP_NET_ADMIN` were ready, while `UserAttribution=false` remained honest
- Linux Traffic Control substrate apply: BLOCKED; public RC-2 returned `Parent Qdisc doesn't exists` while installing its ingress redirect filter because it never creates the required ingress qdisc
- Linux quota/policy lifecycle and feature-disable cleanup: NOT RUN after the source blocker
- Linux host reboot persistence: NOT RUN; GitHub Actions cannot reboot its host without invalidating the runner
- Local Docker Linux observation: NOT RUN; Docker Desktop Linux engine was unavailable and no local repair was pursued

### Docker verification result

- Workflow YAML, CatX GHCR ownership, forbidden-target absence, and preserved multi-architecture list: PASS
- Local no-push multi-architecture build: cancelled after Docker Desktop stopped answering while emulated ARM/386 stages were active; no image was pushed
- GitHub `workflow_dispatch`: not run because the GitHub CLI/API dispatch credential was unavailable

### Stability window

Record actual observations only:

| Field | Value |
| --- | --- |
| Start (UTC) | 2026-09-30 13:24:38Z |
| End (UTC) | 2026-09-30 13:39:43Z |
| Duration | 15m 05s |
| Runtime restarts | 0 during soak; panel and Xray restarted before soak |
| Panel/Xray restarts | 1 panel process restart; 1 Xray restart before soak |
| Synthetic requests/flows | 30 proxy probes plus panel/API, subscription, backup/restore, policy, portal, audit, and fleet flows |
| CPU/RAM trend | Panel status remained responsive; no monotonic failure trend observed |
| Log growth or recurring errors | No recurring runtime error observed during the soak |
| Database marker after restart | Inbound/client/subscription and traffic data persisted |

## Decision

- Source-level blocker: confirmed in public `v0.1.0-rc.2`; Traffic Control redirect setup omits the required ingress qdisc, so a focused fix and new fully-qualified `v0.1.0-rc.3` are required
- Current-branch focused fix: the backend now creates the owned ingress qdisc, refuses a pre-existing administrator-owned ingress/clsact qdisc, and removes the owned ingress qdisc during rollback/disable cleanup; focused Traffic Control tests pass
- Release-only/infrastructure blocker: host reboot persistence remains `NOT RUN` because a GitHub-hosted workflow cannot reboot its runner; this is separate from the RC-2 source blocker
- Stable-qualification evidence package: not started

If a source-level blocker is confirmed, require a focused fix and a new
fully-qualified `v0.1.0-rc.3`. If no source-level blocker is found, close this
record and stop before the separate Stable Qualification package.
