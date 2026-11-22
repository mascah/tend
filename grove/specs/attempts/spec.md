# Attempts

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

`tend <repo> <cycle-id>` takes one approved Grove cycle, or an approved change no cycle lists, through one attempt, claiming it or carrying on the attempt it has, implementing it until Git shows a finished candidate, and recovering after any stop or kill without prompting an agent twice. An attempt is one execution of a work item; its run folder is `TEND_STATE/runs/<attempt>/`, and TEND_STATE is `$TEND_STATE`, else `$XDG_STATE_HOME/tend`, else `~/.local/state/tend`. The owner running tend at a terminal, and `tend watch` running it for them, rely on this.

## Boundaries

This covers the command line and exit statuses of an attempt, the lock of
each work item and the setting up tends take in turn, finding or claiming the attempt of a Grove cycle or of a change in no
cycle, where its worktree starts, implementing until the branch holds a
finished candidate, recovery after a stop or kill, and the run folder and its `status` record. These hold for
every attempt, whatever its work item. Reviewing the head and the one repair
are [review](../review/spec.md); pushing and the pull request are
[publication](../publication/spec.md); a GitHub or Jira issue as the work item
is [work-items](../work-items/spec.md); how agents are started, prompted and
observed in Herdr or rex is [terminals](../terminals/spec.md); `tend status`,
`tend watch` and tidying finished worktrees are
[watch-and-status](../watch-and-status/spec.md); `tend propose` is
[intake](../intake/spec.md). Grove, not tend, owns approval, eligibility and
the claim.

## Requirements

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
end with a worktree, a branch and a record of their own; a tend shall record
its worktree in its run folder, and an intake pass a request's worktree in
its intake folder, before it lets the next go; and a tend shall wait for one
that is setting up before it checks what "Work stays on the attempt's branch"
says.

#### Scenario: Two started together

- GIVEN approved changes A and B in no cycle that share no planned file
- WHEN the owner starts `tend <repo> <A's ID>` and `tend <repo> <B's ID>` in
  the same second
- THEN Grove holds one record for each, each attempt has its own worktree,
  both reach a verdict, and the repository has no `tend-*` branch afterwards

#### Scenario: A turn ends while another attempt sets up

- GIVEN a tend on cycle K whose implementer's turn ends while a tend on
  change C has made its worktree and is still claiming
- WHEN K's tend checks the repository
- THEN it waits for C's tend to finish setting up, K's head is reviewed, and
  tend does not stop for C's worktree or branch

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

### Requirement: Carry on the attempt or claim the cycle

When tend runs on a Grove cycle, the system shall carry on the newest attempt
record for the cycle in `grove changes --json` whose checkout still holds its
branch, and otherwise make a new worktree, run `grove claim <cycle-id>` inside
it and delete the temporary `tend-HHMMSS-<pid>` branch the worktree was made
with, named for the time and tend's process so that tends setting up in the
same second each make their own, stopping for a person if the claim is
refused; with or without origin.

#### Scenario: A rerun after a stop

- GIVEN a repository with origin where tend claimed the cycle and then stopped
  because the implementer's prompt failed
- WHEN the owner runs tend on the cycle again
- THEN standard error says `carrying on attempt <id>`, Grove holds one record
  for the cycle, and tend exits 0

#### Scenario: Without origin

- GIVEN a repository with no remote, where tend claimed the cycle and stopped
- WHEN the owner runs tend on the cycle again
- THEN it carries on the same attempt and exits 0

#### Scenario: The temporary branch goes

- WHEN tend claims a cycle
- THEN the repository has no `tend-*` branch afterwards

### Requirement: A change in no cycle is a work item

When tend runs on the ID of an approved change that no cycle not done lists,
the system shall take it through one attempt as it takes a cycle of that one
change: carrying on the newest attempt record in `grove changes --json` that
names the change and whose checkout still holds its branch, and otherwise
making a new worktree and running `grove claim <change-id>` inside it,
stopping for a person with Grove's own sentence if the claim is refused;
prompting the implementer with `/grove-implement --runner <change-id>`; and
holding the candidate finished once the change is under
`<root>/changes/archive/` on the branch. What the specs say of a Grove cycle
and its attempt shall hold for such a change and its attempt, a change
archived on the target counting as a cycle done.

#### Scenario: A change claimed by name

- GIVEN a repository with origin and approved change `261016-c0001`, which no
  cycle lists and which waits on nothing
- WHEN the owner runs `tend <repo> 261016-c0001` and the reviewer passes the
  head
- THEN Grove holds one record, naming `261016-c0001` alone and no cycle, the
  implementer was prompted `/grove-implement --runner 261016-c0001`, and tend
  exits 0 printing a verdict whose `conclusion` is `success`

#### Scenario: A rerun after a stop

- GIVEN tend claimed `261016-c0001` and then stopped because the
  implementer's prompt failed
- WHEN the owner runs `tend <repo> 261016-c0001` again
- THEN standard error says `carrying on attempt <id>`, Grove holds one record
  for the change, and tend exits 0

#### Scenario: A change that waits

- GIVEN approved changes A and B that no cycle lists, B with `depends_on: [A]`,
  and A not archived
- WHEN the owner runs `tend <repo> <B's ID>`
- THEN tend exits 3, standard error holds Grove's sentence that B waits on A,
  and no implementer is prompted

#### Scenario: A member of a cycle

- GIVEN cycle K, not done, lists approved change A
- WHEN the owner runs `tend <repo> <A's ID>`
- THEN tend exits 3, standard error holds Grove's sentence that A is in cycle
  K, and no implementer is prompted

#### Scenario: Its status

- GIVEN tend took `261016-c0001` to a verdict
- WHEN a person runs `tend status <repo> 261016-c0001`
- THEN it prints `attempt <id> of 261016-c0001: <branch> in <checkout>` and
  the attempt's recorded status

#### Scenario: Its pull request

- GIVEN `TEND_STATE/publish` lists the origin
- WHEN tend publishes the reviewed head of `261016-c0001`
- THEN one pull request is opened onto the target, and its description ends
  `Grove: 261016-c0001`

#### Scenario: The change landed some other way

- GIVEN an attempt at `261016-c0001` with no pull request and no commits,
  after which the target holds the change archived
- WHEN `tend propose` runs
- THEN the attempt's worktree and local branch are removed

### Requirement: New work starts from origin's target

When an attempt needs a new worktree, the system shall start it from
`origin/<target>` after fetching the target from origin, or from the local
target when the repository has no origin, where the target is the branch
Grove names (an issue build's is the repository's default branch), and shall
show the reviewer the diff from that same base; and if the fetch fails it
shall stop for a person.

#### Scenario: The local target is behind

- GIVEN someone else pushed a commit to origin's `main` that the local `main`
  lacks
- WHEN the owner runs tend on the cycle
- THEN the attempt's branch contains that commit, and tend exits 0

#### Scenario: A project targeting trunk

- GIVEN a repository without origin whose `grove.yaml` names target `trunk`
- WHEN the owner runs tend on the cycle
- THEN the worktree is created with base `trunk`, and the reviewer is told
  the change is `git diff trunk...HEAD`

### Requirement: Implement until Git shows a finished candidate

When the attempt's branch has uncommitted files or a change of the cycle that
is not under `<root>/changes/archive/` on it, the system shall prompt the
implementer `impl-<attempt>` with `/grove-implement --runner <cycle-id>`
and, once its turn ends, stop for a person naming what is still missing; and
it shall never prompt the implementation of a branch that already holds a
finished candidate, judging that from Git, never from the agent's state.

#### Scenario: The documents moved on the target

- GIVEN an attempt already built and reviewed, after which the target moved
  its documents from `grove/` to `agreed/`
- WHEN the owner runs tend on the cycle again
- THEN it exits 0 and the implementer has been prompted once in all

#### Scenario: Under --runner

- WHEN tend implements cycle `261114-k0000`
- THEN the implementer is prompted `/grove-implement --runner 261114-k0000`

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

### Requirement: Recover without prompting twice

When tend finds an agent it would prompt still working, as after tend was
killed mid-turn, the system shall wait for it and not prompt it again; if the
agent is blocked or its state is unknown, the system shall stop for a person
without prompting it; and a prompt that timed out or stalled shall count as
possibly delivered, so no rerun sends it again, a timeout (`TIMEOUT_MS`,
default 4 hours) never stopping the agent.

#### Scenario: The implementer outlives a timeout

- GIVEN tend stopped with exit 3 because the implementer's prompt timed out,
  and `status` says `needs-you: impl-<attempt>: timeout`
- WHEN the implementer is still working and the owner runs tend again
- THEN tend says the implementer is still working on an earlier prompt, waits
  for it, prompts `/grove-implement` once in all, and exits 0

#### Scenario: The reviewer outlives a timeout

- GIVEN tend stopped because the reviewer's prompt timed out, and the
  reviewer is still working
- WHEN the owner runs tend again
- THEN the reviewer is started once and prompted once, and tend exits 0 with
  a passing verdict

### Requirement: Run files stay out of the candidate

The system shall keep an attempt's own files in its run folder under
TEND_STATE, resolving a relative TEND_STATE from where tend starts, and never
in the candidate, so the candidate stays clean with no Git excludes; and
where a checkout holds a `.tend/` folder from the prototype, it shall move
each file in it to the run folder unless the run folder holds one of the same
name, read a `repaired` mark as `repair-asked`, and stop for a person listing
any file it could not move.

#### Scenario: No excludes

- GIVEN a repository with no Git excludes
- WHEN tend takes a cycle to a verdict
- THEN `git status --porcelain` in the candidate prints nothing and the
  verdict is under `TEND_STATE/runs/`

#### Scenario: A relative TEND_STATE

- WHEN the owner runs tend with `TEND_STATE=rel`
- THEN it reaches a verdict and the candidate stays clean

#### Scenario: An attempt from the prototype

- GIVEN a reviewed attempt whose verdict and a `repaired` mark are in
  `.tend/` in its checkout
- WHEN the owner runs tend on the cycle
- THEN it exits 0 reusing that verdict with no second review,
  `repair-asked` is in the run folder, and the checkout has no `.tend/`

### Requirement: The attempt's status

While an attempt runs, the system shall keep `status` in its run folder as a
line `<time> <kind>: <what>`, where kind is `working`, `needs-you` or `done`,
followed by `next: <next action>`, replacing the file whole so an interrupted
write never leaves part of one.

#### Scenario: A finished attempt

- GIVEN tend took a cycle to a passing verdict after one repair
- WHEN the owner reads the run folder's `status`
- THEN its first line ends `done: success for <branch> at <short head>,
  verdict <file>` and its second says a person reviews the branch

## Known discrepancies

None recorded.

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
- [261117-bez30](../../changes/archive/261117-bez30-the-implementer-leaves-delivery-to-tend/) The implementer leaves delivery to tend
- [261122-nh8aa](../../changes/archive/261122-nh8aa-build-an-approved-change-by-name/) Build an approved change by name
- [261122-9vbmc](../../changes/archive/261122-9vbmc-several-attempts-at-once/) Several attempts at once
