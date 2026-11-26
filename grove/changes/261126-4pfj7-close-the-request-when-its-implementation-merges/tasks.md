## T-01 — Intake records the changes a proposal drafted

Result: a request's intake folder names the changes its proposal adds, as
the intake change spec's "The intake folder records the changes drafted"
says.

Files:
- `intake.go`
- `intake_test.go`

Interfaces: `changes` in an intake folder, one change ID per line, written
before `linked`.

Acceptance:
- Each scenario of the intake change spec passes as a test driving the built
  binary.
- A pass that finishes an interrupted handoff writes `changes` too.
- The maintained intake spec's scenarios still pass.

## T-02 — The implementation pull request closes its requests

Result: publishing a Grove attempt's head adds `Closes #<n>` for each GitHub
request that drafted one of its changes and comments once on each such Jira
request, as the publication change spec says.

Files:
- `publication.go`
- `attempt.go`
- `request.go`
- `jira.go`
- `publication_test.go`
- `scenario_test.go`
- `testdata/bin/gh`
- `testdata/bin/curl`

Interfaces: a lookup giving, for the attempt's change IDs, the requests
whose intake folders list them, each with its closing line or its Jira
link; the pull request's description is the message's body plus the
missing closing lines, wherever tend opens the pull request, rewrites its
description or compares it to tell a person's edit.

Constraints: an issue build is published as before, its message still
ending `Closes #<n>` or `Refs <KEY>`; a description a person edited is left
as it is.

Acceptance:
- Each scenario of the publication change spec passes as a test driving the
  built binary.
- The maintained publication and work-items specs' scenarios still pass.

## T-03 — Documentation

Result: the README and guide say that intake records the changes a proposal
drafted, and that an implementation pull request closes the GitHub requests
its changes were drafted from and comments on the Jira ones, noting that
GitHub honours `Closes` only on pull requests into the default branch.

Files:
- `README.md`
- `guide/concepts/tend-state.md`
- `guide/concepts/intake.md`
- `guide/concepts/publication.md`

Acceptance:
- `tend-state.md` lists `changes` among an intake folder's files.
- The publication and intake pages describe the closing line and the Jira
  comment, and say a change no request drafted gets neither.
