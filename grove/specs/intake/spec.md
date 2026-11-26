# Intake

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

`tend propose <repo>` turns each open request labelled `tend`, a GitHub issue of the repository or an issue of a Jira project mapped to it, into a Grove proposal that Claude drafts with `/grove-propose --unattended` and the publishing account opens as a pull request, or into a comment on the request saying why none could be drafted. A request's identity is its issue number or Jira key, and its proposal branch is `tend/issue-<n>` or `tend/jira-<KEY>`. Each request has an intake folder, `TEND_STATE/intake/<owner>_<repo>-<n>` or `…-jira-<KEY>`. People filing requests, and people reviewing the proposals, rely on this; labelling a request hands its text to an agent in auto mode.

## Boundaries

This covers one pass of `tend propose`: which requests are taken, drafting
and following up, the proposal's pull request and summary review, declines,
the cap, and finishing an interrupted handoff. The publishing account and
its token are as [publication](../publication/spec.md) says; the Jira
mapping, credentials and how a Jira description becomes text are
[work-items](../work-items/spec.md); how the drafter is started and prompted
is [terminals](../terminals/spec.md); the runner lock is
[attempts](../attempts/spec.md); `tend watch` running `tend propose` and
tidying finished proposal worktrees are
[watch-and-status](../watch-and-status/spec.md). Approving a proposal
(merging it) and planning its cycle stay with people.

## Requirements

### Requirement: Only listed repositories take requests

If the repository's origin is not listed in `TEND_STATE/publish`, then
`tend propose` shall stop for a person, saying tend takes no requests from
it, and draft nothing.

#### Scenario: An unlisted repository

- GIVEN a repository with origin not in `TEND_STATE/publish`, and issue #1
  labelled `tend`
- WHEN the owner runs `tend propose <repo>`
- THEN it exits 3 saying tend takes no requests from it, and nothing is
  drafted

### Requirement: Each labelled request, once

When `tend propose` runs, the system shall take each open GitHub issue of
the repository labelled `tend`, lowest number first, then each issue labelled
`tend` whose status category is not Done in each Jira project mapped to the
repository's origin, oldest first and through every page of results, and
draft one only while no pull request from its proposal branch exists in any
state; and the pass shall exit 0.

#### Scenario: GitHub issues

- GIVEN issues #1 labelled `tend`, #2 unlabelled and #3 labelled `tend,bug`
- WHEN the owner runs `tend propose <repo>` twice
- THEN #1 and #3 are each drafted once, #2 is left alone, both passes exit 0,
  and the second opens and reviews nothing again

#### Scenario: Jira issues

- GIVEN mapped project SCRUM with SCRUM-1, SCRUM-4 and SCRUM-5 labelled `tend`
  and not done, SCRUM-2 labelled and done, and SCRUM-3 unlabelled, two to a
  search page
- WHEN the owner runs `tend propose <repo>`
- THEN SCRUM-1, SCRUM-4 and SCRUM-5 are each drafted once and SCRUM-2 and
  SCRUM-3 get no pull request

### Requirement: Draft, commit and open the proposal

When drafting a request, the system shall work on its proposal branch in a
worktree started from the target as origin has it, fetched first; prompt the
drafter `prop-<n>-<cksum>` or `prop-<key in lower case>-<cksum>`, where cksum
is POSIX `cksum` of the repository's path, with `/grove-propose --unattended`
followed by the request's title, description and URL; then ask it once to
commit the draft as one commit whose body ends `Refs #<n>` or `Refs <KEY>`
and to write the open questions, consequential ones first, to `summary`; push
the branch as the account and open a pull request onto the target titled and
described by that commit; and post a comment review on its head listing the
open questions and saying that merging approves the change with them open and
how to revise it. If the drafter leaves uncommitted files, or commits nothing
under `<root>/changes/`, then the system shall stop for a person.

#### Scenario: A proposal from an issue

- GIVEN issue #1, "Request 1", "Make thing 1 better.", labelled `tend`
- WHEN the owner runs `tend propose <repo>`
- THEN `prop-1-<cksum of the repository's path>` is prompted `/grove-propose
  --unattended Request 1. Make thing 1 better. (GitHub issue <url>)`, branch
  `tend/issue-1` on origin holds the change, the account opens a pull request
  onto `main` from it titled `Propose a thing` whose description ends `Refs
  #1`, and its review lists "Which thing?", names `/grove-propose
  261115-p0000` and says the implementation may then choose those answers

#### Scenario: Origin's target moved

- GIVEN someone else pushed to origin's `main` after the local clone
- WHEN `tend propose` drafts issue #1
- THEN the proposal's branch contains their commit and its pull request opens

#### Scenario: A Jira request

- GIVEN Jira issue SCRUM-1 labelled `tend`
- WHEN `tend propose` drafts it
- THEN the drafter `prop-scrum-1-<cksum>` is asked to end the commit with
  `Refs SCRUM-1`, the account opens a pull request from `tend/jira-SCRUM-1`,
  and its review says `Proposal drafted from SCRUM-1`

#### Scenario: Two repositories with the same issue number

- GIVEN two repositories that each have an issue #1 labelled `tend`, in one
  Herdr, and the first repository's drafter still open after its pass
- WHEN the owner runs `tend propose` on the second repository
- THEN tend starts a drafter named for the second repository, prompts only
  it, and the second repository's proposal is opened as a pull request

### Requirement: A request no proposal can come from

If the drafter commits nothing and writes why to `declined` in the intake
folder, then the system shall comment that reason on the request once, as the
account on GitHub, with a marker it finds the comment by again; take the
`tend` label off; set the intake folder aside as `<folder>.declined-<id>`;
and remove the worktree with its branch; so a later pass leaves the request
until a person labels it again, when it is drafted afresh.

#### Scenario: A GitHub request declined

- GIVEN issue #1 labelled `tend`, and a drafter that writes "It names no
  outcome." and no document
- WHEN the owner runs `tend propose <repo>`, then runs it again
- THEN no pull request opens, the issue has one comment by the account with
  that reason, it no longer has the `tend` label, the worktree and branch
  `tend/issue-1` are gone, and the second pass drafts nothing; once the label
  is back, a third pass drafts it and opens a pull request

#### Scenario: The label removal fails

- GIVEN a declined request whose label removal failed, so the pass exited 3
- WHEN the owner runs `tend propose <repo>` again
- THEN it exits 0, the label is off, and the issue has one comment and was
  drafted once

#### Scenario: A Jira request declined

- GIVEN Jira issue SCRUM-1 labelled `tend`, and a drafter that declines it
- WHEN `tend propose` runs
- THEN SCRUM-1 has a comment starting "No proposal could be drafted from this
  request", with the reason and a line `[tend:declined <id>]`, its `tend`
  label is gone, and its intake folder is set aside as
  `o_r-jira-SCRUM-1.declined-<id>`

### Requirement: A cap on open proposals

While `CAP` (default 5) pull requests by the account from `tend/issue-` or
`tend/jira-` branches are open in the repository, the system shall draft no
further request in that pass, saying the cap holds and the rest wait.

#### Scenario: The cap reached

- GIVEN issues #1, #2 and #3 labelled `tend` and `CAP=2`
- WHEN `tend propose` runs
- THEN #1 and #2 are drafted, #3 is not, and tend says the rest wait; once
  #1's pull request is merged, the next pass drafts #3

### Requirement: Finish an interrupted handoff

When an intake folder of the repository records a pull request in `pr` but
not `linked`, the next `tend propose` shall finish it before taking new
requests, whatever the cap, the request's label or its state: post the
summary review if the account has none on the head, comment the pull
request's URL on a Jira issue once, found by that URL in its comments, and
record `linked`.

#### Scenario: The Jira comment failed

- GIVEN SCRUM-1's proposal pull request opened but the comment on SCRUM-1
  failed, so the pass exited 3 saying to run `tend propose` again to finish,
  and SCRUM-1's `tend` label has since been removed
- WHEN the owner runs `tend propose <repo>` again
- THEN SCRUM-1 has one comment `tend opened <pr url> …`, no second pull
  request is opened, and the intake folder records `linked`

### Requirement: Drafting prompts are sent once

The system shall write `asked` and `followed` in the intake folder just
before sending the drafting and the follow-up prompts, never sending either
again while its mark exists, but for a follow-up whose drafter is gone; a
drafter still working shall be waited for; if the follow-up was asked and the
drafter neither committed a draft nor wrote `declined` and is there and not
working, the system shall stop for a person on every pass until a person
commits the draft, writes `declined`, or deletes `followed`; and if the
follow-up was asked, left neither, and the drafter is gone, the system shall
ask it once more of a new session, as the flows spec's "Any step recovers
without being prompted twice" says.

#### Scenario: The follow-up stalled

- GIVEN the follow-up prompt stalled and the pass exited 3
- WHEN the owner runs `tend propose` again
- THEN it exits 3 without asking again, saying to delete `followed` to have
  it asked again; once `followed` is deleted, the next pass asks a second
  time, opens the pull request, and the draft was asked for once in all

#### Scenario: The drafter outlived a timeout

- GIVEN the drafting prompt timed out and the drafter is still working
- WHEN the owner runs `tend propose` again
- THEN it waits for the drafter, has it commit, opens the pull request, and
  asked for the draft once

#### Scenario: The drafter is gone

- GIVEN a pass was killed after asking the follow-up, the drafter committed
  nothing and wrote no `declined`, and its pane is closed
- WHEN the owner runs `tend propose` again
- THEN the follow-up is asked once more of a new drafter session, and the
  pull request opens or the request is told why there is none

### Requirement: A proposal head of Grove documents only passes the rule

When `tend propose` opens a proposal pull request, or finds an open pull
request that the account opened from a `tend/issue-` or `tend/jira-` branch
whose head has no `tend/review` status set by the account, the system shall
set `tend/review` on that head, as the account, linking the pull request:
`success`, described as Grove documents only and that merging approves the
change, when every path the head changes against the target as fetched lies
in one change's folder under `<root>/changes/`; and no `success` otherwise.

#### Scenario: A proposal just opened

- GIVEN issue #1 labelled `tend`, and a drafter that commits a change only
  under `grove/changes/261115-p0000-thing/`
- WHEN the owner runs `tend propose <repo>`
- THEN the pull request's head has `tend/review` `success` set by the
  account, described "Grove documents only: merging approves this change"

#### Scenario: A proposal opened before this change

- GIVEN an open proposal pull request from `tend/issue-1` by the account,
  whose head changes only one change's folder and has no status
- WHEN the owner runs `tend propose <repo>`, with no labelled request left
  and the cap reached
- THEN that head has `tend/review` `success` set by the account

#### Scenario: A person revises the proposal

- GIVEN a proposal whose head has `tend/review` `success`, and a person
  pushes a commit editing only that change's `proposal.md` to its branch
- WHEN the owner runs `tend propose <repo>`
- THEN the new head has its own `tend/review` `success`, and the earlier
  head's status is not set again

#### Scenario: A head that touches code

- GIVEN an open proposal pull request whose head also changes `intake.go`
- WHEN the owner runs `tend propose <repo>`
- THEN its head has no `tend/review` `success`

#### Scenario: Two changes in one head

- GIVEN an open proposal pull request whose head adds both
  `grove/changes/261115-p0000-thing/` and `grove/changes/261115-p0001-other/`
- WHEN the owner runs `tend propose <repo>`
- THEN its head has no `tend/review` `success`

#### Scenario: Pull requests that are not proposals

- GIVEN a merged proposal pull request, and an open pull request from
  `tend/issue-2` opened by a person, not the account
- WHEN the owner runs `tend propose <repo>`
- THEN neither head gets a `tend/review` status

## Known discrepancies

None recorded.

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
- [261122-hjhmn](../../changes/archive/261122-hjhmn-recover-with-no-one-there/) Recover with no one there
- [261124-60tny](../../changes/archive/261124-60tny-name-the-drafter-after-its-repository/) An intake drafter is named after its request and its repository
- [261126-c16zs](../../changes/archive/261126-c16zs-give-proposals-a-review-status/) Give proposal pull requests a tend/review status
