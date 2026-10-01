[English](/README.md) | [فارسی](/README.fa_IR.md) | [العربية](/README.ar_EG.md) | [中文](/README.zh_CN.md) | [Español](/README.es_ES.md) | [Русский](/README.ru_RU.md) | [Türkçe](/README.tr_TR.md)

<p align="center">
  <strong>CatX-UI</strong> — a maintained downstream fork of <a href="https://github.com/MHSanaei/3x-ui">3x-ui</a> for operating Xray panels with independent Policy, Traffic Control, Insight, and Operations capabilities.
</p>

<p align="center">
  <a href="https://github.com/CatCodeArbelin/CatX-UI/releases/latest"><img src="https://img.shields.io/github/v/release/CatCodeArbelin/CatX-UI?sort=semver" alt="Latest CatX-UI release"></a>
  <a href="https://github.com/CatCodeArbelin/CatX-UI/actions/workflows/ci.yml"><img src="https://github.com/CatCodeArbelin/CatX-UI/actions/workflows/ci.yml/badge.svg" alt="CatX Actions CI"></a>
  <a href="https://www.gnu.org/licenses/gpl-3.0.html"><img src="https://img.shields.io/badge/license-GPLv3-blue.svg" alt="GPLv3 license"></a>
  <a href="https://github.com/CatCodeArbelin/CatX-UI/releases"><img src="https://img.shields.io/github/downloads/CatCodeArbelin/CatX-UI/total" alt="CatX-UI downloads"></a>
</p>

## What CatX-UI is

CatX-UI preserves the upstream 3x-ui panel, APIs, subscriptions, database
support, and Xray integration, then adds first-party fork modules behind
independent settings. CatX is maintained in its own repository and release
channel; it is not an official upstream project and is not endorsed by
MHSanaei or the 3x-ui maintainers.

Stable `v0.1.0` is based on **MHSanaei/3x-ui v3.8.5** and bundles **Xray
26.9.9**. The internal Go module/import path remains
`github.com/mhsanaei/3x-ui/v3` for upstream compatibility.

## What is included

- **Policy** — human-friendly policies, groups and client overrides, schedules,
  domain/category decisions, simulation, explanation, quarantine, and managed
  DNS controls where supported.
- **Insight** — Activity, traffic history, DNS Intelligence, destination
  metadata, service/category classification, retention, privacy controls, and
  anomaly signals.
- **Traffic Control** — quotas, windows, speeds, and soft-throttle workflows
  using capability detection and explicit supported/degraded/unsupported state.
- **Operations** — audit, webhooks, metrics, self-service, fleet visibility,
  backup/restore, and a checksum-verified transactional updater with recovery.
- **Upstream panel** — Xray protocols and transports, subscriptions, clients,
  inbounds, routing, nodes, SQLite/PostgreSQL, bots, and the familiar 3x-ui UI.

Fork modules are disabled by default where applicable. Traffic Control is not
universally per-user kernel shaping: generic Xray-user attribution is not
supported, so the panel must report capability limits instead of claiming
enforcement it cannot prove.

## Install CatX-UI

The stable installer is served from this repository and verifies CatX release
identity and checksums before activation:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh)
```

Install a specific release, including stable `v0.1.0`:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/CatCodeArbelin/CatX-UI/main/install.sh) v0.1.0
```

`dev-latest` is an opt-in rolling development channel. RC tags are for
advanced testing only; stable installations do not follow prereleases.
Unattended/cloud-init examples are in [`deploy/`](deploy/).

After installation, `x-ui` opens the management menu. `x-ui update` follows
the stable channel and `x-ui update-dev` selects the development channel.
Updates use snapshot → validate → stage → activate → healthcheck; a failed
activation attempts automatic restoration of the previous known-good binary,
configuration, and database. Back up before planned changes and retain the
recovery snapshot if an update reports rollback failure.

## Docker

The published stable image is:

```text
ghcr.io/catcodearbelin/catx-ui:v0.1.0
```

The stable aliases `v0.1.0`, `0.1.0`, and `latest` are published for
`linux/amd64`, `linux/arm64`, `linux/arm/v7`, `linux/arm/v6`, and `linux/386`.

```bash
docker run -d --name catx-ui \
  --cap-add=NET_ADMIN --cap-add=NET_RAW \
  -v "$PWD/db:/etc/x-ui" -v "$PWD/cert:/root/cert" \
  -p 2053:2053 --restart unless-stopped \
  ghcr.io/catcodearbelin/catx-ui:v0.1.0
```

The repository [`docker-compose.yml`](docker-compose.yml) builds the same
CatX image locally and includes an optional PostgreSQL profile. Keep the
capabilities when using Fail2ban/IP-limit enforcement; without them a ban may
be recorded but cannot be applied.

## Supported platforms and databases

Stable release archives are published for Linux `amd64`, `arm64`, `armv7`,
`armv6`, `armv5`, `s390x`, and `386`, plus Windows `amd64`. The installer
supports the distributions listed in its current release contract; verify the
target OS before unattended deployment. Docker has the narrower manifest
listed above.

SQLite is the default zero-setup database. PostgreSQL is supported for larger
or multi-node deployments through `XUI_DB_TYPE=postgres` and
`XUI_DB_DSN`. Use the panel migration/backup paths; do not replace database
files by hand.

## Security and privacy boundaries

CatX collects only the metadata needed for its documented panel, insight, and
policy functions: identities, timestamps, domains/DNS, SNI when visible,
destination IP/port, protocol/network, traffic, node/inbound/outbound, and
policy decisions. It does **not** use TLS MITM, decrypt HTTPS, or collect
HTTP bodies, cookies, Authorization headers, credentials, passwords, or
messages. Risk signals are heuristics and do not auto-ban users from one
signal. CatX is not a billing platform or a generic plugin ecosystem.

Capability-dependent behavior is shown explicitly: supported, degraded, or
unsupported is preferable to silently simulating enforcement. Optional Xray
APIs, Linux privileges, network topology, and database services can limit a
feature without changing the upstream-compatible baseline.

## Documentation and support

- [CatX repository documentation](docs/README.md) — local/in-repository docs;
  no independent public CatX documentation site is claimed yet.
- [Stable v0.1.0 release](https://github.com/CatCodeArbelin/CatX-UI/releases/tag/v0.1.0)
- [All CatX releases](https://github.com/CatCodeArbelin/CatX-UI/releases)
- [Issues and feature requests](https://github.com/CatCodeArbelin/CatX-UI/issues)
- [Security policy](SECURITY.md)
- [Contributing guide](CONTRIBUTING.md)

When reporting a problem, include the CatX version, Xray version, install
method, OS/architecture, database backend, and a redacted log excerpt. Report
upstream-only behavior to the upstream project when appropriate, while noting
that CatX is a downstream fork.

## Upstream acknowledgment and license

CatX-UI is based on and remains compatible with
[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui). Upstream copyright and
GPL notices are preserved; this project does not imply endorsement by upstream
maintainers. CatX-UI is distributed under [GPLv3](LICENSE), and corresponding
source is provided for modified GPL-covered binaries.
