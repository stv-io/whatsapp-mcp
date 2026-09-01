# 0004 — Tag convention and immutability

**Status:** accepted

## Decision

Images are tagged `<upstream-version>-<packaging-revision>`:

```
ghcr.io/stv-io/whatsapp-mcp:0.6.0-1    first image built from upstream 0.6.0
ghcr.io/stv-io/whatsapp-mcp:0.6.0-2    same application, packaging changed
ghcr.io/stv-io/whatsapp-mcp:0.7.0-1    upstream moved
```

Every build also gets `sha-<short-git-sha>`, which is what CI reports.

Rules:

- **No `latest`.** It is the tag that makes "what is running in production"
  unanswerable.
- **A published tag is never re-pushed.**
- **Consumers pin the digest.** The tag is human-readable metadata; the digest
  is the reference.

## Why this shape

The application version and the image version are different things: fixing the
Dockerfile rebuilds identical application code. One number cannot carry both
without a judgement call every time about whether packaging counts as a "patch"
release. Debian's `upstream-revision` convention already solves this, sorts
correctly, and reads unambiguously.

## On immutability

GHCR does **not** enforce tag immutability — nothing stops a push from moving a
published tag. So the convention is enforced where it can be: before pushing a
tag, CI queries the registry and **fails if that tag already exists**. That
turns an agreement into something that cannot happen by accident.

Digest pinning downstream is the real guarantee. The registry cannot lie about
a digest; it can absolutely lie about a tag.

## Keeping the upstream version honest

The convention only works while the left-hand number really is upstream's
release. Upstream runs `release-please`, which would happily bump this fork's
version on our own conventional commits and make that number a lie. Both
release workflows are therefore guarded with
`if: github.repository == 'verygoodplugins/whatsapp-mcp'` — guarded rather than
deleted, so they rebase cleanly and still work upstream.

## Consequences

- Bumping upstream means resetting the revision to `1`, not incrementing it.
- Re-releasing an identical build under a new tag is impossible by design — if
  a tag was wrong, the next revision fixes it and the wrong one is left as
  history.

## Revisit if

GHCR gains real immutable tags, at which point the CI guard becomes redundant
and should be replaced rather than kept alongside.
