# Changes to flows

The changes that change 261122-rpkqa proposes to the requirements of `flows`,
the maintained spec at
[grove/specs/flows/spec.md](../../../../specs/flows/spec.md). This is future
work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## ADDED Requirements

### Requirement: Landing is tend's own and a flow places it

The system shall land a candidate only at a step named `land`, as the landing
spec says, with no agent prompted to merge; `land` shall hold only what
follows its outcomes `ok`, `conflict` and `failed`; a flow that has `land`
and no `publish`, or whose `land` holds more than what follows, shall stop as
"A flow that is not valid" says; a loop that comes round only through
`publish`'s `approved` shall need no limit, as people bound it; and the
built-in flows shall hold no `land`, so that tend merges nothing in a project
whose flow has none.

#### Scenario: A build that lands

- GIVEN a `build` flow whose `publish` sends `ok` to `land`, and `land`
  sends `ok` to the end
- WHEN tend takes a cycle to a passing verdict and its checks pass
- THEN the pull request is merged, and no agent was prompted to merge it

#### Scenario: A flow without land

- GIVEN a repository with no flow file
- WHEN tend takes a cycle to a passing verdict
- THEN the pull request is open and tend merged nothing

#### Scenario: Land without publish

- GIVEN a `build` flow that has `land` and no `publish`
- WHEN the owner runs tend
- THEN tend exits 3 with one sentence naming `tend/flow.yaml` and `land`,
  and no agent is started

#### Scenario: Land given a prompt

- GIVEN a `build` flow whose `land` names a prompt
- WHEN the owner runs tend
- THEN tend exits 3 saying landing is tend's own, and no agent is started

#### Scenario: A conflict step with no limit

- GIVEN a `build` flow whose `publish` sends `ok` to `land`, whose `land`
  sends `conflict` to a step that leads back to `review`, and no step of
  that loop has a limit
- WHEN the owner runs tend
- THEN tend exits 3 saying `tend/flow.yaml` loops with no limit

#### Scenario: A flow names tend's own prompts

- GIVEN a `build` flow whose steps after `conflict` and `failed` name the
  prompt files tend ships for them, and holds no such file beside the flow
  file
- WHEN the owner runs tend
- THEN the flow is valid, and those steps are prompted with tend's own text
