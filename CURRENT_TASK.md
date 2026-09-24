# CURRENT TASK

## Work Package

`WP-2A — DNS & Evidence Fusion`

## Goal

Correlate DNS, destination IP/domain, visible SNI/TLS metadata, and existing
analytics events into a unified evidence model without TLS MITM or content
inspection.

## Required

- DNS observation ingestion interfaces;
- correlation between DNS answers and subsequent destination IP connections;
- destination domain/IP evidence fusion;
- visible SNI evidence where available;
- provenance/source tracking for every inferred domain/service hint;
- confidence score/level;
- deduplication and expiration/TTL semantics;
- client-aware correlation;
- safe behavior when evidence is ambiguous or conflicting;
- persistence using the existing WP-1A analytics layer;
- no duplicate client/group/node models;
- no duplicate traffic counters;
- analytics OFF = exact no-op.

Evidence must distinguish at minimum:

- directly observed domain;
- DNS-derived domain;
- SNI-derived domain;
- IP-only destination;
- inferred correlation.

## Hard constraints

- no TLS MITM;
- no decrypted HTTPS;
- no cookies;
- no Authorization headers;
- no credentials;
- no request/response body capture;
- do not implement service/category classification yet — WP-2B;
- do not implement DNS UI/privacy controls yet — WP-2C;
- evidence-fusion failure must never affect Xray/panel operation;
- reuse WP-1A/WP-1B analytics storage and session pipeline;
- minimal upstream-touch points;
- `make verify-fork` must remain green.

## Required tests

- DNS → IP correlation;
- TTL expiry;
- ambiguous/multiple-domain evidence;
- SNI precedence/provenance;
- client isolation;
- deduplication;
- restart/persistence behavior;
- race/concurrency where relevant;
- SQLite/PostgreSQL where applicable.

## Restrictions

- do not modify `main`;
- do not start WP-2B;
- do not implement WP-2C;
- do not merge WP-2A into `develop` until its merge gate is satisfied.
