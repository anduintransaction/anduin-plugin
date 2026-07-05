# Claude Managed Agents API Reference

All endpoints require the `managed-agents-2026-04-01` beta header.
The Python and TypeScript SDKs set this automatically.

## Canonical References

Verify against the official docs whenever anything here looks stale:

- Overview: https://platform.claude.com/docs/en/managed-agents/overview
- Quickstart: https://platform.claude.com/docs/en/managed-agents/quickstart
- Vaults: https://platform.claude.com/docs/en/managed-agents/vaults
- MCP connector: https://platform.claude.com/docs/en/managed-agents/mcp-connector
- Reference: https://platform.claude.com/docs/en/managed-agents/reference

## Agents

### Create Agent
```
POST /v1/agents
```

Parameters:
- `name` (string, required) — Display name
- `model` (string, required) — Model ID (e.g. `claude-sonnet-5`)
- `system` (string) — System prompt
- `tools` (array) — Toolsets: the built-in `agent_toolset_20260401`, plus one `mcp_toolset` entry (with `mcp_server_name` matching a declared server) per MCP server
- `mcp_servers` (array) — MCP server connections: `{"type": "url", "name": ..., "url": ...}`. No auth tokens here — credentials come from vaults at session time
- `skills` (array) — Agent skills

Tool access is controlled per toolset entry in `tools` (there is no `allowedTools`/`disallowedTools`):
- `default_config.permission_policy` — `{"type": "always_allow"}` or `{"type": "always_ask"}`. MCP toolsets default to `always_ask`; the agent toolset defaults to `always_allow`
- `default_config.enabled` plus a per-tool `configs` array (`{"name": ..., "enabled": ..., "permission_policy": ...}`) — enable/disable or override policy for individual tools. See "Tool Permissions & Filtering" below

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
- `config` (object) — Sandbox configuration, e.g. `{"type": "cloud", "packages": {...}, "networking": {...}}`
  - `packages` (object) — Pre-installed packages keyed by package manager (e.g. `{"pip": ["pandas"], "npm": ["express"]}`)
  - `networking` (object) — Outbound network access. Declared on the ENVIRONMENT (there is no agent-level `network_policy`):
    - `{"type": "unrestricted"}` — full outbound access except a safety blocklist (default)
    - `{"type": "limited", "allowed_hosts": [...], "allow_mcp_servers": bool, "allow_package_managers": bool}` — restrict the sandbox to the listed hosts; `allow_mcp_servers` additionally permits the agent's declared MCP endpoints. Recommended for production

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
- `vault_ids` (array) — Credential vault IDs. Credentials from these vaults are matched to the agent's `mcp_servers` entries by URL and injected automatically (see Vaults below)

### Send Events
```
POST /v1/sessions/{session_id}/events
```

Event types (client → session):
- `user.message` — User turn with text/image content
- `user.tool_confirmation` — Approve or deny a tool call paused by an `always_ask` permission policy (`result: "allow" | "deny"`)
- `user.custom_tool_result` — Response to an `agent.custom_tool_use` call (custom tools only; Anduin tools are MCP tools executed server-side — the vault credential matching the server URL is injected automatically, no client-side result needed)
- `user.tool_result` — Pre-built `agent_toolset` results, `self_hosted` environments only (the SDK/CLI provide these automatically)

### Stream Events (SSE)
```
GET /v1/sessions/{session_id}/events
```

Event types received (SSE stream):
- `agent.thinking` — Agent reasoning
- `agent.message` — Agent response text
- `agent.tool_use` / `agent.tool_result` — Pre-built agent tool (bash, file ops) invocation and result
- `agent.mcp_tool_use` / `agent.mcp_tool_result` — MCP server tool invocation and result (Anduin tool calls surface here)
- `session.status_idle` — Agent finished the turn and is awaiting input; carries a `stop_reason`. There is no `agent.completed` event — treat `session.status_idle` as the turn-complete signal. A `stop_reason.type` of `requires_action` means the session is paused on `always_ask` tool calls and waits for `user.tool_confirmation` events

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
- `display_name` (string, required) — Vault name
- `metadata` (object) — Optional tags to map the vault back to your own user records

### Add Credentials
```
POST /v1/vaults/{vault_id}/credentials
```

MCP credentials are keyed by `mcp_server_url` and injected by URL match: when the agent
connects to an MCP server whose declared `url` exactly matches a credential's
`mcp_server_url`, that credential authenticates the connection automatically. There is
no proxy layer, and no tokens ever appear in the agent definition.

Parameters:
- `display_name` (string) — Credential name
- `auth` (object, required) — One of:
  - `{"type": "static_bearer", "mcp_server_url": ..., "token": ...}` — fixed bearer token (API key / PAT). Does NOT auto-refresh
  - `{"type": "mcp_oauth", "mcp_server_url": ..., "access_token": ..., "expires_at": ..., "refresh": {...}}` — OAuth 2.0. With a `refresh` block (`token_endpoint`, `client_id`, `scope`, `refresh_token`, `token_endpoint_auth`), Anthropic refreshes the access token automatically when it expires — use this for long-running/scheduled agents

Secret values (`token`, `access_token`, `refresh_token`, `client_secret`) are write-only
and never returned by the API. See the vaults doc (Canonical References above) for the
full credential shapes and rotation/validation endpoints.

## MCP Server Configuration

### Declaring the Anduin MCP server on an agent
```python
agent = client.beta.agents.create(
    name="anduin-agent",
    model="claude-sonnet-5",
    mcp_servers=[
        {"type": "url", "name": "anduin", "url": "https://mcp.anduin.app/mcp"},
    ],
    tools=[
        {"type": "agent_toolset_20260401"},
        {"type": "mcp_toolset", "mcp_server_name": "anduin"},
    ],
)
```

No `Authorization` header is configured on the agent. Authentication is supplied at
session creation via `vault_ids`: the vault credential whose `mcp_server_url` matches
the declared `url` is injected automatically (see Vaults above). Never hardcode tokens
in agent definitions.

### Tool Naming Convention
MCP tools follow the pattern: `mcp__<server-name>__<tool-name>`

Examples:
- `mcp__anduin__list_funds`
- `mcp__anduin__query_dashboard`
- `mcp__anduin__dr_list_datarooms`

### Tool Permissions & Filtering
There is no `allowedTools`/`disallowedTools`. Access is controlled on the agent's
`mcp_toolset` entry via `permission_policy` and per-tool `configs`:

```python
# Auto-approve all Anduin tools. MCP toolsets default to always_ask, which
# pauses the session for a user.tool_confirmation — unusable for unattended
# scheduled/event-driven agents, so trusted servers are set to always_allow.
{
    "type": "mcp_toolset",
    "mcp_server_name": "anduin",
    "default_config": {"permission_policy": {"type": "always_allow"}},
}

# Least privilege: enable only specific read tools, everything else off
{
    "type": "mcp_toolset",
    "mcp_server_name": "anduin",
    "default_config": {"enabled": False},
    "configs": [
        {"name": "list_funds", "enabled": True},
        {"name": "get_fund_info", "enabled": True},
        {"name": "list_orders", "enabled": True},
        {"name": "query_dashboard", "enabled": True},
    ],
}
```

`configs` entries use the bare tool name as reported by the server (`list_funds`, not
`mcp__anduin__list_funds`). Each entry can also carry its own `permission_policy` to
override the toolset default for that one tool.

## Pricing

| Component | Cost |
|-----------|------|
| Session runtime | $0.08 per session-hour (metered only while `running`; idle time free) |
| Claude Sonnet 5 input | $2 per million tokens through Aug 31, 2026, then $3 |
| Claude Sonnet 5 output | $10 per million tokens through Aug 31, 2026, then $15 |
| Claude Opus 4.8 input | $5 per million tokens |
| Claude Opus 4.8 output | $25 per million tokens |
| Web search (optional) | $10 per 1,000 searches |

Check current rates at https://platform.claude.com/docs/en/about-claude/pricing before
estimating costs — model pricing changes over time.

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
