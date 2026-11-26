## T-01 — Review to the bar, with notes and earlier verdicts

Result: the built-in review prompt states the bar and the earlier-rounds
rule, a verdict may carry `notes`, and every step can be given the
attempt's earlier verdicts as `{earlier}`, as the review change spec's
"Fail a head only for findings that matter", "A later review judges what
changed", "A fresh reviewer for each head" and "Only a valid verdict counts"
say.

Files:
- `flows/prompts/review.md`
- `main.go`
- `flow_run.go`
- `attempt_test.go`
- `flow_test.go`
- `testdata/bin/turn.sh`

Interfaces: the verdict's optional `notes` string; the `{earlier}`
placeholder, filled for every step.

Constraints: `{findings}` stays the verdict's `summary` alone; a verdict
written before this change, with no `notes`, stays valid.

Acceptance:
- The prompt a reviewer receives from the built-in flow names the four kinds
  of finding that fail a head, asks for every other finding in `notes`, and
  holds the earlier verdicts or says there are none.
- A passing verdict with `notes` leads to no repair, `done: success` and
  exit 0.
- After an actionable failure, the repair prompt holds the summary and not
  the notes, and the review of the repaired head names the failed head, its
  conclusion and its verdict file.
- A verdict for a commit the branch no longer contains is not named in
  `{earlier}`.
- A verdict without `notes` counts; a new verdict whose `notes` is not a
  string makes tend exit 3 saying the reviewer settled without a valid
  verdict.
- A project's own `tend/prompts/review.md` is still the text the reviewer
  receives.

## T-02 — Notes on the pull request

Result: the comment review tend posts for a head shows the verdict's notes
under their own heading, and notes never change the `tend/review` status, as
the publication change spec's "A review and a status for each reviewed head"
says.

Files:
- `publication.go`
- `publication_test.go`

Acceptance:
- A passing head whose verdict has `notes` gets status `success` and one
  review showing the note under a heading saying it does not fail the head.
- A head whose verdict has no `notes` gets a review with no notes heading.
- A failing head's review shows its notes after its summary, and still says
  whether the failure was judged repairable or needs a decision.

## T-03 — Retire tend's own review prompt

Result: tend's own repository reviews with the built-in prompt, its stopgap
`tend/prompts/review.md` removed, so its reviews get `{earlier}` and
`notes` too.

Files:
- `tend/prompts/review.md`

Acceptance:
- `tend/prompts/review.md` no longer exists, `tend/flow.yaml` still names
  `prompts/review.md`, and the flow is valid, its review prompted with the
  built-in text.
