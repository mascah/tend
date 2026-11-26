# Changes to intake

<!-- The changes that change 261126-c16zs proposes to the requirements of
`intake`. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled. -->

Starts from the maintained spec [intake](../../../../specs/intake/spec.md) at
c6505b1.

## ADDED Requirements

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

## Observed current behavior

- Intake opens a proposal pull request and posts a summary review on its
  head, and sets no commit status (intake.go, `proposal`, at c6505b1).
- `tend/review` is set only by the publish step, for each verdict
  (publication.go at c6505b1), and by `land` for a head carried by a clean
  merge (landing.go, `carry`, at c6505b1).
- A pass looks at a request's pull request only to skip the request once one
  exists, or to finish a handoff whose folder records `pr` but not `linked`
  (intake.go, `propose`, at c6505b1); it never reads a proposal's later
  heads.
