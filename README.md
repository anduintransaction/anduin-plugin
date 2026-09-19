# Anduin Plugin for Claude

Manage Anduin fund subscriptions and data rooms from **Claude Code** or **Cowork**, using your existing Anduin account and permissions.

This package includes two shared workflow skills and Claude agents that load them automatically. ChatGPT/Codex installation packaging is not included.

## Capabilities

- **GP Assistant:** review LP status, forms, documents and AML/KYC results; analyze fund reports and dashboards; update form fields, order tags and custom columns; post comments and invite fund managers.
- **Data Room:** browse and search rooms/files; create folders and rename items; manage rooms and participants; read PDFs, images and Excel files; view engagement analytics. Analytics require a joined Admin and the room's Insights plan.
- **Presentation:** charts, tables and form-style summaries where the host supports widgets, with text/Markdown fallback otherwise. Widgets are display-only: they do not submit forms or change records.

Available actions depend on the connected server, granted scopes and your product access. The assistants preview changes for confirmation and report partial or uncertain outcomes. They do not provide subscription approval/countersigning, file upload/move, or folder restoration.

## Install and connect

### Choose your region

Install **one** plugin — the one for the region that hosts your Anduin account. If you are unsure which region you use, ask your Anduin contact.

| Your Anduin account is on | Install | Connects to |
|---|---|---|
| Production (US) | `anduin` | `https://mcp.anduin.app/mcp` |
| Production (EU) | `anduin-eu` | `https://mcp.eu.anduin.app/mcp` |

Both are the same plugin; the EU build differs only in the server it connects to. Do not install both: they share the same names and only one connection would load. If your organization provides Anduin through its own plugin list, remove any copy you installed yourself first.

### Claude Code

Run inside Claude Code, using the plugin for your region on the second line:

```text
/plugin marketplace add anduintransaction/anduin-plugin
/plugin install anduin@anduin-marketplace
```

For EU accounts, install `anduin-eu@anduin-marketplace` instead.

Follow the activation prompt, then use `/mcp` to select the Anduin connection and sign in through your browser. See Anthropic's [plugin installation guide](https://code.claude.com/docs/en/discover-plugins) and [MCP authentication guide](https://code.claude.com/docs/en/mcp#authenticate-with-remote-mcp-servers).

### Cowork

Open **Customize → Plugins**, add `anduintransaction/anduin-plugin` as a marketplace, and install **anduin** (US) or **anduin-eu** (EU). Follow the connection prompts to sign in. See Anthropic's [plugin guide](https://support.claude.com/en/articles/13837440-use-plugins-in-claude).

### Claude organizations

An organization Owner can offer the plugin to every member through the organization's plugin marketplace. Admins: follow the printable step-by-step guide in [docs/Anduin_MCP_Enablement_Guide_Claude.html](docs/Anduin_MCP_Enablement_Guide_Claude.html). In short:

1. In your private marketplace repository, add this entry to `.claude-plugin/marketplace.json`, with the `ref` line set to the release tag for your region:

   ```json
   {
     "name": "anduin",
     "description": "Data room management and fund subscription review for the Anduin platform",
     "source": {
       "source": "git-subdir",
       "url": "https://github.com/anduintransaction/anduin-plugin.git",
       "path": "plugins/anduin",
       "ref": "v0.10.0"
     }
   }
   ```

   | Region | `ref` |
   |---|---|
   | Production (US) | `"ref": "v0.10.0"` |
   | Production (EU) | `"ref": "v0.10.0-eu"` |

2. Connect the repository in **Organization settings → Plugins**, try the plugin yourself, then choose who gets it. Members see a single Anduin connection for the right region.

**Releases.** Every release publishes a `vX.Y.Z` tag and its `vX.Y.Z-eu` twin. The [releases page](https://github.com/anduintransaction/anduin-plugin/releases) lists both `ref` lines ready to paste, the matching commits, and a ZIP per region for organizations without GitHub.

**Updating.** Claude syncs an organization marketplace only when your own marketplace repository changes, so a new Anduin release never arrives by itself:

1. Change only the Anduin entry's `ref` to the new tag. Keep your other plugin entries and the marketplace name and owner as they are.
2. Merge the change through a pull request.
3. If automatic sync is enabled, the merge may trigger the sync; otherwise click **Update** on the marketplace in **Organization settings → Plugins**.
4. Check that the plugin shows the new version. To go back, restore the previous tag the same way.

**Strict pinning (advanced).** For a pin that cannot move, add `"sha": "<commit>"` next to `ref`, using the commit listed on the release. The `sha` takes precedence over `ref`, so from then on change **both** lines on every update, rollback or region change — a new `ref` with an old `sha` still installs the old commit.

### Sign in

Whichever way you installed the plugin, sign in with the Anduin account that has access to the funds or rooms you need; do not paste passwords or tokens into chat.

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
- **Sign-in fails or your funds and rooms are missing:** check the connection URL that `/mcp` shows against your account's region (see [Choose your region](#choose-your-region)); a wrong region means the wrong plugin, or an older copy, is active — remove it and install the right one. If the URL is right, check that you can sign in to Anduin itself and have access to the fund or room.
- **Expired login or insufficient scopes:** reconnect through the host and approve the required access. For product-role or plan denials, contact your fund/room administrator instead.
- **No widget appears:** ask for the result as text or a Markdown table; a successful rendering call does not guarantee the host displayed it.
- **Update your own install:** Claude Code does not auto-update third-party marketplaces by default. Run `/plugin`, open **Marketplaces**, select `anduin-marketplace` and choose **Enable auto-update** (see [auto-updates](https://code.claude.com/docs/en/discover-plugins#configure-auto-updates)), or, when you want the latest release, run `/plugin marketplace update anduin-marketplace` and then update the plugin from the **Installed** tab of `/plugin` (or `claude plugin update anduin@anduin-marketplace`; use `anduin-eu@anduin-marketplace` for EU). In Cowork, manage the installed plugin in Plugins settings. Follow any reload/restart prompt.
- **Update a plugin provided by your organization:** your admin controls the version (see [Claude organizations](#claude-organizations)); you receive it in your next session.
- **What changed:** see [CHANGELOG.md](CHANGELOG.md).

## Building the admin guide (PDF)

The organization admin guide is maintained as HTML in [docs/Anduin_MCP_Enablement_Guide_Claude.html](docs/Anduin_MCP_Enablement_Guide_Claude.html). To produce the PDF that is sent to customers, run this from the repository root:

```bash
scripts/render-guide.sh
```

It writes `docs/Anduin_MCP_Enablement_Guide_Claude.pdf` (Letter size; the PDF is not committed). Pass a path to write it elsewhere, for example `scripts/render-guide.sh ~/Desktop/Anduin_Guide.pdf`.

The script needs Google Chrome or Chromium. If it cannot find one, point it at the binary:

```bash
CHROME="/path/to/chrome" scripts/render-guide.sh
```

Render the PDF only after the release it names has been published, so the tags in the guide exist. `scripts/build-eu-plugin.sh --check` confirms the guide names the current release.

## License

[MIT](LICENSE).
