---
name: build-landing-page
description: Builds or changes the user's Anduin landing page in their open builder tab. Use when they ask to create, design, edit or restyle a landing page, or mention the Anduin landing page builder.
---

The user's Anduin landing page builder is open in their browser. This plugin's `anduin-builder` tools reach it:

1. Get the 6-character code. It is usually in the user's message ("…with code K7Q2PX…"). If not, ask them to click
   the sparkles button in the builder's left rail, then "Connect your agent", and read you the code it shows. Call
   `connect_builder` with it.
2. Call `list_builder_tools`, then `call_builder_tool` with `set_working` (`{"working": true}`) so the builder shows
   you are working, then `get_guide`, and follow the guide.
3. Make the change with `call_builder_tool` (`read_page`, `get_schema`, `edit_page`, …), passing each tool's input as
   an object.
4. Check your work: `call_builder_tool` with `check_page` for `desktop` and `mobile` on the sections you changed. Fix
   what it finds and check again.
5. When you finish, or stop to ask the user something, call `set_working` with `{"working": false}`. Then tell the
   user to review the page, click Save, then Exit AI mode. Nothing is saved until they click Save.

Change the page only through these tools. If a tool says the builder disconnected or that no builder is connected,
call `connect_builder` again with the code you were given, without asking the user; then call `read_page`. If it says no
builder is waiting with the code, wait a few seconds and try again: the builder keeps looking for agents, including one
started after it showed the code. Only if that keeps failing, ask the user to click "Connect your agent" and give you
the new code.
