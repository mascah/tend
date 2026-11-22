# Changes to flows

The changes that change 261122-9vbmc proposes to the requirements of `flows`,
the maintained spec at
[grove/specs/flows/spec.md](../../../../specs/flows/spec.md). This is future
work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## ADDED Requirements

### Requirement: The flow file sets how many attempts watch keeps going

Where the flow file holds `watch` with `at_once`, a whole number of at least
1, the system shall read it from the target branch as it reads the flows,
`tend watch` using it as the watch-and-status spec says and `tend flow`
printing it; and if `watch` holds any other field, or `at_once` is not such a
number, the system shall stop as "A flow that is not valid" says.

#### Scenario: A count and no flow

- GIVEN a flow file that holds only `watch` with `at_once: 2`
- WHEN the owner runs `tend flow <repo>`, then tend on a cycle
- THEN `tend flow` prints `watch` with `at_once: 2` and both built-in flows,
  and the cycle is built as the built-in `build` flow says

#### Scenario: No count

- GIVEN a repository with no flow file
- WHEN the owner runs `tend flow <repo>`
- THEN it prints no `watch`

#### Scenario: A count of 0

- GIVEN a flow file whose `watch` holds `at_once: 0`
- WHEN the owner runs tend
- THEN tend exits 3 with one sentence naming `tend/flow.yaml` and `at_once`,
  and no agent is started

#### Scenario: The implementer raises the count

- GIVEN a target whose flow file sets `at_once: 1`, and an attempt whose
  implementer commits a flow file with `at_once: 5`
- WHEN a pass of `tend watch` runs
- THEN it keeps at most one attempt being worked on
