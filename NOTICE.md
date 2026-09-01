# Notice

This project is MIT licensed. It carries three layers of authorship, and the
`LICENSE` file records all three.

## Original work

[**lharries/whatsapp-mcp**](https://github.com/lharries/whatsapp-mcp) — created
by [Luke Harries](https://github.com/lharries). The WhatsApp MCP server, the
whatsmeow bridge and the tool surface all originate here.
_Copyright (c) 2025 Luke Harries._

## Maintained fork

[**verygoodplugins/whatsapp-mcp**](https://github.com/verygoodplugins/whatsapp-mcp)
— maintained by [Very Good Plugins](https://verygoodplugins.com/). Continued
maintenance after the original went quiet: remote transport support,
configurable database paths, on-demand history sync, read-state tracking,
pairing fixes and much else. Essentially everything this fork builds on.
_Copyright (c) 2026 Very Good Plugins._

## This fork

[**stv-io/whatsapp-mcp**](https://github.com/stv-io/whatsapp-mcp) — a
**packaging fork**. It adds a container image and the pipeline that builds,
scans, signs and publishes it, plus one fix to the HTTP transport. It makes no
attempt to fork the project's direction.
_Copyright (c) 2026 Stephen Attard._

Additions in this fork:

| Path | What |
|---|---|
| `Dockerfile`, `.dockerignore`, `entrypoint.sh` | Container packaging |
| `.github/workflows/container.yml` | Build, smoke test, scan, sign, publish |
| `docs/adr/` | Architecture decision records |
| `whatsapp-mcp-server/{main,mcp_config}.py` | DNS-rebinding allowlist fix, offered upstream as [verygoodplugins/whatsapp-mcp#216](https://github.com/verygoodplugins/whatsapp-mcp/issues/216) |

**Bugs in the application itself belong upstream**, not here. Please report
them at
[verygoodplugins/whatsapp-mcp/issues](https://github.com/verygoodplugins/whatsapp-mcp/issues).
Issues here should be about packaging: the image, the pipeline, the transport
fix.
