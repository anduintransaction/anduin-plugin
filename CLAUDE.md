# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A Claude Code plugin that provides AI assistants for two Anduin platform domains:
- **GP Assistant** — fund subscription management (LP review, forms, AML/KYC, dashboards, tagging)
- **Data Room Agent** — virtual data room management (rooms, participants, files, analytics)

The plugin connects to Anduin's MCP server (configured via `ANDUIN_MCP_URL` env var) and uses OAuth2 for authentication. It works on both Claude Code (CLI) and Cowork (desktop app).

## Local Development

```bash
# Install from local path for testing (run from repo root)
/plugin add .

# Install from marketplace (for users)
/plugin marketplace add anduintransaction/anduin-plugin
/plugin install anduin@anduin-marketplace
```

## Testing Changes

Since this is a content-only plugin with no test suite, verify manually:

```bash
# Reinstall after changes
/plugin add .

# Verify skills load
/anduin:setup           # Should show the setup wizard
/anduin:gp-assistant    # Should load GP domain knowledge
/anduin:dataroom        # Should load Data Room domain knowledge

# Verify hook fires on session start
bash hooks/scripts/check-config.sh   # Should warn if ANDUIN_MCP_URL is unset
```

## Plugin Architecture

This is a **content-only plugin** — no build step, no dependencies, no tests. All files are markdown or JSON.

```
.claude-plugin/plugin.json   — Plugin manifest (name, version, description)
.claude-plugin/marketplace.json — Copy of marketplace.json for Cowork discovery
.mcp.json                    — MCP server config (uses $ANDUIN_MCP_URL)
marketplace.json             — Marketplace distribution metadata

agents/                      — Autonomous agent definitions (spawned as subagents)
  gp-assistant.md            — Fund subscription agent (model: sonnet, tools: mcp__anduin__*)
  dataroom-agent.md          — Data room agent (model: sonnet, tools: mcp__anduin__*)

skills/                      — Domain knowledge loaded into context on demand
  setup/SKILL.md             — Interactive setup wizard (/anduin:setup)
  gp-assistant/SKILL.md      — GP domain terminology, tool catalog, workflows
  dataroom/SKILL.md          — Data room domain terminology, tool catalog, workflows

hooks/
  hooks.json                 — SessionStart hook config
  scripts/check-config.sh    — Warns if ANDUIN_MCP_URL is unset
```

**Key pattern:** Each domain (GP, Data Room) has both an agent (`.md` in `agents/`) and a skill (`.md` in `skills/`). The agent defines behavior, model, and tool access. The skill provides domain knowledge that gets loaded into context. The agent references MCP tools prefixed `dr_` (data room) or unprefixed (fund subscription).

## MCP Server Environments

See `skills/setup/SKILL.md` or README.md for the full environment URL table. The setup skill (`/anduin:setup`) handles configuration interactively.

## Development Notes

- Version is tracked in `.claude-plugin/plugin.json` and git tags (e.g., `v0.1.0`). Bump both when releasing.
- Plugin name is `anduin` (in plugin.json). Marketplace name is `anduin-marketplace`.
- Agent frontmatter fields: `name`, `description` (with examples), `model`, `color`, `tools`.
- Skill frontmatter fields: `name`, `description`, and optionally `argument-hint`, `allowed-tools`.
- The `.mcp.json` uses `${ANDUIN_MCP_URL}` env var substitution — the URL is not hardcoded.
- OAuth2 scopes: `fundsub:read/write/admin`, `dataroom:read/write/admin`. Scope hierarchy: admin > write > read.
- MCP tools are filtered by the user's approved OAuth2 scopes at runtime.
- Cowork only supports public URLs (not local dev).
- `.claude/*.local.md` files are gitignored (per-user local config).
