# CURRENT TASK

## Work Package

`WP-3B — Policy Decision Engine & Xray Compiler`

## Goal

Build a deterministic policy decision engine and compile its decisions into
Xray routing/config behavior through the existing authoritative Xray config
path. Existing upstream `GetXrayConfig()` remains authoritative; policy
compilation must attach through the existing fork decorator seam near the end
of config construction.

## Required

- Reuse the WP-3A policy data contracts and upstream client/group identities;
- separate policy resolution, inspectable decision representation, and pure
  Xray compilation;
- preserve the established precedence contract: temporary override, explicit
  override, client assignment, group assignment, then priority and the
  established deterministic tie-break rules;
- ignore expired temporary overrides and disabled policies;
- retain explanation metadata for future WP-3C simulator/explain behavior;
- compile only authorized policy dimensions, at minimum durable allow/deny
  targets and conservatively supported domain/category/service targets;
- preserve upstream inbounds, outbounds, and routing; append only stable,
  fork-owned fragments in deterministic order;
- make decoration idempotent, duplicate-safe, side-effect free where possible,
  and an exact no-op when policies are disabled or no effective policy exists;
- fail explicitly and safely on malformed policy state without returning a
  partially corrupted candidate configuration;
- use the existing candidate validation and recovery path where practical;
- add comprehensive decision, compiler, structural, idempotence, malformed
  input, isolation, concurrency/race, and feature-disabled tests;
- run the strongest local checks available and use GitHub Actions as the
  canonical Linux verification environment.

## Hard constraints

- do not create a second Xray config generator;
- do not rewrite upstream controller/service/database layers or spread policy
  logic through upstream core files;
- do not invent packet inspection or infer a domain solely from a shared IP;
- unknown or ambiguous classifications must not cause destructive blocking;
- no TLS MITM, decrypted HTTPS, payload inspection, cookies, credentials,
  Authorization headers, or request bodies;
- no duplicate client/group/node/inbound models;
- do not implement WP-3C simulator/UI, schedules UI, quarantine UI, managed
  DNS enforcement unless explicitly authorized here, SafeSearch, QoS,
  throttling, traffic quotas, security anomaly actions, or multi-node policy
  distribution;
- do not modify or merge into `main`;
- do not merge WP-3B into `develop` automatically; stop on the feature branch
  after green CI for review.

## Recovery and safety

Use the existing snapshot, candidate validation, apply, healthcheck, and
rollback behavior for Xray changes. A compiler error must leave the original
upstream configuration untouched and be explicit/testable. Feature-disabled
behavior must remain as close to upstream as possible.
