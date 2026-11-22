# Changes to watch-and-status

The changes that change 261122-9vbmc proposes to the requirements of
`watch-and-status`, the maintained spec at
[grove/specs/watch-and-status/spec.md](../../../../specs/watch-and-status/spec.md).
This is future work: nothing here describes the target branch until the change
is implemented and the maintained spec is reconciled.

An attempt is being worked on while a tend holds its work item's lock,
whoever started that tend.

## ADDED Requirements

### Requirement: Watch keeps a count of attempts going

Where a repository's flow file sets `watch.at_once`, `tend watch` shall, in
place of what "Watch acts on a person's behalf" says of starting a cycle and
of running an attempt again, keep at most that many of the repository's
attempts on this runner being worked on: on each pass, while fewer are, it
shall first run tend again on each attempt it would run again, oldest first,
and then start each cycle not done and each change in no cycle that Grove
calls claimable (`claimable` in `grove changes --json`), marked `unattended`
or not, oldest `created` first, saying which it starts; it shall wait for
none of them before its next pass, run intake beside them and outside the
count, and leave alone an attempt that needs a person; and with `WATCH_ONCE`
set it shall wait for what it started and then exit.

#### Scenario: Two changes that share nothing

- GIVEN a flow file with `at_once: 2`, and approved changes A and B in no
  cycle that share no planned file, neither marked
- WHEN one pass of `tend watch <repo>` runs
- THEN it says it is starting each, and both are claimed, built and reviewed,
  each in its own worktree

#### Scenario: A count of 1

- GIVEN a flow file with `at_once: 1`, and approved changes A and B in no
  cycle that share no planned file, A's `created` the earlier
- WHEN one pass of `tend watch <repo>` runs, then another
- THEN the first starts A only, and the second, A's head reviewed, starts B

#### Scenario: Two that overlap

- GIVEN a flow file with `at_once: 2`, and approved changes A and B that
  both name `a.txt`
- WHEN one pass of `tend watch <repo>` runs, then another once the target
  holds A archived
- THEN the first builds A and makes no worktree for B, and the second starts
  B

#### Scenario: An attempt that needs a person

- GIVEN a flow file with `at_once: 1`, A's attempt stopped with `status`
  saying `needs-you`, and approved change B, which shares nothing with A
- WHEN a pass of `tend watch <repo>` runs
- THEN it starts B and does not run tend on A

#### Scenario: A run by hand counts

- GIVEN a flow file with `at_once: 1`, and a tend the owner started on A
  whose implementer is on its turn
- WHEN a pass of `tend watch <repo>` runs
- THEN it starts nothing

#### Scenario: A cycle not marked

- GIVEN a flow file with `at_once: 1`, and a claimable cycle whose plan is
  not marked `unattended`
- WHEN a pass of `tend watch <repo>` runs
- THEN the cycle is started

#### Scenario: Changes requested go first

- GIVEN a flow file with `at_once: 1`, a published attempt with a change
  request on its head from a collaborator, and claimable change B
- WHEN a pass of `tend watch <repo>` runs
- THEN it runs tend on the attempt, whose request is asked of the
  implementer, and does not start B in that pass

#### Scenario: No setting

- GIVEN a repository with no flow file, an approved change in no cycle and a
  cycle not marked `unattended`
- WHEN a pass of `tend watch <repo>` runs
- THEN it starts nothing

### Requirement: Tidying leaves what a tend is on

While another tend holds a work item's lock, or another intake pass a
repository's, the system shall remove no worktree of that attempt, or of that
repository's requests, as "Finished worktrees are tidied" would have it do,
and shall say it left it because a tend is on it; a tend's own lock shall not
keep it from tidying what it would otherwise remove.

#### Scenario: A pull request closed under a tend

- GIVEN a tend on cycle K that has published its head and is still running,
  whose pull request a person closes
- WHEN the owner runs tend on change C
- THEN K's worktree and branch stay, and tend says it left the worktree as a
  tend is on it; once K's tend has exited, the next run removes them

#### Scenario: A tend tidies its own finished work

- GIVEN issue #7's pull request is merged
- WHEN the owner runs `tend <repo> issue-7`
- THEN the worktree of `tend/build-7` is removed, though that tend holds the
  lock of `issue-7`

### Requirement: The runner may lower the count

Where `AT_ONCE` is set to a whole number in `tend watch`'s environment, the
system shall, in each repository whose flow file sets the count, keep at most
the smaller of the two, starting and running again nothing at 0 while
attempts already being worked on finish, and never more than the flow file's
count.

#### Scenario: A lower number on this machine

- GIVEN a flow file with `at_once: 2` and two claimable changes that share
  nothing
- WHEN a pass of `tend watch <repo>` runs with `AT_ONCE=1`
- THEN one is started

#### Scenario: A higher number on this machine

- GIVEN a flow file with `at_once: 1` and two claimable changes that share
  nothing
- WHEN a pass of `tend watch <repo>` runs with `AT_ONCE=5`
- THEN one is started

#### Scenario: Nothing new

- GIVEN a flow file with `at_once: 2` and a claimable change
- WHEN a pass of `tend watch <repo>` runs with `AT_ONCE=0`
- THEN nothing is started
