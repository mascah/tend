# Watch and status

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

`tend status` says, reading only, what one attempt last recorded or what in a repository waits for whom; `tend watch` runs tend and `tend propose` on a person's behalf; and runs and intake passes tidy away the worktrees of finished pull requests. The owner and their teammates rely on these to know what needs them without reading run folders. "Needs a person" means the run or intake folder's `status` says `needs-you`. An attempt is being worked on while a tend holds its work item's lock, whoever started that tend.

## Boundaries

This covers `tend status`, with and without a work item, `tend watch`, and
removing the worktrees of merged or closed pull requests. What an attempt
records in `status` is [attempts](../attempts/spec.md); acting on requested
changes is [publication](../publication/spec.md); a pass of `tend propose` is
[intake](../intake/spec.md); closing a worktree's Herdr workspace or rex
session is [terminals](../terminals/spec.md).

## Requirements

### Requirement: Status of one work item

When a person runs `tend status <repo> <item>`, the system shall print the
newest attempt at it, as `attempt <id> of <item>: <branch>` and, for a Grove
cycle, ` in <checkout>`; then `tend is on it: <lock line>` when a tend holds
the lock for that item in that repository; then the attempt's recorded
`status`, or `nothing recorded yet`; and if the status says `working` and no
tend is on it, it shall say instead that tend stopped before this finished
and to run it again; with no attempt it shall print `no attempt at <item>`;
taking no lock, prompting no one and exiting 0.

#### Scenario: While tend runs

- GIVEN a tend whose implementer is on its turn
- WHEN a person runs `tend status <repo> 261114-k0000`
- THEN it prints `tend is on it: pid <pid>, …` and a line `working:
  impl-<attempt> implements 261114-k0000`

#### Scenario: After a kill

- GIVEN that tend was killed with `kill -9` mid-turn
- WHEN a person runs `tend status <repo> 261114-k0000`
- THEN it says tend stopped before this finished and to run it again

#### Scenario: A published attempt

- GIVEN tend published a cycle's passing head as a pull request
- WHEN a person runs `tend status <repo> 261114-k0000`
- THEN its next action says a person reviews `https://…` the pull request

#### Scenario: An issue's attempt

- GIVEN issue #7 built in a repository without Grove
- WHEN a person runs `tend status <repo> issue-7`
- THEN it prints `attempt issue-7-<cksum> of issue-7: tend/build-7`

### Requirement: What waits for whom

When a person runs `tend status <repo>` without a work item, the system shall
print, each under its heading: approved changes in no cycle, each with
Grove's verdict on whether `grove claim` would claim it; cycles not done,
with Grove's verdict; this runner's attempts in that repository at cycles and
changes not done and at issues, with their last recorded stage; open proposal
pull requests (from `tend/issue-` or `tend/jira-`); open implementation pull
requests (from `grove/` or `tend/build-`), marked where changes are
requested; and requests labelled `tend`, on GitHub and in mapped Jira
projects, with no proposal pull request in any state; saying GitHub is not
readable when `gh` fails there, and exiting 0.

#### Scenario: A repository with work at every stage

- GIVEN approved change `261016-c0001` in no cycle, cycle `261114-k0000`
  claimed and published with changes requested on its head, issue #1 drafted
  as a proposal and issue #2 labelled `tend` and not yet drafted
- WHEN a person runs `tend status <repo>`
- THEN it lists `261016-c0001 Another` with Grove's verdict under "Approved
  changes in no cycle", `261114-k0000 One: 261114-k0000 is in flight in
  attempt…`, the attempt `(261114-k0000): done: success…`, the proposal
  `Propose a thing` under "Proposals to review", `Add the file` with
  "changes requested" under "Implementations to review", and `#2 Request 2`
  but not `#1` under "Requests waiting for tend propose"

#### Scenario: A change that waits

- GIVEN approved changes A and B in no cycle, B with `depends_on: [A]`, and A
  not archived
- WHEN a person runs `tend status <repo>`
- THEN B's line holds Grove's verdict that it waits on A

#### Scenario: An attempt at a change

- GIVEN tend took `261016-c0001`, a change in no cycle, to a verdict
- WHEN a person runs `tend status <repo>`, then runs it again once the target
  holds the change archived
- THEN the first lists the attempt `(261016-c0001): done: success…`, and the
  second no longer lists it

#### Scenario: A cycle done

- GIVEN that cycle's change is archived on the target
- WHEN a person runs `tend status <repo>`
- THEN its attempt is no longer listed

#### Scenario: Jira requests

- GIVEN mapped Jira project SCRUM, whose SCRUM-1 has a proposal and SCRUM-6 is
  labelled `tend` with none
- WHEN a person runs `tend status <repo>`
- THEN "Requests waiting" lists `SCRUM-6 Request SCRUM-6
  https://jira.test/browse/SCRUM-6` and no other SCRUM key, and "Proposals to
  review" lists SCRUM-1's pull request

### Requirement: Watch acts on a person's behalf

While `tend watch [<repo>...]` runs, the system shall, every `WATCH_SECONDS`
(default 300), first fast-forward each repository's local target to origin's,
fetched (only a fast-forward, and only of a clean checkout), since Grove reads
the local target; run `tend propose` on each repository given unless one of its
intake folders needs a person; then start on each repository given the oldest
cycle that Grove would claim (`runs` in `grove changes --json`) and whose plan
marks it `unattended`, saying so; and run tend again on each attempt whose
pull request is open with a change request on its head, from a person with
write access as the publication spec's "Requested changes are acted on once"
says, that tend has not answered, unless the attempt needs a person, saying
which pull request it acts on.

#### Scenario: Changes requested

- GIVEN a published attempt and no change requested
- WHEN one pass of `tend watch` runs (`WATCH_ONCE` set), then a teammate who
  is a collaborator requests changes and another pass runs
- THEN the first runs nothing, and the second says changes were requested on
  the pull request and the request is asked of the implementer

#### Scenario: An attempt that needs a person

- GIVEN a new change request on an attempt whose `status` says `needs-you`
- WHEN a pass of `tend watch` runs
- THEN the request is not asked

#### Scenario: A request already answered

- GIVEN a request the implementer answered with no change
- WHEN a pass of `tend watch` runs
- THEN it runs nothing; once a teammate requests changes again on the same
  head, the next pass runs tend and the new request is asked

#### Scenario: A stranger's request

- GIVEN a published attempt, and an account with no access to the repository
  requests changes on its head
- WHEN a pass of `tend watch` runs
- THEN it runs nothing

#### Scenario: Intake

- GIVEN issue #1 labelled `tend`
- WHEN a pass of `tend watch <repo>` runs
- THEN its proposal pull request opens; and once a request's follow-up has
  stalled and needs a person, a further pass asks nothing more

#### Scenario: A cycle marked unattended on GitHub

- GIVEN a cycle marked `unattended` on origin's target, which the repository's
  local target does not yet hold
- WHEN a pass of `tend watch <repo>` runs
- THEN the local target is fast-forwarded to origin's and the cycle is started

#### Scenario: A cycle marked unattended

- GIVEN a repository whose only cycle is not marked `unattended`
- WHEN one pass of `tend watch <repo>` runs, then the plan is marked
  `unattended` on the target and another pass runs, then a third
- THEN the first starts nothing; the second says it is starting the cycle,
  which is claimed, built and reviewed; the third starts nothing, as the
  cycle is claimed

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

### Requirement: Finished worktrees are tidied

When `tend <repo> <item>` or `tend propose <repo>` runs, the system shall
remove each worktree tend made for that repository whose pull request is
merged or closed, when its head is the pull request's head, and each
attempt's worktree with no pull request whose Grove cycle is done and whose
head the target holds, each with its local branch and its agents' place,
when it holds no uncommitted files, no worktree is inside it and its
implementer is not working, and otherwise leave it and say why; and a
declined request's worktree shall go at once.

#### Scenario: A merged attempt and an open proposal

- GIVEN a published attempt whose pull request is merged and an open proposal
- WHEN `tend propose` runs
- THEN the attempt's worktree and local branch are gone, and the proposal's
  worktree stays

#### Scenario: Uncommitted files

- GIVEN the proposal's pull request is merged and its worktree holds an
  uncommitted file
- WHEN `tend propose` runs
- THEN the worktree stays and tend says it has uncommitted files; once the
  file is gone, the next pass removes it

#### Scenario: A merged issue build

- GIVEN issue #7's pull request is merged
- WHEN the owner runs `tend <repo> issue-7`
- THEN the worktree of `tend/build-7` is removed

#### Scenario: The cycle landed some other way

- GIVEN an attempt with no pull request and no commits, whose cycle the
  target then shows done, with a worktree inside the attempt's
- WHEN `tend propose` runs, then runs again once that worktree is gone
  while the implementer is working, then again once it is idle
- THEN the first two leave the worktree, saying a worktree is inside it and
  that its implementer is working, and the third removes it with its local
  branch

#### Scenario: The cycle is not done

- GIVEN an attempt with no pull request whose cycle is not done
- WHEN `tend propose` runs
- THEN its worktree stays

### Requirement: Watch runs an approved attempt again

While `tend watch` runs, the system shall run tend again on each attempt
whose flow's `publish` gives a next step for `approved` and whose open pull
request's head has an approval that step would act on, as the landing spec's
"A person's approval is an outcome of publishing" says, unless the attempt
needs a person, saying which pull request it acts on; and in a repository
whose flow file sets the count, it shall do so within the count and before
starting new pieces.

#### Scenario: A teammate approves

- GIVEN a flow whose `publish` sends `approved` to `land`, and a published
  candidate that passed review
- WHEN one pass of `tend watch` runs, then a teammate who is a collaborator
  approves the head and another pass runs
- THEN the first runs nothing, and the second says the pull request was
  approved and lands it

#### Scenario: The built-in flow

- GIVEN a repository with no flow file, and a teammate approves a published
  head
- WHEN a pass of `tend watch` runs
- THEN it runs nothing

#### Scenario: An attempt that needs a person

- GIVEN an approval on the head of an attempt whose `status` says
  `needs-you`
- WHEN a pass of `tend watch` runs
- THEN it runs nothing

## Known discrepancies

None recorded.

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
- [261117-bez30](../../changes/archive/261117-bez30-the-implementer-leaves-delivery-to-tend/) The implementer leaves delivery to tend
- [261122-fh9h2](../../changes/archive/261122-fh9h2-requested-changes-count-only-from-write-access/) Requested changes count only from people with write access
- [261122-nh8aa](../../changes/archive/261122-nh8aa-build-an-approved-change-by-name/) Build an approved change by name
- [261122-9vbmc](../../changes/archive/261122-9vbmc-several-attempts-at-once/) Several attempts at once
- [261122-rpkqa](../../changes/archive/261122-rpkqa-land-reviewed-candidates-one-at-a-time/) Land reviewed candidates one at a time
