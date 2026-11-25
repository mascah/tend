# Landing

## Purpose

At a `land` step of a project's flow, tend lands one reviewed candidate at a
time: it merges the target into the candidate when the target has moved,
waits for every check on that head and has GitHub squash-merge exactly that
head, a head made by a clean merge of the target keeping the verdict of the
head it came from. Projects that opt into landing rely on this, and people
approving candidates in a supervised flow.

Landing is merging an attempt's pull request into the target. A head passed
review when its verdict's `conclusion` is `success`. A person with write
access is who the publication spec's "Requested changes are acted on once"
says. `land` is the step of a flow where tend lands, as the flows spec's
"Landing is tend's own and a flow places it" says.

The protected paths in effect are what the flows spec's "The flow file names
the protected paths" says. A head's changed paths are those that differ
between the head and the commit where the head and the target's tip, as
origin holds it, fetched, meet. A renamed file counts under both its old
and its new name.

## Boundaries

This covers what tend does at `land` and the `approved` outcome of
`publish` that a flow may send there. Where `land` sits in a flow and what
makes such a flow valid is [flows](../flows/spec.md); the push, pull request,
reviews and statuses are [publication](../publication/spec.md); reviewing a
head is [review](../review/spec.md); `tend watch` running an approved attempt
again is [watch-and-status](../watch-and-status/spec.md). Several candidates
tested at once, GitHub's merge queue, setting or reading a repository's
branch rule, reading an approval again before the merge (except for a head
that touches a protected path), an order of landing
across machines, running a failed check again and deleting a merged branch
are outside it: each runner lands one candidate at a time, and between
runners GitHub's branch rule decides.

## Requirements

### Requirement: Land only a head that passed review

When a flow reaches `land`, the system shall land the attempt's head only if
it passed review or is carried from a head that did, its pull request is open
and the head is the pull request's; if the head did not pass, the
repository's origin is not listed for publishing, or the pull request is
closed, it shall stop for a person saying which and merge nothing; and a pull
request already merged shall give `ok`.

#### Scenario: A failed head reaches land

- GIVEN a flow that sends every outcome of `review` to `publish` and
  `publish`'s `ok` to `land`, and a reviewer that fails the head with
  `actionable` false
- WHEN tend runs the attempt
- THEN the failed head is published, tend exits 3 saying the head did not
  pass review, and nothing is merged

#### Scenario: An origin not listed

- GIVEN a flow with `land`, and a repository whose origin is not listed in
  `TEND_STATE/publish`
- WHEN tend takes a cycle to a passing verdict
- THEN tend exits 3 saying there is no pull request to land, and runs no `gh`
  command

#### Scenario: A person merged it first

- GIVEN a candidate that passed review, whose pull request a person merges
  before tend reaches `land`
- WHEN tend runs the attempt again
- THEN `land` gives `ok`, nothing more is merged, and tend exits 0

### Requirement: One landing at a time

While one tend lands a candidate of a repository, the system shall have any
other tend using that TEND_STATE that reaches `land` for the same repository
wait its turn, its `status` saying it waits to land, and then land against
the target as the earlier landing left it.

#### Scenario: Two candidates ready

- GIVEN two attempts whose heads passed review, each reaching `land`
- WHEN both tends run
- THEN the first pull request is merged; the second attempt's branch then
  gains a merge of the target that holds the first's commit, its checks are
  waited for, and it is merged; and the target holds both

### Requirement: The target is merged in

When the candidate's head does not hold the target's tip as origin has it,
fetched, the system shall merge the target into the attempt's branch in its
worktree as a merge commit and push it, never forced; and if the two do not
merge cleanly, it shall leave the branch and its worktree as they were and
give `conflict`.

#### Scenario: The target moved

- GIVEN a candidate that passed review, after which a commit that touches
  another file reached the target
- WHEN tend reaches `land`
- THEN the branch's head is a merge commit of the reviewed head and the
  target's tip, origin's branch is at it, and nothing was pushed by force

#### Scenario: The target did not move

- GIVEN a candidate whose head holds the target's tip
- WHEN tend reaches `land`
- THEN no merge commit is made

#### Scenario: A conflict

- GIVEN a candidate that passed review, after which a commit that changes
  the same line reached the target, and a flow that sends `conflict` to a
  step continuing the implementer
- WHEN tend reaches `land`
- THEN the branch is at the reviewed head with no merge under way, that step
  is asked of the implementer, and the head the implementer makes is reviewed
  before anything lands

### Requirement: The verdict carries across a clean merge

The system shall hold a head carried when it is a merge commit whose first
parent passed review or is itself carried, whose second parent is a commit of
the target, and whose tree is exactly Git's merge of the two; for a carried
head it shall set `tend/review` to `success` on it, described as carried from
the reviewed head, post no review and prompt no reviewer; and it shall hold
any other head without a verdict as not reviewed.

#### Scenario: A carried head

- GIVEN a candidate whose head passed review, into which `land` merged the
  target
- WHEN tend goes on to its checks
- THEN the merge commit has the status `tend/review` `success`, whose
  description names the reviewed head, the pull request has no review of the
  merge commit, and one reviewer was started in the attempt

#### Scenario: A rerun after a kill

- GIVEN tend was killed after pushing the merge commit and before the pull
  request merged
- WHEN the owner runs tend again
- THEN no reviewer is prompted, and the pull request is landed

#### Scenario: A merge with an edit of its own

- GIVEN an implementer that merges the target into the branch and changes a
  file in the same commit
- WHEN tend runs the attempt
- THEN that head is not carried, and a reviewer is prompted for it

### Requirement: Checks decide

Before merging, the system shall wait, for at most `TIMEOUT_MS`, until every
check GitHub reports on the head, check runs and commit statuses alike, has
finished; a check that finished as success, neutral or skipped passed, and
one that finished any other way, as cancelled or timed out, failed; if one
failed, it shall give `failed`, giving the step that follows the failed
checks' names and links for its prompt; and if GitHub reports no check on the
head but `tend/review`, or one is still unfinished when the wait ends, it
shall stop for a person saying so, and merge nothing.

#### Scenario: A check fails on the merged head

- GIVEN a candidate into which `land` merged the target, a check that then
  fails on the merge commit, and a flow that sends `failed` to a step
  continuing the implementer
- WHEN tend waits for the checks
- THEN nothing is merged, that step's prompt names the failed check, and the
  head the implementer makes is reviewed before anything lands

#### Scenario: A cancelled check

- GIVEN a candidate on whose head one check passed and another was cancelled,
  and a flow that sends `failed` to a step continuing the implementer
- WHEN tend waits for the checks
- THEN nothing is merged, and that step's prompt names the cancelled check

#### Scenario: A skipped check

- GIVEN a candidate on whose head one check passed and another was skipped
- WHEN tend reaches `land`
- THEN the pull request is merged

#### Scenario: No CI

- GIVEN a candidate that passed review, on whose head GitHub reports no check
  but `tend/review`
- WHEN tend reaches `land`
- THEN tend exits 3 saying the head has no check but `tend/review`, and
  nothing is merged

#### Scenario: The checks pass

- GIVEN a candidate whose head holds the target's tip and whose checks all
  pass
- WHEN tend reaches `land`
- THEN the pull request is merged

### Requirement: Squash-merge exactly that head

When the head holds the target's tip and its checks passed, the system shall
have GitHub squash-merge the pull request at exactly that head, as the
publishing account, with the pull request's title followed by its number as
the commit's subject and its description as the body; if GitHub refuses
because the target or the head moved, it shall begin again from merging the
target in; if GitHub refuses only because a check it requires is pending or
not yet reported, it shall go back to waiting for the checks, within the same
wait; if GitHub refuses for any other reason, it shall stop for a person with
GitHub's reason; and once merged it shall give `ok`, `status` saying the pull
request landed.

#### Scenario: Landed

- GIVEN `TEND_STATE/publish` lists the origin with account `the-bot`, and a
  candidate titled `Add the file` whose head passed review and whose checks
  pass
- WHEN tend reaches `land`
- THEN the target gains one commit, `Add the file (#1)`, whose body ends
  `Grove: 261015-c0000`, the pull request is merged by `the-bot`, `status`
  says `done:` and that the pull request landed, and tend exits 0

#### Scenario: The target moved while the checks ran

- GIVEN a candidate whose checks pass, and a commit that reaches the target
  before the merge
- WHEN GitHub refuses the merge
- THEN tend merges the target in again, waits for the new head's checks, and
  lands it

#### Scenario: A required check reports late

- GIVEN a repository whose rule requires the checks `title` and `check`, and
  a candidate on whose head only `title` is reported, and passed, when tend
  first looks
- WHEN GitHub refuses the merge because `check` is expected
- THEN tend stops for no one, waits, and lands the pull request once `check`
  passes

#### Scenario: A description a person edited

- GIVEN a candidate whose pull request description a person replaced with
  their own words
- WHEN tend lands it
- THEN the commit's body is their words

#### Scenario: GitHub refuses for a rule

- GIVEN a repository whose rule requires an approving review, and a candidate
  with none
- WHEN tend reaches `land` and GitHub refuses the merge
- THEN tend exits 3 quoting GitHub's reason, and the pull request stays open

### Requirement: A person's approval is an outcome of publishing

Where a flow's `publish` gives a next step for `approved`, the system shall
give that outcome when the newest review on the pull request's head by a
person with write access approves it, an approval of a head holding for a
head carried from it; and in a flow whose `publish` names no `approved`, an
approval shall change nothing.

#### Scenario: A teammate approves

- GIVEN a flow whose `publish` sends `approved` to `land` and `ok` to the
  end, and a published candidate that passed review
- WHEN the owner runs tend, then a teammate who is a collaborator approves
  the head and the owner runs tend again
- THEN the first run merges nothing and exits 0, and the second lands the
  pull request

#### Scenario: A stranger approves

- GIVEN that flow, and an account with no access to the repository approves
  the head
- WHEN the owner runs tend again
- THEN nothing is merged, and tend exits 0

#### Scenario: A new head after the approval

- GIVEN that flow, a teammate's approval of the head, and a commit the
  implementer then makes for a requested change
- WHEN tend publishes the new head
- THEN nothing is merged until a person with write access approves the new
  head

#### Scenario: The built-in flow

- GIVEN a repository with no flow file, and a teammate approves a published
  head
- WHEN the owner runs tend again
- THEN nothing is merged, and tend exits 0 as before

### Requirement: A head that touches a protected path waits for a person

When a flow reaches `land` with a head whose changed paths include a
protected path, the system shall land it only if the newest review that
approves or requests changes, by a person with write access, on that head or
on a head it is carried from, approves it; and otherwise it shall merge
nothing, push nothing, prompt no agent, and stop for a person, naming each
protected path the head changes and saying that a person with write access
must approve the head. The system shall decide this from Git and GitHub's
reviews alone, never from a verdict or any file an agent writes, and shall
decide it again, with the protected paths read afresh from the target,
before it has GitHub merge the pull request.

#### Scenario: An autonomous flow and an edit to the flow file

- GIVEN a flow whose `publish` sends `ok` to `land`, and a candidate that
  passed review, whose change edits `tend/flow.yaml` and `main.go`, with no
  approving review
- WHEN tend reaches `land`
- THEN tend exits 3 naming `tend/flow.yaml` and not `main.go`, the pull
  request stays open, and no merge commit is pushed to the branch

#### Scenario: A person approves the held head

- GIVEN that held candidate, whose head a teammate who is a collaborator
  then approves
- WHEN the owner runs tend again
- THEN the pull request is landed

#### Scenario: A stranger approves the held head

- GIVEN that held candidate, whose head an account with no access to the
  repository approves
- WHEN the owner runs tend again
- THEN tend exits 3 naming `tend/flow.yaml`, and nothing is merged

#### Scenario: No protected path

- GIVEN a flow whose `publish` sends `ok` to `land`, and a candidate that
  passed review, whose change touches only `main.go`, with no approving
  review
- WHEN tend reaches `land` and its checks pass
- THEN the pull request is landed

#### Scenario: A nested agent instruction file

- GIVEN a flow whose `publish` sends `ok` to `land`, no `protect` in the
  target's flow file, and a candidate that passed review, whose change adds
  `web/CLAUDE.md`
- WHEN tend reaches `land`
- THEN tend exits 3 naming `web/CLAUDE.md`, and nothing is merged

#### Scenario: A protected file moved away

- GIVEN a flow whose `publish` sends `ok` to `land`, and a candidate that
  passed review, whose change renames `.github/workflows/ci.yml` to
  `ci.yml.bak`
- WHEN tend reaches `land`
- THEN tend exits 3 naming `.github/workflows/ci.yml`, and nothing is merged

#### Scenario: The candidate unprotects a path

- GIVEN a target whose flow file sends `publish`'s `ok` to `land` and
  protects `schema.sql`, and a candidate that passed review, whose change
  edits `schema.sql` and drops it from `protect` in its own flow file
- WHEN tend reaches `land`
- THEN tend exits 3 naming `schema.sql` and `tend/flow.yaml`, and nothing is
  merged

#### Scenario: The target moves in a carried head

- GIVEN a flow whose `publish` sends `ok` to `land`, a candidate that passed
  review and edits `CLAUDE.md`, a teammate's approval of that head, and a
  commit that touches another file reaching the target
- WHEN tend merges the target in and the checks on the merge commit pass
- THEN the carried head is landed without a second approval, and the
  target's own commit is not counted as the candidate's change

#### Scenario: A head made at fix

- GIVEN a flow whose `publish` sends `ok` to `land` and whose `land` sends
  `failed` to `fix`, a candidate touching only `main.go` whose checks fail,
  and an implementer whose commit at `fix` edits `.github/workflows/ci.yml`
  and passes review
- WHEN tend reaches `land` with that head
- THEN tend exits 3 naming `.github/workflows/ci.yml`, and nothing is merged

#### Scenario: An approval before the implementer's new head

- GIVEN a flow whose `publish` sends `ok` to `land` and whose `land` sends
  `conflict` to `resolve`, a candidate that edits `grove.yaml` with a
  teammate's approval, and a conflict the implementer resolves at `resolve`
  in a new head that passes review
- WHEN tend reaches `land` with that head
- THEN tend exits 3 naming `grove.yaml`, and nothing is merged until a
  person with write access approves the new head

#### Scenario: A supervised flow

- GIVEN a flow whose `publish` sends `approved` to `land` and `ok` to the
  end, and a candidate that edits `tend/flow.yaml`, which a teammate
  approves
- WHEN the owner runs tend again
- THEN the pull request is landed as before

## Known discrepancies

None recorded.

## Open questions

- How long should `land` wait for a first check before it says the project
  has no CI on pull requests? Today the wait is the `TIMEOUT_MS` bound, four
  hours by default, during which the repository's landing lock is held.

## History

- [261122-rpkqa](../../changes/archive/261122-rpkqa-land-reviewed-candidates-one-at-a-time/) Land reviewed candidates one at a time
- [261125-6dvbs](../../changes/archive/261125-6dvbs-hold-candidates-that-touch-protected-paths/) Hold a candidate that touches a protected path for the owner
