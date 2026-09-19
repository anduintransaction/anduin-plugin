#!/usr/bin/env bash
# Renders docs/Anduin_MCP_Enablement_Guide_Claude.html to a Letter PDF (gitignored) for sending to customers.
# Needs Chrome or Chromium; set CHROME=/path/to/binary if it is not found. Optional arg: output path.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
src="$root/docs/Anduin_MCP_Enablement_Guide_Claude.html"
out="${1:-${src%.html}.pdf}"

chrome="${CHROME:-}"
for c in "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" google-chrome chromium chromium-browser; do
  [[ -n "$chrome" ]] && break
  if [[ -x "$c" ]] || command -v "$c" >/dev/null 2>&1; then chrome="$c"; fi
done
if [[ -z "$chrome" ]]; then
  echo "Chrome/Chromium not found; set CHROME=/path/to/binary" >&2
  exit 1
fi

"$chrome" --headless=new --disable-gpu --no-pdf-header-footer --print-to-pdf="$out" "file://$src" 2>/dev/null
echo "wrote $out"
