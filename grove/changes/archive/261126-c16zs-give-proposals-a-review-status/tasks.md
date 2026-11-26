## T-01 — Set tend/review on proposal heads

Result: each open proposal pull request's head carries a `tend/review`
status that intake set from its changed paths, as the intake change spec's
"A proposal head of Grove documents only passes the rule" says.

Files:
- `intake.go`
- `intake_test.go`
- `testdata/bin/gh`

Constraints: the decision reads Git alone, never a file the drafter or the
run folder holds; no status is set by any other route than the account's
`gh`; publication's and landing's handling of `tend/review` is unchanged.

Acceptance:
- Each scenario of that requirement passes as an acceptance test driving
  the built binary.
- A second pass with no new head sets no status again.
- The existing intake, publication and landing tests still pass.
