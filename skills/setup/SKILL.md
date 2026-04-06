---
name: setup
description: "Advanced: switch the Anduin MCP server to a non-production environment (EU, Staging, Minas Tirith, Local Dev). By default the plugin connects to Anduin Production (US) with no setup needed. Use this when the user wants to change environments, or encounters MCP connection or OAuth2 issues."
argument-hint: "[environment]"
allowed-tools:
  - Bash
  - Read
  - Edit
  - Write
---

# Anduin Environment Setup (Advanced)

The plugin connects to **Anduin Production (US)** by default — no configuration needed. This skill is for advanced users (typically developers) who need to switch to a different environment.

**Important:** This only works in **Claude Code** (CLI). Cowork users connect to Production automatically.

## Step 1: Check current configuration

Check if the user has a custom environment configured:

1. Run `echo $ANDUIN_MCP_URL` to check for an env var override
2. Read the plugin's `.mcp.json` to see the current URL

- If `ANDUIN_MCP_URL` is set: tell the user which environment they're connected to (match URL against the table below) and ask if they want to change it.
- If not set: tell the user they're using the default (Production US) and ask which environment they want to switch to.

## Step 2: Ask which environment

If the user provided an environment as an argument (e.g., `/anduin:setup staging`), map it directly. Otherwise, ask the user to choose:

| Choice | Environment | URL |
|---|---|---|
| 1 | Production (US) *(default)* | `https://mcp.anduin.app/mcp` |
| 2 | Production (EU) | `https://mcp.eu.anduin.app/mcp` |
| 3 | Staging | `https://mcp-staging.anduin.dev/mcp` |
| 4 | Minas Tirith (daily bounce) | `https://mcp-minas-tirith.anduin.dev/mcp` |
| 5 | Local Development | `http://gondor-local.io:8080/mcp` |

Keyword mapping (case-insensitive): "prod"/"production"/"us" -> 1, "eu" -> 2, "staging"/"internal" -> 3, "minas"/"minas-tirith"/"daily" -> 4, "local"/"dev" -> 5.

If the user picks Production (US), tell them that's already the default — no changes needed. They can remove any existing `ANDUIN_MCP_URL` from their shell profile to revert to default.

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
   - Append to the end of the file:
     ```
     # Anduin MCP server (overrides plugin default of Production US)
     export ANDUIN_MCP_URL="<chosen-url>"
     ```

## Step 4: Update local MCP config

The plugin's `.mcp.json` hardcodes Production US. To override it for this user, add the custom URL to their user-level MCP settings:

1. Read `~/.claude.json` (create if it doesn't exist)
2. Add or update the `anduin` entry under `mcpServers`:
   ```json
   {
     "mcpServers": {
       "anduin": {
         "type": "url",
         "url": "<chosen-url>"
       }
     }
   }
   ```
3. Merge carefully — don't overwrite existing MCP entries in the file.

## Step 5: Confirm and instruct restart

Tell the user:

> Your Anduin server is now configured for **[environment name]**.
>
> **To activate the change:**
> 1. Close this Claude Code session
> 2. Open a **new terminal window** (so your shell loads the updated config)
> 3. Start Claude Code from that new terminal
>
> After restarting, the Anduin MCP will connect to **[environment name]** and a browser window will open for you to sign in.
>
> To revert to Production US, remove the `ANDUIN_MCP_URL` line from your shell profile and the `anduin` entry from `~/.claude.json`, then restart.

## Troubleshooting

If the user reports issues, help with these:

### Anduin MCP not showing in `/mcp`
- The plugin hardcodes Production US, so the MCP should always appear after install
- If missing: reinstall the plugin (`/plugin install anduin@anduin-marketplace`)
- If using a custom environment: check that `~/.claude.json` has the correct `anduin` MCP entry

### "MCP server not found" or connection errors
- Check reachability: `curl -s -o /dev/null -w "%{http_code}" https://mcp.anduin.app/mcp/.well-known/oauth-protected-resource`
- If using a custom URL: `curl -s -o /dev/null -w "%{http_code}" $ANDUIN_MCP_URL/.well-known/oauth-protected-resource`
- Local Development URLs only work in Claude Code (not Cowork)

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
