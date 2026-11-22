# Implementation: Land reviewed candidates one at a time

<!-- Progress on change 261122-rpkqa, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — A flow can hold land

Status: done

`flow.go`: `land` has the outcomes `ok`, `conflict` and `failed`; `publish`
in `build` gains `approved`. A `land` step with anything but `next`, a `land`
with no `publish`, a `land` in the `propose` flow and a step continuing
`land` are refused. A loop through `publish`'s `approved` needs no limit, as
one through `requested` does. tend ships `flows/prompts/conflict.md` and
`flows/prompts/failed.md`, the latter filled with `{checks}`, the failed
checks' names and links. The built-in flows are unchanged, so `tend flow`
with no flow file prints what it did.

Checks: `go test -run 'TestLandFaults|TestBuiltInFlowsAreValid|TestFlowFaults|TestRequestedLoopNeedsNoLimit'`
and `go test -run 'TestFlows/flow-land-faults$'`, passed.

Check: a `land` step in the `propose` flow is refused ("where tend lands nothing"); the change places `land` only in `build` and says nothing of `propose`.
Check: T-01's `flow.go` and `flow_test.go` changed on main since base 40ca101 (changes 261122-9vbmc, several attempts at once, and 261122-fh9h2, write access); built on them as they are now, which only added the `watch` setting and the writer rule this task reuses.
Check: T-01 to T-04 are one commit: `flow.go` calls code in `main.go` that T-02 and T-03 add, so no earlier split builds.

### T-02 — tend lands a candidate

Status: done

`main.go` `landStep`: stops for a person when origin is not listed (before
any `gh`), when the head neither passed review nor is carried, when the pull
request is closed or its head is not the branch's; gives `ok` for a merged
pull request. It takes a per-repository flock (`locks/<cksum>-land`, left out
of watch's count), saying in `status` that it waits to land. Each round reads
the pull request, the head and the fetched target afresh: it merges the
target in with `git merge --no-ff` and a plain push (a conflict is aborted
and gives `conflict`), reads check runs and commit statuses through
`gh api .../commits/<sha>/check-runs` and `.../status`, and once all pass
runs `gh pr merge --squash --match-head-commit <head> --subject "<title> (#<n>)" --body-file -`
as the publishing account. A refusal after which the target lacks in the
head starts again; one naming a required check expected, pending or in
progress waits within the same deadline; any other stops quoting GitHub.
`status` says `done: landed <pr>`. `testdata/bin/gh` gains check runs from a
`ci` file, the combined status, `pr merge` and the knobs the scenarios need.

Checks: `go test -run TestLanding` (21 scenarios) and the full `go test ./...`, passed.

Check: the open question, how long `land` waits for a first check, is built as the change says: the `TIMEOUT_MS` bound (four hours by default), polled every 20 seconds or a quarter of it if less. A repository with no CI therefore holds the landing lock that long before tend stops, and every other candidate of it waits behind; a shorter first-check wait would need a decision.
Check: "A person merged it first" is shown by a merge made while the rerun waits for a check. A rerun started after the merge never reaches `land`: the run's tidy removes the attempt's own merged worktree first, as watch-and-status's "a tend's own lock shall not keep it from tidying" says (see Found).
Check: whether GitHub refused because the target or head moved is judged from Git after a fresh fetch, not from GitHub's wording; only the required-check case reads the message, matched as `required status check … expected|pending|in progress`, which could not be checked against a live GitHub.
Check: a pull request whose head differs from the attempt's branch stops for a person, rather than starting again from the merge, as the requirement "Land only a head that passed review" asks the head to be the pull request's.
Check: no real GitHub call was made; `gh pr merge` flags, the check-runs and status endpoints and GitHub's refusal texts are as documented, seen only through the stand-in.
Check: T-02's `main.go`, `flow.go`, `testdata/bin/gh` and `tend_test.go` changed on main since base; built on them as they are now.
Also changed: `testdata/bin/turn.sh` — the implementer stand-in answers tend's conflict prompt by recording how the branch was left, then merging the target keeping its own side.
Found: `tend <repo> <id>` run again after a person merged the attempt's pull request exits 3 saying the work item is in flight: tidy removes the attempt's worktree, then the attempt claims again and Grove refuses; seen debugging the scenario land-no-ci.

### T-03 — A clean merge of the target keeps the verdict

Status: done

`carried(h)` reads from Git alone that `h` is a merge commit whose second
parent is on the target, whose tree is `git merge-tree --write-tree` of its
parents, and whose first parent passed review (a valid success verdict of
the flow's verdict steps) or is itself carried. A verdict step reached at a
carried head is judged by the reviewed head's verdict, so no reviewer is
started. `carry` sets `tend/review` success on it, "Carried from <short>,
which passed independent review, by a clean merge of <target>", once
(`carried-<head>` in the run folder), from `land` and from `publish` on a
rerun. `publish` does not redescribe a head carried from the described one,
and posts no review for it.

Checks: `TestLanding/land-target-moved`, `land-rerun-after-kill`, `land-merge-edit`, passed.

Check: T-03's `main.go`, `flow.go` and `tend_test.go` changed on main since base; built on them as they are now.

### T-04 — An approval sends a candidate on

Status: done

`publish` gives `approved`, where the flow names it, when the newest review
that approves or requests changes, by a person with write access (the
`writer` rule of 261122-fh9h2), on the head or a head it is carried from,
approves; a change request still comes first. Each attempt's run records
`approves` when its flow's `publish` names `approved`, and `tend watch`
(`rerun`, used by both the plain pass and `keep`) runs such an attempt again
when the newest such review on its open pull request's head approves,
saying `<pr> was approved`, unless it needs a person.

Checks: `TestLanding/land-target-moved` (first run merges nothing, the
approved run lands), `land-check-fails` (a new head waits for a new
approval), `approve-stranger`, `approve-built-in`, `approve-watch`,
`approve-needs-you`, passed.

Check: watch knows a flow acts on approvals from the `approves` mark the attempt's last run wrote, not by reading the target's flow file, and looks for an approval on the pull request's head only, not on heads it is carried from; a flow changed since that run is seen at the attempt's next run.
Check: watch under a count runs approved attempts with requested ones, inside the count and before new pieces, by the same code; no scenario shows the count with an approval.
Check: T-04's files changed on main since base; built on them as they are now.

### T-05 — The README and the publication spec say when tend merges

Status: done

`README.md` gains "Landing: merging one candidate at a time" under Flows:
the three steps, one landing per CI run, no CI no landing, the supervised
and autonomous flow files with limited `resolve` and `fix` steps, the
branch rule, bypassing, dismissed approvals and the approval read once. Its
"tend never merges" line now names `land`. The publication spec's
Boundaries says tend merges only at a `land` step, as the landing spec says.

Check: how the README reads is for a person to judge.
Check: T-05's `README.md` and `grove/specs/publication/spec.md` changed on main since base; edited as they are now.

## Scenarios

- flows "A build that lands": tend_test.go TestLanding/land-landed, passed
- flows "A flow without land": tend_test.go TestLanding/land-none, passed
- flows "Land without publish": flow_test.go TestFlows/flow-land-faults and TestLandFaults, passed
- flows "Land given a prompt": flow_test.go TestFlows/flow-land-faults and TestLandFaults, passed
- flows "A conflict step with no limit": flow_test.go TestFlows/flow-land-faults and TestLandFaults, passed
- flows "A flow names tend's own prompts": flow_test.go TestFlows/flow-land-faults, and TestLanding/land-conflict and land-check-fails prompted with tend's text, passed
- landing "A failed head reaches land": TestLanding/land-failed-head, passed
- landing "An origin not listed": TestLanding/land-unlisted, passed
- landing "A person merged it first": TestLanding/land-no-ci, merged while the rerun waits, passed
- landing "Two candidates ready": TestLanding/land-two, passed
- landing "The target moved": TestLanding/land-target-moved, passed
- landing "The target did not move": TestLanding/land-landed, passed
- landing "A conflict": TestLanding/land-conflict, passed
- landing "A carried head": TestLanding/land-target-moved, passed
- landing "A rerun after a kill": TestLanding/land-rerun-after-kill, passed
- landing "A merge with an edit of its own": TestLanding/land-merge-edit, passed
- landing "A check fails on the merged head": TestLanding/land-check-fails, passed
- landing "A cancelled check": TestLanding/land-cancelled, passed
- landing "A skipped check": TestLanding/land-skipped, passed
- landing "No CI": TestLanding/land-no-ci, passed
- landing "The checks pass": TestLanding/land-landed, passed
- landing "Landed": TestLanding/land-landed, passed
- landing "The target moved while the checks ran": TestLanding/land-moved-before-merge, passed
- landing "A required check reports late": TestLanding/land-late-check, passed
- landing "A description a person edited": TestLanding/land-edited-description, passed
- landing "GitHub refuses for a rule": TestLanding/land-refused, passed
- landing "A teammate approves": TestLanding/land-target-moved, passed
- landing "A stranger approves": TestLanding/approve-stranger, passed
- landing "A new head after the approval": TestLanding/land-check-fails, passed
- landing "The built-in flow": TestLanding/approve-built-in, passed
- publication "A new head after a requested change": TestTend/request, passed
- publication "A description a person edited": TestTend/request-edited, passed
- publication "A carried head": TestLanding/land-rerun-after-kill, passed
- review "The reviewer's one writable folder": TestTend/remote-claim, passed
- review "A reviewer left working": TestFlows/flow-reviewer-writing, passed
- review "A carried head": TestLanding/land-rerun-after-kill, passed
- watch-and-status "A teammate approves": TestLanding/approve-watch, passed
- watch-and-status "The built-in flow": TestLanding/approve-built-in, passed
- watch-and-status "An attempt that needs a person": TestLanding/approve-needs-you, passed

## Reconciliation

<!-- One line per change spec, once applied to its maintained spec:

    - record-editing: applied to specs/record-editing/spec.md
-->
