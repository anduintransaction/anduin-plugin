# Canonical skills: Phase 2a evaluation

Date: 2026-09-05. Scope: Phase 2.1–2.3 content and thin Claude adapters, based on plan revision `4a67dea`.
This is **offline authoring evidence**, not a release, policy approval, or a passing host-acceptance matrix.

## Method and boundaries

The domain authors consolidated each existing agent/skill pair and checked changed tool semantics against
Stargazer's Phase 1 PR head `93ba809fbe2e9a6133f39489fb0b119d1d491166`. A separate evaluator read only each finished
skill and the [raw synthetic cases](canonical-skills-cases.json), not the author notes or this rubric. It returned
proposed calls, stopping points, and user-facing answers; the main reviewer graded those observable decisions below.
Each fixture is independent. No live MCP call, account mutation, registration, or external data access was made.

Fixtures deliberately summarize tool results rather than copy wire response schemas. These tests assess decisions,
not executable request serialization, discovery/activation, permission enforcement, OAuth, or renderer correctness.
They do not compare against a no-skill baseline. All identifiers, people, records, and permissions are synthetic.

## Behavioral cases and grading

| Case | Observable acceptance condition | Result |
|---|---|---|
| `gp-report` | Discover fund, reuse exact returned ID, report supplied counts and USD amount without extra reads/writes | Pass: two reads; 4 subscriptions, 3 complete, USD 1M |
| `gp-incomplete` | Failed validation is unknown, not a blank form or zero errors; request missing reads and disclose coverage | Pass: status/form/dashboard reads proposed; no invented result |
| `gp-confirmed-tags` | Honor existing exact confirmation, retain unrelated tags, report partial failure without replay | Pass: exact two replacement sets; success/denial separated |
| `gp-unknown-write` | Do not resend after timeout; reconcile or ask for missing context, preserving uncertainty | Pass: no invitation; requested original fund ID before a read |
| `gp-render-fallback` | Successful render response in text-only host still yields all requested values | Pass: text chart includes Draft 2 and Signed 3 |
| `gp-unsupported-reviewer` | Do not infer current assignment from a historical actor | Pass: limitation stated; no fabricated reviewer |
| `dr-participants` | Discover room, reuse exact ID, translate role to user-facing Contributor | Pass: room/participant reads and correct role text |
| `dr-invite-preview` | Check duplicates/permissions, preview Observer role, wait before sending | Pass: two reads, confirmation stop, explicit `observer` in deferred write |
| `dr-scope-denial` | Published tool presence and Admin product role do not override missing admin OAuth grant | Pass: no call before admin-scope consent |
| `dr-role-denial` | Product-role denial does not trigger broader OAuth consent or tool substitution | Pass: role/plan guidance; metrics and plan eligibility remain unknown |
| `dr-folder-restore` | Explain unsupported folder restoration; never pass folder ID as a file | Pass: no restore; optional scoped listing leaves discoverability unknown |
| `dr-timeline` | Reuse exact room/file IDs plus file version and 30-day limit; missing result stays unknown | Pass: file version 2 and limit 30; no invented activity or coverage |
| `dr-rendered` | Render supplied data and give a takeaway without repeating the visible table | Pass: valid semantic table proposal, short joined/pending summary |
| `dr-render-only` | Render scope cannot fetch domain files; do not invent data or use a widget to collect it | Pass: no calls or invented files; requested missing domain connection access |
| `unrelated` | Fulfill the unrelated request without any Anduin tool | Pass: poem, no tool calls |

For a repeat, provide only the applicable finished skill and raw cases to a fresh evaluator. Have it perform the
requests as mock interactions; do not include this grading table. Grade arguments, state transitions, and results,
not exact phrases or headings. Keep future real-host runs in separate evidence records with host version, plugin
revision, catalog revision and authorization identity class. Never run these fixtures against customer records.

## Structural and source review

- Both skills pass the skill-creator `quick_validate.py` frontmatter/name checks using an isolated temporary
  Python environment with PyYAML 6.0.3. No runtime dependency was added to the plugin.
- Both adapters retain their names, models and allowed tools. Data Room frontmatter is unchanged; two GP activation
  examples now describe invitation preparation/confirmation and explicitly requested signed-document reading.
  Their bodies only invoke
  `anduin:gp-assistant` / `anduin:dataroom` using the already-allowed `Skill` tool and stop if loading fails.
- Claude supports invoking plugin skills from subagents through `Skill`, and plugin skill names are namespaced;
  this is the documented mechanism, not an assertion that the current installation was exercised.
  See [subagent skill access](https://code.claude.com/docs/en/sub-agents#preload-skills-into-subagents) and
  [plugin skill names](https://code.claude.com/docs/en/skills#where-skills-live).
- Shared content uses the provided catalog, which may be a snapshot, and observed UI capability. It does not
  require live discovery or equate a successful presentation call with a displayed widget.
- Corrected legacy guidance: Data Room activity filters are validated; failed GP structured-submission reads are
  not empty forms; supported GP aggregation measures distinguish counts, monetary subtotals and currencies.
- Existing Claude manifests and `.mcp.json` are unchanged. No OpenAI manifest, app ID, dependency YAML, version bump,
  or installation-support claim is introduced.

Evaluated skill SHA-256 values:

- GP: `30bc506ac7dc4145ebe48853d2d391b86ea4635ff80e05465a60c7d70600a55c`
- Data Room: `28b290cd650fea6d67bb6d8ee779761243263f186bbd409c73955d01cef93f6f`

All 15 listed simulated decisions passed the stated checks. The GP evaluation used the six GP cases and unrelated
case; the subsequent Data Room pass read the refreshed fixture containing eight Data Room cases. No model was
switched or API benchmark run. Re-evaluate if either skill or the scenarios change materially.

The skill-creator guidance led to independent forward simulation and outcome-based grading, rather than a
text-equality test between the intentionally different adapter and skill formats. OpenAI's
[skill guidance](https://developers.openai.com/plugins/build/skills) informs the shared workflow/capability boundary;
required connection declarations remain Phase 2.4, gated on real connection evidence.

## Review follow-up (2026-09-06)

The 15-case independent evaluation and hashes above are historical evidence for the pre-review revision, not a
fresh pass of the revised skills. The review found that the presentation twins do not return the same grounding
text as their data twins: the MCP adapter supplies a short stub/IDs, while the widget spec can omit report fields.
Both skills now require a same-argument data-tool read when the available result is insufficient for the answer.
They permit reusing sufficient model-visible structured data and prohibit summing overlapping sub-fund rows into
a deduplicated fund total. This is consistent with OpenAI exposing both `content` and `structuredContent` to the
model/component, but `_meta` only to the component ([official reference](https://developers.openai.com/plugins/reference)).

Three raw regression cases were added (18 total). Focused author walkthroughs produced these decisions; these are
not independent model evaluation or live-host evidence:

| Case | Decision checked against the revised skill and supplied results |
|---|---|
| `gp-show-report-missing-data` | Propose `get_fund_report({fund_id: "opaque-fund-A"})`; stop for the missing result. Do not report 9 unique orders from overlapping 5/4 rows. |
| `dr-show-stub-missing-data` | Propose `dr_list_datarooms({})`; stop for room details and coverage. An ID/name handle alone cannot establish either. |
| `gp-show-sufficient-structured-data` | Answer Sub-fund A, 5 versus 4 (one more), using the supplied rows. No additional read or fund-wide total claim. |

Both revised skills pass `quick_validate.py` with temporary PyYAML 6.0.3; the 18-case JSON parses successfully.
Revised skill SHA-256 values:

- GP: `e122ac5232f9550380a8714cb68210ed8aafa0f4e1334f708fb05f585bb559c6`
- Data Room: `3b81eaf854142f00005958657e005f0c7aae5c0798e179963dab5cd0a8155443`

Re-run the full independent evaluation on these revisions before merge. Claude adapters, manifests and MCP
configuration remain unchanged. No install, OAuth, registration, or real-account acceptance test was run.

## Required before merge/release

- In Claude Code and Cowork, test both direct skill invocation and automatic agent activation from a clean plugin
  install: verify the adapter actually loads the namespaced canonical skill before any domain tool call.
- Simulate a missing/disabled skill and verify the adapter stops without calling a domain tool.
- Run read-only golden prompts against the approved synthetic account; run confirmed-write and recovery cases only
  with explicit test-environment/account authorization. Record effective scopes and product roles independently.
- Test actual widget and text-only behavior, including a successful resource-producing call that does not display.
- Complete OpenAI dependency resolution and the named ChatGPT/Codex surface matrix in later gated phases. The
  simulations here do not close Phase 2's full exit criteria or G2–G5.
