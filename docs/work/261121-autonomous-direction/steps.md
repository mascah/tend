# Steps toward autonomous development

The work that [direction.md](direction.md) implies, in order, so none of it is
lost. Update a step's status line when it is proposed, approved or built. The
reasons for each step are in the direction document; this file only tracks
them.

Written 2026-11-21.

## 1. tend: run steps from a flow file

Status: change 261121-gapde, approved in pull request #12, built on branch
`grove/261121-gapde-run-steps-from-a-flow-file`, waiting for review.

tend is rewritten around steps and a flow file the project holds
(`tend/flow.yaml`), with today's build and intake loops as built-in flows and
Go tests in place of `dev/test.sh`.

Nothing left to settle before it is built.

## 2. Grove: scheduling leaves merges

Status: not started. Needs its own explore, in Grove's repository.

Approval stays a merge. Planning a cycle and marking it `unattended` stop
being merges: an approved change is ready when everything it waits on is
archived as finished on the target and nothing in flight is planned to touch
the same requirements or files. More than one piece of work can be claimed.
This changes Grove's constitution, rules 10 and 11.

To settle first:
- What is claimed: a cycle or a change. Whether a cycle survives as an
  optional "build these together", or goes away.

## 3. tend: several pieces of work at once, and a merge train

Status: not started. Needs steps 1 and 2.

How many pieces run at once is a setting in the flow file. Finished
candidates land through a merge train: each is tested by CI combined with the
target and the candidates ahead of it. A candidate that fails in the train
goes back to its implementer and its new head is reviewed again.

To settle first:
- Who runs the train: GitHub's merge queue, or tend as a step. The merge
  queue waits for every required check on the train's own commit, so
  `tend/review` would have to be reported there too, and whether a repository
  owned by a personal account can use it is not verified.
- Whether the reviewer's verdict stays with the head it reviewed and is not
  repeated for each position in the train. The direction document assumes so;
  the owner has not confirmed it.
- What the attempts spec says about the one-tend lock and about worktrees
  that appear during a turn, both of which parallel work changes.

## 4. tend: the autonomous flow

Status: not started. Needs step 1.

The steps a project opts into so that work flows without a person: decompose
a goal into requests, check a proposal, plan. With them, what still needs the
owner:

- The rules agents work under, as protected paths: the flow file, CI, the
  ruleset, CODEOWNERS, the principles file, `grove.yaml`, agent instructions,
  skills, deploy config.
- What the product is, and the parts of money and data no file shows: judged,
  on both the proposal and the built change.
- Spend: a daily cap counting every prompt tend sends unattended, checked
  before each.
- Agent ideas: work needed for an outcome the owner accepted flows; an
  outcome an agent adds is a pitch and waits. A pull request that approves a
  proposal holds only Grove documents; a built pitch stays on an unmerged
  branch.

To settle first:
- Where goals, pitches and discovered work are written down.
- What a built pitch is: a throwaway to try, or the real change parked on an
  unmerged branch.
- When no person approves a proposal, who answers its open questions, and
  which step makes the judgments above.
- Whether a project can ask for a person's approval before production. The
  direction document assumes so; the owner has not confirmed it.

## 5. tend: the standing agent

Status: not started. Needs step 4.

One agent the owner talks to. It writes down goals and requests, reads status
and reports. It holds no authority to approve, plan or merge.

To settle first:
- Where it runs, how the owner reaches it, what it may write, and how it
  recovers lost context.

## 6. The typing app

Status: not started. Needs steps 3 to 5. It has no repository yet.

A touch-typing practice site, fun to use and not a trainer, as the first
project in the autonomous mode. What makes it fun is its own explore, in its
own repository. The setup guide at
`docs/work/261116-runner-contract-and-pipeline/new-project-setup.md` was
written before this direction and will need revising against it.

## Parked

- A database for the work record or the engine's state (Dolt, Doltgres). Left
  out. Revisit when there is a second runner, a team sharing one view, or the
  standing agent's answers get slow or wrong.
- Two ceilings accepted for now: agents run as the OS user that holds the
  bot's token, so a determined agent could post its own `tend/review`; and
  two changes in flight that cross a boundary only in combination are tested
  together in the train but not judged together.

## Done on the way

- The 60-second wait under rex before Claude's dialog for adding the run
  folder is answered: fixed by change 261121-7a3cz, in pull request #13.
