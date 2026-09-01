#!/bin/sh
# Run the Go bridge and the Python MCP server in one container.
#
# The bridge owns the WhatsApp session and the SQLite files; the MCP server
# reads them and serves tools over streamable-HTTP. Neither is useful without
# the other, so if either exits we bring the whole container down and let the
# orchestrator restart the pair. See docs/adr/0001-single-container.md.
set -eu

mkdir -p /data/store

# The bridge writes its DBs to ./store/, so run it from /data.
cd /data
whatsapp-bridge &
BRIDGE=$!

# If the bridge exits, signal PID 1 (tini) to tear the container down.
( wait "$BRIDGE"; echo "whatsapp-bridge exited — stopping container"; kill -TERM 1 2>/dev/null || true ) &

# Give the bridge a moment to create ./store before the server opens the DBs.
sleep 3

cd /app/whatsapp-mcp-server
exec python main.py
