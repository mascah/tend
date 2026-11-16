# Watch and status

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

`tend status` says, reading only, what one attempt last recorded or what in a repository waits for whom; `tend watch` runs tend and `tend propose` on a person's behalf; and runs and intake passes tidy away the worktrees of finished pull requests. The owner and their teammates rely on these to know what needs them without reading run folders. "Needs a person" means the run or intake folder's `status` says `needs-you`.

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
print, each under its heading: approved changes without a cycle, marked `not
ready` where Grove says so; cycles not done, with Grove's verdict; this
runner's attempts in that repository at cycles not done and at issues, with
their last recorded stage; open proposal pull requests (from `tend/issue-` or
`tend/jira-`); open implementation pull requests (from `grove/` or
`tend/build-`), marked where changes are requested; and requests labelled
`tend`, on GitHub and in mapped Jira projects, with no proposal pull request
in any state; saying GitHub is not readable when `gh` fails there, and
exiting 0.

#### Scenario: A repository with work at every stage

- GIVEN approved change `261016-c0001` without a cycle, cycle `261114-k0000`
  claimed and published with changes requested on its head, issue #1 drafted
  as a proposal and issue #2 labelled `tend` and not yet drafted
- WHEN a person runs `tend status <repo>`
- THEN it lists `261016-c0001 Another`, `261114-k0000 One: claimed by
  attempt…`, the attempt `(261114-k0000): done: success…`, the proposal
  `Propose a thing` under "Proposals to review", `Add the file` with "changes
  requested" under "Implementations to review", and `#2 Request 2` but not
  `#1` under "Requests waiting for tend propose"

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
(default 300), run `tend propose` on each repository given unless one of its
intake folders needs a person, and run tend again on each attempt whose pull
request is open with a change request on its head that tend has not answered,
unless the attempt needs a person, saying which pull request it acts on.

#### Scenario: Changes requested

- GIVEN a published attempt and no change requested
- WHEN one pass of `tend watch` runs (`WATCH_ONCE` set), then a teammate
  requests changes and another pass runs
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

#### Scenario: Intake

- GIVEN issue #1 labelled `tend`
- WHEN a pass of `tend watch <repo>` runs
- THEN its proposal pull request opens; and once a request's follow-up has
  stalled and needs a person, a further pass asks nothing more

### Requirement: Finished worktrees are tidied

When `tend <repo> <item>` or `tend propose <repo>` runs, the system shall
remove each worktree tend made for that repository whose pull request is
merged or closed, with its local branch and its agents' place, when its head
is the pull request's head and it holds no uncommitted files, and otherwise
leave it and say why; and a declined request's worktree shall go at once.

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

## Known discrepancies

None recorded.

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
