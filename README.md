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

Choose one option:

**From Marketplace (recommended):**
```bash
claude marketplace add https://raw.githubusercontent.com/anduintransaction/anduin-plugin/main/marketplace.json
claude plugin install anduin-plugin
```

**From GitHub:**
```bash
claude plugin add --source github anduintransaction/anduin-plugin
```

### Step 2: Set your Anduin server URL

You need to tell the plugin which Anduin server to connect to. Pick the one that matches your environment:

| Environment | URL |
|---|---|
| **Production (US)** | `https://mcp.anduin.app/mcp` |
| **Production (EU)** | `https://mcp.eu.anduin.app/mcp` |
| **Staging** | `https://mcp-staging.anduin.dev/mcp` |
| **Minas Tirith** (daily bounce) | `https://mcp-minas-tirith.anduin.dev/mcp` |
| **Local dev** (developers only) | `http://gondor-local.io:8080/mcp` |

<details>
<summary><strong>Cowork users</strong> (desktop app)</summary>

Set the environment variable in your system settings or shell profile, then restart the Cowork app.

**macOS:** Add to `~/.zshrc`:
```bash
export ANDUIN_MCP_URL="https://mcp.anduin.app/mcp"
```

**Windows:** Set via System Properties > Environment Variables, or in PowerShell:
```powershell
[Environment]::SetEnvironmentVariable("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp", "User")
```

> **Note:** Cowork runs in the cloud, so only publicly accessible URLs work (Production, Staging, Minas Tirith). Local dev URLs will not work with Cowork.

</details>

<details>
<summary><strong>Claude Code users</strong> (terminal)</summary>

Add to your shell profile (`~/.zshrc` or `~/.bashrc`):
```bash
export ANDUIN_MCP_URL="https://mcp.anduin.app/mcp"
```

Then restart Claude Code and run `/mcp` to verify the `anduin` server is connected.

All URLs work with Claude Code, including local dev.

</details>

### Step 3: Sign in

On first use, a browser window opens for you to sign in with your Anduin credentials. After that, authentication is handled automatically — no tokens or passwords to manage.

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

## Troubleshooting

| Problem | What to do |
|---|---|
| **Can't find the Anduin server** | Make sure `ANDUIN_MCP_URL` is set and restart the app. In Claude Code, run `echo $ANDUIN_MCP_URL` to check. |
| **"Unauthorized" or login issues** | Your session may have expired. Restart the app to sign in again. |
| **"Insufficient scopes" error** | You need broader permissions. Restart and approve additional scopes when prompted, or ask your admin for access. |
| **Tools not showing up** | Check that the server is connected (in Claude Code: run `/mcp`). You only see tools matching your approved scopes. |
| **Cowork: connection failed** | Only public URLs work with Cowork. Make sure you're not using a local dev URL. |

Need help? Ask Claude: *"How do I set up the Anduin MCP connection?"*

## Updating

```bash
claude marketplace update
```

## For Developers

<details>
<summary>Plugin structure and local development</summary>

### Local installation

```bash
claude plugin add /path/to/anduin-plugin
```

### Plugin structure

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

### Compatibility

| Platform | Supported | Notes |
|---|---|---|
| Claude Code (terminal) | Yes | All server URLs work, including local dev |
| Cowork (desktop app) | Yes | Only publicly accessible server URLs |

</details>

## License

MIT License. See [LICENSE](LICENSE) for details.
