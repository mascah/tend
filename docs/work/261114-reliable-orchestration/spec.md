# Tend: reliable execution before broader automation

Created 2026-11-14. Direction agreed by the owner after reviewing the Herdr
prototype, Tend, and Grove #163/#164. This is the canonical plan for this
effort; it describes intended behavior, not capabilities already delivered.
Delivery and the session handoff below own execution progress. Update this
document when scope changes rather than maintaining a competing roadmap.

## Start the next session here

Read this document and the repository instructions, then inspect the current
branch, worktrees, diff and relevant implementation before acting. The baseline
below is historical evidence, not a claim about the checkout you find.

**Next Tend work: run tend for real on the owner's mini (grove with GitHub issues,
memex with Jira, tend itself). Tend now uses Grove: its agreed behaviour is in
`grove/specs/`, and new work is proposed and planned as Grove changes, so this
document is the delivery record up to adopting Grove, not the place for new scope.**
The known gaps are closed and shown live.
Slice 4 is complete and merged into `main`: the non-Grove work item, the port to
Go, the terminal interface with Herdr and rex (shown live) and Jira as the second
tracker (shown live); tmux is dropped. Before that the order was drawn from Herdr and rex (the
owner's next multiplexer) or tmux if rex cannot report agent state; then the
remaining terminal; then a second tracker. Slices 1–3 are merged into `main`. Slices 1, 2
and 2a are done on `tend/slice1-claim-verdict` (not merged; merging is the
owner's decision), slices 2 and 2a shown live on the private sandbox
`mascah/tend-sandbox`, publishing as `mascah-bot`, with the owner as the teammate
(see the session handoff). Preserve existing attempts and candidate branches. Do
not start proposal polling or multi-repository scheduling first.

Grove #164's implementation merged as #165 (`0083bba`). Check Grove's own record for
its acceptance before treating G as finished, and coordinate with its owner before
changing that checkout. Tend slice 2 does not depend on it. Do not automatically rerun the full model trial suite: its implementation
record says the owner deferred it to conserve usage.

At the end of each implementation session, update the Delivery status and
Session handoff with the revision, evidence, remaining failures and next action.
Distinguish code written, local tests passed, live integrations exercised, and
work merged. Link PRs when they exist; do not infer completion from a commit title.

This plan records direction and acceptance; it is not an instruction to execute
every future slice whenever it is read. Follow the current session's mandate.
Later slices may need their open decisions settled before implementation. Recording
this plan does not itself publish changes, authorize a merge, or start live agents.

## Problem and intended outcome

The owner wants a small orchestrator that makes several coding harnesses work
together, supports solo and team use, integrates with existing trackers, and
leaves Grove independently useful for spec-driven development. The first team
deployment is one shared runner, not cooperating orchestrators on many laptops.

The prototype proved a useful Claude implementation / Codex review / Claude
repair loop through Herdr. Promoting its shortcuts into permanent architecture
would leave runtime state implicit, recovery unreliable, and Tend coupled to
Grove cycles and Herdr. More proposal sources and polling would amplify those
problems before solving the operator's workflow.

Tend should take agreed work, run an implementer, obtain independent review of
the candidate commit, allow one automatic repair round, and return a branch or
PR with evidence and explicit human decisions. A teammate should be able to
understand what needs attention without reconstructing terminal conversations.

The broader workflow is:

```text
Request -> proposal -> review/revision -> approval -> cycle planning
        -> implementation -> independent review -> repair -> human merge
```

Proposal revision and cycle planning may remain manual initially. Tend must
identify those handoffs rather than imply that unattended drafting closes the
whole loop. A failed review, interrupted run and successful candidate are
different outcomes even when an agent's turn has ended.

## Future product versus first delivery

The destination is an optional orchestrator around development work, with Grove
as a first-class integration. Users can choose Grove alone for SDD, Grove with
Tend, or eventually Tend with another source of scoped work. A separate repository
alone does not establish that independence; the non-Grove execution proves it.

The broader aim is useful cooperation between harnesses throughout development,
not a permanent restriction to code review. Implementation and review are the
first bounded proof. Proposal/design review by another harness is a possible
later application of the same artifact handoff; it is not committed scope for
the current slices, nor a mandate to automate human product decisions.

| Dimension | Initial delivery | Intended direction / boundary |
| --- | --- | --- |
| Work source | Explicitly selected, approved Grove cycle | Grove is optional; another source can supply scoped, eligible work |
| Harness roles | Claude implements, Codex independently reviews | Harness roles are replaceable; conversations are not a shared protocol |
| Terminal | Herdr | tmux also works without requiring Herdr; terminal and harness choice stay separate |
| People | One person operating locally or teammates using one shared runner | Shared requests, evidence and human decisions; distributed dispatch is deferred |
| Capacity | One active implementation and bounded repair | Serialization is an initial simplification, not a permanent product identity |
| Planning | Human-assisted proposal revision and Grove cycle planning | Handoffs are explicit; future automation requires separate scope |
| Tracker | One provider after the execution loop works | GitHub, GitLab, Jira and Linear are intended integrations, not all promised by the first provider slice |
| Repository coverage | One repository | Multi-repository polling follows a demonstrated workflow |

The original Beads/Gas Town comparison expresses the coordination problem, not
a feature-parity target. Borrow durable work identity, handoffs, isolation and
visibility. Grove and an existing tracker already own much of the work record;
do not add a competing issue tracker, agent society or memory system. Symphony
provided a useful reference for a deterministic runner, not a requirement to
adopt its implementation or rebuild its full specification.

Success is less operator reconstruction and fewer lost or duplicated handoffs.
Smallness is measured by the concepts and responsibilities needed to operate
the tool, not by keeping the prototype under a particular number of lines.

## Decisions and ownership

| Owner | Responsibility |
| --- | --- |
| Grove | Agreed scope, approval, dependencies, planning, implementation evidence and archive |
| Tend | Execution attempts, stages, runtime identity, interruptions, budgets, review results and publication |
| Terminal integration | Process/terminal access and runtime observations, initially through Herdr |
| Harness integration | Start, observe, stop and optionally resume a particular coding agent |
| Tracker | Requests, discussion, priority, shared status and links |
| Git and PR | Candidate revision, review target, human approval and merge |

- Keep Tend separate from Grove. Grove must remain useful without Tend, and
  Tend's execution lifecycle must eventually work with a non-Grove work item.
- Keep `grove claim` for now, inside the Grove integration. Use the existing
  `grove changes --json` contract rather than inventing a second readiness model.
  Claiming cycles pauses that project until the cycle lands; accept that limit
  initially and revisit it only for a concrete workflow that needs more.
- Tend owns execution state. A controller lock and small durable run records
  are sufficient starting mechanisms; neither a database nor a general workflow
  engine is required. Runtime records should live outside candidate worktrees.
- Herdr is the first terminal integration. tmux remains a later required proof
  of portability. Terminal status alone never proves work or review succeeded.
- Start with Claude implementing and Codex reviewing. Keep harness operations
  together so the lifecycle does not depend on one harness's conversation format.
- Retain one automatic repair round and bounded execution. A saved session is
  useful for continuity, but artifacts must permit recovery if the session is gone.
- Publication authority (settled 2026-11-15): a runner publishes only to origins
  it lists, each with the GitHub account to publish as. That account, a bot user
  (`mascah-bot` for the owner), makes every remote write: the push, the PR, its
  reviews and statuses. People approve and merge. The PR description says what
  the change is for and is written once; each verdict is a separate comment
  review on its commit, never folded into the description. A GitHub App (check
  runs, short-lived tokens) waits for a need a machine user cannot meet.
- Humans retain scope approval and merge authority. A tracker label requesting
  execution does not approve an unapproved Grove proposal.
- Use deterministic code for identity, deduplication, polling and routine
  publication; use agents for interpretation, drafting and review. Provider APIs
  belong in integrations. Do not build a universal plugin framework in advance.
- Factory and Crew are abandoned foundations. Neither revival nor archival is
  a prerequisite. Keep the useful Tend experiment; do not restart from scratch.

### Terminology and execution ownership

Keep these meanings separate in code, status and documentation:

| Term | Meaning |
| --- | --- |
| Work item | Scope supplied by a work source; a Grove cycle is the first concrete case, not the universal schema |
| Approved | Scope accepted under that source's rules; for Grove, held by the target branch |
| Eligible | Currently allowed to start under dependencies, plan freshness and source policy; approval alone does not guarantee this |
| Attempt | One execution of a work item; it may be interrupted or fail without changing the approved scope |
| Stage | A part of the attempt such as implementation, review, repair or publication; restarting a controller is not automatically a new attempt |
| Candidate | The concrete code revision being checked or reviewed |
| Reviewed | A validated review exists for that candidate; it may conclude failure |
| Archived on the candidate | Grove's finishing artifacts exist on that branch; this does not prove review success or merge |
| Merged | The target has accepted the candidate; a local completed turn, archive or passing verdict does not imply it |

Retain references to source records instead of copying Grove's spec/task graph
into Tend's own tracker. Tend's state records execution facts, not a second
authority for scope or approval. The non-Grove integration must define its own
eligibility and completion evidence rather than emulate Grove folders or cycles.

For the initial cycle-based loop, Tend advances implementation/review/repair
stages; Grove's cycle instructions may direct the implementing harness's own
subagents. Do not simultaneously introduce a second Tend scheduler dispatching
the same cycle's individual changes. If finer-grained or cross-harness scheduling
within cycles becomes necessary, explicitly reconsider that ownership boundary.

Grove's constitution and maintained specs govern changes to Grove, including its
current claim behavior. This Tend plan does not amend them or impose Grove's code
and instruction budgets on Tend. Extracting claim from Grove is not required;
neither is keeping it there forever if an observed need later justifies a change.

## Acceptance

### Recoverable execution

One controller owns advancement on the shared runner. A duplicate invocation
cannot prompt the same attempt again or advance it concurrently. On restart,
Tend reconciles recorded state, the workspace and any surviving agent before
deciding to observe, resume, retry or request human attention.

An attempt records enough to identify the work and its input revision, workspace
and branch, stage, runtime identity, review target, repair allowance, outcome
and next action. Persistence must not treat an interrupted write as completed
work. The precise file format and command spelling are implementation choices.

Status explains the stage, interruption or question, relevant artifacts and next
action. Cancel and resume preserve useful work. A wait timeout is not assumed to
have stopped an agent: Tend establishes whether it is still running before
restarting or releasing ownership. Human intervention pauses automatic advancement
until the resulting state has been reconciled.

Demonstrate interruption during implementation, review and repair, including
controller death while an agent survives. Resuming must neither lose the branch
nor silently skip a repair whose prompt was never delivered. An uncertain repair
outcome must be visible instead of triggering an unbounded retry.

Runtime artifacts must not dirty the candidate, require fixture-only Git excludes,
or enter its commits. Preserve or explicitly reconcile legacy `.tend/` artifacts;
do not silently discard a prior verdict or reset a consumed repair allowance.

### Review integrity

New and cached verdicts receive the same validation: complete expected fields,
valid conclusion, matching candidate commit and a suitable candidate checkout.
A filename alone is insufficient. Missing, partial, malformed or wrong-head
verdicts do not count as a completed review. Any later code change invalidates
the old review for the new candidate.

The reviewer receives the agreed scope, candidate diff and validation evidence.
Scope revisions on the candidate remain visible to human review. Reviewer or
terminal completion is not equivalent to passing acceptance. A failed review
remains explicit when repair is exhausted or a human decision is needed.

### Grove integration and repository configuration

Both a repository without a remote and one with `origin` can be claimed and
resumed. A successful remote claim is discoverable without a local `grove/runs`
branch. A claim followed by controller failure is reconciled as the same attempt.
Use Grove's published project/run data where it supplies the answer.

Respect the configured target branch for workspace creation and review; do not
hardcode `main`. Grove continues to own eligibility, stale-plan and dependency
semantics. Ordinary repositories work without the ledger fixture's private setup.

### Shared review and publication

For an already approved and planned cycle on a real remote, Tend produces or
updates one PR with the candidate branch, reviewed commit, validation evidence,
findings and unresolved questions. Repeating publication after interruption must
not duplicate the PR or its logical verdict publication. A changed head cannot
retain a passing result for the previous head.

A second person can inspect the outcome and act through the PR, with human merge
authority intact. Team use initially means several people sharing one runner and
tracker, not several dispatchers or simultaneous implementations.

### Request intake and explicit handoffs

Initially accept explicitly labelled requests from one tracker. Keep stable
source IDs, source-to-proposal links, deduplication and a cap on unmerged proposal
PRs. Requests that cannot be drafted have a visible outcome instead of being
repeated on every poll. Do not overwrite teammates' tracker prose indiscriminately.

An unattended Grove proposal uses #164 after it is available. The user can see
whether a proposal needs answers or manual revision, whether approved work needs
cycle planning, or whether an implementation needs review. Merging a proposal
with unresolved questions has Grove's documented consequence: implementation may
choose answers within its rules. Surface that consequence, and distinguish routine
choices from consequential scope decisions.

### Portability

Before calling Tend generic, demonstrate one non-Grove work item and one execution
through a second terminal (rex, 2026-11-15; tmux was dropped). The non-Grove path supplies its own scope and eligibility without
requiring Grove files. The same lifecycle, review and recovery guarantees hold.

GitHub, GitLab, Jira and Linear remain the intended tracker integrations. Prove
one first, then use a second provider to test the boundary before claiming broad
support. Harness, terminal, tracker and work-source choices must not be one
inseparable integration.

## Evidence and baseline

These are observations from the review, not claims that the problems are fixed.

| Artifact | Reviewed revision / observation |
| --- | --- |
| Tend | `60e9f4a`, initial script, README and fixture |
| Grove prototype | `prototype/herdr-loop` at `4c5804d`, local throwaway branch, never merged |
| Grove constitution | [#163](https://github.com/mascah/grove/pull/163), merged as `ea77cde` |
| Grove unattended-propose specification | [#164](https://github.com/mascah/grove/pull/164), merged as `df86b0b`; merging approved future work, not implementation |
| Implementation reviewed | `propose-unattended` at `70875ec`; `just check` passed on an isolated snapshot after allowing the loopback server test |
| Later implementation handoff | `560e907`, inspected while documenting this plan; records partial trials and T-03 still in progress |

The original prototype's author reported a successful implementation/review loop,
a planted spec violation caught despite passing modified tests, a repair and a
passing review of the new commit. A usage-limit stop was detected by the missing
verdict. This did not exercise a remote, PR, second person or agent killed mid-turn.
The renamed Tend script had not been run end to end at the reviewed revision.

The subsequent review reproduced these failures with disposable repositories.
Agent calls were stubbed; the remote-claim probe used real Grove and a local bare
Git remote. No live agents or external publications were involved.

| Finding | Reproduction and consequence | Owner |
| --- | --- | --- |
| Remote record lookup | `grove claim` succeeds; `git log -1 grove/runs` fails because only `origin/grove/runs` exists. Tend stops after claiming. | Tend, slice 1 |
| Cached verdict bypass | A verdict file naming the wrong head, or missing its conclusion, is accepted on restart with exit 0. | Tend, slice 1 |
| Interrupted repair | `.tend/repaired` is written before prompt delivery; after delivery fails, a rerun skips repair and returns the failed verdict. | Tend, slice 1 |
| Fixture-only excludes | A valid first review in a repository without the fixture's `.tend/` exclusion stops as a changed checkout. | Tend, slice 1 |
| Quoted schema arguments | `grove instructions propose "--unattended --schema bugfix …"` selects default; separate arguments select bugfix. Generated skills use the quoted form. | Grove #164 implementation |

Source inspection also found a hardcoded `main`, no controller lock, and no
reconciliation of a surviving agent before prompting on restart. These were
not live concurrency/termination experiments. The review's temporary fixtures
are not required for future execution: reproduce the scenarios in durable tests.

The later Grove handoff records one case-(f) run per harness: Codex passed and
Claude failed two open-question checks. Both agents also interpreted the
instruction about questions requiring resolution before review too broadly.
The full trial suite was deferred at the owner's request to conserve usage.
Consult the current Grove implementation record for the authoritative results;
do not mark #164 complete from this plan or run the full suite automatically.

## Delivery

Keep the following order for Tend. Grove work can finish independently. Status
here describes delivery, not how strongly the architectural direction is agreed.

| Slice | Status | Prerequisites | Acceptance |
| --- | --- | --- | --- |
| G — Finish Grove #164 | Implementation merged as #165 (`0083bba`); acceptance per Grove's record | Existing Grove implementation and owner coordination | Invocation fix, outstanding behavioral findings and Grove's own acceptance |
| 1 — Recoverable execution | Done on `tend/slice1-claim-verdict`, live interruptions demonstrated; not merged | Existing Tend prototype | [Recovery](#recoverable-execution), [review](#review-integrity), [Grove integration](#grove-integration-and-repository-configuration) |
| 2 — One real shared-runner workflow | Done on `tend/slice1-claim-verdict`, live on sandbox PRs #1–#3; not merged | Slice 1 | [Publication](#shared-review-and-publication) |
| 2a — Act on requested changes | Done on `tend/slice1-claim-verdict`, live on sandbox PR #3; not merged | Slice 2 | [Below](#2a--act-on-requested-changes) |
| 3 — Explicit request-to-proposal intake | Done on `tend/slice3-intake`, live on the sandbox; not merged | Slice 2 and G available | [Intake and handoffs](#request-intake-and-explicit-handoffs) |
| 4 — Prove portability | Done and merged into `main`: non-Grove work item, Go port, the rex terminal and Jira as the second tracker; tmux dropped | Earlier lifecycle demonstrated | [Portability](#portability) |

### G — Finish Grove #164 independently

Keep the current scope. Pass on the quoted-schema reproduction, address the
reported behavioral findings, and finish or explicitly revise its outstanding
verification acceptance with the owner. The two original open questions have
reasonable recommended answers; they are not reasons to reopen the design.
Do not add scheduling or tracker behavior to this Grove change.

### 1 — Recoverable execution

Work primarily in `bin/tend`, supporting runtime/tests as needed, and the README.
Begin with regressions for the observed failures. Implement explicit run state,
single-controller ownership, status and reconciliation before resume. Separate
Grove, harness and Herdr operations only as far as this behavior requires.

Account for existing prototype artifacts before changing their location or
meaning. Demonstrate failures and recovery first with disposable Git repositories
and controlled process/agent substitutes, then exercise selected real interruption
boundaries. Record what remains untested; stub success is not live validation.

### 2 — One real shared-runner workflow

Start from an approved cycle rather than inventing intake simultaneously. Add
publication for one remote repository and show implementation, independent review,
bounded repair, PR publication and a second person's review. Exercise interruption
between remote side effects and local acknowledgment to prove deduplication.

### 2a — Act on requested changes

A teammate who requests changes on Tend's PR should not have to start Tend by
hand. On a run, Tend reads review comments on its PR from people other than the
publishing account since the last reviewed commit, gives them to the implementer's
own local session as a repair, has Codex review the new commit, and publishes it.
Each request gets one repair round; anything unclear or outside scope stops for
a person. Then a small watch loop on the runner runs this for Tend's own open PRs
every few minutes. This is narrower than intake polling, which stays deferred.

Claude Code's `/autofix-pr` was considered and not chosen: it is a cloud session
driven by GitHub webhooks and the Claude GitHub App, so it would fix the branch
outside Tend's lock, records, review and repair bound, as another identity.

### 3 — Explicit request-to-proposal intake

Settled with the owner on 2026-11-15:

- First tracker: GitHub Issues.
- Cap: at most 5 unmerged proposal PRs per repository. Proposals need a person's
  revision and approval; a longer queue is not reviewed, only accumulated.
- Request to repository: an issue's own repository receives its proposal, and a
  runner takes requests only from repositories it lists for publication. A tracker
  that spans repositories (Jira, Linear) will need an explicit mapping when added.

Select the first tracker before implementation; GitHub is the practical candidate
given the current repositories, but the provider is not a prerequisite for slice 1.
Add intake identity, cap, outcomes and links. Integrate unattended drafting after
G is available. Keep manual proposal revision and cycle planning visible. Only
then add polling that advances this established workflow.

### 4 — Prove portability

Use real second callers to guide extraction: a non-Grove work item, tmux, and a
second tracker provider. Preserve the same state and review contract. Broader
tracker coverage and multi-repository polling follow demonstrated integration
needs; they are not implied completed by adding adapter names or configuration.

Settled with the owner on 2026-11-15: tmux is dropped for now (rex and Herdr both
report agent state; under tmux tend would infer it from the screen). The second
tracker is Jira, as a source of requests and work items; code, pull requests,
reviews and publication stay on GitHub, unchanged.

#### 4b — Jira as the second tracker (design, before implementation)

- Mapping. A Jira project spans repositories, so `TEND_STATE/jira` lists
  `<site> <project-key> <origin-url>` a line, such as
  `https://mikeascah03.atlassian.net SCRUM git@github.com:mascah/tend-sandbox.git`.
  A project is taken only for an origin also listed in `TEND_STATE/publish`. One
  project maps to one repository; a second line for the same project stops tend.
- Trust. In Jira any user with Edit Issues permission can add `tend` or
  `tend-build` and change an issue's text, and anyone with Create Issues can file
  an issue already carrying them when Labels is on the create screen, unlike
  GitHub, where labelling needs triage access. Mapping a project is therefore a
  statement that everyone who can create or edit its issues may start agents;
  check both before mapping one. True of SCRUM, whose only member is the owner.
- Credentials. `JIRA_EMAIL` and `JIRA_API_TOKEN` from the environment, as Basic
  authentication over HTTPS. tend calls Jira's REST API v3 with `curl`, the token
  given on curl's standard input (`-K -`), never in its arguments, so the
  regression suite stands in for Jira on `PATH` as it does for gh, Git, Herdr and
  rex.
- Requests (`tend propose <repo>`). Besides the repository's GitHub issues, each
  mapped project's issues labelled `tend` whose status category is not Done, oldest
  first (`/rest/api/3/search/jql`). Identity is the issue key: branch
  `tend/jira-<KEY>` (such as `tend/jira-SCRUM-5`); a pull request from it in any
  state means handled. Handled requests keep their label, so the search is followed
  page by page (`nextPageToken`) past them until the cap is reached or results
  end. The shared `CAP` (default 5) counts GitHub and Jira proposals together. The
  drafter is given the summary, the description as text and the issue's browse
  URL; the commit ends `Refs <KEY>`. The description (Atlassian Document Format)
  becomes text that keeps what carries scope: a line per paragraph, heading and
  list item, code blocks verbatim, and every link's destination and smart link's
  URL. Once the pull request exists, tend comments its URL on the Jira issue, once,
  found by that URL in the issue's comments, read page by page (`startAt`). A
  request no proposal can be drafted from gets a Jira comment saying why, with a
  marker, and loses its `tend` label; the run folder is set aside as for GitHub.
- Finishing a handoff. A run folder records `linked` once the Jira comment, and the
  proposal's summary review, are posted. A pull request means "do not draft
  again", not "finished": each `tend propose` completes every run folder with a
  pull request and no `linked`, whatever the cap, the label or the issue's status.
  This closes the same gap in GitHub proposals, whose summary review is lost today
  if tend stops between opening the pull request and posting it.
- Identity is the key. A project-key rename changes every key, so its open
  requests would be drafted again; accepted, as renames are rare and readable
  branches are worth more.
- Work items (`tend <repo> <KEY>`, such as `SCRUM-5`), for repositories without
  Grove, as `issue-<n>`: approved by the `tend-build` label, done when the status
  category is Done, scope from summary and description. The project must map to
  the repository's origin. Branch `tend/build-<KEY>`, attempt
  `jira-<KEY>-<cksum of repo path>`, commit message ending `Refs <KEY>` (GitHub's
  `Closes` does not reach Jira). After publication, the pull request's URL is
  commented on the issue once. Requested changes, watch, recovery and status work
  as for any attempt.
- Not in this unit: moving Jira issues through their workflow (people do; tend only
  comments and removes its label), Jira Data Center, webhooks, and a Jira account
  for tend apart from the owner's (comments are the owner's until one exists).
- Evidence: a Jira stand-in in `dev/test.sh` (through `curl`), covering search and
  comment pages, descriptions with links, smart links, lists and code, and a
  handoff finished after an interruption; then one live proposal from project
  SCRUM on `mascah/tend-sandbox`. Proposals need Grove and direct builds refuse it,
  so Jira work items are shown by regressions only, as the owner settled for
  `issue-<n>` (no live run in a repository without Grove).

## Deferred scope and unresolved choices

Deferred: distributed runners, multiple concurrent implementations, agent
supervisor hierarchies, a new issue tracker/memory product, generic workflow/plugin
languages, automatic merging, unattended proposal revision, automatic cycle
planning, and phone-notification guarantees. Notifications are not the sole status
surface. No agent should propose arbitrary new work merely because it can.

Automatic proposals from review findings and `grove check` diagnostics wait until
intake distinguishes new requests, existing proposals, rejected suggestions and
diagnostics that need correction rather than new product scope. A cap alone does
not prevent repeated or unwanted work.

Language, settled 2026-11-15: Go. Bash carried slices 1–3 but showed its limits
(global variables silently overwritten, fragile quoting and JSON handling, adapters
becoming case statements), and several terminals (Herdr, rex, tmux) and trackers
are coming. Go gives one static binary for a runner, plain subprocess and JSON
work, and small interfaces for adapters. `dev/test.sh` drives Tend only through
its command line with stand-ins on `PATH`, so it is the acceptance suite for the
port: the Go binary must pass it unchanged before it replaces the script. Other
routine choices (state encoding, command spelling, module layout) remain open. Before slice 3, settle
the first tracker, cap and mapping of requests to repositories. Before enabling
automatic publication on a runner, establish its configured repositories and
publication authority. These later decisions do not block slice 1.

## Session handoff

2026-11-14 — Plan recorded after the owner accepted the review direction.
Tend remains at `60e9f4a`; no runtime fix from this plan is implemented. Grove's
implementation advanced from the reviewed `70875ec` to the progress-record commit
`560e907`; its full model trials remain deferred. This documentation session ran
no agents, made no tracker writes and changed no Grove files.

Next action: implement Tend slice 1, starting with durable regressions for remote
claim discovery and cached verdict validation, then resolve the interrupted-run
lifecycle. Keep G separate and record its final merged revision when available.

Context audit after recording the plan: added explicit distinctions between
initial delivery and the future product, terminology, and ownership of cycle
subagent scheduling. These clarify the agreed direction; no additional slice
was implemented or made ready by this documentation update.

2026-11-14 — Slice 1, first unit: remote claim discovery, verdict validation
and the configured target branch. Code written and local regressions passed;
committed on branch `tend/slice1-claim-verdict`, one commit on `60e9f4a`
together with this plan and the earlier README edits; not merged. Changed `bin/tend` and `README.md`; added
`dev/test.sh`.

- Remote claim: the attempt is now read from `grove changes --json` `.runs`
  (Grove's published data: `origin/grove/runs`, or local `grove/runs` without
  `origin`). After a claim, the record is the one whose branch is the checkout's.
  The lookup failure was reproduced first with real Grove 0.10.0 and a bare remote.
- Verdicts: cached and new verdicts get the same check, which needs a head that
  matches, a `success`/`failure` conclusion, a boolean `actionable` and a string `summary`.
  An invalid cached file is renamed `.rejected-<time>` (kept) and the head is reviewed
  again. An invalid new one stops for a person. A suitable checkout is the existing
  clean-and-archived check before review, plus the unchanged-head check after it.
- Target: the worktree base and the reviewer's diff use Grove's `target`, not `main`.
- Checks: `dev/test.sh` passes all 19 checks: remote claim, remote and local resume
  after an interruption past the claim (same attempt, one record), cached
  wrong-head and incomplete verdicts, an incomplete new verdict, and target `trunk`.
  Before the fix, 14 checks failed and the no-remote control passed. Bash syntax was checked;
  shellcheck is not installed.
- Not exercised: live Herdr, Claude or Codex, and the ledger fixture end to end. All agents were
  stubbed and no remote beyond a local bare repository was used. The tests still add
  `.tend/` to `.git/info/exclude` (the fixture-only-exclude finding remains open).
- Remaining slice 1 failures: interrupted repair (`.tend/repaired` before
  delivery), runtime artifacts inside the candidate, no controller lock, no
  reconciliation of a surviving agent (including a leftover reviewer name after a
  stop), and status/cancel. None of the interruption demonstrations are done.
- Grove: not touched. `grove-propose-unattended` is now at `413e0fe` (was
  `560e907` in this plan); G status is unchanged here.

Next action: move the attempt's runtime record out of the candidate and take a
controller lock, reconciling legacy `.tend/` verdicts and `repaired` without
discarding them. Then use that record to make repair delivery and surviving-agent
reconciliation explicit, with the interruption demonstrations. Merging the branch
is the owner's decision.

2026-11-14 — Slice 1, second unit: runtime files outside the candidate and a
controller lock. Code written and local regressions passed; committed on
`tend/slice1-claim-verdict` as `fc8b3e9`, with a test guard in the next commit; not merged.

- Run files: `TEND_STATE` (default `~/.local/state/tend`) holds `lock` and
  `runs/<attempt>/` with verdicts, `.rejected-*` and `repaired`. Not under `.git`:
  Codex's `workspace-write` sandbox refuses writes inside `.git` and outside the
  workspace (probed with `codex sandbox`, no model), so the reviewer starts with
  `--add-dir <run folder>`. The repair prompt now carries the findings text
  rather than a path, so the implementer reads nothing outside its worktree.
- Legacy `.tend/`: on carrying on an attempt, its files move to the run folder.
  A name already present there is not overwritten; tend stops for a person. A
  prior verdict is reused and a consumed repair stays consumed.
  The fixture no longer excludes `.tend/`, and the tests use no Git excludes.
- Lock: one tend per runner (`ponytail:` note: per-attempt locks when capacity
  grows). A forked perl watcher holds a `flock` on a descriptor nothing else has
  and exits when tend's pid is gone, so children such as a `herdr agent prompt
  --wait` cannot keep it after a kill. (A descriptor held by tend itself was
  tried and failed: command-substitution subshells inherit it.) A refused run
  exits 4 and touches nothing; holder pid, cycle and repo are in the lock file.
- Checks: `dev/test.sh` passes all 31 checks. New this unit: a clean candidate with no excludes,
  the verdict outside the candidate, `--add-dir`, legacy reuse, a refused second
  tend that leaves the implementer prompted once, and takeover after `kill -9` of
  the holder. Before the change, 18 checks failed. The unreadable-lock path was checked by hand.
- Codex review of the branch (2 rounds, `NO_OPEN_ISSUES`): fixed a relative
  `TEND_STATE` that put verdicts in the candidate, and a regression that read
  Grove's root from the caller's checkout instead of the attempt's. Both have
  regressions; `dev/test.sh` now passes all 35 checks.
- Not exercised: live Codex writing through `--add-dir` (only the sandbox
  policy was probed), live Herdr/Claude, a legacy name clash, Linux.
- Remaining slice 1: repair still marked before delivery; after takeover, the
  new controller re-prompts an agent that may still be running (the lock test
  shows this: the stub implementer was prompted again); status/cancel; live
  interruption demonstrations during implementation, review and repair.

Next action: record repair delivery so an undelivered or uncertain repair is
visible rather than skipped or retried without bound. Before any prompt, establish
whether the named agent is still working, and observe or ask a person instead of
re-prompting. Add a minimal status that shows stage and next action from the run
folder. Then run the live interruption demonstrations.

2026-11-14 — Slice 1, third unit: repair delivery, surviving agents and status.
Code written and local regressions passed; committed on `tend/slice1-claim-verdict`
after `03d7ddf`; not merged.

- Surviving agent: before any prompt, tend reads the agent's Herdr status. If it is
  `working`, tend waits for it (`agent wait`) and does not prompt it again. If it is
  `blocked` or `unknown`, tend stops for a person. A reviewer left by an earlier run
  is reused instead of started again under its live name. No prompt is ever resent
  after a timeout or stall; Herdr's own guidance says those do not prove non-delivery.
- Repair: `repair-asked` (the reviewed head, written atomically just before the
  prompt is sent) consumes the one repair. If that head has no commit since and the
  implementer is not working, the outcome is unknown: tend stops on every run until a
  person commits a fix or deletes `repair-asked`. A prompt refused before sending
  (blocked implementer) does not consume it. A legacy `repaired` is read as asked.
- Status: tend writes `runs/<attempt>/status` atomically (time, kind
  `working`/`needs-you`/`done`, what, next action). `tend status <repo> <cycle>`
  shows it, takes no lock, and says whether a tend is on the attempt now. If none
  is while the kind is `working`, it says to run tend again.
- Cancel: stopping tend leaves agents and work intact; resume is running tend again.
  No separate cancel command; stopping an agent is done in its pane. (README.)
- Checks: `dev/test.sh` passes all 63 checks. New this unit: a normal repair; a stalled
  repair prompt neither skipped nor resent, and asked again only after a person
  deletes the marker; a blocked implementer not consuming the repair; surviving
  implementer, reviewer and repairer waited for and not re-prompted; status while
  running, after a stop and after `kill -9`. Against the previous `bin/tend`, 19
  of the new checks fail, including the reproduced bug where a rerun after an undelivered
  repair exited 0 with the failed verdict. The stand-in Herdr now refuses prompts
  to a blocked agent, as Herdr documents.
- Not exercised: any of this with live Herdr, Claude or Codex; Herdr's real
  `agent get`/`agent wait` output was checked by shape only, read-only against
  an idle agent of the owner's.

Next action: slice 1's remaining acceptance is the live demonstrations. With the
owner's agreement to spend agent usage, run the ledger fixture through Herdr and
interrupt it during implementation, review and repair. Include `kill -9` of tend
while an agent keeps working, with a planted defect for the repair. Record what
each showed and what stays untested. Then slice 1 can be marked done and slice 2
started.

2026-11-15 — Slice 1, live interruptions. With the owner's go-ahead, ran the ledger
fixture (`dev/fixture.sh`) through `bin/tend` at `03f2a2a` with Herdr 0.9.5, Claude
Code 2.1.325, codex-cli 0.169.0 and Grove `0083bba`; `TEND_STATE` in a scratch folder.
One attempt, `261115-8t6j4`, carried through eight runs, all on the same branch and run folder:

- Implementation: `kill -9` of tend while Claude worked. Claude carried on; `tend
  status` said to run again; the rerun waited ("still working ... not prompting
  again") and the transcript holds `/grove-implement` once. Candidate `692d2d5`.
- Review: `kill -9` of tend while Codex reviewed. The rerun reused and waited for
  `rev-692d2d5`; Codex wrote a valid `success` verdict through `--add-dir` (live now).
- Repair: a defect planted by hand as `f63d3fa` (the message reworded, its test
  changed to match, tests passing). Codex failed it as actionable, citing the
  spec's exact message. `kill -9` of the Claude process mid-repair: tend stopped
  (no repair commit), and the next run stopped again as an uncertain repair rather
  than asking again. After `repair-asked` was deleted by hand, a new Claude session
  took the name and was asked once; `kill -9` of tend while it worked; the rerun
  waited, Claude committed `f9996fb`, and Codex passed it. The transcript holds
  the repair prompt once.
- The candidate stayed clean throughout (no `.tend/`, nothing but `__pycache__`
  ignored); the run folder holds three verdicts, one per head.

Found, not fixed:

- Codex opened on a "hook needs review" screen (a hook the owner had just added).
  Herdr reported it `idle`, not `blocked`, so tend prompted into it; the
  missing verdict stopped tend (exit 3) without a false result. The owner trusted
  the hook and the rerun reused that reviewer. Owner's view: rare, so tend does not
  read the screen for unknown dialogs.
- `kill -9` of tend leaves its `herdr agent prompt --wait` child running until the
  agent settles; it then exits by itself and holds no lock.

Not exercised live: a remote (origin claim was shown with a bare repo only), a
blocked agent, an agent killed mid-implementation or mid-review, Linux, tmux.
`dev/test.sh` still passes 63 checks; no code changed this session.

Next action: slice 2. Settle with the owner the remote repository and publication
authority, then add one PR publication with deduplication across interruptions.

2026-11-15 — Slice 2, publication. The owner chose a small private throwaway
repository and authorized pushing branches, opening and updating PRs and
commenting/statuses there, never merging; the owner plays the teammate. Code
written and local regressions passed; exercised live; committed on `tend/slice1-claim-verdict` as `5e43af9`;
not merged.

- Publication (end of `bin/tend`): only when `origin`'s URL is a line of
  `TEND_STATE/publish` (a side review pointed out that without this, any repository
  with a remote would be published to). Push `HEAD:refs/heads/<branch>`, never
  forced; a rejected push stops for a person. One PR per branch, looked up on GitHub
  each run (`gh pr list --head`, all states), so there is no local acknowledgment to
  lose; created onto Grove's target with the cycle's title from `grove changes
  --json`, else its body edited. The verdict goes on the reviewed head as the
  `tend/review` commit status, so a later commit has no passing result. The body
  holds the verdict, earlier verdicts of the attempt, a note pointing to each
  `implementation.md` for open-question choices, and `grove review`'s summary.
- Checks: `dev/test.sh` passes 92 (a `gh` stand-in added). New: an unlisted origin
  gets no push or PR; push, one PR onto the target, title, body, status; a repeated
  run edits, not creates; a PR created but not acknowledged (stand-in fails after
  creating) is found by the rerun; a later commit's failure goes on the new head with
  the earlier pass kept as history; a diverged remote branch is not overwritten.
  Against the previous `bin/tend` 22 of them fail.
- Live (`mascah/tend-sandbox`, private, from `dev/fixture.sh`; same versions as
  slice 1's runs): attempt `261115-n5xeb` implemented and passed at `c0c26aa`. A
  wrapper around the real `gh` paused after `gh pr create` succeeded; tend was
  killed there, with PR #1 open and no status. The rerun found PR #1, opened none,
  and set the status. Then a planted defect (`8991d29`, message and test reworded)
  failed review, was repaired (`3d107aa`) and passed; PR #1 was updated to lead with
  that pass and list both earlier reviews. Statuses: `c0c26aa` and `3d107aa`
  success; `8991d29` none, since tend publishes at the end of a run.
- Not exercised: the second person's review and merge through the PR; a remote
  branch someone else pushed to (stand-in only); `gh` failures other than a lost
  create; a failed final verdict on GitHub (stand-in only).
- Noticed: `git push -q` still prints GitHub's "create a pull request" hint to
  tend's stderr. A failed head mid-run never gets a status (only the final head is
  published); the PR body is where earlier failures show.

Next action: the owner reviews PR #1 in `mascah/tend-sandbox` as the teammate and
merges or requests changes through it. Record what they could and could not
judge from the PR alone, then mark slice 2 done or add what was missing.

2026-11-15 — Slice 2, publication as a separate account. The owner found the
verdict-in-description clunky (it replaced the PR's intent) and asked that a
separate account, `mascah-bot`, publish rather than the owner. Committed on
`tend/slice1-claim-verdict` as `518c34e` plus the docs commit after it; not merged.

- `TEND_STATE/publish` lines are `<origin-url> <account>`. Tend takes the account's
  token from `gh`'s keyring (`gh auth token --user`) for its own calls; the owner's
  account stays active. Pushes go over HTTPS (`pushInsteadOf` for an SSH origin,
  `gh auth git-credential`). The description is written once, at creation, from
  Grove's JSON and `grove review`; each verdict of the attempt is a COMMENT review
  on its commit (a PR's author may not approve or request changes), found by
  commit and account so it is posted once; each reviewed commit gets a status.
- Owner decisions: comment reviews for now; the bot pushes too; commits keep the
  implementer's local Git identity. `mascah-bot` was invited with write access and
  the invitation accepted with its token (an earlier note here wrongly said it
  already had access).
- Checks: `dev/test.sh` passes 101. The `gh` stand-in now records which account's
  token made each write and applies the caller's `--jq` to stored reviews. New:
  PR authored by the account; description has no verdict and survives a person's
  edit; one review per head, also after a lost create; a repaired failure's review
  on its own commit; each head's status.
- Live: PR #1 (the earlier format, authored by the owner) was merged by the owner,
  which unpaused Grove. A second sandbox cycle, `261115-pr0t1` (`list` says `no
  records`), was added and re-planned by hand after that merge made its plan stale.
  Attempt `261115-fktq6` passed review at `bfff106` and opened PR #2: author,
  branch push, review and status all `mascah-bot`. A rerun posted no second review
  and left the description unchanged.
- Not exercised live: a failed or repaired head through the bot (stand-in only),
  a teammate's approval of a bot PR, requested changes (slice 2a).

Next action: the owner approves and merges PR #2 as the teammate, or requests
changes (which Tend will not act on until slice 2a). Record what could and could
not be judged from the PR alone, mark slice 2 done, then start slice 2a.

2026-11-15 — PR description as the squash commit, and slice 2a. The owner squash-
merges, so the PR title and description become the commit; they found the
Grove-index description unhelpful for that (tool state, not what changed and
why; unclean in history). Committed on `tend/slice1-claim-verdict`: `7964c29`
(commit message), `150e694` (2a), and the docs commit after; not merged.

- Commit message: before opening the PR, tend asks the implementer's session
  (now started with `--add-dir <run folder>`) to write `message-<head>` there: a
  subject in the repository's commit style, a body on what the merged change
  does and why, no tool, review or attempt talk, ending `Grove: <change IDs>`.
  It becomes the PR title and description, set once. `grove review`'s summary
  moved into the bot's review of the head, collapsed.
- 2a: after publishing, a `CHANGES_REQUESTED` review on the open PR's head is
  given, with its line comments, to the implementer once (`request-<review id>`
  written just before, as `repair-asked` is); the new head is reviewed by Codex
  and published, the implementer's reply (`reply-<head>`) quoted in that review.
  Asked with no commit and nobody working, it stops for a person. A request on an
  older head counts as answered. No further automatic repair follows a requested
  change; a failed Codex review of it is published for a person.
- `tend watch` (`WATCH_SECONDS`, default 300): for each run folder with `pr` and
  `attempt` files whose status is not `needs-you`, runs tend when `gh pr view`
  shows the PR open with a change request on its head. Watch holds no lock; the
  tend it runs does. Fixed on the way: the publish loop reused `$state`.
- Checks: `dev/test.sh` passes 124. New: message asked once, PR title and body
  from it, `--add-dir`; a requested change asked once with its body and line
  comment, the new head reviewed and published with the reply, not asked again;
  a stalled request neither repeated nor skipped until its mark is deleted; none
  on a merged PR; watch idle without a request, running tend on one, skipping an
  attempt that needs a person. Against `7964c29`, 14 of the new checks fail.
- Not exercised live: any of this commit's behaviour. Known limits: a requested
  change does not rewrite the PR description, so a scope-changing request leaves
  it stale; when the implementer commits nothing (for example a request it judges
  out of scope) tend stops for the runner's person, and nothing is said on the PR.

Next action: the owner merges PR #2. Then plan a third sandbox cycle, run tend
on it with `tend watch` in its own pane, and have the owner request changes on
the bot's PR; record the commit message, the requested change and its reply.

2026-11-15 — Slices 2 and 2a live. The owner merged PR #2 (approved as the
teammate), which closed cycle `261115-pr0t1`. A third cycle was planned in the
sandbox; a first version (refuse a negative amount) was wrong, as `validate`
already refused it, and was withdrawn by a new commit; `261115-pr0t3` (refuse an
invalid date on `add`) replaced it, after checking the gap was real.

- Attempt `261115-28vwx`: implemented, Codex passed `a271ab2`, and `mascah-bot`
  opened PR #3. Its title and description were the implementer's commit message:
  subject "Refuse an invalid date when adding a record", three short paragraphs
  (the problem, what `add` now does, what stays out of scope), `Grove: 261115-e0002`;
  no tool or review talk. The head's review carried `grove review` collapsed.
- `tend watch` (`WATCH_SECONDS=60`) ran in its own Herdr pane. The owner used
  "Request changes" on `a271ab2` (format in the error message). Watch ran tend,
  which wrote `request-5912015116` and asked Claude's existing session once; Claude
  committed `09b9c1a` and wrote its reply. The watch was then stopped by hand
  (Ctrl-C, killing that tend before the review) after a mistaken diagnosis that it
  had not fired: it had, its sleep just had not ended. A one-pass watch then carried
  on without asking again (the request is on the old head), Codex passed
  `09b9c1a`, and `mascah-bot` posted that review with Claude's reply quoted (what
  it changed, why it left the spec) and the `success` status.
- Slice 2's acceptance, all live on the sandbox: implementation, independent
  review, bounded repair (PR #1's planted defect), one PR with deduplication after
  an interruption between the remote side effect and the local acknowledgment
  (PR #1), and a second person reviewing and acting through the PR (approval and
  merge of #2, requested changes on #3). Merging stayed the owner's.
- Still not live: a failed final verdict through the bot; a branch someone else
  pushed to; a teammate's request the implementer declines (no commit: tend stops
  for the runner's person, and nothing is said on the PR); a request that changes
  scope (the description is not rewritten). `tend watch` was left stopped.

Next action: settle slice 3's open decisions with the owner before implementing
it. Merging `tend/slice1-claim-verdict` is the owner's decision.

2026-11-15 — Slice 3, first unit: `tend propose <repo>`. `tend/slice1-claim-verdict`
was fast-forwarded into `main` and pushed (`34e5c49`) at the owner's word, keeping
the per-unit commits this handoff cites; sandbox PR #3 was squash-merged
(`7e43590`) after setting the sandbox's squash default to the PR title and
description. Code written, local regressions passed, live; committed as `6182dca`
on `tend/slice3-intake` plus the docs commit after; not merged.

- Intake: open issues labelled `tend` in a listed repository, oldest first. The
  issue number is the identity: branch `tend/issue-<n>`; a PR from it in any state
  means handled. Cap: `CAP`, default 5, open proposal PRs by the account. Claude
  drafts with `/grove-propose --unattended <title. body (url)>` (Grove's quoted
  argument form works at `0083bba`); a second prompt has it commit the draft with a
  squash-ready message ending `Refs #<n>` and write a summary of open questions,
  consequential first, or write why there is none. Both prompts are marked before
  they go (`asked`, `followed` under `TEND_STATE/intake/<owner>_<repo>-<n>/`); a
  drafter still working is waited for. Drafted: pushed and opened by the account,
  its review the summary plus Grove's consequence (merging approves with questions
  open; implementation may then choose) and how to revise or plan. Declined: a
  comment carrying the attempt's marker (so a rerun does not repeat it), then the
  label removed and the run folder set aside, so a relabel starts afresh.
- Checks: `dev/test.sh` passes 162; 36 of the new ones fail against `34e5c49`. New:
  drafting from labelled issues only, PR by the account from the issue's branch,
  title and body from the drafter's commit, the change on the branch, the review's
  questions and consequence, no repeat on a second pass; a declined request
  commented once and unlabelled, relabelling drafting afresh; a decline whose label
  removal failed finished without a second comment; the cap, and a merged proposal
  freeing a place; a stalled follow-up neither resent nor skipped until its mark is
  deleted; a drafter that outlived tend waited for; an unlisted repository refused.
- Live on the sandbox: issue #4 (`edit --date` should refuse an invalid date) became
  PR #6 by `mascah-bot`, title "Propose refusing an invalid date when editing a
  record", body a commit message ending `Grove: 261115-2az0q` / `Refs #4`; its review
  listed three consequential and two routine questions with recommendations. The
  drafter noticed the issue's example was already refused for another reason and
  used another record. Issue #5 ("Make the ledger nicer") got a comment saying why
  no proposal could be drafted, with concrete suggestions, and lost its label. A
  second pass changed nothing.
- Not done: tend does not fetch the target before branching (nor does an attempt);
  a declined request's worktree is left; a closed, unmerged proposal is never
  redrafted; comments on a proposal PR are not acted on (revision stays manual,
  as the plan says). Slice 3 also asks that a person can see which proposals need
  answers or revision, which approved changes need a cycle, and which
  implementations need review: not yet built beyond the PRs themselves. Polling
  intake comes after that.

Next action: a read-only overview of those handoffs for a repository (open
proposal PRs and their questions, approved changes without a cycle, cycles ready
to claim, execution PRs awaiting review), from Grove's JSON and GitHub; then let
`tend watch` run intake. PR #6 is the owner's to review, revise or merge.

2026-11-15 — Slice 3, second unit: the handoff overview and polling intake. The
owner merged proposal PR #6, approving change `261115-2az0q`. Committed as
`a5536bc` on `tend/slice3-intake` plus the docs commit after; not merged.

- `tend status <repo>` (no cycle; reads only): approved changes without a cycle,
  cycles not done with Grove's verdict, this runner's attempts for those cycles,
  open proposal PRs (`tend/issue-*`) and implementation PRs (`grove/*`, marked
  when changes are requested), and labelled issues no PR, in any state, came from.
- `tend watch [<repo>...]` also runs `tend propose` for each repository given,
  except one with an intake run folder waiting for a person.
- Checks: `dev/test.sh` passes 170. New: each overview section, a request drafted
  or merged left out, an attempt for a done cycle left out (Grove judges done from
  `origin`'s target); watch running intake and leaving a request that needs a
  person. The `gh` stand-in's `pr list` now builds every PR and narrows by head,
  author or state.
- Live: the sandbox overview showed exactly the approved `261115-2az0q` waiting for
  a cycle. Its first version listed issue #4 as waiting (only open PRs were
  checked) and a finished attempt; both fixed, with checks. A live watch pass ran
  intake with nothing to do.
- Slice 3's acceptance, as shown: labelled requests from one tracker, stable
  issue identity, links (`Refs #n`), no repeats, the cap, a visible outcome for an
  undraftable request, no edits to teammates' prose (a comment and a label only),
  Grove's unattended propose, the handoffs visible (proposal questions on its PR
  and in the overview, planning, review), and the consequence of merging with
  open questions, consequential and routine apart. Polling came last.
- Not done: see the first unit's list (no fetch before branching, leftover
  worktrees, closed proposals, manual revision). Cycle planning stays manual.

Next action: agree slice 4's order with the owner: a non-Grove work item, tmux
without Herdr, or a second tracker. Merging `tend/slice3-intake` is the owner's.

2026-11-15 — Slice 4, first unit: a GitHub issue as a work item. The owner had
`tend/slice3-intake` fast-forwarded into `main` and pushed (`9fe3d93`), asked for
the cycle of `261115-2az0q` to be planned (cycle `261115-qq9tq`, committed to the
sandbox's `main`; `grove claim` would claim it; not built), and agreed slice 4's
order. Committed as `7e3d8f6` on `tend/slice4-issue-items` plus the docs commit
after; not merged.

- `tend <repo> issue-<n>`: approval is the `tend-build` label (triage access to
  put on); eligible while open and labelled; closed is done (exit 0). Scope: the
  issue's title and body. Branch `tend/build-<n>`, found by its worktree; attempt
  id `issue-<n>-<cksum of the repo path>`. Built when the branch has a commit and
  the implementer wrote `implemented` (notes for the reviewer) in the run folder.
  The reviewer is given the issue and the notes instead of change folders; repair
  and requested-change prompts point at the notes; the commit message ends
  `Closes #<n>`; no `grove review`. Lock, surviving agents, the one repair,
  publication, requested changes, status and watch are shared. The second caller
  moved only this into per-kind code: the attempt, the completion check and the
  prompts' scope (`what`, `notes`, `trailer`).
- Checks: `dev/test.sh` passes 189, in a repository without Grove: build, review
  and publish; scope given to implementer and reviewer; `Closes #7`; no grove
  review; status and overview; no rerun repeats; an unlabelled issue refused and a
  closed one done; missing notes stopping; a surviving implementer waited for; a
  repair pointing at the notes. Fixed on the way: the requested-change loop reused
  `$id`, which now names the attempt.
- Live: sandbox issue #7 (`list` ends with a count) became PR #8 by `mascah-bot`,
  "End ledger list with a count of records", body ending `Closes #7`; Codex passed
  `72b84aa` against the issue and notes (14 tests, 0/1/6-record runs). The sandbox
  has Grove, so this did not prove a repository without it; the regressions do.
  The implementer's notes observed that a direct build leaves Grove's maintained
  spec silent on the new behavior.
- Open for the owner: whether `issue-<n>` builds should be refused in a repository
  that has Grove (the spec would otherwise drift), and whether a live run in a
  repository without Grove is wanted (a second throwaway repository).

Next action: tmux without Herdr. Herdr supplies agent state (working, idle,
blocked); tmux does not, so tend must establish it itself.

2026-11-15 — Owner decisions after the first slice 4 unit. Tend and Grove are meant
to be used together: in a repository with Grove, `issue-<n>` builds are now refused
with a pointer to the `tend` label (a proposal first), so maintained specs keep up;
direct builds are for repositories without Grove (regression added; 191 checks).
No live run in a Grove-free repository is wanted. Order from here: port to Go
behind `dev/test.sh`, then the terminal interface from Herdr and rex (or tmux), then
the other terminal, then a second tracker.

2026-11-15 — Go port. Tend is now `main.go` (Go 1.26, standard library only),
built with `go build -o bin/tend .`; `bin/tend` is build output and ignored, and
the Bash script is removed.
- Behavior: a line-for-line port. It still shells out to `herdr`, `gh`, `git` and
  `grove` with the same arguments and the same `--jq` filters, so the stand-ins
  and GitHub see identical calls. It no longer needs `jq` or `perl`. The runner
  lock is an `flock` on a descriptor no child inherits (Go opens files
  close-on-exec), replacing the forked perl watcher: it still goes when tend
  exits or is killed while an agent's wait outlives it. JSON is decoded as jq
  reads it (`null` interpolated as `null`, `-r` printing nothing for it), and
  prompts keep the script's exact wording and spacing.
- Checks: the script and the Go build each passed every check before the swap
  (`TEND=<build> dev/test.sh`). The suite changed in three ways only: `TEND`
  selects a build, by default it builds this checkout itself, and one new check
  pins an issue attempt's name to `cksum` of the repository path, so existing run
  folders keep matching. 192 pass.
- Live, read only: on the sandbox, with the real Grove, gh and run folders,
  `tend status` (overview, `issue-7` and a cycle) printed exactly what the script
  printed, and found `issue-7-2789477195`. No live attempt has run on the Go build
  yet; the next live run is its first.
- Found on the way: an empty repository argument was read as the current folder;
  it now fails as the script's `cd ""` did.

Next action: the terminal interface. Read what Herdr gives tend today (start,
get, wait, prompt, read, send-keys, panes and worktrees) as the interface, and
check whether rex can report agent state (`rex` skill); if not, tmux, where tend
must establish working, idle and blocked itself.

2026-11-15 — Terminal interface; rex. rex (0.1.0-preview.19, from the Ghostty
makers) reports agent state: Claude Code 2.1.325 and Codex 0.169.0 send the program
status protocol (OSC 7501), and `rex block call … program_status` returns it. Probed
live in a throwaway rex session: Claude reported idle after its trust dialog,
working then done for a turn, and blocked with `kind=question` and the question as
its message; during the trust dialog it reported nothing, and its record cleared
when it exited. Codex reported idle, working, done; at a dialog the owner approved
by hand it reported nothing. So tmux is not needed for agent state.
- `herdr.go`: a `terminal` interface (worktree, open, pane, start, read, keys,
  status, prompt, wait, closePane, notify) with Herdr behind it, every call as
  before; shared code keeps the agent rules (wait for working, stop on blocked or
  unknown, the trust dialog, marks).
- `rex.go`: a session per worktree, named after its folder; each agent a block named
  after it, running the agent as the block's program. The state is the root record;
  no record while the agent runs is `unknown`, so a dialog the agent does not report
  stops tend instead of being prompted into (Herdr reports such a dialog idle).
  Prompts go as a bracketed paste then Enter, delivered only if the record changes
  within five seconds, else `agent_prompt_stalled`. Git makes worktrees under
  `TEND_STATE/worktrees/`. tend reports its own needs-you and done to its tab with
  OSC 7501. Chosen by `TEND_TERMINAL`, else rex when `REX_SESSION` is set outside
  Herdr.
- Checks: 215 pass, 23 of them new rex scenarios (claim, build, review and rerun;
  repair; a stalled prompt then a working agent waited for and not prompted again;
  blocked and silent agents stopped before the repair; the trust dialog answered).
  The stand-ins share one file of agent turns.
- Not shown: a live attempt under rex; whether the first Enter after a paste is
  ever lost (seen once in the probe, likely the owner's dialog); `rex block close`
  syntax against the real CLI; rex is not on PATH by default (the CLI is inside
  `Rex Beta.app`).

Next action: a live attempt under rex on the sandbox, watched in the Rex app.

2026-11-15 — Live under rex. `TEND_TERMINAL=rex` on the sandbox took planned cycle
261115-qq9tq (refuse an invalid date on edit) through as attempt 261115-7gphq in
about 3 min 20 s: Claude built in its rex block, tend read `done` and found the
change archived, Codex passed `207105a` in a block of its own (15 tests, both
scenarios reproduced), the reviewer's block was closed (`rex block close` works),
and Claude wrote the message. PR #9 by `mascah-bot`, "Refuse an invalid date when
editing a record", ends `Grove: 261115-2az0q`, carries `tend/review: SUCCESS` and
one COMMENT review. No dialog stopped it and no Enter was lost.
- Found: Claude renamed the rex session (`grove-implement-edit-date`), as the rex
  skill tells agents to, so finding a session by name would open a second one on
  a rerun. tend now records each worktree's session in `TEND_STATE/rex-sessions`
  and reuses it while rex lists it, matching the folder by identity (`/var` and
  `/private/var` are one); a regression renames the session before a rerun. A live
  rerun of the attempt then reused the renamed session, asked nothing and posted
  no second review.

Next action: the remaining terminal, then a second tracker.

2026-11-15 — Jira as the second tracker (4b). The owner dropped tmux and chose Jira;
`tend/slice4-issue-items` was merged into `main` and pushed first. The design in 4b
was reviewed with Codex over three rounds: fixed the label trust boundary (create
and edit), unfinished handoffs, a live proof both repository states cannot give,
search and comment paging, links lost from descriptions, and two wording points;
key renames accepted. Implemented in `jira.go`: a `request` (GitHub or Jira issue)
that intake and work items share, the GitHub calls unchanged; Jira through `curl`
with the token on standard input; descriptions as text keeping link targets and
smart-link URLs; `linked` marks a finished handoff and every pass completes folders
with a `pr` and no `linked`; a proposal is pushed only before its pull request
exists, so a merged and deleted branch is never pushed again.
- Checks: 250 pass, 35 new for Jira (drafting past the first search page; done and
  unlabelled left alone; description links, smart links, lists and code reaching
  the drafter; `Refs <KEY>`; the comment once, though on a later comment page; the
  token never in curl's arguments; the overview; a failed comment finished on the
  next pass with the label gone; a decline; a build with its comment and attempt
  name; unlabelled, done, unmapped and wrongly mapped issues refused).
- Live: project SCRUM on mikeascah03.atlassian.net mapped to `mascah/tend-sandbox`,
  agents in rex. SCRUM-5 ("ledger list should filter by category", with a link to
  the record-listing spec) became PR #10 by `mascah-bot`, "Propose filtering ledger
  list by category", body ending `Refs SCRUM-5`, a review with six open questions
  (consequential first), and one comment on SCRUM-5 naming the PR. A second pass
  repeated nothing. Credentials are the owner's own Jira account, in the repository's
  ignored `.envrc`, so the Jira comment is the owner's until tend has an account.

Next action: agree the next unit with the owner. Candidates from the record: the
known gaps (no fetch of the target before branching, leftover worktrees, a declined
change request silent on the PR, a description stale after a scope-changing
request), a Jira account for tend, and multi-repository polling.

2026-11-15 — The known gaps. The owner merged `tend/slice4-jira` into `main`
(`17048e0`, pushed) and asked for the recorded gaps to be closed; on
`tend/known-gaps`, not merged.
- Fetch: new work branches from `origin/<target>`, fetched first (a fetch that fails
  with an origin stops), and proposals and issue builds count their commits from
  it, as the reviewer's diff does; without an origin, the local target as before.
  Grove's claim already branched from origin for cycles; proposals and issue builds
  did not.
- Leftover worktrees: each run folder names its worktree in `worktree`. At the start
  of `tend propose` and of a run, a worktree whose pull request is merged or closed
  is removed with its local branch and the terminal's place (Herdr's `worktree
  remove`, or the rex sessions recorded for it, killed and forgotten), only when its
  head is the pull request's and it is clean; otherwise left and said. A declined
  request's worktree goes at once. The `tend-HHMMSS` branch a cycle's worktree is
  made with is deleted after the claim. An issue build whose pull request is merged
  is done, so tidying cannot lead to building it again. Folders from before this
  change name no worktree and are not tidied.
- A declined change request: the implementer is told to write its reply even when
  it changes nothing. With no commit and a reply, the reply goes on the pull request
  as the account's comment, once (found by a marker naming the review), and
  `answered-<review node id>` keeps tend and `tend watch` from asking again; a new
  change request on the same head is asked. With no commit and no reply, tend
  stops for a person as before.
- A stale description: tend records in `described` the head whose message is on
  the pull request. After a later head is reviewed, the implementer writes the
  message again, given the current one to keep what holds, and tend puts it on the
  pull request with `gh pr edit`, unless the title or description differ from what
  tend set, in which case a person has edited them and tend leaves them for good.
  Pull requests opened before this change are not rewritten.
- Checks: 281 pass, 31 new; against `main`'s build 20 of them fail (the fetch of a
  cycle passes on both, as Grove's claim fetched).
- Not exercised live. Not done: a closed, unmerged proposal is still never
  redrafted, and comments on a proposal are not acted on (revision stays manual).

The owner merged `tend/known-gaps` into `main` (`38b21d0`).

2026-11-16 — The known gaps, live, and a rex fix. With the owner's approval, the
session acted as the teammate on the sandbox with the owner's account. A new change
`261116-r0001` (report refuses an invalid month) and cycle `261116-k0001` were
committed to the sandbox's `main` as the fixture owner; tend, agents in rex, built
it as PR #12 by `mascah-bot`.
- First run: the reviewer's prompt stopped as `agent_prompt_stalled`, though Codex
  wrote a success verdict 30 s later. Codex in rex reports no state while it works,
  only `done` at the end of its turn (record at 19:53:06, verdict at 19:53:03), so
  the five-second delivery gate could not hold. Fixed on `tend/rex-first-report`
  (`8944fc8`, not merged): in rex a prompt is stalled only when the agent reports
  nothing within the prompt's timeout. The rerun reused the verdict and published.
- Out of scope: the owner asked for a `--year` option. One `tend watch` pass ran tend;
  the implementer committed nothing and replied that the change's proposal puts other
  report options out of scope; tend posted the reply on PR #12 as `mascah-bot`. A
  second pass ran nothing; still one comment.
- Scope-changing: the owner asked for the error to end with an example month. The
  implementer committed `8c79f5d`, Codex passed it, and tend rewrote the PR's
  description to quote the new error.
- Merged: the owner's account squash-merged PR #12; the next `tend propose` removed
  the worktree, its branch and its rex session. The worktree of the earlier PR #9,
  from before run folders named their worktree, stays.

Next action: run tend for real on the owner's mini, in a dedicated `tend` account
with Herdr: grove (GitHub issues), memex (Jira) and tend itself, which adopts Grove
on `tend/adopt-grove`. Merging `tend/rex-first-report` is the owner's.

A suitable prompt for a future implementation session is:

> Read `docs/work/261114-reliable-orchestration/spec.md` in Tend and the repository
> instructions. Revalidate the current handoff; slice 4 and the known gaps are
> done, so show the gaps' behaviour live, then agree the next unit with the owner. Preserve unrelated work and active Grove
> implementation. Update the delivery status with the revision, checks and next
> action; do not advance into proposal intake or broader polling.

When a slice is complete, update this example and the starting pointer above to
the next eligible slice so that stale handoff text does not restart finished work.
