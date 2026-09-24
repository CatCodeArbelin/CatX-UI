# CURRENT TASK

## Work Package

`WP-1C — Client Activity API & UI`

## Goal

Expose the analytics produced by WP-1A/WP-1B in the panel through a read-only
client activity API and UI.

## Scope

- read-only analytics API endpoints;
- per-client recent activity timeline;
- session history;
- destination/domain display when known;
- protocol, port, and source metadata;
- first seen, last seen, and session counts;
- traffic metadata only when reusable from existing upstream counters;
- pagination and time-range filtering;
- empty, loading, and error states;
- analytics-disabled UI behavior;
- frontend navigation through the WP-0A fork registry where appropriate;
- backend/API tests and frontend tests.

- reuse the WP-1A repositories;
- do not create another analytics persistence layer;
- do not duplicate upstream traffic counters;
- do not add service/category classification; WP-2 owns enrichment;
- raw domains/IPs may be displayed only when observed;
- expose provenance/source where useful;
- no TLS MITM or content inspection;
- analytics failures must not affect Xray or panel core operation;
- do not modify `main`;
- do not merge WP-1C into `develop`.
