# Grove and tend: direction for autonomous development

Direction from the owner, including the follow-up review of Grove step 2.
[Steps](steps.md) tracks delivery. The background describes the starting
point before the tend rewrite. Lines marked (inference) await confirmation.

## Background

- **Grove** keeps what is agreed, as Markdown in Git: maintained specs, changes
  (proposals with tasks) and cycles (changes planned to be built together). A
  change is approved when its proposal is merged into the target branch. Grove
  launches no agent. `grove claim` records an attempt on the branch
  `grove/runs` with a compare-and-swap push, and refuses while the latest
  attempt's cycle is not done, so a project builds one cycle at a time.
- **tend** is a Go program (about 2,400 lines, standard library only) that takes
  one approved cycle through one attempt: a worktree, a Claude implementer in a
  terminal multiplexer, a separate Codex reviewer that writes a verdict file
  tend validates, one repair, then a pull request published by a separate bot
  account with a `tend/review` commit status. It never merges. It also drafts
  proposals from labelled GitHub or Jira issues, and `tend watch` runs all of
  this on a timer. It recovers after a kill without prompting an agent twice.
  Its agreed behaviour is seven maintained specs: attempts, review, publication,
  intake, terminals, watch-and-status, work-items.
- **Today's human steps per change:** label an issue; merge the proposal; plan a
  cycle and mark it `unattended` (a merge); merge the implementation.
- **tend's acceptance suite** is one Bash file (1,329 lines, 72 scenarios, 326
  checks) that drives the binary through stand-ins for the multiplexers, `gh`
  and `curl`. It passes on `main` (ed385ad). It runs serially in 7 min 37 s on
  the owner's machine, and the CI job takes about 15 minutes. There is no way
  to run one scenario. Its fixture hand-writes Grove's documents: a Grove
  release that changed the cycle file format made about half the checks fail
  until the fixture was updated, although tend itself was unaffected. CI pins
  Grove's version.
- **Comparison studied:** beads (a work-item store that knows nothing of
  orchestration; its charter gives scheduling and assignment to the layer
  above), Gas Town (standing agents hold authority: a Mayor decomposes and
  assigns, a Refinery merges, no approval step, no gating independent review)
  and Gas City (an orchestrator drives workflows defined as data outside any
  session; judgment is done by fresh agents in steps whose artifacts are
  checked; gates are check scripts, never an agent; the Mayor is a window).

## Decisions

1. **Autonomy.** Agents propose, plan, build, merge and file the work they
   discover. The owner keeps direction: goals and principles.
2. **Reach.** Support both supervised and autonomous work. The owner chose
   “one claim model for both workflows.” A person can select a change or tend
   can select it under the project's flow. Human approval and the number of
   simultaneous builds are separate choices. Compatibility with the old
   timer is a migration question.
3. **What still needs the owner in that mode:**
   - (a) The rules agents work under: the flow file (decision 7), CI, the
     repository ruleset, CODEOWNERS, the principles file, `grove.yaml`, agent
     instructions, skills, deploy config. These are protected paths. tend reads
     the flow file and these rules from the target branch, never from a
     candidate, so a candidate's edit to them cannot authorize that candidate.
   - (b) What the product is: a change that would shift its identity. This
     needs judgment.
   - (c) Money and data: a database, auth, a paid API, a new dependency, storing
     what a visitor typed. Protected paths hold the parts a file shows (the
     lockfile, deploy config). The rest needs judgment, like (b), on both the
     proposal and the built change, because an ordinary code change can store
     data. A change that crosses (b) or (c) waits for the owner.
   - (d) Spend: a daily cap that counts every prompt tend sends unattended, in
     any step, to a fresh agent or a continuing one, checked before each
     prompt. Building, repairing, drafting, decomposing and prototyping all
     stop when it is reached.

   Approval before production is not needed for the first project but should be
   something a project can ask for (inference).
4. **Agent-originated ideas** wait for the owner's yes or no. Work needed for an
   outcome the owner accepted flows without the owner. An outcome an agent adds
   is a pitch and waits; a broad goal approves no new outcome by itself. Some
   pitches need to be built or prototyped before they can be judged. A pull
   request that approves a proposal holds only Grove documents; a built pitch
   stays on an unmerged branch until accepted, and then goes through the
   implementation flow like any change. Later, with trust, agent ideas may ship
   without waiting.
5. **A standing agent** the owner talks to is part of the first version of the
   autonomous mode. It is a window: it writes down goals and requests, reads
   status and reports. It holds no authority to approve, plan or merge.
6. **tend drives.** Every delegated job is a step: an agent given a prompt, a
   file it must write that tend checks, and a rule for what that file allows
   next (go on, retry up to a limit, stop for a person). A step either starts a
   fresh agent or continues the agent of an earlier step: review is fresh, and
   a repair or a person's requested changes continue the implementer's
   conversation, as the review and publication specs say today. The flow file
   must be able to say which. Today only review works this way.
7. **How work flows is a file the project holds**, not code in tend. Adding a
   step, removing a human approval, or running a different flow in another
   project is an edit to that file, made by the owner (decision 3a). tend
   becomes the thing that runs steps.
8. **Approval stays a merge; scheduling stops being one.** An approved change is
   ready when everything it waits on is archived as finished on the target
   (Grove's existing meaning; finished on an attempt's branch releases nothing)
   and nothing in flight is planned to touch the same requirements or files.
   That overlap check only reduces collisions, since an implementer may touch
   files its plan did not name; the merge train (decision 9) is the guarantee.
   How many pieces run at once, and which first, is a setting in the flow file.
9. **Several pieces of work run at once**, easily controlled by the person
   running it, and finished candidates land through a merge train: each is
   tested by CI combined with the target and the candidates ahead of it. A
   candidate that fails in the train or no longer merges goes back to its
   implementer, and its new head is reviewed again. Amended 2026-11-21: the
   train takes one candidate at a time, where this first said "with several
   tested at once" (decision 16); and the owner confirmed that the reviewer's
   verdict stays with the head it reviewed and is not repeated for each
   position in the train (decision 17).
10. **No Dolt or other database for now.** Revisit only if parallel work shows a
    need for a store with many writers and queries.
11. **Order.** tend is rewritten around steps before the new project (a typing
    practice web app) starts. The owner first chose full parity with today's
    suite as the bar, then reversed it: "rewrite them all! `tend` is a project
    I literally started a week ago, we can rewrite the whole thing." The seven
    specs still say what the built-in flows do; the old suite is not a bar to
    pass first.
12. **Test suite.** Rewritten in Go with the rewrite, replacing `dev/test.sh`.
    The earlier plan, to split the Bash suite per spec first and move to Go
    afterwards, is dropped.
13. **What a cycle is for.** Claim approved changes individually by default.
    Keep a cycle when there is a concrete reason to review and land several
    changes together. A dependency can express build order on its own.
    Changes contributing to the same outcome can land independently.
14. **Outcome identity and authority.** The owner chose “Grove ID and
    owner-approved text in Git.” A pitch gets a permanent Grove ID when
    first recorded. Acceptance keeps that ID. A title or path can change,
    and several GitHub or Jira requests can refer to the same outcome.
    Store the owner's accepted scope on the target branch. The standing
    agent can draft an outcome or an amendment. Only the owner's approval
    gives that text authority; a proposed expansion stays pending. A
    change referring to an outcome still follows the proposal approval
    and owner-boundary rules in decisions 3 and 4.
15. **What the graph expresses.** A change's outcome reference says why it
    exists. Its `depends_on` references say which changes must land first.
    A discovery link records where a request or pitch came from. Following
    that link grants no authority to build a new outcome. The graph should
    answer “Which changes serve this outcome, and what is holding them up?”
    Read current attempts and planned overlap when answering readiness.
    These can change during execution without changing the relationships.

    For example, both proposals in Grove PR 182 serve the outcome of
    removing scheduling merges. Named claims depend on reconciliation,
    which can land on its own. A new product idea found while building
    reconciliation can point back to it and wait for the owner's decision.
    This example illustrates the intended graph; the drafts predate it.

Decisions 16 to 22 settle step 3. The owner made them in an explore on
2026-11-21, each by choosing the option quoted or by "Yes" to the round
before the proposals were written.

16. **tend runs the train, one candidate at a time.** The owner chose "tend,
    one at a time" over tend testing several at once and over GitHub's merge
    queue. A `land` step a flow opts into merges the target into the front
    candidate, waits for CI on that head, then squash-merges it. GitHub's
    "branch must be up to date" rule is the guarantee. The ceiling is one
    landing for each CI run. tend merges only in a project whose flow file
    has the step. GitHub documents its merge queue only for repositories an
    organization owns; tend and Grove are owned by a personal account.
17. **The verdict carries across a clean merge of the target.** "Carried if
    the merge is clean": tend checks the new head is exactly Git's merge of
    the reviewed head and the target, and only CI judges the combination. Any
    commit the implementer makes is a new head and is reviewed again.
18. **A person's approval is an approving review.** In a supervised flow a
    review that approves the head, from a person with write access, sends the
    candidate to the train. An autonomous flow sends every head that passed
    review.
19. **The count is attempts being worked on**, on this runner, in that
    repository: building, in review, repairing or in the train. An attempt
    that waits for a person, or is reviewed and waits for a merge, frees its
    slot. One setting, `watch: {at_once: N}` in the flow file, opts a project
    in: watch then starts every piece Grove calls claimable, oldest first.
    The person running watch can lower N for their machine and never raise
    it. Without the setting watch behaves as before.
20. **Containment allows for tend's other attempts.** A worktree or branch
    another attempt or intake pass of the runner made is not the
    implementer's doing; anything else new still stops the attempt. The lock
    is one tend per work item, in place of one per runner.
21. **A step is asked again with no person only when its agent is gone.**
    Watch runs again an attempt whose tend died; a surviving agent is waited
    for; a step whose agent has no pane is asked once more in a new session.
    An agent that is idle, blocked or unknown with no result stays a
    person's. Parking needs nothing new: a stopped attempt keeps its change
    in flight, so nothing claims it again.
22. **Requested changes count only from people with write access.** tend
    acted on a change request from any GitHub account. The owner chose a
    bugfix, built first; approvals follow the same rule.

## Implied order (inference)

1. tend: rewrite around steps and a flow file, with Go tests. Built as
   change 261121-gapde.
2. Grove: reconcile stale requirements, then allow independent approved
   changes and optional cycles to be claimed by name. Built and released in
   Grove 0.14.0.
3. tend: parallel work and the merge train, built as changes 261122-fh9h2,
   261122-nh8aa, 261122-9vbmc, 261122-rpkqa and 261122-hjhmn and released
   in tend 0.4.0; then the autonomous flow's new steps (decompose a goal,
   check a proposal against decision 3, plan); the standing agent. Each
   with its own tests.
4. The typing app, as the first project in the autonomous mode.

Revised 2026-11-25 by the owner: the typing app starts before the
autonomous flow's steps and the standing agent, supervised, and those are
explored once it has produced real pitches. [Steps](steps.md#order-revised-2026-11-25)
records the order.

## Step 2 review findings

The review covered [Grove PR 182](https://github.com/mascah/grove/pull/182)
at `9043b19` against `07d319f`, and its companion
[constitution amendment](https://github.com/mascah/grove/pull/181).
These findings describe the drafts at that revision.

- **Within a cycle**, two members can name different task files and rewrite
  the same requirement. Both builders can read the old text before either
  archives. Apply the overlap check to members before they start and read
  the maintained requirement again after the earlier member archives.
- **Across changes**, B can depend on A and still omit a clause A gained
  during implementation. Two changes can also add different text under the
  same new requirement name. Compare B with what actually landed in both
  cases. Include an example where B already incorporates the landed text.
- **At claim time**, today's watcher can read a cycle's `runs` value and
  lose a race before it calls `grove claim <cycle-id>`. If a project
  requires serial execution, its limit must hold when the claim is recorded
  across participating runners. An earlier readiness read is advisory.
- **After a stop**, scanning all historical unarchived claims can revive a
  superseded pre-upgrade attempt and pause a previously working timer.
  Retrying replaces the attempt and keeps the change in flight. Specify
  the upgrade behavior, how to release a particular attempt for retry, and
  how to park a change for an owner answer while unrelated work proceeds.
  A late release must leave a newer attempt alone.

## What the comparison established

Beads gives work separate links for grouping, dependency and discovery.
Claiming, returning work to the queue and closing it have distinct effects.
Its Dolt-backed implementation versions an issue's owner and status and
keeps lease renewals local to the store. A shared database can coordinate
writers. Independent replicas see one another's updates after sync.
[Beads work graph](https://github.com/gastownhall/beads/blob/main/docs/core-concepts/index.md),
[claim transaction](https://github.com/gastownhall/beads/blob/main/internal/storage/issueops/claim.go)
and [lease recovery](https://github.com/gastownhall/beads/blob/main/cmd/bd/reclaim.go).

Gas City supports a file provider (`GC_BEADS=file`) as well as Beads. Its
file store coordinates processes sharing a file and persists ownership
revisions. Its controller compares the recorded work with the actual agent
session when deciding to resume or restart. This is relevant to tend's
existing promise to recover without prompting a surviving agent twice.
[Gas City file store](https://github.com/gastownhall/gascity/blob/main/internal/beads/filestore.go)
and [health patrol](https://github.com/gastownhall/gascity/blob/main/engdocs/architecture/health-patrol.md).

Multiple hosts claiming different work need one authority for reservations.
Grove already conditionally updates a shared Git ref. Taking over work from
a lost host also needs an accessible candidate commit and step checkpoint,
and a way to prevent the returning attempt from delivering. Support for a
work graph takes priority. Dolt, Beads integration and cross-host takeover
remain options to evaluate against a concrete need.

## Open questions

- How Grove records evidence of owner acceptance for an outcome revision,
  how a runner identifies the revision used for a particular decision, and
  the file format and reference syntax. Decision 14 settles the identity
  and authority.
- Lease timing, takeover by another machine and the choice of a shared
  store. Recovery on one machine is settled by decision 21.
- Where to keep broad goals and discovered requests before they become
  outcome pitches or changes, and how trackers expose their Grove links.
- What a built pitch is: a throwaway to try, or the real change parked on an
  unmerged branch.
- When no person approves a proposal, who answers its open questions, and which
  step judges decision 3b, the judged part of 3c, and whether work is needed for
  an accepted outcome or is a new one (decision 4).
- The standing agent: where it runs, how the owner reaches it, what it may
  write, and how it recovers lost context.
- Known ceiling carried over: agents run as the same OS user that holds the
  bot's token, so a determined agent could post its own `tend/review` status.
- Known ceiling, accepted for now: decisions 3b and 3c are judged on each
  change against its own base. Two changes in flight that cross a boundary only
  in combination are tested together by CI in the train but not judged
  together. Judging every train position would slow the train to an agent's
  pace. Revisit when a project holds real user data.
