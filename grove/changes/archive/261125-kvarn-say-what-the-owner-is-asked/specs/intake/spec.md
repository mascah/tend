# Changes to intake

The changes that change 261125-kvarn proposes to the requirements of
[intake](../../../../specs/intake/spec.md), as that spec stands at 7e339d9.
This is future work: nothing here describes the target branch until the
change is implemented and the maintained spec is reconciled.

The examples use the title wording and the counting of consequential open
questions that the proposal's decisions 1 and 2 fix.

## ADDED Requirements

### Requirement: The owner is told of a proposal

When intake has opened a proposal pull request and posted its review, the
system shall notify the person, as
[terminals](../../../../specs/terminals/spec.md) says for a wait, with a
title naming the pull request and saying it waits for them, and a body
naming the request it was drafted from, how many of its open questions the
drafter marked consequential with the follow prompt's fixed marker, and to merge to approve or revise on its
branch, ending with the pull request's URL; and shall notify each proposal
pull request once.

#### Scenario: A proposal with consequential questions

- GIVEN issue #43 of `o/r` labelled `tend`, drafted with a summary that
  marks two open questions consequential
- WHEN `tend propose` opens pull request #44 for it
- THEN the notification is titled `tend: o/r#44 waits for you`, and its
  body names #43, says two open questions are consequential, says to merge
  to approve or revise on its branch, and ends with the pull request's URL

#### Scenario: A summary with none marked

- GIVEN a drafted proposal whose summary marks no question consequential
- WHEN `tend propose` opens its pull request
- THEN the body gives no count and says to see its open questions

#### Scenario: The next intake pass

- GIVEN that proposal notified once, its pull request still open
- WHEN `tend watch` runs `tend propose` again
- THEN no second notification is shown

## Observed current behavior

- `intake.go` at 7e339d9 opens the proposal pull request, posts its review
  and records `done` in the request's status ("a person answers its open
  questions, revises or merges it"), and sends no notification.
- `flows/prompts/follow.md` at 7e339d9 asks the drafter to write the open
  questions to `summary`, "the consequential ones ... first and marked so",
  with no fixed marker.
