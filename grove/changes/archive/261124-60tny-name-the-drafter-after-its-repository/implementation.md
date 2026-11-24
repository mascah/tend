# Implementation: An intake drafter is named after its request and its repository

<!-- Progress on change 261124-60tny, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — The drafter's name holds its repository

Status: done

- `intake_test.go`: new scenario `propose-same-number`. It runs `tend
  propose` on a repository with issue #1, then on a second repository with
  its own issue #1 in the same Herdr, and checks that the second pass starts
  a drafter of its own name and prompts only it. Written first: before the
  fix it failed "each repository's drafter has its own name", "the second
  repository starts its own drafter" and "and prompts it, not the first's".
  `propose` now expects `prop-1-<cksum>`.
- `scenario_test.go`: `other()` gives a scenario a second repository.
- `intake.go`: `draft` names the drafter `prop-<request>-<cksum of the
  repository's path>`, the number `issueRun` gives an issue's attempt.
- `testdata/bin/turn.sh`: the drafter stand-in takes the issue's number from
  its name less the cksum, for the `Refs #<n>` line it commits.
- `jira_test.go`: `jira-propose` expects `prop-scrum-<n>-<cksum>`.
- `grove/specs/terminals/spec.md`: the Purpose names the drafter
  `prop-<request>-<cksum of the repository's path>`.
- Checks: `go test -count=1 ./...` passes (87 s).

Check: the second repository in `propose-same-number` has its own stand-in
GitHub and its own `TEND_STATE`, and shares only Herdr's open agents, copied
from the first. The stand-ins keep one GitHub for each state folder, so two
repositories in one could not each hold an issue #1 and a pull request from
`tend/issue-1`. On a real machine both share one `TEND_STATE`; their intake
folders differ by the repository's name on GitHub.
Check: the cksum is of the repository's path as tend was given it, made
absolute, as for an issue's attempt. The same repository reached by another
path (a symlink) gets another drafter name; tend then treats the first as
gone and asks once more in a new session.
Check: not tried against a live Herdr. That `herdr agent get <name>` finds
an agent in any workspace is read from `herdr.go`, which passes no
workspace, and from issue #26.

## Scenarios

- intake "A proposal from an issue": intake_test.go `TestTend/propose`, passed
- intake "Origin's target moved": intake_test.go `TestTend/fetch-propose`, passed
- intake "A Jira request": jira_test.go `TestTend/jira-propose`, passed
- intake "Two repositories with the same issue number": intake_test.go `TestTend/propose-same-number`, passed

## Reconciliation

- intake: applied to grove/specs/intake/spec.md
