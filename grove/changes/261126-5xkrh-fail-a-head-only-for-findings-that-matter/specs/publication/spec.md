# Changes to publication

<!-- The changes that change 261126-5xkrh proposes to the requirements of
`publication`. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled. -->

Maintained spec: [publication](../../../../specs/publication/spec.md), as at
e80da30.

## MODIFIED Requirements

### Requirement: A review and a status for each reviewed head

For each valid verdict of the attempt, oldest first, whose head is on the
branch and that the account has not reviewed, the system shall set the
`tend/review` commit status on that head, `success` for a passing verdict and
`failure` otherwise, whatever the verdict's notes hold, then post a comment
review on that commit by the account giving the conclusion and summary, the
verdict's notes, when it has any, under a heading saying they do not fail the
head, whether a failure was judged repairable or needs a person's decision,
the implementer's reply to a requested change, and, on the head of a Grove
cycle, what `grove review` says merging approves; and it shall post no review
twice.

#### Scenario: A passing head

- WHEN tend publishes a cycle's head that passed review
- THEN the head has one comment review by the account saying the review
  passed and showing what merging approves, and the status `tend/review`
  `success`

#### Scenario: A passing head with notes

- GIVEN a head whose verdict concludes `success` with `notes` naming a minor
  finding
- WHEN tend publishes it
- THEN the status `tend/review` is `success`, and the head's one review
  shows the note under a heading saying it does not fail the head

#### Scenario: A verdict without notes

- GIVEN a head whose verdict has no `notes`
- WHEN tend publishes it
- THEN its review has no notes heading

#### Scenario: A repair within one run

- GIVEN a head failed as actionable and its repair passed in the same run
- WHEN tend publishes
- THEN the failed head has its own review, saying tend repairs once, and
  status `failure`, and the repaired head has one review and status `success`

#### Scenario: A later commit that needs a decision

- GIVEN a published attempt, then a new commit whose review fails with
  `actionable` false
- WHEN the owner runs tend again
- THEN the new head gets status `failure` and a review saying a finding needs
  a decision, the earlier head keeps its one review, and there is still one
  pull request
