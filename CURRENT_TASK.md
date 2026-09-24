# CURRENT TASK

## Work Package

`WP-1B — access.log & Session Pipeline`

## Goal

Implement the metadata-only access.log and session ingestion pipeline on top of
the WP-1A analytics foundation.

## Scope

- incremental access.log tailer;
- persistent cursor/offset;
- file identity/inode handling;
- log rotation and truncation;
- partial-line handling;
- bounded queues and backpressure;
- batching;
- graceful shutdown;
- normalization into the WP-1A DestinationEvent model;
- client/session correlation;
- deduplication;
- parser fuzz tests;
- race/concurrency tests;
- analytics OFF = exact no-op;
- ingestion failure must never break Xray/panel;
- no TLS MITM and no content capture.

Do not implement WP-1C or unrelated policy, DNS collection, or UI work in this
package.
