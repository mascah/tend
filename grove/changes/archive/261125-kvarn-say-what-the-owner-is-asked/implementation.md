# Implementation: Say what the owner is asked to do in each notification

<!-- Progress on change 261125-kvarn, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — An attempt's notifications say what the owner is asked to do

Status: done

What changed:

- `terminal.go`: the one function every notification goes through, `tell`,
  given the file that records it, the title and the body. It skips a
  notification identical to the one recorded there, and records each one it
  sends. `fit` applies the length bound: a title of at most 60
  characters, a body of at most 240, whitespace collapsed, the link whole on
  the body's last line, the text before it cut with `…`. `waits` builds a
  wait's title, `tend: <ref> waits for you`. `subject` gives `<ref>` and the
  link: the pull request, else the request (a GitHub issue or a Jira key),
  else the work item's ID with no link. Where the stop's own text holds the
  link, `waits` replaces it with `<ref>`, so the link is given once.
- `main.go`: `stop` now goes through `halt(msg, body)`. It still says the
  message, writes `needs-you` with the same `next:` and exits 3. The
  notification's body is `body`, or the message when that is empty, then the
  link.
- `attempt.go`: an ended flow notifies `ended(verdict)` as a wait, or
  `tend: <ref> landed` with `Squash-merged at <head>.` and the link.
  `ended` words each case: passed under a `publish` that acts on an
  approval, passed under one that does not, findings that stand (with
  "after one repair" once the step that `actionable` leads to has reached
  its limit), checks still failing after the one fix, and an unpublished
  branch.
- `publication.go`: `publishStep` records, in `spentChecks`, the names of
  the checks that failed after the fix was spent, for `ended`. It resets
  them on each publish.
- `landing.go`: the protected-path hold now stops through `halt`. The
  status and stderr message are as before. The notification says `Held at
  land: <head> changes the protected path <p>. Approve this head to land
  it.`, then the link.
- `rex.go`: `notify` now writes what `rexMessage(title, body, sound)`
  returns, which is the state and the one-line message it built before.
- `testdata/bin/herdr`: `notification show` appends `{title, body, sound}`
  as one JSON line to `$HERDR_STATE/notifications`.
- `terminal_test.go`: the `notify-*` scenarios and `TestRexStopMessage`.

Checks: `go test -run 'TestRexStopMessage|TestTend/^notify' ./...` passed,
and so did `go test ./...` (all scenarios).

Also changed: `scenario_test.go` — it registers the new `notifyScenarios` list in `TestTend`.
Also changed: `guide/run.md` — its "Done" step and its "A person is needed" section quoted the old titles, `tend <id> needs you` and "ready for you".
Check: in the gh stand-in, a pull request's URL is `https://example.test/pr/<branch>`, not a GitHub URL. So `<ref>` comes from the URL when it is a github.com pull request or issue, and otherwise from `gh repo view --json nameWithOwner` plus the number (`gh pr view --json number`, or the issue's own number). In the scenarios that gives `o/r#1` for every pull request, not `o/r#42` as the spec's examples have it.
Check: "after one repair" is read from the attempt, not from the last verdict. The step that `actionable` leads to has a limit and has been asked that many times. So a repair followed by a failure that needs a decision still says "after one repair" (`notify-repair-then-decision`). A custom flow with a limit of 2 reads "after 2 repairs".
Check: an unpublished branch's choices are worded "merge the branch yourself, or change it" for a pass and "fix the branch, or drop it" for a failure. The spec asks only that the body names the branch and its head.
Check: a stop before the attempt's run folder exists, such as an issue not labelled `tend-build`, is recorded in `TEND_STATE/notified/<name of the work item's lock file>`. So a second run with the same stop notifies nothing (`notify-stop`). An issue stopped because its repository uses Grove is named by its number, and by its URL where gh gives it, before that stop (`notify-stop-in-grove`). A Jira issue stopped there is named by its key, with no link.
Check: the title keeps ` waits for you` or ` landed` whole. A ref too long for 60 characters loses its start, marked `…`, so its number stays (`TestNotificationTitle`).
Check: the record is the last notification only. Wait A, then wait B, then wait A again notifies A a second time, as decision 5 says.
Check: a dead tend run again by watch is checked by writing `working` into the status file and running `tend watch` once. The test does not kill a real tend.

### T-02 — Intake tells the owner of a proposal

Status: done

What changed:

- `intake.go`: once a proposal pull request is opened, reviewed and linked
  on its request, intake notifies it through T-01's `waits` and `tell`,
  with the request's run folder as the record. It notifies before it writes
  `linked` and `done`, so a pass stopped in between sends the notification
  on the next pass, which picks up a handoff with no `linked`. A
  notification already sent is not sent twice. The body reads `Proposal from <#n or
  KEY>: <n> open questions are consequential. Merge to approve, or revise
  it on <branch>.`, then the pull request's URL. With no question marked,
  it reads `see its open questions` in place of the count. `consequential`
  counts the summary's lines that begin, after any list marker, with
  `[consequential]`. `draft` now sets `item` and clears `pr` for each
  request, so a stop or notification names that request and not the
  previous one.
- `flows/prompts/follow.md`: the drafter is asked to begin each
  consequential question with the fixed marker `[consequential]` and to
  leave the routine ones unmarked.
- `intake_test.go`: `propose-notify` and `propose-notify-unmarked`, and a
  check in `propose-declined` that a declined request notifies no one.

Checks: `go test -run 'TestTend/^propose' ./...` passed, and so did
`go test ./...`.

Also changed: `testdata/bin/turn.sh` — the drafter stand-in writes `$SUMMARY` as its summary when it is set, so a test can mark questions.
Also changed: `scenario_test.go` — `SUMMARY` is added to the environment variables a scenario drops from the caller's.
Check: the fixed marker's text is `[consequential]`, chosen under the proposal's open question 1. The stand-in's default summary keeps its old `(consequential)` wording, so it counts as unmarked.
Check: `intake.go`, `intake_test.go`, `main.go`, `publication.go` and `publication_test.go` changed on main after this change's base, in the `tend/review` status on proposals and in reviewer notes. Neither touches what either task says. T-01 and T-02 were built on the files as main has them, and the change was re-based to 32e85a9, its `base` the only line changed, so that Grove's `check` passes.
Check: the proposal's notification sounds as `request`, the same as a stop, since the owner has to act on it.

## Scenarios

- terminals "A stop under Herdr": terminal_test.go `TestTend/notify-stop`, passed
- terminals "A stop under rex": terminal_test.go `TestRexStopMessage`, passed
- terminals "A landing": terminal_test.go `TestTend/notify-landed`, passed
- terminals "A supervised head that passed review": terminal_test.go `TestTend/notify-supervised`, passed
- terminals "A head that passed review where approval lands nothing": terminal_test.go `TestTend/notify-built-in`, passed
- terminals "Findings that stand after the repair": terminal_test.go `TestTend/notify-repair-spent`, passed
- terminals "A failure the reviewer did not judge repairable": terminal_test.go `TestTend/notify-decision`, passed
- terminals "No pull request": terminal_test.go `TestTend/notify-unpublished`, passed
- terminals "A protected path": terminal_test.go `TestTend/notify-held`, passed
- terminals "Still red after the fix": terminal_test.go `TestTend/notify-still-red`, passed
- terminals "A stop after publishing": terminal_test.go `TestTend/notify-other-stop`, passed
- terminals "A died tend run again by watch": terminal_test.go `TestTend/notify-once`, passed (the dead tend is simulated by a `working` status file)
- terminals "A new head": terminal_test.go `TestTend/notify-once`, passed
- terminals "A long stop reason": terminal_test.go `TestTend/notify-long`, passed
- intake "A proposal with consequential questions": intake_test.go `TestTend/propose-notify`, passed (the stand-in's pull request is `o/r#1`, not `#44`)
- intake "A summary with none marked": intake_test.go `TestTend/propose-notify-unmarked`, passed
- intake "The next intake pass": intake_test.go `TestTend/propose-notify`, passed (also with a pass stopped before and after the notification went)
- Whether each notification is short and plain enough to act on from a phone's lock screen: unverified, since only a person reading one on a phone can judge it

## Reconciliation

- terminals: applied to specs/terminals/spec.md
- intake: applied to specs/intake/spec.md
