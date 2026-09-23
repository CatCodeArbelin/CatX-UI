# ADR-0001: Maintained Downstream Fork

Status: Accepted

## Context

Custom features must coexist with frequent upstream 3x-ui releases.

## Decision

Maintain a downstream fork with minimal integration hooks and isolated custom modules.

Production follows upstream stable tags.

## Consequences

Positive:

- easier upstream sync;
- clear custom ownership;
- feature isolation;
- removable modules.

Negative:

- a few permanent hook points must be maintained;
- some features may be postponed if they require invasive upstream changes.
