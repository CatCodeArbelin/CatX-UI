# ADR-0002: No TLS MITM

Status: Accepted

## Decision

The product will not decrypt HTTPS and will not distribute trusted root CAs for traffic inspection.

## Allowed approach

Use DNS, destination metadata, SNI where visible, protocol metadata, IP/ASN/GeoIP, traffic counters, and session correlation.

## Reason

TLS MITM introduces disproportionate privacy, security, liability, and misuse risk.
