# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A Claude Code plugin that provides AI assistants for two Anduin platform domains:
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
/anduin:gp-assistant    # Should load GP domain knowledge
/anduin:dataroom        # Should load Data Room domain knowledge

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
    gp-assistant.md                   — Fund subscription agent (model: sonnet, tools: mcp__anduin__*)
    dataroom-agent.md                 — Data room agent (model: sonnet, tools: mcp__anduin__*)
  skills/                             — Domain knowledge loaded into context on demand
    gp-assistant/SKILL.md             — GP domain terminology, tool catalog, workflows
    dataroom/SKILL.md                 — Data room domain terminology, tool catalog, workflows
```

**Key pattern:** Each domain (GP, Data Room) has both an agent (`.md` in `agents/`) and a skill (`.md` in `skills/`). The agent defines behavior, model, and tool access. The skill provides domain knowledge that gets loaded into context. The agent references MCP tools prefixed `dr_` (data room) or unprefixed (fund subscription).

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
- OAuth2 scopes: `fundsub:read/write/admin`, `dataroom:read/write/admin`. Scope hierarchy: admin > write > read.
- MCP tools are filtered by the user's approved OAuth2 scopes at runtime.
- Cowork only supports public URLs (not local dev).
- `.claude/*.local.md` files are gitignored (per-user local config).
