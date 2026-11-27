# Design review: evidence

Part of [the Tend and Grove design review](README.md): read that first. It
says what is decided and which file holds which section. Section numbers are
those of the single file these were split from.

Written 2026-11-26 by Claude, from one working session with the owner. This
file holds what was checked, how, and what it showed, so a later conversation
can continue the design discussion without repeating the work. It decides
nothing.

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
