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

Since this is a content-only plugin with no test suite, verify manually:

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

This is a **content-only plugin** — no build step, no dependencies, no tests. All files are markdown or JSON.

```
.claude-plugin/marketplace.json       — Marketplace manifest (lists all plugins)

plugins/anduin/                       — Anduin platform plugin
  .claude-plugin/plugin.json          — Plugin manifest (name, version, description)
  .mcp.json                           — MCP server config (hardcoded to Production US)
  agents/                             — Autonomous agent definitions (spawned as subagents)
    gp-assistant.md                   — Fund subscription agent (model: sonnet, tools: mcp__plugin_anduin_anduin__*)
    dataroom-agent.md                 — Data room agent (model: sonnet, tools: mcp__plugin_anduin_anduin__*)
  skills/                             — Domain knowledge loaded into context on demand
    gp-assistant/SKILL.md             — GP domain terminology, tool catalog, workflows
    dataroom/SKILL.md                 — Data room domain terminology, tool catalog, workflows
```

**Key patterns:**
- Each domain (GP, Data Room) has both an agent (`.md` in `agents/`) and a skill (`.md` in `skills/`). The agent defines behavior, model, and tool access. The skill provides domain knowledge that gets loaded into context. The agent references MCP tools prefixed `dr_` (data room) or unprefixed (fund subscription).

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
- UI render tools (`render_chart`, `render_table`, `render_ui`) are cross-domain, UNPREFIXED, and gated by `mcp:render`. They are MCP Apps tools: each returns `structuredContent` + a text fallback and points at a `ui://anduin/{chart,table,form}` resource that UI-capable hosts (Claude Code, Cowork) fetch and render in a sandboxed iframe (`text/html;profile=mcp-app`). All three are display-only (`interactive: false`, read-only) — documented in both the gp-assistant and dataroom skills/agents.
- Render-tool availability is **per-environment and runtime-discovered, NOT coupled to a plugin version.** A server build that hasn't shipped the render feature advertises neither the `io.modelcontextprotocol/ui` capability, the `mcp:render` scope, nor the render tools — so the same published plugin is correct against local/staging/production simultaneously, exposing render only where the server supports it (envs roll out at different times). The docs are deliberately written **capability-first** (the skills/agents tell the assistant to rely on the live `tools/list` and degrade to markdown when a render tool is absent) rather than asserting the tools always exist. When editing render docs, keep this framing: never make the plugin hard-depend on a render tool, and don't fork the plugin per environment.
- Cowork only supports public URLs (not local dev).
- `.claude/*.local.md` files are gitignored (per-user local config).
