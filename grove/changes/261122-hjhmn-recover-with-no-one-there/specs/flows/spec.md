# Changes to flows

The changes that change 261122-hjhmn proposes to the requirements of `flows`,
the maintained spec at
[grove/specs/flows/spec.md](../../../../specs/flows/spec.md). This is future
work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

An agent is gone when the terminal holds no pane for it. An agent that is
there may be working, idle, blocked or in a state tend does not know.

## MODIFIED Requirements

### Requirement: Any step recovers without being prompted twice

When tend is about to prompt a step, the system shall first record in the run
folder that the step was asked, and at which head. On a later run, an agent
still working on that prompt shall be waited for; a step recorded as asked
whose result is there shall be taken as done; a step recorded as asked with
no result whose agent is gone shall be asked once more in a new session, that
second ask recorded too; and a step recorded as asked with no result whose
agent is there and not working, or that has already been asked once more,
shall stop for a person on every run, until the person supplies the result or
deletes the record to have it asked once more.

#### Scenario: tend is killed during a project's own step

- GIVEN a flow with a step `check`, and tend killed while the `check` agent
  is working
- WHEN the owner runs tend again and the agent finishes
- THEN `check` was prompted once in all, its result is taken, and the flow
  goes on

#### Scenario: The outcome of a step is unknown

- GIVEN `check` recorded as asked, no result, and its agent there and not
  working
- WHEN the owner runs tend again
- THEN tend stops with exit 3 saying `check` was asked and left nothing, and
  prompts no agent

#### Scenario: The agent is gone

- GIVEN `check` recorded as asked, no result, and its agent's pane closed
- WHEN the owner runs tend again
- THEN `check` is prompted a second time, in a new session, its result is
  taken, and the flow goes on

#### Scenario: Gone a second time

- GIVEN `check` asked once more because its agent was gone, still no result,
  and its agent's pane closed again
- WHEN the owner runs tend again
- THEN tend stops with exit 3 saying `check` was asked twice and left
  nothing, and prompts no agent
