---
name: dataroom
description: Use when the user asks about virtual data rooms, VDRs, deal rooms, document sharing, data room participants, file management in data rooms, data room analytics or insights. Provides terminology, tool chaining rules, and workflow patterns for Anduin Data Room operations via MCP.
---

# Anduin Data Room Domain Knowledge

## Terminology

- **data room** (synonyms: VDR, virtual data room, deal room) — a secure online repository for sharing and managing documents
- **participant** (synonyms: member, collaborator, user) — a person with access to a data room
- **entity** (synonyms: organization, company, firm, org) — a business organization registered on the platform, identified by `entity_id`

## Participant Roles

| Role | Can Do | Cannot Do |
|------|--------|-----------|
| **Admin** | All operations (view, create, upload, delete, manage participants, archive) | None (full access) |
| **Member** | View, search, create folders, upload/rename/delete files, invite, view insights | Remove participants, modify permissions, rename/archive data room |
| **Contributor** (internally Guest) | View, search, create folders, upload/rename/delete files, view insights | Remove participants, modify permissions, rename/archive data room |
| **Observer** (internally Restricted) | View (read-only), search | Create, upload, delete, invite, manage, archive |

Roles surface to users as Admin / Member / Contributor / Observer. Call `dr_check_my_permissions` for the authoritative, per-user set of allowed/disallowed actions (Contributor invite capability, for example, is a conditional per-user flag). The "view insights" capability shown for Member/Contributor mirrors `dr_check_my_permissions`, but the analytics **tools** (`dr_get_insights`/`dr_get_timeline`/`dr_get_activity_log`/`dr_get_dataroom_summary`) are gated to **Admins on a premium Insights plan** — see Analytics & Insights below.

## MCP Tools by Category

All tools require OAuth2 scope `dataroom:read` or `dataroom:write`. Tools are prefixed with `dr_` in the MCP server.

### Entity & Data Room Discovery (dataroom:read)
- `dr_list_entities` — list all entities the user belongs to, with subscription plans and data room counts
- `dr_list_datarooms` — list all data rooms accessible to the user
- `dr_get_dataroom_detail` — get detailed information about a specific data room

### Data Room Operations (dataroom:write)
- `dr_create_dataroom` — create a new data room; `name` (req), `entity_id` (optional, pattern `^ent[a-z0-9]{13}$`) — auto-resolves for single-entity users, so only call `dr_list_entities` first when the user has multiple entities
- `dr_rename_dataroom` — rename an existing data room
- `dr_archive_dataroom` — archive or unarchive a data room

### Participant Management
- `dr_list_participants` — list all participants with roles and status (dataroom:read)
- `dr_invite_users` — invite users by email; `dataroom_id` (req), `emails` (req, array, ≥1), `role` (optional — admin / member / guest (=Contributor) / observer (=Observer); defaults to Contributor (internally Guest) when omitted) (dataroom:write)
- `dr_remove_users` — remove users from a data room (dataroom:write)
- `dr_modify_user_permissions` — change a user's role; `dataroom_id`, `user_id` (from dr_list_participants), `role` (admin / member / guest (=Contributor) / observer (=Observer)) (dataroom:write)
- `dr_check_my_permissions` — check your current role and the actions you can/cannot perform (dataroom:read)
- `dr_list_groups` — list user groups in a data room (dataroom:read)

### File & Folder Management
- `dr_list_files` — list files and folders in a directory (dataroom:read)
- `dr_search` — search for files and folders by name (dataroom:read)
- `dr_get_file_download_url` — get a temporary presigned download URL for a data room file (expires after 15 minutes); takes `file_id` (from dr_list_files or dr_search), NOT dataroom_id (dataroom:read)
- `dr_create_folder` — create a new folder (dataroom:write)
- `dr_rename_item` — rename a file or folder (dataroom:write)
- `dr_delete_items` — delete files and/or folders (moves to trash); `dataroom_id` (req), `file_ids` (array, optional), `folder_ids` (array, optional) — at least one of the two must be non-empty (dataroom:write)
- `dr_restore_items` — restore previously deleted FILES; `dataroom_id` (req), `file_ids` (array). Folder restoration is NOT supported — passing `folder_ids` errors the entire call (dataroom:write)

### Document Processing (dataroom:read)
- `dr_convert_document_to_markdown` — convert an uploaded document (PDF, JPEG, PNG, GIF, WebP) to markdown using OCR; takes `file_id` plus `force_regenerate` (default false; conversions are cached). The large-doc trigger is size/character-based (~50,000 chars, often 50+ pages): for large documents it returns a page index instead of full content. The file must belong to the scoped data room
- `dr_read_document_pages` — read specific page ranges from a previously-converted large document; `file_id` (req), `start_page` (default 1), `end_page` (default start_page+29, max 30 pages per call). Pages are 1-indexed; over-budget output returns continuation hints
- `dr_convert_spreadsheet_to_markdown` — convert an uploaded Excel spreadsheet (XLS, XLSX) to markdown; each sheet becomes a markdown section headed by the sheet name. For large spreadsheets returns a sheet index — use `dr_read_spreadsheet_sheet`
- `dr_read_spreadsheet_sheet` — read a specific sheet from a previously-converted spreadsheet; `file_id` (req), `sheet_index` (optional, default 1), `start_row`/`end_row` (optional, 1-based positional, default 1/last). Rows are positional (Nth non-empty row), not spreadsheet line numbers; over-budget output returns fewer rows with a continuation hint

### Analytics & Insights (dataroom:read)

> **Access gate:** all four analytics tools below additionally require the caller to be a **joined Admin** on a data room whose plan includes the **Insights** premium feature (service-enforced via `checkJoinedAdmin` + `checkPremiumPlanInsight`). Members, Contributors, Observers — and Admins on a non-premium plan — are rejected, even though `dr_check_my_permissions` lists "view insights and analytics" for Members/Contributors. Surface a clear "requires an Admin on a premium Insights plan" message instead of retrying.

- `dr_get_insights` — query user/file/group engagement metrics; `dataroom_id` (req), `dimension` (req: user|file|group), `id` (optional specific user/file/group filter), `top_n` (optional, default 10), `sort_by` (optional: view_download [default] | time_spent | access). Use `dr_list_groups` to discover group IDs before `dr_get_insights(dimension="group")`
- `dr_get_timeline` — view activity trends over time (day buckets) for a single entity; `dataroom_id` (req), `dimension` (req: user|file|group), `id` (REQUIRED — no aggregate/whole-room mode; discover via dr_list_participants/dr_list_files/dr_list_groups), `version_index` (required for `dimension="file"`, from `dr_get_insights`), `limit` (optional, default 30 day buckets)
- `dr_get_activity_log` — see recent events and audit trail; `dataroom_id` (req), `time_range_days` (optional, default 7), `activity_type` (optional free-form filter — not an enforced enum; e.g. create/rename/archive/invite/join/permission_change/remove_users/request_access), `limit` (optional, default 50)
- `dr_get_dataroom_summary` — get a health check snapshot

## UI Rendering (mcp:render scope)

Three **display-only** render tools turn structured data into interactive `ui://` widgets (MCP Apps). They render as sandboxed iframes in UI-capable hosts (Claude Code, Cowork); in text-only / headless contexts they are not shown, so ALWAYS also summarize the data in markdown. These tools are NOT prefixed with `dr_`.

**Availability is environment-dependent — rely on your live tool list, never assume.** These tools exist only on Anduin servers that have shipped UI rendering (rolled out per environment — local/staging ahead of production) AND only when your grant includes the **`mcp:render`** OAuth scope (independent of `dataroom:*`). Your available tools are the source of truth: before offering a rendered view, confirm the specific render tool is actually present; if it is not, your server/environment simply hasn't enabled it yet — quietly fall back to a markdown table/list (don't announce a missing tool unless asked).

These tools are a **presentation layer only**: values are shown for viewing and CANNOT be edited or sent back (`interactive: false`).

| Tool | Renders (`ui://`) | Key inputs |
|------|------|-----------|
| `render_chart` | ECharts chart (`ui://anduin/chart`) | `title` (req), `echarts_option` (req — ECharts option object: series/xAxis/yAxis/tooltip/legend), `width` (opt, 200–1200, default 600), `height` (opt, 150–800, default 400) |
| `render_table` | Data table (`ui://anduin/table`) | `title` (req), `columns` (req — `[{id, label, type?: text\|number\|tag-list\|badge\|link}]`), `rows` (req — `[{<column id>: value, …, id}]`; tag-list cells are JSON string arrays) |
| `render_ui` | Form-layout view (`ui://anduin/form`) | `component: "form"` (req), `title` (req), `description` (opt), `sections` (req — `[{title, fields:[{alias, label, type: text\|number\|select\|checkbox\|date\|textarea, value, required?, options?}]}]`) |

Limits (over-limit/malformed input returns `{ "error": ... }` — fall back to markdown): chart `echarts_option` ≤100 KB; table ≤20 columns / ≤200 rows; form ≤20 sections / ≤50 fields.

**When to render (vs. plain markdown):** render when a visual genuinely helps — a chart of file engagement or activity over time, a sortable table of participants or files, a form-style snapshot of a data room's details. Prefer plain markdown for a single fact or a short list, and when running headless. Build the data with the read tools FIRST, then pass it to a render tool, and STILL give a one-line text summary so non-UI clients stay functional.

- "Chart the most-viewed files" → `dr_get_insights(dimension="file")` → `render_chart` (bar)
- "Show participants as a table" → `dr_list_participants` → `render_table` (name / role / status columns)
- "Summarize this data room" → `dr_get_dataroom_detail` → `render_ui` (form sections, display-only)

## Tool Chaining Rules

IDs flow between tools in a strict order. ALWAYS copy IDs exactly as returned — never shorten, modify, or invent IDs.

```
1. dr_list_entities → entity_id → dr_create_dataroom (entity_id optional — auto-resolves for single-entity users; entity-scoped, no dataroom_id)
2. dr_list_datarooms → dataroom_id → dataroom-scoped tools (detail, participants, files, search, insights, timeline, activity log, summary, groups, create/rename/archive, invite/remove/modify, folder/item ops)
3. dr_list_files → file_id/folder_id → dr_rename_item, dr_delete_items; dr_restore_items takes file_id ONLY (folders cannot be restored)
4. dr_list_participants → user_id → dr_remove_users, dr_modify_user_permissions
5. dr_list_groups → group_id → dr_get_insights(dimension="group")
6. dr_list_files/dr_search → file_id → dr_get_file_download_url, dr_convert_document_to_markdown, dr_convert_spreadsheet_to_markdown (file-scoped, NOT dataroom-scoped)
7. (large doc) dr_convert_document_to_markdown → page index → dr_read_document_pages(start_page, end_page)
8. (multi-sheet) dr_convert_spreadsheet_to_markdown → sheet_index → dr_read_spreadsheet_sheet
```

## Workflows

### Data Room Discovery
1. Call `dr_list_entities` to understand the user's organizations
2. Call `dr_list_datarooms` to discover accessible data rooms
3. Summarize: "You have X data rooms across Y entities"
4. Offer suggestions: create a data room, manage participants, organize files

### Participant Management
1. Ask for email addresses (comma-separated or one at a time)
2. Ask for role: Admin, Member, Contributor, or Observer
3. Check `dr_list_participants` to avoid duplicates
4. Confirm the invite list before proceeding
5. Call `dr_invite_users` after confirmation
6. Always confirm destructive actions (remove, role change) before executing

### File Organization
1. Call `dr_list_files` to see current structure
2. For folder creation: suggest structure based on common patterns (by date, type, project)
3. For cleanup: always confirm before deleting — deleted files go to trash and are recoverable via `dr_restore_items`, but deleted folders cannot currently be restored
4. Use `dr_search` to quickly find specific files

### Analytics Exploration
> Requires an **Admin** caller on a **premium Insights-plan** data room — every analytics tool (summary, insights, timeline, activity log) is gated. If a call errors with an access/plan message, tell the user this needs an Admin on a premium plan rather than retrying.
1. Start with `dr_get_dataroom_summary` for the health check overview
2. Ask what to explore: participants, files, or activity
3. For participants: `dr_get_insights(dimension="user")`
4. For files: `dr_get_insights(dimension="file")`
5. For groups: `dr_list_groups` then `dr_get_insights(dimension="group")`
6. For activity: `dr_get_activity_log` for recent events
7. For trends: `dr_get_timeline` for time-based patterns
8. Present data as markdown tables for clarity

### Document Reading Workflow
1. Navigate to the file: `dr_list_files` to find the file_id
2. For PDFs/images: `dr_convert_document_to_markdown` with the file_id
3. If a large document (size/character-based trigger, ~50,000 chars ≈ 50+ pages): review the page index, then `dr_read_document_pages` for specific ranges (max 30 pages per call)
4. For spreadsheets: `dr_convert_spreadsheet_to_markdown` with the file_id
5. If multi-sheet: `dr_read_spreadsheet_sheet` with sheet_index for specific sheets
6. Present the extracted content to the user

## Best Practices
- Use `dr_list_entities` first to understand the user's organization context
- Use `dr_list_datarooms` to discover data room IDs rather than guessing
- Use `dr_list_files` with folder_id to navigate directory structure
- Check `dr_list_participants` before inviting to avoid duplicates
- Use `dr_archive_dataroom` instead of deleting when data should be preserved
