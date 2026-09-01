# Architecture decision records

Short records of decisions that are not obvious from the code, and that someone
will otherwise re-litigate. Each states the decision, why, and what would make
us revisit it.

| ADR | Decision |
|---|---|
| [0001](0001-single-container.md) | Bridge and MCP server share one container |
| [0002](0002-amd64-only.md) | `linux/amd64` only |
| [0003](0003-packaging-fork.md) | Package in a fork, not a separate repo |
| [0004](0004-image-tags.md) | Tag convention and immutability |
| [0005](0005-transport-security-patch.md) | Carrying the DNS-rebinding fix |
