# ChatGPT and Codex Support Plan

**Status:** In progress — Phase 0A released; Phase 1 review fixes published in [PR #55982](https://github.com/anduintransaction/stargazer/pull/55982), under verification/review;
live host validation and public-distribution approval remain open. See the readiness gates below.
**Last updated:** 2026-09-05
**Repositories:** `mcp-ui-scala`, `stargazer`, `anduin-plugin`
**Primary production endpoint:** `https://mcp.anduin.app/mcp`

## Summary

Support ChatGPT and Codex through one OpenAI plugin package backed by the existing Anduin MCP server and the existing GP Assistant and Data Room skills. Preserve Claude Code and Cowork support by retaining the Claude manifests, MCP configuration, and thin Claude-specific agent adapters.

ChatGPT and Codex should share domain behavior and server implementation. A universal public listing remains the
target, but package/import availability differs by surface. Do not assume a local MCP declaration or a successful
direct Codex login proves the registered-app path. Keep one canonical skill source; generate distribution-specific
packaging only if the host-validation gate demonstrates that it is necessary and the product owner approves it.

The work is split into three code tracks followed by deployment and publication:

1. Export exact structured-output schemas and OpenAI widget CSP compatibility from `mcp-ui-scala`.
2. Make the public MCP server compliant with OpenAI OAuth, tool metadata, snapshot discovery, and review requirements in `stargazer`.
3. Canonicalize the skills and add OpenAI packaging in `anduin-plugin` without removing Claude support.
4. Deploy the server, run cross-host acceptance tests, and submit one universal OpenAI listing.

## Readiness gates and next work

"Implemented", "merged", "deployed", "host-validated", and "approved for public distribution" are distinct states.
Unit tests, a real local Hydra, and a published SDK do not prove a live ChatGPT connection or production readiness.

| Gate | Current state | Exit evidence / accountable role |
|---|---|---|
| G0 — SDK | Delivered: PR #30, `v0.5.0` | Released artifacts consumed by Stargazer; SDK maintainer |
| G1 — Phase 1 code | PR #55982 open; local verification passed, CI/review pending | Exact head, green required CI, reviewer approval; server maintainer |
| G2 — Deployed compatibility | Not proven | Selected public non-production revision/config, edge-header test, real OAuth success/error redirects, exchange/refresh/reconnect, Inspector and Claude regression; deployment owner + host tester |
| G3 — Distribution and data | Open; blocks public launch | Per-tool data inventory, approved distribution catalog and tested server-side controls, legal/security sign-off; product + security/privacy owners |
| G4 — Package and hosts | Not implemented/proven | Real app mapping, clean-install connection/skill tests per named surface, stored scan comparison, dedicated UI origin; plugin maintainer + host tester |
| G5 — Production/public launch | Blocked on G2–G4 | Exact deployed revision, production synthetic-account evaluations, required approvals, monitoring/rollback rehearsal, accepted listing; release owner |

Next: close G1, then deploy the default-off issuer candidate to the selected non-production environment and execute
G2. Phase 2.1–2.3 (canonical skill content and thin Claude adapters) can proceed independently now. Phase 2.4 and
Phase 3 wiring need the real connection and host observations. Begin G3 in parallel: policy controls are a separately
scoped implementation track if needed, not functionality already supplied by Phase 1.

A workspace-private pilot is a separate, explicitly approved milestone, not completion of the universal-public
goal. If a host or public listing cannot be supported, record the evidence and obtain an explicit scope decision;
do not silently drop Codex cloud or substitute desktop-only distribution.

For each gate, record date, named owner, repository/PR head, deployed image/config digest where applicable, host
version, test identity class (no credentials), result and evidence link. Assign actual people before deployment;
role labels here are not assignments. Historical probes below retain their original dates.

### G1 revision evidence (2026-09-05)

- [Stargazer PR #55982](https://github.com/anduintransaction/stargazer/pull/55982), branch `codex/openai-phase-1`,
  head `93ba809fbe2e9a6133f39489fb0b119d1d491166`, base `323f1c0dcb5d329b59d0b0ed8bdd0c7837dcc8e9`.
- The original implementation and review-fix commits rebased without changes (`git range-diff` equality).
  The unrelated AI Usage dashboard edit is excluded; its original checkout/patch is preserved separately because
  it conflicts with upstream's dashboard work.
- Local verification on that rebased head: **669 Scala tests passed** — all MCP unit tests (458), selected Hydra
  consent/DCR/metadata tests (58), shared file/folder metadata tests (16), Data Room insight/widget/failure tests (33),
  LP review execution/schema tests (89), and real-Hydra `McpOAuthConsentFlowInteg` (15). The latter uses real Hydra
  with production consent logic and a Tapir HTTP stub; login is accepted via the Admin API, not a browser/edge test.
- Gateway: TypeScript compile-check, both edge-header regression tests, and scoped Biome checks passed.
  Affected Scala modules pass `checkStyleCached`; `git diff --check` is clean.
- CI and reviewer approval remain required. Draft checks skipped build/test jobs; skipped checks are not green
  evidence. Inspect the current required checks on the ready PR before merge. No deployment, host registration,
  production account change or public publication was performed by this implementation pass.

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
        ├── .mcp.json                # Existing Claude input; OpenAI import collision must be tested
        ├── openai.mcp.json          # Conditional direct-client artifact, not a universal fallback
        ├── agents/                  # Thin Claude-specific adapters
        ├── skills/                  # Canonical cross-host behavior
        │   ├── gp-assistant/
        │   │   ├── SKILL.md
        │   │   └── agents/openai.yaml # Required Anduin MCP dependency
        │   └── dataroom/
        │       ├── SKILL.md
        │       └── agents/openai.yaml # Required Anduin MCP dependency
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

Record design decisions before implementation. Validate uncertain host behavior after the necessary candidate server
slice is available; Phase 0's live-validation gates are not a prerequisite for writing that slice. Keep the issuer
flag off outside the explicitly selected spike environment until D5 passes.

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
  data. A review-only account without interactive MFA requires a documented, narrowly scoped security exception;
  never disable organization-wide MFA. Keep the identity valid for review and retire it through the approved process.
- If authenticated scanning cannot enumerate the approved distribution catalog, stop the launch and choose an
  explicit package/review redesign. Do not expose an anonymous catalog by default. The baseline 83-tool inventory is
  a code-contract count, not an automatic approval to publish all 83 tools (see G3).
- Changes to tools, schemas, security schemes, annotations, tool `_meta`, server `instructions`, UI resource URIs, or
  linked UI metadata/CSP require a compatible deployment, new draft, fresh **Scan Tools**, regression testing, and
  resubmission. Origin changes require a new plugin. Breaking removals, renames, schemas, and UI resources require an
  additive migration without breaking any active published snapshot. Do not retire an old contract merely because a
  rollback window elapsed. Server-only fixes that preserve captured contracts may ship independently.
- Skills are also attached as submission artifacts. Skill changes require a new plugin version/draft, reattachment, evaluation, and resubmission.
- Read-only, destructive, and open-world classifications are reviewed per tool rather than inferred from naming
  conventions. Each tool advertises its minimal exact scope from the existing allowlist bucket; materialized scope
  closure lets higher grants satisfy lower-scope tools, and recovery challenges request only the missing minimal scope.
- Existing Claude manifests and agents remain supported.
- The OpenAI manifest uses `"apps": "./.app.json"`. Preserve the existing Claude `.mcp.json` source, but test whether
  its mere presence affects each OpenAI importer; omission of `mcpServers` is not proof it will be ignored.
- A separate `openai.mcp.json` may support a direct-client path, but is not a universal fallback: the documented
  workspace import path marks MCP-declaring packages desktop-only, including remote HTTPS servers. If necessary,
  produce an OpenAI import artifact excluding Claude MCP configuration from the same canonical source, with an
  explicit manifest/file allowlist and reproducibility test. Approve the layout after the spike, before publishing.
  See [workspace plugin management](https://learn.chatgpt.com/docs/enterprise/plugin-management).
- ChatGPT renders MCP Apps UI. Codex workflows must remain fully useful from text and structured tool results without relying on widget rendering.
- V1 widget links are clickable only for the exact origin `https://deals.anduintransact.com`; fallback, customer,
  non-US, and arbitrary model-supplied destinations render as inert text. No wildcard CSP is used.
- Inspect the rendered effective DCR allowlist in every spike and production environment because out-of-repo overrides
  replace the defaults wholesale. Empty or missing configuration must fail closed; it must never disable redirect-domain
  validation.
- Choose and record the OpenAI distribution mode only after a tool-by-tool data-policy review. A universal public
  listing remains blocked while any tool can solicit, process, return, or expose OpenAI Restricted Data, including
  government identifiers, payment-card data, health information, or credentials. W-9/W-8BEN, AML/KYC, arbitrary
  form, and OCR workflows require explicit server-enforced exclusion or pre-boundary redaction, not prompt-only
  instructions, output-only redaction, or a generic PII audit. Regulated sensitive data additionally requires
  documented strict necessity, legally adequate consent, and prominent disclosure. If those controls cannot be proven,
  request approval for a workspace-private pilot and do not claim public-listing readiness or project completion.
  Private distribution does not waive
  the remaining privacy and security requirements. See the
  [OpenAI plugin guidelines](https://developers.openai.com/plugins/app-guidelines).

Run the human-dependent work in this order; the detailed log is
`docs/architecture/openai-compatibility-contract.md`:

1. Create a disposable developer-mode registration and capture its actual identifiers and callback-specific redirect
   URI. Distinguish the `plugin_asdk_app...` URL/creator identifier from the underlying app ID. Current package-builder
   and workspace-import docs describe different ID forms; validate the exact mapping accepted by each intended path
   rather than mechanically copying a URL ID into every manifest. Do not commit placeholders.
2. Deploy the narrow candidate callback/issuer slice to the selected public non-production environment. Capture
   successful and error authorization redirects from ChatGPT and direct Codex. If any returned `iss` differs from the
   candidate metadata issuer, reject the candidate design; do not promote a metadata-only issuer change to production.
3. After the issuer design passes, complete OAuth with the non-production synthetic account, run **Scan Tools**, and
   verify the complete intended test contract. Approve G3's distribution inventory before the final production scan.
4. Install a package that references only `.app.json` and verify connection, registration method, exact callback,
   PKCE, reconnect, tool/catalog behavior, and fallback requirements separately in Codex App, CLI, IDE extension, and
   cloud. Test registered-app wiring and direct MCP separately, from isolated profiles without ambient Anduin
   connections; CLI/App/IDE can share local MCP credentials, so one login is not three independent validations.

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
- The distribution mode, per-tool data classification, and server-enforced Restricted Data exclusions or pre-boundary
  redactions are recorded; otherwise public submission is explicitly blocked.
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

The original gap was that the eight public UI-producing tools returned `structuredContent` without exported SDK
schemas. Phase 0A closed that gap; the following describes the delivered contract, not remaining implementation.

SDK-owned, versioned JSON Schema constants accompany the corresponding validators:

- `TableSpec` schema for `render_table`, `show_funds`, `show_orders`, `show_fund_report`, `show_datarooms`, and `show_dataroom_insights`.
- `ChartSpec` schema for `render_chart`.
- `FormSpec` schema for `render_ui`.

The schemas must be sound for every structured result emitted on success (required fields, unions, nullability,
`additionalProperties`) and must reject every validator-rejected payload that JSON Schema can express. Constraints JSON
Schema cannot express (unique column ids and field aliases, byte and character size caps, the `editable_columns`
cross-check, the form-wide field total) stay validator-only and are listed in each schema's `description`. Tests must
validate representative success payloads, reject expressible validator rejections, and pin the validator-only gaps.

The SDK encoder supports `_meta["openai/widgetCSP"].redirect_domains` in addition to standard MCP Apps CSP fields,
plus `_meta.ui.domain` and its OpenAI compatibility alias for the submitted-UI origin.

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
**Status (2026-09-05):** implementation and review fixes are published on `codex/openai-phase-1` in PR #55982.
The four latest findings cover scheme-header spoofing, swallowed Data Room operation failures, audience validation
before access guards, and the missing empty fund-report schema fixture. The implementation now fixes these paths;
G1 requires passing regression evidence and a focused upstream PR. "Code implemented" below never means deployed or
host-validated. Section 1.2 remains a default-off candidate. The dependency is released SDK `0.5.0`.

### 1.1 Accept ChatGPT OAuth callbacks

The DCR redirect-domain allowlist currently contains only localhost and Anthropic-controlled domains:

- `platform/stargazerConfig/shared/src/com/anduin/stargazer/service/GondorBackendConfig.scala`
- `platform/stargazerConfig/jvm/resources/reference.conf`

Allow the documented `chatgpt.com` callback domain in the DCR policy and keep environment overrides in sync. This domain-level change can be implemented before the final app registration. Out-of-repository environment overrides replace the default list wholesale, so capture and review the rendered effective allowlist in the selected public non-production environment and in production; both must retain localhost and Anthropic callbacks while adding `chatgpt.com`. For the MVP, register and test only the callback-specific form emitted by app management: `https://chatgpt.com/connector/oauth/{callback_id}`.

Do not enable the stable `https://chatgpt.com/connector_platform_oauth_redirect` callback in the MVP. It requires RFC 9207 issuer identification on every successful and error authorization response, which the current Hydra flow does not provide consistently.

**Code status:** implemented; deployed configuration unverified. `chatgpt.com` is in the in-repo default (`GondorBackendConfig.HydraConfig.dcrAllowedRedirectDomains`
and `reference.conf`), but unlike the Anthropic domains it is matched exact-host and only on
`/connector/oauth/{callback_id}`, so the stable callback, other paths and subdomains are rejected. Redirect URIs are
now parsed once (`java.net.URI`): userinfo, fragments and non-loopback plain http are rejected on the PARSED host.
`HydraDcrServerTestSpec` covers the callback-specific URI, mixed lists, lookalikes, userinfo smuggling, malformed
ports, fragments, and the preserved Claude/loopback cases. The rendered effective allowlist per environment (D16) is
still a deployment check. An empty or missing effective allowlist now fails closed: `HydraPublicServer` admits only
loopback callbacks in that state and `HydraDcrServerTestSpec` pins it. DCR also registers the configured MCP resources
as the client's `audience` whitelist, which Hydra requires for refresh.

Add tests for:

- The callback-specific ChatGPT connector redirect.
- Mixed valid redirect lists if OpenAI registers more than one callback.
- Rejection of lookalike domains, non-HTTPS remote callbacks, fragments, and unapproved domains.
- Preservation of the Claude Code, Cowork, and localhost cases.
- Rejection of every remote redirect when the configured allowlist is empty or missing, preserving the explicit
  loopback exception. Fail-closed remote registration does not mean local Codex/Claude callbacks are disabled.

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
- Authorization requests carry and validate the MCP resource value; token requests echo it. **Observed:** Hydra 2.2.0
  ignores `resource` at the token endpoint (no `invalid_target`), so the binding rests on consent plus the client
  audience whitelist — a code exchange or refresh naming a wrong resource still yields a token bound ONLY to the
  consented resource, never the requested one (pinned by `McpOAuthConsentFlowInteg`).
- PKCE advertises and enforces `S256`.
- Tokens with the wrong issuer or resource audience are rejected. Audience matching must bind to the exact trusted
  canonical MCP scheme, host, effective port, and `/mcp` path, with no userinfo, query, or fragment; a relative `/mcp`
  reference or another host with that path must fail. The authorization/consent validation path and resource-server
  validation must use the same canonicalization and trusted-host source.
- Existing Hydra-issued tokens continue to validate against the configured token issuer. Scope closure remains
  backward-compatible, but exact audience enforcement can reject legacy tokens minted for a different host. Inventory
  existing Claude/base-host/custom-domain audiences before rollout, test representative refreshes, and document any
  necessary reconnect. Do not weaken audience validation to preserve incorrectly bound sessions.
- Wrong-resource tokens return 401 before product entitlement checks; a valid audience with insufficient product
  access returns 403 and an entitlement outage remains a non-401 failure. Exercise both through the HTTP boundary.

The test configuration must explicitly set `advertisedMcpHost` and exercise the trusted `x-anduin-forwarded-host` path; defaults and localhost-shaped fixtures are insufficient evidence for production discovery.

**Status:** candidate implemented behind `hydraConfig.envScopedMetadataIssuer` (default `false`,
`STARGAZER_SERVICES_HYDRA_ENV_SCOPED_METADATA_ISSUER`). `HydraMetadataIssuerTestSpec` and `OAuthDiscoveryGoldenSpec`
pin, in production shape (`id.anduin.app`, `mcp.anduin.app` via `x-anduin-forwarded-host`), that the flagged issuer
equals `authorization_servers[0]`, that RFC 9207 is not advertised, S256, and the endpoint set. Token-validation
semantics are otherwise untouched. Audience binding is now exact in both paths: `McpResourceIdentifier` canonicalizes
`http(s)://host[:port]/mcp` (no userinfo/query/fragment), the resource server requires `aud` to equal the edge-derived
request resource, and consent (`McpResourceAudiencePolicy`) accepts only the configured base URLs or a registered
custom domain as its default https origin (`https://host/mcp`; any other port must be an explicitly configured
resource). The gateway overwrites `x-anduin-forwarded-host` from `%REQ(:authority)%` and sets public
`x-forwarded-proto=https`. When a trusted edge host is present, MCP derives the resource scheme as HTTPS explicitly;
it never consults client `cf-visitor`, `Forwarded`, or `x-forwarded-proto` to choose that scheme. Direct/local traffic
uses configured fallback URLs. The deployment owner must verify TLS listeners, host routing, header overwrite, and
that untrusted clients cannot reach the backend directly; header names alone confer no trust. Still open: the real
success/error redirect gate (Spike 2) and live tokens in a deployed environment (local real-Hydra coverage:
`McpOAuthConsentFlowInteg`, including the custom-domain audience patch and the refresh that depends on it).

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

Materialize hierarchical scope closure when issuing new grants: admin contains admin/write/read, and write contains
write/read. Continue resource-server expansion for legacy scopes, while enforcing the exact audience contract in
1.2. Advertise/request only the minimal bucket scope. Test consent text, stored grants, access/refresh-token claims,
descriptors, discovery, execution and challenges together so client metadata cannot diverge from authorization.

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

**Code status:** implemented; live scan unverified. `McpToolDefinition` gained `title`, `outputSchema`, `annotations`, `securitySchemes` (all optional,
so internal and golden descriptors are byte-identical). `PublicMcpServerConfig.publicToolMetadata` is the reviewed
table (83 entries), `requiredScopeFor` derives the minimal scope from the allowlist buckets, `uiToolOutputSchemas`
assigns the SDK display-only schemas to the eight UI tools, and `describeTool` is the public server's descriptor
adapter (`McpServerConfig.toolDefinition`). `PublicToolMetadataContractTestSpec` pins table↔catalog equality, the
hint rules, the `$id`s, and validates real builder output against the advertised schema with a draft 2020-12
validator (test-only dependency), and executes the three render tools for real through the adapter; the five
source-backed `show_*` projections are executed for real in their source modules (`DataRoomShowToolsWidgetSpec`,
`LpReviewToolExecutionTestSpec`), populated and empty-state, and validated against the same display-only table schema
they advertise. D4 closure:
DCR registers the hierarchical closure of explicit resource scopes, and both consent paths (interactive and
auto-approve) grant the approved scopes plus their closure bounded by the client's REGISTERED scope list
(`HydraConsentServer.materializeGrantScopes`), since Hydra v2 copies `grant_scope` verbatim into the token. The
resource server keeps expanding old tokens. `McpOAuthConsentFlowInteg` proves the `scp` closure and the bound `aud` on
the issued AND refreshed token against a real Hydra, through `HydraConsentServer.approveConsentRequest` (the same
function the consent route calls). It also surfaced that Hydra re-validates the granted audience against the client's
`audience` whitelist on refresh: DCR now registers the configured MCP resources there and consent APPENDS a validated
tenant resource (re-read → `add /audience/-` → re-read to verify with bounded retries;
`HydraConsentServerTestSpec`), otherwise every DCR client's refresh failed. Mocked interleavings are not proof of
distributed atomicity: verify parallel consents and subsequent refreshes against the deployed Hydra before claiming
concurrent audience retention. The local integ test proves the patch and the
refresh it enables for a custom-domain resource absent from the DCR audience. Every repeated RFC 8707 `resource`
parameter is kept and validated. Hints: tag tools (replace/clear), renames, field and custom-data updates and the four
admin tools are destructive; `draft_comment` (public to the LP by default), invitations and download/invitation links
are open-world; both sets are pinned exactly in the contract test. Tool failures: the public Data Room tools no longer
return "Error …" strings as successful results — validation, scope denials, `GeneralServiceException` and
`DataRoomException` fail as `CleanToolError` (rendered `isError=true`), anything else propagates to the adapter's
opaque correlation-id error (`DataRoomAgentTools.surfaceFailure`; `DataRoomToolFailureSurfacingSpec`,
`PublicDataRoomToolErrorContractTestSpec`). Operation-level analytics, summary/detail counts, search branches and
metadata reads must propagate failures rather than synthesize empty/zero success. `DataRoomOperationalFailureContractSpec`
injects failures below the real operations. Shared `FileFolderInfoTools` intentionally makes missing, inaccessible,
operationally failed and defective metadata reads the same clean **failed** result to avoid an existence oracle;
interruption propagates. `FileFolderInfoToolsSpec` pins this privacy exception. Empty fund-report output is explicitly
validated alongside populated reports. These are code contracts, not evidence of live-host error presentation.

### 1.4 Complete the authentication error contract

- Preserve HTTP `WWW-Authenticate` challenges containing the protected-resource metadata URL.
- Return in-band `ToolsCallResult` errors with `isError = true` and `_meta["mcp/www_authenticate"]` when a call can be recovered by authenticating or granting a broader OAuth scope. The challenge must include the protected-resource metadata URL plus `error` and `error_description`; insufficient-scope responses must also identify the required scope.
- Distinguish unauthenticated, expired-token, insufficient-scope, and product-authorization failures.
- Change the current scope-filtered lookup behavior: when a requested name belongs to the known public catalog but is absent from the caller's filtered registry, return an `insufficient_scope` tool error rather than the current generic “unknown tool” protocol error. Truly unknown names remain unknown-tool errors.
- A product-role denial with otherwise sufficient OAuth scopes must remain a normal authorization error and must not tell the host to request broader scopes.
- Confirm that reconnecting or approving broader scopes exposes only the newly authorized tools.
- Never include tokens, client secrets, or authorization codes in logs or tool results.

**Code status:** implemented; host recovery unverified. `McpIdentity.hiddenToolScope` (OAuth2: catalog tool outside the token's scope allowlist → its
minimal bucket scope) turns a known-but-hidden call into `isError` + `_meta["mcp/www_authenticate"]` =
`Bearer error="insufficient_scope", error_description=…, scope=<minimal>, resource_metadata=<PRM URL>`; the HTTP
`WWW-Authenticate` challenges share the same PRM URL helper. Truly unknown names stay `-32602`; a visible tool's
product-role denial carries no challenge. New audit outcome `DenyInsufficientScope`. Covered by `McpServerTestSpec`,
`OAuth2McpIdentityTestSpec`, `OAuthDiscoveryGoldenSpec`. Authentication ordering is token validation → exact request
audience → product access guard; real-Hydra HTTP coverage rejects a wrong audience without calling the failing guard.

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

**Code status:** implemented; actual UI origin and live acceptance pending. `WidgetOriginPolicy` (from `mcpConfig.widgetUiDomain` /
`widgetRedirectOrigins`; the default is EMPTY, fail-closed, and production US sets
`STARGAZER_SERVICES_MCP_WIDGET_REDIRECT_ORIGIN=https://deals.anduintransact.com` in its rivendell config) feeds
`UiResourceRegistry`
(`_meta.ui.domain` + derived `openai/widgetDomain`, `openai/widgetPrefersBorder`, `openai/widgetCSP` with exact
`redirect_domains`) and the server-side deep-link filter that downgrades non-allowlisted `link` cells of every table
`structuredContent` to inert text (normalized-origin equality, no wildcard; every `href`/`url` occurrence in a cell
must be allowlisted). The internal PAT server stays unrestricted. Covered by `WidgetOriginPolicySpec` and
`UiResourceRegistrySpec`. The `?v=` content hash now covers the widget HTML only, so it is policy-independent.
`widgetUiDomain` stays unset until the submitted UI origin exists (Phase 0 spikes); when set it must be exactly one
canonical https origin or startup fails (`WidgetOriginPolicy.validateUiDomain`); the emitted `_meta.ui.domain` and
`openai/widgetDomain` are asserted with a fixture value.

Set the actual dedicated UI origin before the final scan, and test resource fetch, initialization, CSP and navigation
in ChatGPT. Non-production defaults remain inert-link/fail-closed: never route a synthetic staging widget into real
production records. If a test origin is needed, approve an exact non-production-only override separately; the
production v1 allowlist remains unchanged. Final navigation proof uses the production synthetic organization.

### Phase 1 exit criteria

- ChatGPT can discover the OAuth server and complete DCR, consent, PKCE, token exchange, refresh, and reconnect against a deployed test environment.
- The Phase 0 candidate issuer has passed real success/error redirect validation; authorization-server identifier and
  issuer then match exactly in the deployed environment.
- New hierarchical grants materialize their implied lower scopes; legacy scope expansion and correctly bound
  audiences pass. Any old cross-host token incompatibility has an explicit reconnect/rollout plan.
- Every public tool has reviewed security schemes and annotations, and all eight structured UI tools advertise schemas that their success responses satisfy.
- Missing-scope calls produce a recoverable in-band OAuth challenge; product-role denials do not request broader scopes.
- Widget deep links are limited to the exact v1 origin `https://deals.anduintransact.com` and work against the
  dedicated UI origin.
- The rendered effective DCR allowlist is recorded for the deployed test environment and retains existing localhost
  and Anthropic callbacks alongside `chatgpt.com`.
- Empty or missing DCR allowlist configuration fails closed and is covered by a regression test.
- Consent and resource-server token audience checks require the same exact trusted MCP scheme, host, effective port,
  and `/mcp` path with no userinfo, query, or fragment; relative and foreign-host `/mcp` values are rejected. Done in
  code; a registered custom domain is trusted only on its default https origin.
- The gateway overwrites `x-anduin-forwarded-host` and `x-forwarded-proto` on every route (checked in and rendered);
  MCP ignores spoofable scheme headers. A real edge request and reachability check confirm the trust boundary.
- A failed public tool call is an `isError=true` result, never a success string. Done in code for the Data Room tools.
- `update_order_tags` and `batch_update_order_tags` advertise `destructiveHint = true`, with contract tests for their
  replacement and clear semantics.
- Existing Claude OAuth and MCP integration tests still pass.
- MCP Inspector reports no blocking protocol or schema failures.

## Phase 2: Make Skills the Canonical Behavior

**Repository:** `anduin-plugin`

**Dependency:** Sections 2.1–2.3 can start before deployed OAuth validation. Section 2.4 is complete only after G4
proves connection resolution; writing a YAML file alone does not close it.

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

### 2.4 Declare each skill's MCP dependency

Add `agents/openai.yaml` beside each canonical `SKILL.md`. Because both skills require the Anduin MCP server, each
file must declare an MCP tool dependency with `type: mcp`, a stable Anduin connection `value`, a user-facing
description, `transport: streamable_http`, and the production endpoint URL. Reconcile the dependency `value` with
the connection exposed by `.app.json` during Phase 3; do not invent an app identifier or include credentials.

Validate the dependency file and its referenced connection as package contracts. Enabling either skill in ChatGPT or
Codex must make the required Anduin MCP tools available rather than relying on an already configured ambient server.
See [Build skills](https://developers.openai.com/plugins/build/skills).

### Phase 2 exit criteria

- ChatGPT, Codex, and Claude receive the same domain and safety behavior from the skills.
- Claude-specific agents contain no independent copy of mutable tool catalogs or authorization rules.
- Both skills contain valid `agents/openai.yaml` files whose MCP dependencies resolve to the intended Anduin
  connection without secrets.
- Existing Claude skill activation and tool access continue to work.

## Phase 3: Add OpenAI Packaging

**Repository:** `anduin-plugin`
**Dependency:** Requires the Phase 0 app identifier and host/package spike results.

### 3.1 Finalize the registered app and discovery snapshot

- Re-register or update the app to point to the deployed production US MCP endpoint.
- Complete OAuth with the dedicated production-US full-scope synthetic review account from Phase 0; do not reuse the
  non-production spike identity.
- Run **Scan Tools** and verify that the draft contains G3's complete approved distribution catalog, titles, descriptions,
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

Map the package to the real registered connection, set its `required` flag to `true`, and test the exact ID form for
each distribution path. The creator guide accepts a `plugin_asdk_app...` URL identifier; workspace import documents
an underlying `asdk_app_...`/supported app ID and explicitly rejects a plugin ID. This discrepancy is an unresolved
host-contract question, not permission to invent or normalize IDs blindly. Record both observed values, validate
the generated file and installation, and use the same connection in skill dependencies. The mapping neither creates
an app nor grants access: admins must enable it and members authenticate. Never include secrets or tokens.
See [packaging](https://developers.openai.com/plugins/build/plugins) and
[workspace app references](https://learn.chatgpt.com/docs/enterprise/plugin-management).

Keep the existing Claude `.mcp.json` source unchanged. Its `mcpServers` wrapper is not the documented native OpenAI
bundled-server shape, and some import paths detect MCP files independently of manifest references. Validate the
actual artifact contents and resulting web/desktop eligibility. If required, generate a separate OpenAI import
artifact without the Claude MCP file. A direct-client `openai.mcp.json` fallback must use the observed supported shape
and be tested separately; it cannot close hosted/web/cloud gates or create duplicate active Anduin connections.

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
- Each `skills/*/agents/openai.yaml` parses, declares the required MCP dependency fields, uses the production
  Streamable HTTP URL, and maps to the `.app.json` connection without embedding credentials.
- The actual release/import artifact contains only the approved files. Test unintended `.mcp.json` discovery,
  app-ID shape, `required: true`, and absence of duplicate app/direct-server bindings. A generated artifact must be
  reproducible from the tag and preserve canonical skill hashes.

Run the validator in CI or the repository’s release workflow.

### Phase 3 exit criteria

- The OpenAI package installs from the local repository marketplace in Codex App, CLI, and IDE, and the supported cloud installation path is verified separately.
- ChatGPT recognizes the registered app and both skills.
- Enabling either OpenAI skill resolves its declared Anduin MCP dependency and exposes the required tools in ChatGPT
  and each supported Codex surface.
- Claude installation still works from the existing marketplace.
- Package validation passes from a clean checkout.

## Phase 4: Cross-platform Verification

Maintain a versioned evaluation matrix in the repository or the release checklist.

### 4.1 Golden prompt set

Include at least five positive cases:

Use only synthetic data and G3-approved tools. If policy excludes a workflow below, replace its positive case with
an approved equivalent and test the exclusion as a negative case; never re-enable a restricted tool just for review.

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
| Codex App | Registered-app package path; separately test direct MCP/loopback | Clean install, skills, OAuth, reconnect, tools, text fallback; label each path's evidence |
| Codex CLI | Supported package/app path must be observed; separately test direct MCP/loopback | Same cases in an isolated profile; no inference from App credentials |
| Codex IDE extension | Supported package/app path must be observed; separately test direct MCP/loopback | Same cases in an isolated profile; no inference from CLI credentials |
| Codex cloud | Hosted installation/connection support and callback are unproven | Prove the supported path and full cases, or obtain an explicit scope decision; local loopback is not evidence |
| Claude Code | Existing Claude marketplace | No regression |
| Cowork | Existing Claude marketplace | No regression |

Do not treat one Codex result as evidence for the other surfaces. Record callback, package-resolution, and OAuth behavior for each. ChatGPT must render and exercise the widgets; Codex acceptance depends on complete text/structured fallbacks and does not assume widget rendering.

Record client versions and whether the catalog is published-snapshot or live-direct. Test with ambient Anduin MCP
configuration disabled and no reusable login from another surface. For each path, measure startup/tool timeout,
cancellation, refresh races, and a deliberately slow operation. Direct Codex documents a 60-second default tool
timeout; do not assume the server's longer timeout fits every host. Use a verified configurable timeout or a
separately designed async workflow if needed, and never automatically retry an ambiguously completed write.
See [Codex MCP configuration](https://learn.chatgpt.com/docs/extend/mcp).

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

### 5.1 OpenAI data-policy and submission gate

- Complete a tool-by-tool classification of inputs, outputs, logs, widgets, and linked resources against the current
  OpenAI plugin guidelines. A generic PII review is insufficient.
- Public submission is blocked for any workflow that can solicit, process, return, or expose Restricted Data. In
  particular, W-9/W-8BEN, AML/KYC, arbitrary forms, document OCR, previews, downloads, and search results must have
  server-enforced exclusion or redaction before the data enters the plugin/model boundary for government identifiers,
  payment-card data, health information, and credentials. Output-only redaction, prompt text, and model discretion are
  not enforcement controls.
- For other regulated sensitive data, record strict necessity, legally adequate consent, and explicit prominent
  disclosure before exposure through the plugin. Minimize returned data and avoid placing sensitive values in logs,
  errors, widgets, or model-visible metadata.
- Approve a distribution-specific tool inventory **before** the final scan. It may be narrower than the 83-tool
  baseline, but exclusions must be enforced at execution and data access, not only hidden from discovery or described
  in prompts. Classify generic search/OCR/download/render paths as well as named tax/KYC tools. Any per-distribution
  distinction needs a trusted server-side registration/endpoint policy, not spoofable host labels; design and test
  it as separate work without changing existing Claude access accidentally.
- Record the approved distribution mode. If public controls cannot be proven, seek approval for a workspace-private
  pilot; do not silently redefine the public milestone. Privacy and security obligations still apply.

- Complete organization and business verification.
- Use an OpenAI project eligible for an MCP-backed public submission.
- Ensure the submitter has the required Apps Management permissions.
- Provide public privacy policy, terms, support, and homepage URLs.
- Provide a stable production endpoint and confirm that the dedicated production-US synthetic full-scope review
  account established in Phase 0 has a security-approved review-only authentication arrangement. Keep the
  non-production spike account separate and never relax global MFA controls.
- Provide public-facing disclosures that accurately describe every data category and server-side control in scope.

### 5.2 Submit one universal listing

- Submit the integration as a plugin with MCP.
- Create a new draft, scan the production MCP server with the dedicated production-US synthetic full-scope review
  account, and resolve all metadata, schema, annotation, and UI findings before submission.
- Compare the stored scan snapshot against G3's approved distribution contract, not blindly all 83 tools: inventory, titles,
  descriptions, input/output schemas, security schemes, annotations, tool `_meta`, server `instructions`, UI resource
  URIs, and linked UI metadata/CSP. Fail the release if anything is absent, unintended, or stale.
- Attach the exact two canonical skill revisions evaluated for this release.
- Provide starter prompts, at least five positive tests, and at least three negative tests.
- Supply accurate release notes and supported-platform claims.
- Publish one listing for ChatGPT and Codex.

For every subsequent release, repeat draft creation, **Scan Tools**, complete-contract snapshot comparison, skill
attachment, evaluations, and submission whenever any captured field or either skill changes. Changes to the MCP
scheme, hostname, or port require a new plugin. Breaking tool, schema, or UI-resource changes must first ship
additively and remain compatible with every active published snapshot. Current review rules do not support breaking
updates; approval of a replacement or the end of a rollback window alone does not authorize removal. Retire contracts
only through a supported migration with evidence no supported consumer needs them. Compatible HTML changes at an
unchanged URI may be cached for up to an hour; a changed content-hash URI is a captured-contract change requiring a
rescan. Do not assume a live deployment updates published discovery.
See [MCP server review requirements](https://developers.openai.com/plugins/deploy/app-review).

### 5.3 Rollout sequence

1. Merge G1 only after required CI/review; deploy the candidate to selected non-production with recorded config.
2. Close G2 and verify Claude, including existing token audiences; approve G3 before production/public exposure.
3. Promote the validated configuration and test the unpublished registration with authorized synthetic users.
4. Release the multi-host plugin package and synchronized version.
5. Capture and approve the production tool/skill snapshot, then submit or enable the OpenAI listing for a limited cohort where possible.
6. Expand availability after observing stable authentication, tool selection, and error rates.

### 5.4 Observability

Track, without recording sensitive values:

- DCR failures by client and reason.
- OAuth discovery, authorization, token, refresh, and reconnect failures.
- Insufficient-scope and product-authorization failures.
- Known-but-scope-hidden tool calls separately from truly unknown tool names.
- Tool invocation success, clean errors, and unexpected errors; host/client labels are diagnostic only, may be
  missing or spoofed, and must not drive authorization or distribution policy.
- Tool-selection misses and unsupported-request false positives from evaluation or feedback.
- Widget render failures and fallback usage.
- Confirmation cancellation rates for write and destructive tools.

Before canary, assign dashboard/on-call owners, record baseline rates and explicit stop/expand thresholds, and
rehearse the pause procedure. Do not claim instrumentation exists merely because metrics are listed here. Avoid
response-body logging and redact OAuth secrets; assess whether model-visible correlation IDs are strictly necessary
under response-minimization policy before public review.

### 5.5 Rollback

- Preserve supported contracts and independently pause OpenAI availability. Do not roll back security fixes such as
  exact audience enforcement to avoid reconnecting legacy sessions.
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

Deploy and verify this PR before publishing/enabling the OpenAI package. Canonical skill-only changes may merge
earlier with Claude regression evidence. `McpUi.version` already pins released `0.5.0`.

### PR 2a: `anduin-plugin` — Canonical skills (can start before G2)

- Consolidated provider-neutral skills.
- Thin Claude agent adapters.
- Golden prompts, confirmation/partial-failure cases, and Claude regression evidence.

### PR 2b: `anduin-plugin` — Multi-host package (gated on observed connection behavior)

- `.codex-plugin/plugin.json`.
- `.app.json` using the real registered app identifier.
- Per-skill `agents/openai.yaml` MCP dependency declarations, reconciled with the app connection.
- Proven app mapping and conditional direct-client packaging; preserve Claude source and prevent import collisions.
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
| Empty DCR configuration silently disables redirect validation | Make empty/missing configuration fail closed and test it at the configuration-to-validator boundary |
| Path-only audience checks accept tokens for a foreign or relative MCP resource | Canonicalize and compare the exact trusted scheme, host, effective port, and path in consent and resource-server validation |
| Scan Tools captures an incomplete scope-filtered contract | Scan with the dedicated production full-scope account and compare every captured tool, instruction, and linked UI field before each submission |
| Published metadata or skills become stale after a live change | Treat the published snapshot as authoritative in both ChatGPT and Codex; require a new draft, complete rescan or skill reattachment, evaluation, and resubmission as part of the release checklist |
| Incorrect tool annotations suppress warnings or cause excessive confirmations | Maintain explicit reviewed classifications and fail tests for missing metadata |
| Restricted or regulated data makes the plugin ineligible or unsafe for public distribution | Classify every tool and response, enforce exclusions or pre-boundary redaction server-side, and keep public submission blocked until the policy gate passes |
| Structured output drifts from advertised schemas | Own schemas next to SDK validators and validate every UI tool's success output in `stargazer` contract tests |
| Widget links escape the intended Anduin origin | Allow only `https://deals.anduintransact.com` in v1 CSP and render fallback, customer, non-US, and arbitrary URLs as inert text |
| Skills and Claude agents drift | Make skills canonical and reduce agents to thin host adapters with no duplicated mutable policy |
| A skill installs without its required MCP connection | Declare and validate the Anduin dependency in each skill's `agents/openai.yaml` and test clean-install resolution |
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
- New admin/write grants expose their complete hierarchical scope closure; correctly audience-bound old tokens
  remain compatible, and any necessary legacy reconnect is documented. Descriptors/challenges name minimal scopes.
- Every public MCP tool has correct security schemes and reviewed annotations; all eight UI-producing tools have sound, passing output schemas.
- The published OpenAI snapshot used by ChatGPT and Codex contains the complete approved distribution catalog, server
  instructions, and linked UI contract, while live direct-client discovery and execution authorization match approved
  scopes and Anduin product roles.
- Known-but-scope-hidden calls return a recoverable in-band OAuth challenge; product-role denials do not ask for broader OAuth scopes.
- Destructive actions retain appropriate confirmation and admin gates.
- Tag replacement/clear tools are classified as destructive and retain confirmation behavior.
- The selected distribution mode and per-tool OpenAI data-policy review are recorded; public availability is blocked
  unless Restricted Data exclusions/pre-boundary redactions and regulated-data requirements are enforced and proven.
- Each OpenAI skill declares and resolves its required Anduin MCP dependency from a clean install.
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
- [Build skills](https://developers.openai.com/plugins/build/skills)
- [Package your plugin](https://developers.openai.com/plugins/build/plugins)
- [Workspace plugin import and app references](https://learn.chatgpt.com/docs/enterprise/plugin-management)
- [Authenticate users](https://developers.openai.com/plugins/build/auth)
- [Plugin and MCP Apps reference](https://developers.openai.com/plugins/reference)
- [Build ChatGPT UI](https://developers.openai.com/plugins/build/chatgpt-ui)
- [Connect and test your plugin](https://developers.openai.com/plugins/deploy/connect-chatgpt)
- [Submit and publish](https://developers.openai.com/plugins/deploy/submission)
- [MCP server review requirements](https://developers.openai.com/plugins/deploy/app-review)
- [Plugin guidelines](https://developers.openai.com/plugins/app-guidelines)
- [Connect Codex to MCP servers](https://learn.chatgpt.com/docs/extend/mcp)
- [Submit a Claude Code plugin](https://developers.openai.com/plugins/guides/submit-claude-plugin)
- [Optimize tool metadata](https://developers.openai.com/plugins/guides/optimize-metadata)
