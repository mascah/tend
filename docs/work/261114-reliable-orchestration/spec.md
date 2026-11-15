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

**Next Tend work: finish slice 2, one real shared-runner workflow.** Slice 1 is
done on `tend/slice1-claim-verdict` (not merged; merging is the owner's decision).
Slice 2's publication is live on the private sandbox `mascah/tend-sandbox`, PR #1;
what remains is the owner, as the teammate, reviewing and acting through that PR
(see the session handoff). Tend publishes only to origins the runner lists. Preserve existing attempts and candidate branches. Do not
start proposal polling or multi-repository scheduling first.

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
through tmux. The non-Grove path supplies its own scope and eligibility without
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
| 2 — One real shared-runner workflow | In progress: publication live on a sandbox PR; the second person's review remains | Slice 1 | [Publication](#shared-review-and-publication) |
| 3 — Explicit request-to-proposal intake | Not started | Slice 2 and G available | [Intake and handoffs](#request-intake-and-explicit-handoffs) |
| 4 — Prove portability | Not started | Earlier lifecycle demonstrated | [Portability](#portability) |

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

### 3 — Explicit request-to-proposal intake

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

Routine implementation choices remain open: language, state encoding/location,
exact status/resume commands and integration module layout. Choose the smallest
solution that meets the scenarios; Bash remains an option. Before slice 3, settle
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

A suitable prompt for a future implementation session is:

> Read `docs/work/261114-reliable-orchestration/spec.md` in Tend and the repository
> instructions. Revalidate the current handoff and implement the next unfinished
> unit of slice 2 within its acceptance. Preserve unrelated work and active Grove
> implementation. Update the delivery status with the revision, checks and next
> action; do not advance into proposal intake or broader polling.

When slice 1 is complete, update this example and the starting pointer above to
the next eligible slice so that stale handoff text does not restart finished work.
