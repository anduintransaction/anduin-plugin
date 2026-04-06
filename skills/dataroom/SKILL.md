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
| **Contributor** | View, search, create folders, upload/rename/delete files | Remove participants, modify permissions, rename/archive data room |
| **Observer** | View (read-only), search | Create, upload, delete, invite, manage, archive |

## MCP Tools by Category

All tools require OAuth2 scope `dataroom:read` or `dataroom:write`. Tools are prefixed with `dr_` in the MCP server.

### Entity & Data Room Discovery (dataroom:read)
- `dr_list_entities` — list all entities the user belongs to, with subscription plans and data room counts
- `dr_list_datarooms` — list all data rooms accessible to the user
- `dr_get_dataroom_detail` — get detailed information about a specific data room

### Data Room Operations (dataroom:write)
- `dr_create_dataroom` — create a new data room under a specific entity
- `dr_rename_dataroom` — rename an existing data room
- `dr_archive_dataroom` — archive or unarchive a data room

### Participant Management
- `dr_list_participants` — list all participants with roles and status (dataroom:read)
- `dr_invite_users` — invite users by email (dataroom:write)
- `dr_remove_users` — remove users from a data room (dataroom:write)
- `dr_modify_user_permissions` — change a user's role (dataroom:write)
- `dr_check_my_permissions` — check your current role and permissions (dataroom:read)
- `dr_list_groups` — list user groups in a data room (dataroom:read)

### File & Folder Management
- `dr_list_files` — list files and folders in a directory (dataroom:read)
- `dr_search` — search for files and folders by name (dataroom:read)
- `dr_create_folder` — create a new folder (dataroom:write)
- `dr_rename_item` — rename a file or folder (dataroom:write)
- `dr_delete_items` — delete files and/or folders (moves to trash) (dataroom:write)
- `dr_restore_items` — restore previously deleted items (dataroom:write)

### Analytics & Insights (dataroom:read)
- `dr_get_insights` — query user/file/group engagement metrics (dimension param)
- `dr_get_timeline` — view activity trends over time
- `dr_get_activity_log` — see recent events and audit trail
- `dr_get_dataroom_summary` — get a health check snapshot
- `dr_list_groups` — discover groups for group analytics

## Tool Chaining Rules

IDs flow between tools in a strict order. ALWAYS copy IDs exactly as returned — never shorten, modify, or invent IDs.

```
1. dr_list_entities → entity_id → dr_create_dataroom
2. dr_list_datarooms → dataroom_id → all other tools
3. dr_list_files → file/folder_id → dr_rename_item, dr_delete_items, dr_restore_items
4. dr_list_participants → user_id → dr_remove_users, dr_modify_user_permissions
5. dr_list_groups → group info → dr_get_insights(dimension="group")
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
3. For cleanup: always confirm before deleting — explain items go to trash (recoverable)
4. Use `dr_search` to quickly find specific files

### Analytics Exploration
1. Start with `dr_get_dataroom_summary` for the health check overview
2. Ask what to explore: participants, files, or activity
3. For participants: `dr_get_insights(dimension="user")`
4. For files: `dr_get_insights(dimension="file")`
5. For groups: `dr_list_groups` then `dr_get_insights(dimension="group")`
6. For activity: `dr_get_activity_log` for recent events
7. For trends: `dr_get_timeline` for time-based patterns
8. Present data as markdown tables for clarity

## Best Practices
- Use `dr_list_entities` first to understand the user's organization context
- Use `dr_list_datarooms` to discover data room IDs rather than guessing
- Use `dr_list_files` with folder_id to navigate directory structure
- Check `dr_list_participants` before inviting to avoid duplicates
- Use `dr_archive_dataroom` instead of deleting when data should be preserved
