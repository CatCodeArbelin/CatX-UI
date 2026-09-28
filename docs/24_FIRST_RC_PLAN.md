# First public CatX-UI RC plan

This plan deliberately creates no tag or GitHub release. The first public RC
is proposed only after the exact RC-3 commit passes all gates.

## Proposed identity

- base version: 0.1.0
- prerelease version: 0.1.0-rc.1
- tag: v0.1.0-rc.1
- title: CatX-UI v0.1.0-rc.1 — First Public Release Candidate
- channel: explicit rc

The RC version is checked in separately from the stable base version. The
stable updater continues to resolve only releases/latest and rejects
prerelease metadata. RC updates resolve the exact RC tag, while dev-latest
continues to be the opt-in rolling development channel.

The RC release must be marked prerelease=true and latest=false. It must never
move the GitHub releases/latest pointer.

## Planned release body

Use the verified contents of docs/23_RC2_RELEASE_NOTES.md, the exact upstream
base version, migration notes, rollback instructions, known limitations, and
the final CI/staging evidence. Do not copy an unverified checklist into the
public release.

## Artifact list

Publish only assets emitted by the release workflow:

- catx-ui-linux-amd64.tar.gz
- catx-ui-linux-arm64.tar.gz
- catx-ui-linux-armv7.tar.gz
- catx-ui-linux-armv6.tar.gz
- catx-ui-linux-386.tar.gz
- catx-ui-linux-armv5.tar.gz
- catx-ui-linux-s390x.tar.gz
- catx-ui-windows-amd64.zip
- the CatX installer/updater control assets and their checksum sidecars

The final publication step must include the checksum list generated from the
exact release assets and must be preceded by the release identity and artifact
qualification jobs.

The release metadata must identify channel=rc, releaseVersion=0.1.0-rc.1,
releaseTag=v0.1.0-rc.1, prerelease=true, and latest=false. Every archive,
control asset, and checksum sidecar must come from that same tagged workflow
run.

## Upgrade and rollback text

Tell operators to back up the database and installation directory, use the
CatX installer/update path, wait for panel and Xray health checks, and retain
the update evidence (runId, state, exit code, finish time, rollback flags).
If rollback is unhealthy, stop further retries and use the recorded recovery
state rather than manually replacing files on a production node.

## Final public cut

After RC-3 is green on one exact SHA:

1. create the tag v0.1.0-rc.1 at that exact SHA;
2. push the tag to origin;
3. wait for the release workflow and its artifact, updater, Linux staging, and
   PostgreSQL qualification jobs;
4. confirm the release is prerelease and not latest;
5. publish only the workflow-emitted assets and their sidecars.

Do not retarget the tag, edit artifacts manually, promote the RC to latest, or
update production as part of the RC cut.
