# Terminals

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

tend runs its agents in Herdr or rex, where the owner can watch and take over, and this covers which one, where a worktree's agents run, how an agent is started, found and observed, how a prompt is delivered, the dialogs tend answers, how tend tells a person, and how a worktree's place is closed. A place is where one worktree's agents run: a Herdr workspace, or a rex session. An agent is found by its name: `impl-<attempt>` implements, `rev-<short head>` reviews, `prop-<request>-<cksum of the repository's path>` drafts a proposal.

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
`--disallowedTools "Bash(git push:*)" "Bash(gh pr:*)"` and
`--append-system-prompt` with tend's terms: the session works for tend on
its branch in this worktree, pushes nothing, opens or edits no pull request
and creates no branch or worktree, as tend delivers the work, and this wins
over any other instruction about delivering it, an implementer also being
told to commit its work on the branch and a drafter to commit only when tend
asks; a reviewer as Codex with `-s
workspace-write -c sandbox_workspace_write.network_access=true -a never
--add-dir <run folder>`, so the tests it runs may serve on localhost; and if an agent's pane is
gone, a new session shall take its name and find the work in Git.

#### Scenario: The implementer's flags

- WHEN tend starts the implementer of an attempt
- THEN Herdr is asked to start `impl-<attempt>` with `--add-dir
  TEND_STATE/runs/<attempt>`, `git push` and `gh pr` denied, and a system
  prompt naming the attempt's branch and telling it to commit there

#### Scenario: The drafter's flags

- WHEN `tend propose` starts a drafter
- THEN `git push` and `gh pr` are denied to it, and its system prompt says
  tend publishes the work and to commit only when tend asks

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

### Requirement: The trust dialogs are answered

If an agent starts at the dialog asking whether to trust its folder, at
Claude Code's dialog asking whether to add tend's run folder with
`--add-dir`, or at a notice whose hint offers `esc skip`, such as Codex's
"Update available", then the system shall answer it, by choosing to trust or
add the folder, Down then Enter at "Yes, I trust this folder" or "Yes, add
it", or Enter at "Trust and continue", or by pressing Esc alone at the notice
and saying on standard error that it skipped that agent's startup notice, and
wait for the agent to be ready, answering each in turn when one follows
another; it shall pause before answering, as
Claude Code refuses keys that reach a dialog before it takes focus, and answer
again a dialog still on screen after an answer, even when the terminal calls
the agent ready; and if an agent fails to start for any other reason, the
system shall stop for a person, saying it did not start.

If an agent starts at Codex's dialog asking to review new or changed hooks,
then the system shall stop for a person, saying so, and send it no key, even
when the terminal calls the agent ready or the screen offers `esc skip`: a trusted hook runs outside the
sandbox, and a hook in the candidate is the implementer's to write.

#### Scenario: A new folder under rex

- GIVEN under rex agents that start at the trust dialog
- WHEN tend takes a cycle to a verdict
- THEN rex is sent the keys `down enter`, and tend exits 0 with a passing
  verdict

#### Scenario: Adding the run folder under rex

- GIVEN under rex agents that start at the dialog for adding a directory
- WHEN tend takes a cycle to a verdict
- THEN rex is sent the keys `down enter`, and tend exits 0 with a passing
  verdict

#### Scenario: A dialog that refuses the first answer under rex

- GIVEN under rex agents at the dialog for adding a directory, which refuses
  the first answer and is called idle behind it
- WHEN tend takes a cycle to a verdict
- THEN rex is sent `down enter` twice for an agent, and tend exits 0 with a
  passing verdict

#### Scenario: New hooks under rex

- GIVEN under rex an agent that starts at the dialog asking to review hooks
- WHEN tend starts it
- THEN tend exits 3 saying it asks to review new or changed hooks, and rex
  is sent no key

#### Scenario: A startup notice under rex

- GIVEN under rex agents that start at Codex's "Update available" notice,
  whose hint is "enter continue · esc skip"
- WHEN tend takes a cycle to a verdict
- THEN rex is sent the key `esc` for each agent and never `enter` at the
  notice, tend says it skipped each agent's startup notice, and it exits 0
  with a passing verdict

#### Scenario: A screen tend does not know under rex

- GIVEN under rex an agent that starts at a screen that is neither a trust
  dialog nor a notice offering `esc skip`
- WHEN tend starts it
- THEN tend exits 3 saying the agent did not start, and rex is sent no key

### Requirement: Telling a person

When tend stops for a person or finishes an attempt, the system shall tell
them through the terminal: under Herdr a notification whose title names the
pull request, else the request, else the work item, and says it waits for
them (`tend: <ref> waits for you`), or that it landed (`tend: <ref>
landed`), with a body saying what the person is asked to do and ending with
the link to that pull request or request where there is one; under rex by reporting tend's own state, `blocked` for a stop or
`done` for an end, with that title and body as its message, to the terminal
tend runs in.

#### Scenario: A stop under Herdr

- GIVEN tend under Herdr stops because issue #7 of `o/r` lacks `tend-build`
- WHEN it exits 3
- THEN Herdr shows a notification `tend: o/r#7 waits for you` whose body
  says the issue is not labelled tend-build and ends with the issue's URL

#### Scenario: A stop under rex

- GIVEN the same stop under rex
- WHEN it exits 3
- THEN tend reports the state `blocked` to its terminal with a message that
  begins `tend: o/r#7 waits for you` and holds the issue's URL

#### Scenario: A landing

- GIVEN a flow that lands pull request #42 of `o/r`
- WHEN tend lands it
- THEN the notification is titled `tend: o/r#42 landed` and does not say it
  waits for anyone

### Requirement: An ended attempt says what the owner is asked to do

When an attempt's flow ends with a published head waiting for a person,
the system shall say in the notification's body what that head's review
concluded and the choices tend acts on: for a head that passed review,
approve to land or request changes where the flow's `publish` acts on an
approval, and merge or request changes where it does not; for a head that
failed review, that Codex's findings stand, after one repair where the
repair was spent, and to request changes to send them to the implementer
or close the pull request.

#### Scenario: A supervised head that passed review

- GIVEN a flow whose `publish` sends `approved` to `land`, and head 4df1bd8
  of pull request #42 of `o/r` that passed review
- WHEN the flow ends
- THEN the notification is titled `tend: o/r#42 waits for you` and its body
  says review passed at 4df1bd8, to approve to land or request changes, and
  ends with the pull request's URL

#### Scenario: A head that passed review where approval lands nothing

- GIVEN the built-in build flow and a published head that passed review
- WHEN the flow ends
- THEN the body says to merge it or request changes, and says nothing of
  approving

#### Scenario: Findings that stand after the repair

- GIVEN a published head whose review failed after the attempt's one repair
  was spent
- WHEN the flow ends
- THEN the body says Codex's findings stand after one repair, to request
  changes to send them to the implementer or close, and ends with the pull
  request's URL

#### Scenario: A failure the reviewer did not judge repairable

- GIVEN a published head whose review failed and was not marked actionable,
  so no repair was asked
- WHEN the flow ends
- THEN the body says Codex's findings stand, to request changes to send
  them to the implementer or close, and does not mention a repair

#### Scenario: No pull request

- GIVEN a repository whose origin is not listed in `TEND_STATE/publish`,
  and a head of branch `tend/issue-7` that passed review
- WHEN the flow ends
- THEN the body names the branch and its head, and ends with the URL of
  issue #7, the request the work came from

### Requirement: A hold at land says what held it

When tend stops at `land` because the head may land only with a person's
approval, the system shall say in the notification's body what held it and
to approve this head to land it, ending with the pull request's URL.

#### Scenario: A protected path

- GIVEN a head of pull request #42 of `o/r` that passed review and changes
  `tend/flow.yaml`, a protected path, with no approval
- WHEN tend reaches `land` and stops
- THEN the notification is titled `tend: o/r#42 waits for you`, and its
  body names `tend/flow.yaml`, says to approve this head to land it, and
  ends with the pull request's URL

### Requirement: Failed checks after the one fix are said

When an attempt's flow ends at a published head on which a check still
failed after the step `publish` gives for `failed` was spent, as the
publication spec's "A spent fix is said on the pull request" says, the
system shall say in the notification's body which checks failed after the
one fix, and to fix the branch, run the checks again or close the pull
request, ending with the pull request's URL, and shall not say the head is
ready to approve or merge.

#### Scenario: Still red after the fix

- GIVEN a flow whose `publish` sends `failed` to a `fix` step with limit 1,
  and pull request #42 of `o/r` whose check `test` fails on its head and
  again on the head `fix` makes
- WHEN the flow ends
- THEN the notification is titled `tend: o/r#42 waits for you`, and its
  body names `test`, says it failed after the one fix, says to fix the
  branch, run the checks again or close, ends with the pull request's URL,
  and says nothing of approving

### Requirement: Another stop gives its reason and the link

When tend stops for a person for any other reason, the system shall give
the stop's reason as the notification's body, followed by the link to the
attempt's pull request where it has one and otherwise to its request where
it has one.

#### Scenario: A stop after publishing

- GIVEN an attempt that published pull request #42 of `o/r`, whose
  implementer then committed nothing for a requested change and is not
  working on it
- WHEN tend stops
- THEN the notification is titled `tend: o/r#42 waits for you`, and its
  body gives that reason and ends with the pull request's URL

### Requirement: One notification for each wait

If a notification is the same as the last one sent for the same attempt or
request, then the system shall not send it again.

#### Scenario: A died tend run again by watch

- GIVEN an attempt whose flow ended at a published head that passed review,
  notified once, and whose status was then left `working` by a tend that
  died before writing `done`
- WHEN `tend watch` runs tend on it again and the flow ends at the same head
- THEN no second notification is shown

#### Scenario: A new head

- GIVEN that attempt notified for head 4df1bd8, and a change a person
  requested that the implementer commits as a new head that passes review
- WHEN the flow ends again
- THEN a notification for the new head is shown

### Requirement: Short enough for a lock screen

The system shall keep a notification's title to at most 60 characters and
its body to at most 240, with the link on the body's last line and whole,
shortening the text before the link and marking the cut with an ellipsis.

#### Scenario: A long stop reason

- GIVEN a stop whose reason is 600 characters long, on an attempt with a
  pull request
- WHEN tend notifies
- THEN the body is at most 240 characters, ends with the pull request's
  whole URL, and the reason is cut with an ellipsis before it

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
- [261117-bez30](../../changes/archive/261117-bez30-the-implementer-leaves-delivery-to-tend/) The implementer leaves delivery to tend
- [261121-m9bxm](../../changes/archive/261121-m9bxm-skip-startup-notices/) Skip startup notices
- [261125-kvarn](../../changes/archive/261125-kvarn-say-what-the-owner-is-asked/) Say what the owner is asked to do in each notification
