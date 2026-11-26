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
  approval is read. A failed check gives `failed` while `canAsk` says that
  step can still be asked. Otherwise `spentFix` comments once per head and
  `publish` gives what it would have given without the wait. With no check
  but `tend/review`, or one still unfinished, when the wait ends, tend
  stops for a person.
- `flows/prompts/failed.md`: worded for a failure found at either step.
- Tests: `publication_test.go` ci-red, ci-pass, ci-failed-review,
  ci-requested-first, ci-approved-red, ci-none, ci-unfinished, ci-no-flow,
  ci-still-red, ci-one-fix, ci-rerun; `flow_test.go` flow-publish-failed;
  `watch_test.go` watch-at-once-checks; `landing_test.go` updated for the
  prompt's new wording. `go test ./...` passed. Breaking the `tend/review`
  exclusion made ci-failed-review fail, so that test catches it.

Check: Main changed `publication.go`, `landing.go`, `flow.go`, `landing_test.go`, `flow_test.go` and `testdata/bin/gh` after this change's base, in the protected-path change 261125-6dvbs (`hold`, `approvedOn`, `protect`, `DURING_CHECKS`). The work builds on those files as they are now and changes none of that behavior; `testdata/bin/gh` needed no change, since `FIX_ALSO` in `testdata/bin/turn.sh` lets a test's fix make the checks pass.
Check: When a step can still be asked (open question 8): it can when it has no limit or is under it, and also when it was asked at this very head and is still working on it. Without that last part, a rerun during the fix would read the fix as spent.
Check: The spent-fix comment is found on GitHub by a `<!-- tend:checks-failed <head> -->` marker, and a `checks-said-<head>` file in the run folder skips the lookup. Its wording names the head, the failed checks with their links, and the step's limit.
Check: The status while waiting is "waiting for the checks on <branch> at <head>". `publish` does not apply the protected-path hold before sending a failure to fix (`land` does). The change does not ask for that, and in a supervised flow a person approves the head anyway.
Check: `prompts/failed.md` now starts "Checks GitHub ran on {branch} at {head} failed, so tend neither asks a person to approve it nor lands it". A project's own copy of the old wording still works, since it is only a prompt.

### T-02 — tend's own flow and its documentation

Status: not started

## Scenarios

## Reconciliation
