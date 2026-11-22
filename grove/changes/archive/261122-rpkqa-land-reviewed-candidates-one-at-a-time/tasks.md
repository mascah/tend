## T-01 — A flow can hold land

Result: `land` is a step tend knows as its own, with the outcomes `ok`,
`conflict` and `failed`; a flow that misuses it is refused; and tend ships
the prompts for the steps a conflict and a failed check lead to, as the flows
requirement "Landing is tend's own and a flow places it" says.

Files:
- `flow.go`
- `flows/prompts/`
- `flow_test.go`

Delegated: test layout and private helpers; the names of the two prompt
files and their wording.

Interfaces: the step `land` and its outcomes; the outcome `approved` of
`publish`; what a prompt after `failed` is filled in with, the failed checks'
names and links.

Constraints: the built-in flows hold no `land`; `tend flow` in a repository
with no flow file prints what it printed before.

Acceptance:
- The scenarios "Land without publish", "Land given a prompt", "A conflict
  step with no limit" and "A flow names tend's own prompts" are tests that
  pass.
- A loop that comes round only through `publish`'s `approved` is valid with
  no limit.
- `TestBuiltInFlowsAreValid` still passes.

## T-02 — tend lands a candidate

Result: at `land`, tend lands one reviewed candidate at a time: the target
merged in when it moved, the checks waited for, and the pull request
squash-merged at exactly that head, as the landing requirements "Land only a
head that passed review", "One landing at a time", "The target is merged in",
"Checks decide" and "Squash-merge exactly that head" say.

Files:
- `main.go`
- `flow.go`
- `testdata/bin/gh`
- `tend_test.go`

Delegated: test layout and private helpers; how tend reads a head's checks
from GitHub; the names of files under TEND_STATE; the wording of
one-sentence messages.

Constraints: nothing is pushed by force; every write to GitHub is the
publishing account's; a conflict leaves the worktree with no merge under
way.

Acceptance:
- Each scenario of those five requirements, and the flows scenarios "A build
  that lands" and "A flow without land", is a test that passes.
- With no flow file, every test of publication passes unchanged.

## T-03 — A clean merge of the target keeps the verdict

Result: a head `land` made by merging the target cleanly is held as carried:
it gets `tend/review` as carried, no review, no reviewer and no new commit
message, as the landing requirement "The verdict carries across a clean
merge", the review requirement "A fresh reviewer for each head" and the
publication requirement "The description follows later heads until a person
edits it" in this change say.

Files:
- `main.go`
- `flow.go`
- `tend_test.go`

Delegated: test layout and private helpers; how a carried head is recorded
in the run folder.

Constraints: whether a head is carried is read from Git, never from a mark
an agent could write.

Acceptance:
- The landing scenarios "A carried head", "A rerun after a kill" and "A
  merge with an edit of its own", the review scenario "A carried head" and
  the publication scenario "A carried head" are tests that pass.

## T-04 — An approval sends a candidate on

Result: in a flow whose `publish` names `approved`, an approving review by a
person with write access on the head gives that outcome, and `tend watch`
runs the attempt again when it sees one, as the landing requirement "A
person's approval is an outcome of publishing" and the watch-and-status
requirement "Watch runs an approved attempt again" say.

Files:
- `main.go`
- `flow.go`
- `testdata/bin/gh`
- `tend_test.go`

Delegated: test layout and private helpers; the wording of one-sentence
messages.

Constraints: who has write access is decided as for a change request, by
the one rule change 261122-fh9h2 adds.

Acceptance:
- Each scenario of those two requirements is a test that passes.

## T-05 — The README and the publication spec say when tend merges

Result: the README shows how a project opts into landing, and the
publication spec no longer says tend never merges.

Files:
- `README.md`
- `grove/specs/publication/spec.md`

Acceptance:
- The README's flows section shows a supervised flow file, where `publish`
  sends `approved` to `land`, and an autonomous one, where it sends `ok`
  there, each with limited steps for `conflict` and `failed`.
- It says to set a branch rule on the target that requires the project's CI
  and `tend/review` and requires branches to be up to date, that the
  publishing account must be allowed to merge and must not be able to
  bypass the rule, so that for an account that administers the repository
  "Do not allow bypassing the above settings" is on, and that a rule
  dismissing approvals on new commits undoes an approval whenever tend
  merges the target in.
- It says tend reads an approval once, at `publish`, and that a supervised
  project which wants an approval withdrawn, or changes requested, while the
  candidate waits to hold the merge sets the rule to require an approving
  review, so that GitHub refuses and tend stops with its reason.
- It says one candidate lands for each CI run, and that a project with no
  CI on pull requests cannot land.
- Once the change is archived, the publication spec's Boundaries says tend
  merges a pull request only at a `land` step, as the landing spec says.
