# CURRENT TASK

## Work Package

`WP-2B — Enrichment & Classification`

Recommended model: `GPT-5.6 Luna Medium`

## Goal

Turn normalized destination evidence from WP-2A into useful service/category metadata while preserving provenance and confidence.

## Required

- domain/service enrichment pipeline;
- IP → ASN/country metadata;
- service identification;
- category identification;
- confidence/provenance for every classification;
- first-party vs inferred classification distinction;
- caching;
- bounded refresh/update behavior;
- safe fallback to `unknown`;
- persistence using existing analytics storage;
- no duplicate traffic/client/node models.

Initial useful categories should support examples such as:

- social;
- video/streaming;
- messaging;
- gaming;
- cloud/CDN;
- search;
- software/update;
- advertising;
- adult;
- gambling;
- unknown.

## Hard constraints

- no TLS MITM;
- no payload/content inspection;
- classification must never be presented as certain when evidence is weak;
- conflicting providers/evidence must remain explainable;
- no Policy Engine yet;
- no blocking yet;
- no DNS UI yet;
- enrichment failure must never affect Xray or panel core;
- analytics OFF = exact no-op;
- minimal upstream-touch points;
- `make verify-fork` must remain green.

## Required tests

- exact domain classification;
- subdomain inheritance where appropriate;
- conflicting classification;
- ASN/GeoIP enrichment;
- cache expiry;
- unknown fallback;
- confidence/provenance;
- SQLite/PostgreSQL where applicable;
- concurrency/race where relevant.

## Restrictions

- do not modify `main`;
- do not start WP-2C;
- do not start the Policy Engine;
- do not merge WP-2B into `develop` until its merge gate is satisfied.
