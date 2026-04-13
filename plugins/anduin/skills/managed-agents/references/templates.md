# Managed Agent Deployment Templates

Ready-to-use Python scripts for deploying Anduin agents as Claude Managed Agents.
Each template creates the agent, environment, vault, and runs a test session.

## Template 1: GP Assistant (Interactive)

Deploy the GP fund subscription assistant as an API-callable managed agent.

```python
"""
Anduin GP Assistant — Managed Agent Deployment
Interactive fund subscription management via API.
"""
import os
import anthropic

client = anthropic.Anthropic()

# --- Configuration ---
ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")
ANDUIN_TOKEN = os.environ["ANDUIN_OAUTH_TOKEN"]

# --- Step 1: Create the Agent ---
agent = client.beta.agents.create(
    name="anduin-gp-assistant",
    model="claude-sonnet-4-6",
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
    mcp_servers={
        "anduin": {
            "type": "http",
            "url": ANDUIN_MCP_URL,
            "headers": {
                "Authorization": f"Bearer {ANDUIN_TOKEN}"
            }
        }
    }
)
print(f"Agent created: {agent.id}")

# --- Step 2: Create Environment ---
environment = client.beta.environments.create(
    name="anduin-gp-env",
    packages=["python3"],
)
print(f"Environment created: {environment.id}")

# --- Step 3: Start a Session ---
session = client.beta.sessions.create(
    agent=agent.id,
    environment_id=environment.id,
)
print(f"Session started: {session.id}")

# --- Step 4: Send a Test Message ---
client.beta.sessions.events.send(
    session.id,
    events=[{
        "type": "user.message",
        "content": [{"type": "text", "text": "List my funds and show a summary of each."}]
    }]
)

# --- Step 5: Stream Results ---
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
import anthropic

client = anthropic.Anthropic()

ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")
ANDUIN_TOKEN = os.environ["ANDUIN_OAUTH_TOKEN"]

# --- Create Agent (once, reuse across reviews) ---
agent = client.beta.agents.create(
    name="anduin-subscription-reviewer",
    model="claude-sonnet-4-6",
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
    mcp_servers={
        "anduin": {
            "type": "http",
            "url": ANDUIN_MCP_URL,
            "headers": {"Authorization": f"Bearer {ANDUIN_TOKEN}"}
        }
    }
)
print(f"Reviewer agent created: {agent.id}")

environment = client.beta.environments.create(
    name="anduin-reviewer-env",
    packages=["python3"],
)


def review_order(order_id: str) -> str:
    """Review a single LP order. Call this from a webhook handler."""
    session = client.beta.sessions.create(
        agent=agent.id,
        environment_id=environment.id,
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
import anthropic

client = anthropic.Anthropic()

ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")
ANDUIN_TOKEN = os.environ["ANDUIN_OAUTH_TOKEN"]

agent = client.beta.agents.create(
    name="anduin-fund-health-reporter",
    model="claude-sonnet-4-6",
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
    mcp_servers={
        "anduin": {
            "type": "http",
            "url": ANDUIN_MCP_URL,
            "headers": {"Authorization": f"Bearer {ANDUIN_TOKEN}"}
        }
    }
)

environment = client.beta.environments.create(
    name="anduin-reporter-env",
    packages=["python3"],
)


def generate_report() -> str:
    """Generate a fund health report. Call from a cron job."""
    session = client.beta.sessions.create(
        agent=agent.id,
        environment_id=environment.id,
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
import anthropic

client = anthropic.Anthropic()

ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")
ANDUIN_TOKEN = os.environ["ANDUIN_OAUTH_TOKEN"]

agent = client.beta.agents.create(
    name="anduin-compliance-monitor",
    model="claude-sonnet-4-6",
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
    mcp_servers={
        "anduin": {
            "type": "http",
            "url": ANDUIN_MCP_URL,
            "headers": {"Authorization": f"Bearer {ANDUIN_TOKEN}"}
        }
    }
)

environment = client.beta.environments.create(
    name="anduin-compliance-env",
    packages=["python3"],
)


def run_compliance_check() -> str:
    """Run daily compliance sweep. Call from cron."""
    session = client.beta.sessions.create(
        agent=agent.id,
        environment_id=environment.id,
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
import anthropic

client = anthropic.Anthropic()

ANDUIN_MCP_URL = os.environ.get("ANDUIN_MCP_URL", "https://mcp.anduin.app/mcp")
ANDUIN_TOKEN = os.environ["ANDUIN_OAUTH_TOKEN"]

agent = client.beta.agents.create(
    name="anduin-dataroom-agent",
    model="claude-sonnet-4-6",
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
dr_list_entities → entity_id → dr_create_dataroom
dr_list_datarooms → dataroom_id → all other tools
dr_list_files → file/folder_id → dr_rename_item, dr_delete_items
dr_list_participants → user_id → dr_remove_users, dr_modify_user_permissions

Participant Roles:
- Admin: Full access (view, create, upload, delete, manage participants, archive)
- Member: View, search, create folders, upload/rename/delete files, invite
- Contributor: View, search, create folders, upload/rename/delete files
- Observer: Read-only (view and search only)

Critical ID Rules:
1. IDs are opaque strings — NEVER construct, guess, or modify them
2. ALWAYS obtain IDs from tool outputs
3. Call dr_list_entities first to understand the organization context""",
    mcp_servers={
        "anduin": {
            "type": "http",
            "url": ANDUIN_MCP_URL,
            "headers": {"Authorization": f"Bearer {ANDUIN_TOKEN}"}
        }
    }
)
print(f"Data room agent created: {agent.id}")

environment = client.beta.environments.create(
    name="anduin-dataroom-env",
    packages=["python3"],
)
print(f"Environment created: {environment.id}")
```
