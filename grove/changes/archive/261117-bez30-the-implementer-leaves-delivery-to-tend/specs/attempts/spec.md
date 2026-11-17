# Changes to attempts

## ADDED Requirements

### Requirement: Work stays on the attempt's branch

When an implementer turn that builds, repairs or makes requested changes
ends, and on every run before the head is reviewed, if the attempt's
worktree is on a branch other than the attempt's, or the repository has a
worktree, or a local branch holding commits the target lacks, that it did
not have just before the implementer was last prompted, the system shall
stop for a person naming what it found, and review and publish nothing.

#### Scenario: The implementer adds a worktree

- GIVEN an implementer that builds the cycle and also adds a worktree
- WHEN the owner runs tend on the cycle
- THEN tend exits 3 naming the worktree, and no reviewer is prompted

#### Scenario: A rerun finds it still there

- GIVEN tend stopped because the implementer added a worktree, which is
  still there
- WHEN the owner runs tend again
- THEN tend exits 3 again, and no reviewer is prompted

#### Scenario: The implementer makes a branch with commits

- GIVEN an implementer that builds the cycle and also makes a branch at its
  head
- WHEN the owner runs tend on the cycle
- THEN tend exits 3 naming the branch, and no reviewer is prompted

## MODIFIED Requirements

### Requirement: Implement until Git shows a finished candidate

When the attempt's branch has uncommitted files or a change of the cycle that
is not under `<root>/changes/archive/` on it, the system shall prompt the
implementer `impl-<attempt>` with `/grove-implement --runner <cycle-id>`
and, once its turn ends, stop for a person naming what is still missing; and
it shall never prompt the implementation of a branch that already holds a
finished candidate, judging that from Git, never from the agent's state.

#### Scenario: The documents moved on the target

- GIVEN an attempt already built and reviewed, after which the target moved
  its documents from `grove/` to `agreed/`
- WHEN the owner runs tend on the cycle again
- THEN it exits 0 and the implementer has been prompted once in all

#### Scenario: Under --runner

- WHEN tend implements cycle `261114-k0000`
- THEN the implementer is prompted `/grove-implement --runner 261114-k0000`

### Requirement: Recover without prompting twice

When tend finds an agent it would prompt still working, as after tend was
killed mid-turn, the system shall wait for it and not prompt it again; if the
agent is blocked or its state is unknown, the system shall stop for a person
without prompting it; and a prompt that timed out or stalled shall count as
possibly delivered, so no rerun sends it again, a timeout (`TIMEOUT_MS`,
default 4 hours) never stopping the agent.

#### Scenario: The implementer outlives a timeout

- GIVEN tend stopped with exit 3 because the implementer's prompt timed out,
  and `status` says `needs-you: impl-<attempt>: timeout`
- WHEN the implementer is still working and the owner runs tend again
- THEN tend says the implementer is still working on an earlier prompt, waits
  for it, prompts `/grove-implement` once in all, and exits 0

#### Scenario: The reviewer outlives a timeout

- GIVEN tend stopped because the reviewer's prompt timed out, and the
  reviewer is still working
- WHEN the owner runs tend again
- THEN the reviewer is started once and prompted once, and tend exits 0 with
  a passing verdict
