# Anduin Plugin for Claude

Manage Anduin fund subscriptions and data rooms from **Claude Code** or **Cowork**, using your existing Anduin account and permissions.

This package includes two shared workflow skills and Claude agents that load them automatically. ChatGPT/Codex installation packaging is not included.

## Capabilities

- **GP Assistant:** review LP status, forms, documents and AML/KYC results; analyze fund reports and dashboards; update form fields, order tags and custom columns; post comments and invite fund managers.
- **Data Room:** browse and search rooms/files; create folders and rename items; manage rooms and participants; read PDFs, images and Excel files; view engagement analytics. Analytics require a joined Admin and the room's Insights plan.
- **Presentation:** charts, tables and form-style summaries where the host supports widgets, with text/Markdown fallback otherwise. Widgets are display-only: they do not submit forms or change records.

Available actions depend on the connected server, granted scopes and your product access. The assistants preview changes for confirmation and report partial or uncertain outcomes. They do not provide subscription approval/countersigning, file upload/move, or folder restoration.

## Install and connect

### Claude Code

Run inside Claude Code:

```text
/plugin marketplace add anduintransaction/anduin-plugin
/plugin install anduin@anduin-marketplace
```

Follow the activation prompt, then use `/mcp` to select the Anduin connection and sign in through your browser. See Anthropic's [plugin installation guide](https://code.claude.com/docs/en/discover-plugins) and [MCP authentication guide](https://code.claude.com/docs/en/mcp#authenticate-with-remote-mcp-servers).

### Cowork

Open **Customize → Plugins**, add `anduintransaction/anduin-plugin` as a marketplace, and install **anduin**. Follow the connection prompts to sign in. See Anthropic's [plugin guide](https://support.claude.com/en/articles/13837440-use-plugins-in-claude).

The bundled connection uses **Anduin Production (US)** at `https://mcp.anduin.app/mcp`. Sign in with the Anduin account that has access to the funds or rooms you need; do not paste passwords or tokens into chat.

## Usage

Describe your task naturally, or invoke `/anduin:gp-assistant` or `/anduin:dataroom` in Claude Code. In Cowork, select the corresponding skill from the skills menu.

- “Which LPs in Venture Fund III have incomplete forms?”
- “Summarize the fund's subscription report and chart order counts by close.”
- “Preview adding Reviewed to these orders, keeping their existing tags.”
- “Who has access to the Acme data room?”
- “Read the NDA and summarize the pages you reviewed.”
- “Prepare invitations for these participants as Observers.”

## Permissions

Consent shows the scopes requested by the connection, not every supported permission or one entry per tool. Approve only what your task needs.

| Scope | Access |
|---|---|
| `fundsub:read` | Subscription data, forms, documents and reports |
| `fundsub:write` | Form/comment/tag/custom-column changes and manager invitations |
| `fundsub:admin` | Includes fund write/read; currently adds no tools beyond write |
| `dataroom:read` | Room contents, participants and permitted analytics |
| `dataroom:write` | Create/rename rooms and folders, rename items, invite users and restore eligible files |
| `dataroom:admin` | Archive/unarchive rooms, delete items, remove users and change roles |
| `mcp:render` | Display-only visuals; no domain-data access |

Within each domain, admin includes write and read; write includes read. Rendering is separate. OAuth consent does **not** grant fund/room roles or upgrade your plan.

## Troubleshooting and updates

- **Missing connection or tools:** check that the plugin is enabled and connected (`/plugin` and `/mcp` in Claude Code). Availability can also depend on scopes and the server's tool catalog.
- **Expired login or insufficient scopes:** reconnect through the host and approve the required access. For product-role or plan denials, contact your fund/room administrator instead.
- **No widget appears:** ask for the result as text or a Markdown table; a successful rendering call does not guarantee the host displayed it.
- **Update the plugin:** use Claude Code's [marketplace update controls](https://code.claude.com/docs/en/discover-plugins#configure-auto-updates), or manage the installed plugin in Cowork's Plugins settings. Follow any reload/restart prompt.

## License

[MIT](LICENSE).
