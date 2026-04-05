---
name: Anduin MCP Setup
description: Use when the user asks how to set up, configure, or connect to the Anduin MCP server, or when they encounter OAuth2 authentication issues, MCP connection errors, or need to change the Anduin server URL.
---

# Anduin MCP Server Setup

## Prerequisites

- Claude Code CLI installed
- Access to an Anduin platform environment (production, staging, or development)
- A user account on the Anduin platform with appropriate permissions

## Configuration

### Set the MCP Server URL

The plugin connects to the Anduin MCP server via the `ANDUIN_MCP_URL` environment variable. Set it to your environment's MCP endpoint:

**Production:**
```bash
export ANDUIN_MCP_URL="https://gondor-public.anduintransact.com/mcp"
```

**Staging/Feature:**
```bash
export ANDUIN_MCP_URL="https://mordor.anduin.dev/mcp"
```

**Local Development:**
```bash
export ANDUIN_MCP_URL="http://gondor-local.io:8080/mcp"
```

Add the export to your shell profile (`~/.zshrc`, `~/.bashrc`) for persistence.

### Verify MCP Connection

After setting the URL, restart Claude Code and check the MCP connection:
1. Run `/mcp` in Claude Code to see connected MCP servers
2. Look for `anduin` in the server list
3. The server should show available tools prefixed with `mcp__anduin__`

### OAuth2 Authentication

The Anduin MCP server uses OAuth2 with PKCE for authentication. Claude Code handles the entire flow automatically:

1. **Auto-discovery**: Claude Code reads `/.well-known/oauth-protected-resource` to find the authorization server
2. **Dynamic registration**: Claude Code registers itself as an OAuth2 client via RFC 7591 DCR
3. **Authorization**: A browser window opens for you to log in with your Anduin credentials
4. **Token management**: Claude Code manages access and refresh tokens automatically

No manual OAuth2 configuration is needed.

### Available Scopes

When authorizing, you may be asked to approve these scopes:

| Scope | Description |
|-------|-------------|
| `fundsub:read` | View fund subscription data (orders, forms, documents) |
| `fundsub:write` | Modify fund subscriptions (update forms, tags, invite managers) |
| `fundsub:admin` | Full fund subscription admin access |
| `dataroom:read` | View data room contents (files, participants, analytics) |
| `dataroom:write` | Modify data rooms (create, invite, upload, delete) |
| `dataroom:admin` | Full data room admin access |

Scope hierarchy: `admin` implies `write` implies `read`.

## Troubleshooting

### "MCP server not found"
- Verify `ANDUIN_MCP_URL` is set: `echo $ANDUIN_MCP_URL`
- Restart Claude Code after setting the variable
- Check the URL is reachable: `curl -s $ANDUIN_MCP_URL/.well-known/oauth-protected-resource`

### "Token validation failed" or 401 errors
- Your OAuth2 token may have expired — Claude Code should auto-refresh
- Try restarting Claude Code to trigger a fresh OAuth2 flow
- Verify your Anduin account has the necessary permissions

### "Insufficient scopes" or 403 errors
- The tool requires a scope you didn't authorize
- Re-authorize with broader scopes (the MCP server will prompt)
- Check with your fund administrator if you need additional access

### Tools not appearing
- Verify the MCP server is connected via `/mcp`
- Tools are filtered by your OAuth2 scopes — you only see tools you have access to
- DataRoom tools are prefixed `dr_` (e.g., `mcp__anduin__dr_list_datarooms`)
- FundSub tools have no prefix (e.g., `mcp__anduin__list_funds`)
