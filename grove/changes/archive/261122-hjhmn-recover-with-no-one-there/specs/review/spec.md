# Changes to review

The changes that change 261122-hjhmn proposes to the requirements of
`review`, the maintained spec at
[grove/specs/review/spec.md](../../../../specs/review/spec.md). This is future
work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## MODIFIED Requirements

### Requirement: An uncertain repair goes to a person

If `repair-asked` names the head under review, no commit has followed and the
implementer is there and not working, then the system shall stop for a person
on every run, saying the repair's outcome is unknown, until a person commits
a fix or deletes `repair-asked`, which has the repair asked once more; an
implementer that is gone shall be asked the repair once more in a new
session, as the flows spec's "Any step recovers without being prompted twice"
says; an implementer still working on the repair shall be waited for; and a
repair refused before its prompt went, as by a blocked implementer, shall
leave the repair unused.

#### Scenario: A stalled repair prompt

- GIVEN tend stopped with exit 3 because the repair prompt stalled, and the
  implementer is idle with no new commit
- WHEN the owner runs tend again
- THEN it exits 3 without asking the repair again, saying the repair has no
  commit and the implementer is not working on it, and `tend status` shows
  `needs-you: the repair asked`

#### Scenario: The mark deleted

- GIVEN that stopped attempt, and a person deletes `repair-asked`
- WHEN the owner runs tend again
- THEN the repair is asked a second time, the new head passes, and tend exits 0

#### Scenario: The implementer still repairing

- GIVEN the repair prompt timed out and the implementer is still working
- WHEN the owner runs tend again
- THEN tend waits for it, has the new head reviewed, exits 0, and the repair
  was asked once

#### Scenario: A blocked implementer

- GIVEN the implementer is blocked when the head fails as actionable
- WHEN tend would ask the repair
- THEN it exits 3 with no `repair-asked` written; once the implementer is
  idle, the next run asks the repair and passes

#### Scenario: The implementer is gone

- GIVEN tend was killed after asking the repair, no commit has followed, and
  the implementer's pane is closed
- WHEN the owner runs tend again
- THEN the repair is asked once more of a new implementer session, the new
  head is reviewed, and tend exits 0
