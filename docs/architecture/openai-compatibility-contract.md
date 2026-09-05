# OpenAI Compatibility Contract (Phase 0 record)

**Status:** Phase 0A released (`v0.5.0`); Phase 1 fixes published in [PR #55982](https://github.com/anduintransaction/stargazer/pull/55982), local checks passed, CI/review pending. Live D5 validation,
spikes 1–3, D17 distribution approval, and packaging remain open. The plan's G0–G5 table is the readiness index.
**Recorded:** 2026-09-03, updated 2026-09-05
**Plan:** `docs/plans/chatgpt-codex-support.md` (Phase 0)
**Review:** `docs/reviews/chatgpt-codex-support-review.md`
**Repositories:** `mcp-ui-scala`, `stargazer`, `anduin-plugin`

This note is the "architecture note" Phase 0 asks for. Section 1 locks the decisions. Sections 2–4 record the
facts they rest on, with each probe's original date retained (not a fresh production verification). Section 6 is
the runbook and log for the spikes that need real ChatGPT/Codex accounts.

## 1. Locked decisions

| # | Decision | Value |
|---|---|---|
| D1 | Packaging target | One canonical behavior source and one universal public listing for ChatGPT and Codex. Identical package/import wiring across all surfaces is unproven (D13). A private pilot or a reduced host matrix requires explicit approval and is not completion of the public goal. |
| D2 | Target environment | First release targets production US only: `https://mcp.anduin.app/mcp`. EU (`mcp.eu.anduin.app`) is out of scope until US is stable. |
| D3 | Client registration | Dynamic Client Registration (RFC 7591) with PKCE S256. Client ID Metadata Documents deferred. |
| D4 | OAuth scopes | Scope names and authority are unchanged: `fundsub:read/write/admin`, `dataroom:read/write/admin` (hierarchical), `mcp:render` (flat). New grants materialize lower scopes; old tokens still receive scope expansion. This does not promise compatibility for tokens with a different resource audience: inventory base/custom-host legacy tokens and plan reconnects before rollout. Destructive dataroom tools retain `dataroom:admin`. |
| D5 | Issuer — **provisional** | Candidate design: authorization-server metadata `issuer` becomes that environment's scoped identifier (production example: `https://id.anduin.app/hydra/env/default`) so it equals the PRM `authorization_servers` entry; RFC 9207 is not advertised and token validation stays strict. Do not promote this change beyond the public non-production spike environment until successful and error authorization redirects prove that any returned `iss` is either absent or exactly environment-scoped. Codex rejects every mismatched returned `iss`, even when issuer support is not advertised. If Hydra returns the bare-host issuer, stop and redesign; a metadata-only production change is forbidden. |
| D6 | ChatGPT redirect | Callback-specific `https://chatgpt.com/connector/oauth/{callback_id}` (the stable `connector_platform_oauth_redirect` needs RFC 9207, excluded by D5). DCR allowlist gains the `chatgpt.com` domain only, and the validator pins it to that exact host and path (Phase 1). |
| D7 | Discovery model | Scan Tools output is the authoritative catalog for the published OpenAI plugin in both ChatGPT and Codex. Live `tools/list` stays authoritative for Claude and direct Codex MCP connections. Execution authorization is enforced server-side for every host. |
| D8 | Test and review accounts | Separate synthetic non-production and production-US demo identities, with only the scope/roles needed for the approved test catalog and no customer data. Any review-only no-MFA arrangement needs a narrowly scoped security exception, never a global policy change. Named owners, lifecycle and credentials live in the approved release system. |
| D9 | Catalog completeness | Authenticated scanning must enumerate the complete D17-approved distribution catalog, not blindly the baseline 83 tools. Otherwise launch is blocked; no anonymous catalog fallback. |
| D10 | Versioning and resubmission | Any change to the tool list, names, titles, descriptions, input/output schemas, security schemes, annotations, tool `_meta`, server `instructions`, UI resource URI, or linked UI metadata/CSP requires a compatible server deployment, new draft, fresh Scan Tools, regression run, and resubmission. Skill changes require a new plugin version, reattachment, evaluation, and resubmission. Origin changes require a new plugin. Removals, renames, incompatible schemas, and removed/incompatible UI resources use an additive compatibility window; never break the live published snapshot. Server-only fixes that preserve captured contracts ship independently. |
| D11 | Tool classification and advertised scopes | `readOnlyHint`, `destructiveHint`, and `openWorldHint` are reviewed per tool and declared in `PublicMcpServerConfig`; they are never inferred from names. Each tool advertises the minimal exact OAuth scope derived from its existing read/write/admin/render bucket, mirrored in compatibility metadata, and pinned by contract tests. D4's materialized scope closure makes higher grants satisfy lower-scope descriptors. Recoverable challenges request only the missing minimal scope. |
| D12 | Claude compatibility | `.claude-plugin/`, `.mcp.json` (`mcpServers` wrapper), and `agents/*.md` remain supported and unchanged in shape. |
| D13 | OpenAI MCP config — **validation required** | Use `"apps": "./.app.json"` and a required registered connection. Verify the actual accepted app-ID form and app access separately for every distribution path. Preserve Claude `.mcp.json` source, but test importer auto-discovery and desktop-only classification; omission from the manifest is insufficient proof. A separate generated OpenAI artifact may need to exclude it. `openai.mcp.json` is a conditional direct-client path, not a web/cloud fallback. |
| D14 | UI | ChatGPT renders the existing MCP Apps widgets. Codex must be fully useful from text and `structuredContent`. |
| D15 | Deep links | V1 permits widget navigation only to the exact stable origin `https://deals.anduintransact.com`. Fallback `*.anduin.io`, customer custom domains, non-US environments, and arbitrary `render_table` destinations are not allowlisted and are emitted as inert text by the server. No wildcard or release-time enumeration is used. |
| D16 | Effective deployment configuration | Before each spike or production rollout, inspect the rendered effective DCR allowlist for that environment. Out-of-repo overrides replace the default list wholesale; each target must retain `localhost`, `127.0.0.1`, `claude.ai`, and `claude.com`, and add `chatgpt.com` where the candidate OpenAI flow is enabled. Empty or missing configuration must fail closed rather than disable domain validation. |
| D17 | OpenAI data policy and distribution | Public launch is blocked until tool inputs, outputs, resources, logs and generic document/search/OCR paths are classified and required restrictions are enforced before the plugin/model boundary. Approve the distribution catalog before final scan. New exclusions/redaction and trusted distribution identification are separate implementation work, not supplied by Phase 1 or prompt text. Preserve existing Claude access deliberately. Regulated sensitive data additionally requires necessity, adequate consent and disclosure. A private pilot needs explicit approval and does not waive privacy/security or fulfill D1. |
| D18 | Skill MCP dependencies | Each canonical skill that requires Anduin tools carries `agents/openai.yaml` with an explicit MCP dependency (`type`, stable connection `value`, description, `streamable_http` transport, and production URL). The dependency is reconciled with `.app.json`, validated without credentials, and tested from a clean install; ambient MCP configuration is not assumed. |

## 2. Verified current state

### 2.0 Review update (2026-09-05)

The latest code review fixes pin HTTPS when a trusted edge host is present (ignore `cf-visitor` and other caller
scheme hints), propagate Data Room operation failures, bind token audience before product access checks, and validate
empty fund-report output. Local regression evidence belongs to G1; deployed edge trust and actual ChatGPT/Codex
behavior still belong to G2/G4. `FileFolderInfoTools` deliberately returns one clean failed result for missing,
inaccessible and failed metadata reads, preserving privacy without pretending the operation succeeded.

Current PR branch: `codex/openai-phase-1`, head `93ba809fbe2e9a6133f39489fb0b119d1d491166`, rebased onto upstream
`323f1c0dcb5d329b59d0b0ed8bdd0c7837dcc8e9`. The historical `chatgpt-codex/phase-1` checkout is kept separately to
preserve unrelated dashboard edits. See the plan's G1 evidence block for final checks and PR status.

Official docs currently differ on identifier shape: the builder accepts a `plugin_asdk_app...` URL/creator ID,
whereas workspace import documents the underlying `asdk_app_...` app ID and rejects plugin IDs. Record actual
identifiers and validate both paths in Spike 3. Workspace import can mark MCP-declaring packages desktop-only even
for remote HTTPS; retaining the Claude MCP file in the same artifact is therefore a compatibility risk, not just a
JSON-schema issue. References: [builder](https://developers.openai.com/plugins/build/plugins),
[workspace import](https://learn.chatgpt.com/docs/enterprise/plugin-management).

### 2.1 Live production probe (2026-09-03, unauthenticated)

```
GET https://mcp.anduin.app/.well-known/oauth-protected-resource/mcp   (identical body at the un-suffixed path)
{"resource":"https://mcp.anduin.app/mcp",
 "authorization_servers":["https://id.anduin.app/hydra/env/default"],
 "scopes_supported":["dataroom:admin","dataroom:read","dataroom:write","fundsub:admin","fundsub:read","fundsub:write","mcp:render"],
 "bearer_methods_supported":["header"]}

GET https://id.anduin.app/.well-known/oauth-authorization-server/hydra/env/default
 issuer                                  https://id.anduin.app            <-- mismatch with PRM entry (D5 fixes this)
 authorization_endpoint                  https://id.anduin.app/oauth2/auth
 token_endpoint                          https://id.anduin.app/oauth2/token
 registration_endpoint                   https://id.anduin.app/hydra/env/default/oauth2/register
 code_challenge_methods_supported        [S256]
 grant_types_supported                   [authorization_code, refresh_token]
 token_endpoint_auth_methods_supported   [none, client_secret_post]
 authorization_response_iss_parameter_supported   (absent)

POST https://mcp.anduin.app/mcp  initialize  -> HTTP 401
 www-authenticate: Bearer resource_metadata="https://mcp.anduin.app/.well-known/oauth-protected-resource", scope="..."
 body: {"jsonrpc":"2.0","id":1,"error":{"code":-32001,"message":"Authentication required"}}
```

Consequences:

- The issuer mismatch the plan predicted is live on production, not only in source.
- `initialize` itself requires a token, so any ChatGPT scan must complete OAuth first. There is no anonymous
  catalog to accidentally rely on (supports D7/D9).
- The 401 challenge advertises the un-suffixed PRM path while RFC 9728 clients derive the suffixed one. Both
  resolve to the same document today. Keep both working through Phase 1.
- `mcp-staging.anduin.dev` did not resolve from the office network used for this probe. Minas Tirith
  (`https://minas-tirith.anduin.dev`, PRM resource `https://mcp-minas-tirith.anduin.dev/mcp`) and EU answered.

### 2.2 Environment hosts (from `stargazer/ci/rivendell-v2/src/apps/appconfig/environments/*/config.ts`)

| Environment | `server.baseUrl` | Portal | Hydra advertised host | MCP advertised host |
|---|---|---|---|---|
| gondor-public (prod US) | `https://deals.anduintransact.com` | `https://portal.anduin.app` | `https://id.anduin.app` | `https://mcp.anduin.app` |
| gondor-eu-public | `https://deals.eu.anduin.app` | `https://portal.eu.anduin.app` | `https://id.eu.anduin.app` | `https://mcp.eu.anduin.app` |
| gondor-minas-tirith | `https://minas-tirith.anduin.dev` | `https://portal-minas-tirith.anduin.dev` | derived | derived |
| gondor-internal (staging) | `https://staging.anduin.dev` | `https://portal-staging.anduin.dev` | derived | derived |
| gondor-demo | `https://demo.anduin.dev` | `https://portal-demo.anduin.dev` | `https://id-demo.anduin.dev` | derived |

Environment fallback subdomain (`STARGAZER_ENVIRONMENT_FALLBACK_SUBDOMAIN`): `anduin.io` (US), `eu.anduin.io` (EU). `target.anduin.app` /
`target.eu.anduin.app` is the CNAME target customers point their own domains at; it never appears in a deep link.

### 2.3 Source facts the plan depends on

- `stargazer` Phase 1 review base: `upstream/master` `74a3ca15a92511fb34ae4e562d8baaf6c693dcde`;
  implementation commit `c05643a1368e0a1608ed6824f8d5936fab87a957`
- `mcp-ui-scala`: `337c7082fd5ebdc08089a4bcb86bf992c6601236` (`master`, `v0.4.0`) for the facts below; Phase 0A merged via
  [PR #30](https://github.com/anduintransaction/mcp-ui-scala/pull/30) and released as `v0.5.0` (tag `313147b`).

| Fact | Where |
|---|---|
| Default DCR redirect allowlist is `["localhost","127.0.0.1","claude.ai","claude.com"]`; matching is domain-wide including subdomains. Out-of-repo per-environment overrides may replace the list wholesale, so rendered effective configuration must be inspected (D16). **Phase 1 adds `chatgpt.com` (exact host, `/connector/oauth/{callback_id}` only). An empty list now admits ONLY loopback callbacks — no remote domain — and `HydraDcrServerTestSpec` pins that fail-closed state.** | `platform/stargazerConfig/jvm/resources/reference.conf:1827`, `GondorBackendConfig.scala:1182`, `HydraPublicServer.dcrAllowedDomains`, `HydraPublicServer.validateRedirectUris`, `HydraDcrServerTestSpec` |
| AS metadata `issuer` is `urls.publicUrl` (host only) while PRM advertises `envScopedAuthServerUrl`. **Phase 1 branch: `hydraConfig.envScopedMetadataIssuer` (default off) switches it to the env-scoped URL; `HydraMetadataIssuerTestSpec` pins both modes.** | `HydraPublicServer.authorizationServerMetadata`, `OAuth2McpAuthHttpExtension.protectedResourceMetadata`, `HydraUrls.scala` |
| Consent and resource-server checks share `McpResourceIdentifier`: exact absolute scheme/host/effective port and `/mcp`, no userinfo/query/fragment. Consent admits configured bases and registered custom domains at default HTTPS only. All repeated resources are validated. The gateway overwrites the trusted authority header; MCP pins HTTPS in the edge path and ignores client scheme hints. Backend isolation and actual edge configuration remain deployment checks. Wrong-audience tokens fail before product guards. Hydra 2.2.0 ignores token-endpoint `resource`, but exchange/refresh tokens stay bound to the consented audience in local integration tests. | `McpResourceIdentifier`, `McpResourceAudiencePolicy`, `McpEndpointServer`, `OAuth2McpAuthenticator.bindResource`, `McpOAuthConsentFlowInteg` |
| Hydra refresh checks the client's audience whitelist. DCR seeds configured resources; consent appends validated custom-domain resources with bounded re-read/verify retries. Local real-Hydra integration proves custom-host patch and refresh, while mocked tests cover selected interleavings. Do not infer distributed race freedom; parallel deployed consent/refresh remains an explicit acceptance case. | `HydraPublicServer.registerHydraClient`, `HydraConsentServer.whitelistClientAudience`, `HydraConsentServerTestSpec`, `McpOAuthConsentFlowInteg` |
| `McpToolDefinition` has `name, description, inputSchema, _meta` only — no `title`, `annotations`, `outputSchema`, `securitySchemes`. **Phase 1 branch adds all four as optional fields; `PublicMcpServerConfig.publicToolMetadata` / `describeTool` populate them for public tools only.** | `modules/mcp/mcp/jvm/src/anduin/mcp/protocol/McpTypes.scala`, `PublicMcpServerConfig.scala` |
| Reviewed hint sets are EXACT and pinned by `PublicToolMetadataContractTestSpec`. Destructive (replace or remove existing data): `update_form_fields`, `update_order_tags`, `batch_update_order_tags`, `update_order_custom_data`, `dr_rename_dataroom`, `dr_rename_item`, `dr_archive_dataroom`, `dr_remove_users`, `dr_modify_user_permissions`, `dr_delete_items`. Open-world (minted external capability or a message that reaches an LP/invitee): `get_file_download_url`, `dr_get_file_download_url`, `get_invitation_link`, `invite_fund_managers`, `dr_invite_users`, `draft_comment` (public to the LP by default). | `PublicMcpServerConfig.publicToolMetadata`, `PublicToolMetadataContractTestSpec` |
| `RuntimeTool.isReadOnly` exists but no destructive/open-world flags | `modules/reagent/reagentCore/jvm/src/anduin/reagent/core/TypedTool.scala:121` |
| Public Data Room tools FAIL their effect on failure: validation, scope denials, `GeneralServiceException` and `DataRoomException` become `CleanToolError` (adapter renders `isError=true`, no correlation id); anything else propagates to the adapter's opaque correlation-id error. No "Error …" success strings remain in the five tool files. | `DataRoomAgentTools.failClean` / `surfaceFailure`, `DataRoomToolFailureSurfacingSpec`, `PublicDataRoomToolErrorContractTestSpec` |
| All eight UI tools' REAL output is validated against the schema they advertise: the three render tools through the adapter (`PublicToolMetadataContractTestSpec`), the five `show_*` projections in their source modules, populated and empty-state (`DataRoomShowToolsWidgetSpec`, `LpReviewToolExecutionTestSpec`; shared test-only dependency `JsonSchemaValidatorTestDeps`). | `build/dependency.mill`, the three specs |
| SDK resource meta models `csp` and `domain`; since Phase 0A `ResourcesRead.contents(uri, body, meta, openAi = Some(OpenAiWidgetOptions(redirectDomains, description)))` also emits `openai/widgetCSP` (with `redirect_domains`), `openai/widgetDomain`, and `openai/widgetPrefersBorder`, derived from `_meta.ui` | `mcp-ui-scala/core/src/mcpui/ResourcesRead.scala`, `OpenAiCompat.scala` |
| SDK exports sound draft 2020-12 output schemas per widget and variant (`TableSpec`/`ChartSpec`/`FormSpec` `DisplayOnlyOutputSchema` / `InteractiveOutputSchema`), versioned `$id`, no `$schema`; validator-only rules are listed in each `description` | `mcp-ui-scala/widgets/src/mcpui/widgets/OutputSchema.scala` and the three `*Spec.scala` |
| SDK table renderer accepts any http(s) href and routes clicks through `app.openLink`; no origin allowlist | `mcp-ui-scala/widgets/ts/src/render/cells.ts` (`renderLink`), `widgets/ts/src/table.ts` |
| Stargazer pins `McpUi.version = "0.4.0"` on `upstream/master`; the Phase 1 branch pins the released `0.5.0` | `stargazer/build/versions.mill` |

### 2.4 Official OpenAI constraints used by the decisions

- Codex validates any authorization-response `iss` it receives against authorization-server metadata even when the
  server does not advertise issuer support. This makes the real success/error redirect check in D5 a hard gate:
  [Connect Codex to MCP servers](https://learn.chatgpt.com/docs/extend/mcp).
- A published plugin uses the metadata captured by Scan Tools. Changes to server instructions, tool metadata, UI
  resource URIs, or linked UI metadata require a new draft and scan; an origin change requires a new plugin, and
  breaking contract changes must remain additively compatible with the active version:
  [MCP server review requirements](https://developers.openai.com/plugins/deploy/app-review).
- Tool `securitySchemes` tell the host which scopes to request. This is why D11 advertises the minimal bucket scope and
  D4 makes hierarchical grants client-visible:
  [Authenticate users](https://developers.openai.com/plugins/build/auth).
- Widget navigation uses `_meta["openai/widgetCSP"].redirect_domains`. The reference documents explicit redirect
  domains but no wildcard contract, so D15 intentionally chooses one exact v1 origin:
  [Plugin and MCP Apps reference](https://developers.openai.com/plugins/reference).
- OpenAI's plugin guidelines prohibit collection, solicitation, or processing of Restricted Data and impose
  necessity, consent, disclosure, and minimization requirements on regulated sensitive data. D17 therefore blocks a
  public submission until server-enforced controls are proven:
  [Plugin guidelines](https://developers.openai.com/plugins/app-guidelines).
- A skill that needs an MCP server declares that dependency in `agents/openai.yaml`. D18 makes this a package and
  clean-install contract rather than relying on ambient configuration:
  [Build skills](https://developers.openai.com/plugins/build/skills).

## 3. Production base revision (historical snapshot: 2026-09-03)

These release-train counts and tags are historical context, not current deployment evidence. Phase 1 is reconciled
against `upstream/master` `323f1c0dcb5d329b59d0b0ed8bdd0c7837dcc8e9` fetched on 2026-09-05; record the PR's final base
and head after reconciliation. Refresh train/tag information and record the deployed image before any promotion.

- **Upstream:** `anduintransaction/stargazer` (`upstream` remote). `origin` is the personal fork `cmpham/stargazer`; ignore it for base decisions.
- **Release flow:** `upstream/release` is the train (currently `beta-410.0-candidate-18-ge809db4ae3c`, 2026-08-28). Finals are archived as `archive/beta-N.0` with tag `beta-N.0-final`; newest final is `beta-409.0-final` (`89168c35ebd`, 2026-08-20).
- **Divergence:** `upstream/master` is 620 commits ahead of `upstream/release`. 21 of those touch `modules/mcp`, `modules/reagent`, or `modules/heimdall`, including PL-412 (conformance suite) and PL-413 (derived capabilities, auth error codes). None of them are on the release branch or in `beta-409.0-final`.
- **Deployed tag:** the prod image tag (`stargazerSprintTag`) is set outside the repo. It is exposed as the `version` label on the gondor Deployment (`ci/rivendell-v2/src/apps/gondor/common/index.ts:53`). Record it in the spike log with:

  ```
  kubectl --context <prod-ctx> -n <gondor-namespace> get deploy -l app=gondor -o jsonpath='{.items[*].metadata.labels.version}'
  ```

- **Implications for PR 1:** target `upstream/master`; the change reaches production only when the next train is cut, so the Phase 1 "deploy and verify before PR 2" gate is a release-train dependency, not a merge. Phase 1 golden tests must run against the metadata shape production actually serves (Section 2.1), not against local defaults.
- The plan's remark that the local checkout was "ahead of `upstream/master`" is obsolete: on 2026-09-03 the local `master` is 35 commits behind `upstream/master` and 0 ahead.

## 4. Deep-link origin inventory (spike 4)

### 4.1 What emits links today

| Producer | Link source | Origin control |
|---|---|---|
| `show_funds`, `show_orders` | `LpReviewServiceLive.resolveFundSubDeepLinkBase` + `fundAdminSpaPath` / order doc paths, via `FundSubToolHelpers.openLinkField` | server-resolved |
| `show_fund_report` | no link column today | n/a |
| `show_datarooms` | `DataRoomAgentServiceLive.resolveDataRoomDeepLinkBase` + `dataRoomHomeSpaPath` (`#/detail/home/<id>`) | server-resolved |
| `show_dataroom_insights` | same resolver + `dataRoomInsightsSpaPath` | server-resolved |
| `render_table` (`link` column) | **model-supplied** `{text, href}` or bare URL | none — any https URL |
| `render_chart`, `render_ui` | no links | n/a |

### 4.2 How the server resolves a deep-link host

`EnvironmentService.resolvePath` / `CustomDomainService.resolveOfferingCustomDomainPath`, in priority order:

1. Environment **offering** primary custom domain, else environment **platform** primary custom domain → customer-owned hostnames (unbounded set).
2. Environment **fallback** domain → `<10 random alphanumerics>.<fallbackSubdomain>`, i.e. `*.anduin.io` in prod US, `*.eu.anduin.io` in EU (`EnvironmentService.generateFallbackSubdomain`).
3. No environment, custom domains enabled → the offering's global primary custom domain, else default.
4. Default → `server.baseUrl` + offering prefix: `https://deals.anduintransact.com/fundsub/#/…` and `https://deals.anduintransact.com/dataroom/#/…` in prod US.

### 4.3 Allowlist decision (D15)

| Origin | In v1 allowlist | Notes |
|---|---|---|
| `https://deals.anduintransact.com` | **yes** | The only stable fund-sub and data-room origin admitted in v1. |
| `https://*.anduin.io` | **no** | Environment fallback hostnames are random and wildcard behavior is not part of the documented OpenAI contract. Rows using them render without a clickable link. |
| Customer custom domains | **no** | Unbounded and customer-controlled. Rows whose resolved base is not allowlisted are emitted without an `open` cell and render as the existing muted em-dash. |
| Model-supplied `render_table` links | filtered | the server downgrades any `link` cell whose origin is outside the allowlist to plain text before returning the spec |
| `https://portal.anduin.app`, `https://id.anduin.app`, `https://mcp.anduin.app` | **no** | Never emitted by a widget. |

EU, staging, Minas, fallback, and customer-owned hosts are not in the production allowlist. The production OpenAI
resource metadata advertises exactly `https://deals.anduintransact.com`; it is not derived from a changing environment
inventory. A future regional or custom-domain expansion needs its own reviewed exact-origin set and a new metadata
snapshot.

The server-side filter is mandatory even when the host enforces `_meta["openai/widgetCSP"].redirect_domains`, so
widgets behave identically on every host. Parse and compare normalized URL origins; do not use string-prefix matching.

Implementation (Phase 1 branch): `anduin.mcp.ui.WidgetOriginPolicy`, configured from `mcpConfig.widgetRedirectOrigins`
(fail-closed default `[]`; production US sets `STARGAZER_SERVICES_MCP_WIDGET_REDIRECT_ORIGIN=https://deals.anduintransact.com`
in `ci/rivendell-v2/.../gondor-public/config.ts`, never derived from the base URL) and `mcpConfig.widgetUiDomain`
(when set, it must be exactly one canonical https origin — `WidgetOriginPolicy.validateUiDomain` — or startup fails).
`McpServer` applies `filterStructuredContent` to every table `structuredContent` (a link cell is kept only when every
`href`/`url` value it carries is allowlisted; otherwise it becomes `{"text": <label or href>, "href": ""}`, which the
SDK renders inert) and `UiResourceRegistry` emits the same list as `openai/widgetCSP.redirect_domains`. The internal
PAT server stays unrestricted.

## 5. Resubmission rules (D10, operational form)

Before every OpenAI submission:

1. Diff the reviewed server contract against the last captured Scan Tools snapshot: tool list, names, titles,
   descriptions, `inputSchema`, `outputSchema`, `securitySchemes`, annotations, tool `_meta`, server `instructions`, UI
   resource URIs, and linked UI metadata including CSP.
2. Any diff requires a backward-compatible deployment, new draft, Scan Tools with the production D8 review account,
   snapshot comparison, Phase 4 golden prompts and authorization matrix, and resubmission.
3. Any change under `plugins/anduin/skills/` → bump plugin version, reattach skills, run skill evaluations, resubmit.
4. Record the production `version` label (Section 3), plugin tag, and listing version together in the release evidence.
5. A change to MCP scheme, hostname, or port requires a new plugin rather than a new version. An endpoint-path-only
   change uses the normal new-version flow.
6. Keep changes additive and honor every active published tool schema and UI URI. A replacement's approval or an
   elapsed rollback window is not permission to remove an old contract: current review rules do not support breaking
   updates. Require a supported migration and evidence that no supported consumer needs the old contract.
7. Compatible HTML at an unchanged resource URI may stay cached for up to an hour. A changed content-hash URI is a
   captured-contract change requiring a rescan. Test both existing and new clients during rollout.

## 6. Spike runbook and log

Spikes 1–3 need a person with ChatGPT developer mode, an OpenAI organization, and Codex on each surface. D17 also
requires a named product, security/privacy, and legal owner before public submission. Fill in
the **Result** lines and keep secrets out of this file. Use one selected public non-production environment and its
own synthetic account; the production D8 account is reserved for the final registration, scan, and review.

The executable order is:

1. Create the disposable developer-mode registration and capture its app identifier and ChatGPT callback URI.
2. Finish G1: verify the review fixes, preserve unrelated work, reconcile upstream and publish the focused PR.
   Require reviewer/CI approval before merge. The issuer candidate stays default-off; local real-Hydra tests do not
   replace deployed redirects, edge reachability or host tests. Concurrent audience updates need real parallel
   consent/refresh validation before claiming distributed race safety.
3. Deploy the narrow candidate OAuth slice to the selected public non-production environment: the full effective DCR
   allowlist from D16 plus the provisional D5 metadata issuer. Do not deploy the candidate issuer to production.
4. Complete Spike 2 by recording actual successful and error authorization redirects. If any returned `iss` differs
   from metadata, reject D5 and redesign before continuing.
5. Only after D5 passes, run the authenticated Scan Tools and Codex package/surface spikes.
6. Complete and approve D17 before preparing any public submission; a private evaluation does not count as public
   policy approval.
7. Phase 3 repeats the successful flow against production with the production D8 review account after the complete
   backward-compatible Phase 1 server change is deployed.

### Spike 1 — Scan Tools behavior

After Spike 2 passes, complete OAuth to the spike endpoint with the non-production full-scope synthetic account, run
Scan Tools, and export the captured metadata.

Record:
- Did the scanner require OAuth before enumerating? (expected: yes; Section 2.1 shows `initialize` is 401 unauthenticated)
- Which scopes were granted and whether the snapshot equals the approved test catalog; final production scan must
  equal D17's distribution inventory, with exclusions enforced at execution as well as discovery.
- Whether the scanner captured the exact tool titles, descriptions, input/output schemas, security schemes,
  annotations, tool `_meta`, server `instructions`, and linked UI metadata/CSP.
- Result: _pending_

### Spike 2 — App identifier, redirect, and issuer preflight

Record (in the release system, referenced here by name only):
- The generated URL/creator identifier and underlying registered app ID locations, with observed mapping format.
- The callback-specific redirect URI `https://chatgpt.com/connector/oauth/{callback_id}` location. Before the candidate
  allowlist deployment, DCR is expected to reject it; after deployment, it must accept it.
- The candidate metadata issuer and the actual `iss` query parameter, or confirmed absence of `iss`, on both a
  successful authorization redirect and a user-denied/error redirect.
- The same returned-`iss` check for a direct Codex loopback flow. Record the exact callback URI shown by Codex,
  including its server-specific callback ID and variable loopback port behavior.
- The rendered effective DCR allowlist for the spike environment.
- D5 verdict: _pending — accept only if every returned `iss` is absent or exactly matches metadata_
- Result: _pending_

### Spike 3 — Codex surfaces with `.app.json` only

After Spike 2 passes, install a package that references only `.app.json` (no OpenAI MCP file). Include the required
per-skill `agents/openai.yaml` MCP dependency declarations. Per surface record the
connection path, actual registration method, exact callback, PKCE, reconnect after token expiry, tool call, catalog
behavior, whether enabling a skill resolved its dependency from a clean install, and whether a direct
`openai.mcp.json` was required.

Disable ambient Anduin connections and avoid reusing CLI/App/IDE credentials for clean-install evidence. Record
whether the existing Claude `.mcp.json` was discovered, whether the import became desktop-only, required-app access,
duplicate connections and the accepted app-ID shape. Test direct MCP separately from registered-app wiring. Cloud
support and callbacks are unproven; obtain an explicit scope decision if no supported path can be demonstrated.
Record per-host timeouts/cancellation and ambiguous-write recovery; a direct client's 60-second default may be
shorter than a server operation. No automatic retry of potentially completed writes.

| Surface | Connect | Registration + callback | PKCE | Reconnect | Skill dependency + tool/catalog result | Needs `openai.mcp.json`? |
|---|---|---|---|---|---|---|
| Codex App | _pending_ | | | | | |
| Codex CLI | _pending_ | | | | | |
| Codex IDE extension | _pending_ | | | | | |
| Codex cloud | _pending_ | | | | | |

### Spike 4 — Deep-link origins

Origin decision recorded in Section 4, not live UI acceptance. Set the dedicated UI origin before final scan and
verify fetch/CSP/navigation in ChatGPT. Non-production links stay inert unless a separate exact test origin is
approved; never point staging records at production. Final navigation tests use the production synthetic account.

### Production base

- `upstream/master` at time of record: `0422db31589` (2026-09-03).
- Newest final tag: `beta-409.0-final`; release head: `beta-410.0-candidate-18-ge809db4ae3c`.
- Deployed prod `version` label: _pending (needs cluster access, Section 3)_.

## 7. Exit criteria status

| Criterion | Status |
|---|---|
| D5 issuer design proven against successful and error redirects | **Open** — spike 2; production promotion prohibited until it passes |
| Full-catalog scan strategy proven with the non-production synthetic account, or launch blocked | **Open** — spike 1 |
| Real app identifier and callback redirects recorded without credentials | **Open** — spike 2 |
| `.app.json` behavior and `openai.mcp.json` need known per Codex surface | **Open** — spike 3 |
| Effective DCR allowlist verified in the spike environment and production | **Open** — spike environment in spike 2; production before rollout |
| Empty or missing DCR allowlist fails closed | **Done in code** — empty admits loopback only (`HydraDcrServerTestSpec`) |
| Token audience bound to the exact trusted MCP origin and path, with no userinfo/query/fragment, in consent and resource-server validation | **Done in code** — `McpResourceIdentifier` + `McpResourceAudiencePolicy` + `bindResource`; refresh proven by `McpOAuthConsentFlowInteg` |
| Tag replacement/clear tools classified as destructive | **Done in code** — exact destructive/open-world sets pinned (`PublicToolMetadataContractTestSpec`) |
| D17 per-tool data classification, server controls, disclosures, and distribution verdict recorded | **Open** — public listing remains blocked |
| D18 per-skill `agents/openai.yaml` dependencies resolve from a clean install | **Open** — Phase 2/3 package work |
| Upstream and release base revisions documented | **Done** (Section 3) |
| Exact deployed production revision recorded | **Open** — needs cluster read before production rollout |
| Deep-link origin allowlist documented | **Done** — one exact origin, no wildcard dependency (Section 4) |
| Resubmission rules documented | **Done** (Section 5) |
| Phase 0A SDK contracts (output schemas, `redirect_domains`, UI domain alias) | **Done** — PR #30 merged, released as `v0.5.0` |
| Phase 1 server changes (1.1, 1.3, 1.4, 1.5, D4 closure; 1.2 as a config-gated candidate) | **PR #55982 open; local checks passed, CI/review pending** — `codex/openai-phase-1`, SDK `0.5.0`; no deployed-host claim. See the plan's G1 evidence. |

Dependencies to start now, in parallel with G1/G2 and canonical skill work (PR 0 is already released): OpenAI
organization and business verification, legal/support URLs, the non-production spike account, the production D8
demo organization and review account, and a disposable developer-mode registration for spikes 1–2.
