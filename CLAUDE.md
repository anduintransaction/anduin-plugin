# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A Claude Code plugin that provides AI assistants for Anduin platform domains:
- **GP Assistant** — fund subscription management (LP review, forms, AML/KYC, dashboards, tagging)
- **Data Room Agent** — virtual data room management (rooms, participants, files, analytics)
- **Managed Agents Deployer** — deploy Anduin agents as Claude Managed Agents for scheduled, event-driven, and API-accessible workflows

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
/anduin:managed-agents     # Should load Managed Agents deployment guide

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
    managed-agents-deployer.md        — Managed Agents deployment agent (model: sonnet, tools: Read, Write, Bash, Skill)
  skills/                             — Domain knowledge loaded into context on demand
    gp-assistant/SKILL.md             — GP domain terminology, tool catalog, workflows
    dataroom/SKILL.md                 — Data room domain terminology, tool catalog, workflows
    managed-agents/                   — Managed Agents deployment guide
      SKILL.md                        — Deployment workflow, setup, MCP environments
      references/templates.md         — Python deployment templates (5 use cases)
      references/api-reference.md     — Managed Agents API endpoints and pricing
```

**Key patterns:**
- Each domain (GP, Data Room) has both an agent (`.md` in `agents/`) and a skill (`.md` in `skills/`). The agent defines behavior, model, and tool access. The skill provides domain knowledge that gets loaded into context. The agent references MCP tools prefixed `dr_` (data room) or unprefixed (fund subscription).
- The managed-agents deployer agent uses local tools (Read, Write, Bash) instead of MCP tools. It generates deployment scripts and reads skill references for templates. It does not connect to the Anduin MCP server directly.

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

- Version is tracked in `.claude-plugin/plugin.json` and git tags (e.g., `v0.1.0`). Bump both when releasing.
- Plugin name is `anduin` (in plugin.json). Marketplace name is `anduin-marketplace`.
- Agent frontmatter fields: `name`, `description` (with examples), `model`, `color`, `tools`.
- Skill frontmatter fields: `name`, `description`, and optionally `argument-hint`, `allowed-tools`.
- OAuth2 scopes: `fundsub:read/write/admin`, `dataroom:read/write/admin` (hierarchy admin > write > read), plus `mcp:render` (flat, non-hierarchical) which grants ONLY the three display-only UI render tools and no data access.
- MCP tools are filtered by the user's approved OAuth2 scopes at runtime.
- UI render tools (`render_chart`, `render_table`, `render_ui`) are cross-domain, UNPREFIXED, and gated by `mcp:render`. They are MCP Apps tools: each returns `structuredContent` + a text fallback and points at a `ui://anduin/{chart,table,form}` resource that UI-capable hosts (Claude Code, Cowork) fetch and render in a sandboxed iframe (`text/html;profile=mcp-app`). All three are display-only (`interactive: false`, read-only) — documented in both the gp-assistant and dataroom skills/agents.
- Cowork only supports public URLs (not local dev).
- `.claude/*.local.md` files are gitignored (per-user local config).
- Managed Agents skill uses progressive disclosure: lean SKILL.md with detailed templates and API reference in `references/`. The deployer agent loads these on demand via the Skill tool.
