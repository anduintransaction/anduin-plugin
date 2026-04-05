# Anduin Plugin for Claude Code

Claude Code plugin for managing Anduin fund subscriptions and data rooms via the Model Context Protocol (MCP).

## Features

### GP Assistant
AI assistant for fund managers (General Partners) that helps with:
- LP order review and subscription form inspection
- Fund reporting and dashboard queries
- Order tagging and cross-order analysis
- Fund manager invitations
- AML/KYC compliance checks
- Form filling assistance

### Data Room Agent
AI assistant for virtual data room management:
- Data room creation and organization
- Participant management (invite, remove, change roles)
- File and folder operations
- Search and navigation
- Activity analytics and insights

## Installation

### 1. Install the plugin

```bash
claude plugin add /path/to/anduin-mcp-plugins
```

Or add to your project's `.claude/settings.json`:
```json
{
  "plugins": ["/path/to/anduin-mcp-plugins"]
}
```

### 2. Configure the MCP server URL

Set the `ANDUIN_MCP_URL` environment variable:

```bash
# Production
export ANDUIN_MCP_URL="https://gondor-public.anduintransact.com/mcp"

# Staging
export ANDUIN_MCP_URL="https://mordor.anduin.dev/mcp"

# Local development
export ANDUIN_MCP_URL="http://gondor-local.io:8080/mcp"
```

Add to your shell profile for persistence.

### 3. Verify connection

Restart Claude Code and run `/mcp` to verify the `anduin` MCP server is connected.

## Usage

### Proactive triggering
The agents activate automatically when you mention relevant topics:
- "Review the LPs in my fund" triggers the GP Assistant
- "Create a data room" triggers the Data Room Agent
- "Check AML status for this investor" triggers the GP Assistant

### Explicit invocation
Use slash commands to invoke agents directly:
- Ask about fund subscriptions, LP review, fund reports
- Ask about data rooms, participants, file management

## OAuth2 Authentication

Authentication is handled automatically by Claude Code:
1. On first use, a browser opens for Anduin login
2. Approve the requested OAuth2 scopes
3. Claude Code manages tokens (access + refresh) automatically

No manual token configuration needed.

## Available Scopes

| Scope | Description |
|-------|-------------|
| `fundsub:read` | View fund subscription data |
| `fundsub:write` | Modify fund subscriptions |
| `dataroom:read` | View data room contents |
| `dataroom:write` | Modify data rooms |

## Plugin Structure

```
anduin-plugin/
├── .claude-plugin/
│   └── plugin.json          # Plugin manifest
├── agents/
│   ├── dataroom-agent.md    # Data Room autonomous agent
│   └── gp-assistant.md      # GP Assistant autonomous agent
├── skills/
│   ├── dataroom/
│   │   └── SKILL.md         # Data Room domain knowledge
│   ├── gp-assistant/
│   │   └── SKILL.md         # GP Assistant domain knowledge
│   └── setup/
│       └── SKILL.md         # MCP connection setup guide
├── .mcp.json                # MCP server configuration
├── README.md
└── .gitignore
```
