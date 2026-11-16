# Attempts

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

`tend <repo> <cycle-id>` takes one approved Grove cycle through one attempt, claiming it or carrying on the attempt it has, implementing it until Git shows a finished candidate, and recovering after any stop or kill without prompting an agent twice. An attempt is one execution of a work item; its run folder is `TEND_STATE/runs/<attempt>/`, and TEND_STATE is `$TEND_STATE`, else `$XDG_STATE_HOME/tend`, else `~/.local/state/tend`. The owner running tend at a terminal, and `tend watch` running it for them, rely on this.

## Boundaries

This covers the command line and exit statuses of an attempt, the runner
lock, finding or claiming a Grove cycle's attempt, where its worktree starts,
implementing until the branch holds a finished candidate, recovery after a
stop or kill, and the run folder and its `status` record. These hold for
every attempt, whatever its work item. Reviewing the head and the one repair
are [review](../review/spec.md); pushing and the pull request are
[publication](../publication/spec.md); a GitHub or Jira issue as the work item
is [work-items](../work-items/spec.md); how agents are started, prompted and
observed in Herdr or rex is [terminals](../terminals/spec.md); `tend status`,
`tend watch` and tidying finished worktrees are
[watch-and-status](../watch-and-status/spec.md); `tend propose` is
[intake](../intake/spec.md). Grove, not tend, owns approval, eligibility and
the claim.

## Requirements

### Requirement: Exit statuses

When an attempt ends, the system shall exit 0 once the head is reviewed,
whatever the verdict's conclusion, printing the verdict JSON on standard
output and notifying that the branch or pull request is ready; exit 3 when a
person is needed, saying why on standard error, recording it in `status` as
`needs-you` and notifying through the terminal; exit 4 when another tend holds
the runner; and exit 2 with `usage: tend <repo> <cycle-id>` when no
repository or work item is given.

#### Scenario: A cycle reaches a verdict

- GIVEN a repository with origin and cycle `261114-k0000` approved and planned
- WHEN the owner runs `tend <repo> 261114-k0000` and the reviewer passes the
  head
- THEN tend exits 0 and standard output is the verdict, whose `conclusion` is
  `success`

#### Scenario: The reviewer leaves no valid verdict

- GIVEN a reviewer that writes a verdict with no `conclusion`
- WHEN the owner runs tend on the cycle
- THEN tend exits 3, saying the reviewer settled without a valid verdict

#### Scenario: No work item named

- WHEN the owner runs `tend <repo>`
- THEN tend prints `usage: tend <repo> <cycle-id>` and exits 2

### Requirement: One tend per runner

While a tend holds the lock in TEND_STATE, the system shall refuse another
`tend <repo> <item>` or `tend propose <repo>` using that TEND_STATE, after
waiting about two seconds for the lock to come free, naming the holder's pid,
work item and repository, prompting no agent and exiting 4; and the lock
shall be released when its holder exits or is killed, whatever agents outlive
it.

#### Scenario: A second tend while one runs

- GIVEN a tend whose implementer is still on its turn
- WHEN the owner runs tend on the same cycle again
- THEN it exits 4, standard error names `pid <first tend's pid>`, and the
  implementer has been prompted once

#### Scenario: The holder is killed

- GIVEN that first tend is killed with `kill -9`
- WHEN the owner runs tend on the cycle again
- THEN it takes the lock, says it is carrying on the attempt, and exits 0

### Requirement: Carry on the attempt or claim the cycle

When tend runs on a Grove cycle, the system shall carry on the newest attempt
record for the cycle in `grove changes --json` whose checkout still holds its
branch, and otherwise make a new worktree, run `grove claim <cycle-id>` inside
it and delete the temporary `tend-HHMMSS` branch the worktree was made with,
stopping for a person if the claim is refused; with or without origin.

#### Scenario: A rerun after a stop

- GIVEN a repository with origin where tend claimed the cycle and then stopped
  because the implementer's prompt failed
- WHEN the owner runs tend on the cycle again
- THEN standard error says `carrying on attempt <id>`, Grove holds one record
  for the cycle, and tend exits 0

#### Scenario: Without origin

- GIVEN a repository with no remote, where tend claimed the cycle and stopped
- WHEN the owner runs tend on the cycle again
- THEN it carries on the same attempt and exits 0

#### Scenario: The temporary branch goes

- WHEN tend claims a cycle
- THEN the repository has no `tend-*` branch afterwards

### Requirement: New work starts from origin's target

When an attempt needs a new worktree, the system shall start it from
`origin/<target>` after fetching the target from origin, or from the local
target when the repository has no origin, where the target is the branch
Grove names (an issue build's is the repository's default branch), and shall
show the reviewer the diff from that same base; and if the fetch fails it
shall stop for a person.

#### Scenario: The local target is behind

- GIVEN someone else pushed a commit to origin's `main` that the local `main`
  lacks
- WHEN the owner runs tend on the cycle
- THEN the attempt's branch contains that commit, and tend exits 0

#### Scenario: A project targeting trunk

- GIVEN a repository without origin whose `grove.yaml` names target `trunk`
- WHEN the owner runs tend on the cycle
- THEN the worktree is created with base `trunk`, and the reviewer is told
  the change is `git diff trunk...HEAD`

### Requirement: Implement until Git shows a finished candidate

When the attempt's branch has uncommitted files or a change of the cycle that
is not under `<root>/changes/archive/` on it, the system shall prompt the
implementer `impl-<attempt>` with `/grove-implement <cycle-id>` and, once its
turn ends, stop for a person naming what is still missing; and it shall never
prompt the implementation of a branch that already holds a finished
candidate, judging that from Git, never from the agent's state.

#### Scenario: The documents moved on the target

- GIVEN an attempt already built and reviewed, after which the target moved
  its documents from `grove/` to `agreed/`
- WHEN the owner runs tend on the cycle again
- THEN it exits 0 and the implementer has been prompted once in all

### Requirement: Recover without prompting twice

When tend finds an agent it would prompt still working, as after tend was
killed mid-turn, the system shall wait for it and not prompt it again; if the
agent is blocked or its state is unknown, the system shall stop for a person
without prompting it; and a prompt that timed out or stalled shall count as
possibly delivered, so no rerun sends it again, a timeout (`TIMEOUT_MS`,
default 30 minutes) never stopping the agent.

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

### Requirement: Run files stay out of the candidate

The system shall keep an attempt's own files in its run folder under
TEND_STATE, resolving a relative TEND_STATE from where tend starts, and never
in the candidate, so the candidate stays clean with no Git excludes; and
where a checkout holds a `.tend/` folder from the prototype, it shall move
each file in it to the run folder unless the run folder holds one of the same
name, read a `repaired` mark as `repair-asked`, and stop for a person listing
any file it could not move.

#### Scenario: No excludes

- GIVEN a repository with no Git excludes
- WHEN tend takes a cycle to a verdict
- THEN `git status --porcelain` in the candidate prints nothing and the
  verdict is under `TEND_STATE/runs/`

#### Scenario: A relative TEND_STATE

- WHEN the owner runs tend with `TEND_STATE=rel`
- THEN it reaches a verdict and the candidate stays clean

#### Scenario: An attempt from the prototype

- GIVEN a reviewed attempt whose verdict and a `repaired` mark are in
  `.tend/` in its checkout
- WHEN the owner runs tend on the cycle
- THEN it exits 0 reusing that verdict with no second review,
  `repair-asked` is in the run folder, and the checkout has no `.tend/`

### Requirement: The attempt's status

While an attempt runs, the system shall keep `status` in its run folder as a
line `<time> <kind>: <what>`, where kind is `working`, `needs-you` or `done`,
followed by `next: <next action>`, replacing the file whole so an interrupted
write never leaves part of one.

#### Scenario: A finished attempt

- GIVEN tend took a cycle to a passing verdict after one repair
- WHEN the owner reads the run folder's `status`
- THEN its first line ends `done: success for <branch> at <short head>,
  verdict <file>` and its second says a person reviews the branch

## Known discrepancies

- Exit statuses: README.md at 38b21d0 lists exits 0, 3 and 4 only. The code
  (main.go) also exits 2 with a usage line when no work item is given, and 1
  when a command it relies on fails unexpectedly.
- One tend per runner: README.md at 38b21d0 says exit 4 means "another tend
  is running on this machine". The lock is the file `lock` in TEND_STATE
  (main.go `takeLock`), so tends with different TEND_STATE folders run
  together.
- Purpose: README.md at 38b21d0 gives TEND_STATE's default as
  `~/.local/state/tend`; the code uses `$XDG_STATE_HOME/tend` when
  `XDG_STATE_HOME` is set (main.go `main`).
- README.md at 38b21d0, "What it asks of Grove", names three things:
  `grove changes --json`, `grove claim` and `/grove-implement`. tend also
  runs `grove review` for the head's published review
  ([publication](../publication/spec.md)), prompts `/grove-propose
  --unattended` ([intake](../intake/spec.md)), and takes a successful `grove
  changes --json` as the sign a repository uses Grove
  ([work-items](../work-items/spec.md)).

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
