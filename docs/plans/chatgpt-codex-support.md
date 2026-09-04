# ChatGPT and Codex Support Plan

**Status:** In progress — Phase 0A delivered (`mcp-ui-scala` PR #30); Phase 1 implemented on `stargazer` branch `chatgpt-codex/phase-1` (uncommitted, see status notes per section)
**Last updated:** 2026-09-04
**Repositories:** `mcp-ui-scala`, `stargazer`, `anduin-plugin`
**Primary production endpoint:** `https://mcp.anduin.app/mcp`

## Summary

Support ChatGPT and Codex through one OpenAI plugin package backed by the existing Anduin MCP server and the existing GP Assistant and Data Room skills. Preserve Claude Code and Cowork support by retaining the Claude manifests, MCP configuration, and thin Claude-specific agent adapters.

ChatGPT and Codex should not receive separate implementations. OpenAI plugins can package skills, MCP servers, and optional UI once for use in both products. The portable skills become the source of truth; host-specific configuration remains at the edges.

The work is split into three code tracks followed by deployment and publication:

1. Export exact structured-output schemas and OpenAI widget CSP compatibility from `mcp-ui-scala`.
2. Make the public MCP server compliant with OpenAI OAuth, tool metadata, snapshot discovery, and review requirements in `stargazer`.
3. Canonicalize the skills and add OpenAI packaging in `anduin-plugin` without removing Claude support.
4. Deploy the server, run cross-host acceptance tests, and submit one universal OpenAI listing.

## Goals

- Make both Anduin domains available in ChatGPT and Codex:
  - Fund subscription review and management.
  - Virtual data room discovery, management, documents, and analytics.
- Reuse the production US MCP server and its existing OAuth scopes.
- Preserve the existing permission model, confirmation rules, live scope-filtered discovery for direct clients, and execution-time authorization for every host.
- Render existing MCP Apps widgets where supported and preserve text fallbacks everywhere.
- Keep Claude Code and Cowork behavior working throughout the migration.
- Maintain one canonical copy of domain workflows, tool-chaining rules, and safety guidance.
- Publish one OpenAI listing that serves both ChatGPT and Codex.

## Non-goals

- Replacing or deprecating the Claude plugin.
- Forking the skills or MCP server by AI host.
- Changing Anduin product authorization or weakening existing OAuth scope gates.
- Adding new business capabilities or public MCP tools as part of the compatibility work.
- Supporting the EU MCP endpoint in the first public OpenAI submission. Start with production US and expand only after the US integration is stable and any regional submission constraints are resolved.
- Replacing Dynamic Client Registration with Client ID Metadata Documents in the first release. DCR remains the MVP registration flow.

## Current State

The plugin is currently Claude-specific at its packaging and documentation boundaries:

- `.claude-plugin/marketplace.json` defines the repository marketplace.
- `plugins/anduin/.claude-plugin/plugin.json` defines the Claude plugin.
- `plugins/anduin/.mcp.json` points at the production US MCP endpoint.
- `plugins/anduin/agents/*.md` contain Claude-specific `model`, `tools`, and activation configuration.
- `plugins/anduin/skills/*/SKILL.md` already contain most of the portable domain knowledge needed by ChatGPT and Codex.
- `README.md` documents only Claude Code and Cowork installation and troubleshooting.

The server already provides most of the required foundation:

- Public HTTPS Streamable HTTP MCP endpoint.
- OAuth protected-resource and authorization-server discovery.
- Dynamic Client Registration and PKCE S256.
- Resource audience binding and OAuth scope-based tool filtering.
- MCP Apps resources for chart, table, and form widgets.
- Structured widget output plus a text fallback.

The remaining blocking SDK and server gaps are described in Phase 0A and Phase 1.

## Target Architecture

```text
anduin-plugin/
├── .agents/
│   └── plugins/
│       └── marketplace.json        # Explicit OpenAI/Codex repository marketplace
├── .claude-plugin/
│   └── marketplace.json            # Existing Claude marketplace
├── docs/
│   └── plans/
│       └── chatgpt-codex-support.md
└── plugins/
    └── anduin/
        ├── .codex-plugin/
        │   └── plugin.json          # OpenAI plugin manifest
        ├── .claude-plugin/
        │   └── plugin.json          # Existing Claude manifest
        ├── .app.json                # OpenAI app registration mapping
        ├── .mcp.json                # Claude-only; never parsed as OpenAI configuration
        ├── openai.mcp.json          # Conditional fallback if the app mapping is insufficient
        ├── agents/                  # Thin Claude-specific adapters
        ├── skills/                  # Canonical cross-host behavior
        └── assets/                  # Shared branding and listing assets
```

The shared runtime architecture is:

```text
ChatGPT ─┐
Codex ───┼─> canonical Anduin skills ─> https://mcp.anduin.app/mcp ─> Anduin APIs
Claude ──┘             │
                      └─> MCP Apps widgets when supported, text fallback otherwise
```

## Phase 0: Record and Validate the Compatibility Contract

Before implementation, record the following decisions in the first PR description or an architecture note and validate the uncertain host behavior with a disposable registration:

- One OpenAI package and listing serves ChatGPT and Codex.
- The first release targets production US only.
- DCR with PKCE S256 remains the OAuth client registration flow.
- Existing OAuth scope names and authority remain unchanged. New token grants materialize implied lower scopes so
  client-visible scopes match the existing hierarchy, while the resource server continues expansion for old tokens:
  - `fundsub:read`, `fundsub:write`, `fundsub:admin`.
  - `dataroom:read`, `dataroom:write`, `dataroom:admin`.
  - `mcp:render`.
- For the published OpenAI plugin in both ChatGPT and Codex, the metadata captured by **Scan Tools** is the
  authoritative discovery snapshot. The live `tools/list` response remains authoritative for Claude and direct Codex
  MCP connections, and remains a server contract, but it does not dynamically update the published plugin snapshot.
- Use two dedicated synthetic full-scope accounts: one in the selected public non-production spike environment and a
  separate production-US demo-organization account for the final scan and OpenAI review. Neither contains customer
  data. The production review account avoids MFA or manual-verification dependencies and remains valid for review.
- If authenticated scanning cannot enumerate the full catalog, stop the launch and choose explicitly between a tightly isolated catalog-only discovery mechanism and a package/review redesign. Do not expose an anonymous full catalog by default.
- Changes to tools, schemas, security schemes, annotations, tool `_meta`, server `instructions`, UI resource URIs, or
  linked UI metadata/CSP require a compatible deployment, new draft, fresh **Scan Tools**, regression testing, and
  resubmission. Origin changes require a new plugin. Breaking removals, renames, schemas, and UI resources require an
  additive compatibility window. Server-only fixes that preserve captured contracts may ship independently.
- Skills are also attached as submission artifacts. Skill changes require a new plugin version/draft, reattachment, evaluation, and resubmission.
- Read-only, destructive, and open-world classifications are reviewed per tool rather than inferred from naming
  conventions. Each tool advertises its minimal exact scope from the existing allowlist bucket; materialized scope
  closure lets higher grants satisfy lower-scope tools, and recovery challenges request only the missing minimal scope.
- Existing Claude manifests and agents remain supported.
- The OpenAI manifest references `.app.json`. The existing `.mcp.json` retains its Claude `mcpServers` wrapper and must not be referenced as OpenAI MCP configuration.
- If `.app.json` does not provide the required remote OAuth behavior in a Codex surface, add a separately named `openai.mcp.json` in an OpenAI-supported direct-server shape and reference only that file from the OpenAI manifest. Never make one `.mcp.json` serve incompatible schemas.
- ChatGPT renders MCP Apps UI. Codex workflows must remain fully useful from text and structured tool results without relying on widget rendering.
- V1 widget links are clickable only for the exact origin `https://deals.anduintransact.com`; fallback, customer,
  non-US, and arbitrary model-supplied destinations render as inert text. No wildcard CSP is used.
- Inspect the rendered effective DCR allowlist in every spike and production environment because out-of-repo overrides
  replace the defaults wholesale.

Run the human-dependent work in this order; the detailed log is
`docs/architecture/openai-compatibility-contract.md`:

1. Create a disposable developer-mode registration and capture the generated `plugin_asdk_app...` identifier and
   callback-specific redirect URI. Do not commit placeholders.
2. Deploy the narrow candidate callback/issuer slice to the selected public non-production environment. Capture
   successful and error authorization redirects from ChatGPT and direct Codex. If any returned `iss` differs from the
   candidate metadata issuer, reject the candidate design; do not promote a metadata-only issuer change to production.
3. After the issuer design passes, complete OAuth with the non-production full-scope account, run **Scan Tools**, and
   verify the complete stored contract.
4. Install a package that references only `.app.json` and verify connection, registration method, exact callback,
   PKCE, reconnect, tool/catalog behavior, and fallback requirements separately in Codex App, CLI, IDE extension, and
   cloud.

Implementation targets `stargazer` `upstream/master`; the release train and deployed production revision are separate
evidence. Record the exact deployed `version` label before production rollout rather than inferring it from a local or
remote branch.

### Phase 0 exit criteria

- The candidate issuer is proven against successful and error redirects, including Codex's rule for every returned
  `iss`, or the design is rejected before production.
- The full-catalog scan strategy is proven with the non-production synthetic account or the launch is explicitly blocked.
- The real app identifier and callback-specific redirects are recorded in the release system without storing credentials.
- `.app.json` behavior and any `openai.mcp.json` fallback requirement are known for every Codex surface.
- The upstream/release bases, exact v1 deep-link origin, effective DCR configuration, and resubmission rules are documented.
- The exact deployed production revision is recorded before production rollout.

## Phase 0A: Export MCP UI Contracts

**Repository:** `mcp-ui-scala`
**Dependency:** Merge and release before the `stargazer` compatibility PR.
**Status:** Done — merged via [PR #30](https://github.com/anduintransaction/mcp-ui-scala/pull/30) and released as
`v0.5.0` (`mcp-ui-core_3`/`mcp-ui-widgets_3` `0.5.0` on Anduin Artifactory).

Delivered API (see the SDK README sections "ChatGPT compatibility metadata" and "Output schemas for
`structuredContent`"):

- `TableSpec.DisplayOnlyOutputSchema` / `InteractiveOutputSchema`, `ChartSpec.*OutputSchema`, `FormSpec.*OutputSchema`
  (`McpJson`), versioned `$id` `https://mcpui.dev/schemas/widgets/v1/<widget>.<variant>.json`, no `$schema`.
- `ResourcesRead.contents(uri, body, meta?, openAi?: OpenAiWidgetOptions)`. `OpenAiWidgetOptions(redirectDomains,
  description)` carries the only ChatGPT-only inputs; `openai/widgetCSP`, `openai/widgetDomain`, and
  `openai/widgetPrefersBorder` are derived from the standard `_meta.ui` so the two surfaces cannot disagree.
- `Constants.OpenAiMetaKey` and `OpenAiWidgetMeta.isConsistentWith` for contract tests.

The eight public UI-producing tools return `structuredContent`, but the SDK currently keeps the canonical `TableSpec`, `ChartSpec`, and `FormSpec` validation rules in Scala code without exporting matching JSON Schemas. OpenAI requires a declared `outputSchema` whenever a tool returns `structuredContent`.

Add SDK-owned, versioned JSON Schema constants next to the corresponding validators:

- `TableSpec` schema for `render_table`, `show_funds`, `show_orders`, `show_fund_report`, `show_datarooms`, and `show_dataroom_insights`.
- `ChartSpec` schema for `render_chart`.
- `FormSpec` schema for `render_ui`.

The schemas must be sound for every structured result emitted on success (required fields, unions, nullability,
`additionalProperties`) and must reject every validator-rejected payload that JSON Schema can express. Constraints JSON
Schema cannot express (unique column ids and field aliases, byte and character size caps, the `editable_columns`
cross-check, the form-wide field total) stay validator-only and are listed in each schema's `description`. Tests must
validate representative success payloads, reject expressible validator rejections, and pin the validator-only gaps.

Extend the SDK CSP model and encoder to support the OpenAI compatibility key `_meta["openai/widgetCSP"].redirect_domains` in addition to the standard MCP Apps CSP fields. Also provide the resource metadata needed for a dedicated submitted-UI origin, including `_meta.ui.domain` and any required compatibility alias.

### Phase 0A exit criteria

- Sound Table, Chart, and Form output schemas are exported and covered by conformance tests. **Done** (PR #30:
  `OutputSchemaSpec`, e2e Layer 3b against Ajv and the real MCP SDK client).
- Widget CSP can encode the v1 OpenAI redirect allowlist containing only
  `https://deals.anduintransact.com`. **Done** — the SDK stays provider-neutral and is tested with `example.com`
  origins; the production origin is asserted in `stargazer` (Phase 1.5).
- A released SDK version is available for `stargazer` to consume. **Done** — `v0.5.0` tagged and published;
  `stargazer` pins `McpUi.version = "0.5.0"`.

## Phase 1: Make the MCP Server OpenAI-ready

**Repository:** `stargazer`
**Dependency:** Must be deployed before the OpenAI package is published.
**Status (2026-09-04):** 1.1, 1.3, 1.4, 1.5 and the D4 scope closure are implemented with unit and contract tests on
branch `chatgpt-codex/phase-1` (from `upstream/master` `74a3ca15a92`). 1.2 ships as a config-gated candidate
(`hydraConfig.envScopedMetadataIssuer`, default off) so the spike environment can run it before production. The SDK
pin is the released `McpUi.version = "0.5.0"`.

### 1.1 Accept ChatGPT OAuth callbacks

The DCR redirect-domain allowlist currently contains only localhost and Anthropic-controlled domains:

- `platform/stargazerConfig/shared/src/com/anduin/stargazer/service/GondorBackendConfig.scala`
- `platform/stargazerConfig/jvm/resources/reference.conf`

Allow the documented `chatgpt.com` callback domain in the DCR policy and keep environment overrides in sync. This domain-level change can be implemented before the final app registration. Out-of-repository environment overrides replace the default list wholesale, so capture and review the rendered effective allowlist in the selected public non-production environment and in production; both must retain localhost and Anthropic callbacks while adding `chatgpt.com`. For the MVP, register and test only the callback-specific form emitted by app management: `https://chatgpt.com/connector/oauth/{callback_id}`.

Do not enable the stable `https://chatgpt.com/connector_platform_oauth_redirect` callback in the MVP. It requires RFC 9207 issuer identification on every successful and error authorization response, which the current Hydra flow does not provide consistently.

**Status:** done. `chatgpt.com` is in the in-repo default (`GondorBackendConfig.HydraConfig.dcrAllowedRedirectDomains`
and `reference.conf`), but unlike the Anthropic domains it is matched exact-host and only on
`/connector/oauth/{callback_id}`, so the stable callback, other paths and subdomains are rejected. Redirect URIs are
now parsed once (`java.net.URI`): userinfo, fragments and non-loopback plain http are rejected on the PARSED host.
`HydraDcrServerTestSpec` covers the callback-specific URI, mixed lists, lookalikes, userinfo smuggling, malformed
ports, fragments, and the preserved Claude/loopback cases. The rendered effective allowlist per environment (D16) is
still a deployment check.

Add tests for:

- The callback-specific ChatGPT connector redirect.
- Mixed valid redirect lists if OpenAI registers more than one callback.
- Rejection of lookalike domains, non-HTTPS remote callbacks, fragments, and unapproved domains.
- Preservation of the Claude Code, Cowork, and localhost cases.

### 1.2 Make protected-resource and issuer metadata consistent

The protected-resource metadata currently advertises an environment-scoped authorization-server URL, while the authorization-server metadata returns the base public URL as its issuer:

- `modules/mcp/mcp/jvm/src/anduin/mcp/auth/OAuth2McpAuthHttpExtension.scala`
- `modules/heimdall/heimdallApp/jvm/src/anduin/oauth2/hydra/HydraPublicServer.scala`

The candidate MVP design makes the authorization-server metadata `issuer` exactly equal the environment-scoped identifier advertised in protected-resource metadata and does not advertise `authorization_response_iss_parameter_supported`. First deploy this narrow metadata and callback slice to the selected public non-production environment. Promote it into the full Phase 1 change only after real successful and error authorization redirects prove that every returned `iss` is absent or exactly equals the environment-scoped metadata issuer. Codex rejects a mismatched returned `iss` even when issuer support is not advertised. If Hydra returns the current bare-host issuer, stop and redesign rather than shipping a metadata-only production change. The candidate continues to use the callback-specific ChatGPT redirect described in Phase 1.1.

This metadata correction must not silently change access-token semantics. Preserve JWT validation against the actual Hydra token issuer, then prove with real tokens that existing Claude clients and new OpenAI clients still pass issuer and MCP resource-audience validation. Treat a mismatch discovered in that end-to-end test as a design blocker, not as permission to loosen validation.

Add production-shaped golden tests that prove:

- Protected-resource `resource` equals the canonical MCP URL.
- `authorization_servers[0]` exactly equals the fetched metadata `issuer`.
- `authorization_response_iss_parameter_supported` is absent for the MVP.
- Successful and error authorization redirects either omit `iss` or return the exact metadata issuer, matching the
  Phase 0 observation for ChatGPT and direct Codex.
- Authorization, token, registration, and JWKS endpoints resolve correctly.
- Authorization and token requests echo and validate the MCP resource value.
- PKCE advertises and enforces `S256`.
- Tokens with the wrong issuer or resource audience are rejected.
- Existing Hydra-issued tokens continue to validate against the configured token issuer.

The test configuration must explicitly set `advertisedMcpHost` and exercise the trusted `x-anduin-forwarded-host` path; defaults and localhost-shaped fixtures are insufficient evidence for production discovery.

**Status:** candidate implemented behind `hydraConfig.envScopedMetadataIssuer` (default `false`,
`STARGAZER_SERVICES_HYDRA_ENV_SCOPED_METADATA_ISSUER`). `HydraMetadataIssuerTestSpec` and `OAuthDiscoveryGoldenSpec`
pin, in production shape (`id.anduin.app`, `mcp.anduin.app` via `x-anduin-forwarded-host`), that the flagged issuer
equals `authorization_servers[0]`, that RFC 9207 is not advertised, S256, and the endpoint set. Token-validation
semantics are untouched. Still open: the real success/error redirect gate (Spike 2) and the live-token tests.

### 1.3 Add complete tool security metadata

The current MCP tool wire model exposes `name`, `description`, `inputSchema`, and UI `_meta`, but not the standard security and annotation fields:

- `modules/mcp/mcp/jvm/src/anduin/mcp/protocol/McpTypes.scala`
- `modules/mcp/mcp/jvm/src/anduin/mcp/adapter/McpToolAdapter.scala`

Keep business/runtime tool definitions provider-neutral. Do not add OpenAI-only or MCP-only fields to `RuntimeTool`. Instead, add one exhaustively reviewed public-tool metadata table in `PublicMcpServerConfig`, adjacent to the existing read/write/admin allowlist buckets, and use it when adapting public tools to MCP descriptors.

Each public descriptor must include:

- A human-readable `title`.
- `securitySchemes` containing the exact OAuth scopes required by that tool.
- `annotations.readOnlyHint`.
- `annotations.destructiveHint`.
- `annotations.openWorldHint`.
- `outputSchema` for every tool that returns `structuredContent`.
- Any compatibility mirror of security metadata required by the OpenAI Apps SDK scanner.

Derive OAuth scope requirements from the existing public allowlist buckets so scope membership is not maintained twice. Use existing `RuntimeTool.isReadOnly` as the source for `readOnlyHint`. Keep explicit reviewed fields for destructive and open-world semantics in the public MCP metadata table instead of deriving them from tool names.

Materialize hierarchical scope closure when issuing new grants: an admin grant contains admin, write, and read, while a write grant contains write and read. Continue expanding scopes at the resource server so previously issued tokens remain valid during the transition. Advertise only the minimal allowlist-bucket scope on each tool and request only that missing minimal scope in a recovery challenge. Test consent text, stored grants, access- and refresh-token scope claims, descriptors, live discovery, execution authorization, and challenges together so the client-visible contract cannot diverge from server authorization.

Assign the SDK schemas from Phase 0A (`TableSpec.DisplayOnlyOutputSchema`, `ChartSpec.DisplayOnlyOutputSchema`,
`FormSpec.DisplayOnlyOutputSchema`; the public tools run display-only) to all eight UI-producing tools:

- Table: `render_table`, `show_funds`, `show_orders`, `show_fund_report`, `show_datarooms`, and `show_dataroom_insights`.
- Chart: `render_chart`.
- Form: `render_ui`.

Success `structuredContent` must conform exactly to the advertised schema. Error responses must not emit `structuredContent` that pretends to be a successful result.

At minimum, review these categories carefully:

- Archive, delete, remove-user, role/permission changes, and overwrite-style updates.
- Invitations or other operations that communicate with people outside the current account context.
- Temporary download URLs and any other action that creates an externally usable capability.
- Render tools, which should remain read-only and display-only.

Add contract tests that fail whenever a public tool lacks metadata, whenever a metadata entry has no public tool or vice versa, or whenever structured output violates its declared schema. At minimum, assert that every public tool has all standard annotations, every tool in `publicDataroomAdminToolNames` has `destructiveHint = true`, and every `isReadOnly` tool has `destructiveHint = false`.

**Status:** done. `McpToolDefinition` gained `title`, `outputSchema`, `annotations`, `securitySchemes` (all optional,
so internal and golden descriptors are byte-identical). `PublicMcpServerConfig.publicToolMetadata` is the reviewed
table (83 entries), `requiredScopeFor` derives the minimal scope from the allowlist buckets, `uiToolOutputSchemas`
assigns the SDK display-only schemas to the eight UI tools, and `describeTool` is the public server's descriptor
adapter (`McpServerConfig.toolDefinition`). `PublicToolMetadataContractTestSpec` pins table↔catalog equality, the
hint rules, the `$id`s, and validates real builder output against the advertised schema with a draft 2020-12
validator (test-only dependency), and executes the three render tools for real through the adapter. D4 closure:
DCR registers the hierarchical closure of explicit resource scopes, and both consent paths (interactive and
auto-approve) grant the approved scopes plus their closure bounded by the client's REGISTERED scope list
(`HydraConsentServer.materializeGrantScopes`), since Hydra v2 copies `grant_scope` verbatim into the token. The
resource server keeps expanding old tokens. Still open: a Hydra integration test asserting the token `scp` claim.

### 1.4 Complete the authentication error contract

- Preserve HTTP `WWW-Authenticate` challenges containing the protected-resource metadata URL.
- Return in-band `ToolsCallResult` errors with `isError = true` and `_meta["mcp/www_authenticate"]` when a call can be recovered by authenticating or granting a broader OAuth scope. The challenge must include the protected-resource metadata URL plus `error` and `error_description`; insufficient-scope responses must also identify the required scope.
- Distinguish unauthenticated, expired-token, insufficient-scope, and product-authorization failures.
- Change the current scope-filtered lookup behavior: when a requested name belongs to the known public catalog but is absent from the caller's filtered registry, return an `insufficient_scope` tool error rather than the current generic “unknown tool” protocol error. Truly unknown names remain unknown-tool errors.
- A product-role denial with otherwise sufficient OAuth scopes must remain a normal authorization error and must not tell the host to request broader scopes.
- Confirm that reconnecting or approving broader scopes exposes only the newly authorized tools.
- Never include tokens, client secrets, or authorization codes in logs or tool results.

**Status:** done. `McpIdentity.hiddenToolScope` (OAuth2: catalog tool outside the token's scope allowlist → its
minimal bucket scope) turns a known-but-hidden call into `isError` + `_meta["mcp/www_authenticate"]` =
`Bearer error="insufficient_scope", error_description=…, scope=<minimal>, resource_metadata=<PRM URL>`; the HTTP
`WWW-Authenticate` challenges share the same PRM URL helper. Truly unknown names stay `-32602`; a visible tool's
product-role denial carries no challenge. New audit outcome `DenyInsufficientScope`. Covered by `McpServerTestSpec`,
`OAuth2McpIdentityTestSpec`, `OAuthDiscoveryGoldenSpec`.

### 1.5 Preserve MCP Apps behavior

- Retain `ui://anduin/chart`, `ui://anduin/table`, and `ui://anduin/form` resources.
- Retain structured content and text fallbacks.
- Preserve restrictive CSP and sandbox behavior. Populate standard MCP Apps CSP fields and the OpenAI compatibility
  `_meta["openai/widgetCSP"].redirect_domains` with exactly `https://deals.anduintransact.com` for v1.
- The table widget currently uses `openLink`. Emit a clickable deep link only when its normalized origin is exactly
  `https://deals.anduintransact.com`; render fallback `*.anduin.io`, customer, non-US, and arbitrary
  model-supplied destinations as inert text. Do not use wildcard redirect domains or string-prefix matching.
- Set and test the dedicated submitted-UI origin through `_meta.ui.domain` and its `openai/widgetDomain` alias, both
  produced by `ResourcesRead.contents(..., openAi = Some(OpenAiWidgetOptions(redirectDomains = ...)))`.
- Keep rendering capability-discovered and independent of plugin version.
- Add OpenAI-host contract cases without weakening the existing Claude contract tests.

**Status:** done except the UI origin value. `WidgetOriginPolicy` (from `mcpConfig.widgetUiDomain` /
`widgetRedirectOrigins`; the default is EMPTY, fail-closed, and production US sets
`STARGAZER_SERVICES_MCP_WIDGET_REDIRECT_ORIGIN=https://deals.anduintransact.com` in its rivendell config) feeds
`UiResourceRegistry`
(`_meta.ui.domain` + derived `openai/widgetDomain`, `openai/widgetPrefersBorder`, `openai/widgetCSP` with exact
`redirect_domains`) and the server-side deep-link filter that downgrades non-allowlisted `link` cells of every table
`structuredContent` to inert text (normalized-origin equality, no wildcard; every `href`/`url` occurrence in a cell
must be allowlisted). The internal PAT server stays unrestricted. Covered by `WidgetOriginPolicySpec` and
`UiResourceRegistrySpec`. The `?v=` content hash now covers the widget HTML only, so it is policy-independent.
`widgetUiDomain` stays unset until the submitted UI origin exists (Phase 0 spikes); the emitted `_meta.ui.domain` and
`openai/widgetDomain` are asserted with a fixture value.

### Phase 1 exit criteria

- ChatGPT can discover the OAuth server and complete DCR, consent, PKCE, token exchange, refresh, and reconnect against a deployed test environment.
- The Phase 0 candidate issuer has passed real success/error redirect validation; authorization-server identifier and
  issuer then match exactly in the deployed environment.
- New hierarchical grants materialize their implied lower scopes while old tokens continue to work through
  resource-server expansion.
- Every public tool has reviewed security schemes and annotations, and all eight structured UI tools advertise schemas that their success responses satisfy.
- Missing-scope calls produce a recoverable in-band OAuth challenge; product-role denials do not request broader scopes.
- Widget deep links are limited to the exact v1 origin `https://deals.anduintransact.com` and work against the
  dedicated UI origin.
- The rendered effective DCR allowlist is recorded for the deployed test environment and retains existing localhost
  and Anthropic callbacks alongside `chatgpt.com`.
- Existing Claude OAuth and MCP integration tests still pass.
- MCP Inspector reports no blocking protocol or schema failures.

## Phase 2: Make Skills the Canonical Behavior

**Repository:** `anduin-plugin`

OpenAI supports skills but does not import Claude `agents/` definitions. Move reusable behavior into the two existing skills while keeping Claude adapters.

### 2.1 Consolidate domain guidance

Compare each agent with its corresponding skill:

- `plugins/anduin/agents/dataroom-agent.md`
- `plugins/anduin/skills/dataroom/SKILL.md`
- `plugins/anduin/agents/gp-assistant.md`
- `plugins/anduin/skills/gp-assistant/SKILL.md`

Move any reusable workflow, presentation, safety, confirmation, and recovery rules found only in the agents into the skills. The skills must remain the authoritative definitions for:

- Domain terminology and ambiguous terms.
- Tool selection and tool chaining.
- Opaque ID handling.
- OAuth scope and product-role distinctions.
- Confirmation before writes and destructive actions.
- Batch preview and partial-failure behavior.
- Capability-based widget rendering and markdown fallback.
- Unsupported requests and safe failure behavior.

### 2.2 Remove host-specific assumptions from skills

- Refer to “the host” or “the current client” when behavior depends on UI capability.
- Name ChatGPT, Codex, Claude Code, and Cowork only in installation or platform-specific sections.
- Do not hardcode host-specific MCP tool prefixes in canonical workflow prose unless they are part of the wire-level tool name.
- Do not claim render tools always exist; use the current host's provided tool catalog and UI capabilities. Do not assume that ChatGPT's published catalog is live or that Codex renders widgets.

### 2.3 Retain thin Claude adapters

Keep `plugins/anduin/agents/*.md` so current Claude activation behavior remains available. Reduce them to Claude-specific activation metadata, model/tool configuration, and an instruction to use the relevant canonical skill.

Do not maintain duplicate permission, destructive-tool, scope, or workflow prose in the adapters. Review their thinness during skill changes instead of adding a fragile text-drift check between intentionally different file formats.

### Phase 2 exit criteria

- ChatGPT, Codex, and Claude receive the same domain and safety behavior from the skills.
- Claude-specific agents contain no independent copy of mutable tool catalogs or authorization rules.
- Existing Claude skill activation and tool access continue to work.

## Phase 3: Add OpenAI Packaging

**Repository:** `anduin-plugin`
**Dependency:** Requires the Phase 0 app identifier and host/package spike results.

### 3.1 Finalize the registered app and discovery snapshot

- Re-register or update the app to point to the deployed production US MCP endpoint.
- Complete OAuth with the dedicated production-US full-scope synthetic review account from Phase 0; do not reuse the
  non-production spike identity.
- Run **Scan Tools** and verify that the draft contains the complete intended public catalog, titles, descriptions,
  input/output schemas, security schemes, annotations, tool `_meta`, server `instructions`, UI resource URIs, and
  linked UI metadata/CSP.
- Confirm the generated `plugin_asdk_app...` technical identifier and callback-specific redirect URI match the tested configuration.
- Do not commit a fabricated or placeholder app identifier.

The Phase 0 spike uses the selected public non-production environment, but the final package and submission must
reference a production registration and a snapshot captured from the deployed production-compatible metadata.

### 3.2 Add `.codex-plugin/plugin.json`

Include:

- Name, description, semantic version, author, homepage, repository, license, and keywords.
- Explicit references to both skills.
- The app mapping reference.
- Interface metadata and shared assets.
- Privacy policy, support, and terms links needed for publication.

Reference the app mapping through `.app.json`. Omit the OpenAI manifest's `mcpServers` field when Phase 0 proves that the app mapping supplies remote OAuth correctly on every intended Codex surface.

Keep the version aligned with the Claude plugin manifest.

### 3.3 Add `.app.json`

Map the package to the registered OpenAI app technical identifier. Treat the identifier as public package configuration, not a credential. Never include OAuth client secrets or user tokens.

Keep the existing `plugins/anduin/.mcp.json` unchanged and Claude-only; its `mcpServers` wrapper is not the OpenAI plugin-package schema. If the Phase 0 host spike proves a direct server declaration is also necessary, create `plugins/anduin/openai.mcp.json` using a currently documented OpenAI direct-server shape and reference that distinct file from the OpenAI manifest. Validate both paths independently.

### 3.4 Add an explicit OpenAI repository marketplace

Create `.agents/plugins/marketplace.json` pointing to `plugins/anduin`. Keep `.claude-plugin/marketplace.json` for Claude.

The two marketplace files may share name, description, source path, and version, but should carry host-appropriate policy, category, and interface metadata where supported.

### 3.5 Update documentation

Revise `README.md` to:

- Describe the project as the Anduin plugin rather than a Claude-only plugin.
- Provide separate installation instructions for ChatGPT, Codex, Claude Code, and Cowork.
- Document the shared OAuth scopes once.
- Explain host-dependent widget support and universal text fallbacks.
- Add reconnect and insufficient-scope troubleshooting for OpenAI hosts.
- Provide a compatibility matrix that distinguishes ChatGPT, Codex App, Codex CLI, Codex IDE extension, Codex cloud, Claude Code, and Cowork.
- Preserve developer instructions for changing environments without implying that every host supports local URLs.

Update `CLAUDE.md` only where its repository architecture or release instructions become stale. It may remain Claude-focused by design.

### 3.6 Add package validation

Add a small, dependency-light validator that checks:

- All JSON manifests parse.
- Referenced skills, apps, and assets exist.
- OpenAI-required skill frontmatter includes valid `name` and `description` values. Claude-only frontmatter remains optional and is validated only when present.
- Versions match across both plugin manifests, both marketplace entries, and the intended Git tag.
- No credential-shaped values appear in package files.
- The production MCP URL is consistent wherever it is intentionally duplicated.
- The Claude `.mcp.json` is not referenced by the OpenAI manifest; any conditional `openai.mcp.json` uses the OpenAI schema selected in Phase 0.

Run the validator in CI or the repository’s release workflow.

### Phase 3 exit criteria

- The OpenAI package installs from the local repository marketplace in Codex App, CLI, and IDE, and the supported cloud installation path is verified separately.
- ChatGPT recognizes the registered app and both skills.
- Claude installation still works from the existing marketplace.
- Package validation passes from a clean checkout.

## Phase 4: Cross-platform Verification

Maintain a versioned evaluation matrix in the repository or the release checklist.

### 4.1 Golden prompt set

Include at least five positive cases:

1. List funds and show a fund report.
2. Find LP orders with incomplete forms.
3. List data rooms and inspect participants or activity.
4. Perform a confirmed write, such as tagging orders or inviting a participant.
5. Render a chart or table when available and return an equivalent text result when unavailable.

Include at least three negative cases:

1. An unrelated request that must not select an Anduin tool.
2. An unsupported operation that must be explained without inventing a tool.
3. A request outside the approved scopes that must not be attempted through a broader or unrelated tool.

Add follow-up cases that reuse opaque IDs returned by earlier calls without reconstructing them.

### 4.2 Authorization matrix

Test each host with:

- No authentication.
- Read-only fund subscription scope.
- Fund subscription write scope.
- Read-only data room scope.
- Data room write without admin.
- Data room admin.
- Render scope alone and combined with each domain.
- Expired access token with valid refresh token.
- Expired or revoked refresh token requiring reconnect.
- Valid OAuth scope but insufficient Anduin product role.

Verify both tool visibility and server-side enforcement. A hidden tool is not a substitute for authorization at execution time.

Apply that assertion according to each discovery model:

- For Claude and direct Codex MCP connections, verify that live `tools/list` remains scope-filtered and changes after reconnect.
- For the published OpenAI plugin in both ChatGPT and Codex, verify that the full submitted snapshot remains stable,
  and that invoking a tool beyond the current grant produces the recoverable `insufficient_scope` tool result defined
  in Phase 1.4.
- For every host, verify execution-time scope and Anduin product-role enforcement independently of discovery.

### 4.3 Host matrix

Run the golden prompts and authorization cases in:

| Host | OAuth/installation path | Required result |
|---|---|---|
| ChatGPT | Registered app, callback-specific `chatgpt.com` redirect, and universal plugin | OAuth, stored tool snapshot, skill selection, widgets, and fallback pass |
| Codex App | Local repository marketplace; direct MCP uses a loopback redirect | Skills, DCR/PKCE, reconnect, tools, and text fallback pass |
| Codex CLI | Local repository marketplace; direct MCP uses a loopback redirect | Skills, DCR/PKCE, reconnect, tools, and text fallback pass |
| Codex IDE extension | Local repository marketplace; direct MCP uses a loopback redirect | Skills, DCR/PKCE, reconnect, tools, and text fallback pass |
| Codex cloud | Published-plugin or cloud connection path proven in Phase 0; use the observed ChatGPT-controlled callback | Skills, DCR/PKCE, reconnect, tools, and text fallback pass |
| Claude Code | Existing Claude marketplace | No regression |
| Cowork | Existing Claude marketplace | No regression |

Do not treat one Codex result as evidence for the other surfaces. Record callback, package-resolution, and OAuth behavior for each. ChatGPT must render and exercise the widgets; Codex acceptance depends on complete text/structured fallbacks and does not assume widget rendering.

### 4.4 Safety and interaction checks

- Writes require the existing confirmation behavior.
- Destructive data room operations require both `dataroom:admin` and the correct product role.
- Read tools are never mislabeled as destructive.
- Destructive or open-world tools are never mislabeled as harmless.
- Batch operations preview targets and report partial failures.
- Tool errors do not expose sensitive data.
- Widget content matches the machine-readable and text results.
- Table deep links open only for the exact v1 origin `https://deals.anduintransact.com`; other values render safely as
  non-clickable text.
- Every success `structuredContent` result validates against the output schema stored in the submitted tool snapshot.

### Phase 4 exit criteria

- All positive and negative evaluations pass on ChatGPT and each Codex surface.
- The Claude regression matrix passes.
- OAuth reconnect and insufficient-scope flows are understandable in every host.
- ChatGPT's stored catalog and live direct-client discovery behave according to their distinct contracts.
- No observed tool-selection failure can cause an unconfirmed write or authorization bypass.

## Phase 5: Release and Publication

### 5.1 OpenAI submission prerequisites

- Complete organization and business verification.
- Use an OpenAI project eligible for an MCP-backed public submission.
- Ensure the submitter has the required Apps Management permissions.
- Provide public privacy policy, terms, support, and homepage URLs.
- Provide a stable production endpoint and confirm that the dedicated production-US synthetic full-scope review
  account established in Phase 0 still has no MFA or manual-verification dependency. Keep the non-production spike
  account separate.
- Audit tool inputs, outputs, logs, and widgets for PII, credentials, authorization codes, and secrets.

### 5.2 Submit one universal listing

- Submit the integration as a plugin with MCP.
- Create a new draft, scan the production MCP server with the dedicated production-US synthetic full-scope review
  account, and resolve all metadata, schema, annotation, and UI findings before submission.
- Compare the stored scan snapshot against the complete reviewed contract: public-tool inventory, titles,
  descriptions, input/output schemas, security schemes, annotations, tool `_meta`, server `instructions`, UI resource
  URIs, and linked UI metadata/CSP. Fail the release if anything is absent, unintended, or stale.
- Attach the exact two canonical skill revisions evaluated for this release.
- Provide starter prompts, at least five positive tests, and at least three negative tests.
- Supply accurate release notes and supported-platform claims.
- Publish one listing for ChatGPT and Codex.

For every subsequent release, repeat draft creation, **Scan Tools**, complete-contract snapshot comparison, skill
attachment, evaluations, and submission whenever any captured field or either skill changes. Changes to the MCP
scheme, hostname, or port require a new plugin. Breaking tool, schema, or UI-resource changes must first ship
additively and remain compatible with the active published snapshot until the replacement is approved and the rollback
window closes. Do not assume that a live server deployment updates the published discovery snapshot in ChatGPT or
Codex.

### 5.3 Rollout sequence

1. Deploy the backward-compatible server changes.
2. Verify Claude against the deployed server.
3. Test the unpublished OpenAI registration with internal users.
4. Release the multi-host plugin package and synchronized version.
5. Capture and approve the production tool/skill snapshot, then submit or enable the OpenAI listing for a limited cohort where possible.
6. Expand availability after observing stable authentication, tool selection, and error rates.

### 5.4 Observability

Track, without recording sensitive values:

- DCR failures by client and reason.
- OAuth discovery, authorization, token, refresh, and reconnect failures.
- Insufficient-scope and product-authorization failures.
- Known-but-scope-hidden tool calls separately from truly unknown tool names.
- Tool invocation success, clean errors, and unexpected errors by host.
- Tool-selection misses and unsupported-request false positives from evaluation or feedback.
- Widget render failures and fallback usage.
- Confirmation cancellation rates for write and destructive tools.

### 5.5 Rollback

- Keep all server changes backward-compatible so they need not be reverted merely to pause OpenAI availability.
- Pause or unpublish the OpenAI listing independently of Claude releases.
- Retain the previous plugin tag and package metadata for a fast package rollback.
- Do not remove OpenAI OAuth clients until active sessions and rollback needs have been assessed.

## Pull Request Sequence

### PR 0: `mcp-ui-scala` — Structured-output and widget CSP contracts

Opened as [PR #30](https://github.com/anduintransaction/mcp-ui-scala/pull/30) (`phase-0a/openai-ui-contracts`).

- Export sound Table, Chart, and Form JSON Schemas next to their validators. **Done**
- Add schema conformance tests. **Done**
- Encode OpenAI `widgetCSP.redirect_domains` compatibility metadata. **Done**
- Model the submitted UI domain metadata. **Done**
- Release a new SDK version (`v0.5.0`). **Done**

### PR 1: `stargazer` — OpenAI MCP compatibility

- ChatGPT redirect allowlist and tests.
- Candidate environment-scoped metadata issuer correction without RFC 9207 advertisement, promoted only after the
  Phase 0 success/error redirect gate passes.
- New-grant hierarchical scope closure with old-token resource-server compatibility.
- Tool security schemes and annotations.
- Output schemas for all eight UI-producing tools using the new SDK release.
- Authentication error metadata and known-but-scope-hidden call handling.
- Exact v1 widget deep-link origin and submitted UI domain.
- OpenAI-shaped protocol and OAuth contract tests.
- Claude regression coverage.

Deploy and verify this PR before merging the public package changes. Bump `McpUi.version` in `stargazer/build/versions.mill` to the released PR 0 version.

### PR 2: `anduin-plugin` — Canonical skills and multi-host package

- Consolidated provider-neutral skills.
- Thin Claude agent adapters.
- `.codex-plugin/plugin.json`.
- `.app.json` using the real registered app identifier.
- Conditional `openai.mcp.json` only if the Phase 0 Codex matrix requires it; preserve the Claude `.mcp.json` unchanged.
- `.agents/plugins/marketplace.json`.
- Shared assets and legal/support metadata.
- README compatibility and installation updates.
- Package/version validator.
- Golden prompt and host acceptance checklist.

### Release operation

- Verify exact deployed server and plugin revisions.
- Run the full host and authorization matrix.
- Create a synchronized semantic-version tag.
- Publish the Claude update.
- Submit or publish the OpenAI universal listing.

## Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Strict OpenAI OAuth discovery rejects the current issuer mismatch | Gate the candidate design on real successful and error redirects in public non-production, then require identifier equality in production-shaped tests |
| Hydra returns a bare-host `iss` even when RFC 9207 support is not advertised | Treat every returned mismatch as a Codex blocker and redesign before production; do not ship the metadata-only issuer change |
| Environment-scoped metadata issuer diverges from Hydra token issuer | Keep token validation strict and gate release on real-token ChatGPT, Codex, and Claude tests |
| Adding ChatGPT callbacks broadens unauthenticated DCR exposure | Allow only the documented `chatgpt.com` domain, use the callback-specific flow, retain HTTPS and URI validation, and test lookalikes |
| Scan Tools captures an incomplete scope-filtered contract | Scan with the dedicated production full-scope account and compare every captured tool, instruction, and linked UI field before each submission |
| Published metadata or skills become stale after a live change | Treat the published snapshot as authoritative in both ChatGPT and Codex; require a new draft, complete rescan or skill reattachment, evaluation, and resubmission as part of the release checklist |
| Incorrect tool annotations suppress warnings or cause excessive confirmations | Maintain explicit reviewed classifications and fail tests for missing metadata |
| Structured output drifts from advertised schemas | Own schemas next to SDK validators and validate every UI tool's success output in `stargazer` contract tests |
| Widget links escape the intended Anduin origin | Allow only `https://deals.anduintransact.com` in v1 CSP and render fallback, customer, non-US, and arbitrary URLs as inert text |
| Skills and Claude agents drift | Make skills canonical and reduce agents to thin host adapters with no duplicated mutable policy |
| A package release precedes server readiness | Deploy and verify PR 1 before publishing PR 2 |
| Large tool inventory causes poor selection | Improve titles/descriptions and use direct, indirect, follow-up, and negative prompt evaluations |
| Codex surfaces resolve packages, OAuth, or callbacks differently | Test App, CLI, IDE, and cloud independently; retain complete text fallbacks |
| `.mcp.json` schemas collide between Claude and OpenAI | Keep the Claude file private to its manifest and use a separate OpenAI-format file only when the Phase 0 spike requires it |
| Hierarchical grants look narrower to clients than the access they confer | Materialize lower-scope closure in new grants, retain resource-server expansion for old tokens, advertise minimal tool scopes, and test consent, token claims, discovery, execution, and challenges together |
| OpenAI review or organization setup delays launch | Start business verification, legal URLs, disposable app registration, and synthetic review-account preparation in Phase 0 |

## Definition of Done

The project is complete when all of the following are true:

- One published OpenAI plugin is available in both ChatGPT and Codex.
- Both canonical skills activate reliably for direct and indirect requests.
- ChatGPT completes OAuth discovery, DCR, PKCE, consent, refresh, and reconnect against production.
- The candidate issuer design is proven with real successful and error redirects before production; protected-resource
  authorization-server identifiers then exactly match authorization-server issuers.
- The MVP uses the callback-specific ChatGPT redirect and does not advertise unsupported RFC 9207 response issuer behavior.
- The rendered effective DCR allowlists are recorded for the spike and production environments and preserve existing
  Claude/loopback callbacks while adding `chatgpt.com`.
- New admin/write grants expose their complete hierarchical scope closure to clients, while old tokens remain valid and
  every tool descriptor/challenge names only its minimal required scope.
- Every public MCP tool has correct security schemes and reviewed annotations; all eight UI-producing tools have sound, passing output schemas.
- The published OpenAI snapshot used by ChatGPT and Codex contains the complete reviewed public catalog, server
  instructions, and linked UI contract, while live direct-client discovery and execution authorization match approved
  scopes and Anduin product roles.
- Known-but-scope-hidden calls return a recoverable in-band OAuth challenge; product-role denials do not ask for broader OAuth scopes.
- Destructive actions retain appropriate confirmation and admin gates.
- Widgets work in ChatGPT, equivalent text is always available, and v1 deep links can target only the exact
  `https://deals.anduintransact.com` origin.
- Codex App, CLI, IDE, and cloud each pass their own package, OAuth, reconnect, skill, tool, and fallback checks.
- Claude Code and Cowork pass the regression suite.
- All package versions and references validate automatically.
- Production server revision, plugin tag, and OpenAI listing version are recorded in the release evidence.
- Monitoring and independent OpenAI rollback procedures are in place.

## Official References

- [Plugin architecture](https://developers.openai.com/plugins/concepts/plugins)
- [Skills](https://developers.openai.com/plugins/concepts/skills)
- [Package your plugin](https://developers.openai.com/plugins/build/plugins)
- [Authenticate users](https://developers.openai.com/plugins/build/auth)
- [Plugin and MCP Apps reference](https://developers.openai.com/plugins/reference)
- [Build ChatGPT UI](https://developers.openai.com/plugins/build/chatgpt-ui)
- [Connect and test your plugin](https://developers.openai.com/plugins/deploy/connect-chatgpt)
- [Submit and publish](https://developers.openai.com/plugins/deploy/submission)
- [MCP server review requirements](https://developers.openai.com/plugins/deploy/app-review)
- [Connect Codex to MCP servers](https://learn.chatgpt.com/docs/extend/mcp)
- [Submit a Claude Code plugin](https://developers.openai.com/plugins/guides/submit-claude-plugin)
- [Optimize tool metadata](https://developers.openai.com/plugins/guides/optimize-metadata)
