# Changes to attempts

The changes that change 261122-9vbmc proposes to the requirements of
`attempts`, the maintained spec at
[grove/specs/attempts/spec.md](../../../../specs/attempts/spec.md). This is
future work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## ADDED Requirements

### Requirement: One tend for each work item

While a tend holds the lock of a work item of a repository in TEND_STATE, the
system shall refuse another `tend <repo> <item>` on that work item in that
repository using that TEND_STATE, after waiting about two seconds for the
lock to come free, naming the holder's pid, work item and repository,
prompting no agent and exiting 4; it shall likewise allow one `tend propose`
pass for a repository at a time; a tend on another work item or another
repository, and an intake pass beside attempts, shall run at the same time;
and a lock shall be released when its holder exits or is killed, whatever
agents outlive it.

#### Scenario: A second tend on the same work item

- GIVEN a tend whose implementer is still on its turn
- WHEN the owner runs tend on the same cycle again
- THEN it exits 4, standard error names `pid <first tend's pid>`, and the
  implementer has been prompted once

#### Scenario: The holder is killed

- GIVEN that first tend is killed with `kill -9`
- WHEN the owner runs tend on the cycle again
- THEN it takes the lock, says it is carrying on the attempt, and exits 0

#### Scenario: Two work items at once

- GIVEN a tend on cycle K whose implementer is still on its turn, and
  approved change C in no cycle, which shares no planned file with K
- WHEN the owner runs `tend <repo> <C's ID>`
- THEN it claims C in a worktree of its own, both attempts reach a verdict,
  and each implementer was prompted once

#### Scenario: Intake beside a build

- GIVEN a tend on cycle K whose implementer is still on its turn, and issue
  #1 labelled `tend`
- WHEN the owner runs `tend propose <repo>`
- THEN the proposal is drafted and its pull request opens, and K's attempt
  reaches its verdict

#### Scenario: A second intake pass

- GIVEN a `tend propose <repo>` whose drafter is still on its turn
- WHEN the owner runs `tend propose <repo>` again
- THEN it exits 4 naming the first pass's pid

### Requirement: Attempts set up one at a time

While one tend fetches the target, tidies finished worktrees, makes its
worktree and claims, the system shall have any other tend using that
TEND_STATE wait before doing the same, so that tends started together each
end with a worktree, a branch and a record of their own.

#### Scenario: Two started together

- GIVEN approved changes A and B in no cycle that share no planned file
- WHEN the owner starts `tend <repo> <A's ID>` and `tend <repo> <B's ID>` in
  the same second
- THEN Grove holds one record for each, each attempt has its own worktree,
  both reach a verdict, and the repository has no `tend-*` branch afterwards

### Requirement: A refused claim leaves nothing behind

If a piece has no attempt to carry on and `grove changes --json` says Grove
would not claim it, or Grove refuses its claim, then the system shall stop
for a person with Grove's sentence and leave no worktree or branch it made
for the claim.

#### Scenario: A piece that overlaps one in flight

- GIVEN change A is in flight and names `a.txt`, and approved change B, in no
  cycle, names `a.txt`
- WHEN the owner runs `tend <repo> <B's ID>`
- THEN tend exits 3, standard error holds Grove's sentence that B shares
  `a.txt` with A, and the repository has the worktrees and branches it had
  before

## MODIFIED Requirements

### Requirement: Exit statuses

When an attempt ends, the system shall exit 0 once the head is reviewed,
whatever the verdict's conclusion, printing the verdict JSON on standard
output and notifying that the branch or pull request is ready; exit 3 when a
person is needed, saying why on standard error, recording it in `status` as
`needs-you` and notifying through the terminal; exit 4 when another tend holds
the work item's lock; and exit 2 with `usage: tend <repo> <id>` when no
repository or work item is given.

#### Scenario: A cycle reaches a verdict

- GIVEN a repository with origin and cycle `261114-k0000` approved and planned
- WHEN the owner runs `tend <repo> 261114-k0000` and the reviewer passes the
  head
- THEN tend exits 0 and standard output is the verdict, whose `conclusion` is
  `success`

#### Scenario: The reviewer leaves no valid verdict

- GIVEN a reviewer that writes a verdict with no `conclusion`
- WHEN the owner runs tend on the cycle
- THEN tend exits 3, saying the reviewer settled without a valid verdict

#### Scenario: No work item named

- WHEN the owner runs `tend <repo>`
- THEN tend prints `usage: tend <repo> <id>` and exits 2

### Requirement: Work stays on the attempt's branch

When an implementer turn that builds, repairs or makes requested changes
ends, and on every run before the head is reviewed, if the attempt's
worktree is on a branch other than the attempt's, or the repository has a
worktree, or a local branch holding commits the target lacks, that it did
not have just before the implementer was last prompted and that no other
attempt or intake pass of this runner made, the system shall stop for a
person naming what it found, and review and publish nothing. What another
attempt or intake pass made is the worktree its run or intake folder records
and the branch that worktree is on.

#### Scenario: The implementer adds a worktree

- GIVEN an implementer that builds the cycle and also adds a worktree
- WHEN the owner runs tend on the cycle
- THEN tend exits 3 naming the worktree, and no reviewer is prompted

#### Scenario: A rerun finds it still there

- GIVEN tend stopped because the implementer added a worktree, which is
  still there
- WHEN the owner runs tend again
- THEN tend exits 3 again, and no reviewer is prompted

#### Scenario: The implementer makes a branch with commits

- GIVEN an implementer that builds the cycle and also makes a branch at its
  head
- WHEN the owner runs tend on the cycle
- THEN tend exits 3 naming the branch, and no reviewer is prompted

#### Scenario: Another attempt starts during the turn

- GIVEN a tend on cycle K whose implementer is on its turn, during which the
  owner runs tend on change C, which makes its worktree and branch and
  commits on it
- WHEN K's implementer's turn ends
- THEN K's head is reviewed, and tend does not stop for C's worktree or
  branch

#### Scenario: A stray worktree beside another attempt's

- GIVEN K's implementer adds a worktree of its own during a turn in which
  an attempt at change C also starts
- WHEN K's implementer's turn ends
- THEN tend exits 3 naming the implementer's worktree and not C's

## REMOVED Requirements

### Requirement: One tend per runner

Replaced by "One tend for each work item": the lock is now a work item's, so
that tends on different work items run at the same time.
