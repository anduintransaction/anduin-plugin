---
name: managed-agents
description: |
  This skill should be used when the user asks about "managed agents", "deploy agents",
  "Anthropic managed agents", "scheduled agents", "automated review", "agent API",
  "run agents in production", "agent deployment", or wants to deploy Anduin's AI agents
  as Claude Managed Agents for scheduled, async, or API-driven workflows.
---

# Anduin Managed Agents Deployment

Deploy Anduin's fund subscription and data room AI agents as Claude Managed Agents
for scheduled, event-driven, and API-accessible workflows that the embedded reagent
agents cannot handle (no cron, no webhooks, no external API access).

## When to Use Managed Agents vs Embedded Agents

<table>
| Use Case | Embedded (Reagent) | Managed Agents |
|----------|-------------------|----------------|
| Interactive chat in Anduin UI | Yes | No |
| Scheduled/cron tasks | No | Yes |
| Event-driven (webhook) triggers | No | Yes |
| Long-running batch processing | Limited | Yes |
| External API access | No | Yes |
| Tool approval UI | Yes | No |
</table>

## Core Concepts

- **Agent** — Reusable config: model, system prompt, MCP servers. Versioned via API.
- **Environment** — Container template with packages and network access.
- **Session** — Running agent instance. $0.08/session-hour (idle time free).
- **Vault** — Secure credential storage. Manages OAuth token refresh automatically.

## Setup Prerequisites

1. Anthropic API key with managed agents access
2. Python 3.10+ or Node.js 18+ with the Anthropic SDK
3. OAuth2 credentials for Anduin MCP server (client\_id, client\_secret)
4. Access to an Anduin environment (production, staging, or local)

## MCP Server Environments

| Environment | URL |
|---|---|
| Production (US) | `https://mcp.anduin.app/mcp` |
| Production (EU) | `https://mcp.eu.anduin.app/mcp` |
| Staging | `https://mcp-staging.anduin.dev/mcp` |
| Minas Tirith | `https://minas-tirith.anduin.dev/mcp` |
| Local | `http://gondor-local.io:8080/mcp` |

## Deployment Workflow

### Step 1: Install the SDK

```bash
# Python
pip install anthropic

# TypeScript/Node
npm install @anthropic-ai/sdk
```

### Step 2: Create a Vault for OAuth Credentials

```python
import anthropic
client = anthropic.Anthropic()

vault = client.beta.vaults.create(name="anduin-credentials")
client.beta.vaults.credentials.create(
    vault_id=vault.id,
    provider="anduin",
    access_token="<OAUTH_ACCESS_TOKEN>",
    refresh_token="<OAUTH_REFRESH_TOKEN>",
)
```

### Step 3: Create the Agent

Reuse system prompts from the anduin-plugin agent definitions. Reference templates in
`references/templates.md` for ready-to-use Python scripts for each use case:

- **GP Assistant** — Interactive fund subscription management
- **Data Room Agent** — Data room operations
- **Subscription Reviewer** — Automated LP form review (event-driven)
- **Compliance Monitor** — Scheduled AML/KYC sweep
- **Fund Health Reporter** — Scheduled fund status reports

### Step 4: Create an Environment

```python
environment = client.beta.environments.create(
    name="anduin-agent-env",
    packages=["python3"],
)
```

### Step 5: Start a Session and Send Work

```python
session = client.beta.sessions.create(
    agent=agent.id,
    environment_id=environment.id,
    vault_ids=[vault.id],
)

client.beta.sessions.events.send(
    session.id,
    events=[{
        "type": "user.message",
        "content": [{"type": "text", "text": "Review LP order ord_abc123"}]
    }]
)

for event in client.beta.sessions.events.stream(session.id):
    print(event)
```

## Available MCP Tools

All 72 Anduin MCP tools are available to managed agents via the HTTP MCP server.

**FundSub tools (45):** list\_funds, get\_fund\_info, list\_orders, query\_dashboard,
get\_lp\_status, get\_form\_schema, get\_form\_markdown, get\_form\_validation\_errors,
update\_form\_fields, draft\_comment, convert\_document\_to\_markdown, and more.

**DataRoom tools (27):** dr\_list\_entities, dr\_list\_datarooms, dr\_create\_dataroom,
dr\_invite\_users, dr\_list\_files, dr\_search, dr\_get\_insights, dr\_get\_file\_download\_url, and more.

Consult the `anduin:gp-assistant` and `anduin:dataroom` skills for full tool lists,
chaining rules, and workflow patterns.

## Key Constraints

- Beta API: requires `managed-agents-2026-04-01` header (SDKs set this automatically)
- Cloud-only (Anthropic infrastructure), US and EU regions
- $0.08/session-hour + standard Claude token pricing
- Multi-agent orchestration is in research preview

## Additional Resources

### Reference Files

- **`references/templates.md`** — Complete Python deployment templates for each use case
- **`references/api-reference.md`** — Managed Agents API endpoints and parameters
