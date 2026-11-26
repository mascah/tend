# Implementation: Fail a head in review only for findings that matter

<!-- Progress on change 261126-5xkrh, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — Review to the bar, with notes and earlier verdicts

Status: done

- `flows/prompts/review.md` states the bar (the four kinds of finding that
  fail a head; contrived corners, the reason a stop gives, wording and polish
  go in `notes`), gives `{earlier}`, asks for the earlier failures to be
  judged fixed and for the commits since the latest earlier head to be
  judged, and not to fail for code unchanged since a head an earlier review
  saw unless it is a reachable security or data hole; the verdict shape adds
  `notes`, and `actionable` is judged on the findings that fail the head.
- `main.go` `valid`: `notes`, where present, must be a string.
- `flow_run.go`: `earlier(h)` fills `{earlier}` in `reach`, so for every
  step: each valid verdict file a step of the flow names, for a commit that
  is an ancestor of the head and not the head, oldest first, as
  `<short head> <conclusion> <file>`, joined by `; `, or `none`.
  `{findings}` is still the verdict's `summary` alone.
- `testdata/bin/turn.sh`: the reviewer stand-in writes to the verdict path
  in its prompt that does not yet exist (so an earlier verdict the prompt
  names is not overwritten), and gains `REVIEW=notes`, `badnotes` and
  `repairable-notes`.
- Tests: `attempt_test.go` `review-bar`, `review-notes`,
  `review-notes-not-text`, `repair-notes`, `review-earlier-gone`;
  `flow_test.go` `flow-own-review-bar`. All pass, as does `go test ./...`.

Check: `{earlier}` is one `; `-separated line (`<short head> <conclusion> <file>`, or `none`) rather than one line per verdict, since every built-in prompt is a single line sent to the agent's terminal; the proposal leaves its form to the implementer.
Check: which files count as the attempt's verdicts: those matching any step's `verdict <name>` result in the flow (with `{head}` as a wildcard), so a project's own verdict step's verdicts are named too; a verdict for the head under review itself is not "earlier" and is left out.

## Scenarios

## Reconciliation
