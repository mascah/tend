# Implementation: Fix failing CI before a person is asked

<!-- Progress on change 261125-rf036, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — publish waits for checks and sends a failure to the implementer once

Status: done

- `flow.go`: the build flow's `publish` has the outcome `failed`.
- `landing.go`: `checked` is split into `readChecks` (with `ciOnly`, which
  leaves `tend/review` out entirely) and `standing`; `land` reads checks
  as before.
- `publication.go`: in a flow whose `publish` gives a step for `failed`,
  `publishStep` acts on a change a person requested first, then
  `waitChecks` waits for the head's checks within `TIMEOUT_MS`, before an
  approval is read. Nothing is decided until every check has finished, and
  a read that ends after the wait decides nothing: tend stops for a person.
  Once all have finished, a failed check gives `failed` while `canAsk` says that
  step can still be asked. Otherwise `spentFix` comments once per head and
  `publish` gives what it would have given without the wait. With no check
  but `tend/review`, or one still unfinished, when the wait ends, tend
  stops for a person.
- `flows/prompts/failed.md`: worded for a failure found at either step.
- Tests: `publication_test.go` ci-red, ci-pass, ci-failed-review,
  ci-requested-first, ci-approved-red, ci-none, ci-unfinished, ci-mixed,
  ci-past-deadline, ci-no-flow,
  ci-still-red, ci-one-fix, ci-rerun; `flow_test.go` flow-publish-failed;
  `watch_test.go` watch-at-once-checks; `landing_test.go` updated for the
  prompt's new wording. `go test ./...` passed. Breaking the `tend/review`
  exclusion made ci-failed-review fail, so that test catches it.

Check: Main changed `publication.go`, `landing.go`, `flow.go`, `landing_test.go`, `flow_test.go` and `testdata/bin/gh` after this change's base, in the protected-path change 261125-6dvbs (`hold`, `approvedOn`, `protect`, `DURING_CHECKS`). The work builds on those files as they are now and changes none of that behavior; `testdata/bin/gh` needed no change, since `FIX_ALSO` in `testdata/bin/turn.sh` lets a test's fix make the checks pass.
Check: When a step can still be asked (open question 8): it can when it has no limit or is under it, and also when it was asked at this very head and is still working on it. Without that last part, a rerun during the fix would read the fix as spent.
Check: The spent-fix comment is found on GitHub by a `<!-- tend:checks-failed <head> -->` marker, and a `checks-said-<head>` file in the run folder skips the lookup. Its wording names the head, the failed checks with their links, and the step's limit.
Check: The status while waiting is "waiting for the checks on <branch> at <head>". `publish` does not apply the protected-path hold before sending a failure to fix (`land` does). The change does not ask for that, and in a supervised flow a person approves the head anyway.
Check: Before archiving, Grove's `check` called the change stale because files its tasks name changed on main. It was re-based on this branch through propose: `base` moved to 98f4515, and the proposal's out-of-scope line on the protected-path check now says that check is on main, that `land` keeps its hold and that `publish` adds none. No task or requirement changed, so this task stays done.
Check: An independent review of 312d846 found two defects, both fixed. First, a failed check decided before an unfinished one was waited for. `standing` in `landing.go` now waits until every check has finished, which also changes `land`, as the landing spec's "Checks decide" asks ("until every check ... has finished"). Second, `publish` checked its deadline before reading GitHub, so a pass read after the deadline was taken. `waitChecks` now checks the deadline after each read. If that read ended past the deadline, tend stops for a person, naming what the last read within the wait saw unfinished. New tests: ci-mixed and ci-past-deadline (`testdata/bin/gh` gained `CHECKS_SLEEP`). Without the `standing` fix, ci-mixed fails.
Found: `land` has the same deadline gap: `landStep` in `landing.go` passes `time.Now().After(until)` to `checked` before the read, so a pass read after `TIMEOUT_MS` still lands. Seen while fixing the same defect in `publish`. Left as it is: the landing spec owns it, and no task of this change asks for it.
Check: `prompts/failed.md` now starts "Checks GitHub ran on {branch} at {head} failed, so tend neither asks a person to approve it nor lands it". A project's own copy of the old wording still works, since it is only a prompt.

### T-02 — tend's own flow and its documentation

Status: done

- `tend/flow.yaml`: `publish` gives `failed: fix`, and the header comment
  says so.
- `README.md`: `publish`'s outcomes include `failed`. The supervised example
  names `failed: fix`, with how it shares the one `fix` with `land`. The
  autonomous example says such a flow may name `failed` to keep the first
  CI run out of the landing lock, and need not, since `land` waits for checks.
- `guide/flows-reference.md`: the outcomes table, `{checks}` and the
  `prompts/failed.md` text. `guide/concepts/publication.md`: the outcome
  `failed`, the wait, the spent-fix comment, the `at_once` slot, and the
  sequence diagram. `guide/concepts/landing.md`: the shared fix.
- The file is valid: parsed by the flow checker, it has no fault and
  `publish` gives `approved: land, failed: fix, ok: end, requested: request`.

Also changed: `guide/setup.md` — it holds the guide's supervised flow example, which the acceptance says names `failed: fix` on `publish`.
Check: `tend flow` on this repository reads the flow file from origin's main, so until this branch merges it prints main's `publish` without `failed`. I could not push the file to a scratch origin to run `tend flow` on it (that command was refused). Instead, a throwaway test ran the flow checker (`faultOf`) on the file in this working tree: no fault, and `publish` gives `failed: fix`. The test was deleted afterwards. The test flow-publish-failed shows `tend flow` printing a flow with `failed: fix` on `publish`.

## Scenarios

- publication "A red pull request in a supervised flow": publication_test.go TestTend/ci-red, passed
- publication "The checks pass": publication_test.go TestTend/ci-pass, passed
- publication "Failed review is not a failed check": publication_test.go TestTend/ci-failed-review, passed
- publication "A change requested first": publication_test.go TestTend/ci-requested-first, passed
- publication "Approved but red": publication_test.go TestTend/ci-approved-red, passed
- publication "No CI": publication_test.go TestTend/ci-none, passed (and ci-unfinished, ci-mixed and ci-past-deadline for checks not all finished within the wait, passed)
- publication "A flow without failed": publication_test.go TestTend/ci-no-flow, passed
- publication "Still red after the fix": publication_test.go TestTend/ci-still-red, passed
- publication "One fix in all": publication_test.go TestTend/ci-one-fix, passed
- publication "A rerun": publication_test.go TestTend/ci-rerun, passed
- watch-and-status "Waiting for checks counts": watch_test.go TestTend/watch-at-once-checks, passed
- watch-and-status, the requirement's other scenarios (unchanged): watch_test.go TestTend/watch-at-once-*, passed

## Reconciliation

- publication: applied to grove/specs/publication/spec.md
- watch-and-status: applied to grove/specs/watch-and-status/spec.md
