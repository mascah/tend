# Review

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

Each candidate head of an attempt is reviewed by a fresh Codex session apart from the implementer, whose only output is a verdict file tend validates, and a failure the reviewer judges repairable gets one repair in the implementer's conversation before the new head is reviewed. A verdict is `verdict-<head>.json` in the attempt's run folder; actionable means every finding that fails the head is an in-scope code fix that needs no decision from the owner. The owner, and anyone reading a published review, rely on a verdict meaning a review of exactly that commit happened.

## Boundaries

This covers starting and prompting the reviewer, what makes a verdict count,
and the one repair with its `repair-asked` mark. When the attempt runs at
all, its recovery in general and its exit statuses are
[attempts](../attempts/spec.md); posting verdicts on GitHub is
[publication](../publication/spec.md); what the reviewer is told an issue
build must do is [work-items](../work-items/spec.md). Whether the reviewer's
judgement is right is not checked.

## Requirements

### Requirement: A fresh reviewer for each head

When an attempt's head has no valid verdict and is not carried from a head
that passed review, as the landing spec's "The verdict carries across a clean
merge" says, the system shall start reviewer `rev-<short head>` as Codex with
`-s workspace-write -c sandbox_workspace_write.network_access=true -a never
--add-dir <run folder>`, reusing one an earlier run left, and prompt it once
with the branch, the head, the change as `git diff <base>...HEAD`, what the
work item must do, where the implementer's claims are and the attempt's
earlier verdicts, asking it to fail the head only as "Fail a head only for
findings that matter" says, for no edits or commits, and for the verdict
written to `verdict-<head>.json` in the run folder as `head`, `conclusion`,
`actionable`, `summary` and `notes`; and once the verdict is valid it shall
close the reviewer's pane.

#### Scenario: The reviewer's one writable folder

- WHEN tend has a cycle's head reviewed
- THEN the reviewer is started with `--add-dir TEND_STATE/runs/<attempt>`,
  and the verdict file is there

#### Scenario: A reviewer left working

- GIVEN tend stopped because the reviewer's prompt timed out, and the reviewer
  is still working
- WHEN the owner runs tend again
- THEN the reviewer is started once and prompted once in all

#### Scenario: A carried head

- GIVEN an attempt whose head is a clean merge of the target into a head
  that passed review, made at `land`
- WHEN the owner runs tend again
- THEN no reviewer is started for the merge commit, and the flow goes on as
  from a passing review

### Requirement: Only a valid verdict counts

The system shall count a verdict, new or cached, only when its file is valid
JSON naming the head it is for, with `conclusion` `success` or `failure`,
`actionable` a boolean, `summary` a string, and `notes`, where present, a
string; if a cached verdict is not valid, it shall keep it as
`verdict-<head>.json.rejected-<time>` and review the head again; and if the
reviewer settles without writing a valid verdict, it shall stop for a person,
naming the reviewer's pane.

#### Scenario: A cached verdict for another head

- GIVEN a reviewed attempt whose verdict file now names head `0000…0000`
- WHEN the owner runs tend again
- THEN the head is reviewed a second time and the verdict names the head

#### Scenario: A cached verdict without a conclusion

- GIVEN a reviewed attempt whose verdict file has lost its `conclusion`
- WHEN the owner runs tend again
- THEN the head is reviewed again and the set-aside file is kept as
  `verdict-<head>.json.rejected-<time>`

#### Scenario: A new verdict without a conclusion

- GIVEN a reviewer that writes a verdict with no `conclusion`
- WHEN the owner runs tend
- THEN tend exits 3, saying the reviewer settled without a valid verdict

#### Scenario: A verdict without notes

- GIVEN a reviewer that writes a passing verdict with no `notes`
- WHEN the owner runs tend
- THEN the verdict counts, and tend exits 0

#### Scenario: Notes that are not text

- GIVEN a reviewer that writes a verdict whose `notes` is a list
- WHEN the owner runs tend
- THEN tend exits 3, saying the reviewer settled without a valid verdict

### Requirement: Fail a head only for findings that matter

The system's built-in review prompt shall ask the reviewer to conclude
`failure` only for a defect a realistic use would hit, a requirement missed
in its main case, a security or data hole that can be reached, or a claim in
the implementer's notes the code does not support; to write every other
finding, such as a corner that needs contrived setup to reach, which of two
reasons a stop gives, wording or polish, in the verdict's `notes` and
conclude `success`; and to judge `actionable` on the findings that fail the
head alone; and when a verdict concludes `success`, the system shall go on
as from a passing review whatever its `notes` hold, passing no note to the
repair.

#### Scenario: The bar in the prompt

- WHEN tend prompts a reviewer with its built-in review prompt
- THEN the prompt names the four kinds of finding that fail a head, says
  that contrived corners, the reason a stop gives, wording and polish do
  not, and asks for those in `notes`

#### Scenario: A passing head with notes

- GIVEN a reviewer that writes a verdict concluding `success` with `notes`
  naming a minor finding
- WHEN the owner runs tend on the cycle
- THEN no repair is asked, `status` says `done: success`, and tend exits 0

#### Scenario: A failure with notes

- GIVEN a reviewer that fails the first head as actionable, with `notes`
  naming a minor finding, and passes the next
- WHEN tend asks the repair
- THEN the repair prompt holds the verdict's `summary` and not its `notes`

#### Scenario: A project's own bar

- GIVEN a project whose flow names `prompts/review.md` and whose
  `tend/prompts/review.md` on its target branch states another bar
- WHEN tend has a head reviewed
- THEN the reviewer is prompted with the project's text, not the built-in
  bar

### Requirement: A later review judges what changed

When tend prompts a review of a head and the attempt holds a valid verdict
for an earlier commit of the branch, the system shall fill `{earlier}` with
each such verdict's head, conclusion and file, and the built-in review prompt
shall ask the reviewer to judge whether the findings that failed an earlier
head are fixed and what the commits since the latest earlier head changed,
and not to fail the head for code unchanged since a head an earlier review
saw, unless it is a security or data hole that can be reached; and when the
attempt holds no such verdict, `{earlier}` shall say there is none.

#### Scenario: The review after a repair

- GIVEN a reviewer that fails the first head as actionable, and a repair
  that commits
- WHEN tend prompts the review of the new head
- THEN the prompt names the first head, its conclusion `failure` and its
  verdict file, and asks for the commits since that head to be judged

#### Scenario: The first review

- WHEN tend prompts the first review of an attempt
- THEN the prompt says there is no earlier verdict

#### Scenario: A verdict for a commit no longer on the branch

- GIVEN an attempt whose run folder holds a verdict for a commit the branch
  no longer contains
- WHEN tend prompts the review of the head
- THEN that verdict is not named in the prompt

### Requirement: The reviewer changes nothing

If the head moved or the tree has uncommitted files once the reviewer's turn
ends, then the system shall stop for a person, saying the reviewer changed the
checkout, and publish nothing for that head.

#### Scenario: A reviewer that commits

- GIVEN a reviewer that commits a fix of its own as well as its verdict
- WHEN its turn ends
- THEN tend exits 3, saying `rev-<short head>` changed the checkout

### Requirement: One repair

When the head's verdict is a failure marked actionable and the attempt has
asked no repair, the system shall write the reviewed head to `repair-asked`
in the run folder just before prompting the implementer once with the
verdict's findings, stop for a person if no commit follows, and have the new
head reviewed; it shall ask no second repair in the attempt, and a failure
that is not actionable, or that remains after the repair, shall exit 0 with
that verdict for a person to decide.

#### Scenario: Repaired and passed

- GIVEN a reviewer that fails the first head as actionable and passes the next
- WHEN the owner runs tend on the cycle
- THEN the repair is asked once, the new head passes, `status` says `done:
  success`, and tend exits 0

#### Scenario: A finding that needs a decision

- GIVEN a reviewer that fails the head with `actionable` false
- WHEN tend has it reviewed
- THEN no repair is asked, and tend exits 0 printing the failed verdict

### Requirement: An uncertain repair goes to a person

If `repair-asked` names the head under review, no commit has followed and the
implementer is there and not working, then the system shall stop for a person
on every run, saying the repair's outcome is unknown, until a person commits
a fix or deletes `repair-asked`, which has the repair asked once more; an
implementer that is gone shall be asked the repair once more in a new
session, as the flows spec's "Any step recovers without being prompted twice"
says; an implementer still working on the repair shall be waited for; and a
repair refused before its prompt went, as by a blocked implementer, shall
leave the repair unused.

#### Scenario: A stalled repair prompt

- GIVEN tend stopped with exit 3 because the repair prompt stalled, and the
  implementer is idle with no new commit
- WHEN the owner runs tend again
- THEN it exits 3 without asking the repair again, saying the repair has no
  commit and the implementer is not working on it, and `tend status` shows
  `needs-you: the repair asked`

#### Scenario: The mark deleted

- GIVEN that stopped attempt, and a person deletes `repair-asked`
- WHEN the owner runs tend again
- THEN the repair is asked a second time, the new head passes, and tend exits 0

#### Scenario: The implementer still repairing

- GIVEN the repair prompt timed out and the implementer is still working
- WHEN the owner runs tend again
- THEN tend waits for it, has the new head reviewed, exits 0, and the repair
  was asked once

#### Scenario: A blocked implementer

- GIVEN the implementer is blocked when the head fails as actionable
- WHEN tend would ask the repair
- THEN it exits 3 with no `repair-asked` written; once the implementer is
  idle, the next run asks the repair and passes

#### Scenario: The implementer is gone

- GIVEN tend was killed after asking the repair, no commit has followed, and
  the implementer's pane is closed
- WHEN the owner runs tend again
- THEN the repair is asked once more of a new implementer session, the new
  head is reviewed, and tend exits 0

## Known discrepancies

None recorded.

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
- [261122-rpkqa](../../changes/archive/261122-rpkqa-land-reviewed-candidates-one-at-a-time/) Land reviewed candidates one at a time
- [261122-hjhmn](../../changes/archive/261122-hjhmn-recover-with-no-one-there/) Recover with no one there
- [261126-5xkrh](../../changes/archive/261126-5xkrh-fail-a-head-only-for-findings-that-matter/) Fail a head in review only for findings that matter
