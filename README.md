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
- **Read documents** — convert subscription documents and spreadsheets to readable text using OCR
- **Visualize data** — see charts, tables, and form-style summaries as interactive widgets (in apps that support them)

### Data Room Agent — Virtual Data Room Management

For deal teams and anyone managing shared documents:

- **Create and organize data rooms** — set up new rooms, create folder structures
- **Manage participants** — invite, remove, or change roles (Admin, Member, Contributor, Observer)
- **Search and browse files** — find documents and navigate folder structures
- **View analytics** — see who's accessing what, activity trends, and engagement metrics
- **Read documents** — convert PDFs, images, and spreadsheets to readable text using OCR
- **Visualize data** — see engagement charts, file/participant tables, and summaries as interactive widgets (in apps that support them)

### Managed Agents Deployer — Deploy Agents to Production

For developers and platform teams:

- **Deploy Anduin agents as Claude Managed Agents** — run agents server-side via Anthropic's hosted infrastructure
- **Automate subscription reviews** — set up event-driven agents that review LP submissions automatically
- **Schedule fund reports** — create cron-triggered agents for nightly health reports and compliance monitoring
- **Generate deployment scripts** — get ready-to-run Python scripts customized for your use case

## Getting Started

### Step 1: Install the plugin

**Cowork (desktop app):**

Follow [Anthropic's guide to using plugins in Cowork](https://support.claude.com/en/articles/13837440-use-plugins-in-cowork). When adding a marketplace, use `anduintransaction/anduin-plugin`.

**Claude Code (terminal):**
```
/plugin marketplace add anduintransaction/anduin-plugin
/plugin install anduin@anduin-marketplace
```

### Step 2: Sign in

After installing, restart the app. On first use, a browser window opens for you to sign in with your Anduin credentials. After that, authentication is handled automatically — no tokens or passwords to manage.

The plugin connects to **Anduin Production (US)** by default. No additional setup is needed.

You'll be asked to approve access scopes:

| Scope | What it allows |
|---|---|
| `fundsub:read` | View fund subscription data (orders, forms, documents) |
| `fundsub:write` | Make changes to subscriptions (update forms, tags, invite managers) |
| `dataroom:read` | View data rooms (files, participants, analytics) |
| `dataroom:write` | Make changes to data rooms (create, invite, upload, delete) |
| `mcp:render` | Show interactive charts, tables, and forms as visual widgets (display-only) |

Approve whichever scopes match the work you need to do. You only see tools relevant to your approved scopes. The visual widgets from `mcp:render` appear in apps that support them (Claude Code, Cowork); elsewhere you still get the same data as text.

## Usage

Just describe what you need in plain language. The right assistant activates automatically:

**Fund subscriptions:**
- *"Show me the fund report for Venture Fund III"*
- *"Which LPs have incomplete forms?"*
- *"Compare the commitment amounts across all LPs"*
- *"Tag these orders as reviewed"*
- *"Invite john@acme.com as a fund manager"*
- *"Check AML status for the LP orders in Close 2"*
- *"Read the subscription agreement for Acme Capital"*
- *"Chart the commitment totals by close"*
- *"Show the orders as a table"*

**Data rooms:**
- *"List all my data rooms"*
- *"Create a data room for the Series B deal"*
- *"Invite sarah@example.com as an Admin to the Acme data room"*
- *"Organize the files into folders by document type"*
- *"Show me the activity analytics for this data room"*
- *"Who has access to our deal room?"*
- *"Show me the contents of the NDA document"*
- *"Chart the most-viewed files in this data room"*
- *"Show the participants as a table"*

**Managed agents:**
- *"Deploy the GP assistant as a managed agent"*
- *"Set up an automated agent that reviews LP submissions"*
- *"Create a daily compliance check across all funds"*
- *"Generate a managed agent script for our data room"*

## Troubleshooting

| Problem | What to do |
|---|---|
| **Anduin MCP not showing** | Reinstall the plugin and restart the app. |
| **"Unauthorized" or login issues** | Your session may have expired. Restart the app to sign in again. |
| **"Insufficient scopes" error** | You need broader permissions. Restart and approve additional scopes when prompted, or ask your admin for access. |
| **Tools not showing up** | Check that the server is connected (in Claude Code: run `/mcp`). You only see tools matching your approved scopes. |

Need help? Ask Claude: *"How do I connect to Anduin?"*

## Updating

Run inside Claude Code or Cowork:
```
/plugin marketplace update
```

## For Developers

<details>
<summary>Advanced configuration, plugin structure, and local development</summary>

### Switching environments

The plugin defaults to Production (US). To connect to a different environment, remove and re-add the MCP server in Claude Code:

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

To revert to Production US, just reinstall the plugin — it will restore the default.

### Local installation

```
/plugin add ./plugins/anduin
```

### Plugin structure

```
anduin-plugin/
├── .claude-plugin/
│   └── marketplace.json         # Marketplace manifest
├── plugins/anduin/
│   ├── .claude-plugin/
│   │   └── plugin.json          # Plugin manifest
│   ├── .mcp.json                # MCP server config (Production US)
│   ├── agents/
│   │   ├── dataroom-agent.md              # Data Room autonomous agent
│   │   ├── gp-assistant.md                # GP Assistant autonomous agent
│   │   └── managed-agents-deployer.md     # Managed Agents deployment agent
│   └── skills/
│       ├── dataroom/
│       │   └── SKILL.md                   # Data Room domain knowledge
│       ├── gp-assistant/
│       │   └── SKILL.md                   # GP Assistant domain knowledge
│       └── managed-agents/
│           ├── SKILL.md                   # Managed Agents deployment guide
│           └── references/
│               ├── templates.md           # Python deployment templates
│               └── api-reference.md       # Managed Agents API reference
├── LICENSE
└── README.md
```

### Compatibility

| Platform | Supported | Notes |
|---|---|---|
| Claude Code (terminal) | Yes | All server URLs work, including local dev |
| Cowork (desktop app) | Yes | Connects to Production automatically |

</details>

## License

MIT License. See [LICENSE](LICENSE) for details.
