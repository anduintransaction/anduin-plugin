# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A Claude Code plugin that provides AI assistants for Anduin platform domains:
- **GP Assistant** — fund subscription management (LP review, forms, AML/KYC, dashboards, tagging)
- **Data Room Agent** — virtual data room management (rooms, participants, files, analytics)

The plugin connects to Anduin's MCP server (hardcoded to Production US by default) and uses OAuth2 for authentication. It works on both Claude Code (CLI) and Cowork (desktop app) with no setup required.

## Local Development

```bash
# Install from local path for testing (run from repo root)
/plugin add ./plugins/anduin

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
# Reinstall after changes
/plugin add ./plugins/anduin

# Verify skills load
/anduin:gp-assistant       # Should load GP domain knowledge
/anduin:dataroom           # Should load Data Room domain knowledge

# Verify MCP connects
# After reinstall, check /mcp — "anduin" should appear pointing to Production US
```

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
```

**Key patterns:**
- Each domain has one canonical skill and a thin Claude adapter. The adapter preserves activation, model, and tool
  access, then invokes `anduin:gp-assistant` or `anduin:dataroom` with the already-allowed `Skill` tool before domain
  work. If loading fails, it stops rather than running without the shared rules. Keep workflows, permission rules,
  confirmation, recovery, and presentation guidance in the skill, not a second copy in the adapter.
- Canonical skills use wire names (`dr_` for Data Room, unprefixed for fund subscription). Resolve them against the
  current host's provided Anduin catalog; do not copy Claude-specific MCP prefixes into shared behavior.
- Provider-neutral content alone does not establish support for additional hosts. Keep installation and
  compatibility claims aligned with the supported platforms documented in `README.md`.

## MCP Server Configuration

The `.mcp.json` hardcodes the Production US URL (`https://mcp.anduin.app/mcp`). This means the plugin works out of the box for both Cowork and Claude Code — no environment variables or setup needed.

Advanced users (developers) can switch to a different environment by removing and re-adding the MCP server manually:

```bash
claude mcp remove anduin
claude mcp add --transport http anduin <environment-url>
```

Available environments:

| Environment | URL |
|---|---|
| Production (US) *(default)* | `https://mcp.anduin.app/mcp` |
| Production (EU) | `https://mcp.eu.anduin.app/mcp` |
| Staging | `https://mcp-staging.anduin.dev/mcp` |
| Minas Tirith (daily bounce) | `https://minas-tirith.anduin.dev/mcp` |
| Local Development | `http://gondor-local.io:8080/mcp` |

## Development Notes

- Version is tracked in `plugins/anduin/.claude-plugin/plugin.json`, mirrored in the plugin entry of the root `.claude-plugin/marketplace.json`, and git tags (e.g., `v0.1.0`). Bump all three in lockstep when releasing — plugin.json and marketplace.json must never diverge.
- Plugin name is `anduin` (in plugin.json). Marketplace name is `anduin-marketplace`.
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
