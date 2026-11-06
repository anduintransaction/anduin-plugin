#!/usr/bin/env python3
"""Prints GitHub Release notes: paste-ready refs per region plus this version's changelog section."""
import re
import sys

if len(sys.argv) not in (4, 5) or sys.argv[4:] not in ([], ["--immutable"]):
    sys.exit(f"usage: {sys.argv[0]} <version> <us-sha> <eu-sha> [--immutable]")
version, us_sha, eu_sha = sys.argv[1:4]
# --immutable: the release is published and GitHub has signed its attestation, so tell customers how to verify.
immutable = sys.argv[4:] == ["--immutable"]
changelog = open("CHANGELOG.md").read()
section = re.search(rf"^## {re.escape(version)}\b.*?\n(.*?)(?=^## |\Z)", changelog, re.S | re.M)
if not section:
    sys.exit(f"CHANGELOG.md has no '## {version}' section")

verify = (
    "\nThis release is immutable: GitHub locked its tag and assets when it was published and signed an attestation for them. "
    f"Check a downloaded file with `gh release verify-asset v{version} <file> --repo anduintransaction/anduin-plugin`.\n"
    if immutable else ""
)

print(f"""Organization admins: set the `ref` line of your Anduin marketplace entry to the value for your region.

| Region | `ref` line | Commit (optional `sha` pin) |
|---|---|---|
| Production (US) | `"ref": "v{version}"` | `{us_sha}` |
| Production (EU) | `"ref": "v{version}-eu"` | `{eu_sha}` |

No GitHub? Upload `anduin-us-{version}.zip` or `anduin-eu-{version}.zip` from the assets below. The admin guide for this release is `Anduin_MCP_Enablement_Guide_Claude.pdf`, also below.
{verify}
## Changes

{section.group(1).strip()}""")
