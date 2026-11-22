# Changes to watch-and-status

The changes that change 261122-nh8aa proposes to the requirements of
`watch-and-status`, the maintained spec at
[grove/specs/watch-and-status/spec.md](../../../../specs/watch-and-status/spec.md).
This is future work: nothing here describes the target branch until the change
is implemented and the maintained spec is reconciled.

## MODIFIED Requirements

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
  changes in no cycle", `261114-k0000 One: claimed by attempt…`, the attempt
  `(261114-k0000): done: success…`, the proposal `Propose a thing` under
  "Proposals to review", `Add the file` with "changes requested" under
  "Implementations to review", and `#2 Request 2` but not `#1` under
  "Requests waiting for tend propose"

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
