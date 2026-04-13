# Claude Managed Agents API Reference

All endpoints require the `managed-agents-2026-04-01` beta header.
The Python and TypeScript SDKs set this automatically.

## Agents

### Create Agent
```
POST /v1/agents
```

Parameters:
- `name` (string, required) — Display name
- `model` (string, required) — Model ID (e.g. `claude-sonnet-4-6`)
- `system` (string) — System prompt
- `tools` (array) — Built-in tools (`agent_toolset_20260401`)
- `mcp_servers` (object) — MCP server connections
- `skills` (array) — Agent skills
- `allowedTools` (array) — Tool whitelist patterns (e.g. `mcp__anduin__*`)
- `disallowedTools` (array) — Tool blacklist patterns

### List Agents
```
GET /v1/agents
```

### Get Agent
```
GET /v1/agents/{agent_id}
```

### Update Agent
```
POST /v1/agents/{agent_id}
```
Creates a new version. Pass `version` to ensure consistency.

### Archive Agent
```
POST /v1/agents/{agent_id}/archive
```
Makes agent read-only. Existing sessions continue.

### List Versions
```
GET /v1/agents/{agent_id}/versions
```

## Environments

### Create Environment
```
POST /v1/environments
```

Parameters:
- `name` (string, required) — Display name
- `packages` (array) — Pre-installed packages (e.g. `["python3", "nodejs"]`)
- `network_policy` (object) — Allowed/denied domains

### List Environments
```
GET /v1/environments
```

## Sessions

### Create Session
```
POST /v1/sessions
```

Parameters:
- `agent` (string, required) — Agent ID
- `environment_id` (string) — Environment ID
- `vault_ids` (array) — Credential vault IDs

### Send Events
```
POST /v1/sessions/{session_id}/events
```

Event types:
- `user.message` — User turn with text/image content
- `tool.result` — Tool execution result (for custom tools)

### Stream Events (SSE)
```
GET /v1/sessions/{session_id}/events
```

Event types received:
- `agent.thinking` — Agent reasoning
- `agent.tool_call` — Tool invocation
- `agent.tool_result` — Tool result
- `agent.message` — Final response text
- `agent.completed` — Session complete

### Delete Session
```
DELETE /v1/sessions/{session_id}
```

## Vaults

### Create Vault
```
POST /v1/vaults
```

Parameters:
- `name` (string, required) — Vault name

### Add Credentials
```
POST /v1/vaults/{vault_id}/credentials
```

Parameters:
- `provider` (string, required) — Provider name (e.g. `anduin`)
- `access_token` (string) — OAuth access token
- `refresh_token` (string) — OAuth refresh token
- `expires_at` (string) — Token expiry ISO timestamp

## MCP Server Configuration

### HTTP/SSE Transport (for Anduin MCP)
```python
mcp_servers={
    "anduin": {
        "type": "http",
        "url": "https://mcp.anduin.app/mcp",
        "headers": {
            "Authorization": "Bearer ${ANDUIN_TOKEN}"
        }
    }
}
```

### Tool Naming Convention
MCP tools follow the pattern: `mcp__<server-name>__<tool-name>`

Examples:
- `mcp__anduin__list_funds`
- `mcp__anduin__query_dashboard`
- `mcp__anduin__dr_list_datarooms`

### Tool Filtering
```python
# Allow all tools from Anduin MCP server
allowedTools=["mcp__anduin__*"]

# Allow only read tools
allowedTools=[
    "mcp__anduin__list_funds",
    "mcp__anduin__get_fund_info",
    "mcp__anduin__list_orders",
    "mcp__anduin__query_dashboard",
]
```

## Pricing

| Component | Cost |
|-----------|------|
| Session runtime | $0.08 per session-hour (idle time free) |
| Claude Sonnet 4.6 input | $3 per million tokens |
| Claude Sonnet 4.6 output | $15 per million tokens |
| Claude Opus 4.6 input | $5 per million tokens |
| Claude Opus 4.6 output | $25 per million tokens |
| Web search (optional) | $10 per 1,000 searches |

## Error Handling

Common errors:
- `401 Unauthorized` — Invalid API key or expired token
- `403 Forbidden` — Insufficient permissions
- `404 Not Found` — Agent/session/vault not found
- `429 Rate Limited` — Too many requests
- `500 Internal Error` — Server-side failure (retry with backoff)

For MCP-specific errors, the agent receives JSON-RPC error codes:
- `-32001` — Authentication failed (OAuth token expired)
- `-32002` — Insufficient scope (need additional OAuth scopes)
