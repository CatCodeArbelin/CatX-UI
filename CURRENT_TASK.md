# CURRENT TASK

## Work Package

`WP-2C — DNS UI, Privacy & Retention`

## Goal

Expose the existing metadata-only DNS/evidence and analytics controls through
a safe, privacy-preserving UI while keeping retention behavior explicit,
bounded, and compatible with the upstream panel.

## Required

- DNS intelligence and enrichment visibility in the fork analytics UI;
- privacy disclosures and safe defaults;
- retention configuration and enforcement using the existing analytics layer;
- feature-disabled compatibility and no-op behavior;
- SQLite/PostgreSQL compatibility where persistence changes are required;
- focused tests and `make verify-fork` before merge.

## Hard constraints

- no TLS MITM, decrypted HTTPS, payload inspection, cookies, credentials, Authorization headers, or request bodies;
- no Policy Engine, blocking, or traffic enforcement;
- do not duplicate upstream client/group/node/session/traffic models;
- analytics/DNS intelligence OFF must remain a no-op;
- enrichment failure must never affect Xray or panel core;
- minimal upstream-touch points;
- do not modify `main`.

## Restrictions

- do not start a later work package;
- do not merge into `main`;
- preserve the low-divergence downstream fork architecture.
