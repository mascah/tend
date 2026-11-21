## T-01 — The step engine and the flow file

Result: tend can run a flow read from `tend/flow.json` on the target branch:
prompt each step once, check its result, follow its outcome, hold its limit,
recover after a kill, and refuse a flow that is not valid.

Files:
- `flow.go`
- `flow_test.go`

Interfaces: the flow file `tend/flow.json` and its prompt files; the step
record in an attempt's run folder.

Constraints: standard library only. A flow names a harness and none of its
flags.

Acceptance:
- Each scenario of "The flow is read from the target branch", "A step prompts
  once and its result is checked", "A fresh agent or a continuing one", "An
  outcome chooses what follows", "A step is asked no more often than its
  limit", "Any step recovers without being prompted twice" and "A flow that
  is not valid" is a test that passes with stand-in agents.

## T-02 — The build flow is built in

Result: a Grove cycle, a GitHub issue and a Jira issue are built by the
built-in `build` flow through the engine, and the code that ran that loop
directly is gone.

Files:
- `main.go`
- `flow.go`
- `flows/build.json`
- `flows/prompts/`

Interfaces: the built-in flow `build`, with its step `publish`.

Constraints: the attempts, review, publication, work-items and terminals
specs hold as written, including the names of run files.

Acceptance:
- "The built-in flows", "A project's flow replaces the built-in one of its
  name" and "Publishing is tend's own and a flow places it" each have a test
  that passes.
- Every scenario of the attempts, review, publication and work-items specs
  passes as a test, with no flow file in the fixture.

## T-03 — The propose flow is built in

Result: `tend propose` drafts each request through the built-in `propose`
flow, and the code that ran that loop directly is gone.

Files:
- `main.go`
- `flows/propose.json`
- `flows/prompts/`

Interfaces: the built-in flow `propose`.

Constraints: the intake spec holds as written.

Acceptance:
- Every scenario of the intake spec passes as a test, with no flow file in
  the fixture.

## T-04 — `tend flow` prints the flows in effect

Result: `tend flow <repo>` prints the flows a repository's attempts would
run, as text a project can save as its flow file.

Files:
- `main.go`
- `README.md`

Interfaces: the command `tend flow <repo>`.

Acceptance:
- The scenario of "Print the flows in effect" is a test that passes.
- The README says what a flow file is, where it lives, what a step can say,
  and that it is read from the target branch.

## T-05 — Go tests replace `dev/test.sh`

Result: tend's acceptance tests are Go tests that drive the built binary
through the stand-ins, run in parallel, and can be run one at a time.

Files:
- `tend_test.go`
- `testdata/`
- `dev/test.sh`
- `.github/workflows/ci.yml`
- `README.md`
- `grove.yaml`

Acceptance:
- Each of the 72 scenarios in `dev/test.sh` at ed385ad has a Go test of the
  same name, or a line in `implementation.md` saying why it has none.
- `go test ./...` passes, and CI runs it in place of `dev/test.sh`.
- One scenario runs alone by name.
- `dev/test.sh` is removed, and the README and the `context` line of
  `grove.yaml` name the Go tests.
