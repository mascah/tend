# Changes to intake

The changes that change 261122-hjhmn proposes to the requirements of
`intake`, the maintained spec at
[grove/specs/intake/spec.md](../../../../specs/intake/spec.md). This is future
work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## MODIFIED Requirements

### Requirement: Drafting prompts are sent once

The system shall write `asked` and `followed` in the intake folder just
before sending the drafting and the follow-up prompts, never sending either
again while its mark exists, but for a follow-up whose drafter is gone; a
drafter still working shall be waited for; if the follow-up was asked and the
drafter neither committed a draft nor wrote `declined` and is there and not
working, the system shall stop for a person on every pass until a person
commits the draft, writes `declined`, or deletes `followed`; and if the
follow-up was asked, left neither, and the drafter is gone, the system shall
ask it once more of a new session, as the flows spec's "Any step recovers
without being prompted twice" says.

#### Scenario: The follow-up stalled

- GIVEN the follow-up prompt stalled and the pass exited 3
- WHEN the owner runs `tend propose` again
- THEN it exits 3 without asking again, saying to delete `followed` to have
  it asked again; once `followed` is deleted, the next pass asks a second
  time, opens the pull request, and the draft was asked for once in all

#### Scenario: The drafter outlived a timeout

- GIVEN the drafting prompt timed out and the drafter is still working
- WHEN the owner runs `tend propose` again
- THEN it waits for the drafter, has it commit, opens the pull request, and
  asked for the draft once

#### Scenario: The drafter is gone

- GIVEN a pass was killed after asking the follow-up, the drafter committed
  nothing and wrote no `declined`, and its pane is closed
- WHEN the owner runs `tend propose` again
- THEN the follow-up is asked once more of a new drafter session, and the
  pull request opens or the request is told why there is none
