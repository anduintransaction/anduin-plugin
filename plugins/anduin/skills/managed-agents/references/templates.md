# Managed Agent Deployment Templates

Ready-to-use Python scripts for deploying Anduin agents as Claude Managed Agents.
Each template creates a credential vault, the agent, and an environment, and (except
Template 5) runs a test session. Credentials always flow through a vault — injected by
URL match at session time — never hardcoded into the agent definition.

## Template 1: GP Assistant (Interactive)

Deploy the GP fund subscription assistant as an API-callable managed agent.

```python
"""
Anduin GP Assistant — Managed Agent Deployment
Interactive fund subscription management via API.
"""
import os
from urllib.parse import urlparse

import anthropic

client = anthropic.Anthropic()

# --- Configuration ---
ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")

# --- Step 1: Create a Vault with the Anduin Credential ---
# The credential is injected by URL match when the agent connects to the MCP
# server. Never put tokens in the agent definition.
vault = client.beta.vaults.create(display_name="anduin-gp-credentials")

# Preferred: mcp_oauth — Anthropic refreshes the access token automatically,
# so long-running sessions survive token expiry.
client.beta.vaults.credentials.create(
    vault_id=vault.id,
    display_name="Anduin OAuth",
    auth={
        "type": "mcp_oauth",
        "mcp_server_url": ANDUIN_MCP_URL,
        "access_token": os.environ["ANDUIN_ACCESS_TOKEN"],
        "expires_at": os.environ["ANDUIN_TOKEN_EXPIRES_AT"],  # ISO 8601
        "refresh": {
            "token_endpoint": os.environ["ANDUIN_TOKEN_ENDPOINT"],
            "client_id": os.environ["ANDUIN_CLIENT_ID"],
            "scope": "fundsub:read fundsub:write",
            "refresh_token": os.environ["ANDUIN_REFRESH_TOKEN"],
            # Match your OAuth client's token_endpoint_auth_method
            # ("client_secret_basic", "client_secret_post", or "none").
            "token_endpoint_auth": {
                "type": "client_secret_basic",
                "client_secret": os.environ["ANDUIN_CLIENT_SECRET"],
            },
        },
    },
)
# Quick-start alternative: static_bearer — a fixed token with NO auto-refresh
# (sessions start failing once the token expires; fine for a one-off test):
# client.beta.vaults.credentials.create(
#     vault_id=vault.id,
#     display_name="Anduin token",
#     auth={
#         "type": "static_bearer",
#         "mcp_server_url": ANDUIN_MCP_URL,
#         "token": os.environ["ANDUIN_OAUTH_TOKEN"],
#     },
# )
print(f"Vault created: {vault.id}")

# --- Step 2: Create the Agent ---
agent = client.beta.agents.create(
    name="anduin-gp-assistant",
    model="claude-sonnet-5",
    system="""You are an AI assistant specialized in reviewing LP (Limited Partner) \
subscriptions for fund administration on the Anduin platform.

Your Capabilities:
- Discovering funds, LPs, and orders
- Reviewing LP subscription forms, documents, and compliance status
- Analyzing orders across a fund (cross-order comparison, search)
- Managing order tags, custom data, and workflow status
- Reporting fund-level statistics and dashboards
- Inviting fund managers to groups
- Reading documents via OCR conversion

Tool Chaining — IDs Flow Between Tools:
Step 1: list_funds → fund_id
Step 2: query_dashboard → order_id (PREFERRED for browsing)
        OR list_orders → order_id (lightweight discovery)
Step 3: get_order_workflow_data → tags, contacts, commitments
Step 4: get_order_subscription_docs → file_id
Step 5: get_form_schema → field_alias → get_form_field_value

Critical ID Rules:
1. IDs are opaque strings — NEVER construct, guess, or modify them
2. ALWAYS obtain IDs from tool outputs
3. Copy IDs exactly as returned""",
    mcp_servers=[
        {"type": "url", "name": "anduin", "url": ANDUIN_MCP_URL},
    ],
    tools=[
        {"type": "agent_toolset_20260401"},
        {
            "type": "mcp_toolset",
            "mcp_server_name": "anduin",
            # MCP toolsets default to always_ask, which pauses unattended
            # sessions for approval — auto-approve the trusted Anduin tools.
            "default_config": {"permission_policy": {"type": "always_allow"}},
        },
    ],
)
print(f"Agent created: {agent.id}")

# --- Step 3: Create Environment ---
environment = client.beta.environments.create(
    name="anduin-gp-env",
    config={
        "type": "cloud",
        # Least privilege: the sandbox only needs to reach the Anduin MCP server.
        "networking": {
            "type": "limited",
            "allowed_hosts": [urlparse(ANDUIN_MCP_URL).hostname],
            "allow_mcp_servers": True,
        },
    },
)
print(f"Environment created: {environment.id}")

# --- Step 4: Start a Session (vault_ids injects the Anduin credential) ---
session = client.beta.sessions.create(
    agent=agent.id,
    environment_id=environment.id,
    vault_ids=[vault.id],
)
print(f"Session started: {session.id}")

# --- Step 5: Send a Test Message ---
client.beta.sessions.events.send(
    session.id,
    events=[{
        "type": "user.message",
        "content": [{"type": "text", "text": "List my funds and show a summary of each."}]
    }]
)

# --- Step 6: Stream Results ---
for event in client.beta.sessions.events.stream(session.id):
    if hasattr(event, 'content'):
        print(event.content, end="", flush=True)
    elif hasattr(event, 'type'):
        print(f"\n[Event: {event.type}]")
```

---

## Template 2: Subscription Reviewer (Event-Driven)

Automated LP subscription review triggered by form submission webhooks.

```python
"""
Anduin Subscription Reviewer — Managed Agent
Automatically reviews LP submissions for completeness.
Trigger via webhook when an LP submits a form.
"""
import os
from urllib.parse import urlparse

import anthropic

client = anthropic.Anthropic()

ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")

# --- Create Vault (once) ---
# mcp_oauth auto-refreshes the token, so webhook-triggered reviews keep
# working long after the initial access token would have expired.
# (See Template 1 for the static_bearer quick-start alternative — no refresh.)
vault = client.beta.vaults.create(display_name="anduin-reviewer-credentials")
client.beta.vaults.credentials.create(
    vault_id=vault.id,
    display_name="Anduin OAuth",
    auth={
        "type": "mcp_oauth",
        "mcp_server_url": ANDUIN_MCP_URL,
        "access_token": os.environ["ANDUIN_ACCESS_TOKEN"],
        "expires_at": os.environ["ANDUIN_TOKEN_EXPIRES_AT"],  # ISO 8601
        "refresh": {
            "token_endpoint": os.environ["ANDUIN_TOKEN_ENDPOINT"],
            "client_id": os.environ["ANDUIN_CLIENT_ID"],
            "scope": "fundsub:read fundsub:write",
            "refresh_token": os.environ["ANDUIN_REFRESH_TOKEN"],
            "token_endpoint_auth": {
                "type": "client_secret_basic",
                "client_secret": os.environ["ANDUIN_CLIENT_SECRET"],
            },
        },
    },
)

# --- Create Agent (once, reuse across reviews) ---
agent = client.beta.agents.create(
    name="anduin-subscription-reviewer",
    model="claude-sonnet-5",
    system="""You are an automated subscription review agent for the Anduin platform. \
When given an order ID, perform a comprehensive review:

Review Process:
1. Call get_lp_status to understand the LP's current state
2. Call get_form_validation_errors to identify incomplete fields
3. Call get_form_schema and get_form_markdown to review form content
4. Call get_order_subscription_docs to list documents
5. For each key document: convert_document_to_markdown to review content
6. Call get_aml_check to verify compliance status
7. Analyze findings and draft_comment with a structured review

Output Format:
Provide a structured review report with:
- Overall Status: PASS / NEEDS_ATTENTION / CRITICAL_ISSUES
- Form Completeness: percentage and missing fields
- Document Status: which docs are present/missing/expired
- AML/KYC Status: current compliance state
- Issues Found: numbered list with severity (Critical/Warning/Info)
- Recommended Actions: what the GP should do next

Critical Rules:
- NEVER fabricate IDs. Always obtain from tool outputs.
- ALWAYS call get_lp_status before any other order-specific tools.
- Report findings objectively. Do not approve or reject orders.
- Use draft_comment to post the review summary on the order.""",
    mcp_servers=[
        {"type": "url", "name": "anduin", "url": ANDUIN_MCP_URL},
    ],
    tools=[
        {"type": "agent_toolset_20260401"},
        {
            "type": "mcp_toolset",
            "mcp_server_name": "anduin",
            # Auto-approve: webhook-driven runs have no human to confirm calls.
            "default_config": {"permission_policy": {"type": "always_allow"}},
        },
    ],
)
print(f"Reviewer agent created: {agent.id}")

environment = client.beta.environments.create(
    name="anduin-reviewer-env",
    config={
        "type": "cloud",
        "networking": {
            "type": "limited",
            "allowed_hosts": [urlparse(ANDUIN_MCP_URL).hostname],
            "allow_mcp_servers": True,
        },
    },
)


def review_order(order_id: str) -> str:
    """Review a single LP order. Call this from a webhook handler."""
    session = client.beta.sessions.create(
        agent=agent.id,
        environment_id=environment.id,
        vault_ids=[vault.id],
    )

    client.beta.sessions.events.send(
        session.id,
        events=[{
            "type": "user.message",
            "content": [{
                "type": "text",
                "text": f"Review LP order {order_id} for completeness. "
                        f"Check form validation, documents, and AML status. "
                        f"Post your findings as a draft comment on the order."
            }]
        }]
    )

    result_parts = []
    for event in client.beta.sessions.events.stream(session.id):
        if hasattr(event, 'content'):
            result_parts.append(event.content)

    return "".join(result_parts)


# --- Test Run ---
if __name__ == "__main__":
    import sys
    order_id = sys.argv[1] if len(sys.argv) > 1 else "test-order-id"
    print(f"Reviewing order: {order_id}")
    result = review_order(order_id)
    print(result)
```

---

## Template 3: Fund Health Reporter (Scheduled)

Nightly/weekly fund health report generation.

```python
"""
Anduin Fund Health Reporter — Managed Agent
Runs on a schedule to generate fund health reports.
"""
import os
from urllib.parse import urlparse

import anthropic

client = anthropic.Anthropic()

ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")

# --- Create Vault (once) ---
# Scheduled agents MUST use mcp_oauth: the refresh block lets Anthropic renew
# the access token between runs, so cron jobs don't die on token expiry.
# (See Template 1 for the static_bearer quick-start alternative — no refresh.)
vault = client.beta.vaults.create(display_name="anduin-reporter-credentials")
client.beta.vaults.credentials.create(
    vault_id=vault.id,
    display_name="Anduin OAuth",
    auth={
        "type": "mcp_oauth",
        "mcp_server_url": ANDUIN_MCP_URL,
        "access_token": os.environ["ANDUIN_ACCESS_TOKEN"],
        "expires_at": os.environ["ANDUIN_TOKEN_EXPIRES_AT"],  # ISO 8601
        "refresh": {
            "token_endpoint": os.environ["ANDUIN_TOKEN_ENDPOINT"],
            "client_id": os.environ["ANDUIN_CLIENT_ID"],
            "scope": "fundsub:read",
            "refresh_token": os.environ["ANDUIN_REFRESH_TOKEN"],
            "token_endpoint_auth": {
                "type": "client_secret_basic",
                "client_secret": os.environ["ANDUIN_CLIENT_SECRET"],
            },
        },
    },
)

agent = client.beta.agents.create(
    name="anduin-fund-health-reporter",
    model="claude-sonnet-5",
    system="""You are a fund health reporting agent for the Anduin platform. \
Generate comprehensive fund health reports by analyzing all accessible funds.

Report Generation Process:
1. Call list_funds to discover all funds
2. For each fund:
   a. get_fund_info for fund details
   b. get_fund_report for aggregate statistics
   c. query_dashboard with page_size=100 to get all orders
   d. Identify: stale orders (no activity > 14 days), incomplete forms,
      expiring AML/KYC checks, unsigned documents
3. Generate a structured health report

Report Format:
# Fund Health Report — [Date]

## Executive Summary
- Total funds: N
- Total LPs across all funds: N
- Issues requiring attention: N

## Per-Fund Analysis
### [Fund Name]
- Status: HEALTHY / NEEDS_ATTENTION / AT_RISK
- Total LPs: N (Completed: N, In Progress: N, Not Started: N)
- Form Completion Rate: N%
- Stale Orders (>14 days no activity): list
- Expiring AML/KYC: list with expiry dates
- Missing Documents: list

## Recommended Actions
Numbered list of priority actions across all funds.

Rules:
- Never fabricate data. Only report what tools return.
- Always start with list_funds. Never guess fund IDs.
- Process ALL funds, not just the first one.""",
    mcp_servers=[
        {"type": "url", "name": "anduin", "url": ANDUIN_MCP_URL},
    ],
    tools=[
        {"type": "agent_toolset_20260401"},
        {
            "type": "mcp_toolset",
            "mcp_server_name": "anduin",
            # Auto-approve: scheduled runs have no human to confirm calls.
            "default_config": {"permission_policy": {"type": "always_allow"}},
        },
    ],
)

environment = client.beta.environments.create(
    name="anduin-reporter-env",
    config={
        "type": "cloud",
        "networking": {
            "type": "limited",
            "allowed_hosts": [urlparse(ANDUIN_MCP_URL).hostname],
            "allow_mcp_servers": True,
        },
    },
)


def generate_report() -> str:
    """Generate a fund health report. Call from a cron job."""
    session = client.beta.sessions.create(
        agent=agent.id,
        environment_id=environment.id,
        vault_ids=[vault.id],
    )

    client.beta.sessions.events.send(
        session.id,
        events=[{
            "type": "user.message",
            "content": [{
                "type": "text",
                "text": "Generate a comprehensive fund health report for all "
                        "accessible funds. Check for stale orders, incomplete "
                        "forms, expiring AML/KYC, and missing documents."
            }]
        }]
    )

    result_parts = []
    for event in client.beta.sessions.events.stream(session.id):
        if hasattr(event, 'content'):
            result_parts.append(event.content)

    return "".join(result_parts)


if __name__ == "__main__":
    print("Generating fund health report...")
    report = generate_report()
    print(report)
```

---

## Template 4: Compliance Monitor (Scheduled Daily)

Daily AML/KYC compliance sweep across all funds.

```python
"""
Anduin Compliance Monitor — Managed Agent
Daily sweep for expiring AML/KYC and missing compliance docs.
"""
import os
from urllib.parse import urlparse

import anthropic

client = anthropic.Anthropic()

ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")

# --- Create Vault (once) ---
# Scheduled agents MUST use mcp_oauth so the token refreshes between daily
# runs. (See Template 1 for the static_bearer quick-start alternative.)
vault = client.beta.vaults.create(display_name="anduin-compliance-credentials")
client.beta.vaults.credentials.create(
    vault_id=vault.id,
    display_name="Anduin OAuth",
    auth={
        "type": "mcp_oauth",
        "mcp_server_url": ANDUIN_MCP_URL,
        "access_token": os.environ["ANDUIN_ACCESS_TOKEN"],
        "expires_at": os.environ["ANDUIN_TOKEN_EXPIRES_AT"],  # ISO 8601
        "refresh": {
            "token_endpoint": os.environ["ANDUIN_TOKEN_ENDPOINT"],
            "client_id": os.environ["ANDUIN_CLIENT_ID"],
            "scope": "fundsub:read",
            "refresh_token": os.environ["ANDUIN_REFRESH_TOKEN"],
            "token_endpoint_auth": {
                "type": "client_secret_basic",
                "client_secret": os.environ["ANDUIN_CLIENT_SECRET"],
            },
        },
    },
)

agent = client.beta.agents.create(
    name="anduin-compliance-monitor",
    model="claude-sonnet-5",
    system="""You are a compliance monitoring agent for the Anduin platform. \
Perform a daily sweep across all funds to identify compliance risks.

Monitoring Process:
1. Call list_funds to discover all funds
2. For each fund: query_dashboard to get all active orders
3. For orders with compliance flags: get_aml_check for detailed status
4. Identify risks:
   - AML/KYC checks expiring within 30 days
   - Missing required compliance documents
   - Orders stuck in compliance review > 7 days
   - Failed or expired background checks

Output Format:
# Compliance Alert Report — [Date]

## Critical (Action Required Today)
- [List items requiring immediate attention]

## Warning (Action Required This Week)
- [List items approaching deadlines]

## Info (Monitor)
- [Items to keep an eye on]

## Statistics
- Total orders checked: N
- Clean: N | Warning: N | Critical: N

Rules:
- Report ONLY issues found by tools. Never fabricate compliance data.
- Always start with list_funds. Process every fund.
- For each flagged order, provide the order ID and LP name for easy lookup.""",
    mcp_servers=[
        {"type": "url", "name": "anduin", "url": ANDUIN_MCP_URL},
    ],
    tools=[
        {"type": "agent_toolset_20260401"},
        {
            "type": "mcp_toolset",
            "mcp_server_name": "anduin",
            # Auto-approve: scheduled runs have no human to confirm calls.
            "default_config": {"permission_policy": {"type": "always_allow"}},
        },
    ],
)

environment = client.beta.environments.create(
    name="anduin-compliance-env",
    config={
        "type": "cloud",
        "networking": {
            "type": "limited",
            "allowed_hosts": [urlparse(ANDUIN_MCP_URL).hostname],
            "allow_mcp_servers": True,
        },
    },
)


def run_compliance_check() -> str:
    """Run daily compliance sweep. Call from cron."""
    session = client.beta.sessions.create(
        agent=agent.id,
        environment_id=environment.id,
        vault_ids=[vault.id],
    )

    client.beta.sessions.events.send(
        session.id,
        events=[{
            "type": "user.message",
            "content": [{
                "type": "text",
                "text": "Run a compliance sweep across all funds. Check for "
                        "expiring AML/KYC, missing compliance docs, and stalled "
                        "compliance reviews. Report findings by severity."
            }]
        }]
    )

    result_parts = []
    for event in client.beta.sessions.events.stream(session.id):
        if hasattr(event, 'content'):
            result_parts.append(event.content)

    return "".join(result_parts)


if __name__ == "__main__":
    print("Running compliance sweep...")
    report = run_compliance_check()
    print(report)
```

---

## Template 5: Data Room Agent (Interactive)

Deploy the data room management agent as an API-callable service.

```python
"""
Anduin Data Room Agent — Managed Agent Deployment
Data room management via API.
"""
import os
from urllib.parse import urlparse

import anthropic

client = anthropic.Anthropic()

ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")

# --- Create Vault ---
# mcp_oauth auto-refreshes the token so the deployment keeps working after
# token expiry. (See Template 1 for the static_bearer quick-start alternative.)
vault = client.beta.vaults.create(display_name="anduin-dataroom-credentials")
client.beta.vaults.credentials.create(
    vault_id=vault.id,
    display_name="Anduin OAuth",
    auth={
        "type": "mcp_oauth",
        "mcp_server_url": ANDUIN_MCP_URL,
        "access_token": os.environ["ANDUIN_ACCESS_TOKEN"],
        "expires_at": os.environ["ANDUIN_TOKEN_EXPIRES_AT"],  # ISO 8601
        "refresh": {
            "token_endpoint": os.environ["ANDUIN_TOKEN_ENDPOINT"],
            "client_id": os.environ["ANDUIN_CLIENT_ID"],
            "scope": "dataroom:read dataroom:write",
            "refresh_token": os.environ["ANDUIN_REFRESH_TOKEN"],
            "token_endpoint_auth": {
                "type": "client_secret_basic",
                "client_secret": os.environ["ANDUIN_CLIENT_SECRET"],
            },
        },
    },
)
print(f"Vault created: {vault.id}")

agent = client.beta.agents.create(
    name="anduin-dataroom-agent",
    model="claude-sonnet-5",
    system="""You are an AI assistant specialized in managing virtual data rooms \
on the Anduin platform.

Your Capabilities:
- Discovering entities, data rooms, and files
- Creating new data rooms and organizing folder structures
- Managing participants (inviting, removing, changing roles)
- Searching and navigating files within data rooms
- Analyzing data room activity, user engagement, and file metrics
- Reading documents via OCR conversion

Tool Chaining — IDs Flow Between Tools:
dr_list_entities → entity_id → dr_create_dataroom (entity_id optional; auto-resolves for single-entity users)
dr_list_datarooms → dataroom_id → dataroom-scoped tools (detail, participants, files, search, insights, timeline, activity log, summary, groups, create/rename/archive, invite/remove/modify, folder/item ops)
dr_list_files/dr_search → file_id → dr_get_file_download_url, dr_convert_document_to_markdown, dr_convert_spreadsheet_to_markdown
dr_list_files → file_id/folder_id → dr_rename_item, dr_delete_items; dr_restore_items takes file_id ONLY (folders cannot be restored)
dr_list_participants → user_id → dr_remove_users, dr_modify_user_permissions

Participant Roles:
- Admin: Full access (view, create, upload, delete, manage participants, archive)
- Member: View, search, create folders, upload/rename/delete files, invite, view insights
- Contributor (internally Guest): View, search, create folders, upload/rename/delete files, view insights
- Observer (internally Restricted): Read-only (view and search only)
Note: the analytics tools (dr_get_insights/dr_get_timeline/dr_get_activity_log/dr_get_dataroom_summary) require an Admin on a premium Insights plan — Members/Contributors are rejected despite the "view insights" capability listed above.

Critical ID Rules:
1. IDs are opaque strings — NEVER construct, guess, or modify them
2. ALWAYS obtain IDs from tool outputs
3. Call dr_list_entities first to understand the organization context""",
    mcp_servers=[
        {"type": "url", "name": "anduin", "url": ANDUIN_MCP_URL},
    ],
    tools=[
        {"type": "agent_toolset_20260401"},
        {
            "type": "mcp_toolset",
            "mcp_server_name": "anduin",
            # Auto-approve the trusted Anduin tools for API-driven use.
            "default_config": {"permission_policy": {"type": "always_allow"}},
        },
    ],
)
print(f"Data room agent created: {agent.id}")

environment = client.beta.environments.create(
    name="anduin-dataroom-env",
    config={
        "type": "cloud",
        "networking": {
            "type": "limited",
            "allowed_hosts": [urlparse(ANDUIN_MCP_URL).hostname],
            "allow_mcp_servers": True,
        },
    },
)
print(f"Environment created: {environment.id}")

# When starting sessions for this agent, pass vault_ids so the Anduin
# credential is injected:
# session = client.beta.sessions.create(
#     agent=agent.id,
#     environment_id=environment.id,
#     vault_ids=[vault.id],
# )
```
