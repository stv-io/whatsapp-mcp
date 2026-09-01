# 0005 — Carrying the DNS-rebinding fix

**Status:** accepted, expected to be temporary

## Decision

This fork patches `whatsapp-mcp-server/{main,mcp_config}.py` so the MCP SDK's
DNS-rebinding allowlist reflects the host the server actually binds to,
configured via `WHATSAPP_MCP_ALLOWED_HOSTS`.

## The bug

The SDK enables DNS-rebinding protection automatically when `FastMCP` is
constructed with the default host, pinning `allowed_hosts` to loopback:

```python
if transport_security is None and host in ("127.0.0.1", "localhost", "::1"):
    transport_security = TransportSecuritySettings(
        enable_dns_rebinding_protection=True,
        allowed_hosts=["127.0.0.1:*", "localhost:*", "[::1]:*"], ...)
```

`main.py` constructs the server at import time — deliberately, to keep imports
free of side effects — and applies `WHATSAPP_MCP_HOST` afterwards in
`__main__`. Mutating `settings.host` does not revisit the allowlist, so a
server listening on `0.0.0.0` still carries a loopback-only policy and answers
**421 Misdirected Request** to every caller that addresses it by any other
name. Remote transport is the reason this fork's parent exists, so in practice
the HTTP transport only worked from localhost.

## The fix

Re-derive the policy in `__main__`, next to where host and port are already
applied:

- `WHATSAPP_MCP_ALLOWED_HOSTS` set → protection on, with those hosts. Entries
  are `host:port`, or `host:*` for any port.
- Unset and bound to a non-loopback address → protection off, because the
  operator has deliberately opened the server up and a loopback-only allowlist
  can then only lock them out.
- Unset and bound to loopback → untouched; the SDK's default is correct.

Keeping the allowlist configurable matters: an operator who binds `0.0.0.0`
should not have to choose between "unreachable" and "protection off".

## Why here and not a shim

The previous workaround was a `sitecustomize.py` monkey-patching
`FastMCP.__init__` from `site-packages` — invisible, and fragile against SDK
changes. As a patch it is also the form that can be offered upstream, which it
has been: [verygoodplugins/whatsapp-mcp#216](https://github.com/verygoodplugins/whatsapp-mcp/issues/216).

## Delete this when

Upstream merges a fix. Then drop the patch, drop this ADR, and switch to
whatever environment variable upstream settles on.
