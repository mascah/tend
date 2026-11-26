# Changes to watch-and-status

<!-- The changes that change 261125-rf036 proposes to the requirements of
`watch-and-status`. This is future work: nothing here describes the target
branch until the change is implemented and the maintained spec is
reconciled. -->

Starts from the maintained [watch-and-status spec](../../../../specs/watch-and-status/spec.md)
at 2a8ae80. The requirement's sentence is unchanged; one scenario is added,
"Waiting for checks counts", saying that an attempt whose `publish` waits for
checks is being worked on.

## MODIFIED Requirements

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

#### Scenario: Waiting for checks counts

- GIVEN a flow file with `at_once: 1` whose `publish` gives a step for
  `failed`, A's attempt published and its tend waiting for the checks on
  A's head, and claimable change B, which shares nothing with A
- WHEN a pass of `tend watch <repo>` runs, then another once A's checks
  passed and its tend exited waiting for a person's approval
- THEN the first starts nothing, and the second starts B

#### Scenario: No setting

- GIVEN a repository with no flow file, an approved change in no cycle and a
  cycle not marked `unattended`
- WHEN a pass of `tend watch <repo>` runs
- THEN it starts nothing
