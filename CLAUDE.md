# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A Claude Code plugin that provides AI assistants for Anduin platform domains:
- **GP Assistant** — fund subscription management (LP review, forms, AML/KYC, dashboards, tagging)
- **Data Room Agent** — virtual data room management (rooms, participants, files, analytics)

The plugin has one source folder, `plugins/anduin`, hardcoded to Production US. Each release tag `vX.Y.Z` also gets a CI-generated, immutable `vX.Y.Z-eu` tag whose copy points at Production EU. Users install exactly one region. Authentication is OAuth2. It works on both Claude Code (CLI) and Cowork (desktop app) with no setup required.

## Local Development

```bash
# Load the working tree for testing (run from repo root)
claude --plugin-dir ./plugins/anduin

# Test the EU build locally
scripts/build-eu-plugin.sh /tmp/anduin-eu && claude --plugin-dir /tmp/anduin-eu

# Install from marketplace (for users)
/plugin marketplace add anduintransaction/anduin-plugin
/plugin install anduin@anduin-marketplace
```

## Testing Changes

For content changes, verify direct skill invocation, automatic agent activation, permission handling, and
widget/text fallback in the supported hosts. Verify that adapters stop without calling domain tools if the
canonical skill cannot load. Structural validation does not replace installation, OAuth, or renderer tests.
Use an explicitly authorized test account for write checks; do not execute synthetic writes against the default
production connection.

For an authorized host smoke test, verify manually:

```bash
# Load the working tree, then /reload-plugins after further edits
claude --plugin-dir ./plugins/anduin

# Verify skills load
/anduin:gp-assistant       # Should load GP domain knowledge
/anduin:dataroom           # Should load Data Room domain knowledge

# Verify MCP connects
# Check /mcp — "anduin" should appear pointing to Production US
# (Production EU when testing an EU build)
```

`--plugin-dir` loading does not exercise marketplace caching or updates; test those with a marketplace install.

## Plugin Architecture

This is a **content-only plugin** — no runtime build step or application dependencies. Runtime files are Markdown
or JSON. Keep repository documentation customer-facing; internal plans, reviews, and test-run records belong
outside this repository.

```
.claude-plugin/marketplace.json       — Marketplace manifest (lists all plugins)

plugins/anduin/                       — Anduin platform plugin
  .claude-plugin/plugin.json          — Plugin manifest (name, version, description)
  .mcp.json                           — MCP server config (hardcoded to Production US)
  agents/                             — Thin Claude adapters (spawned as subagents)
    gp-assistant.md                   — Fund subscription agent (model: sonnet, tools: mcp__plugin_anduin_anduin__*)
    dataroom-agent.md                 — Data room agent (model: sonnet, tools: mcp__plugin_anduin_anduin__*)
  skills/                             — Canonical behavior loaded into context on demand
    gp-assistant/SKILL.md             — GP workflows, terminology, permissions, and safety
    dataroom/SKILL.md                 — Data Room workflows, terminology, permissions, and safety

scripts/build-eu-plugin.sh            — Rewrites a copy of plugins/anduin for EU (URL + `-eu` version); `--check` verifies marketplace
                                        sources and that the guide, README and CHANGELOG name the current release
scripts/release-notes.py              — Writes the GitHub Release notes from CHANGELOG.md
scripts/render-guide.sh               — Renders the admin guide to docs/*.pdf (gitignored) with headless Chrome

.github/workflows/check.yml           — PR/push: `--check` and a trial EU build
.github/workflows/release.yml         — On tag vX.Y.Z: validates tag == plugin.json version, commits the EU build, tags vX.Y.Z-eu
                                        (no-op if that tag exists and holds the EU build; fails if it exists from a different
                                        commit), then publishes a GitHub Release with both refs and one ZIP per region

docs/Anduin_MCP_Enablement_Guide_Claude.html — Customer guide for organization admins (source of the PDF we send)
CHANGELOG.md                          — One `## X.Y.Z` section per release (required by `--check`)
```

**Key patterns:**
- Each domain has one canonical skill and a thin Claude adapter. The adapter preserves activation, model, and tool
  access, then invokes `anduin:gp-assistant` or `anduin:dataroom` with the already-allowed `Skill` tool before domain
  work. If loading fails, it stops rather than running without the shared rules. Keep workflows, permission rules,
  confirmation, recovery, and presentation guidance in the skill, not a second copy in the adapter.
- Canonical skills use wire names (`dr_` for Data Room, unprefixed for fund subscription). Resolve them against the
  current host's provided Anduin catalog; do not copy Claude-specific MCP prefixes into shared behavior.
- Regional builds: `main` holds only the US plugin. The EU build is never committed to `main`; CI derives it on
  each release tag. Both keep plugin.json name `anduin` and server name `anduin`, so skills (`anduin:*`) and the
  tool prefix (`mcp__plugin_anduin_anduin__*`) are identical in either region. Installing both collides; users
  install one. The EU version carries a `-eu` suffix so that switching an organization entry's `ref` between
  regions registers as an update instead of being skipped as the same version.
- Provider-neutral content alone does not establish support for additional hosts. Keep installation and
  compatibility claims aligned with the supported platforms documented in `README.md`.

## MCP Server Configuration

`plugins/anduin/.mcp.json` hardcodes Production US (`https://mcp.anduin.app/mcp`); the `-eu` tags carry the same folder with Production EU (`https://mcp.eu.anduin.app/mcp`). The region is chosen by which ref is installed — the marketplace entry `anduin` (US, `./plugins/anduin`) or `anduin-eu` (EU, the current `vX.Y.Z-eu` tag), or the `ref` in an organization's `git-subdir` source (`vX.Y.Z` or `vX.Y.Z-eu`). There are no moving release branches: claude.ai syncs an organization marketplace only when the customer's own repository has a new commit, so a moving ref would not deliver updates there. This means the plugin works out of the box for both Cowork and Claude Code — no environment variables or setup needed.

To test against another environment, load a development copy of the plugin whose `.mcp.json` points at that URL, so the server keeps the `plugin:anduin:anduin` identity the agents allow:

```bash
cp -R plugins/anduin /tmp/anduin-dev   # then edit the url in /tmp/anduin-dev/.mcp.json
claude --plugin-dir /tmp/anduin-dev    # confirm the URL in /mcp before use
```

`claude mcp remove anduin` does not affect a plugin-provided server, and a manually added `anduin` server gets a different tool prefix that the agents' `tools:` lists do not allow.

Available environments:

| Environment | URL |
|---|---|
| Production (US) *(`main`, tags `vX.Y.Z`)* | `https://mcp.anduin.app/mcp` |
| Production (EU) *(tags `vX.Y.Z-eu`)* | `https://mcp.eu.anduin.app/mcp` |
| Staging | `https://mcp-staging.anduin.dev/mcp` |
| Minas Tirith (daily bounce) | `https://minas-tirith.anduin.dev/mcp` |
| Local Development | `http://gondor-local.io:8080/mcp` |

## Releasing

Prerequisite: GitHub Actions enabled on the repository with **Read and write** workflow permissions (the release workflow pushes a tag and creates a Release).

1. Bump `version` in `plugins/anduin/.claude-plugin/plugin.json` and add a `## X.Y.Z` section to `CHANGELOG.md`.
2. Update the tags that name the release: the `anduin-eu` entry's `ref` in `.claude-plugin/marketplace.json`, the README organization example, and the guide's two JSON blocks. `scripts/build-eu-plugin.sh --check` fails until all of them name the new release.
3. Merge to `main`, then push an annotated tag `vX.Y.Z` by hand. A tag pushed with `GITHUB_TOKEN` would not trigger the workflow.
4. Confirm the Action published `vX.Y.Z-eu` and the GitHub Release (paste-ready `ref` lines, commit SHAs, one ZIP per region). If it fails, re-run it from the Actions tab: a re-run reuses the `-eu` tag only if it matches a fresh EU build, keeps assets already uploaded, adds missing ones and publishes a leftover draft. Until it succeeds, the `anduin-eu` entry, README and guide point at a tag that does not exist.
5. Run `scripts/render-guide.sh` and send the PDF with the release announcement.

Never move or delete `v*` tags by hand: customers pin them.

## Development Notes

- Version lives only in `plugins/anduin/.claude-plugin/plugin.json`. Marketplace entries carry no `version`: Claude Code silently prefers plugin.json, so a second copy could only go stale. See [Releasing](#releasing).
- Plugin name is `anduin` in every build; the EU marketplace entry is named `anduin-eu` but still loads under the `anduin` namespace. Marketplace name is `anduin-marketplace`.
- Agent frontmatter fields: `name`, `description` (with examples), `model`, `color`, `tools`.
- Skill frontmatter fields: `name`, `description`, and optionally `argument-hint`, `allowed-tools`.
- OAuth2 scopes: `fundsub:read/write/admin`, `dataroom:read/write/admin` (hierarchy admin > write > read), plus `mcp:render` (flat, non-hierarchical) which grants ONLY the three display-only UI render tools and no data access. The four destructive dataroom tools (`dr_archive_dataroom`, `dr_delete_items`, `dr_remove_users`, `dr_modify_user_permissions`) are gated on `dataroom:admin` — a `dataroom:write` token cannot call them. `fundsub:admin` currently unlocks nothing beyond `fundsub:write` on the public server.
- MCP tools are filtered by the user's approved OAuth2 scopes at runtime.
- UI render tools (`render_chart`, `render_table`, `render_ui`) are cross-domain, UNPREFIXED, and gated by
  `mcp:render`. They are display-only MCP Apps tools returning structured content and a text fallback, with
  `ui://anduin/{chart,table,form}` resources. A successful tool call alone does not prove an iframe displayed.
- Rendering is capability-based, not coupled to a plugin version. Use the host-provided catalog, which may be
  live-filtered or a published snapshot, plus actual UI support. Missing tools can reflect deployment, scopes, or
  host packaging; do not diagnose the cause from absence alone. Read tools return data, explicit `show_*` tools
  may produce domain widgets, and generic render tools grant no domain-data access. Keep complete Markdown
  fallback when widgets cannot be shown, and a short takeaway when they can. Maintain this behavior in the skills.
- Cowork only supports public URLs (not local dev).
- `.claude/*.local.md` files are gitignored (per-user local config).
