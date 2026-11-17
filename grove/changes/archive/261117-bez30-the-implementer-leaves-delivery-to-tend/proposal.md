# The implementer leaves delivery to tend

## Problem

On 2026-11-16 tend claimed memex cycle 261116-dmj60. memex's
`rules.skills.implement` told the agent to open a stacked pull request per
change, so the implementer made nested worktrees and branches, pushed them
and opened #141 to #143 itself. The attempt's own branch held no commits;
Codex never reviewed the work, and the attempt stalled until the 30-minute
timeout (`docs/work/261116-runner-contract-and-pipeline/design.md`, §1).
Grove gives delivery to the project's rules, and tend gave its implementer
nothing that outranked them.

## Outcome

Whatever a project's rules say, an attempt's work stays committed on the
attempt's branch in its worktree, where tend reviews and publishes it. An
implementer that leaves work elsewhere stops the attempt for a person,
naming what it found, and nothing is reviewed or published.

## Scope

Following design §5 (A1–A5):

- The implementer is prompted `/grove-implement --runner <cycle-id>`, so
  Grove leaves out `rules.delivery` (grove 261117-w31az, Grove 0.12.0).
- The implementer and the drafter start with tend's terms in their system
  prompt: work on this branch in this worktree, push nothing, open or edit
  no pull request, create no branch or worktree, and this wins over any
  other instruction about delivering it; the implementer is also told to
  commit. A system prompt carries it rather than prompt text, as text after
  a skill's name is that skill's arguments.
- Both start with `git push` and `gh pr` denied, a guard against mistakes.
- Before each implementer prompt, tend records the repository's worktrees
  and local branches; after a turn that builds, repairs or makes requested
  changes, and on every run before review, it compares.
- `TIMEOUT_MS` defaults to 4 hours.
- Tidy also removes the worktree of an attempt with no pull request whose
  cycle is done and whose branch the target holds. Neither path removes one
  with a worktree inside it or whose implementer is working.

Out of scope: credential separation, which the design defers (§5, "Known
ceiling"); tend merging anything.

## Decisions

1. The design's Fix A as written. The owner, on the reviewed design: "Yes
   cleanup and ok to the rest", then "ok".
2. Local state is compared, not GitHub pull requests, so tend's own
   proposals cannot be mistaken for the implementer's (design §5, A3).
3. The deny rules are a guard, not a boundary; the required `tend/review`
   check is the gate (design §5, A2).

## Open questions

None.

## Related documents

- [attempts](specs/attempts/spec.md), [terminals](specs/terminals/spec.md),
  [watch-and-status](specs/watch-and-status/spec.md)
- [tasks](tasks.md)
