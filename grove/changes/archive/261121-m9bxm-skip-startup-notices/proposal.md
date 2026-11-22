# Skip startup notices

## Problem

An agent that starts at a screen tend does not know does not start, and tend
stops for a person: the terminals spec at 0a4e821, "The trust dialogs are
answered", says so for any startup failure but the trust dialogs, and
`start` in `main.go` at 0a4e821 does it. Codex shows such a screen whenever it
has a one-off notice.

A supervised run on the owner's MacBook on 2026-11-21 (tend at 43d100a on the
branch of change 261121-gapde, Herdr 0.9.5, codex-cli 0.169.0, Grove's ledger
fixture) met one. The run was not committed; what it showed:

- The reviewer started at Codex's "Update available" screen: "› 1. Update
  now … 2. Skip 3. Skip until next version", with the hint "enter continue ·
  esc skip". Herdr reported the agent blocked, its launch pending, the start
  failed as `agent_not_ready`, and tend exited 3 with
  "rev-139fa3d did not start: agent_not_ready".
- Esc skipped the notice. Codex's folder trust dialog followed, which tend
  answers when it sees it.

The runner on the owner's Mac mini sets `check_for_update_on_startup = false`
for its `tend` user, so it does not see that notice. Codex has no one setting
for the rest: each notice has its own (codex-cli 0.169.0 names
`hide_rate_limit_model_nudge` and `hide_gpt5_1_migration_prompt`, among
others), and a model announcement is counted per model in
`[tui.model_availability_nux]`, where the mini's config holds
`"gpt-6.1-sol" = 4` while Codex now offers GPT-6.2-Sol. An unattended runner
stops at the next notice and every attempt waits until someone answers it.

## Intended outcome

An agent that starts at a notice offering to be skipped is skipped past, as
tend answers the trust dialogs, and the attempt goes on. tend never chooses
anything at such a notice: it does not update the agent or change its model.
A screen it does not know still stops for a person, and the hooks review still
stops with no key sent.

## Scope

### In scope

- The terminals requirement "The trust dialogs are answered", for a startup
  screen whose hint offers `esc skip`, under Herdr and rex and for either
  harness.
- Saying on standard error which agent's notice was skipped.
- A test with a stand-in agent that starts at such a notice, and the README.

### Out of scope

- Updating Codex or Claude Code, or choosing a model at a notice: a person
  does that.
- Any other screen an agent starts at, which still stops for a person.
- The Codex hooks review, which still stops with no key sent.
- The runner's own Codex settings on the mini, which already turn off the
  update check.

## Decisions

The owner answered "yes" to the round of decisions on 2026-11-21:

1. A notice is a startup screen whose hint line offers `esc skip`, and tend
   presses Esc and nothing else at it, so it always skips and never updates
   or picks a model. The screen itself says Esc skips it, and Esc alone can
   choose nothing.
2. Every other screen tend does not know still stops for a person, saying the
   agent did not start, and the hooks review still stops with no key sent.
3. The notice is told by its screen's text, for either harness, not only
   Codex. Claude Code shows no such screen today.
4. tend says on standard error that it skipped an agent's startup notice, so
   a run's output shows it.
5. The change depends on 261121-gapde, which replaces `dev/test.sh` with Go
   tests and stand-ins in `testdata/bin`: the owner chose "a", to build it
   against those after 261121-gapde lands, over testing through `dev/test.sh`
   and revising whichever change lands second.

## Open questions

None.

## Related documents

- Requirements: [specs/terminals/spec.md](specs/terminals/spec.md)
- Tasks: [tasks.md](tasks.md)
