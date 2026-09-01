# 0003 — Package in a fork, not a separate repo

**Status:** accepted

## Decision

Container packaging lives in a fork of `verygoodplugins/whatsapp-mcp`, not in a
standalone repo that pulls the project in at build time.

## Why

The alternative — a thin `whatsapp-mcp-image` repo cloning a pinned upstream ref
inside the Dockerfile — was how this started, and it had three problems:

1. **The build was not hermetic.** It ran `git clone` mid-build, so a GitHub
   blip failed the build and the pinned SHA was the only thing between us and a
   moving target.
2. **Pinning a SHA assumed upstream history is stable. It is not.** The ref we
   pinned (`7f518e2`) is no longer reachable from upstream `main`: it and the
   current `f5de02d` are both "chore(main): release 0.6.0" with an **identical
   tree and committer date but different SHAs**, so `main` was rewritten at some
   point. Content was unchanged and nothing broke, but a pin that can silently
   detach from its branch is not a pin. In a fork we control the history.
3. **The code and its packaging drifted apart.** The Dockerfile hand-mirrored
   the project's Python dependencies instead of installing from its
   `pyproject.toml`. Upstream tightened `mcp[cli]` to `<2`; the copy did not;
   `mcp` 2.0 removed the module `main.py` imports on its first line, and the
   container crash-looped 1878 times before anyone connected the two.

In-tree, the build context is the repository, the dependency pins come from the
project's own metadata, and the transport fix is a real patch to `main.py`
rather than a shim wedged into `site-packages` — which is also the form it can
be offered upstream in.

## Consequences

- Periodic rebase onto upstream. The delta is small: a Dockerfile, an
  entrypoint, one workflow, some docs, and one patch.
- The fork carries the whole project, so issues arriving here may be about the
  application rather than the packaging. `NOTICE.md` redirects them.

## Revisit if

Upstream accepts the Dockerfile and the transport fix, at which point the fork
has nothing left to carry and should be retired in favour of upstream's own
image.
