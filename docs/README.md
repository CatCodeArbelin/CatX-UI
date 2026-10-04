# CatX-UI repository documentation

This directory contains inherited 3x-ui documentation-site source plus the
canonical CatX planning, architecture, release, and operations documents. The
site material originated in the upstream `MHSanaei/3x-ui` project and may
describe upstream behavior; it is retained for low-divergence compatibility
and is not presented here as an official CatX documentation website.

For the current public product contract, start with the [CatX-UI root
README](../README.md), [stable v0.2.0 release notes](32_V0_2_0_RELEASE_NOTES.md),
[CatX Sponsors contract](33_CATX_SPONSORS.md),
[contributing guide](../CONTRIBUTING.md), and [security policy](../SECURITY.md).
CatX installs and updates come only from
[CatCodeArbelin/CatX-UI](https://github.com/CatCodeArbelin/CatX-UI); do not use
upstream install commands or images for CatX deployments.

## Inherited documentation site

`content/`, `app/`, `components/`, and the site tooling preserve the upstream
documentation structure and translations. Any public deployment of this site
must be labeled as inherited/upstream-derived until a separately approved CatX
documentation deployment exists. No independent CatX public URL is claimed by
this repository.

The inherited pages cover installation, first login, Xray configuration,
subscriptions, operations, API reference, troubleshooting, and database
topics. CatX-specific claims take precedence from the root README and the
stable release notes; capability-dependent features must not be advertised as
universal upstream behavior.

## Local development

```bash
cd docs
pnpm install
pnpm dev
pnpm build
pnpm typecheck
pnpm lint
pnpm test
```

The documentation source and its GPL notices remain available under this
directory. Upstream attribution is intentional and does not imply endorsement.
