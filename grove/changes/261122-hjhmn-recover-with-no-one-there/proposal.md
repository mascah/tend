# Recover with no one there

## Problem

When a tend dies, or an agent's pane goes, nothing carries the attempt on
until a person runs tend again, and some of those reruns stop for a person by
rule.

- A tend that died mid-turn leaves `status` saying `working`. `tend status`
  then says "tend stopped before this finished" and to run it again
  (watch-and-status spec at 40ca101, "Status of one work item"). `watch` in
  `main.go` at 40ca101 runs tend again only on an attempt with a change
  request, so nothing does.
- A step that was asked and left no result stops for a person on every run:
  "a step recorded as asked with no result and no agent working shall stop
  for a person on every run, until the person supplies the result or deletes
  the record" (flows spec at 40ca101, "Any step recovers without being
  prompted twice"). The review spec's "An uncertain repair goes to a person",
  the publication spec's "A request answered with no change" and the intake
  spec's "Drafting prompts are sent once" say the same of their steps. The
  rule exists so that no prompt is acted on twice.
- After a reboot, or when the terminal that holds the agents restarts, every
  agent is gone. Each attempt that was mid-turn then waits for a person, and
  with several attempts running unattended that is all of them.
- tend can tell an agent that is gone from one that is idle: the terminal
  gives it no pane (`pane` in `herdr.go` and `rex.go` at 40ca101), and tend
  already starts a new session of the same name for it (terminals spec at
  40ca101, "How agents start": "if an agent's pane is gone, a new session
  shall take its name and find the work in Git").

[direction.md](https://github.com/mascah/tend/blob/main/docs/work/261121-autonomous-direction/direction.md)
lists this as open: "How tend detects a lost worker, reattaches to a
surviving session and retries from a durable step. Automatic local recovery
matters for unattended work."

## Intended outcome

`tend watch` runs again an attempt whose tend died. A surviving agent is
waited for and not prompted again, as today. A step whose agent is gone is
asked once more in a new session, which finds the work in Git, since nothing
is left that could act on the earlier prompt; if that second ask also leaves
nothing, a person is needed. An agent that is there but idle, blocked or
unknown still stops for a person. After a reboot, attempts carry on without
anyone running them.

## Scope

### In scope

- `tend watch` running again an attempt whose `status` says `working` and
  that no tend is on.
- A step recorded as asked, with no result, whose agent has no pane: asked
  once more, for any step of a flow, the repair, requested changes and
  intake's follow-up.
- The README.

### Out of scope

- An agent that is there and idle with no result, blocked, or in an unknown
  state. It stays a person's, as a prompt it received may have been acted on.
- Another machine taking over an attempt, a lease on a claim, and
  `grove claim --retry` run by tend. A lost machine's attempt stays a
  person's to retry.
- An attempt whose worktree is gone.
- The drafting prompt of intake, whose step has no result to miss.
- A way to stop watch from running an attempt again other than the attempt
  needing a person.

## Decisions

The owner settled the first in an explore on 2026-11-21, by choosing the
option quoted, and answered "Yes" to the round of decisions before writing,
which holds the rest.

1. When a step is asked again with no person: "Only when its agent is gone".
   "If the step's agent no longer has a pane, nothing can act on the earlier
   prompt, so tend asks once more in a new session, which finds the work in
   Git. If that also leaves nothing, a person. An agent that is idle, blocked
   or unknown with no result stays a person's." The question said that in
   every option "tend watch runs again an attempt whose tend died, and a
   surviving agent is waited for, as today."
2. In every project, watch runs again an attempt whose status says
   `working`, that no tend is on, and that Grove shows in flight. On upgrade,
   an attempt left that way is run again on watch's first pass. Reason: it is
   what `tend status` already tells a person to do.
3. The second ask covers the repair, requested changes and intake's
   follow-up, as well as any step of a project's flow.
4. This change depends on 261122-rpkqa, so that it is built after landing
   and its account of what watch runs again is written once.
5. From a review on 2026-11-22, approved by the owner: the second ask does
   not count toward a step's limit. Reason: the built-in repair has limit 1,
   and a repair whose agent went before it committed would otherwise be both
   asked again and refused.
6. Test layout, private helpers, file names under TEND_STATE that no spec
   names, and the wording of one-sentence messages are the implementer's to
   choose.

## Open questions

None.

## Related documents

- Requirements: [specs/flows/spec.md](specs/flows/spec.md),
  [specs/review/spec.md](specs/review/spec.md),
  [specs/publication/spec.md](specs/publication/spec.md),
  [specs/intake/spec.md](specs/intake/spec.md),
  [specs/watch-and-status/spec.md](specs/watch-and-status/spec.md)
- Tasks: [tasks.md](tasks.md)
