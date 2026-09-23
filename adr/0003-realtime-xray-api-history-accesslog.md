# ADR-0003: Xray API for Real-time, access.log for History

Status: Accepted

## Decision

Use Xray online/stat APIs as the primary real-time source where available.

Use access.log as one analytics/history source.

## Reason

Modern upstream already moved real-time IP tracking away from access.log. Re-scanning logs for online state duplicates upstream and scales poorly.
