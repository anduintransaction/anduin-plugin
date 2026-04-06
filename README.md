# Anduin Plugin for Claude

AI-powered assistants for managing Anduin fund subscriptions and data rooms. Works with both **Claude Code** (for developers) and **Cowork** (for everyone).

## What You Can Do

### GP Assistant — Fund Subscription Management

For fund managers and operations teams:

- **Review LP orders** — check subscription status, form completeness, and compliance
- **Browse fund dashboards** — see fund reports, commitment summaries, and activity logs
- **Manage order tags** — tag and organize orders across your fund
- **Invite fund managers** — add team members to fund manager groups
- **Check AML/KYC** — review compliance status for investors
- **Assist with forms** — help fill or correct subscription form fields

### Data Room Agent — Virtual Data Room Management

For deal teams and anyone managing shared documents:

- **Create and organize data rooms** — set up new rooms, create folder structures
- **Manage participants** — invite, remove, or change roles (Admin, Member, Contributor, Observer)
- **Search and browse files** — find documents and navigate folder structures
- **View analytics** — see who's accessing what, activity trends, and engagement metrics

## Getting Started

### Step 1: Install the plugin

**Cowork (desktop app):**
1. Open the Plugins panel (left sidebar)
2. Click **+** next to **Marketplaces**
3. Enter `anduintransaction/anduin-plugin` and click **Sync**
4. Find **anduin** in the plugin list and enable it

**Claude Code (terminal):**
```
/plugin marketplace add anduintransaction/anduin-plugin
/plugin install anduin@anduin-marketplace
```

### Step 2: Connect to your Anduin server

After installing, run the setup command:

```
/anduin:setup
```

Claude will ask which Anduin environment you use and configure everything automatically. Available environments:

| Environment | Who it's for |
|---|---|
| **Production (US)** | Most users — live Anduin platform |
| **Production (EU)** | EU region users |
| **Staging** | Internal testing |
| **Minas Tirith** | Daily bounce server for QA |
| **Local Development** | Developers running Anduin locally (Claude Code only) |

> If you skip this step, Claude will remind you to run setup when you start a new session.

After setup, restart the app for the change to take effect.

### Step 3: Sign in

On first use after restarting, a browser window opens for you to sign in with your Anduin credentials. After that, authentication is handled automatically — no tokens or passwords to manage.

You'll be asked to approve access scopes:

| Scope | What it allows |
|---|---|
| `fundsub:read` | View fund subscription data (orders, forms, documents) |
| `fundsub:write` | Make changes to subscriptions (update forms, tags, invite managers) |
| `dataroom:read` | View data rooms (files, participants, analytics) |
| `dataroom:write` | Make changes to data rooms (create, invite, upload, delete) |

Approve whichever scopes match the work you need to do. You only see tools relevant to your approved scopes.

## Usage

Just describe what you need in plain language. The right assistant activates automatically:

**Fund subscriptions:**
- *"Show me the fund report for Venture Fund III"*
- *"Which LPs have incomplete forms?"*
- *"Compare the commitment amounts across all LPs"*
- *"Tag these orders as reviewed"*
- *"Invite john@acme.com as a fund manager"*
- *"Check AML status for the LP orders in Close 2"*

**Data rooms:**
- *"List all my data rooms"*
- *"Create a data room for the Series B deal"*
- *"Invite sarah@example.com as an Admin to the Acme data room"*
- *"Organize the files into folders by document type"*
- *"Show me the activity analytics for this data room"*
- *"Who has access to our deal room?"*

## Changing Your Environment

To switch to a different Anduin server (e.g., from staging to production), run:

```
/anduin:setup
```

Claude will update your configuration. Restart after switching.

## Troubleshooting

| Problem | What to do |
|---|---|
| **Can't find the Anduin server** | Run `/anduin:setup` to configure your server. |
| **"Unauthorized" or login issues** | Your session may have expired. Restart the app to sign in again. |
| **"Insufficient scopes" error** | You need broader permissions. Restart and approve additional scopes when prompted, or ask your admin for access. |
| **Tools not showing up** | Check that the server is connected (in Claude Code: run `/mcp`). You only see tools matching your approved scopes. |
| **Cowork: connection failed** | Only public URLs work with Cowork. Run `/anduin:setup` and pick a non-local environment. |

Need help? Ask Claude: *"How do I set up the Anduin MCP connection?"*

## Updating

Run inside Claude Code or Cowork:
```
/plugin marketplace update
```

## For Developers

<details>
<summary>Plugin structure and local development</summary>

### Local installation

```
/plugin add /path/to/anduin-plugin
```

### Manual configuration (alternative to `/anduin:setup`)

Set the `ANDUIN_MCP_URL` environment variable directly:

```bash
# Production (US)
export ANDUIN_MCP_URL="https://mcp.anduin.app/mcp"

# Production (EU)
export ANDUIN_MCP_URL="https://mcp.eu.anduin.app/mcp"

# Staging
export ANDUIN_MCP_URL="https://mcp-staging.anduin.dev/mcp"

# Minas Tirith (daily bounce)
export ANDUIN_MCP_URL="https://mcp-minas-tirith.anduin.dev/mcp"

# Local development
export ANDUIN_MCP_URL="http://gondor-local.io:8080/mcp"
```

### Plugin structure

```
anduin-plugin/
├── .claude-plugin/
│   └── plugin.json          # Plugin manifest
├── agents/
│   ├── dataroom-agent.md    # Data Room autonomous agent
│   └── gp-assistant.md      # GP Assistant autonomous agent
├── hooks/
│   ├── hooks.json           # SessionStart config detection
│   └── scripts/
│       └── check-config.sh  # Checks if MCP URL is configured
├── skills/
│   ├── dataroom/
│   │   └── SKILL.md         # Data Room domain knowledge
│   ├── gp-assistant/
│   │   └── SKILL.md         # GP Assistant domain knowledge
│   └── setup/
│       └── SKILL.md         # Interactive setup (/anduin:setup)
├── .mcp.json                # MCP server configuration
├── marketplace.json         # Marketplace distribution config
├── LICENSE
└── README.md
```

### Compatibility

| Platform | Supported | Notes |
|---|---|---|
| Claude Code (terminal) | Yes | All server URLs work, including local dev |
| Cowork (desktop app) | Yes | Only publicly accessible server URLs |

</details>

## License

MIT License. See [LICENSE](LICENSE) for details.
