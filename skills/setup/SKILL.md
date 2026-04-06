---
name: setup
description: Set up the Anduin MCP server connection. Run this to configure which Anduin environment to connect to (Production, Staging, Minas Tirith, or Local Dev). Also use when the user encounters MCP connection errors, OAuth2 authentication issues, or asks how to change the Anduin server URL.
argument-hint: "[environment]"
allowed-tools:
  - Bash
  - Read
  - Edit
  - Write
---

# Anduin MCP Setup

You are guiding the user through connecting the Anduin plugin to their Anduin server. Follow these steps in order.

## Step 1: Check current configuration

Run `echo $ANDUIN_MCP_URL` to see if a server URL is already configured.

- If set: tell the user which environment they're connected to (match URL against the table below) and ask if they want to change it.
- If empty: tell the user the plugin is not configured yet and proceed to Step 2.

## Step 2: Ask which environment

If the user provided an environment as an argument (e.g., `/anduin-plugin:setup staging`), map it directly. Otherwise, ask the user to choose:

| Choice | Environment | URL |
|---|---|---|
| 1 | Production (US) | `https://mcp.anduin.app/mcp` |
| 2 | Production (EU) | `https://mcp.eu.anduin.app/mcp` |
| 3 | Staging | `https://mcp-staging.anduin.dev/mcp` |
| 4 | Minas Tirith (daily bounce) | `https://mcp-minas-tirith.anduin.dev/mcp` |
| 5 | Local Development | `http://gondor-local.io:8080/mcp` |

Keyword mapping (case-insensitive): "prod"/"production"/"us" -> 1, "eu" -> 2, "staging"/"internal" -> 3, "minas"/"minas-tirith"/"daily" -> 4, "local"/"dev" -> 5.

If the user is on Cowork, note that only options 1-4 work (Local Development requires direct network access).

## Step 3: Write to shell profile

1. Detect the shell: check `$SHELL`
   - Contains "zsh" -> `~/.zshrc`
   - Contains "bash" -> `~/.bashrc`
   - Otherwise -> `~/.zshrc` (default)

2. Read the profile file to check if `ANDUIN_MCP_URL` is already defined.

3. If an existing `export ANDUIN_MCP_URL=` line exists:
   - Use the Edit tool to replace that line with the new URL
   - Tell the user you updated the existing setting

4. If no existing line:
   - Use the Edit tool or Bash to append these two lines to the end of the file:
     ```
     # Anduin MCP server
     export ANDUIN_MCP_URL="<chosen-url>"
     ```
   - Tell the user you added the setting

## Step 4: Confirm and instruct restart

Tell the user:

> Your Anduin server is now configured for **[environment name]**.
>
> Please restart Claude Code (or Cowork) for the change to take effect. After restarting, a browser window will open for you to sign in with your Anduin credentials.

## Troubleshooting

If the user reports issues after setup, help with these:

### "MCP server not found" or connection errors
- Verify the URL is set: `echo $ANDUIN_MCP_URL`
- Check reachability: `curl -s -o /dev/null -w "%{http_code}" $ANDUIN_MCP_URL/.well-known/oauth-protected-resource`
- If using Cowork with a local dev URL, explain that Cowork only works with public URLs

### "Unauthorized" or 401 errors
- Session may have expired — restart to trigger a fresh sign-in
- Verify Anduin account has the necessary permissions

### "Insufficient scopes" or 403 errors
- User needs broader permissions — restart and approve additional scopes
- Contact fund administrator for access

### Tools not appearing
- Run `/mcp` in Claude Code to check server status
- Tools are filtered by approved scopes — user only sees what they have access to
- DataRoom tools are prefixed `dr_`, FundSub tools have no prefix

### OAuth2 authentication

Authentication is fully automatic:
1. Claude reads `/.well-known/oauth-protected-resource` to find the authorization server
2. Claude registers itself as an OAuth2 client automatically
3. A browser window opens for sign-in
4. Tokens are managed and refreshed automatically

Available scopes:

| Scope | What it allows |
|---|---|
| `fundsub:read` | View fund subscription data |
| `fundsub:write` | Modify fund subscriptions |
| `fundsub:admin` | Full fund subscription admin access |
| `dataroom:read` | View data room contents |
| `dataroom:write` | Modify data rooms |
| `dataroom:admin` | Full data room admin access |

Scope hierarchy: `admin` implies `write` implies `read`.
