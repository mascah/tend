## T-01 — The drafter's name holds its repository

Result: `tend propose` names each drafter after its request and its
repository, as the intake requirement "Draft, commit and open the proposal"
in this change says, so a pass never finds another repository's drafter.

Files:
- `intake_test.go`
- `scenario_test.go`
- `jira_test.go`
- `testdata/bin/turn.sh`
- `intake.go`
- `grove/specs/terminals/spec.md`

Constraints: the test comes first and fails before the fix. The name is the
same on every pass over a repository, so a later pass finds its own drafter.

Acceptance:
- The intake scenario "Two repositories with the same issue number" is a
  test that fails at 63fc1a0 and passes with the fix.
- The intake scenarios "A proposal from an issue" and "A Jira request" pass
  with the new names, and every other scenario still passes.
- The terminals spec's Purpose names the drafter as the intake requirement
  does.
