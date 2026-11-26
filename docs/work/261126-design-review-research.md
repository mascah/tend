# Tend and Grove design review: research record

Written 2026-11-26 by Claude, from one working session with the owner. This
file holds what was checked, how, and what it showed, so a later conversation
can continue the design discussion without repeating the work. It decides
nothing. Sections 1 to 10 are evidence. Section 11 is Claude's analysis and is
opinion. Section 12 lists what the owner has said. Section 13 lists what the
session got wrong and corrected. Sections 14 to 17 were added later the same
day from a second session: they record what the owner decided in threads 1
and 2 of the design discussion, the evidence behind it, and how to continue.

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
| Rulesets and protected branches on a private personal repository need GitHub Pro; bypass lists and merge queues need an organization | docs.github.com, about protected branches and rulesets |

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
  protected branches are available on private keyborg. Bypass lists and
  merge queues still need an organization.
- How the mini's Claude login behaves outside a desktop session. Guess,
  untested: `setup-token` avoids the question.
- Rebuilding the record from GitHub trusts the `tend/review` status.
  Whether an agent can forge it is thread 3.
- Not reproduced: the overlap finding and the `--retry` bypass in 14.1.

## 17. How to continue

Threads 1 and 2 are decided. Remaining, in order:

3. Authorization without Claude Approvals: what moves to GitHub rulesets
   and Actions, where a judged gate runs, whether `land` shrinks.
4. Intake batching: the triage step's contract, and Grove's
   one-change-per-proposal assumptions.
5. The queue and a TUI, with the multiplexer as an optional view.
6. The smallest laptop mode and what it drops.
7. Docs: which two or three diagrams once states are explicit.
8. What must be true before Tend runs unattended on keyborg.

Read this file first, sections 14 to 16 before 11. Then the Codex review
and the direction document. Start thread 3 with the open fact in
section 16 that bears on it most: who can post `tend/review`.
