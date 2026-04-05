# Anduin Plugin for Claude Code

Claude Code plugin for managing Anduin fund subscriptions and data rooms via the Model Context Protocol (MCP).

## Features

### GP Assistant
AI assistant for fund managers (General Partners):
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

### Option A: From Marketplace (recommended)

Add the Anduin marketplace and install:

```bash
claude marketplace add https://raw.githubusercontent.com/anduintransaction/anduin-plugin/main/marketplace.json
claude plugin install anduin-plugin
```

To update later:

```bash
claude marketplace update
```

### Option B: From GitHub

```bash
claude plugin add --source github anduintransaction/anduin-plugin
```

### Option C: Local (for development)

```bash
claude plugin add /path/to/anduin-plugin
```

## Configuration

### Set the MCP server URL

Set the `ANDUIN_MCP_URL` environment variable to point to your Anduin environment:

```bash
# Production
export ANDUIN_MCP_URL="https://gondor-public.anduintransact.com/mcp"

# Staging
export ANDUIN_MCP_URL="https://mordor.anduin.dev/mcp"

# Local development (HTTP only for local)
export ANDUIN_MCP_URL="http://gondor-local.io:8080/mcp"
```

Add to your shell profile (`~/.zshrc` or `~/.bashrc`) for persistence.

### Verify connection

Restart Claude Code and run `/mcp` to verify the `anduin` MCP server is connected.

## Usage

### Proactive triggering

The agents activate automatically when you mention relevant topics:

- *"Review the LPs in my fund"* — triggers GP Assistant
- *"Create a data room for the Series B deal"* — triggers Data Room Agent
- *"Check AML status for this investor"* — triggers GP Assistant
- *"Who has access to our deal room?"* — triggers Data Room Agent

### Example tasks

**GP Assistant:**
- "Show me the fund report for Venture Fund III"
- "Which LPs have incomplete forms?"
- "Compare the commitment amounts across all LPs"
- "Tag these orders as reviewed"
- "Invite john@acme.com as a fund manager"

**Data Room Agent:**
- "List all my data rooms"
- "Invite sarah@example.com as an Admin to the Acme data room"
- "Organize the files into folders by document type"
- "Show me the activity analytics for this data room"

## OAuth2 Authentication

Authentication is handled automatically by Claude Code:

1. On first use, a browser opens for Anduin login
2. Approve the requested OAuth2 scopes
3. Claude Code manages tokens (access + refresh) automatically

No manual token configuration needed.

### Available scopes

| Scope | Description |
|-------|-------------|
| `fundsub:read` | View fund subscription data |
| `fundsub:write` | Modify fund subscriptions |
| `dataroom:read` | View data room contents |
| `dataroom:write` | Modify data rooms |

Scope hierarchy: `admin` implies `write` implies `read`.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| MCP server not found | Check `echo $ANDUIN_MCP_URL` is set, restart Claude Code |
| 401 Unauthorized | Token expired — restart Claude Code to re-authenticate |
| 403 Insufficient scopes | Re-authorize with broader scopes, or check with your admin |
| Tools not appearing | Run `/mcp` to check connection; tools are filtered by your scopes |

For detailed setup help, ask Claude: *"How do I set up the Anduin MCP connection?"*

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
├── marketplace.json         # Marketplace distribution config
├── LICENSE
└── README.md
```

## License

MIT License. See [LICENSE](LICENSE) for details.
