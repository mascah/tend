# Implementation: Recover with no one there

<!-- Progress on change 261122-hjhmn, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — A step whose agent is gone is asked once more

Status: done

`flow.go`: `again` asks a step once more when its ask left no result and
`term.pane` finds no pane for its agent, in both places a rerun used to stop
for a person: `reachAt` (verdict and file steps) and `reachInOrder` (commit
steps, the repair, requested changes and intake's follow-up, which run there
as the propose flow's `follow` step). `promptStep` now takes the head to
record, so the second ask is recorded by the head of the ask it repeats in
`again-<record>` beside the step's record (`again-check-asked`,
`again-repair-asked`, `again-request-<id>`, `again-followed`). That file is
not the step's record, so `timesAsked` never counts it toward a limit, and
`request-*` does not match it. An ask whose head is already in it is never
repeated: tend stops for a person, saying the step was "asked twice". An
agent still working is waited for before any of this, as before; blocked and
unknown agents still stop in `promptStep` before a prompt goes; an agent there
and idle still stops, since `again` asks only when the pane is gone.

Tests, all passing (`go test ./...`, 92s): `flow_test.go`
`flow-agent-gone`, `flow-agent-gone-once` and `flow-repair-once-more`;
`tend_test.go` `repair-gone`, `request-gone` and `propose-follow-gone`, with a
`gone` helper that closes an agent's pane in the Herdr stand-in. Each of the
six fails with `flow.go` as it was on main. `request-uncertain`,
`repair-blocked`, `repair-undelivered`, `propose-follow-lost` and
`flow-unknown-outcome` still pass, their agents there and idle.

Check: deleting a step's record after a second ask has it asked afresh, but
`again-<record>` still holds that head. So if that fresh ask's agent is then
gone too, tend stops for a person and does not ask a third time. The change
does not say whether the deletion also resets the second ask. I chose the
reading that never risks a third prompt; deleting `again-<record>` as well
resets it.
Check: a commit step whose agent is gone is asked once more only when its ask
left no commit (or no reply file), or, for the build's implementation, when
Git shows the candidate unfinished. Those are the same conditions under which
a rerun stopped before.
Check: T-01's `flow.go`, `main.go`, `flow_test.go` and `tend_test.go` changed
on main since the change's base in 261122-rpkqa (landing) and the two changes
before it. I built on the files as main now has them. Landing's `land` step
has no agent and prompts none, so the second ask never applies to it.

### T-02 — Watch runs again an attempt whose tend died

Status: done

`main.go`: `rerun` has a new reason, `stopped`: the attempt's `status` says
`working`, no tend holds its lock (`held`), and, for a Grove cycle or change,
`grove changes --json` lists the run folder's attempt with `in_flight: true`.
Watch then says "tend stopped on attempt <id> before it finished; running tend
on <item>". Both of watch's paths use `rerun`. Under a count, `keep` already
runs reruns within the count and before it starts anything new, and skips an
attempt a tend holds. A new `working` helper reads the status file, and
`status` now uses it too. The pull request's own reasons (changes requested,
approved) are still checked first, so an attempt that has both is named by the
same reason as before.

Tests, all passing (`go test ./...`, 100s): `tend_test.go` `watch-killed`
(a tend is on it, needs a person, after a kill, then finished),
`watch-killed-agents-gone`, `watch-killed-retried` and
`watch-at-once-killed`. Without the change, `watch-killed` and
`watch-killed-agents-gone` fail. With the in-flight check taken out,
`watch-killed-retried` fails. `watch`, `watch-unattended` and
`status-interrupted` still pass.

Check: the first version checked `stopped` before the pull request's reasons.
`land-rerun-after-kill` then failed: it expects the reason "was approved" for
an attempt whose tend was also killed. So the pull request's reasons now come
first and `stopped` is the fallback. The change does not say which reason
watch gives when both hold.
Check: T-02's `main.go` and `tend_test.go` changed on main since the change's
base, in landing (261122-rpkqa) and several attempts at once (261122-9vbmc).
Those changes brought `keep` and its count. This task builds on them: the
dead-tend rerun goes through `rerun`, which `keep` already calls within the
count.
Found: a tend that exits through `must` (exit 1, an unexpected command
failure) leaves `status` saying `working`. Watch now runs that attempt again
on every pass until it gets further or stops for a person. I saw this by
reading `must` in `main.go`. The change rules out any other way to stop watch
here.

### T-03 — The README says what recovers by itself

Status: done

`README.md`, in four places. The recovery list ("After exit 3, or a kill")
now says that a step whose agent is gone is asked once more in a new session,
recorded in `again-<record>` and not counted toward the step's limit. It says
that an agent there and idle with nothing to show still stops for you, and
that the repair and requested changes stop only when the implementer is
there. The `tend watch` paragraph adds attempts whose tend died, plus a
paragraph on what carries on with no one there: what still stops for a person
(idle, blocked, unknown, or a second ask that left nothing), and that the
first pass after upgrading runs again any attempt left `working`. "How work
reaches tend" gains a fifth kind of work, "A tend that died". The flows
section's "never asks a step twice" now reads "a second time only when its
agent is gone, never a third".

Check: whether the README reads well is for a person to judge. Checked by
reading it against the task's acceptance; every point it lists is stated.

## Scenarios

- flows "A second repair": flow_test.go TestFlows/flow-second-repair, passed
- flows "The built-in one repair": flow_test.go TestFlows/flow-one-repair, passed
- flows "A repair asked once more": flow_test.go TestFlows/flow-repair-once-more, passed
- flows "tend is killed during a project's own step": flow_test.go TestFlows/flow-killed-in-step, passed
- flows "The outcome of a step is unknown": flow_test.go TestFlows/flow-unknown-outcome, passed
- flows "The agent is gone": flow_test.go TestFlows/flow-agent-gone-once, passed
- flows "Gone a second time": flow_test.go TestFlows/flow-agent-gone, passed
- intake "The follow-up stalled": tend_test.go TestTend/propose-follow-lost, passed
- intake "The drafter outlived a timeout": tend_test.go TestTend/propose-survivor, passed
- intake "The drafter is gone": tend_test.go TestTend/propose-follow-gone, passed
- publication "Out of scope": tend_test.go TestTend/request-declined, passed
- publication "A request prompt that stalled": tend_test.go TestTend/request-uncertain, passed
- publication "The implementer is gone": tend_test.go TestTend/request-gone, passed
- review "A stalled repair prompt": tend_test.go TestTend/repair-undelivered, passed
- review "The mark deleted": tend_test.go TestTend/repair-undelivered, passed
- review "The implementer still repairing": tend_test.go TestTend/survivor-repairing, passed
- review "A blocked implementer": tend_test.go TestTend/repair-blocked, passed
- review "The implementer is gone": tend_test.go TestTend/repair-gone, passed
- watch-and-status "After a kill": tend_test.go TestTend/watch-killed, passed
- watch-and-status "After the agents went too": tend_test.go TestTend/watch-killed-agents-gone, passed
- watch-and-status "A tend is on it": tend_test.go TestTend/watch-killed, passed
- watch-and-status "An attempt that needs a person": tend_test.go TestTend/watch-killed, passed
- watch-and-status "An attempt retried since": tend_test.go TestTend/watch-killed-retried, passed

## Reconciliation
