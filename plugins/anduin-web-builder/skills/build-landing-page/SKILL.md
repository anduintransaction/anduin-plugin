---
name: build-landing-page
description: Builds or changes the user's Anduin landing page in their open builder tab. Use when they ask to create, design, edit or restyle a landing page, or mention the Anduin landing page builder.
---

The user's Anduin landing page builder is open in Chrome, and it exposes its own tools over WebMCP. Reach them through the chrome-devtools connector:

1. Call `list_pages`, then `select_page` on the tab that shows the landing page builder.
2. Call `list_webmcp_tools`. Run the builder's tools with `execute_webmcp_tool`, passing their parameters as a JSON string.
3. Call the builder's `get_guide` tool first and follow it, then make the change the user asked for.
4. Check your work with the builder's `show_page` tool and `take_screenshot`.

If `list_webmcp_tools` shows no builder tools, ask the user to open "Build with your AI agent" in the builder's left rail and click "Open for AI to connect". Nothing is saved until the user clicks Save.
