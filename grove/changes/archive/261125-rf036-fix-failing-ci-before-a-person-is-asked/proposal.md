# Fix failing CI before a person is asked

## Problem

In a supervised flow a pull request whose CI fails still waits for the
owner's approval, and nobody fixes it in the meantime
([issue #37](https://github.com/mascah/tend/issues/37)).

- tend reads checks only at `land`: `checksOn` and `checked` in `landing.go`
  at 2a8ae80 are called only from `landStep`.
- `publishStep` in `publication.go` at 2a8ae80 publishes the head, then
  gives `requested`, `approved` or `ok` without looking at checks. A
  project's CI that runs on `pull_request` reports its checks only once the
  pull request is open, so nothing before `publish` could see them either.
- tend's own flow (`tend/flow.yaml` at 2a8ae80) is supervised:
  `publish: {ok: end, approved: land, requested: request}` and
  `land: {ok: end, conflict: resolve, failed: fix}`, with `fix` limited to
  one ask. The attempt ends at `publish` and waits for a person. The owner
  looks at a red pull request and approves it or fixes it by hand; only then
  does `land` see the failure and send it to `fix`, whose new head is
  reviewed again and needs a new approval. The issue's words: "the owner is
  asked twice, once about a head that was never going to land."

## Intended outcome

In a flow whose `publish` names a step for a new outcome `failed`, tend
waits after publishing for the checks on the head, reading them as `land`
does, before anything asks a person. A failed check goes to that step, which
in tend's own flow is the same `fix` step `land` uses, prompted with the
failed checks' names and links (`{checks}`, `prompts/failed.md`); its new
head is reviewed and published again. The owner is asked to approve only a
head whose checks passed, or one whose checks failed after that one fix was
spent, and then the pull request says so. A flow that does not name
`failed` on `publish` behaves as today.

## Scope

### In scope

- A new outcome `failed` of the build flow's `publish`, given when a check
  on the published head failed and the step it leads to can still be asked.
- Waiting at `publish` for the checks on the head, with landing's reading of
  checks, `tend/review` not counted as CI and not counted as a check at all
  there.
- A comment on the pull request, once per head, when the checks failed and
  the step `failed` leads to is spent.
- `prompts/failed.md` worded for a failure found at `publish` as well as at
  `land`.
- tend's own flow file sending `publish`'s `failed` to `fix`.
- README and guide pages that describe `publish`'s outcomes and landing.

### Out of scope

- The built-in flows: they have no `fix` step and no `land`, and stay as
  they are.
- Running a failed check again, or telling a flaky check from a real failure.
- How long a wait for a first check should be before tend says a project has
  no CI: landing's own open question, shared here, not answered here.
- The protected-path check of issue #36, built separately and on main since 98f4515: `land` keeps its hold, and `publish` adds none.

## Decisions

None was made beyond the request. The request asks that tend waits for the
checks after `publish` and before a person is asked; that a failed check goes
to the implementer once, with "the existing `{checks}` and
`prompts/failed.md`"; that the new head is reviewed and published again; that
the owner is asked only about a head whose checks passed "or one where that
single fix was spent and still failed, with the pull request saying so"; that
it works in the autonomous flow too or the proposal says why it is not
needed; and that it "reuses landing's check reading. `tend/review` alone
still does not count as CI."

## Open questions

The first three are the request's own open points; this draft is written to
the recommended answers, and each can wait for review unless marked.

1. **Where the wait lives.** Recommended: a new outcome `failed` of
   `publish`, which a flow places by naming a step for it; `publish` waits
   for checks only in a flow that does. Reason: a flow file already decides
   where landing and fixing go, a project without CI on pull requests is not
   made to wait, and no second tend-owned step is added beside `publish` and
   `land`. The alternatives were a step of its own (one more reserved name
   and its validity rules) or always waiting inside `publish` (changes every
   flow, including those with no CI). Decides T-01.
2. **Whether waiting for CI holds a slot under `watch: {at_once: N}`.**
   Recommended: yes. tend stays running and holds the work item's lock while
   it waits, so the attempt counts as being worked on, as the direction's
   decision 19 counts an attempt "building, in review, repairing or in the
   train"; it is not yet waiting for a person, which is what frees a slot.
   Reason: no new state for `tend watch` to track, and a tend killed while
   waiting is run again as an attempt whose tend died, as today. The cost is
   a slot held for as long as CI runs. Decides T-01, and the watch-and-status
   change spec.
3. **How it shares the single fix with `land`'s `failed`.** Recommended: a
   flow sends both to the same `fix` step, whose limit of 1 counts asks from
   either, so an attempt gets one fix for failed checks in all. When
   `publish` sees failed checks and that step is spent, `publish` comments on
   the pull request naming the failed checks and gives what it would have
   given without the wait (`ok` or `approved`); `land` reaching a spent `fix`
   later stops for a person as it does today. Reason: matches the request's
   "goes to the implementer once", and `publish` going to a spent step
   instead would bring the flow back to `publish` at the same head, which
   tend stops as a loop. Decides T-01.
4. **Checks still unfinished, or none but `tend/review`, when the wait
   ends.** Recommended: stop for a person as `land` does, within the same
   `TIMEOUT_MS` bound, naming the head and the checks. Reason: a flow that
   names `failed` has said it has CI; asking a person to approve a head
   whose checks never finished is what the request avoids. The cost is a
   wait of up to four hours by default for a project with no CI that named
   `failed` anyway. Decides T-01.
5. **Order with a person's requested change and approval.** Recommended: a
   change a person requested on the head is acted on first, without waiting
   for checks, since it makes a new head anyway; checks are waited for before
   an approval is read, and a failed check gives `failed` even where the head
   is approved, as the new head needs a new approval. Decides T-01.
6. **A head that failed review.** Recommended: its checks are waited for and
   fixed like any other head's, since it is published and a person is asked
   about it too; `tend/review`'s `failure` on it is not read as a failed
   check. Decides T-01.
7. **The autonomous flow.** Recommended: allowed but not needed. There
   `publish` sends `ok` to `land`, which already waits for checks and sends
   `failed` to `fix`. Naming `failed` on `publish` there only moves the first
   wait out of the landing lock, so one candidate's first CI run does not
   hold up another's landing. The README says so; tend's own flow is
   supervised, so no autonomous flow in the repository changes. Decides T-02.
8. **Routine choices inside the approach.** Recommended for the implementer
   to settle and record: the run-folder mark that the spent-fix comment was
   posted for a head, the wording of that comment and of `status` while
   waiting, how `checked` is split so `publish` and `land` share it, and the
   tests' names and places. Decides T-01.

## Related documents

- Requirements: [publication](specs/publication/spec.md),
  [watch-and-status](specs/watch-and-status/spec.md)
- Tasks: [tasks.md](tasks.md)
