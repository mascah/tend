# Design review, thread 2: runtime environments

Part of [the Tend and Grove design review](README.md): read that first. It
says what is decided and which file holds which section. Section numbers are
those of the single file these were split from.

From the second session, 2026-11-26: what the owner decided and the
evidence behind it.

## 15. Thread 2, runtime environments: decided 2026-11-26

### 15.1 Decisions

1. **One runtime class: a machine whose disk persists** (laptop, the mini,
   or a Linux VM). Owner: "I can live with a persistent disk for now."
   Hosted CI runs gates, not agents. Anthropic cloud sessions are parked.
   The record, the worktree and the agent's session transcript must share a
   disk until a push.
2. **Agents run headless; Tend's correctness needs no multiplexer.** Owner:
   yes, and "having a multiplexer improves the user experience greatly".
   So a multiplexer stays as an optional view (open a run's session or its
   output in a pane), designed in thread 5, never as the transport.

### 15.2 Facts and how each was checked

| Fact | Source |
| --- | --- |
| `claude` 2.1.332 has `-p`, `--session-id`, `--resume`, `--output-format json`, `--json-schema`, `--permission-mode auto`, `--max-budget-usd`, `--permission-prompts none`, `--bg`, `--cloud` | `claude --help` on the MacBook |
| `claude agents --json` lists every session with `kind`, `status` (idle, busy) or `state` (done, stopped), `cwd`, `pid`, `sessionId` | Run on the MacBook |
| `codex` 0.172.0 has `exec`, `exec resume <id>`, `--json`, `-o`, `--output-schema`, `-s` | `codex --help` on the MacBook |
| `-p` shows no trust dialog; in auto mode a refused action fails and the turn continues | code.claude.com/docs/en/headless.md, permission-modes.md (delegated read) |
| The JSON result has `session_id`, `result`, `total_cost_usd`; `is_error` and `num_turns` unconfirmed | Same |
| A `-p` session resumes headless or interactive by ID | code.claude.com/docs/en/sessions.md |
| Transcripts are `~/.claude/projects/<project>/<session-id>.jsonl`; no documented move between machines | Same |
| `claude -p` and the Agent SDK draw from the subscription's limits. A separate credit announced for 2026-06-15 was paused; a 2026-10-07 update says nothing changes | support.claude.com/en/articles/15036540 |
| `claude setup-token` gives a one-year token that can only make model requests; the GitHub Action accepts it and bills the subscription | code.claude.com/docs/en/authentication.md, github-actions.md |
| Cloud sessions are on Max, count toward plan limits, start with `claude --cloud "<prompt>"`; routines are a research preview | code.claude.com/docs/en/claude-code-on-the-web.md, routines.md |
| A hosted Actions job runs at most 6 hours | docs.github.com/en/actions/reference/limits |
| Linux hosted minutes about $0.006; self-hosted not charged | Secondary sources only (runs-on.com and others) |
| Codex in CI: an API key is the documented path; the ChatGPT login wants a persistent `CODEX_HOME` | developers.openai.com/codex/noninteractive, /codex/auth/ci-cd-auth |
| `mascah/keyborg` is private with no rulesets; `mascah/tend` has one ruleset, `main` | `gh` on 2026-11-26 |
| Rulesets and protected branches on a private personal repository need GitHub Pro; merge queues need an organization. Bypass lists do not (corrected in 18.3) | docs.github.com, about protected branches and rulesets |

### 15.3 What headless changes in Tend

- Claude: `claude -p --session-id <uuid> --permission-mode auto
  --output-format json`, then `--resume` for repair and requested changes.
- Codex: `codex exec -s workspace-write --output-schema <schema> -o <file>`.
  The verdict becomes the process's structured output, captured by Tend, so
  the reviewer needs no write access to the run folder. This narrows
  finding 4; it is no boundary under one OS user.
- Each ask runs detached and writes its result to a file, so a turn
  outlives a killed Tend as a pane does now.
- Goes: `herdr.go`, `rex.go`, `terminal.go` (547 lines), dialog answering,
  the `blocked` and `unknown` states, the was-it-delivered ambiguity
  (`flow_run.go:248-254`).
- Lost: typing into a running turn. Instead: tail its output, or stop it
  and resume the session by hand.
- Keep the agent contract to three calls (start or resume with a prompt,
  is it running, its result) so a pane transport could return if headless
  use is ever billed apart. Do not build one now.
