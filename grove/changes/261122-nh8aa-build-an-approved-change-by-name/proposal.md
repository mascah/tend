# Build an approved change by name

## Problem

tend builds only a cycle, so in a tend project every approved change still
needs a cycle planned and merged before anything builds it.

- The attempts spec at 40ca101 opens "`tend <repo> <cycle-id>` takes one
  approved Grove cycle through one attempt". `attempt` in `main.go` at
  40ca101 finds the attempt to carry on with `firstRun(proj, "cycle", cycle)`,
  and `status` does the same, so a record that names no cycle is never found.
- `overview` in `main.go` at 40ca101 prints approved changes under "Approved
  changes without a cycle (plan one: /grove-plan)".
- Grove 0.14.0 claims an approved change that no cycle lists:
  `grove claim <change-id>` records an attempt "naming C alone and no cycle"
  ([Grove's runs spec at v0.14.0](https://github.com/mascah/grove/blob/v0.14.0/grove/specs/runs/spec.md),
  "Start one attempt", scenario "A change in no cycle"), and
  `grove changes --json` gives such a change `claimable` and a `verdict`
  ([Grove's changes spec at v0.14.0](https://github.com/mascah/grove/blob/v0.14.0/grove/specs/changes/spec.md),
  "Print the project as data").
- tend's CI installs Grove v0.13.0 (`.github/workflows/ci.yml` at 40ca101),
  which has neither.

The merge that plans a cycle of one change only schedules work that is
already approved. [direction.md](https://github.com/mascah/tend/blob/main/docs/work/261121-autonomous-direction/direction.md),
decision 8, reads "Approval stays a merge; scheduling stops being one", and
decision 13 "Claim approved changes individually by default". Grove's part of
that is released. This change is the first of tend's.

## Intended outcome

`tend <repo> <change-id>` takes an approved change that no cycle lists
through an attempt exactly as it takes a cycle of that one change: it claims
the change by name, carries the attempt on after a stop, implements, reviews
and publishes it. `tend status`, the overview and the tidying of finished
worktrees treat such a change as they treat a cycle. A person who merges a
proposal can build it with one command, with no cycle planned.

## Scope

### In scope

- `tend <repo> <change-id>` and `tend status <repo> <change-id>` for an
  approved change that no cycle not done lists.
- The overview's list of approved changes, which gives Grove's verdict for
  each in place of telling the reader to plan a cycle, and its list of
  attempts, which includes attempts at such changes.
- Tidying the worktree of an attempt at such a change that reached the target
  some other way.
- The sentence tend's proposal review ends with, which today says the merged
  change "waits for a cycle: `/grove-plan`" (`proposal` in `main.go` at
  40ca101).
- Grove 0.14.0 as the Grove tend needs: CI's pin and the README.

### Out of scope

- `tend watch` starting such a change by itself: change 261122-9vbmc.
- The usage line `usage: tend <repo> <cycle-id>`. It is part of the attempts
  requirement "Exit statuses", which change 261122-9vbmc rewrites for the
  lock, so that change carries the new line and the requirement is rewritten
  once.
- More than one attempt at once: change 261122-9vbmc.
- Cycles. A cycle is still built as one attempt, and a change a cycle lists
  is built with its cycle; Grove refuses a claim of it by name.
- Grove itself.

## Decisions

The owner settled these in an explore on 2026-11-21 and answered "Yes" to the
round of decisions before writing.

1. Step 3 of the direction is split. Asked how, the owner chose "Four, settle
   all first": four changes in order, "(1) build an approved change by name,
   (2) several attempts at once (lock, containment, watch's count and
   selection in the flow file), (3) landing, (4) automatic recovery", drafted
   together once every open question was answered. Reason: each part can be
   reviewed, built and landed on its own.
2. A change in no cycle is "built exactly as a cycle of that one change,
   covering the claim, carrying on, status, overview, tidying and the pull
   request. watch does not start one until change 2's setting." Reason: Grove
   records its attempt as it records a cycle's, so nothing in tend needs a
   second path.
3. "tend needs Grove 0.14.0 and CI's pin moves to `v0.14.0`. With an older
   Grove the claim is refused and tend stops with Grove's sentence." Reason:
   tend adds no check of Grove's version; Grove's refusal already says what
   is wrong.
4. The overview's heading becomes "Approved changes in no cycle", each change
   with Grove's verdict.
5. This change depends on no other.
6. Test layout, private helpers, file names under TEND_STATE that no spec
   names, and the wording of one-sentence messages are the implementer's to
   choose.

## Open questions

None.

## Related documents

- Requirements: [specs/attempts/spec.md](specs/attempts/spec.md),
  [specs/watch-and-status/spec.md](specs/watch-and-status/spec.md)
- Tasks: [tasks.md](tasks.md)
