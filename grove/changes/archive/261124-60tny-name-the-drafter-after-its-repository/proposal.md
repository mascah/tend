# An intake drafter is named after its request and its repository

## Observed

An intake pass on one repository prompts another repository's drafter when
both have a request of the same number
([issue #26](https://github.com/mascah/tend/issues/26)).

- `draft` in `intake.go` at 63fc1a0 names the drafter `"prop-" + r.agent`,
  and `r.agent` is the issue's number (`request.go`) or the Jira key in lower
  case (`jira.go`). The repository is no part of the name.
- `startAgent` in `flow_run.go` at 63fc1a0 starts no agent when
  `term.pane(name)` finds one of that name.
- Under Herdr, `pane` in `herdr.go` at 63fc1a0 asks `herdr agent get <name>`,
  which names no workspace: an agent of that name anywhere in Herdr is found.
- Under rex, `pane` in `rex.go` at 63fc1a0 looks only among the blocks of the
  worktree's own session, so the same name in another repository is not
  found. The bug is seen under Herdr only.

Since tend takes a lock for each work item (change 261122-9vbmc), intake
passes on two repositories can run at once. Before that, a pass could
already meet a drafter another repository's pass left open.

## Expected

A pass drafts each request with a drafter of its own repository. The owner
asked for the fix ("Lets fix 26") and answered "yes to all four" to these:

1. The change is a bugfix. Reason: the behavior is wrong, and the spec text
   changes with it.
2. The drafter is named `prop-<n>-<cksum of the repository's path>`, in both
   terminals. Reason: the same number already names an issue build's run
   folder, and one rule is simpler than one for Herdr alone.
3. A drafter still open under the old name when tend is updated is not
   found: tend treats it as gone and asks once more in a new session.
   Reason: falling back to the old name would keep the collision.
4. The proposal and the fix reach the target in one pull request, so one
   merge approves the change with its code.

## Reproduction

From a clean checkout at 63fc1a0, with Go and Grove installed, add a scenario
that runs `tend propose` on a repository with issue #1 labelled `tend`, then
on a second repository with its own issue #1, GitHub state and `TEND_STATE`,
in a Herdr stand-in where the first pass's agents are still open. The first
pass's Herdr calls were:

```text
agent start prop-1 --kind claude --pane p0 ...
agent prompt prop-1 /grove-propose --unattended Request 1. Make thing 1 better. (GitHub issue https://example.test/issues/1)
agent prompt prop-1 Without asking anyone: if you wrote a change, commit everything ...
```

and the second pass's:

```text
agent prompt prop-1 /grove-propose --unattended Request 1. Make thing 1 better. (GitHub issue https://example.test/issues/1)
agent prompt prop-1 Without asking anyone: if you wrote a change, commit everything ...
```

The second pass starts no agent and sends its request to the first
repository's `prop-1`. This was run with the stand-ins in `testdata/bin`,
not against a live Herdr.

## The requirement it breaks

None. The code does what the maintained spec says, and the spec is what is
wrong: intake, "Draft, commit and open the proposal" (63fc1a0), prompts "the
drafter `prop-<n>` or `prop-<key in lower case>`". The owner agreed to change
it; the change spec gives the requirement in full. The terminals spec's
Purpose (63fc1a0) repeats the name as `prop-<request>` and is corrected with
it.

## Open questions

None must be answered before review. Routine choices, for T-01: the
scenario's name and place, and how the fixture gives a second repository the
same Herdr. Recommended: the implementer chooses, and records each choice.

## Related documents

- Tasks: [tasks.md](tasks.md)
- Requirements: [specs/intake/spec.md](specs/intake/spec.md)
