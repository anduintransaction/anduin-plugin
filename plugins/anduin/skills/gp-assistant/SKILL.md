---
name: gp-assistant
description: Use when the user asks about fund subscriptions, LP review, fund administration, investor onboarding, subscription forms, AML/KYC checks, fund manager invitations, fund reports, order dashboards, or any GP (General Partner) fund management task. Provides terminology, tool chaining rules, and workflow patterns for Anduin fund subscription operations via MCP.
---

# Anduin GP Assistant Domain Knowledge

## Terminology

- **fund** (synonyms: vehicle, fund vehicle, investment fund) — identified by fund_id
- **close** (synonyms: closing, first close, final close) — a fundraising round within a fund
- **commitment** (synonyms: allocation, capital commitment, pledge) — total amount an LP pledges to invest
- **capital call** (synonyms: drawdown, funding notice) — fund manager request to transfer committed capital
- **investment entity** (synonyms: legal entity, subscribing entity) — legal entity through which an LP invests
- **sub-fund** (synonyms: subfund, feeder fund, parallel fund) — structural subdivision of a fund
- **investor group** (synonyms: LP group) — grouping of LPs within a fund (distinct from a fund manager group)
- **fund manager group** (synonyms: GP group, manager group) — a GP team grouping with fund permissions; role types are Admin / Custom, with four default groups (Fund managers / Fund counsel / Fund admins / Anduin support)
- **subscription document** (synonyms: sub doc, subscription agreement) — formal legal agreement to subscribe
- **side letter** — separate agreement granting investor-specific terms
- **supporting document** (synonyms: tax form, KYC document, W-9, W-8BEN) — documents LP uploads for compliance
- **fund manager** (synonyms: GP, general partner, manager) — firm managing the fund
- **compliance check** (synonyms: AML, KYC, background check) — anti-money laundering/know-your-customer verification
- **order** (synonyms: LP, subscription, investor) — an LP's subscription order, identified by order_id
- **form** (synonyms: subscription form, investor questionnaire) — the subscription form filled by an LP

When a user says "group" unqualified, both meanings are plausible — ask whether they mean an **investor group** (LP grouping) or a **fund manager group** (GP team with permissions) before acting.

## Flow Gating

Both **Flexible-flow** and **Restricted-flow** funds are reachable via MCP — visibility is gated by the tool allowlist + OAuth scope, NOT by fund flow. `get_fund_review_config.reviewers` semantics depend on flow type:
- **Flexible flow** — `reviewers` is empty by design (reviewer identities live per review step). Use `list_fund_members` to find potential reviewers.
- **Restricted flow** — `reviewers` is populated with the Admin-group members from the legacy review package when configured (may still be empty if none assigned).
- The `Form filled` status appears only on Restricted funds. Do not filter for it on Flexible funds.

## LP Status Lifecycle

Two review branches exist, and they lead to different terminal states:

```
LPNotStarted → LPInProgress
  ├─ UNSIGNED-review branch: → LPPendingUnsignedReview → LPFormReviewed → LPRequestedSignature → LPSignedForm
  └─ SIGNED-review branch:   → LPRequestedSignature → LPSignedForm → LPPendingReview → LPSubmitted
                                                                                       ↓
                                                                 (optional) LPPendingSubmission → LPSubmitted
                                                                                       ↓
                                                                                  LPCountersigned → LPCompleted
```

Key points:
- SIGNED-review approval transitions directly to `LPSubmitted`. It does NOT pass through `LPFormReviewed` and does NOT countersign. Countersigning is a separate subsequent action on top of `LPSubmitted`.
- UNSIGNED-review approval transitions to `LPFormReviewed` (pre-signature).
- `LPPendingSubmission` is gated by the `enableLpManualSubmitSubscription` fund switch.
- `LPPendingUnsignedReview` / `LPPendingReview` are gated by review-package flags.

## Status Enum Names for `query_dashboard`

`query_dashboard` filters accept **enum names**, not UI labels. The full closed set of 14 valid values is:
`LPNotStarted`, `LPInProgress`, `LPChangeInProgress`, `LPFilledForm`, `LPPendingUnsignedReview`, `LPRequestedSignature`, `LPSignedForm`, `LPPendingSubmission`, `LPPendingReview`, `LPFormReviewed`, `LPSubmitted`, `LPCountersigned`, `LPCompleted`, `LPRemoved`.

Invalid values are silently ignored (no filter applied), so spelling must be exact. Never pass UI labels like `"Pending review"` or `"Pending approval"`.

## Sort Fields

`query_dashboard` accepts three `sort_by` keys: `status`, `contactName` (orders by investor name — investment entity, else contact name), and `lastActiveAt` (most-recent activity). For activity, the key is `lastActiveAt`, NOT `lastActivityAt`.

## Subscription Agreement vs Form vs Supporting Docs

- **Subscription form** (`get_form_markdown` / `get_form_schema`) — the live form the LP fills. This IS the subscription agreement's content for verification purposes.
- **Supporting documents** (`get_supporting_docs`) — AML/KYC, tax forms (W-9, W-8BEN/W-8BEN-E), formation documents, uploaded by the LP.
- **Subscription documents** (`get_order_subscription_docs`) — the generated/signed subscription booklet PDF. Downloadable via `get_file_download_url`, but its CONTENT is the form — verification questions should read the form, not OCR the PDF.

When a user says "subscription agreement," route to the form unless they explicitly want the signed PDF artifact.

## Activity-Log Identity Rules

`get_order_activity_log` returns `actorName` and `actorRole` for events, so you CAN identify who performed an action (e.g., who approved an unsigned review). What you CANNOT identify from the activity log is the **ASSIGNED reviewer** for a multi-step review stage — assignment metadata is not in the log.

## Search Scan-Limit

`search_orders_by_field` truncates by **scanning the first 100 orders**, not by limiting the result list. When truncated, warn the user that matches beyond the first 100 scanned orders may have been missed.

## MCP Tools by Category

All tools require OAuth2 scope `fundsub:read` or `fundsub:write`.

OAuth scope is necessary but not sufficient — some tools additionally require granular fund-manager permissions enforced at the service layer. `get_fund_feature_switches`, `get_fund_review_config`, `get_fund_report`, and `aggregate_orders` require the `AccessFundReporting` fund permission (ManageFundSetting alone does not grant the two config tools). `list_fund_members` and `list_fund_manager_groups` are role-gated: Admin-role users see the whole team / all groups; Custom-role users see only the groups they have permission on. Use `get_my_fund_permissions` to check the current user's permissions. On a permission denial, point the user to a fund admin rather than treating it as an error or bug.

### Fund & Order Discovery (fundsub:read)
- `list_funds` — list accessible funds with IDs
- `get_fund_info` — detailed fund information (closes, sub-funds, entity info)
- `list_orders` — list LP orders in a fund (limit default 50, max 100; offset default 0)
- `get_order_workflow_data` — detailed order status (tags, contacts, commitments, metadata)
- `get_order_submission_data` (order_id) — transformed form submission data grouped by namespace; a "No submission data available/found" result is returned both when the order is unsubmitted/unfilled (a caught internal error, not surfaced) and when a successful response has no fields — treat it as possibly not-yet-submitted rather than confirmed-absent, and cross-check via `get_lp_status`
- `get_order_subscription_docs` — subscription documents by stage
- `get_file_download_url` — pre-signed download URL for a file
- `get_standard_form_fields` — standard form field definitions
- `get_invitation_link` — self-signup invitation link for a fund

### LP Status & Review (fundsub:read)
- `get_lp_status` — LP subscription status and details
- `get_supporting_docs` (order_id) — requested supporting/AML-KYC docs with status (Uploaded / Not Applicable / Marked as Provided / Pending) and a `[Requested by Admin]` flag, a tax-form submissions section, plus file IDs (the OCR-via-`convert_document_to_markdown` flow is covered in the Document Reading Workflow)
- `get_required_docs` (fund_id) — lists the TAX FORMS configured for the fund (e.g. W-9, W-8BEN); despite the name returns ONLY tax-form types, NOT general supporting docs (use `get_supporting_docs` for per-LP requested supporting documents); returns "No tax forms configured" when empty
- `get_form_markdown` — form content rendered as markdown
- `get_form_comments` — all comments on the order's form; `include_internal` (default false) controls whether GP/admin-only internal comments are returned
- `draft_comment` — create a public (default) or internal comment; field-level when `field_alias` is set, OR order-level/general when `field_alias` is omitted (do not pass a placeholder) (fundsub:write)
- `get_aml_check` (order_id) — AML/KYC check results (status, provider, investor type, entity role); access is enforced at the service layer (ReBAC view check) — unlike sibling order tools it does not run the tool-layer conversation-scope confinement check
- `get_aml_kyc_doc_groups` — AML/KYC document group configuration

### Form Interaction (fundsub:read / fundsub:write)
- `get_form_schema` — form structure and field definitions (fundsub:read)
- `get_form_progress` — form completion progress (fundsub:read)
- `get_form_field_aliases` — field alias mappings (fundsub:read)
- `get_next_fields_to_fill` — prioritized fields: required, docs, recommended (fundsub:read)
- `get_form_field_value` (order_id, field_alias) — read a specific field value by alias (fundsub:read)
- `get_form_validation_errors` — validation errors (fundsub:read)
- `update_form_fields` — update field values (fundsub:write)

### Cross-Order Analysis (fundsub:read)
- `compare_form_fields` — compare a field value across multiple orders (order_ids required, max 20 — graceful error over the limit). Returns PARTIAL results: orders that can't be read (bad/out-of-scope ID, no access) are skipped and listed under a "Could not compare N order(s)" section rather than failing the whole call — surface those skipped orders instead of assuming every requested order was compared
- `search_orders_by_field` — search orders by field value (limit default 20, max 100)

### Activity Log (fundsub:read)
- `get_order_activity_log` (order_id; optional offset, limit [default 50, max 100], category, only_unseen) — LP order activity history, newest-first; offset walks backward in time; category filter (invitation/form/document/review/signature/comment/email/entity/other — an unknown category is rejected with an error, NOT silently ignored) and only_unseen supported
- `get_fund_activity_log` (fund_id; optional offset, limit [default 50, max 100]) — fund-level admin activity log, newest-first; offset walks backward; NO category filter and NO only_unseen

### Dashboard & Reporting (fundsub:read / fundsub:write)
- `query_dashboard` — dashboard with search, filter, sort, pagination (page_size default 50, max 100) (fundsub:read)
- `get_fund_report` — fund subscription report (fundsub:read)
- `get_my_fund_permissions` — current user's role and permissions (fundsub:read)
- `list_fund_members` — fund team members and roles (fundsub:read)
- `update_order_tags` — REPLACE all tags on an order with the provided list (empty array clears; read current tags via `get_order_workflow_data` first) (fundsub:write)
- `batch_update_order_tags` — batch REPLACE tags across multiple orders (max 200 items) (fundsub:write)
- `update_order_custom_data` — MERGE custom columns on an order (only specified columns change; max 20 columns; discover columns/allowed values via `get_fund_info`; metadata columns are read-only) (fundsub:write)

### Fund Configuration & Aggregation (fundsub:read)
- `aggregate_orders` — server-computed deduped order counts grouped by one or two of {status, orderType, close} (an unsupported dimension, or more than two, is rejected with a clean error — not silently dropped), with optional `status_filter` and `order_type_filter` (Online/Offline); counts are fund-wide, deduped by investor; `group_by` defaults to `[status, orderType]`; `order_type_filter` uses the {Online, Offline} vocabulary (distinct from the status enum). Prefer `aggregate_orders` (or `get_fund_report`'s cross-tab) over hand-counting `query_dashboard` rows
- `get_fund_feature_switches` (fund_id) — curated GP-visible feature-family config (review workflow, unsigned review, multi-step review, manual submission, side letter, AML/KYC, supporting-doc review, form lock); each family returns Enabled/Disabled + product meaning; internal rollout flags excluded. Call BEFORE answering "is X enabled for this fund?". The `review_workflow` family always agrees with `get_fund_review_config`'s "Signed review enabled" (same underlying configuration). Requires the `AccessFundReporting` permission (ManageFundSetting alone is insufficient)
- `get_fund_review_config` (fund_id) — fund's signed and unsigned subscription review configuration (whether review enabled, whether unsigned review enabled, reviewer identities). Call BEFORE answering "is review enabled?" / "who reviews orders here?". For UNSIGNED review gating prefer the `unsigned_review` family in `get_fund_feature_switches`: it includes the submit-before-signing enforcement switch, which `isUnsignedReviewEnabled` here reports as configuration only. v1 limitation: multi-step review stages and supporting-doc review config are NOT exposed; requires `AccessFundReporting`

### Fund Manager Invitation (fundsub:read / fundsub:write)
- `list_fund_manager_groups` — list groups the user can invite into (fundsub:read)
- `validate_fund_manager_emails` — check email membership status (fundsub:read)
- `invite_fund_managers` — send invitations to fund managers (fundsub:write)

### Document Processing (fundsub:read)
- `convert_document_to_markdown` — convert an uploaded document (PDF, JPEG, PNG, GIF, WebP) to markdown using OCR. For large documents (50+ pages — the trigger is size/character-based), returns a page index instead of full content
- `read_document_pages` — read specific page ranges from a previously-converted large document. Pages are 1-indexed, max 30 pages per call
- `convert_spreadsheet_to_markdown` — convert an uploaded Excel spreadsheet (XLS, XLSX) to markdown tables (one section per sheet). Do NOT use `convert_document_to_markdown` for Excel files. For large spreadsheets returns a sheet index — use `read_spreadsheet_sheet`
- `read_spreadsheet_sheet` — read a specific sheet from a previously-converted spreadsheet by index

## UI Rendering (mcp:render scope)

Three **display-only** render tools turn structured data into interactive `ui://` widgets (MCP Apps). They render as sandboxed iframes in UI-capable hosts (Claude Code, Cowork); in text-only / headless contexts they are not shown, so ALWAYS also summarize the data in markdown.

**Availability is environment-dependent — rely on your live tool list, never assume.** These tools exist only on Anduin servers that have shipped UI rendering (rolled out per environment — local/staging ahead of production) AND only when your grant includes the **`mcp:render`** OAuth scope (independent of `fundsub:*`). Your available tools are the source of truth: before offering a rendered view, confirm the specific render tool is actually present; if it is not, your server/environment simply hasn't enabled it yet — quietly fall back to a markdown table/list (don't announce a missing tool unless asked).

These tools are a **presentation layer only**: values are shown for viewing and CANNOT be edited or sent back (`interactive: false`). Never use `render_ui` to *collect* input — use the form-filling tools (`update_form_fields`) for that.

| Tool | Renders (`ui://`) | Key inputs |
|------|------|-----------|
| `render_chart` | ECharts chart (`ui://anduin/chart`) | `title` (req), `echarts_option` (req — ECharts option object: series/xAxis/yAxis/tooltip/legend), `width` (opt, 200–1200, default 600), `height` (opt, 150–800, default 400) |
| `render_table` | Data table (`ui://anduin/table`) | `title` (req), `columns` (req — `[{id, label, type?: text\|number\|tag-list\|badge\|link}]`), `rows` (req — `[{<column id>: value, …, id}]`; tag-list cells are JSON string arrays) |
| `render_ui` | Form-layout view (`ui://anduin/form`) | `component: "form"` (req), `title` (req), `description` (opt), `sections` (req — `[{title, fields:[{alias, label, type: text\|number\|select\|checkbox\|date\|textarea, value, required?, options?}]}]`) |

Limits (over-limit/malformed input returns `{ "error": ... }` — fall back to markdown): chart `echarts_option` ≤100 KB; table ≤20 columns / ≤200 rows; form ≤20 sections / ≤50 fields. Each column `id` and each field `alias` must be unique, and every `type` must be one of the values listed above — duplicate ids/aliases or an unrecognized/non-string `type` are rejected.

**When to render (vs. plain markdown):** render when a visual genuinely helps — a bar/pie chart of commitments by close, a sortable table of LP orders, a form-style snapshot of an order's key fields. Prefer plain markdown for a single fact or a short list, and when running headless. Build the data with the read tools FIRST, then pass it to a render tool, and STILL give a one-line text summary so non-UI clients stay functional.

- "Chart commitments by close" → `aggregate_orders(group_by=[close])` → `render_chart` (bar/pie)
- "Show the orders as a table" → `query_dashboard` → `render_table` (entity / status / commitment / tags columns)
- "Summarize this order's key fields" → `get_order_submission_data` → `render_ui` (form sections, display-only)

## Tool Chaining Rules

IDs flow between tools in a strict order. ALWAYS copy IDs exactly — never fabricate, shorten, or modify.

```
Step 1: list_funds → fund_id
Step 2: query_dashboard → order_id + status/entity/contact/activity (preferred for browsing)
        OR list_orders → order_id (lightweight ID discovery)
Step 3: get_order_workflow_data → detailed order info (tags, contacts, commitments)
Step 4: get_order_subscription_docs → file_id → get_file_download_url
Step 5: get_form_schema → field_alias → get_form_field_value, update_form_fields
Step 6: get_order_subscription_docs → file_id → convert_document_to_markdown OR convert_spreadsheet_to_markdown
Step 7: (large doc) convert_document_to_markdown → page index → read_document_pages(start_page, end_page)
Step 8: (multi-sheet) convert_spreadsheet_to_markdown → sheet_index → read_spreadsheet_sheet
```

### Critical ID Rules
1. IDs are opaque strings with internal structure — CANNOT construct or guess them
2. MUST obtain IDs from tool outputs only
3. ALWAYS call `list_orders` or `query_dashboard` FIRST before any tool requiring order_id
4. Copy IDs exactly as they appear — never modify, truncate, or combine parts
5. If "Invalid ID" error: you fabricated the ID — call the appropriate discovery tool

## Workflows

### GP Review Workflow
1. Discover fund and orders (`list_funds`, `get_fund_info`)
2. `get_lp_status` — understand LP state and form progress
3. `get_form_schema` + `get_form_markdown` — review form content
4. `get_form_validation_errors` — identify incomplete fields
5. `get_supporting_docs` + `get_order_subscription_docs` — check documents
6. `get_form_comments` — review existing discussions
7. `draft_comment` — flag issues or provide feedback
8. `compare_form_fields` or `search_orders_by_field` — cross-order analysis

### Fund Manager Invitation Workflow
1. `list_fund_manager_groups` — discover available groups
2. Collect email addresses from user
3. `validate_fund_manager_emails` — check who's already a member
4. Present preview for user confirmation
5. `invite_fund_managers` — send invitations after approval

### Batch Tagging Protocol
1. Gather data: `list_orders` for IDs, then `get_order_workflow_data` for current tags
2. Present preview as markdown table: LP Name | Current Tags | Proposed Tags
3. STOP and wait for user confirmation
4. Call `batch_update_order_tags` with confirmed items
5. NEVER call batch operations without user preview unless they say "skip preview"

### Form Filling Protocol (if assisting LP)
1. `get_form_schema` — learn all field aliases, types, enum values
2. `get_form_field_aliases` — map standard names to form-specific aliases
3. `get_next_fields_to_fill` — get prioritized field list
4. Before updating: verify alias exists in schema, field is not hidden/disabled
5. For enums: value MUST exactly match enumValues
6. For multi-select: value MUST be JSON array `["option1", "option2"]`
7. Call `update_form_fields` to update
8. After update: check `cascadingChanges` response, call `get_form_validation_errors`
9. Report progress percentage from `updatedProgressPercentage`

### Anti-Hallucination Rules for Form Updates
- Reading schema or knowing values DOES NOT equal updating them
- No tool call = no update — NEVER tell user fields were updated without calling `update_form_fields`
- NEVER call `update_form_fields` without first calling `get_form_schema` in the session
- Use field `alias` from `get_form_schema` — never use labels or guessed names
