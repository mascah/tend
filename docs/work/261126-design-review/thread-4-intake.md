# Design review, thread 4: intake batching

Part of [the Tend and Grove design review](README.md): read that first. It
says what is decided and which file holds which section. Section numbers are
those of the single file these were split from.

From the fourth session, 2026-11-26: the evidence for thread 4, three
proposals by Claude and a second opinion from Codex, none adopted. 19.10
says where it stands.

## 19. Thread 4, intake batching (2026-11-26)

19.1 to 19.4 are evidence. 19.5 is Claude's proposal as first put; the
owner has decided nothing in it. 19.6 corrects earlier sections.

### 19.1 How batches have gone so far

**memex, by hand** (`gh` reads; times UTC):

| Pull request | What | Opened | Merged |
| --- | --- | --- | --- |
| #194 | One proposal holding six changes, from two days' usage notes | 11-25 16:13 | 16:29 |
| #195 | The cycle that plans them, one file | 16:35 | 16:41 |
| #196 to #201 | Six implementations, one per change, each based on the branch of the one before | 17:24 to 17:41 | 11-26 00:21:07 to 00:21:13 |

The owner's acts were one approval of six changes, a second merge for the
plan, and one sitting that merged 2,492 added lines in six seconds. The
cycle did not produce one pull request; it produced a stack.

**Tend on Tend, drafted by the bot.** Questions are the list items of the
bot's summary review, counted here from the review bodies:

| Proposal | Request | Change | Questions | Marked consequential | Opened to merged |
| --- | --- | --- | --- | --- | --- |
| #39 | #36 | 261125-6dvbs | 6 | 5 | 24 min |
| #40 | #37 | 261125-rf036 | 8 | 6 | 31 min |
| #44 | #43 | 261125-kvarn | 9 | 4 | 8 h 12 min |
| #53 | #45 | 261126-5xkrh | 8 | 6 | 1 h 27 min |
| #55 | #54 | 261126-4pfj7 | 6 | 5 | still open |
| #56 | #50 | 261126-c16zs | 5 | 3 | 45 min |

Six requests gave the owner 42 questions, 29 of them consequential. Two
of those questions are about another change of the same six (#44 asks
whether 261125-6dvbs lands first; #53 asks the same of 261125-rf036):
drafting each request alone turns coordination between them into
questions for the owner.

**Those six changes share files.** Read from each `tasks.md` (`Files:`
lists) and each change's `specs/` folders: 12 of the 15 pairs name a file
in common, most often `publication.go`, `intake.go`, `landing.go` and
their tests.

**Two of the five built changes waited behind another.** From the claim
ledger (`grove/runs` on origin) and the pull requests' merge times:

| Change | Approved | Claimed | Waited | In flight meanwhile, sharing a file |
| --- | --- | --- | --- | --- |
| 261125-6dvbs | 11-25 17:12:40 | 17:12:52 | 12 s | nothing |
| 261125-rf036 | 11-25 17:23:10 | 11-26 00:03:48 | 6 h 40 min | 6dvbs, merged 23:59:33 |
| 261126-5xkrh | 11-26 02:31:42 | 02:32:43 | 1 min | nothing |
| 261126-c16zs | 11-26 03:25:31 | 03:26:56 | 1 min | nothing (5xkrh shares none) |
| 261125-kvarn | 11-26 02:35:12 | 06:11:41 | 3 h 36 min | 5xkrh, then c16zs, merged 06:10:10 |

`watch.at_once` is 2 in `tend/flow.yaml`, and an attempt waiting for a
person frees its slot (decision 19), so the count did not hold these two
back. Each was claimed within four minutes of the change it shares files
with landing. Most of each wait was the earlier pull request waiting for
the owner (#42 was open 6 h 27 min). Not checked: whether watch tried the
claim and was refused, since the runner's logs were not read; 19.2 test 5
shows Grove refuses such a claim.

### 19.2 What Grove assumes, tested

On Grove `3cfd2ef`, with this `lib.sh` (a project with a bare origin and
a helper that writes one change whose task names one file):

```bash
export GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=test@example.com GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=test@example.com GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
change() { # id slug title file [extra .grove.yaml line]
	local c=grove/changes/$1-$2
	mkdir -p "$c"
	printf 'title: %s\ncreated: 2026-11-26T09:00:00Z\nschema: default\nbase: %s\n%s' "$3" "$(git rev-parse HEAD)" "${5:+$5
}" >"$c/.grove.yaml"
	printf '# %s\n\n## Problem\n\nBecause.\n\n## Scope\n\nA file.\n' "$3" >"$c/proposal.md"
	printf '# Tasks\n\n## T-01 Edit the file\n\nFiles: `%s`\n\nAcceptance: it is edited.\n' "$4" >"$c/tasks.md"
}
project() { # dir
	local d=$1
	git init -q -b main "$d" && cd "$d"
	echo x >a.txt && echo y >b.txt && git add -A && git commit -qm init
	grove init >/dev/null
	sed "s/^target: .*/target: main/" grove.yaml >g && mv g grove.yaml
	git add -A && git commit -qm grove
	git clone -q --bare . "$d.git" && git remote add origin "$d.git" && git fetch -q origin && git branch -q -u origin/main
	cd - >/dev/null
}
```

Tests 1 to 3, one proposal branch:

```bash
set -uo pipefail; . "$(dirname "$0")/lib.sh"
T=$(mktemp -d "$1/g1.XXXX"); cd "$T"; project "$T/repo"; cd "$T/repo"
git switch -q -c tend/batch-1
change 261126-a0000 first "First" a.txt
change 261126-b0000 second "Second" a.txt
git add -A && git commit -qm "propose two"
grove check; grove review | sed -n 1,2p
mkdir -p grove/cycles
printf -- '---\ntitle: "Both"\ncreated: 2026-11-26T10:00:00Z\nchanges: [261126-a0000, 261126-b0000]\nplanned: %s\n---\n' "$(git rev-parse origin/main)" >grove/cycles/261126-k0000-both.md
git add -A && git commit -qm "plan them on the same branch"
grove check
git switch -q main && git merge -q --squash tend/batch-1 2>/dev/null && git commit -qm "approve and plan in one merge" && git push -q origin main
git switch -q --detach && git worktree add -q -b tmp "$T/wt" origin/main
(cd "$T/wt" && grove claim 261126-k0000); echo "claim exit $?"
change 261126-c0000 third "Third" b.txt 'requests: ["#1", "#2"]'
grove check; echo "check exit $?"
```

Tests 4 and 5, three changes approved separately with no cycle:

```bash
set -uo pipefail; . "$(dirname "$0")/lib.sh"
T=$(mktemp -d "$1/g5.XXXX"); cd "$T"; project "$T/repo"; cd "$T/repo"
change 261126-a0000 first "First" a.txt
change 261126-b0000 second "Second" a.txt
change 261126-c0000 third "Third" b.txt
git add -A && git commit -qm "three approved" && git push -q origin main
git switch -q --detach
for id in a0000 b0000 c0000; do
	git worktree add -q -b tmp-$id "$T/wt-$id" origin/main
	(cd "$T/wt-$id" && grove claim 261126-$id); echo "claim 261126-$id exit $?"
done
```

| # | Test | Observed |
| --- | --- | --- |
| 1 | Two drafted changes on one branch | `check`: no problems. `review` says merging approves each as future work |
| 2 | A cycle on that branch listing both, `planned` at the target's commit; then one squash merge; then `grove claim <cycle>` | `check`: no problems before and after the merge. The claim exits 0. One merge approved two changes and planned them |
| 3 | A `requests` field in a change's `.grove.yaml` | `check` fails: "unknown field "requests"; this file takes title, created, schema, description, depends_on and base" |
| 4 | Claim the first change, then a third that names another file | Both exit 0 |
| 5 | Claim the second, which names the first's file, while the first is in flight | Exit 1: "261126-b0000 shares a.txt with 261126-a0000, in flight in attempt …" |

Read, not run:

- **Unattended propose writes one change.** "Write one change … A split
  you would offer … is an open question" (Grove
  `internal/instruct/requires/propose-unattended.md`). The rule of
  `3cfd2ef`, write every part of an agreed split on one branch, applies
  with the owner in the session.
- **Plan forbids what test 2 did.** "List under `changes` only changes the
  project index lists as approved" (`requires/plan.md`). That is an
  instruction to the agent; no code enforces it.
- **`claim` takes one piece**, a cycle or a change in no cycle
  (constitution rule 2; `claim.go:85-96`). Only a cycle builds several
  changes in one attempt.
- **A change is all or nothing at archive**: every task done, or the
  change dropped (rule 3). A cycle's member that cannot finish is left
  out and the others archive (`requires/implement-cycle.md`).

So Grove's code holds no one-change-per-proposal assumption. Its
unattended instructions do, its claim takes one piece, and its change
file has no place for a request.

### 19.3 What Tend assumes

Read from the code, not run. Every one of these is one request to one
proposal:

- The proposal branch is `tend/issue-<n>` or `tend/jira-<KEY>`, and a pull
  request from it in any state means the request is handled
  (`intake.go:82`, `:93`; intake spec, "Each labelled request, once").
- The intake folder, the drafter's name, the commit's `Refs` line and the
  summary's "drafted from" each name one request (`intake.go:50`, `:122`;
  `request.go:19`; `intake.go:179`).
- The drafter is sent one request's title, body and URL
  (`flows/prompts/draft.md`).
- A decline comments on one request and takes its label off
  (`intake.go:190-202`).
- A closed proposal pull request is final. Nothing drafts that request
  again and nothing tells the request.
- `notDocs` fails a head that touches two change folders or a cycle file
  (`intake.go:265-284`). It goes with `vouch` (18.6).
- Drafts run one after another inside one pass (`intake.go:92-102`;
  finding 9).
- Open proposal #55 (261126-4pfj7) joins a request to its change by
  writing change IDs into the intake folder. That is stored local state
  of the kind 14.2 removes.

### 19.4 What the evidence says about where a batch can form

| Where | What it saves | What it costs | Needs |
| --- | --- | --- | --- |
| Before drafting: several requests become one change | Proposals to read; coordination questions between requests; the wait between overlapping builds, since one change is one claim | A larger proposal. Wrong for feature-sized requests: each of Tend's carried 5 to 9 questions alone | Judgment. A `requests` field in Grove. Tend's request identity |
| At the proposal: several changes in one pull request | Merge clicks only; the same documents are read | The owner cannot approve one and refuse another without a revision | Nothing in Grove's code (test 1) |
| At the build: several approved changes in one attempt | The wait between overlapping builds (6 h 40 min and 3 h 36 min above); implementation pull requests | A larger diff to review; one failing member holds the rest | Today a cycle file, so a merge. Otherwise `claim` taking a set |

### 19.5 Proposal as first put (Claude's; nothing here is decided)

1. **A batch is one change.** Requests that are one piece of work become
   one change with several tasks, in one proposal pull request, built by
   one claim. Grove's "write one change" stays. Feature-sized requests
   stay one each. Not recommended: several changes and their cycle in one
   pull request. It works (test 2) and tolerates a member that cannot
   finish, but it gives the owner more documents for the same work and
   needs plan's rule relaxed.
2. **The drafter does the grouping; no triage step yet.** Tend gives the
   drafter the oldest waiting request as the request, and the other
   waiting requests, up to a bound set in the flow file, as ones it may
   fold in where they are the same piece of work. It still writes one
   change or declines the oldest. Reasons: the drafter has read the code,
   so it knows what belongs together better than an agent that skims
   titles; the flow keeps its three steps; and the output is checked from
   the branch (item 3). The limit: one drafter per repository at a time,
   or two could take the same request. The owner's reading is the
   bottleneck at five open proposals, not drafting.

   The alternative is 11.5's step, with this contract: given the waiting
   requests and Grove's open changes, return for each request exactly one
   of `draft {group}` or `decline {reason}`; Tend accepts it only when
   every offered request appears once, no other does, and no group
   exceeds the bound; each group is then drafted in parallel. Add it when
   serial drafting is measurably the wait, or when requests need sorting
   before drafting (a pitch from an agent against a request from the
   owner, decision 4).
3. **Grove gains one field: `requests`, a list, in a change's
   `.grove.yaml`.** Opaque strings to Grove, which gates on nothing in
   them (rule 7); Tend writes each request's URL. The drafter fills it
   in. Tend accepts a draft only when the list holds the oldest request
   and nothing it was not offered. A request is handled when a change on
   the target, in the archive or on an open proposal's head lists it:
   observed each pass, never stored, which keeps 14.2's invariant. The
   implementation pull request says `Closes` for each (issue #54), which
   replaces proposal #55's marker file. The record's subject in 14.3
   takes `requests` in place of `request`.
4. **The owner does not approve the grouping apart from the proposal.**
   11.5 had the owner accept it as one decision. That adds a wait for a
   choice that is cheap to undo. The proposal's title and summary name
   the requests it was drafted from. To take one out, the owner revises
   the change and removes it from `requests`; it is then waiting again.
   A proposal closed unmerged is the owner's no: Tend takes the label off
   its requests and says so on each, as a decline does, and labelling one
   again tries again.
5. **When a batch forms.** Whatever is waiting when a draft starts. The
   cap on open proposals is what lets requests gather; no timer, no quiet
   period. A request that arrives while a related proposal is open is not
   folded into it; it is drafted later on its own. Add either when a
   split burst is seen to cost something.
6. **Batching at the build is deferred, and its cost is measured.**
   Approved changes that share a file keep building one after another;
   when the owner wants several built together, `/grove-plan` and its
   merge are still the way. The wait is real (19.1) and keyborg, a small
   new codebase, will show more of it. The table in 19.1 came from the
   ledger and merge times alone, so Tend can report each change's wait
   and what it waited behind with no new state. Revisit `claim` taking a
   set of approved changes with that number from keyborg, as the Codex
   review asks before any change to the overlap rule.

Out of this thread: one request that should become several changes
(decomposing a goal). Unattended, that split stays an open question; it
belongs with the autonomous steps and thread 9.

Still open from section 4, gap 5: the drafter reads a request's text
when it drafts, and an issue's author can edit it after the owner labels
it. Folding widens what one draft reads to every labelled request
offered. Unlabelled requests are never offered. A check Tend could make,
not tested: skip a request whose last edit is later than its label and
is not by someone with write access (GraphQL `lastEditedAt`, `editor`
and the `LabeledEvent` time).

### 19.6 Corrections to earlier sections

- Section 10 said the five memex issues would be "five builds that
  Grove's file-overlap rule would run in sequence". The issues name five
  different files. Whether their changes would share a file or a
  capability spec is not known until they are drafted.
- Section 10 said "a cycle built them". The cycle was a second merge, and
  the build was six stacked pull requests (19.1).
- Section 4, gap 4, said Grove's propose "now writes every part of a
  split". That holds with the owner in the session. Unattended, a split
  is an open question (19.2).

### 19.7 The owner's question: separate proposals from planning, and Dolt

The owner, on 19.5: not confident; has long wondered whether proposals
and tasks or planning should be separate in Grove, and whether that makes
a case for Dolt. Evidence gathered for it, then Claude's view. Nothing
here is decided.

**How much of a proposal the tasks are.** Lines of each document as
proposed, read from the commit that added it:

| Change | `proposal.md` | `specs/` | `tasks.md` | Tasks' share |
| --- | --- | --- | --- | --- |
| 261125-6dvbs | 136 | 173 | 50 | 14% |
| 261125-rf036 | 137 | 204 | 57 | 14% |
| 261125-kvarn | 161 | 247 | 58 | 12% |
| 261126-5xkrh | 150 | 221 | 71 | 16% |
| 261126-c16zs | 114 | 76 | 20 | 9% |

Of the 29 consequential questions in 19.1, the drafters tagged 5 as
interface, 14 as behavior and 6 as scope; 4 carry no tag.

**How well the approved plans held.** Files the tasks named at approval
against files the implementation's pull request changed outside `grove/`:

| Change | Named | Changed | Named and changed | Changed, not named |
| --- | --- | --- | --- | --- |
| 261125-6dvbs (#42) | 5 | 12 | 5 | 7: four guides, `publication.go`, two test stand-ins |
| 261125-rf036 (#48) | 14 | 15 | 14 | 1 guide |
| 261126-5xkrh (#57) | 9 | 11 | 9 | 2 guides |
| 261126-c16zs (#59) | 3 | 7 | 3 | 3 guides, `scenario_test.go` |
| 261125-kvarn (#60, open) | 13 | 14 | 11 | 1 guide, `scenario_test.go`, a test stand-in |

42 of 44 named files were changed. 17 of the 59 changed files were named
by no task; `guide/setup.md` was changed by three of the five and named
by none, so the overlap rule could not have seen those collisions.

**Grove's code for plans made ahead.** About 400 of 4,833 non-test Go
lines handle cycles, a plan gone stale and file overlap (`check.go`,
`claim.go`, `review.go`, `instructions.go`), plus 84 lines of
instructions.

**The claim record already holds a list.** Each record on `grove/runs`
has `"changes": [...]`; for a cycle it lists every member
(`claim.go:107-109`).

**Beads on Dolt** (web search, project documentation and mirrors, not
Dolt's own reference; not run): its embedded mode is one writer behind a
file lock, and several agents writing at once need a `dolt sql-server`
beside it.

**Claude's view.**

- Tasks as a document are not the weight: about a seventh of what the
  owner reads, and 5 of 29 consequential questions. The plans were right
  about the files that matter. Moving `tasks.md` out of the proposal
  would change little that was measured.
- What is fused is the unit. A change is at once what the owner approves
  and what one attempt builds, so a batch made on either side distorts
  the other: 19.5 item 1 enlarged the approval to get one build, and a
  cycle needs a merge to group a build. The separation the evidence
  supports is: approve small, build grouped.
- That revises 19.5. Item 6 moves from deferred to the main lever:
  `claim` takes a set of approved changes, Tend chooses the set (the
  oldest claimable change and those that would wait behind it, up to a
  bound), and one attempt builds them as a cycle is built today. It
  belongs to the claim contract 14.2 already rewrites, and the record's
  format does not change. Items 1 and 2, folding at intake, become the
  later addition, for the case of several small fixes that are one piece
  of work. Items 3 and 4 stand.
- Dolt is not what this needs. The state that would move there has one
  writer and a handful of rows, and Tend schedules whole changes, never
  single tasks across agents, which is where a store with many writers
  and queries earns its place. On one machine (15.1) embedded Dolt gives
  one writer behind a lock, which a file replaced atomically gives too.
  What authorizes has to stay where GitHub enforces it (18.6), so Dolt
  would be a second versioned store beside Git, joined to it with no
  transaction. Direction decision 10 names the trigger, many writers and
  queries; a second runner on the same repositories, tasks handed out one
  by one, or a queue across many repositories too slow to derive would
  be that.

### 19.8 The owner's objection: one flow for everything, or too much in one group

The owner, on 19.7: approve small and build grouped is in the right
vein, but it fails when everything goes through that flow or when one
group holds too much work. Evidence gathered for it, then Claude's view.
Nothing here is decided.

**How single changes fared in review.** The bot's review comments and the
owner's reviews on the five implementation pull requests (`gh` reads):

| Pull request | Added lines | Independent reviews | Owner asked for changes | Open for |
| --- | --- | --- | --- | --- |
| #42 (6dvbs) | 897 | 5, all failure | 3 times | 6 h 27 min |
| #48 (rf036) | 594 | failure, failure, then success | once | 1 h 22 min |
| #57 (5xkrh) | 389 | success | no | 42 min |
| #59 (c16zs) | 337 | 1, failure; merged with it standing | no | 2 h 35 min |
| #60 (kvarn) | 820 | 3, all failure | no | open, a draft |

One of five passed its first review. Three still failed after the one
repair. The two largest never passed without the owner. The waits in
19.1 were these pull requests being open: #42's 6 h 27 min is rf036's
wait.

**Nothing known before a build predicts this.** Files named by tasks: 5,
14, 9, 3, 13. Tasks: 2, 2, 3, 1, 2. Neither follows the failures; added
lines do, and they are known only afterwards. Five is a small sample.

**Tend has a build approved by a label, refused where Grove is.**
`tend <repo> issue-<n>` builds an issue labelled `tend-build` with no
proposal (`attempt.go:152-187`). In a Grove repository it stops: "where
Grove holds the scope, a request goes through a proposal … so the
maintained specs keep up with what is built" (work-items spec, "An issue
in a Grove project").

**Grove builds a drafted change on its own branch** (test 6, run on
`3cfd2ef` with 19.2's `lib.sh`):

```bash
set -uo pipefail; . "$(dirname "$0")/lib.sh"
T=$(mktemp -d "$1/g6.XXXX"); cd "$T"; project "$T/repo"; cd "$T/repo"
git switch -q -c tend/fix-1
change 261126-f0000 fix "Fix" a.txt
sed -i '' 's/^schema: default/schema: bugfix/' grove/changes/261126-f0000-fix/.grove.yaml
git add -A && git commit -qm "draft a fix"
grove check
grove instructions implement --runner 261126-f0000 | sed -n 3p
grove claim 261126-f0000
```

Observed: `check` passes; `instructions implement` exits 0 and begins
"You are implementing drafted change 261126-f0000 … as this working tree
holds it"; `claim` exits 1, "main at … holds no cycle or change
261126-f0000". What forbids an unattended draft being built with its
code is one sentence of constitution rule 4 and the last line of
`requires/propose-unattended.md`, not code.

**Claude's view.**

- The second half of the objection is borne out. One feature-sized
  change is already at the edge of what review and one repair get
  through, so a group of them would rarely pass and one member would
  hold the rest. 19.7's "`claim` takes a set" is withdrawn for such
  changes. Their waits are the earlier pull request failing review and
  waiting for the owner; grouping lengthens both.
- The first half is answered by lanes the owner picks with a label; no
  agent picks one.
  - **Feature lane**, label `tend`, as today: one request, one proposal,
    one build. Never grouped, at intake or at the build.
  - **Fix lane**, a second label: no proposal to approve. One pull
    request holds a bugfix change and its code, and the owner's one
    approval covers both. The agent declines, saying why, when the
    request is not a small fix, and the owner labels it `tend` instead.
    It takes no claim, so a collision with work in flight is caught at
    landing (decision 8).
  - **Grouping exists only in the fix lane and starts off**: one fix,
    one pull request. A count bound is added there if approving small
    pull requests proves to be the cost.
- Cost: an amendment to Grove rule 4 and its unattended instruction, and
  Tend's refusal in a Grove repository lifted for that label. The
  separate issue-build path (section 4, gap 7) can then go, since a fix
  is a Grove change built by the same flow.
- Of 19.5, items 3 and 4 stand. Items 1, 2, 5 and 6 are replaced by the
  above.
- For memex's five issues: three fixes and two that need the owner come
  to seven approvals, against ten today.
- Not solved here, and larger than batching: four of five builds failed
  their first independent review, and the owner judged the reviewer's
  findings by hand on #42 three times. It belongs to thread 8.

### 19.9 Codex's second opinion, and a check on it

The owner, on 19.8: it still does not feel right or solved; the thread is
too bound to existing constraints; get a second opinion from Codex.
[Its reply](thread-4-codex-second-opinion.md) is kept whole. It was asked
an open question with Claude's proposals cut from its copy of this file.
Nothing here is decided.

**What Codex said.**

- Batching is the wrong frame. The problem is when the owner has to
  exercise judgment, and on what: 42 questions on documents before any
  evidence exists, then failing reviews afterwards.
- Units: a request (who wants what outcome), a decision (a choice only
  the owner can make, its evidence, the answer) and a candidate (the
  smallest change the owner could accept or decline alone). Tasks belong
  to the agent and change as it learns.
- The owner sets priorities, a daily budget and boundaries. Inside them
  Tend investigates, prototypes or builds. It asks only where two
  readings give materially different results, with a concrete example,
  at the last point the answer is cheap. The owner approves a concrete
  candidate.
- Review repeats while it makes progress within budget, parks on a
  repeated failure with nothing new, and sends a real product
  disagreement to the owner with the example. At most two items wait on
  the owner; they are shown in one sitting.
- Retire for keyborg: the proposal merge, the file reservation, the one
  repair. One builder to start. One tool; a product promise in plain
  words lands with the first code that honors it.
- Keep: thread 1's decision records and deciding apart from acting,
  thread 2's headless persistent runner, thread 3's merge authority
  outside the runner.
- Store: local SQLite. Beads if independent workers duplicate work or
  lose dependencies; Dolt if work records edited on several machines
  need versioned history and merging.
- Its strongest objection to itself: a working demo anchors the owner,
  and prose is the cheap place to reject a premise.
- Its experiment: one real keyborg request by hand over two days, the
  owner writing down privately beforehand what would make them reject
  it; measure the owner's active minutes, sittings, rework, and what the
  reviewer missed.

**Checked.** Its three code citations hold: one branch per labelled issue
(`intake.go:80-82`), the overlap refusal (Grove `claim.go:219-238`), and
`limit: 1` on repair (`tend/flow.yaml:30`).

**Did the proposal step change anything?** Commits on the six proposal
pull requests (`gh` reads):

| Proposal | Owner's commits before merging | Its build |
| --- | --- | --- |
| #39 | none; merged as drafted in 24 min | #42: 5 failed reviews |
| #40 | none; merged in 31 min | #48: 2 failed, then passed |
| #44 | "settle the open questions", +124 −95 in the proposal, both specs and the tasks | #60: 3 failed, open |
| #53 | "settle the open questions", +54 −57 in the proposal and both specs | #57: passed first time |
| #56 | none; merged in 45 min | #59: 1 failed, merged |
| #55 | none; still open | not built |

The owner changed two of five before merging, and substantially. Three
went through as drafted. So the early checkpoint is neither empty nor
always used, which is the ground Codex's own objection stands on.

**Claude's view.**

- Codex's frame is better than 19.5, 19.7 and 19.8, which rearranged the
  work between two fixed checkpoints. The evidence gathered for them
  points at the checkpoints: questions answered in advance by the agent
  that asks them, proposals merged unread or rewritten, a review loop
  the owner ran by hand on #42. What memex shows is the owner batching
  a sitting, not the work.
- Gaps in it:
  - Stage 2 of thread 3 approves an implementation by comparing it with
    an approved proposal. Without one, the owner approves every
    candidate until something else is the reference.
  - With one builder and two candidates waiting, the next piece of work
    touching the same files either stacks on a waiting candidate, as
    memex's did, or collides. It does not say which.
  - "Ask only at a real fork" is the judgment that marked 29 of 42
    questions consequential. The cap on what waits and on questions per
    request does the work, not the agent's judgment.
  - It gives no reason for SQLite over 14.2's file per run.
- Two things follow whichever way the rest goes: repair bounded by
  progress and budget in place of a count of one, and the owner's
  waiting items shown together while each stays separately rejectable
  (thread 5).
- The rest turns on one untested question: whether this owner decides
  better from an example and a candidate than from a proposal. Codex's
  experiment tests it with no code.

### 19.10 Where thread 4 stands

Nothing in 19.5, 19.7, 19.8 or Codex's reply is adopted.

**What the owner said.**

- "I generally agree batching is the wrong frame."
- The experiment Codex proposed will not be run: one request is one kind
  of work, and a result from it "doesn't hold across many different
  kinds of work".
- On 19.7, the owner's objection was "when everything goes through that
  flow". On 19.8 and on Codex's reply: too focused on existing
  constraints, and not solved.
- The owner does not know how to move forward on this thread for now,
  and parked it: "we're in a rabbit hole right now and I want to get out
  of it".

**Claude's reading, not agreed.** Two of the owner's three objections are
the same one: each proposal, Claude's and Codex's, was a single rule
for all work, or two lanes. The thread designed a flow without first
listing the kinds of work and where the owner's judgment changes the
outcome in each. The record already shows kinds that differ in this: a
bug whose symptom is the whole scope (memex #202), a small change that
turns on one choice (memex #204, #205), a change to Tend's own machinery
with five to nine forks (two of five proposals rewritten by the owner,
19.9), a reshaping of a whole app (memex #189, three changes), a change
to the rules agents work under (thread 3), work the agents find
themselves (memex #202 to #206), and research that ends in no code.

**The question that replaces the frame.** For each kind of work, where
does the owner's judgment change the outcome, and what is it asked on.
It is not an intake question. It is the same question as thread 5 (what
waits on the owner), thread 8 (what must be true before keyborg runs
unattended) and thread 9 (research before a proposal).

**What does not wait on it.** Threads 1 to 3: Codex, asked to treat them
as open, kept all three. And two findings of this thread: repair bounded
by progress and budget in place of a count of one (19.8, 19.9), and the
`requests` reference from a change to what asked for it (19.2 test 3,
14.3).

**A way to continue, offered and not accepted.** Sort the requests and
pull requests the owner already has, in `mascah/tend`, `mascah/memex`
and `mascah/grove`, by kind, and mark for each where the owner stepped
in (a proposal rewritten, changes requested, a pull request closed, a
fix after merge) and where it passed through. That covers many kinds
from real work and asks no new experiment of the owner.
