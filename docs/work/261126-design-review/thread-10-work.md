# Design review, thread 10: work as a first-class thing

Part of [the Tend and Grove design review](README.md): read that first. It
says what is decided and which file holds which section.

From a session of 2026-11-27, held in the Grove repository: what the owner
raised, the evidence gathered, and Claude's view. Nothing here is decided.

## 20. Thread 10, work as a first-class thing: opened 2026-11-27

### 20.1 What the owner said

- "i also want to consider what it would look like to move the tasks
  aspect out of grove and into a new project that is equivalent to beads.
  I'm open to trying dolt. If tend is the orchestrator, grove is what is
  agreed, maybe we can gain a lot by treating what needs to be done as a
  first class thing. So Tend + Grove + (To be named)."
- Asked to "explore what we could gain from this and the challenges we
  face", and then to record it as a thread: "yes".
- It came up after a review of how Grove's specs grow over time. That
  review produced two Grove pull requests, mascah/grove#205 (trim the
  archive) and mascah/grove#206 (a size limit in `grove check`). Neither
  depends on this thread.

### 20.2 What "tasks" is in Grove today

Read and counted by Claude on Grove `3cfd2ef`. "Tasks" is five things, and
one of them is agreement.

| Today in Grove | Kind | Would live in |
| --- | --- | --- |
| `tasks.md`: the breakdown, the files a change may touch, acceptance | Agreement: approved with the change | Grove |
| `implementation.md`: status, blocked questions, notes for the reviewer, work found along the way | Work | The new tool |
| `depends_on` and cycles | Work: ordering | The new tool |
| `claim` and the branch `grove/runs` | Work: who has what | The new tool |
| `serve`'s readiness and progress views | Work | The new tool or Tend |
| Finished or dropped: the archive folder | A fact of the merge | Read from Git, stored nowhere |

How much of Grove this is:

- The cycles, runs and serve specs are 12,347 of 30,463 spec words (41%).
  With implement, 17,449 (57%).
- Constitution rules 2, 3 and 10 are 604 of the 1,227 words in the
  product-boundary rules 1 to 11. Rule 10 says Grove "is not a workflow
  engine" at the end of a paragraph of gates.
- `claim.go` and `serve.go` are 830 of 4,842 non-test Go lines. The cycle
  and task logic inside `doc.go` and `check.go` was not separated out.
- `tasks.md` and `implementation.md` are 36% of the 202,287 words in the
  archive.

Tend already has work items with no record of their own: a GitHub issue,
a Jira issue, and a Grove cycle or loose change, each with its identity
computed its own way (`grove/specs/work-items`, `intake`, `attempts`).

### 20.3 What would be gained

1. **One record per unit of work.** Section 6: one unit has an issue, a
   proposal branch and pull request, a change ID, an attempt ID, an
   attempt branch and pull request, a run folder and agent names, and
   "Nothing joins them." 14.3 planned the join through the change ID and
   a request reference in `.grove.yaml`.
2. **A smaller Grove.** Two fifths of its spec text and half the words of
   its boundary rules are about scheduling and claiming work, not about
   what is agreed.
3. **The queue as a query.** 14.4 already made a decision its own record;
   thread 5 needs the list of open ones.
4. **A home for kinds of work.** 19.10: each proposal in thread 4 was one
   rule for all work. Work the agents find themselves and research that
   ends in no code do not fit "a change with tasks". This gives the
   question a place; it does not answer it.
5. **Finished work leaves the tree.** Closed work is filtered by a query
   instead of sitting in Markdown that agents search.
6. **Partial progress.** 19.10's repair bounded by progress and budget
   needs progress recorded per attempt.

### 20.4 What stands in the way

1. **Two answers to "done".** Today the code, the folded spec and the
   archived change land in one merge, so they cannot disagree. A store
   outside Git can say done when the branch never landed. Section 8: Gas
   Town's docs name "tasks don't make it through the refinery and land on
   the target branch" as the most common complaint. Beads users lost
   issue changes in the pull request when it moved to Dolt
   (gastownhall/beads#2489, delegated).
2. **No authority.** The bot's credential can write the store, so nothing
   may gate on it (thread 3). Approval stays the merge and the approving
   review; "ready" in the store is advice.
3. **`tasks.md` does two jobs.** Its `Files:` lists are approved scope:
   they fence an unattended build, `review` marks files against them, and
   `claim` reads them for overlap (Grove constitution rules 3 and 10).
   Status is work. The cut runs through the file, not around it.
4. **Status rides the branch today.** Abandon or revert the branch and the
   status goes with it. In a store it must be keyed by attempt and
   released: the release record of 14.4.
5. **Notes for the reviewer are in the pull request diff today.** Out of
   Git, Tend has to copy them into the pull request.
6. **Grove alone must still work** (thread 6). `grove archive` refuses a
   change unless every task says done in `implementation.md`
   (`archive.go:30-33`, checked).
7. **Requests already live in two outside trackers.** A new store is a
   third place for them unless it only refers to them.
8. **Size.** Beads has about 115 command pages and 1,063 open issues
   (delegated); Gas City is about 996,000 lines of Go (section 7). The
   part that is needed is items, edges, a ready query, claim and release,
   and decisions.

### 20.5 Stores compared: facts and how each was checked

"Delegated" means a read-only pass over primary sources on 2026-11-27,
whose citations Claude did not re-read. "Checked" was read or run by
Claude.

**Records on a Git ref**, which is what `grove/runs` is.

| Fact | Source |
| --- | --- |
| A push to the ref is compare-and-swap, so a claim is atomic across machines | Checked: Grove `claim.go:129` pushes with `--force-with-lease`; 11.3 calls it "the one part of the stack that already works from an ephemeral runner" |
| No queries beyond reading the records | By construction |

**Dolt**, embedded in a Go program.

| Fact | Source |
| --- | --- |
| Needs cgo; no pure-Go build | Delegated: `dolthub/dolt` v2.7.7 storage imports `dolthub/gozstd`; Beads's `engdocs/ICU-POLICY.md` |
| `bd` 1.3.4 for darwin-arm64 is 144.5 MB | Delegated: gastownhall/beads release v1.3.4 |
| About 110 ms to open a 2.3 MB database; about 1.8 s per `bd show` on a 482 MB store | Delegated: gastownhall/beads#5397, one user's measurement |
| One exclusive lock per database: a second process opens read-only without saying so, or is locked out | Delegated: `go/store/nbs/journal.go` at v2.7.7 |
| Several writer processes on one machine, or any outside SQL client, need `dolt sql-server` running | Delegated: `dolthub/driver` README; DoltHub blog 2026-07-22 |
| A Git remote can hold the data under `refs/dolt/data`, since 2026-02-13. A plain `git clone` does not bring it, and `git push --mirror` from such a clone deletes it | Delegated: DoltHub blog 2026-02-13; gastownhall/beads#5266 |
| A 56 MB pull over that path took 40 minutes | Delegated: dolthub/dolt#11916, closed |
| Merge is per cell; conflicts go to a table and nothing commits until resolved. Dolt picks no winner | Delegated: DoltHub docs, "merges" |
| A claim is atomic within one database only. Two hosts each claimed all 20 items before syncing; the pull settled it last-write-wins | Delegated: gastownhall/beads#8282 |
| Migrations between tool versions are the tool's to build. Beads's accidental 1.2.x release migrated schemas on any command | Delegated: beads.gascity.com, "accidental-1-2-1-release" |
| `AS OF` and diff tables answer "what changed since yesterday" in one query | Delegated: DoltHub docs, system tables |

**Beads** itself, 1.3.4.

| Fact | Source |
| --- | --- |
| A work tracker by charter. Its author: "Beads are intentionally lightweight for tracking execution, not storing full specifications" | Delegated: `engdocs/PROJECT_CHARTER.md`; gastownhall/beads#976 |
| Edges that affect readiness (`blocks`, `parent-child`, `waits-for`) and ones that do not (`discovered-from`, `supersedes`); `bd ready --claim` | Delegated: `docs/core-concepts/index.md` |
| Session start injects instructions and no issue list; finished work stays out because agents ask for ready work | Delegated: `cmd/bd/prime.go` |
| Last-write-wins is Beads's own merge policy; two open bugs under it drop a comment and reopen a closed item | Delegated: gastownhall/beads#7757, #7783 |

**GitHub Issues**, on private repositories of a personal account.

| Fact | Source |
| --- | --- |
| Sub-issues (100 per parent, 8 levels) and blocked-by edges (50 per issue), with REST, GraphQL and `gh` flags | Delegated: docs.github.com; `gh issue edit` manual |
| Issue types and fields reached personal accounts on 2026-11-11 | Delegated: GitHub changelog |
| Ready work is one list call: each issue carries its count of open blockers | Delegated: REST issues reference |
| No atomic claim: adding an assignee never fails a second writer, and writes take no condition | Delegated: REST assignees reference; REST best practices |
| GitHub's own pattern page says issues "are not queue-storage backends" | Delegated: github.github.com/gh-aw |

Not determined: Dolt server's idle memory, and any size limit on Dolt data
held in a GitHub repository.

### 20.6 Claude's view (opinion; nothing here is agreed)

1. **The split is worth more than the store.** Gains 1 to 6 come from
   giving work one model and one contract. None of them needs SQL at tens
   of open items.
2. **Store only what nothing else can answer.** That is 14.2's rule for
   Tend's record, and it is the guard against two answers to "done":
   landed and approved are read from Git each time.
3. **Cut through `tasks.md`.** The file fence and the acceptance stay in
   Grove, approved with the change. Status, blocked questions, ordering,
   cycles and claims move.
4. **Start inside Tend's rewrite.** Build the model as its own package and
   command set, with records on a Git ref as `grove/runs` keeps them.
   Lift it into a third project when something other than Tend calls it.
5. **Not Dolt first.** For one owner and one to three machines it adds
   cgo, a binary over 100 MB, a server for concurrent writers and a
   migration system to write, and it still leaves cross-machine claims to
   a compare-and-swap push. To feel it before deciding, run Beads itself
   on one repository with Tend as the only writer; that costs no code.
6. **Requests stay where they are.** GitHub and Jira issues are referred
   to by ID, not copied.

### 20.7 What it would touch in what is decided

- **14.2, decision 1** ("No Grove replacement, no shared store"). Grove
  stays. The reason given there, approval as a merge, is untouched so
  long as the store holds no scope and no authority. The store question
  reopens for work only.
- **14.2's contract and 14.4's four points.** A machine claim that writes
  the ledger only, typed results, the release record, the decision record
  and the caller's attempt ID would be built in the new tool. Grove would
  lose `claim`, cycles and the gates of constitution rule 10 that rest on
  them, in place of amending rules 2 and 4. Reads that take a commit stay
  Grove's.
- **Thread 3.** Unaffected if nothing gates on the store.
- **Threads 4, 5 and 6.** See 20.3 points 3 and 4, and 20.4 point 6.

It bears on timing: the rewrite of 14.2 builds the ledger and decision
records that would move, so this is cheaper to settle before it than
after.

### 20.8 Open questions

1. Is this a third project, or a package and commands inside Tend until a
   second caller exists?
2. Which facts does it store? Proposed test: only what Git, GitHub, Grove
   and the agent cannot answer.
3. Does the task breakdown stay approved with the change, with only
   status moving?
4. How does Grove alone know a change is finished once status has left
   `implementation.md`?
5. Do requests stay in GitHub and Jira, referred to by ID?
6. Which store, and what need would justify Dolt: several machines
   writing at once, queries across repositories, or SQL dashboards?
7. Before, with or after Tend's rewrite?
8. Is the Beads trial of 20.6 point 5 worth a week?
