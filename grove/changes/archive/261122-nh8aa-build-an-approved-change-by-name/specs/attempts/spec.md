# Changes to attempts

The changes that change 261122-nh8aa proposes to the requirements of
`attempts`, the maintained spec at
[grove/specs/attempts/spec.md](../../../../specs/attempts/spec.md). This is
future work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## ADDED Requirements

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
