# 0002 — `linux/amd64` only

**Status:** accepted

## Decision

The image is published for `linux/amd64` alone. No multi-arch manifest.

## Why

- The bridge builds with `CGO_ENABLED=1` for `go-sqlite3`. Cross-compiling CGO
  needs a C toolchain per target; building arm64 on GitHub's amd64 runners
  means QEMU emulating a C compile, which is slow and a recurring source of
  flakiness.
- Upstream's own release workflow ships `GOOS=linux GOARCH=amd64` binaries
  only, so amd64 is the arch the project actually tests.
- The only consumer today is an amd64 mini PC.

Publishing an arm64 image nobody runs is a build we would have to keep green
for no one.

## Consequences

- Apple Silicon developers must build locally or run under emulation.
- A future arm64 host cannot pull this image.

## Revisit if

A consumer is arm64 — most plausibly a replacement homelab box. The change is
adding `linux/arm64` to `platforms:`, plus either a cross-toolchain in the
builder stage or patience with QEMU.
