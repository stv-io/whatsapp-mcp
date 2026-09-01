# 0001 — Bridge and MCP server share one container

**Status:** accepted

## Decision

`whatsapp-bridge` (Go) and the Python MCP server run as two processes in one
container, supervised by `entrypoint.sh`, rather than as two containers.

## Why

They are not independently useful. The bridge owns the WhatsApp session and
writes `store/{whatsapp,messages}.db`; the server does nothing but read those
files and expose them as MCP tools. Splitting them would mean a shared volume
and a startup ordering dance, in exchange for isolation neither one can use.

`entrypoint.sh` deliberately takes the **whole container down** when either
process exits — it signals PID 1 (tini) — so the orchestrator restarts the
pair. Partial liveness here is worse than none: a bridge with no server is
unreachable, and a server with no bridge serves a stale database.

## Consequences

- A crash in either process shows up as a container restart. That is intended,
  but it does mean a fault in one obscures the other: a Python import error
  once killed the bridge mid-QR-pairing every few seconds, which read as a
  pairing problem for weeks. Read the logs of both processes.
- No per-process resource limits.

## Revisit if

The server grows independent value — serving a database populated by something
other than this bridge, or needing to scale separately.
