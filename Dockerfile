# Container image for whatsapp-mcp: the Go/whatsmeow bridge and the Python
# FastMCP server in one image, supervised by entrypoint.sh.
#
# Upstream ships no container, so this is packaging added by the fork. Built
# from the repository itself — nothing is fetched from git at build time, so a
# build is reproducible from a checkout alone and cannot drift under us.
#
# linux/amd64 only: the bridge needs CGO for go-sqlite3, and upstream's own
# release binaries are amd64-only. See docs/adr/0002-amd64-only.md.

# ---- build the Go bridge (whatsmeow; CGO for go-sqlite3) ----
FROM golang:1.25-bookworm@sha256:3b4a11519ad929d1e1d261a12cff056f0c85b735253d7d861346b9c6f8b36437 AS bridge-build

WORKDIR /src/whatsapp-bridge

# Dependencies first: this layer is cached unless go.mod/go.sum change, which
# is what makes an app-only rebuild cheap.
COPY whatsapp-bridge/go.mod whatsapp-bridge/go.sum ./
RUN go mod download

COPY whatsapp-bridge/ ./
ENV CGO_ENABLED=1
RUN go build -trimpath -ldflags="-s -w" -o /out/whatsapp-bridge .

# ---- runtime: glibc/bookworm to match the CGO build ----
FROM python:3.11-slim-bookworm@sha256:528257d48c1da0dcecc2e725d1ae34498d60c965f1241e39cd6a85a8859bdf84

# tini reaps zombies and forwards signals; entrypoint.sh relies on being able to
# signal PID 1 to bring the container down when either process dies.
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tini \
 && rm -rf /var/lib/apt/lists/*

COPY --from=bridge-build /out/whatsapp-bridge /usr/local/bin/whatsapp-bridge
COPY whatsapp-mcp-server /app/whatsapp-mcp-server

# Install from the source tree so the dependency pins come from the project's
# own pyproject.toml. Do NOT re-list them here: a hand-mirrored copy once
# drifted to an unbounded `mcp[cli]`, picked up mcp 2.x, and every start died on
# `ModuleNotFoundError: mcp.server.fastmcp`.
RUN pip install --no-cache-dir --no-compile /app/whatsapp-mcp-server \
 && find /usr/local/lib/python3.11 -name '__pycache__' -type d -prune -exec rm -rf {} +

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# The bridge writes ./store/{messages.db,whatsapp.db}; entrypoint.sh cds to
# /data so that becomes /data/store/. The Python server reads the same files via
# the *_DB_PATH vars. All are overridable at run time.
ENV WHATSAPP_DB_PATH=/data/store/messages.db \
    WHATSMEOW_DB_PATH=/data/store/whatsapp.db \
    WHATSAPP_MCP_TRANSPORT=http \
    WHATSAPP_MCP_HOST=0.0.0.0 \
    WHATSAPP_MCP_PORT=8000 \
    WHATSAPP_BRIDGE_PORT=8080 \
    HOME=/data

VOLUME ["/data"]
EXPOSE 8000
ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]
