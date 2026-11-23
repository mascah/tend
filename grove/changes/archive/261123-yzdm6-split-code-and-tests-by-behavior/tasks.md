<!-- Every task shares these constraints, from the proposal's scope and
decisions: tend's behavior does not change; everything stays in `package
main` at the root; no record's path or format changes; code moves as it is
unless an edit is needed, and each edit that is not a move is recorded in
`implementation.md` with its reason. Line numbers are those of `43758c3`. -->

## T-01 — Landing has its own files

Result: the landing code, which is the `// ---- landing` section of
`main.go` (from `main.go:1981`), is in `landing.go`, and `TestLanding`, its
scenarios and their fixtures (from `tend_test.go:2106`) are in
`landing_test.go`.

Files:
- `main.go`
- `landing.go`
- `tend_test.go`
- `landing_test.go`

Delegated: which of this task's files each declaration goes in.

Constraints: `TestLanding` and each of its scenarios keep their names.

Acceptance:
- `go build -o bin/tend .` succeeds and `go test ./...` passes.
- `go test -v -run TestLanding` reports the same scenario names as at
  `43758c3`.
- `main.go` holds no landing lock, reviewed-head, check or merge decision,
  and `tend_test.go` no landing scenario.

## T-02 — The CLI and attempts are apart

Result: `main.go` holds dispatch and startup; starting and resuming a build,
with its locks and its claim, are in `attempt.go`; and creating, finding,
removing and tidying a worktree are in `worktree.go`.

Files:
- `main.go`
- `attempt.go`
- `worktree.go`

Delegated: which of this task's files each declaration goes in.

Constraints: `version` stays a variable of `package main`, as the release
build sets it with `-X main.version`.

Acceptance:
- `go build -o bin/tend .` succeeds and `go test ./...` passes.
- `bin/tend version` prints what it printed at `43758c3`.
- `main.go` holds no declaration that only attempts or only worktrees use.

## T-03 — Watch and status have their own files

Result: scheduling and reruns, which are `watch` and what only it calls, are
in `watch.go`, and reporting, which is `status`, `overview` and what only
they call, is in `status.go`.

Files:
- `main.go`
- `watch.go`
- `status.go`

Delegated: which of this task's files each declaration goes in.

Acceptance:
- `go build -o bin/tend .` succeeds and `go test ./...` passes.
- `main.go` holds no declaration that only watch or only status uses.

## T-04 — Publication has its own file

Result: publishing a pull request and answering it, which is the publish
step, the pull request's title and description, verdict reviews and
statuses, and the response to requested changes, with the GitHub account
and token helpers they use, are in `publication.go`.

Files:
- `main.go`
- `publication.go`

Delegated: which of this task's files each declaration goes in.

Acceptance:
- `go build -o bin/tend .` succeeds and `go test ./...` passes.
- `main.go` holds no declaration that only publication uses.

## T-05 — Intake and terminals have their own files

Result: proposal intake, which is `propose`, `draft` and `proposal`, is in
`intake.go`; the `request` type and what GitHub and Jira requests share are
in `request.go`, leaving `jira.go` with Jira alone; and the `terminal`
interface, with starting, awaiting and prompting an agent and the dialogs
tend answers, is in `terminal.go`, leaving `herdr.go` with Herdr alone.

Files:
- `main.go`
- `intake.go`
- `request.go`
- `jira.go`
- `terminal.go`
- `herdr.go`

Delegated: which of this task's files each declaration goes in.

Acceptance:
- `go build -o bin/tend .` succeeds and `go test ./...` passes.
- `main.go` holds no declaration that only intake or only terminals use.
- `jira.go` holds nothing a GitHub request uses, and `herdr.go` nothing rex
  uses.

## T-06 — Flows are three files

Result: `flow.go` holds a flow's definition and its validation;
`flow_load.go` holds finding the flow file at a revision, the built-in
defaults and their prompts; and `flow_run.go` holds walking a flow, its
records of what was asked, its agents and its recovery.

Files:
- `flow.go`
- `flow_load.go`
- `flow_run.go`

Delegated: which of this task's files each declaration goes in.

Constraints: the built-in flows stay embedded from `flows/`; `tend flow`
prints what it printed at `43758c3`.

Acceptance:
- `go build -o bin/tend .` succeeds and `go test ./...` passes.
- `flow.go` reads no Git revision, loads no embedded file and starts no
  agent.

## T-07 — Tests are filed by behavior

Result: `tend_test.go` is gone; the fixtures every scenario shares, with
`TestMain` and `TestTend`, are in `scenario_test.go`; each scenario is in the
test file of the behavior it checks; and the documents that named
`tend_test.go` name the files that replace it.

Files:
- `tend_test.go`
- `scenario_test.go`
- `attempt_test.go`
- `worktree_test.go`
- `watch_test.go`
- `status_test.go`
- `publication_test.go`
- `intake_test.go`
- `jira_test.go`
- `terminal_test.go`
- `grove.yaml`
- `README.md`

Delegated: which of this task's files each fixture and scenario goes in.

Constraints: `TestTend`, `TestLanding` and `TestFlows` and each scenario keep
their names, and each scenario stays under the test it is under at
`43758c3`.

Acceptance:
- `go test ./...` passes.
- `go test -v ./...` reports the same set of test and scenario names as at
  `43758c3`.
- `go test -run 'TestTend/^publish$'` runs that one scenario.
- No file in the repository outside `grove/changes/` names `tend_test.go`.
