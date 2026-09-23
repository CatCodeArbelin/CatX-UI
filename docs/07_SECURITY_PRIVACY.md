# Security and Privacy

## Product boundary

This fork is a metadata-oriented network control and analytics tool.

It is not a TLS interception product.

## Never implement

- HTTPS MITM;
- root CA distribution for inspection;
- decrypted request/response collection;
- cookies;
- Authorization headers;
- password/form capture;
- decrypted message body capture.

## Allowed metadata

Depending on source:

```text
client
source IP
destination IP
domain
DNS query
SNI when visible
port
TCP/UDP
TLS/QUIC/HTTP classification
inbound/outbound
rule/policy decision
bytes
first/last seen
ASN
country
```

## Visibility limitations

ECH may hide SNI.

Client-controlled DoH/DoT/cached DNS may bypass observed DNS.

Therefore combine multiple evidence sources and preserve confidence.

## Risk engine

Signals are evidence, not proof.

Bad:

```text
4 IPs = account sharing = ban
```

Good:

```text
Risk 78
- 4 concurrent IPs
- 3 ASNs
- new country
- datacenter ASN
```

## Sensitive logging

Never log:

- auth tokens;
- passwords;
- subscription secrets unnecessarily;
- private keys;
- full sensitive request headers.

## Self-service portal

Sensitive actions require:

- authorization;
- confirmation;
- rate limit;
- audit;
- cooldown where appropriate.

## DNS policy

Do not claim perfect DoH filtering.

## Data deletion

Support retention and manual analytics cleanup without damaging upstream accounting.

## Mandatory security review

Required when touching:

- updater;
- authentication;
- subscription token;
- self-service;
- shell execution;
- node remote commands;
- file paths;
- backup restore;
- privileged traffic shaping.
