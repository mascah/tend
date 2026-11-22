# Several attempts at once

## Problem

One tend runs at a time on a runner, and `tend watch` starts one cycle at a
time, though Grove now lets pieces of work that share nothing be claimed
together.

- The attempts spec at 40ca101, "One tend per runner", refuses a second
  `tend <repo> <item>` or `tend propose <repo>` while any tend holds the lock
  in TEND_STATE. `takeLock` in `main.go` at 40ca101 does it, under the note
  "one lock for the runner, so one attempt at a time; lock per attempt when
  capacity grows".
- `watch` in `main.go` at 40ca101 runs each tend and waits for it to exit
  before anything else, and starts only "the oldest cycle that Grove would
  claim (`runs` in `grove changes --json`) and whose plan marks it
  `unattended`" (watch-and-status spec at 40ca101, "Watch acts on a person's
  behalf"). Grove's `runs` is false while anything at all is in flight.
- The attempts spec at 40ca101, "Work stays on the attempt's branch", stops
  an attempt for any worktree, or branch with commits, that the repository
  "did not have just before the implementer was last prompted". `contained`
  in `main.go` at 40ca101 compares the two layouts. A second attempt that
  starts during the turn adds both.
- `attempt` in `main.go` at 40ca101 makes a worktree on a `tend-HHMMSS`
  branch, runs `grove claim` in it and, when Grove refuses, stops with the
  worktree and branch still there. With several tends claiming, a refusal
  for overlap is an ordinary event.
- Grove 0.14.0 records attempts at pieces that share no planned file at the
  same time, each in its own checkout, and says for each cycle and each
  change in no cycle whether `grove claim` would claim it now
  ([Grove's runs spec at v0.14.0](https://github.com/mascah/grove/blob/v0.14.0/grove/specs/runs/spec.md),
  "Refuse what is in flight or overlaps it").

[direction.md](https://github.com/mascah/tend/blob/main/docs/work/261121-autonomous-direction/direction.md),
decision 8, puts "How many pieces run at once, and which first" in the flow
file, and decision 9 has several pieces run at once, "easily controlled by
the person running it".

## Intended outcome

Tends on different work items run at the same time, each with its own lock,
worktree and agents, and an intake pass runs beside them. A project that
sets `watch: {at_once: N}` in `tend/flow.yaml` has `tend watch` start every
cycle and change Grove calls claimable, oldest first, keeping at most N
attempts being worked on; the person running watch can lower that number for
their machine. A project without the setting sees no change. An attempt that
stops for a person frees its slot, so work that does not overlap it goes on.

## Scope

### In scope

- A lock for each work item of a repository and one for each repository's
  intake pass, in place of the runner's one lock; what exit 4 means; and the
  usage line, which becomes `usage: tend <repo> <id>`.
- Tends that start together setting up one at a time.
- The temporary branch a new worktree is made with, named for the time and
  tend's process (`tend-HHMMSS-<pid>`) in place of the time alone, so that
  tends setting up in the same second each make their own worktree. Added in
  implementation, on 2026-11-22: two tends started in the same second made
  the same `tend-HHMMSS` worktree.
- A claim Grove refuses leaving no worktree or branch behind.
- The containment rule allowing for what tend's other attempts and intake
  passes made, and waiting for one that is setting up.
- Cleanup leaving the worktree of an attempt or intake pass a tend is on.
- `watch.at_once` in the flow file: reading it, refusing a value that is not
  valid, and `tend flow` printing it.
- `tend watch` under that setting: which pieces it starts, in what order,
  the count, `AT_ONCE`, not waiting for a run, and `WATCH_ONCE`.
- The README.

### Out of scope

- Landing what is built: change 261122-rpkqa.
- Running again an attempt whose tend died, and asking a step again when its
  agent is gone: change 261122-hjhmn.
- A count that holds across machines. Each runner counts its own attempts;
  Grove's claim still refuses overlap whichever machine asks.
- A cap on reviewed pull requests that wait for a person.
- Several proposals drafted at once: an intake pass still drafts one request
  after another.
- `tend watch` starting issue builds.
- The daily spend cap of direction decision 3d: step 4.
- Cycles, `/grove-plan` and the `unattended` mark, which work as before in a
  project without the setting.
- A new way to park an attempt. One that stopped for a person keeps its
  change in flight in Grove, so nothing claims it again, and watch leaves it
  until a person runs tend on it.

## Decisions

The owner settled the first four in an explore on 2026-11-21, each by
choosing the option quoted, and answered "Yes" to the round of decisions
before writing, which holds the rest.

1. What the count counts: "Attempts being worked on". "Attempts on this
   runner in that repository with a tend on them: building, in review,
   repairing or in the train. An attempt waiting for a person, or reviewed
   and waiting for a merge, frees its slot, so unrelated work proceeds.
   Reviewed pull requests can pile up while you are away, and two machines
   each count their own." Reason: the direction asks that a change parked
   for the owner not hold up unrelated work.
2. What watch starts by itself: "The flow file opts in". "With no setting,
   watch behaves as today. A setting in tend/flow.yaml makes it start every
   piece Grove calls claimable, oldest first, up to the at-once count, so
   merging a proposal is all it takes. Existing projects see no change on
   upgrade."
3. Where the count is set: "Flow file, runner may lower it". "The person
   running tend watch can set a lower number for their machine with an
   environment variable, like WATCH_SECONDS today, but never a higher one,
   so agents and runners cannot exceed the project's rule."
4. Containment: "Except tend's other attempts". "A worktree or branch that
   another attempt or intake pass of this runner recorded is not the
   implementer's doing; anything else new still stops the attempt. Limit:
   tend cannot tell if an implementer wrote into another attempt's worktree,
   though that work would show in that attempt's own diff and review."
5. One key opts in and sets the count: `watch: {at_once: N}`. There is no
   separate key for which pieces watch starts. Reason: a count has no effect
   under today's pause, so the two would only ever be set together.
6. `AT_ONCE` in watch's environment lowers N and never raises it; 0 starts
   nothing new and lets running work finish.
7. The count is attempts in that repository on this runner whose lock a tend
   holds, including ones run by hand. Intake is not counted, and one pass per
   repository runs beside the builds. Each pass, attempts already begun go
   before new pieces.
8. The lock: one tend per work item per repository, and one intake pass per
   repository. Exit 4 means that item is taken.
9. A claim Grove refuses leaves no worktree or branch behind.
10. A short runner-wide lock is held only while an attempt fetches, tidies,
    makes its worktree and claims.
11. Parking needs nothing new: a stopped attempt keeps its change in flight,
    so nothing claims it again, watch leaves it, its slot is free and only
    overlapping pieces wait. To park a running attempt, a person interrupts
    its agent in its pane.
12. This change depends on 261122-nh8aa, which has tend build a change by
    name.
13. Routine: the usage line becomes `usage: tend <repo> <id>`, and
    `WATCH_ONCE` waits for what it started. Test layout, private helpers,
    file names under TEND_STATE that no spec names, and the wording of
    one-sentence messages are the implementer's to choose.

14. From a review on 2026-11-22, each fix approved by the owner:
    - Cleanup leaves the worktree of an attempt or intake pass a tend is
      on. Reason: with a lock for each work item, one tend's cleanup runs
      while another tend still uses its worktree.
    - A tend records its worktree before it lets the next tend set up, and
      waits for one that is setting up before it checks containment. Reason:
      a worktree made and not yet recorded would stop another attempt for a
      person.

## Open questions

None.

## Related documents

- Requirements: [specs/attempts/spec.md](specs/attempts/spec.md),
  [specs/flows/spec.md](specs/flows/spec.md),
  [specs/watch-and-status/spec.md](specs/watch-and-status/spec.md)
- Tasks: [tasks.md](tasks.md)
