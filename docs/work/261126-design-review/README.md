# Tend and Grove design review

A design discussion between the owner and Claude, begun 2026-11-26 and
worked through in numbered threads. This file is the index: where each
thread stands, what is open, and which file holds what. It decides nothing
itself. Each line under "Where each thread stands" points at the section
that records it, and that section is the authority.

Read with: [the Codex architecture review](../261125-architecture-review.md)
and [the autonomous direction](../261121-autonomous-direction/direction.md).

## How to use these files

- To continue a thread, read this file and that thread's file. Open
  another file only for a section a reference points at.
- A reference such as "14.2" or "section 4, gap 5" is a section number.
  The numbers are those of the single file this folder was split from on
  2026-11-26, text unchanged, so every reference still resolves. The table
  says which file holds which numbers.
- A new thread gets its own file and the next section number (21 for the
  next; thread 10 took 20).
- A decision is recorded in its thread's file, with its evidence, and as
  one line under "Where each thread stands" here.

| File | Sections | Holds |
| --- | --- | --- |
| This file | 12, 16, 17 | What the owner has said; open facts; how to continue |
| [evidence.md](evidence.md) | Summary, 1 to 10 | What was examined; the Codex review's 12 findings, checked; reproductions; gaps the review does not cover; Tend's state and control; workflow numbers; Gas City, Gas Town and Claude Approvals; the memex case |
| [analysis.md](analysis.md) | 11, 13 | Claude's analysis from the first session, which is opinion; that session's corrections |
| [thread-1-state-model.md](thread-1-state-model.md) | 14 | Thread 1, the state model |
| [thread-2-runtime.md](thread-2-runtime.md) | 15 | Thread 2, runtime environments |
| [thread-3-authorization.md](thread-3-authorization.md) | 18 | Thread 3, authorization, with the GitHub sandbox results |
| [thread-4-intake.md](thread-4-intake.md) | 19 | Thread 4, intake batching |
| [thread-4-codex-second-opinion.md](thread-4-codex-second-opinion.md) | | Codex's reply on thread 4, kept whole; 19.9 checks it |
| [thread-10-work.md](thread-10-work.md) | 20 | Thread 10, work as a first-class thing, with the stores compared |

## Where each thread stands

**Thread 1, the state model: decided** (14.2, amended by 14.4)

- Grove stays, with scope in Git. No Grove replacement, no shared store.
- The contract between Grove and Tend changes. Reads take a commit. A
  machine claim writes the ledger only and prints the record as JSON: no
  branch made, no ref moved. Results are typed: claimed, refused with a
  reason, failed. Needs amendments to Grove constitution rules 2 and 4.
- Tend keeps one record per run of a flow, build or propose: a local JSON
  file with one writer, replaced atomically, holding past-tense facts only
  Tend knows. Anything Git, GitHub, Grove or the agent can answer is
  observed each pass, never stored.
- The loop is observe, decide, act. `decide` is a pure function returning
  one of a closed set of actions. A failed observation is a retry, never a
  human wait.
- Invariant: after publish the record can be rebuilt from the branch, the
  pull request and the ledger, and a test proves it.
- Adopted from Codex (14.4): the ledger gets a release record naming the
  attempt; a decision is its own record, and the queue is the open
  decisions; the caller supplies the attempt ID; overlap is read at each
  in-flight attempt's own claim commit, and the text sent in each ask is
  kept beside the record.
- 14.3 sketches the record and predates those four.

**Thread 2, runtime environments: decided** (15.1)

- One runtime class: a machine whose disk persists (laptop, the mini, or a
  Linux VM). Hosted CI runs gates, not agents. Anthropic cloud sessions
  are parked.
- Agents run headless, and Tend's correctness needs no multiplexer. A
  multiplexer stays as an optional view, designed in thread 5.

**Thread 3, authorization: decided in its main lines** (18.6; 18.7 lists
what is left)

- Tend holds no authorization.
- The runner holds one credential, the bot's, without `workflow` scope.
- The authorizing act is a required approving review from someone who did
  not push the head, with `check` pinned to GitHub Actions. `tend/review`
  is no longer required.
- Stage 1, where the owner is the only approver, comes first; 18.6 gives
  its ruleset and settings. Stage 2, the hosted gate, comes after Stage 1
  has run on keyborg.
- `land` is two acts: enable auto-merge, and merge the target into the
  candidate at the front when it is behind.
- No GitHub App for now. The owner stays on the bypass list.
- The changes to Tend belong to the rewrite of 14.2, not to a patch.

**Thread 4, intake batching: parked, nothing adopted** (start at 19.10)

- The owner agrees batching is the wrong frame, and parked the thread.
- 19.10 gives the question that replaced the frame, and two findings that
  do not wait on it: repair bounded by progress and budget in place of a
  count of one, and the `requests` reference from a change to what asked
  for it.

**Threads 5 to 9: not started.** Section 17 lists them.

**Thread 10, work as a first-class thing: opened 2026-11-27, nothing
decided** (start at 20.1)

- The owner raised moving "what needs to be done" out of Grove into a
  third tool beside Tend and Grove, in the manner of Beads, and is open
  to Dolt (20.1).
- 20.2 to 20.5 hold the evidence: what "tasks" is in Grove today, what
  would be gained, what stands in the way, and the stores compared.
- 20.6 is Claude's view, which is opinion. 20.8 lists the open questions.
- 20.7 says what it would touch in what is decided: if adopted it amends
  14.2 and moves what 14.4 adds to the ledger, which is why its timing
  depends on Tend's rewrite.

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
(18.6); 18.7 lists what is left of it. Thread 4 is open and its
frame is set aside: 19.10 says what replaced it and what stands.
Remaining, in order:

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
10. Work as a first-class thing: whether "what needs to be done" leaves
    Grove for a third tool, and where it is stored. Raised by the owner
    on 2026-11-27 and opened the same day. Its place in this order is
    not settled: 20.7 says why it bears on the rewrite of 14.2.

Thread 3 continues from 18.7, thread 4 from 19.10, thread 10 from 20.8.
What to read first is at the top of this file.
