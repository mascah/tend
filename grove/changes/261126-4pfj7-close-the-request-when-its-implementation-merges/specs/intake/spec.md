# Changes to intake

<!-- The changes that change 261126-4pfj7 proposes to the requirements of
`intake`. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled. -->

Maintained spec: [intake](../../../../specs/intake/spec.md), at 7e339d9.

## ADDED Requirements

### Requirement: The intake folder records the changes drafted

When the system publishes a request's proposal, including when it finishes
an interrupted handoff, it shall write to `changes` in the request's intake
folder, before it records `linked`, the IDs of the changes the proposal's
branch adds under `<root>/changes/`, one per line, the same IDs its summary
review names.

#### Scenario: A proposal from an issue

- GIVEN issue #1 labelled `tend`, whose drafter commits change
  `261115-p0000` on `tend/issue-1`
- WHEN the owner runs `tend propose <repo>`
- THEN the intake folder `…-1` holds `changes` reading `261115-p0000`, and
  `linked`

#### Scenario: A request no proposal came from

- GIVEN a request the drafter declined
- WHEN `tend propose` says why on it
- THEN its intake folder, renamed as declined, holds no `changes`
