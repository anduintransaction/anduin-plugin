# Review: ChatGPT and Codex Support Plan

**Reviewed document:** `docs/plans/chatgpt-codex-support.md` (Status: Proposed, 2026-09-03)
**Review date:** 2026-09-03
**Method:** Every server-side claim was checked against the `stargazer` checkout (`upstream/master` lineage) and the
`mcp-ui-scala` SDK. Every OpenAI claim was checked against the current pages under
`https://developers.openai.com/plugins/`.

## Verdict

The plan's diagnosis of the server is accurate, and the OpenAI packaging facts hold against the current docs. The
weaknesses are in the design decisions it leaves open and in one architectural collision it does not see: ChatGPT
publishes a **snapshot** of the tool list at submission time, which contradicts the plan's central "live `tools/list`
is authoritative" principle. Items 1 to 3 below should be resolved before Phase 1 is scoped.

## Confirmed as correct

| Claim in plan | Evidence |
|---|---|
| DCR allowlist is localhost plus Anthropic domains only | `GondorBackendConfig.scala:1182`, `reference.conf:1829`; `HydraPublicServer.validateRedirectUris` matches domain-wide including subdomains |
| PRM `authorization_servers` and AS `issuer` disagree | PRM emits `HydraUrls.envScopedAuthServerUrl` (`{publicUrl}/hydra/env/{envId}`, default env when none, so production is affected); AS metadata emits bare `urls.publicUrl` (`HydraPublicServer.scala:55`) |
| Tool wire model lacks title, annotations, securitySchemes, outputSchema | `McpTypes.scala:71` (`McpToolDefinition`: name, description, inputSchema, `_meta`) |
| `RuntimeTool.isReadOnly` exists; no destructive/open-world flag exists | `TypedTool.scala:128`; no `destructive`/`openWorld` symbol anywhere under `modules/mcp` |
| Unauthenticated requests are rejected | `McpEndpointServer.scala:223` fails with `UnauthenticatedError` (HTTP 401 + `WWW-Authenticate`) |
| Agents duplicate skills | `agents/gp-assistant.md` (275 lines) repeats terminology, workflows, and the whole UI-rendering section of the skill |
| Skills contain host-specific prose | `skills/gp-assistant/SKILL.md:153`, `skills/dataroom/SKILL.md:73` name Claude Code and Cowork |
| One universal OpenAI listing; 5 positive + 3 negative tests; agents unsupported | `plugins/deploy/submission`, `plugins/guides/submit-claude-plugin` |
| File layout: `.codex-plugin/plugin.json`, `.app.json`, `.agents/plugins/marketplace.json` | `plugins/build/plugins` |
| Widgets work in ChatGPT | ChatGPT "implements the open MCP Apps standard"; `_meta.ui.resourceUri` is preferred, `openai/outputTemplate` is a compatibility alias. Codex does not render UI (docs: keep tools useful without a component "so ChatGPT and Codex can complete the workflow without UI") |

## Significant gaps and errors

### 1. Snapshot vs. runtime discovery (highest priority)

OpenAI's **Scan Tools** imports tool metadata into the draft, and "the published plugin uses this metadata snapshot
while tool calls and UI resources continue to use your live MCP server." Anduin's `tools/list` is per-identity,
scope-filtered, and returns 401 without a token. The plan never reconciles these. Consequences:

- The scan must be performed by an account holding every scope (`fundsub:admin`, `dataroom:admin`, `mcp:render`), or
  the listing permanently lacks tools.
- A `fundsub:read` user sees write tools in ChatGPT and hits insufficient-scope on call. In-band
  `_meta["mcp/www_authenticate"]` (with both `error` and `error_description`) is therefore **mandatory**, not
  "where required" as Phase 1.4 says.
- Any tool addition or description change requires a rescan and resubmission. The plan's release process has no step
  for this.
- Whether the scanner authenticates at all is unverified. If it is unauthenticated, Anduin faces a security decision
  about serving an anonymous catalog with `securitySchemes`. Make this a Phase 0 spike.

### 2. Issuer fix is left undecided, and the stated constraint is misdirected

Phase 1.2 says "do not simply change the metadata issuer to the scoped path unless Hydra token issuance and validation
use that same issuer." ChatGPT never inspects the access token's `iss`; it only compares the metadata `issuer` string
to `authorization_servers[0]`. Returning the env-scoped URL as `issuer` is the direct fix and does not affect token
validation on the resource server.

The real consequence is elsewhere: `authorization_response_iss_parameter_supported` must **not** be advertised,
because Hydra's RFC 9207 `iss` parameter would still be the bare host. ChatGPT then uses the callback-ID redirect
`https://chatgpt.com/connector/oauth/{callback_id}`, which the existing domain-wide allowlist matching handles. The
plan should commit to this path and record why the stable redirect is out of scope for the first release.

### 3. `.mcp.json` schema collision

Claude's `.mcp.json` wraps servers in `mcpServers`. OpenAI's plugin `.mcp.json` accepts a bare server map or a
`mcp_servers` (snake case) wrapper, and the Claude-conversion guide lists `.mcp.json` as unsupported. The plan keeps a
single file for both hosts without addressing this.

Decide explicitly: omit `mcpServers` from `.codex-plugin/plugin.json` so OpenAI ignores the Claude file, and have both
OpenAI hosts connect via `.app.json`. Then **verify that Codex CLI and IDE honor `.app.json` for a remote OAuth
server**. If they do not, Codex is effectively unsupported and the "one universal package" premise fails.

### 4. Codex OAuth surfaces are not analyzed

Codex CLI and IDE authenticate with loopback redirects (covered by the existing `localhost` allowlist). Codex cloud
goes through chatgpt.com. Phase 4.3 treats Codex as a single row. Split it into CLI, IDE, and cloud, each with its own
OAuth verification.

### 5. `outputSchema` pulls in a third repository

OpenAI: "Declare `outputSchema` for any tool that returns `structuredContent`." Every render and `show_*` tool does.
The spec validators (`TableSpec`, `ChartSpec`, `FormSpec`) live in `mcp-ui-scala`, so the SDK must export JSON Schema,
publish, and stargazer must bump `McpUi.version` in `build/versions.mill`. The plan lists two repositories and two PRs;
it needs three. Per the MCP spec, once `outputSchema` is declared, `structuredContent` must conform exactly.

### 6. Clickable widget links

OpenAI's reference states redirect allowlisting still requires the legacy `_meta["openai/widgetCSP"].redirect_domains`
key even though CSP otherwise moved to `_meta.ui.csp`. The `show_*` tables open deep links via `app.openLink`. Without
that key, links likely fail in ChatGPT. Not mentioned in Phase 1.5. This is another SDK/`UiResourceRegistry` change.

### 7. Where annotations live

Phase 1.3 says "add explicit reviewed fields" without saying where. Adding them to `RuntimeTool` leaks MCP-only
metadata into `reagentCore`, which internal agents share. Recommended:

- A reviewed table in `PublicMcpServerConfig`, next to `publicDataroomAdminToolNames`.
- Derive `securitySchemes` scopes from the existing read/write/admin allowlist sets rather than a second hand-kept list.
- Contract tests: every admin-bucket tool has `destructiveHint=true`; every `isReadOnly` tool has
  `destructiveHint=false`; every public tool has an entry.

### 8. Skills are also a snapshot

ChatGPT skills are attached at submission. Updating canonical skill text means resubmitting. The "skills are the source
of truth" model needs a release-process note about this, alongside item 1.

## Smaller corrections

- **Phase 0 branch note.** The local `origin` is the personal fork `cmpham/stargazer`; the 415-commit lead is against
  the fork's stale master, not necessarily upstream. Name `upstream/master` (`anduintransaction/stargazer`) explicitly.
- **Phase 2.3 drift check is unnecessary** once agents shrink to frontmatter plus one line invoking the skill. There is
  nothing left to drift.
- **Phase 3.6 validator** must not require Claude-only frontmatter (`allowed-tools`, `argument-hint`) on the OpenAI
  side; OpenAI documents only `name` and `description`.
- **Chicken-and-egg between 1.1 and 3.1** is already resolved by the domain-wide allowlist (the callback ID is not
  needed to add `chatgpt.com`). State this so the phases can proceed in parallel.
- **Review account without MFA** conflicts with Anduin org login policies. Provision a dedicated demo org with
  synthetic data early; list it as a Phase 0 dependency, not a Phase 5 item.
- **Phase 1.2 golden tests** must run with production-shaped `advertisedMcpHost` / `x-anduin-forwarded-host`, since the
  PRM `resource` value depends on them.

## Recommended plan changes (summary)

1. Add a Phase 0 spike: register the current server in ChatGPT developer mode, observe whether Scan Tools
   authenticates, and confirm whether Codex honors `.app.json` for a remote OAuth server.
2. Commit to metadata `issuer` = env-scoped URL, no RFC 9207 advertisement, callback-ID redirect.
3. Add `mcp-ui-scala` as a third repository and a PR 0 for `outputSchema` export and `redirect_domains`.
4. Make in-band `_meta["mcp/www_authenticate"]` a hard Phase 1 requirement.
5. Move tool classification into `PublicMcpServerConfig` with derivation from the allowlists and contract tests.
6. Add rescan/resubmission steps to the release process for tool and skill changes.
7. Split Codex into CLI, IDE, and cloud rows in the host matrix.

## Sources checked

- `stargazer`: `modules/mcp/mcp/jvm/src/anduin/mcp/{protocol/McpTypes.scala, auth/OAuth2McpAuthHttpExtension.scala,
  server/McpEndpointServer.scala, server/PublicMcpServerConfig.scala}`,
  `modules/heimdall/heimdall/jvm/src/anduin/oauth2/hydra/HydraUrls.scala`,
  `modules/heimdall/heimdallApp/jvm/src/anduin/oauth2/hydra/HydraPublicServer.scala`,
  `modules/reagent/reagentCore/jvm/src/anduin/reagent/core/TypedTool.scala`,
  `platform/stargazerConfig/{shared/src/.../GondorBackendConfig.scala, jvm/resources/reference.conf}`
- `mcp-ui-scala`: `docs/2026-06-23-mcp-apps-rendering-findings.md` (note: its "ChatGPT needs `openai/outputTemplate`"
  finding is superseded by OpenAI's current docs)
- OpenAI: `plugins/build/{plugins,auth,skills,chatgpt-ui}`, `plugins/deploy/{app-review,submission,connect-chatgpt}`,
  `plugins/guides/{submit-claude-plugin,optimize-metadata}`, `plugins/reference`
