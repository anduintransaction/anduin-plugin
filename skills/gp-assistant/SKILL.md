---
name: Anduin GP Assistant
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
- **investor group** (synonyms: LP group) — grouping of LPs within a fund
- **subscription document** (synonyms: sub doc, subscription agreement) — formal legal agreement to subscribe
- **side letter** — separate agreement granting investor-specific terms
- **supporting document** (synonyms: tax form, KYC document, W-9, W-8BEN) — documents LP uploads for compliance
- **fund manager** (synonyms: GP, general partner, manager) — firm managing the fund
- **compliance check** (synonyms: AML, KYC, background check) — anti-money laundering/know-your-customer verification
- **order** (synonyms: LP, subscription, investor) — an LP's subscription order, identified by order_id
- **form** (synonyms: subscription form, investor questionnaire) — the subscription form filled by an LP

## MCP Tools by Category

All tools require OAuth2 scope `fundsub:read` or `fundsub:write`.

### Fund & Order Discovery (fundsub:read)
- `list_funds` — list accessible funds with IDs
- `get_fund_info` — detailed fund information (closes, sub-funds, entity info)
- `list_orders` — list LP orders in a fund
- `get_order_workflow_data` — detailed order status (tags, contacts, commitments, metadata)
- `get_order_submission_data` — order submission details
- `get_order_subscription_docs` — subscription documents by stage
- `get_file_download_url` — pre-signed download URL for a file
- `get_standard_form_fields` — standard form field definitions
- `get_invitation_link` — self-signup invitation link for a fund

### LP Status & Review (fundsub:read)
- `get_lp_status` — LP subscription status and details
- `get_supporting_docs` — LP supporting/compliance documents
- `get_required_docs` — required documents checklist
- `get_form_markdown` — form content rendered as markdown
- `get_form_comments` — comments on form fields
- `draft_comment` — draft a comment on a form field (fundsub:write)
- `get_aml_check` — AML check results
- `get_aml_kyc_doc_groups` — AML/KYC document group configuration

### Form Interaction (fundsub:read / fundsub:write)
- `get_form_schema` — form structure and field definitions (fundsub:read)
- `get_form_progress` — form completion progress (fundsub:read)
- `get_form_field_aliases` — field alias mappings (fundsub:read)
- `get_next_fields_to_fill` — prioritized fields: required, docs, recommended (fundsub:read)
- `get_form_field_value` — specific field value (fundsub:read)
- `get_form_validation_errors` — validation errors (fundsub:read)
- `update_form_fields` — update field values (fundsub:write)

### Cross-Order Analysis (fundsub:read)
- `compare_form_fields` — compare a field value across multiple orders
- `search_orders_by_field` — search orders by field value

### Activity Log (fundsub:read)
- `get_order_activity_log` — LP order activity history with filtering
- `get_fund_activity_log` — fund-level admin activity log

### Dashboard & Reporting (fundsub:read / fundsub:write)
- `query_dashboard` — dashboard with search, filter, sort, pagination (fundsub:read)
- `get_fund_report` — fund subscription report (fundsub:read)
- `get_my_fund_permissions` — current user's role and permissions (fundsub:read)
- `list_fund_members` — fund team members and roles (fundsub:read)
- `update_order_tags` — update tags on an order (fundsub:write)
- `batch_update_order_tags` — batch tag update across multiple orders (fundsub:write)
- `update_order_custom_data` — update custom columns on an order (fundsub:write)

### Fund Manager Invitation (fundsub:read / fundsub:write)
- `list_fund_manager_groups` — list groups the user can invite into (fundsub:read)
- `validate_fund_manager_emails` — check email membership status (fundsub:read)
- `invite_fund_managers` — send invitations to fund managers (fundsub:write)

## Tool Chaining Rules

IDs flow between tools in a strict order. ALWAYS copy IDs exactly — never fabricate, shorten, or modify.

```
Step 1: list_funds → fund_id
Step 2: query_dashboard → order_id + status/entity/contact/activity (preferred for browsing)
        OR list_orders → order_id (lightweight ID discovery)
Step 3: get_order_workflow_data → detailed order info (tags, contacts, commitments)
Step 4: get_order_subscription_docs → file_id → get_file_download_url
Step 5: get_form_schema → field_alias → get_form_field_value, update_form_fields
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
