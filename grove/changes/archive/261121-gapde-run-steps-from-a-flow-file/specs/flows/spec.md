# Changes to flows

The changes that change 261121-gapde proposes to the requirements of `flows`.
This is future work. There is no maintained spec for `flows` at ed385ad, so
every requirement is ADDED and today's behaviour is under "Observed current
behavior".

A flow is a named set of steps that an attempt goes through once it has its
branch and worktree. A step is an agent given a prompt, a result tend checks,
and what each outcome of that result allows next. The flow file is
`tend/flow.yaml` in the project.

## ADDED Requirements

### Requirement: The built-in flows

Where the target branch holds no flow file, the system shall run its built-in
flows, `build` for a Grove cycle or a work item and `propose` for a request,
whose behaviour is what the attempts, review, publication, intake and
work-items specs say.

#### Scenario: A project without a flow file

- GIVEN a repository whose target branch holds no `tend/flow.yaml`, with one
  approved cycle
- WHEN the owner runs tend on the cycle
- THEN the cycle is implemented, reviewed, repaired at most once and
  published as those specs say, and tend exits 0

### Requirement: A project's flow replaces the built-in one of its name

Where the target branch holds `tend/flow.yaml`, the system shall run each flow
that file names in place of the built-in flow of that name, and the built-in
flow for a name it does not hold.

#### Scenario: A build with one more step

- GIVEN a flow file whose `build` flow has a step `check` between `review` and
  `publish`, and no `propose` flow
- WHEN tend builds a cycle and later drafts a request
- THEN the build prompts `check` after a passing review and before
  publishing, and the request is drafted as the intake spec says

### Requirement: The flow is read from the target branch

The system shall read the flow file and the prompt files it names from the
target branch as origin holds it, fetched before the attempt acts, and never
from the attempt's branch or worktree.

#### Scenario: The implementer edits the flow

- GIVEN an attempt whose implementer commits a `tend/flow.yaml` that drops the
  `review` step
- WHEN tend carries on the attempt
- THEN the head is reviewed as the target's flow says, and the edit is part
  of the candidate like any other file

### Requirement: A step prompts once and its result is checked

When an attempt reaches a step, the system shall prompt the step's agent once
with the step's prompt, filled in with the attempt's branch, base, head, run
folder and work item, wait for the agent's turn to end, and then check the
result the step names: a new commit on the attempt's branch, a file in the run
folder that is not empty, or a verdict file as the review spec defines one. If
the result is not there, the system shall stop for a person, naming the step
and what is missing.

#### Scenario: A step that leaves no result

- GIVEN a step `check` whose result is the file `check-<head>.json`, and an
  agent whose turn ends without writing it
- WHEN tend runs the flow
- THEN tend stops with exit 3 saying `check` wrote no `check-<head>.json`,
  and prompts no later step

### Requirement: A fresh agent or a continuing one

Where a step names an earlier step to continue, the system shall prompt that
step's agent in its own conversation and start none; otherwise it shall start
an agent of the harness the step names, `claude` or `codex`, with the flags
the terminals spec gives that harness, a flow setting none of them.

#### Scenario: A repair continues the implementer

- GIVEN the built-in `build` flow and a review that fails as actionable
- WHEN tend asks the repair
- THEN the prompt goes to the implementer's own agent, and the agents started
  in the attempt are one implementer and one reviewer for each head

#### Scenario: A flow tries to set a flag

- GIVEN a flow file whose step has a field naming a command-line flag
- WHEN the owner runs tend
- THEN tend stops as "A flow that is not valid" says

### Requirement: An outcome chooses what follows

When a step's result is checked, the system shall take the outcome it gives
(`ok` for a commit or a file; `success`, `actionable` or `failure` for a
verdict; `spent` for a step at its limit) and go to what the step gives for
that outcome: another step, the end of the flow, or a person. An outcome the
step gives nothing for shall stop for a person.

#### Scenario: A failing check goes to a person

- GIVEN a step `check` that gives `success` the step `publish` and gives
  `failure` nothing
- WHEN its verdict is a failure
- THEN tend stops with exit 3 naming `check` and the verdict's summary, and
  nothing is published

### Requirement: A step is asked no more often than its limit

Where a step has a limit, the system shall prompt it at most that many times
in one attempt. When the flow reaches it again, the system shall not prompt
it, and shall go to what the step gives for the outcome `spent`, or stop for
a person where it gives nothing.

#### Scenario: A second repair

- GIVEN a `build` flow whose `repair` step has the limit 2 and gives `spent`
  nothing, and a reviewer that fails three heads as actionable
- WHEN tend runs the attempt
- THEN the repair is prompted twice, three heads are reviewed, and tend stops
  with exit 3 at the third failure

#### Scenario: The built-in one repair

- GIVEN the built-in `build` flow and a reviewer that fails both heads as
  actionable
- WHEN tend runs the attempt
- THEN one repair is asked, and tend exits 0 with the failed verdict, as the
  review spec's "One repair" says

### Requirement: Any step recovers without being prompted twice

When tend is about to prompt a step, the system shall first record in the run
folder that the step was asked, and at which head. On a later run, an agent
still working on that prompt shall be waited for; a step recorded as asked
whose result is there shall be taken as done; and a step recorded as asked
with no result and no agent working shall stop for a person on every run,
until the person supplies the result or deletes the record to have it asked
once more.

#### Scenario: tend is killed during a project's own step

- GIVEN a flow with a step `check`, and tend killed while the `check` agent
  is working
- WHEN the owner runs tend again and the agent finishes
- THEN `check` was prompted once in all, its result is taken, and the flow
  goes on

#### Scenario: The outcome of a step is unknown

- GIVEN `check` recorded as asked, no result, and its agent not working
- WHEN the owner runs tend again
- THEN tend stops with exit 3 saying `check` was asked and left nothing, and
  prompts no agent

### Requirement: Publishing is tend's own and a flow places it

The system shall publish a head only at a step named `publish`, as the
publication spec says, with no agent prompted to push or open a pull request,
and shall publish nothing in a flow that has no such step.

#### Scenario: A flow without publish

- GIVEN a `build` flow that ends after `review`
- WHEN tend runs it on a repository whose origin is listed in `publish`
- THEN the verdict is printed, no branch is pushed and no pull request is
  opened

### Requirement: A flow that is not valid

If the flow file is not YAML, holds a field tend does not know, names a step,
harness, prompt file or outcome that does not exist, or holds a loop no step
of which has a limit, then the system shall stop with exit 3 before starting
or prompting any agent, with one sentence naming the file and the fault.

#### Scenario: A loop with no limit

- GIVEN a flow whose `review` gives `actionable` to `repair`, whose `repair`
  gives `ok` to `review`, and neither has a limit
- WHEN the owner runs tend
- THEN tend exits 3 saying `tend/flow.yaml` loops through `review` and
  `repair` with no limit, and no agent is started

### Requirement: Print the flows in effect

When the owner runs `tend flow <repo>`, the system shall print, reading only,
the flows that repository's attempts would run, as flow-file text: the
project's where its target branch holds one, the built-in ones otherwise, and
say which each is.

#### Scenario: Starting a project's file

- GIVEN a repository with no flow file
- WHEN the owner runs `tend flow <repo>` and saves what it prints as
  `tend/flow.yaml`
- THEN the file is valid, and attempts behave as before it was added

## Observed current behavior

- The build loop is Go code: `implementer`, `review`, `publish`, `requested`
  and the functions around them in `main.go` at ed385ad. The intake loop is
  `propose` and `draft` in the same file.
- Review alone has the shape of a step: `review` prompts a fresh reviewer and
  `valid` checks the verdict file before anything follows (`main.go` at
  ed385ad).
- One repair is a file mark, `repair-asked`, written before the prompt goes
  (`grove/specs/review/spec.md` at ed385ad, "One repair"). Requested changes
  and drafting have marks of the same kind (`request-<review id>`, `asked`,
  `followed`).
- No file in a project changes how tend runs. `TEND_STATE/publish` and
  `TEND_STATE/jira` say where it may publish and which Jira projects feed
  which repository, on the runner.
