# OpenAI Compatibility Contract (Phase 0 record)

**Status:** Decisions recorded; Phase 0A released (`mcp-ui-scala` `v0.5.0`); Phase 1 implemented on `stargazer` `chatgpt-codex/phase-1` (uncommitted); Phase 0 remains open pending D5 validation and spikes 1–3
**Recorded:** 2026-09-03, updated 2026-09-04
**Plan:** `docs/plans/chatgpt-codex-support.md` (Phase 0)
**Review:** `docs/reviews/chatgpt-codex-support-review.md`
**Repositories:** `mcp-ui-scala`, `stargazer`, `anduin-plugin`

This note is the "architecture note" Phase 0 asks for. Section 1 locks the decisions. Sections 2–4 record the
facts they rest on, each checked against the live production endpoint or source on the date above. Section 6 is
the runbook and log for the spikes that need real ChatGPT/Codex accounts.

## 1. Locked decisions

| # | Decision | Value |
|---|---|---|
| D1 | Packaging | One OpenAI plugin package and one universal listing serve ChatGPT and Codex. |
| D2 | Target environment | First release targets production US only: `https://mcp.anduin.app/mcp`. EU (`mcp.eu.anduin.app`) is out of scope until US is stable. |
| D3 | Client registration | Dynamic Client Registration (RFC 7591) with PKCE S256. Client ID Metadata Documents deferred. |
| D4 | OAuth scopes | Scope names and authority are unchanged: `fundsub:read/write/admin`, `dataroom:read/write/admin` (hierarchical), `mcp:render` (flat). New grants materialize the hierarchy in the token's `scope` claim (`admin` also carries `write` and `read`; `write` also carries `read`) so client-visible scopes match server authorization. Existing tokens remain valid during transition because the resource server continues expanding scopes. Destructive dataroom tools stay gated on `dataroom:admin`. |
| D5 | Issuer — **provisional** | Candidate design: authorization-server metadata `issuer` becomes that environment's scoped identifier (production example: `https://id.anduin.app/hydra/env/default`) so it equals the PRM `authorization_servers` entry; RFC 9207 is not advertised and token validation stays strict. Do not promote this change beyond the public non-production spike environment until successful and error authorization redirects prove that any returned `iss` is either absent or exactly environment-scoped. Codex rejects every mismatched returned `iss`, even when issuer support is not advertised. If Hydra returns the bare-host issuer, stop and redesign; a metadata-only production change is forbidden. |
| D6 | ChatGPT redirect | Callback-specific `https://chatgpt.com/connector/oauth/{callback_id}` (the stable `connector_platform_oauth_redirect` needs RFC 9207, excluded by D5). DCR allowlist gains the `chatgpt.com` domain only, and the validator pins it to that exact host and path (Phase 1). |
| D7 | Discovery model | Scan Tools output is the authoritative catalog for the published OpenAI plugin in both ChatGPT and Codex. Live `tools/list` stays authoritative for Claude and direct Codex MCP connections. Execution authorization is enforced server-side for every host. |
| D8 | Test and review accounts | Use two synthetic full-scope accounts: one in the selected public non-production spike environment, and one in a dedicated demo organization **on production US** for the final scan and OpenAI review. Both contain no customer data and hold every scope and product role needed to enumerate and exercise review tools. The production review account has no MFA or external verification step and remains valid for review. Owners and credentials live in the approved release system, never in this repo. |
| D9 | Full-catalog fallback | If authenticated scanning cannot enumerate the full catalog, the launch is blocked. No anonymous catalog is exposed. |
| D10 | Versioning and resubmission | Any change to the tool list, names, titles, descriptions, input/output schemas, security schemes, annotations, tool `_meta`, server `instructions`, UI resource URI, or linked UI metadata/CSP requires a compatible server deployment, new draft, fresh Scan Tools, regression run, and resubmission. Skill changes require a new plugin version, reattachment, evaluation, and resubmission. Origin changes require a new plugin. Removals, renames, incompatible schemas, and removed/incompatible UI resources use an additive compatibility window; never break the live published snapshot. Server-only fixes that preserve captured contracts ship independently. |
| D11 | Tool classification and advertised scopes | `readOnlyHint`, `destructiveHint`, and `openWorldHint` are reviewed per tool and declared in `PublicMcpServerConfig`; they are never inferred from names. Each tool advertises the minimal exact OAuth scope derived from its existing read/write/admin/render bucket, mirrored in compatibility metadata, and pinned by contract tests. D4's materialized scope closure makes higher grants satisfy lower-scope descriptors. Recoverable challenges request only the missing minimal scope. |
| D12 | Claude compatibility | `.claude-plugin/`, `.mcp.json` (`mcpServers` wrapper), and `agents/*.md` remain supported and unchanged in shape. |
| D13 | OpenAI MCP config | The OpenAI manifest references `.app.json`. `.mcp.json` is never referenced as OpenAI configuration. If a Codex surface needs a direct server file, add `openai.mcp.json` in OpenAI's shape and reference only that. |
| D14 | UI | ChatGPT renders the existing MCP Apps widgets. Codex must be fully useful from text and `structuredContent`. |
| D15 | Deep links | V1 permits widget navigation only to the exact stable origin `https://deals.anduintransact.com`. Fallback `*.anduin.io`, customer custom domains, non-US environments, and arbitrary `render_table` destinations are not allowlisted and are emitted as inert text by the server. No wildcard or release-time enumeration is used. |
| D16 | Effective deployment configuration | Before each spike or production rollout, inspect the rendered effective DCR allowlist for that environment. Out-of-repo overrides replace the default list wholesale; each target must retain `localhost`, `127.0.0.1`, `claude.ai`, and `claude.com`, and add `chatgpt.com` where the candidate OpenAI flow is enabled. |

## 2. Verified current state

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

- `stargazer` `upstream/master`: `0422db31589604278921a05d7272db9ee3a5eb91`
- `mcp-ui-scala`: `337c7082fd5ebdc08089a4bcb86bf992c6601236` (`master`, `v0.4.0`) for the facts below; Phase 0A merged via
  [PR #30](https://github.com/anduintransaction/mcp-ui-scala/pull/30) and released as `v0.5.0` (tag `313147b`).

| Fact | Where |
|---|---|
| Default DCR redirect allowlist is `["localhost","127.0.0.1","claude.ai","claude.com"]`; matching is domain-wide including subdomains. Out-of-repo per-environment overrides may replace the list wholesale, so rendered effective configuration must be inspected (D16). **Phase 1 branch adds `chatgpt.com`.** | `platform/stargazerConfig/jvm/resources/reference.conf:1827`, `GondorBackendConfig.scala:1182`, `HydraPublicServer.validateRedirectUris` |
| AS metadata `issuer` is `urls.publicUrl` (host only) while PRM advertises `envScopedAuthServerUrl`. **Phase 1 branch: `hydraConfig.envScopedMetadataIssuer` (default off) switches it to the env-scoped URL; `HydraMetadataIssuerTestSpec` pins both modes.** | `HydraPublicServer.authorizationServerMetadata`, `OAuth2McpAuthHttpExtension.protectedResourceMetadata`, `HydraUrls.scala` |
| `McpToolDefinition` has `name, description, inputSchema, _meta` only — no `title`, `annotations`, `outputSchema`, `securitySchemes`. **Phase 1 branch adds all four as optional fields; `PublicMcpServerConfig.publicToolMetadata` / `describeTool` populate them for public tools only.** | `modules/mcp/mcp/jvm/src/anduin/mcp/protocol/McpTypes.scala`, `PublicMcpServerConfig.scala` |
| `RuntimeTool.isReadOnly` exists but no destructive/open-world flags | `modules/reagent/reagentCore/jvm/src/anduin/reagent/core/TypedTool.scala:121` |
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

## 3. Production base revision

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
in `ci/rivendell-v2/.../gondor-public/config.ts`, never derived from the base URL) and `mcpConfig.widgetUiDomain`.
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
6. Never deploy a breaking removal or rename ahead of approval. Add the replacement while continuing to honor the
   published tool schemas and UI resource URIs, publish the approved snapshot, then retain old contracts until the
   replacement is approved and active and the rollback window closes.

## 6. Spike runbook and log

Spikes 1–3 need a person with ChatGPT developer mode, an OpenAI organization, and Codex on each surface. Fill in
the **Result** lines and keep secrets out of this file. Use one selected public non-production environment and its
own synthetic account; the production D8 account is reserved for the final registration, scan, and review.

The executable order is:

1. Create the disposable developer-mode registration and capture its app identifier and ChatGPT callback URI.
2. Deploy the narrow candidate OAuth slice to the selected public non-production environment: the full effective DCR
   allowlist from D16 plus the provisional D5 metadata issuer. Do not deploy the candidate issuer to production.
3. Complete Spike 2 by recording actual successful and error authorization redirects. If any returned `iss` differs
   from metadata, reject D5 and redesign before continuing.
4. Only after D5 passes, run the authenticated Scan Tools and Codex package/surface spikes.
5. Phase 3 repeats the successful flow against production with the production D8 review account after the complete
   backward-compatible Phase 1 server change is deployed.

### Spike 1 — Scan Tools behavior

After Spike 2 passes, complete OAuth to the spike endpoint with the non-production full-scope synthetic account, run
Scan Tools, and export the captured metadata.

Record:
- Did the scanner require OAuth before enumerating? (expected: yes; Section 2.1 shows `initialize` is 401 unauthenticated)
- Which scopes were granted at consent, and whether the captured catalog equals the full public allowlist.
- Whether the scanner captured the exact tool titles, descriptions, input/output schemas, security schemes,
  annotations, tool `_meta`, server `instructions`, and linked UI metadata/CSP.
- Result: _pending_

### Spike 2 — App identifier, redirect, and issuer preflight

Record (in the release system, referenced here by name only):
- The generated `plugin_asdk_app...` identifier location.
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

After Spike 2 passes, install a package that references only `.app.json` (no OpenAI MCP file). Per surface record the
connection path, actual registration method, exact callback, PKCE, reconnect after token expiry, tool call, catalog
behavior, and whether a direct `openai.mcp.json` was required.

| Surface | Connect | Registration + callback | PKCE | Reconnect | Tool/catalog result | Needs `openai.mcp.json`? |
|---|---|---|---|---|---|---|
| Codex App | _pending_ | | | | | |
| Codex CLI | _pending_ | | | | | |
| Codex IDE extension | _pending_ | | | | | |
| Codex cloud | _pending_ | | | | | |

### Spike 4 — Deep-link origins

Done in Section 4. V1 uses one exact origin and has no wildcard-format dependency.

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
| Upstream and release base revisions documented | **Done** (Section 3) |
| Exact deployed production revision recorded | **Open** — needs cluster read before production rollout |
| Deep-link origin allowlist documented | **Done** — one exact origin, no wildcard dependency (Section 4) |
| Resubmission rules documented | **Done** (Section 5) |
| Phase 0A SDK contracts (output schemas, `redirect_domains`, UI domain alias) | **Done** — PR #30 merged, released as `v0.5.0` |
| Phase 1 server changes (1.1, 1.3, 1.4, 1.5, D4 closure; 1.2 as a config-gated candidate) | **Done in code** — `stargazer` branch `chatgpt-codex/phase-1`, uncommitted, pinned to SDK `0.5.0`; needs review, merge, release-train deployment |

Dependencies to start now, in parallel with PR 0 (mcp-ui-scala), so they do not gate Phase 3: OpenAI
organization and business verification, legal/support URLs, the non-production spike account, the production D8
demo organization and review account, and a disposable developer-mode registration for spikes 1–2.
