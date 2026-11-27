# Design review: Claude's first analysis, and corrections

Part of [the Tend and Grove design review](README.md): read that first. It
says what is decided and which file holds which section. Section numbers are
those of the single file these were split from.

From the first session, 2026-11-26. Section 11 is Claude's analysis and is
opinion; 14.5 says where thread 1 departs from it. Section 13 lists what
that session got wrong and corrected.

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
