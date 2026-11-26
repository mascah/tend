# Tend and Grove design review: research record

Written 2026-11-26 by Claude, from one working session with the owner. This
file holds what was checked, how, and what it showed, so a later conversation
can continue the design discussion without repeating the work. It decides
nothing. Sections 1 to 10 are evidence. Section 11 is Claude's analysis and is
opinion. Section 12 lists what the owner has said. Section 13 lists what the
session got wrong and corrected. Sections 14 to 17 were added later the same
day from a second session: they record what the owner decided in threads 1
and 2 of the design discussion, the evidence behind it, and how to continue.
Section 18 comes from a third session: the evidence for thread 3, what a
sandbox on GitHub showed, and what the owner decided.

Read with: [the Codex architecture review](261125-architecture-review.md) and
[the autonomous direction](261121-autonomous-direction/direction.md).

## Summary

- The Codex review's 12 findings hold. Two were reproduced here on disposable
  repositories; one (finding 10) is documented behaviour.
- Finding 1 is worse than the review says: it also blinds a clean, unattended
  runner to newly approved work.
- The review misses four things that matter more for the owner's attention:
  transient failures that page a person, no record of a decision waiting for
  the owner, a limit on agents but none on the owner's queue, and intake that
  maps one issue to one proposal to one pull request.
- About half the findings share one root: Tend's state is implicit. It is
  about 25 kinds of marker file plus Git, GitHub and terminal status,
  re-derived by replaying the flow, with deciding and acting interleaved.
- Gas City and Gas Town both keep one record per work item. Neither has a
  single queue of human decisions, a spend cap, or (Gas Town) a gating review.
- Claude Approvals would move merge authorization off the agents' machine. It
  is not available to the owner (personal Max account).

## 1. What was examined

| Thing | Revision | How |
| --- | --- | --- |
| Tend | `c0e0be0`, branch `codex-arch` | Read every non-test Go file (4,367 lines), the built-in flows and prompts, the guide, and the direction documents. Ran `go vet ./...` (clean) and `go test ./...` (`ok`, 110 s). |
| Grove | `3cfd2ef` (0.15.2 dev build) | Read `internal/cli/claim.go`, `project.go`, `review.go`, `archive.go`, `internal/gitx/gitx.go`, `CONSTITUTION.md`. Ran the installed binary in the reproductions. |
| Gas City | `ebb5ae5` (2026-11-26), shallow clone | A delegated read-only pass answered eight questions with citations. Claude checked the cited lines marked "checked" in section 7. |
| Gas Town | `94cc822` (2026-11-12), shallow clone | Same, section 8. |
| Claude Approvals | Docs as of 2026-11-26 | Five docs pages fetched, plus the policy-file reference bundled with Claude Code. |
| GitHub | `mascah/tend`, `mascah/memex` on 2026-11-26 | `gh` reads of issues and pull requests. |

Not examined: the runner (the OS user `tend` on the mini) and its state
folder, live GitHub rulesets beyond their existence, Jira, a live agent
session, and anything in `mascah/keyborg` (private, holds a README only).
The state folder on the MacBook used for this session was empty.

## 2. The Codex review, finding by finding

"Smaller fix" is Claude's suggestion where the review's remedy is heavier than
the problem needs. It is opinion.

| # | Finding | Result | Evidence | Smaller fix |
| --- | --- | --- | --- | --- |
| 1 | A claim changes the owner's staging area in another worktree | Reproduced, and worse (section 3) | Grove `claim.go:143` runs `update-ref` on the local target whether or not it is checked out | Grove: do not move a checked-out target; read scope at the claim's commit |
| 2 | A failed setup leaves a recorded claim | Reproduced | Grove pushes the record (`claim.go:129`) before `git switch -c` (`:139`). Tend treats any non-zero exit as a refusal and removes the worktree (`attempt.go:229-232`). Tend makes the run folder only after the claim (`attempt.go:249-250`), and watch looks only at run folders | Grove: create the branch first. Tend: watch reruns a record in flight on this host that has no run folder |
| 3 | A passing review can answer a different scope question | Holds | The reviewer is pointed at the archived change folders (`attempt.go:254`). `unbuilt()` accepts any matching archive path (`attempt.go:343-346`). Grove lets an implementer revise the change on its branch (constitution rule 4) and archives dropped changes too. `grove review` prints "approves no new change" and a path list; it does not flag a revised proposal | A mechanical gate: compare the change's `proposal.md` and specs at the claim commit with the archived copies; hold for the owner when they differ |
| 4 | Review evidence shares the implementer's write boundary | Holds | Both agents get `--add-dir <run folder>` (`flow_run.go:179`, `:194`). An existing valid verdict for the head is returned without asking the reviewer (`flow_run.go:389`). The bot token comes from the OS user's `gh` keyring (`publication.go:28-34`) | None that is cheap and real under one OS user. Claude in auto mode can write anywhere its user can |
| 5 | Author association can exceed the person's authority | Holds; moot on personal repositories | `publication.go:411-413`; the comment says a read-only organization member counts | Nothing until an organization repository uses Tend |
| 6 | Recovery replays old evidence under today's flow | Holds; a policy choice | The flow is loaded from the target on every run (`flow_load.go:155`). A verdict is keyed by head only | Record the flow and prompt hash beside the verdict; do not invalidate automatically |
| 7 | Watch spends work on history every pass | Holds | `rerun()` calls `gh pr view` for every run folder with a `pr` file, finished or not (`watch.go:153-164`). Nothing marks a run finished. At 12 passes an hour, about 400 finished attempts would use a 5,000-per-hour API budget | Mark the folder once its pull request reads merged or closed |
| 8 | Intake can starve old issues and miscount capacity | Holds | `gh issue list` with no `--limit` returns 30 (`intake.go:80`). A failed count of open proposals becomes 0 (`intake.go:96-97`) | `--limit`, and stop on a failed count |
| 9 | One request or slow repository stops unrelated scheduling | Holds | Watch runs `tend propose` synchronously (`watch.go:69`) and skips it for a repository when any intake folder says `needs-you` (`watch.go:67-68`). The setup lock is one per state folder (`attempt.go:121-133`). The landing lock is held across CI polling (`landing.go:232`), which is direction decision 16 | Run each draft as its own process |
| 10 | A custom flow passes locally without the required status | Documented behaviour | `guide/flows-reference.md:154-156`: "Only verdicts named `verdict-…` are posted as reviews". Publication globs `verdict-*.json` (`publication.go:242`) | A validation message when a build flow with `publish` has no `verdict-…` step |
| 11 | Repository identity changes between locking and recovery | Holds; issue builds and drafter names only | Locks hash the common Git directory (`attempt.go:95-105`); issue run folders hash the invocation path (`attempt.go:58-63`); status compares path strings (`status.go:35`) | Use the lock's identity for both |
| 12 | An unavailable Grove reads as a repository without Grove | Holds; low risk | `attempt.go:168`; `status.go:15-18`. An issue build also needs the `tend-build` label | Check for `grove.yaml` |

One measurement differs from the review: `grove changes --json` took 0.23 s
for 63,340 bytes here, against the review's 0.92 s for 73,415 bytes.

## 3. Reproductions

Each script makes a disposable Grove project with a bare remote and uses the
real `grove` binary. Run as `bash <script> <an empty folder>`.

Shared setup, `lib.sh`:

```bash
export GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=test@example.com GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=test@example.com GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
project() { # dir
	local d=$1 target=main
	git init -q -b "$target" "$d" && cd "$d"
	echo x >a.txt && git add -A && git commit -qm init
	grove init >/dev/null
	sed "s/^target: .*/target: $target/" grove.yaml >g && mv g grove.yaml
	git add -A && git commit -qm grove
	local c=grove/changes/261015-c0000-a-change
	mkdir -p "$c" grove/cycles
	printf 'title: A change\ncreated: 2026-10-15T09:00:00Z\nschema: default\nbase: %s\n' "$(git rev-parse HEAD)" >"$c/.grove.yaml"
	printf '# A change\n\n## Why\n\nBecause.\n\n## What changes\n\nA file.\n' >"$c/proposal.md"
	printf '# Tasks\n\n- [ ] T-01 Add the file\n' >"$c/tasks.md"
	git add -A && git commit -qm propose
	printf -- '---\ntitle: "One"\ncreated: 2026-11-14T02:00:00Z\nchanges: [261015-c0000]\nplanned: %s\n---\n' "$(git rev-parse HEAD)" >grove/cycles/261114-k0000-one.md
	git add -A && git commit -qm plan
	git clone -q --bare . "$d.git" && git remote add origin "$d.git" && git fetch -q origin && git branch -q -u "origin/$target"
	cd - >/dev/null
}
```

### 3.1 Finding 1 as the review states it

```bash
set -euo pipefail; . "$(dirname "$0")/lib.sh"
T=$(mktemp -d "$1/f1.XXXX"); cd "$T"; project "$T/repo"
git clone -q "$T/repo.git" "$T/other" && (cd "$T/other" && echo new >upstream.txt && git add -A && git commit -qm "upstream adds a file" && git push -q origin main)
cd "$T/repo" && echo edited >>a.txt          # the owner's uncommitted edit on main
git status --porcelain
git fetch -q origin main                      # what Tend does with a dirty target checkout
git worktree add -q -b tend-tmp "$T/wt" origin/main
(cd "$T/wt" && grove claim 261114-k0000)
git status --porcelain; git diff --cached --stat
```

Observed: before, ` M a.txt`. After a successful claim, ` M a.txt` and
`D  upstream.txt`: the upstream file is a staged deletion in the owner's
checkout, and a plain commit there would remove it.

### 3.2 Finding 1 on a clean runner (new)

```bash
set -euo pipefail; . "$(dirname "$0")/lib.sh"
T=$(mktemp -d "$1/f1b.XXXX"); cd "$T"; project "$T/repo"
git clone -q "$T/repo.git" "$T/other"; cd "$T/repo"
git fetch -q origin main && git merge -q --ff-only origin/main   # Tend's sync of a clean checkout
git worktree add -q -b tend-tmp "$T/wt" origin/main
# origin's main moves between Tend's fetch and Grove's own fetch
(cd "$T/other" && echo new >landed.txt && git add -A && git commit -qm "something lands" && git push -q origin main)
(cd "$T/wt" && grove claim 261114-k0000 >/dev/null)
git status --porcelain --untracked-files=no
# the owner then merges a new proposal
(cd "$T/other" && git pull -q && mkdir -p grove/changes/261120-n0000-later && printf 'title: Later\ncreated: 2026-11-20T09:00:00Z\nschema: default\nbase: %s\n' "$(git rev-parse HEAD)" >grove/changes/261120-n0000-later/.grove.yaml && printf '# Later\n\n## Why\n\nBecause.\n' >grove/changes/261120-n0000-later/proposal.md && printf '# Tasks\n\n- [ ] T-01 Do it\n' >grove/changes/261120-n0000-later/tasks.md && git add -A && git commit -qm "owner merges a new proposal" && git push -q origin main)
git fetch -q origin main
git status --porcelain --untracked-files=no   # still dirty, so Tend will not fast-forward
grove changes --json | python3 -c 'import json,sys; print([c["id"] for c in json.load(sys.stdin)["changes"]])'
```

Observed: the clean checkout shows `D  landed.txt` after the claim. On every
later pass Tend's `syncTarget` sees uncommitted changes and leaves `main`
alone (`main.go:309-311`). Grove reads scope at the commit that local and
origin `main` share (`internal/gitx/gitx.go:85-96`), so it lists only
`261015-c0000` and not the newly merged `261120-n0000`. Watch keeps running
and starts nothing new, until another claim moves `main` again or someone
resets the checkout.

The window is the time between Tend's fetch and Grove's fetch inside
`grove claim`, which includes creating the worktree: a few seconds per claim.
Merges and claims are correlated (a merged proposal is what makes work
claimable; a landing frees a slot), so this is not a rare coincidence over
many claims.

Untested workaround: keep the runner's clone on a detached HEAD, so no
worktree has the target checked out.

### 3.3 Finding 2

```bash
set -euo pipefail; . "$(dirname "$0")/lib.sh"
T=$(mktemp -d "$1/f2.XXXX"); cd "$T"; project "$T/repo"; cd "$T/repo"
git branch grove                               # so refs/heads/grove/<id> cannot be created
git worktree add -q -b tend-tmp "$T/wt" origin/main
set +e
(cd "$T/wt" && grove claim 261114-k0000); echo "first claim exit $?"
git --git-dir="$T/repo.git" log --format=%B -1 refs/heads/grove/runs | head -12
git branch -D grove -q
git worktree add -q -b tend-tmp2 "$T/wt2" origin/main
(cd "$T/wt2" && grove claim 261114-k0000); echo "second claim exit $?"
```

Observed: the first claim exits 1 at `git switch` ("cannot lock ref"), yet
origin's `grove/runs` holds the attempt record. The second claim exits 1 with
"is in flight in attempt … on branch grove/…", a branch that never existed.
`grove changes --json` reports the cycle as not claimable and the run as
`in_flight: true`.

The trigger used here is contrived. A kill between the push and Tend's first
write to the run folder leaves the same state; that window is about a second,
because Tend reads the project again before it writes (`attempt.go:238-250`).

## 4. Gaps the Codex review does not cover

1. **Transient failures become human waits.** Of 106 `stop()` call sites, 46
   are a failed read or write of GitHub or Git ("could not read the reviews
   of …", "could not push …", "could not set the tend/review status …").
   Each writes `needs-you`, and watch leaves a `needs-you` attempt alone until
   a person runs Tend on it (`watch.go:150`). `land` polls GitHub every 20 s
   for up to four hours (`landing.go:234-241`), so one failed call parks a
   candidate. Replaying the flow makes a retry safe; nothing uses that.
2. **No record of a decision waiting for the owner.** There is a prose status
   line that watch matches with a regular expression (`watch.go:15-17`), a
   terminal notification, and GitHub. Open issues #24 (dashboard), #43 (say
   what is asked), #51 (answering open questions) and #54 (close the request)
   are each one view of this.
3. **The limit is on agents, not on the owner.** `watch.at_once` counts locks
   held (`watch.go:212-219`), so an attempt waiting for a person frees its
   slot (direction decision 19). `CAP` bounds open proposal pull requests at 5.
   Nothing bounds implementation pull requests or stops waiting on the owner.
4. **Intake maps one issue to one proposal to one pull request.** One branch
   per issue (`intake.go:82`), and since change 261126-c16zs a proposal head
   that touches two change folders gets a failing `tend/review`
   (`intake.go:279-280`). Grove's propose now writes every part of a split
   (Grove `3cfd2ef`). See section 10.
5. **Issue text is read at draft time, from anyone.** `mascah/tend` is public.
   Intake lists open issues labelled `tend` with no author check and reads the
   title and body when it drafts (`intake.go:80-82`), which can be minutes or
   passes after the label. The drafter runs in auto permission mode as the OS
   user that holds the bot's token. An issue's author can edit it after the
   owner labels it.
6. **Handled requests stay open and labelled.** Issue #43 is open and labelled
   `tend` although its proposal (#44) merged and its implementation (#60) is
   open. Intake skips it with one `gh pr list` call per pass (`intake.go:93`).
7. **Modes and code that may be dead weight.** Watch has two scheduling paths,
   with and without a count (`watch.go:61-74`); the uncounted one runs a whole
   build synchronously. Issue builds (`issue-<n>`) and Jira double the attempt
   logic and are where findings 11 and 12 live. `built()` has no caller
   (`attempt.go:351`). Migrations from the prototype remain
   (`attempt.go:266-289`).

## 5. Tend's state and control structure

**Markers in a run or intake folder.** Each is a separate file; there is no
index.

| File | Says |
| --- | --- |
| `attempt` | repository path and work item, for watch |
| `worktree` | the worktree's path, for tidy and containment |
| `status` | one line `<time> <working|needs-you|done>: <text>` and a `next:` line; watch and status parse it |
| `layout` | worktrees and branches before each implementer prompt |
| `<step>-asked` (`asked`, `followed` in intake) | heads at which a step was prompted |
| `<step>-done` | finished asks of the implementation |
| `again-<record>` | asks repeated once in a new session |
| `request-<review id>` | asks for a person's requested changes |
| `answered-<node id>` | a change request answered with no commit |
| `reply-<review id>`, `reply-<head>` | the implementer's reply |
| `verdict-<head>.json`, `….rejected-<time>` | the reviewer's verdict |
| `implemented` | notes for the reviewer (issue builds) |
| `message-<head>` | the commit message the implementer wrote |
| `described` | the head whose message is on the pull request |
| `pr` | the pull request's URL |
| `approves` | the flow's publish acts on an approval, for watch |
| `carried-<head>` | the head a carried verdict came from |
| `checks-said-<head>` | the failed-checks comment was posted |
| `id`, `summary`, `declined`, `linked` | intake |
| `repaired`, `repair-asked` | legacy names |

Outside the folder: `locks/<hash>-<item>` (flock), `setup` (flock),
`publish`, plus Git (branch, commits, worktrees), GitHub (pull request,
reviews, statuses, comments with hidden markers) and the agent's status in
the terminal.

**Control.** `walkFlow` starts at the flow's first step on every run and
re-derives each step's outcome (`flow_run.go:41-97`). A commit step answers
"the k-th time a walk reaches it" with "its k-th ask" (`flow_run.go:429-430`).
Record names are special-cased by flow and step (`flow_run.go:101-111`).
About 40 package-level variables hold the run's state (`main.go:38-48`,
`flow_run.go:15-25`). `stop()` exits the process (`main.go:194-200`) from 106
places, so a decision cannot be tested without running the program; the test
suite is 3,753 lines that spawn the binary.

**Findings that share this root.** 2 (no folder, so watch is blind), 6
(records not tied to a definition), 7 (no finished state), 9 (a stop ends the
whole pass), 10 (publication globs files instead of taking the accepted
result), 11 (identity computed in three places); and gaps 1 and 2 above.

## 6. Workflow numbers

`mascah/tend`, 48 pull requests as of 2026-11-26:

| Opened by | Count | Notes |
| --- | --- | --- |
| `mascah`, by hand | 28 | 7 of them on `grove/` branches, 2026-11-21 to 11-23 |
| `mascah-bot`, proposals | 6 | first on 2026-11-25 |
| `mascah-bot`, implementations | 5 | first on 2026-11-25 |
| release-please | 9 | |

Tend has run on Tend, through the bot, only since 2026-11-25, for a few hours
in total (the owner's statement; the bot's first pull requests agree).

One change touches the owner at: label the issue; read and merge the
proposal (answering open questions needs a local session, issue #51); approve
the implementation pull request; merge the release pull request; close the
issue (issue #54); and, for any stop, look at a pane and run Tend again on
the runner. One unit of work has an issue, a proposal branch and pull
request, a change ID, an attempt ID, an attempt branch and pull request, a
run folder and agent names. Nothing joins them.

## 7. Gas City

About 996,000 lines of non-test Go and 2.03 million of tests (counted here).
Lines marked "checked" were read by Claude; the rest come from the delegated
read and were not re-read.

- **One record.** Everything is a bead: tasks, mail, workflow steps, sessions.
  Three states (`open`, `in_progress`, `closed`); richer state is metadata on
  the same record under a reserved `gc.*` namespace
  (`engdocs/architecture/beads.md`). Stated rule, checked: "No status files —
  query live state" (`AGENTS.md:196`). Exceptions exist: a controller lock
  file, an event log, a file-backed nudge queue.
- **Store.** Checked: the default provider is `bd` over Dolt; `file` is an
  opt-in (`GC_BEADS=file`); the SQLite providers "were removed in favor of
  doltlite" (`engdocs/architecture/beads.md`, "Provider resolution").
- **Reconciliation.** A controller ticks every 30 s, compares desired sessions
  with session records and live processes, and acts. Liveness comes from the
  process tree, never PID files.
- **Deciding apart from acting.** `internal/workladder/decide.go` (exists,
  checked) is a pure decision core: assigned, claimed, started, escalated;
  nudges back off from 3 to 60 minutes; a started claim is recycled once,
  then escalated; the decision is committed before the action.
- **Workflows.** TOML formulas compile to a graph of beads. Checked: "The
  orchestrator executes every control bead… No agent participates in control
  execution" (`docs/reference/specs/formula-spec-v2.md:65-68`).
- **Gates are scripts with typed exits.** Checked
  (`formula-spec-v2.md:722-723`, `:796-812`): exit 0 passes; exit 75 is
  infrastructure trouble and re-runs without spending an attempt; exit 78
  holds the step as misconfigured; any other non-zero spends an attempt.
- **Definition pinning.** The formula's hash is stamped on the run and a
  command detects drift; it does not block. Retries are rebuilt from a frozen
  copy of the step.
- **Human waits.** A human gate is a bead; an order mails once when it is
  created and re-mails at most hourly; closing the bead resumes the flow.
  There is no single decision queue: dashboard, mail, gates and tool
  approvals are separate.
- **Review and merge.** No built-in reviewer or merger; both are formula
  steps. Checked: the refinery is "A workflow step, not a standing role"
  (`docs/getting-started/coming-from-gastown.md:23`) and "Mayor is overseer,
  not worker" (`AGENTS.md:172`). A pull request monitor classifies failed
  checks and conflicts and mints one deduplicated repair record per head.
- **Transport.** Providers for tmux, subprocess, exec, Kubernetes and ACP.
  Prompting is pull-based: the agent is told to claim its work.
- **Limits.** Session caps and restart limits. No spend cap.

## 8. Gas Town

About 249,000 lines of non-test Go and 226,000 of tests (counted here). Nine
named roles and roughly 30 concepts.

- **State is spread out.** One piece of work spans an issue bead, an agent
  bead, a merge-request bead, a molecule, a convoy, mail and files. The docs
  call this "discover, don't track".
- **Recovery puts liveness ahead of safety.** A new session finds work on its
  hook and continues. "Duplicate detection is best-effort… Crash recovery may
  re-execute partial work" (`docs/design/polecat-lifecycle-patrol.md`).
- **Judging stuck agents is given to AI agents**, after false positives.
  Checked: the "murder spree" reference
  (`docs/design/polecat-lifecycle-patrol.md:470`).
- **No gating review.** Checked: "There is no approval step"
  (`internal/templates/polecat-CLAUDE.md:10`); the refinery's quality review
  is "measurement-only (Phase 1): reviews are recorded but do NOT gate merges"
  (`mol-refinery-patrol.formula.toml:476-477`).
- **Human surfaces: at least eight.** Mayor chat, mail, escalations, convoys,
  a terminal feed, a web dashboard, status commands, and email or SMS.
  Escalations are records with severity, routed and re-escalated after four
  hours.
- **Transport.** tmux, with agents started in permission-bypass mode; startup
  dialogs answered by reading the screen; nudges through `send-keys` or a
  file queue drained at turn boundaries.
- **Limits.** Workers uncapped by default. Cost is recorded, never capped.
- **Admitted weakness.** Checked: "The most common complaint: tasks don't make
  it through the refinery and land on the target branch"
  (`docs/design/convoy/roadmap.md:99`).

## 9. Claude Approvals

From `code.claude.com/docs/en/approvals` and its setup, policy, classifier
and pull-request pages.

- A check named Claude Approvals on a pull request, following
  `.claude/approvals.yml`: per path, how many people must approve and whether
  Claude may approve. It can post an approving review. It never merges.
- The default branch's copy of the policy decides, so a pull request cannot
  weaken its own policy.
- A ruleset can require the check with its source pinned to the Claude app.
  A process on the agents' machine cannot forge it.
- Conditions read facts (paths, authors, size) and classifiers. They cannot
  read CI results, file contents, titles or draft state. A custom classifier
  reads the title, description and diff, and the docs warn that author text
  "can try to steer the classifier's answer".
- `mode: observe` reports what the policy would decide and blocks nothing.
- Public beta for Claude organizations on Team and Enterprise plans.

**Not available to the owner**, who is on a personal Max account. Whatever
moves authorization off the agents' machine has to be built from GitHub's own
features or something else.

## 10. The memex batching case

The five newest open issues in `mascah/memex` on 2026-11-26, all filed
2026-11-25:

| Issue | What | Files it names | Needs the owner? |
| --- | --- | --- | --- |
| #202 | Hand-off gives an absolute path `memex summary add` refuses | `crates/memex-core/src/handoff.rs` | No |
| #204 | Hand-off composer does not widen with wide text | `app/src/handoff/handoff.css` | Yes: widen or not |
| #205 | Summarize fails on a recorded note that is not a meeting | `app/src/meeting/Summary.tsx` | Yes: what should it do |
| #206 | Summarize on a transcript with no turns | `app/src/handoff/composer.ts` | No |
| #203 | Notes tab opens scrolled to the end | meeting page editor | No |

Four share the Summarize hand-off path. Three were raised by a Codex review
of pull requests #196 to #201 or found while building them. Through Tend as
built: five labels, five proposals, five builds that Grove's file-overlap
rule would run in sequence, ten merges. As one change: five tasks, two
questions, one proposal, one pull request.

The owner already batches by hand in memex: pull request #194 proposed six
changes from one day's usage notes, and a cycle built them.

## 11. Analysis (Claude's opinion; nothing here is agreed)

1. **Two kinds of state.** Agreed scope is slow, human-reviewed, and needs
   diff, history and approval; Markdown in Git fits it, and approval as a
   merge depends on it. Runtime state is fast, machine-written, and needs
   atomic transitions and queries across repositories; that is what sits in
   Tend's marker files. The defects found in Grove are two lines in `claim`
   and one read-model issue (scope read at what the local clone and origin
   share, where a machine caller wants a named commit). None says files in
   Git are the wrong store for scope.
2. **The implicit state is a core design flaw of Tend.** It is the root of
   half the review's findings, the missing queue, the slow tests and the
   complicated diagrams. The remedy Claude proposed is to rewrite Tend's core
   (`attempt.go`, `flow_run.go`, `watch.go`, about 1,150 lines) around one
   attempt record with explicit states and a pure function from record and
   observations to the next action. The marker files are already "what Tend
   did"; the change consolidates and types them and separates deciding from
   acting.
3. **Where the record lives decides which runtimes are possible.** Claude
   first said one JSON file per attempt or one SQLite file. That assumes one
   long-lived machine. A CI job or a fresh cloud VM has no state folder, so
   the record would have to live in Git or GitHub. Grove's claim ledger,
   commits on a ref with a compare-and-swap push, is the one part of the
   stack that already works from an ephemeral runner. This is unresolved and
   is the link between the state question and the runtime question.
4. **Tend should not decide who may merge.** Authorization that runs as the
   same OS user as the agents cannot be trusted by construction. Moving it to
   GitHub (rulesets, required checks from a pinned source, mechanical gates as
   Actions jobs) would let Tend drop protected paths, `hold`, `approvedOn`
   and `vouch`, and would remove the reason for a second OS user. A workflow
   run on a pull request uses the pull request's copy of the workflow, so
   `.github/workflows/` would itself need a human-approval rule. Without
   Claude Approvals, a judged gate has no hosted home yet.
5. **Batch at intake with a triage step, not a standing agent.** A fresh
   agent reads open requests and what Grove holds, writes a grouping that
   Tend validates, and the owner accepts it as one decision. Both reference
   projects moved judgment out of resident agents.
6. **Build the queue from the record.** One list of what waits on the owner,
   each with one action. A TUI is then a view; without the record it scrapes
   marker files.
7. **Order Claude suggested.** Fix Grove's two claim bugs; settle the state
   model before patching the current core; then authorization, intake
   batching and the queue.

## 12. What the owner has said (2026-11-26)

- The aim is the middle between supervised and fully autonomous, not a dark
  factory. Attention is split across issues, proposal pull requests, Tend's
  own development and terminal panes.
- The goal is to get Tend and Grove to a state where they can be set loose on
  `keyborg`. memex has little development planned and is not the test bed.
- Agrees: Grove moving a checked-out target, and the Tend fixes, need doing.
- Considers the implicit state and the interleaved deciding and acting
  critical flaws, and wants the Tend and Grove approach reconsidered; raised
  starting a replacement for Grove that does not use files in Git.
- Requiring GitHub in the stack is acceptable. Wonders whether a GitHub App
  should separate concerns.
- Thinks the stack may be more complicated than it needs to be: most people
  want one command on a laptop, without several OS users.
- Wants an interface with dashboards and queues, possibly a TUI that works
  with any multiplexer (tmux, Herdr, rex).
- Wants the docs elegant, visual and quick to understand; finds the current
  mermaid diagrams complex.
- Wants to understand the runtime environments that could be supported
  (local machine, CI, cloud VM) and is willing to compromise for the right
  solution.
- Does not have Claude Approvals.

## 13. Corrections made during the session

- Claude first counted 12 Tend-built implementation pull requests by branch
  name. Seven of those were opened by hand. The bot opened 5, all since
  2026-11-25.
- Claude wrote "a week of Tend-on-Tend". It has been a few hours since
  2026-11-25.
- Claude recommended freezing features and running on memex for two weeks.
  The owner rejected this: fixes are plainly needed, and memex is not the
  target. Claude withdrew "freeze"; what remains of the point is in 11.7.
- Claude ranked the state problem third and offered a patch. The owner said
  that undersold it; Claude agreed.
- The owner recalled Gas City reverting to file state by default. Its default
  is still `bd` over Dolt (section 7).
- Claude listed autonomous proposal approval under "not now" without a
  reason. The reason was ordering: proposal merge is the owner's only scope
  checkpoint and finding 3 means it leaks.

## 14. Thread 1, the state model: decided 2026-11-26

The owner said yes to both decisions in 14.2. Codex gave its own take on the
thread afterwards (14.4). The owner then said yes to the four amendments
Claude recommended from it.

### 14.1 New evidence

- **A claim whose attempt ended without landing stays in flight.**
  Reproduced on Grove `3cfd2ef`. The script uses `lib.sh` from section 3
  with the cycle left out (drop the `printf … >grove/cycles/…` line and the
  `plan` commit), so the change is loose:

  ```bash
  set -euo pipefail; . "$(dirname "$0")/lib.sh"
  T=$(mktemp -d "$1/closed.XXXX"); cd "$T"; project "$T/repo"; cd "$T/repo"
  git switch -q --detach
  git worktree add -q -b tend-tmp "$T/wt" origin/main
  (cd "$T/wt" && grove claim 261015-c0000 >/dev/null)
  B=$(git -C "$T/wt" branch --show-current)
  # the pull request is closed; tend tidies the worktree and branch
  git worktree remove --force "$T/wt" && git branch -q -D "$B" tend-tmp || true
  grove changes --json | python3 -c 'import json,sys; v=json.load(sys.stdin); print([(c["id"],c.get("claimable")) for c in v["changes"]], [(r["id"],r["in_flight"]) for r in v["runs"]])'
  git worktree add -q -b tend-tmp2 "$T/wt2" origin/main
  set +e; (cd "$T/wt2" && grove claim 261015-c0000); echo "second claim exit $?"
  ```

  Observed: the change is `claimable: false`, the run `in_flight: true`, and
  the second claim exits 1 with "is in flight in attempt …". On the Tend
  side, read and not run: `rerun` considers only open pull requests
  (`watch.go:163`) and `stopped` needs a `working` status (`watch.go:184`),
  so nothing reruns the attempt and nothing tells the owner. Every change
  that overlaps its files waits too. Only `grove claim --retry` by hand
  clears it.
- **Grove moves the local target because its reads depend on it.** Every
  instruction, check and review reads scope through `agreed()`
  (Grove `project.go:62`), which is the commit the local target and origin
  share (`gitx.go:85-96`). `claim.go:142-145` says so. Skipping the ref move
  alone would make the implementer read stale scope.
- **Stops.** 106 `stop()` sites; 54 are in `publication.go` (22),
  `landing.go` (17) and `intake.go` (15).
- **Rewrite size.** A pure decision needs the GitHub reads hoisted out of
  publication, landing and intake: `attempt`, `flow_run`, `watch`, `intake`,
  `landing`, `publication`, `main`, `status` and `worktree` are about 3,000
  of 4,400 non-test lines. Section 11.2's 1,150 is too low.
- **Overlap is computed against today's plan (Codex; read here, not run).**
  `claim` loads the target as fetched now (`claim.go:74`) and `busy()`
  compares the new piece with the in-flight change's plan as that commit
  holds it (`claim.go:232-238`), not as the in-flight attempt's own commit
  held it. An amendment that lands while A is in flight changes what A is
  taken to reserve.
- **`--retry` skips every `busy()` refusal** (`claim.go:95`), overlap with
  other work included, not only "this piece is in flight".

### 14.2 Decisions (owner: yes to both)

1. **Keep Grove and scope in Git; change the contract between Grove and
   Tend.** Approval as a merge is the only scope checkpoint GitHub enforces
   for free, and thread 3 depends on it. No Grove replacement, no shared
   store. The fault is that an attempt has two owners: Grove records its
   start, Tend its life, nobody its end. The contract changes:
   - Reads take a commit: `grove changes --json --at <commit>`, and
     instructions on an attempt branch read at the claim's commit. Removes
     `syncTarget` and Tend's reason to touch anyone's checkout.
   - A machine claim writes the ledger only and prints the record as JSON:
     no branch made, no ref moved. Tend makes the worktree from the record.
     Removes the temporary-branch dance (`attempt.go:226-237`) and the poll
     in `watch.go:256-266`.
   - Typed results: claimed, refused with a reason, failed.
   - Needs amendments to Grove constitution rules 2 and 4.
2. **Tend keeps one record per run of a flow, build or propose**, as a local
   JSON file with one writer, replaced atomically. It holds past-tense facts
   only Tend knows (what it asked, accepted and decided). Anything Git,
   GitHub, Grove or the agent can answer is observed each pass, never
   stored. The loop is observe, decide, act: `decide` is a pure function
   returning one of a closed set of actions; a failed observation is a
   retry, never a human wait; asking the owner is the only output the queue
   shows. **Invariant:** after publish the record can be rebuilt from the
   branch, the pull request and the ledger, and a test proves it. Today
   `passed()` reads only local verdict files (`landing.go:22`).

### 14.3 The record as sketched

```
runs/<id>/record.json
runs/<id>/def/          the flow and prompts this run started under
runs/<id>/out/<ask>/    the only place that ask's agent may write

id, kind (build | propose), repo identity, origin
subject   {changes, cycle, request}
claim     {commit, branch, host}
def       {flow hash, read from which commit}
asks      [{n, step, agent, session, head, prompt hash, sent, ended, how, cost}]
accepted  [{ask, head, kind, file, sha256, conclusion}]
moves     [{step, head, outcome, to}]
pr
wait      {on, why, act, head, since}
ended     landed | closed | superseded | dropped

observe(record) -> snapshot | error
decide(def, record, snapshot) -> Prompt | Await | Publish | Merge
    | Wait(ci | turn | slot) | Retry(after) | Ask(owner, why, act) | End(disposition)
act(action)     writes the record before anything not repeatable
```

`asks`, `accepted` and `moves` replace about 20 marker kinds and the replay
from the flow's start. `described`, `carried-*`, `checks-said-*` and
`answered-*` go: GitHub answers them. The join across request, change,
attempt and pull request is the Grove change ID; what is missing is the
request reference in the change's `.grove.yaml`, written at propose time.

### 14.4 Codex's take on thread 1, and what Claude would change

Codex agrees on: Grove's Git scope kept, the claim protocol changed, Tend's
core rewritten, a pure decision function, intent written before action, the
Grove claim ID used as Tend's attempt ID.

Four points adopted from it (owner: yes to all four). The sketch in 14.3
predates them: its `wait` field becomes a reference to a decision record,
and `claim` gains the caller's ID.

1. **The ledger gets a release record naming the attempt.** Claude had said
   the ledger needs no end record. That contradicts Claude's own diagnosis
   (nobody records the end), and `--retry` is too broad a substitute.
   Retry should mean "replace this attempt, then check the other
   reservations". Two record kinds: claim and release. Parking needs none:
   a parked attempt is claimed and not released (direction decision 21).
2. **A decision is its own record**, not a `wait` field on a run. It names
   its subject and revision, the question, the one action, and, once
   resolved, the answer and where it came from. Reason: a closed pull
   request ends the attempt, but "retry, drop or amend?" outlives it. The
   queue is then the open decisions. The answer must be observed from
   GitHub with a source, never typed into a local file (thread 3).
3. **The caller supplies the attempt ID**, so Tend writes its intent first
   and a repeated claim finds the same record. Replaces Claude's "adopt an
   orphan by host".
4. **Overlap is read at each in-flight attempt's own claim commit**, and
   the actual text sent in each ask is kept beside the record.

Resolved by thread 2: Codex's open question was whether a lost runner must
let another machine continue the same attempt. With a persistent disk
required, a replacement started from the last published candidate is enough.

### 14.5 Where this departs from section 11

- 11.1: right conclusion, weak reason, and it undersold Grove's defect.
- 11.2: no stored state enum; the record covers intake runs too; the size.
- 11.3: wrong that the record must live in Git or GitHub. What an ephemeral
  machine loses is the worktree and the agent session.

## 15. Thread 2, runtime environments: decided 2026-11-26

### 15.1 Decisions

1. **One runtime class: a machine whose disk persists** (laptop, the mini,
   or a Linux VM). Owner: "I can live with a persistent disk for now."
   Hosted CI runs gates, not agents. Anthropic cloud sessions are parked.
   The record, the worktree and the agent's session transcript must share a
   disk until a push.
2. **Agents run headless; Tend's correctness needs no multiplexer.** Owner:
   yes, and "having a multiplexer improves the user experience greatly".
   So a multiplexer stays as an optional view (open a run's session or its
   output in a pane), designed in thread 5, never as the transport.

### 15.2 Facts and how each was checked

| Fact | Source |
| --- | --- |
| `claude` 2.1.332 has `-p`, `--session-id`, `--resume`, `--output-format json`, `--json-schema`, `--permission-mode auto`, `--max-budget-usd`, `--permission-prompts none`, `--bg`, `--cloud` | `claude --help` on the MacBook |
| `claude agents --json` lists every session with `kind`, `status` (idle, busy) or `state` (done, stopped), `cwd`, `pid`, `sessionId` | Run on the MacBook |
| `codex` 0.172.0 has `exec`, `exec resume <id>`, `--json`, `-o`, `--output-schema`, `-s` | `codex --help` on the MacBook |
| `-p` shows no trust dialog; in auto mode a refused action fails and the turn continues | code.claude.com/docs/en/headless.md, permission-modes.md (delegated read) |
| The JSON result has `session_id`, `result`, `total_cost_usd`; `is_error` and `num_turns` unconfirmed | Same |
| A `-p` session resumes headless or interactive by ID | code.claude.com/docs/en/sessions.md |
| Transcripts are `~/.claude/projects/<project>/<session-id>.jsonl`; no documented move between machines | Same |
| `claude -p` and the Agent SDK draw from the subscription's limits. A separate credit announced for 2026-06-15 was paused; a 2026-10-07 update says nothing changes | support.claude.com/en/articles/15036540 |
| `claude setup-token` gives a one-year token that can only make model requests; the GitHub Action accepts it and bills the subscription | code.claude.com/docs/en/authentication.md, github-actions.md |
| Cloud sessions are on Max, count toward plan limits, start with `claude --cloud "<prompt>"`; routines are a research preview | code.claude.com/docs/en/claude-code-on-the-web.md, routines.md |
| A hosted Actions job runs at most 6 hours | docs.github.com/en/actions/reference/limits |
| Linux hosted minutes about $0.006; self-hosted not charged | Secondary sources only (runs-on.com and others) |
| Codex in CI: an API key is the documented path; the ChatGPT login wants a persistent `CODEX_HOME` | developers.openai.com/codex/noninteractive, /codex/auth/ci-cd-auth |
| `mascah/keyborg` is private with no rulesets; `mascah/tend` has one ruleset, `main` | `gh` on 2026-11-26 |
| Rulesets and protected branches on a private personal repository need GitHub Pro; merge queues need an organization. Bypass lists do not (corrected in 18.3) | docs.github.com, about protected branches and rulesets |

### 15.3 What headless changes in Tend

- Claude: `claude -p --session-id <uuid> --permission-mode auto
  --output-format json`, then `--resume` for repair and requested changes.
- Codex: `codex exec -s workspace-write --output-schema <schema> -o <file>`.
  The verdict becomes the process's structured output, captured by Tend, so
  the reviewer needs no write access to the run folder. This narrows
  finding 4; it is no boundary under one OS user.
- Each ask runs detached and writes its result to a file, so a turn
  outlives a killed Tend as a pane does now.
- Goes: `herdr.go`, `rex.go`, `terminal.go` (547 lines), dialog answering,
  the `blocked` and `unknown` states, the was-it-delivered ambiguity
  (`flow_run.go:248-254`).
- Lost: typing into a running turn. Instead: tail its output, or stop it
  and resume the session by hand.
- Keep the agent contract to three calls (start or resume with a prompt,
  is it running, its result) so a pane transport could return if headless
  use is ever billed apart. Do not build one now.

## 16. Open facts carried forward

- Whether the slash skill `/grove-implement --runner <id>` runs under
  `claude -p`. Fallback: send what the skill runs, `grove instructions
  implement` (constitution rule 9). One real run on a disposable project
  settles it.
- Settled: the owner has GitHub Pro (owner's statement), so rulesets and
  protected branches are available on private keyborg. Merge queues still
  need an organization.
- How the mini's Claude login behaves outside a desktop session. Guess,
  untested: `setup-token` avoids the question.
- Rebuilding the record from GitHub trusts the `tend/review` status.
  Answered in 18.1: any agent can post it.
- Not reproduced: the overlap finding and the `--retry` bypass in 14.1.

## 17. How to continue

Threads 1 and 2 are decided. Thread 3 is decided in its main lines
(18.6); 18.7 lists what is left of it. Remaining, in order:

3. Authorization without Claude Approvals: what moves to GitHub rulesets
   and Actions, where a judged gate runs, whether `land` shrinks.
4. Intake batching: the triage step's contract, and Grove's
   one-change-per-proposal assumptions.
5. The queue and a TUI, with the multiplexer as an optional view.
6. The smallest laptop mode and what it drops.
7. Docs: which two or three diagrams once states are explicit.
8. What must be true before Tend runs unattended on keyborg.
9. Product research: how agents do the research that makes an autonomous
   proposal meaningful. Raised by the owner on 2026-11-26 as something
   keyborg needs before it runs unattended (direction decision 4 and its
   open questions).

Read this file first, sections 14 to 16 before 11. Then the Codex review
and the direction document. Thread 3 continues from 18.7.

## 18. Thread 3, authorization (2026-11-26)

18.1 to 18.3 and 18.5 are evidence. 18.4 is Claude's proposal as first
put, kept for the reasoning. 18.6 is what the owner decided and what the
sandbox changed. 18.7 is what remains.

### 18.1 Who can post `tend/review`

Every agent on the runner, in one API call. Read live on 2026-11-26:

- `mascah/tend` has one ruleset, `main` (id 28050972). Its only rule is
  required status checks: `check` and `tend/review`, branch up to date,
  neither with a source pinned (no `integration_id`). No pull request or
  review rule. Bypass: the repository admin role, always.
- `tend/review` is a commit status. Tend posts it with the bot's token
  (`landing.go:79`, `publication.go:257`, `intake.go:256`), which it reads
  with `gh auth token --user <bot>` (`publication.go:28-34`). Any process of
  that OS user can run the same command; both agents run as that user.
- The bot's token is a classic OAuth token with `repo`, `workflow`, `gist`
  and `read:org`. `mascah-bot` has the write role.
- On the MacBook the same keyring holds `mascah` (admin, `repo` and
  `workflow`) as the active account. Every `gh` call of this session ran as
  the owner. An agent there can approve as the owner, merge by bypass, or
  edit the ruleset.
- The rule is bypassed in practice. All 9 merged pull requests the bot
  opened were merged by `mascah`; 7 of the merged heads carry no
  `tend/review` at all (#39, #40, #42, #44, #53, #56, #59). The two that do
  (#48, #57) show `success`, creator `mascah-bot`.
- `check` on those heads is a check run from GitHub Actions (app 15368).

So `tend/review` records what Tend said and proves nothing about who said
it. Rebuilding a record from it (14.2) rebuilds a claim, not evidence.

### 18.2 What else the bot's token allows today

- **Change CI in its own pull request.** With `workflow` scope it can push
  an edit to `.github/workflows/ci.yml`, and a `pull_request` run uses the
  pull request's copy of the workflow.
- **Approve through Actions.** The repository allows Actions to approve
  pull requests (`can_approve_pull_request_reviews: true`). A workflow
  pushed on any branch can ask for `pull-requests: write` and approve any
  open pull request as `github-actions[bot]`. Harmless today only because
  no rule asks for an approval.
- **Merge.** Write access merges any pull request whose required checks
  pass, and it can post one of the two itself.
- `ci.yml` runs the pull request's code with a read-only token and no
  secrets. `release.yml` holds write permissions and runs on pushes to
  `main` only. No workflow runs pull request code with a write token.
- `mascah/keyborg`: private, `mascah` the only collaborator, no ruleset,
  auto-merge off.

### 18.3 GitHub facts

From a delegated read of docs.github.com and the GitHub changelog on
2026-11-26, each with the quoted sentence. "Inferred" means the docs imply
it and do not say it.

| Fact | Status |
| --- | --- |
| Anyone with write access can set any commit status. A required check with no source pinned accepts it. With a source app pinned, a status from anyone else blocks the merge | Confirmed |
| "If a check and a commit status have the same name, both must pass": a posted `check` status cannot override a failed Actions run | Confirmed |
| Required checks match the job name only, whatever the workflow. A skipped job counts as success. Same-named jobs give "ambiguous" results | Confirmed; what a same-named job in an edited workflow does is not documented |
| On a personal account with Pro: rulesets with required approvals, code-owner review, approval of the most recent push, and required checks. Not available: required workflows (organization), push rules that restrict file paths (Team plan), merge queue (organization) | Confirmed; the four review settings inferred from the absence of a limit |
| Bypass actors on a personal repository: the admin role, a user, a GitHub App, a deploy key. Corrects 15.2 and 16, which said bypass lists need an organization; the live ruleset has one | Confirmed. That an empty list binds the owner is inferred |
| An author cannot approve their own pull request. Approval of the most recent push must come from someone other than the pusher, and earlier approvals are kept | Confirmed |
| An approval by `github-actions[bot]` counts when the repository setting allows it; new personal repositories default to off | Inferred from the setting's stated purpose. Whether another App's approval counts: not found |
| CODEOWNERS is read from the base branch; available on private repositories with Pro; owners are users with write access | Confirmed for review requests; inferred for the merge requirement. Deletions and renames under an owned path: not found |
| A fine-grained token cannot serve a collaborator on another account's repository. Classic `repo` includes commit statuses; no classic scope pushes to a private repository without them | Confirmed |
| Without `workflow` scope a token cannot add or update a workflow file, unless an identical file is on another branch of the repository | Confirmed as a scope rule. That a `git push` is refused, and what a deletion does: not found |
| `pull_request` and `pull_request_review` run the workflow file of the pull request's merge commit. `pull_request_target` and `issue_comment` run the default branch's. A pull request from a branch of the same repository gets secrets, and a write token if its workflow asks | Confirmed; the last sentence inferred (limits are stated for forks only) |
| From 2026-12-02 public repositories block `pull_request_target` unless an Actions event policy allows it. Private repositories are unaffected | Confirmed |
| An environment limited to `main` is closed to `pull_request` jobs and open to `pull_request_target` jobs. Environment secrets and branch limits work on private repositories with Pro; required reviewers do not | Confirmed |
| Auto-merge: available with Pro; a collaborator with write access can enable it; GitHub merges once required reviews and checks pass | Confirmed |
| A personal account can own a GitHub App with separate permissions for contents, pull requests, commit statuses, checks and workflows | Confirmed |
| Whether an approval survives a merge of the base branch into the head, under "dismiss stale approvals" and under "approval of the most recent push" | The docs contradict themselves. Settled by the sandbox (18.5): it survives. `guide/setup.md:173` says the opposite and is wrong |

### 18.4 Proposal as first put (Claude's; see 18.6 for what was decided)

**Frame.** A gate authorizes only when the credential that satisfies it
cannot be read by the agents' OS user. A credential can live in three
places: the runner, a workflow on the default branch, or the owner's own
devices. That gives each kind of gate one home:

| Gate | Home | How GitHub enforces it |
| --- | --- | --- |
| Mechanical (CI; scope unchanged; clean merge) | A workflow the bot cannot edit | Required check pinned to GitHub Actions; an approval from the workflow |
| Human (new scope, protected paths, releases) | The owner's account | Required approval; code owners |
| Judged, when it authorizes | A workflow the bot cannot edit | An approval from the workflow |
| Judged, when it only improves the work | The runner | Nothing; it is advice |

1. **Tend holds no authorization.** What must be true before a merge is a
   GitHub rule. What Tend checks before it acts is for convergence only.
2. **The runner holds one credential: the bot's, without `workflow`
   scope.** This is the linchpin on a personal account, which has neither
   required workflows nor path rules: it is what stops a pull request
   rewriting CI or adding a workflow that approves. Tend refuses to run
   when the bot's token carries `workflow`, and `tend watch` refuses when
   `gh` holds a token with admin on the repository. A machine where the
   owner is signed in beside the agents is supervised by definition
   (thread 6). No second OS user is needed for authorization.
3. **The authorizing act is an approving review, required by the ruleset,
   from someone who did not push the head.** Ruleset: pull request
   required; 1 approval; approval of the most recent push; code-owner
   review; `check` pinned to GitHub Actions; branch up to date; the owner
   the only bypass. `tend/review` stops being required. Two approvers
   exist that an agent cannot impersonate: the owner, and a workflow on
   the default branch.
4. **Protected paths become CODEOWNERS.** Decision 3a's list moves from
   the flow file's `protect` to a file GitHub reads from the base branch.
5. **Stage 1, supervised and enforced.** The owner is the only approver.
   Configuration, plus deletions in Tend.
6. **Stage 2, the hosted gate, is what makes a merge autonomous.** One
   workflow on the default branch that runs nothing from the pull request.
   It approves an implementation when: the branch is a claimed attempt;
   `grove review` says it approves no new change and the proposal and
   specs equal those at the claim commit (finding 3); and a model, given
   the approved proposal and the diff and no tools, answers that the diff
   stays inside the proposal and crosses none of decision 3b and 3c. A
   head that is a clean merge of the target into an approved head is
   approved with no model (decision 17). Anything else waits for the
   owner. The model's credential is the `claude setup-token` token (15.2)
   in an environment limited to `main`, read by a `pull_request_target`
   job. This is the part of Claude Approvals the owner needs, as one file.
7. **The runner's reviewer stays, as advice.** It runs the code, finds
   defects and drives the repair. Its verdict authorizes nothing, so
   finding 4 stops mattering and nothing rebuilds authority from it.
   Alternative not recommended: move the whole review to the hosted gate.
   One reviewer and no verdict files, but it cannot run the code, each
   repair waits for CI first, and Codex there needs an API key.
8. **`land` shrinks to two acts:** merge the target into the front
   candidate when it is behind, and enable auto-merge. GitHub waits for
   the checks and merges. The lock held across CI polling (finding 9), the
   merge call and the parsing of GitHub's refusal go. Decision 16 holds:
   Tend still orders the train, one candidate at a time.
9. **No GitHub App now.** The workflow is already an identity agents
   cannot reach, and an App's key would have to sit in the same Actions
   secret. An App in place of the bot user would make "no workflow
   permission" structural where a token scope is a choice at login; worth
   it when a second runner or an organization appears.

Goes from Tend: `hold`, `approved`, `approvedOn` and `writer`
(finding 5), the `protect` list and its parsing (`flow.go:77-185`),
`vouch` and `notDocs` (`intake.go:207-284`), the `tend/review` posts,
`carry`, and most of `landStep`. About 570 non-test lines carry
authorization or landing today. `cleanMerges` and `notDocs` move into the
hosted gate. Needs from Grove: `grove review --json`, which today prints
text only.

Open in the proposal: whether Stage 2 is built before keyborg runs
unattended or after Stage 1 has run there; and which model the hosted gate
uses.

### 18.5 Sandbox results

Run on 2026-11-26 in `mascah/tend-auth-sandbox`, a private repository made
for this, with `mascah-bot` as a collaborator with write access. The
owner allowed it. It is disposable and still exists: the token used cannot
delete a repository.

Setup. A CI workflow on `pull_request` with one job, `check`, skipped for
paths under `norun/`. A workflow on `workflow_dispatch` that approves the
pull request whose number it is given, with `pull-requests: write`.
`.github/CODEOWNERS`: `/owned/ @mascah`, then `/owned/archive/` with no
owner. Actions allowed to approve pull requests. One ruleset on `main`
with an empty bypass list, its rules changed per test: `check` required
with or without `integration_id` 15368 (GitHub Actions), and a
`pull_request` rule of 1 approval with `dismiss_stale_reviews_on_push`,
`require_last_push_approval` and `require_code_owner_review` on or off.
The bot opened every pull request and attempted every merge unless noted,
with `PUT /pulls/<n>/merge`. Its token had `workflow` scope for tests 1 to
16 and lacked it for 17 to 25 (`X-Oauth-Scopes: gist, read:org, repo`).

| # | Test | Observed |
| --- | --- | --- |
| 1 | `check` pinned to Actions; head has no check run; the bot posts a `check` status of `success` and merges | Refused: "Required status check "check" was not set by the expected GitHub app." |
| 2 | Same head, the owner merges, not on the bypass list | Refused, same message. An empty bypass list binds the owner |
| 3 | Same head, `check` required from any source | Merged on the bot's own status. This is `tend/review` on `mascah/tend` today |
| 4 | Approval rules on; the bot approves its own pull request | "Can not approve your own pull request" |
| 5 | The bot merges with no approval | Refused: "New changes require approval from someone other than the last pusher." |
| 6 | The bot dispatches the approving workflow on its own pull request, then merges | Approved by `github-actions`, decision `APPROVED`, merged. A workflow's approval counts toward the required approval and the most-recent-push rule, and a workflow anyone with write access can dispatch is self-approval |
| 7 | Workflow-approved pull request that edits a file under `/owned/` | Refused: "Waiting on code owner review from mascah." Merged once the owner approved |
| 8 | Workflow-approved, deletes a file under `/owned/` | Refused, waiting on the code owner |
| 9 | Workflow-approved, moves a file from `/owned/` to `/owned/archive/` | Refused, waiting on the code owner |
| 10 | Workflow-approved, adds a file under `/owned/archive/` | Merged. A later line with no owner un-owns a path |
| 11 | Workflow-approved, edits an owned file and removes its line from CODEOWNERS | Refused, waiting on the code owner. The base branch's copy decides |
| 12 | Owner approves; `main` moves; the bot pushes a merge of `main` into the head (a merge commit, as Tend makes). Run three times: most-recent-push rule alone, dismiss-stale alone, both | Approval kept each time, and shown on the new head; merged |
| 13 | Both rules on; owner approves; the bot pushes a commit that changes a file | Approval `DISMISSED`; refused |
| 14 | Both rules on; owner approves; the bot pushes a merge of `main` whose merge commit also changes a file | Approval `DISMISSED`; refused |
| 15 | The bot opens a pull request that adds a workflow with `pull-requests: write`, which approves a different open pull request of the bot's | It ran at once. The other pull request showed `APPROVED` by `github-actions` and the bot merged it. No person took part |
| 16 | Branch must be up to date, both review rules, auto-merge allowed. The bot enables auto-merge on two pull requests; the owner approves both | GitHub merged the first. The second waited, behind. The bot pushed a merge of `main` into it; the approval held; CI passed; GitHub merged it. `mergedBy` is the bot, who enabled auto-merge |
| 17 | Without `workflow` scope: the bot pushes a branch that edits `ci.yml` | Refused: "refusing to allow an OAuth App to create or update workflow `.github/workflows/ci.yml` without `workflow` scope" |
| 18 | Pushes a branch that adds a workflow file found on no other branch | Refused, same message |
| 19 | Pushes a branch that renames a workflow file | Refused, for the new path |
| 20 | Adds a workflow file through the contents API | 404 |
| 21 | Pushes a merge of `main` after the owner changed `ci.yml` there; the merge commit carries that change | Accepted. A further edit of `ci.yml` on that branch: refused |
| 22 | Pushes a branch that adds a workflow file identical, in path and contents, to one on another branch | Accepted. The documented exception |
| 23 | Pushes a branch that puts `ci.yml` back to an older version still held by another branch | Accepted, by the same exception |
| 24 | Pushes a branch that deletes `ci.yml` | Accepted. Such a head has no `check`, and test 1 shows the merge is then refused as "expected" |
| 25 | Pushes a branch that edits `.github/CODEOWNERS`, and one that adds `.github/actions/x/action.yml` | Accepted. The scope covers `.github/workflows/` only |

What the results change:

- **GitHub carries an approval across a clean merge of the target and
  drops it for anything else** (12 to 14). That is decision 17, done by
  the platform and checked against the diff. `cleanMerges`, `carried`,
  `carry` and the `carried-*` marker have no job left, and the hosted gate
  needs no clean-merge approver.
- **The train needs two acts from Tend** (16): enable auto-merge once, and
  merge the target into the candidate that is behind.
- **`grove/changes/` cannot be code-owned** (8, 9): an implementation
  moves its change to the archive, and that would wait for the owner every
  time. "New scope needs the owner" has to come from who approves, not
  from a path.
- **A workflow's approval is real authority** (6, 15). While the bot's
  token has `workflow` scope and Actions may approve, the bot authorizes
  itself. `mascah/tend` has both today; nothing asks for an approval there
  yet. The hosted gate of Stage 2 must take its pull request from the
  event, never from an input.

- **Without `workflow` scope the bot cannot bring new workflow content
  into the repository** (17 to 21), and Tend's merge of the target still
  goes through when the target's CI changed (21). Three limits. It can
  delete a workflow, which fails closed for a required check (24). It can
  put a workflow back to any version another branch still holds (23): a
  downgrade to something the owner once approved, visible in the pull
  request's diff, and held for the owner in Stage 2 once `.github/` is
  code-owned (tests 7 to 9). And files a workflow calls from outside
  `.github/workflows/`, such as a local action or a script, are not
  covered (25), so the hosted gate must call none.

### 18.6 Decided (owner, 2026-11-26)

1. **Yes to items 1 to 3 of 18.4.** Tend holds no authorization. The
   runner holds one credential, the bot's, without `workflow` scope. The
   authorizing act is a required approving review from someone who did not
   push the head, with `check` pinned to GitHub Actions and `tend/review`
   no longer required.
2. **Stage 2 comes after Stage 1 has run on keyborg.** The owner added
   that keyborg waits on more than this: how agents do product research
   that makes an autonomous proposal meaningful (section 17, item 9).
3. The sandbox was allowed.
4. **`land` is two acts** (item 8 of 18.4): enable auto-merge, and merge
   the target into the candidate at the front when it is behind.
5. **No GitHub App for now** (item 9 of 18.4).
6. **The owner stays on the bypass list**, as on `mascah/tend` today.

Stage 1 as it now stands, with the sandbox results applied:

- **Ruleset on the target:** pull request required; 1 approval; dismiss
  stale approvals; approval of the most recent push; `check` required from
  GitHub Actions (`integration_id` 15368); branch up to date. The owner
  bypasses, as on `mascah/tend` today. The bot does not.
- **Repository:** auto-merge allowed; squash only. Actions not allowed to
  approve pull requests where nothing needs it: off by default on a new
  personal repository such as keyborg. `mascah/tend` has it on, which
  release-please needs to open its pull request.
- **Runner:** the bot's token without `workflow`; the owner's token not in
  that OS user's keyring.
- **The owner's act is one approval per pull request.** GitHub merges.
- **CODEOWNERS waits for Stage 2.** While the owner is the only approver
  it adds nothing.
- **Tend:** `publish` enables auto-merge; `land` merges the target into
  the candidate at the front when it is behind; the approval is read as
  GitHub's review decision. `hold`, `approved`, `approvedOn`, `writer`,
  the `protect` list, `vouch`, `notDocs`, `passed` as a landing guard,
  `cleanMerges`, `carried`, `carry`, the `tend/review` posts and the rest
  of `landStep` go. This belongs to the rewrite of 14.2, not to a patch.

### 18.7 What remains of thread 3

- **The bot's token on the mini.** Only the MacBook's was changed. Each
  machine's keyring holds its own token for the bot; the runner's needs
  the same: `gh auth switch -u mascah-bot`, `gh auth refresh -h github.com
  -r workflow` with the browser signed in as the bot, `gh auth switch -u
  mascah` where the owner is signed in too. The owner's token keeps
  `workflow`.

  On the MacBook the first refresh left the scope in place and the second
  removed it; why the first failed was not established. `gh auth status`
  prints one scopes line for each account, which is easy to misread, and
  for a while the scope was believed gone when it was not. So the check
  in item 2 of 18.4 (Tend reads `X-Oauth-Scopes` and refuses to run) is
  needed, not optional.
- **Until Tend handles it:** a change that must edit CI cannot be pushed
  by the bot. Tend stops at "could not push"; the owner pushes that
  branch.
- **Stale branches hold old workflow versions** (test 23). Deleting merged
  branches keeps the set small.
- **Deferred with Stage 2:** items 4, 6 and 7 of 18.4 (code owners for
  protected paths, the hosted gate, the runner's reviewer as advice). The
  sandbox supports item 4.
- **Stage 2's shape** (items 6 and 7 of 18.4) is deferred with Stage 2.
  Carry forward: no clean-merge approver; new scope is held by the gate
  withholding its approval; the gate is never dispatchable.
- **Correct `guide/setup.md:173`** when the guide is next touched.
- **Delete `mascah/tend-auth-sandbox`** once the scope tests are run.
