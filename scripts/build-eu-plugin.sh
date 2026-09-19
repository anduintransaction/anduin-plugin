#!/usr/bin/env bash
# Turns a copy of plugins/anduin (US) into the EU build: only the MCP URL and version differ.
# Usage: scripts/build-eu-plugin.sh <target-dir>   e.g. /tmp/anduin-eu, or plugins/anduin for in place (CI)
#        scripts/build-eu-plugin.sh --check        verify marketplace sources, README and CHANGELOG
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
us="$root/plugins/anduin"

if [[ $# -ne 1 || ( "$1" == -* && "$1" != "--check" ) ]]; then
  echo "Usage: $0 <target-dir> | --check" >&2
  exit 2
fi

if [[ "$1" == "--check" ]]; then
  python3 - "$root" <<'PY'
import json, re, sys
root = sys.argv[1]
want = json.load(open(f"{root}/plugins/anduin/.claude-plugin/plugin.json"))["version"]
if not re.fullmatch(r"\d+\.\d+\.\d+", want):
    sys.exit(f"plugin.json version {want!r} must be a plain X.Y.Z release version")
entries = json.load(open(f"{root}/.claude-plugin/marketplace.json"))["plugins"]
names = [e["name"] for e in entries]
if len(names) != len(set(names)):
    sys.exit(f"duplicate marketplace entry names: {names}")
# plugin.json is the only version source; a marketplace `version` would be silently ignored.
if any("version" in e for e in entries):
    sys.exit("remove `version` from marketplace.json entries; plugin.json is the version source")
src = {e["name"]: e.get("source") for e in entries}
eu = {"source": "git-subdir", "url": "https://github.com/anduintransaction/anduin-plugin.git",
      "path": "plugins/anduin", "ref": f"v{want}-eu"}
if src != {"anduin": "./plugins/anduin", "anduin-eu": eu}:
    sys.exit(f"marketplace.json sources must be ./plugins/anduin and {eu}, got {src}")
if json.load(open(f"{root}/plugins/anduin/.claude-plugin/plugin.json"))["name"] != "anduin":
    sys.exit("plugin.json name must stay `anduin` (agents allow mcp__plugin_anduin_anduin__*)")
if list(json.load(open(f"{root}/plugins/anduin/.mcp.json"))["mcpServers"]) != ["anduin"]:
    sys.exit(".mcp.json must define exactly one server named `anduin`")
# README and CHANGELOG must name this release.
readme = open(f"{root}/README.md").read()
if f'"ref": "v{want}"' not in readme or f'"ref": "v{want}-eu"' not in readme:
    sys.exit(f"README.md organization example must use v{want} and v{want}-eu")
if not re.search(rf"^## {re.escape(want)}\b", open(f"{root}/CHANGELOG.md").read(), re.M):
    sys.exit(f"CHANGELOG.md needs a '## {want}' section")
print(f"release metadata consistent for {want} / {want}-eu")
PY
  exit 0
fi

target="$(python3 -c 'import os,sys; print(os.path.realpath(sys.argv[1]))' "$1")"
if [[ "$target" != "$us" ]]; then
  case "$us/" in "$target"/*) echo "target $target contains the source plugin; refusing" >&2; exit 2 ;; esac
  case "$target" in "$us"/*) echo "target $target is inside the source plugin; refusing" >&2; exit 2 ;; esac
fi

# Stage next to the destination so the final rename stays on one filesystem.
parent="$(dirname "$target")"
mkdir -p "$parent"
tmp="$(mktemp -d "$parent/.anduin-eu-build.XXXXXX")"
trap 'rm -rf "$tmp"' EXIT
cp -R "$us/." "$tmp/"
find "$tmp" -name .DS_Store -delete
python3 - "$tmp" <<'PY'
import json, re, sys
d = sys.argv[1]
us = {"anduin": {"type": "http", "url": "https://mcp.anduin.app/mcp"}}
cfg = json.load(open(f"{d}/.mcp.json"))
if cfg != {"mcpServers": us}:
    sys.exit("plugins/anduin/.mcp.json must hold exactly the US `anduin` server")
cfg["mcpServers"]["anduin"]["url"] = "https://mcp.eu.anduin.app/mcp"
open(f"{d}/.mcp.json", "w").write(json.dumps(cfg, indent=2) + "\n")
p = f"{d}/.claude-plugin/plugin.json"
raw = open(p).read()
ver = json.loads(raw)["version"]
if not re.fullmatch(r"\d+\.\d+\.\d+", ver):
    sys.exit(f"plugin.json version {ver!r} must be a plain X.Y.Z release version")
pat = re.compile(r'("version"\s*:\s*")' + re.escape(ver) + '"')
if len(pat.findall(raw)) != 1:
    sys.exit("could not locate exactly one version field in plugin.json")
out = pat.sub(lambda m: m.group(1) + ver + '-eu"', raw)
if json.loads(out)["version"] != ver + "-eu":
    sys.exit("version rewrite produced unexpected plugin.json")
open(p, "w").write(out)
PY

# Swap in the finished build; restore the previous target if the swap fails midway.
backup=""
if [[ -e "$target" ]]; then
  backup="$parent/.anduin-eu-backup.$$"
  mv "$target" "$backup"
fi
if ! mv "$tmp" "$target"; then
  [[ -n "$backup" ]] && mv "$backup" "$target"
  echo "failed to place EU build at $target" >&2
  exit 1
fi
trap - EXIT
[[ -n "$backup" ]] && rm -rf "$backup"
echo "EU build written to $target"
