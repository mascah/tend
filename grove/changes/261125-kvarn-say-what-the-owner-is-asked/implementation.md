# Implementation: Say what the owner is asked to do in each notification

<!-- Progress on change 261125-kvarn, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — An attempt's notifications say what the owner is asked to do

Status: done

What changed:

- `terminal.go`: the one function every notification goes through, `tell`,
  given the folder that records it, the title and the body. It skips a
  notification identical to the one in that folder's `notified`, and records
  each one it sends. `fit` applies the length bound: a title of at most 60
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
Check: "after one repair" is read from the flow. The step that `actionable` leads to has a limit and has been asked that many times. A custom flow with a limit of 2 reads "after 2 repairs".
Check: an unpublished branch's choices are worded "merge the branch yourself, or change it" for a pass and "fix the branch, or drop it" for a failure. The spec asks only that the body names the branch and its head.
Check: a stop before the attempt's run folder exists, such as an issue not labelled `tend-build`, is recorded in that issue's run folder when an earlier run made it. Otherwise it is not recorded, so it is notified on every run.
Check: the record is the last notification only. Wait A, then wait B, then wait A again notifies A a second time, as decision 5 says.
Check: a dead tend run again by watch is checked by writing `working` into the status file and running `tend watch` once. The test does not kill a real tend.

## Scenarios

## Reconciliation
