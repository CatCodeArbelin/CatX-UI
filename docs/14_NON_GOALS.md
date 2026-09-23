# Non-goals

These exclusions protect maintainability and privacy.

## 1. No TLS MITM

No:

- certificate distribution workflow;
- trusted root CA;
- TLS interception proxy;
- HTTPS decryption;
- decrypted request/response harvesting.

## 2. No Remnawave architecture transplant

Do not redesign 3x-ui around:

- Squads;
- External Squads;
- mandatory Config Profiles;
- replacement subscription control plane;
- replacement node/control-plane model.

## 3. No rewrite

This is not "4x-ui".

Do not rewrite backend/frontend.

## 4. No forked xray-core

Panel features must not require maintaining a custom Xray fork.

## 5. No mandatory external infrastructure

Do not require Redis/Kafka/ClickHouse/Elastic for baseline fork operation.

Optional integrations may exist later.

## 6. No billing platform in core

No payments, invoicing, reseller CRM, or crypto billing in core roadmap.

Expose APIs/webhooks instead.

## 7. No surveillance escalation

Metadata analytics must not evolve into content inspection.

## 8. No automatic punitive action by default

Risk engine informs.

Blocking/ban requires explicit policy.

## 9. No duplicate upstream features

If upstream implements a feature well, remove/deprecate the fork duplicate.

## 10. No unnecessary framework building

A small extension registry is enough.

Do not build a general plugin system before needed.
