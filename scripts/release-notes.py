#!/usr/bin/env python3
"""Prints GitHub Release notes: paste-ready refs per region plus this version's changelog section."""
import re
import sys

version, us_sha, eu_sha = sys.argv[1:4]
changelog = open("CHANGELOG.md").read()
section = re.search(rf"^## {re.escape(version)}\b.*?\n(.*?)(?=^## |\Z)", changelog, re.S | re.M)
if not section:
    sys.exit(f"CHANGELOG.md has no '## {version}' section")

print(f"""Organization admins: set the `ref` line of your Anduin marketplace entry to the value for your region.

| Region | `ref` line | Commit (optional `sha` pin) |
|---|---|---|
| Production (US) | `"ref": "v{version}"` | `{us_sha}` |
| Production (EU) | `"ref": "v{version}-eu"` | `{eu_sha}` |

No GitHub? Upload `anduin-us-{version}.zip` or `anduin-eu-{version}.zip` from the assets below.

## Changes

{section.group(1).strip()}""")
