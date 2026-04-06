#!/bin/bash
set -euo pipefail

if [[ -z "${ANDUIN_MCP_URL:-}" ]]; then
  cat <<'MSG'
Anduin plugin: No server configured. Run /anduin:setup to connect to your Anduin environment.
MSG
fi
