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

Status: not started

### T-03 — The README says what recovers by itself

Status: not started

## Scenarios

## Reconciliation
