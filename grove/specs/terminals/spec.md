# Terminals

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

tend runs its agents in Herdr or rex, where the owner can watch and take over, and this covers which one, where a worktree's agents run, how an agent is started, found and observed, how a prompt is delivered, the one dialog tend answers, how tend tells a person, and how a worktree's place is closed. A place is where one worktree's agents run: a Herdr workspace, or a rex session. An agent is found by its name: `impl-<attempt>` implements, `rev-<short head>` reviews, `prop-<request>` drafts a proposal.

## Boundaries

This covers tend's use of the terminal and of the agents' harnesses. When an
agent is prompted, and what tend does with a state it reports, are
[attempts](../attempts/spec.md), [review](../review/spec.md) and
[intake](../intake/spec.md); when a worktree is removed is
[watch-and-status](../watch-and-status/spec.md). The agents' own behaviour,
and Herdr's and rex's, are outside it.

## Requirements

### Requirement: Herdr or rex

The system shall run the agents in rex when `TEND_TERMINAL` is `rex`, or when
it is unset and tend itself runs in rex (`REX_SESSION` set) and not in Herdr
(`HERDR_ENV` not `1`), and in Herdr otherwise.

#### Scenario: rex chosen

- GIVEN `TEND_TERMINAL=rex`
- WHEN tend takes a cycle to a verdict
- THEN it asks Herdr nothing, and each agent runs in a rex block labelled with
  its name

### Requirement: A place for each worktree

When an attempt or a proposal needs a new worktree, the system shall have
Herdr make it with `herdr worktree create`, opened as a workspace of its own;
or, under rex, make it with Git at `TEND_STATE/worktrees/<repository
folder>-<branch, slashes as hyphens>` and open a rex session named after that
folder, recorded in `TEND_STATE/rex-sessions` as `<session-id> <path>`; and a
later run shall open the existing worktree's place, finding a rex session by
that record while rex still has it, whatever the session is now named.

#### Scenario: A rex worktree

- GIVEN `TEND_TERMINAL=rex`
- WHEN tend claims and builds a cycle
- THEN its worktree is under `TEND_STATE/worktrees/`, a rex session is named
  after the worktree's folder, and the reviewer's block is gone after its
  verdict

#### Scenario: An agent renamed the session

- GIVEN a reviewed attempt under rex whose session an agent renamed
- WHEN the owner runs tend again
- THEN `TEND_STATE/rex-sessions` still has one line and nothing is reviewed
  again

### Requirement: How agents start

When the system starts an agent, it shall start an implementer or drafter as
Claude Code with `--permission-mode auto --add-dir <run or intake folder>`,
and a reviewer as Codex with `-s workspace-write -a never --add-dir <run
folder>`; and if an agent's pane is gone, a new session shall take its name
and find the work in Git.

#### Scenario: The implementer's flags

- WHEN tend starts the implementer of an attempt
- THEN Herdr is asked to start `impl-<attempt>` with `--add-dir
  TEND_STATE/runs/<attempt>`

### Requirement: An agent's state is what its terminal reports

The system shall take an agent's state from Herdr's agent status, or under
rex from what the agent reports with rex's program status protocol, where a
rex agent with no report, as at a dialog it does not report, or reporting
`error`, is unknown and a state the protocol does not know is idle.

#### Scenario: Blocked under rex

- GIVEN under rex an implementer that reports `blocked` when its head fails as
  actionable
- WHEN tend would ask the repair
- THEN it exits 3 saying the implementer is blocked

#### Scenario: Silent under rex

- GIVEN under rex an implementer that reports nothing
- WHEN tend would ask the repair
- THEN it exits 3 saying the implementer is unknown, prompts nothing and
  writes no `repair-asked`; once it reports idle, the next run asks the repair
  and passes

### Requirement: Prompts are delivered once and confirmed

When the system prompts an agent, it shall under Herdr use `herdr agent prompt
--wait` with the `TIMEOUT_MS` timeout; and under rex it shall send the text as
a bracketed paste when the agent has asked for one, then Enter, and count the
prompt delivered only when the agent reports a new state, which may be its
final done, since Codex reports nothing while it works; reporting
`agent_prompt_stalled` when none comes within the prompt's timeout; then wait
until the agent is idle, done or blocked.

#### Scenario: A paste under rex

- WHEN tend prompts the implementer under rex
- THEN rex is sent `/grove-implement …` between bracketed-paste markers

#### Scenario: A prompt never taken up under rex

- GIVEN under rex an implementer that never takes up its prompt
- WHEN tend prompts it
- THEN tend exits 3 saying `agent_prompt_stalled`; once the agent is working,
  the next run says it is still working on an earlier prompt, waits, prompts
  `/grove-implement` once in all, and exits 0

### Requirement: The trust dialog is answered

If an agent starts at the dialog asking whether to trust its folder, then the
system shall answer it by choosing to trust the folder, Down then Enter at
"Yes, I trust this folder" or Enter at "Trust and continue", and wait for the
agent to be ready; and if an agent fails to start for any other reason, the
system shall stop for a person, saying it did not start.

#### Scenario: A new folder under rex

- GIVEN under rex agents that start at the trust dialog
- WHEN tend takes a cycle to a verdict
- THEN rex is sent the keys `down enter`, and tend exits 0 with a passing
  verdict

### Requirement: Telling a person

When tend stops for a person or finishes an attempt, the system shall tell
them through the terminal: under Herdr a notification titled `tend <item>
needs you` with the reason, or `tend <item>: <conclusion>` saying the branch
or pull request is ready; under rex by reporting tend's own state, `blocked`
or `done` with that message, to the terminal tend runs in.

#### Scenario: A stop under Herdr

- GIVEN tend under Herdr stops because an issue lacks `tend-build`
- WHEN it exits 3
- THEN Herdr shows a notification `tend issue-7 needs you` whose body says the
  issue is not labelled tend-build

### Requirement: A worktree's place is closed with it

When the system removes a worktree, it shall first close its place: under
Herdr, remove the worktree's open workspace through Herdr; under rex, kill
every session recorded for the worktree in `TEND_STATE/rex-sessions` and
forget it there.

#### Scenario: A merged attempt under rex

- GIVEN under rex a published attempt whose pull request is merged
- WHEN the next `tend propose` runs in that repository
- THEN the worktree is gone, rex was asked to kill its session, and
  `TEND_STATE/rex-sessions` is empty

## Known discrepancies

None recorded.

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
