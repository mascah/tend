# Steps toward autonomous development

The work that [direction.md](direction.md) implies, in order, so none of it is
lost. Update a step's status line when it is proposed, approved or built. The
reasons for each step are in the direction document; this file only tracks
them.

Written 2026-11-21.

## Order, revised 2026-11-25

The owner confirmed an order that departs from the numbering below, which
had the typing app wait for steps 3 to 5:

1. Run tend on tend itself, supervised, on the runner (the OS user `tend`
   on the mini). Done: tend's flow file landed in pull request #35.
2. Build the next two changes through tend, on tend itself, as the first
   live run since the flow-file rewrite: the protected-path check (issue
   #36, from step 4) and fixing failing CI before a person is asked (issue
   #37).
3. Start the typing app, `keyborg`, now: its explore first, then repository
   setup. The owner merges its proposals; tend lands on its own once the
   protected-path check exists.
4. Explore outcomes, the judge and the standing agent (steps 4 and 5) after
   keyborg has produced real pitches. The spend cap waits until agents can
   originate work.

## 1. tend: run steps from a flow file

Status: change 261121-gapde, approved in pull request #12, built and merged
in pull request #15.

tend is rewritten around steps and a flow file the project holds
(`tend/flow.yaml`), with today's build and intake loops as built-in flows and
Go tests in place of `dev/test.sh`.

Nothing left to settle before it is built.

## 2. Grove: scheduling leaves merges

Status: built, merged and released in Grove 0.14.0. Proposed in
[Grove PR 182](https://github.com/mascah/grove/pull/182); change 261122-6ndwb
(requirement reconciliation) built in PR 183 and 261122-7vht2 (named claims)
in PR 185, after the [constitution amendment](https://github.com/mascah/grove/pull/181).

Merging the proposal approves the change. A runner can claim it directly
once every dependency is archived as finished on the target and its
planned files and capabilities are free. This removes the cycle-planning
and `unattended` merges from the ordinary path and allows disjoint work
to proceed concurrently.
The companion amendment covers rules 2, 3, 4, 10 and 11.

Use one claim model for a person and for tend. Default to one approved
change. An optional cycle groups changes that need a combined candidate.
The [review findings](direction.md#step-2-review-findings) remain to be
resolved in the proposals.

## 3. tend: several pieces of work at once, and a merge train

Status: built and merged as five changes, in pull requests #21, #22, #23,
#27 and #29, and released in tend 0.4.0. tend 0.4.1 adds the fix for issue
#26 (pull request #32). Not yet run against real GitHub; the first live run
is the revised order's item 2.

How many pieces run at once is a setting in the flow file. Finished
candidates land through a merge train, one at a time: each is tested by CI
combined with the target as the candidates ahead of it left it. A candidate
that fails in the train goes back to its implementer and its new head is
reviewed again.

The changes, in the order they are built:
- 261122-fh9h2, a bugfix: requested changes count only from people with
  write access.
- 261122-nh8aa: `tend <repo> <change-id>` builds an approved change that no
  cycle lists.
- 261122-9vbmc: several attempts at once. A lock for each work item, and
  `watch: {at_once: N}` in the flow file.
- 261122-rpkqa: a `land` step that lands reviewed candidates one at a time.
- 261122-hjhmn: watch runs again an attempt whose tend died, and a step
  whose agent is gone is asked once more.

Settled, as direction decisions 16 to 22 record:
- tend runs the train as a step, one candidate at a time. GitHub documents
  its merge queue only for repositories an organization owns.
- The reviewer's verdict stays with the head it reviewed, and carries
  across a clean merge of the target.
- The lock is one tend per work item, and containment allows for what
  tend's other attempts made.
- A dead tend's attempt is run again by watch; a step is asked again only
  when its agent is gone. A parked change stays in flight in Grove, so
  nothing selects it again.

## 4. tend: the autonomous flow

Status: the protected-path check is taken early, as issue #36. The rest
waits until keyborg has produced real pitches; the spend cap waits until
agents can originate work.

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

A pitch gets its permanent Grove ID at creation. Its owner-approved text
on the target defines the accepted scope.

To settle first:
- How to verify that approval, how changes reference outcomes, how
  requests record their origin, and where broad goals and incoming
  requests live.
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

Status: started ahead of steps 4 and 5, by the revised order. Named
`keyborg`, in the new private repository `mascah/keyborg`, which holds only a
README. The old project is archived as `mascah/keyborg-retired`. Next: its
explore, then repository setup once the explore has chosen a stack.

A touch-typing practice site, fun to use and not a trainer. It starts
supervised, with the owner merging its proposals, and becomes the first
project in the autonomous mode once step 4 exists. What makes it fun is its own explore, in its
own repository. The setup guide at
`docs/work/261116-runner-contract-and-pipeline/new-project-setup.md` was
written before this direction and will need revising against it.

## Parked

- Prioritize querying the work graph. Multiple machines claiming different
  changes and another machine taking over a lost attempt are separate
  capabilities. Revisit a shared store such as Dolt if the runner or graph
  queries need it. Cross-machine takeover is a separate decision.
- Two ceilings accepted for now: agents run as the OS user that holds the
  bot's token, so a determined agent could post its own `tend/review`; and
  two changes in flight that cross a boundary only in combination are tested
  together in the train but not judged together.

## Done on the way

- The 60-second wait under rex before Claude's dialog for adding the run
  folder is answered: fixed by change 261121-7a3cz, in pull request #13.
