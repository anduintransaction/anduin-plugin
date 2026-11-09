---
name: build-landing-page
description: Builds or changes the user's Anduin landing page in their open builder tab. Use when they ask to create, design, edit or restyle a landing page, or mention the Anduin landing page builder.
---

The user's Anduin landing page builder is open in their browser. This plugin's `anduin-builder` tools reach it:

1. Get the 6-character code. It is usually in the user's message ("…with code K7Q2PX…"). If not, ask them to click
   the sparkles button in the builder's left rail, then "Connect your agent", and read you the code it shows. Call
   `connect_builder` with it.
2. Call `list_builder_tools`, then `call_builder_tool` with `get_guide`, and follow the guide.
3. Make the change with `call_builder_tool` (`read_page`, `get_schema`, `edit_page`, …), passing each tool's input as
   an object.
4. Check your work: `call_builder_tool` with `check_page` for `desktop` and `mobile` on the sections you changed. Fix
   what it finds and check again.
5. When you finish, tell the user to review the page, click Save, then Exit AI mode. Nothing is saved until they click
   Save.

Change the page only through these tools. If a tool says the builder disconnected, and it still shows a code, call
`connect_builder` with it again (ask the user to click Retry if needed); otherwise ask them to click "Connect your agent"
and give you the new code. Then call `read_page`. If `connect_builder` says no builder is waiting with the code, ask the
user to click Retry under the code in the builder (or, if it shows "No agent found", click Retry after starting you),
then try again.
