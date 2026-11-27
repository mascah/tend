# Design review, thread 1: the state model

Part of [the Tend and Grove design review](README.md): read that first. It
says what is decided and which file holds which section. Section numbers are
those of the single file these were split from.

From the second session, 2026-11-26: what the owner decided and the
evidence behind it.

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
