# Implementation: Run steps from a flow file

Built in the order T-05, T-01 to T-03 together, then T-04: the old loop
was ported to Go tests first, so the rewrite ran against all 72 scenarios.

Check: T-01, T-02, T-03 and T-04's command are one commit naming all
four; T-04's README section is a second commit. The engine replaces
functions `main.go` defined, so neither the engine nor the built-in flows
build or pass alone, and splitting them would leave commits that do
neither.

## Tasks

### T-01 — The step engine and the flow file

Status: done

`flow.go` reads `tend/flow.yaml` from the target as origin holds it
(`git show <target>:tend/flow.yaml`, after the fetch the attempt already
makes), with `gopkg.in/yaml.v3` and unknown fields refused, and falls back
to the built-in flows embedded from `flows/`. It validates the whole file
before any agent starts. Then `walkFlow` runs the steps: it records each
ask, prompts once, checks the result, follows the outcome, holds the limit,
and on a rerun takes a result already there or waits for an agent still
working. `flow_test.go` holds unit tests of validation (`TestFlowFaults`,
`TestBuiltInFlowsAreValid`, `TestRequestedLoopNeedsNoLimit`) and a scenario
for each of the flows spec's (`TestFlows/flow-*`), through the binary and
the stand-ins.

Checks run on 2026-11-21: `gofmt -l .` and `go vet ./...` clean;
`go test ./...` passed in 43.8 s, every `TestTend` and `TestFlows` subtest.

Also changed: `testdata/bin/turn.sh` — gives a project step `check-*` its
turn (`CHECK=success|failure|none`), and adds `REVIEW=actionable` for a
reviewer that fails every head as actionable.
Also changed: `testdata/bin/herdr` — adds `CHECK_FAIL`, and while
`IMPL_SLEEP` holds the implementer's prompt it marks that agent working
with its turn pending, as a real agent is after tend is killed mid-turn.
Also changed: `tend_test.go` — keeps `CHECK*` out of the inherited
environment, and `lock` now also checks that the implementer was prompted
once in all after the takeover.

Check: a prompt the terminal refused (an error code containing neither
`timeout` nor `stall`, and not "no answer") has its head taken back out of
the step's record, so a rerun asks it again. A timed-out or stalled prompt
keeps it, counting as possibly delivered, as the attempts spec says.
Check: a step recorded as asked, with no result and its agent not working,
now stops on every rerun for every step, the implementer's included: "Look
at its pane ... or delete `<record>` to have it asked once more". Before,
tend would have prompted the implementer again. The flows spec requires
this, and so does the attempts spec's "an agent still working ... not
prompted again". The old `IMPL_SLEEP` stand-in left the agent idle while
tend was killed, and the new engine rightly stopped on it; it was the only
check in `dev/test.sh` that failed against the rewrite (`lock: after the
holder is killed, tend takes over`), and the stand-in now models the
working agent.
Check: a step's fields are `agent` (`claude` or `codex`) or `continue`
(a step), `prompt`, `result`, `limit` and `next`. A result is `commit`,
`file <name>`, `verdict <name>` or `commit or file <name>`, and with none the
turn ending is `ok`. Names are in the run folder and may hold `{head}`,
`{review_id}` and the like. A field naming a flag is refused like any
unknown field, by the YAML decoder's own message: `tend/flow.yaml is not a
flow file tend can read: line N: field flags not found in a step`.
Check: records are `<step>-asked`, a head a line, so the built-in `repair`
keeps `repair-asked`. The propose flow keeps `asked` and `followed`, and
requested changes keep `request-<review id>`, the names the specs and
README give.
Check: a prompt file is looked up under `tend/` on the target and then
among tend's built-in prompts, so a project can override one prompt and
keep the rest. A built-in flow reads only built-in prompts. `{kind}`
(`grove` or `issue`) is filled in in prompt paths too, which is how
`implement` picks its prompt.
Check: the loop check exempts a loop through `publish`'s `requested`
outcome, because a person requesting changes bounds it; the built-in
request loop has no limit.
Check: agents are named as before. The start step's agent is the
implementer (`impl-…`, or `prop-…` when proposing), a step continuing
`review` is `rev-<short head>`, and any other fresh step is
`<step>-<short head>`. A fresh agent not continued by a later step is closed
once its result is checked. In `build`, a step whose result is a verdict or
a file only must leave HEAD and the tree as they were, or tend stops saying
the agent changed the checkout.

### T-02 — The build flow is built in

Status: done

`flows/build.yaml` holds `implement`, `review`, `repair` (limit 1, `spent`
to `publish`), `publish` and `request`, and `flows/prompts/` holds their
prompts, copied word for word from the strings `main.go` held. In `main.go`,
`attempt` now loads the flow after fetching the target and calls `walkFlow`
with `publishStep`; `implementer`, `checked` and `review`, which ran the loop
directly, are gone. Grove cycles, GitHub issues and Jira issues all run
through it: every ported scenario passes with no flow file in the fixture.

Also changed: `docs/work/261121-autonomous-direction/steps.md` — step 1's
status line said "not approved, not built"; it now says approved in #12 and
built on this branch.

Check: acting on requested changes is the step `request`, which
`publish` leads to through its outcome `requested`. Its result is
`commit or file reply-{review_id}`. A reply with no commit is posted and
answered, and with a commit, the new head is reviewed. So `publish` has the
outcomes `ok` and `requested`, and a flow places `request` like any step.
Check: with a project step whose result is also a verdict, the verdict
printed at the end is the newest `verdict-*`, the reviewer's, as before.
A flow without a verdict step prints "done … with no verdict" and exits 0.

### T-03 — The propose flow is built in

Status: done

`flows/propose.yaml` holds `draft`, `follow` (`commit or file declined`) and
`publish`. `draft` in `main.go` fills in the request and walks the flow, and
what came after the drafter's turn is now `proposal`, run as the flow's
`publish`. Every intake and Jira scenario passes with no flow file.

Also changed: `flow.go` — the engine `draft` calls (named under T-01 and
T-02, not T-03).

### T-04 — `tend flow` prints the flows in effect

Status: done

`tend flow <repo>` prints a header, then each flow headed by where it comes
from (`# build: built in`, or `# build: tend/flow.yaml on origin/main`), as
flow-file text. `TestFlows/flow-print` saves its output as `tend/flow.yaml`,
finds it valid, and runs a repaired attempt the same as before. The README
has a section "Flows: the steps an attempt runs", whose example flow
(the built-in `build` with a `check` step) was checked to be valid.

Also changed: `flow.go` — `printFlows` sits beside the reading it shares.

Check: `tend flow` fetches the target from origin first, so it prints what
an attempt would run. That moves only origin's remote-tracking ref, and no
branch, worktree or file of the project; that is what I took "reading only"
to mean. The target comes from Grove, else origin's HEAD, else the current
branch.
Check: a project's flow is printed as tend parsed it, so its comments are
not kept; the built-in flows are printed as written, comments included.

### T-05 — Go tests replace `dev/test.sh`

Status: done

`tend_test.go` drives the built binary through stand-ins extracted to
`testdata/bin` (`herdr`, `rex`, `gh`, `curl`, `turn.sh`). Every one of the 72
scenarios is a subtest of `TestTend` with the same name, all run in
parallel. `go test ./...` took about 30 s where `dev/test.sh` took 7 min
37 s, and `go test -run 'TestTend/^remote-claim$'` runs one alone (1.9 s).
`dev/test.sh` is deleted (`dev/fixture.sh` stays); CI runs `go test ./...`;
the README and `grove.yaml`'s `context` line name the Go tests and no longer
say standard library only. A `tend` that only exits 0 failed the scenarios
with their check text, so the checks do bite.

Check: main changed `dev/test.sh` after base ed385ad, in d6b3e7e
(#13). Its `rex-add-dir` gained "without waiting out the agents' first
reports", checking the run takes under 60 s. The port is of `dev/test.sh` as
main holds it, so that check is in `TestTend/rex-add-dir` and passes.
Check: three checks of `jira-ineligible` are stricter than in Bash, where
`check … test $? = 3 && said …` ran its `said` outside `check` and ignored
it. Each now tests the exit code and the message, and passes.
Check: `KEEP` was dropped; `TEND` still names another build to test.
shellcheck was not added for `testdata/bin/*`, since it is not installed
here to confirm it passes; CI still checks `install.sh` only.

## Scenarios

- flows "A project without a flow file": every `TestTend` scenario (no flow file in the fixture), e.g. `TestTend/publish-repair`, passed
- flows "A build with one more step": `TestFlows/flow-one-more-step`, passed
- flows "The implementer edits the flow": `TestFlows/flow-from-target`, passed. The committed flow drops `review` by sending `implement`'s `ok` to `publish`, since removing the step outright leaves `repair` naming it
- flows "A step that leaves no result": `TestFlows/flow-no-result`, passed
- flows "A repair continues the implementer": `TestFlows/flow-repair-continues`, passed
- flows "A flow tries to set a flag": `TestFlows/flow-flag` and `TestFlowFaults`, passed
- flows "A failing check goes to a person": `TestFlows/flow-check-fails`, passed
- flows "A second repair": `TestFlows/flow-second-repair`, passed
- flows "The built-in one repair": `TestFlows/flow-one-repair`, passed
- flows "tend is killed during a project's own step": `TestFlows/flow-killed-in-step`, passed. The kill is modelled as the state it leaves, `check` recorded as asked and its agent working (reached through a prompt that timed out); a real `kill -9` mid-turn is `TestTend/lock`, for the implementer
- flows "The outcome of a step is unknown": `TestFlows/flow-unknown-outcome`, passed
- flows "A flow without publish": `TestFlows/flow-without-publish`, passed
- flows "A loop with no limit": `TestFlows/flow-loop` and `TestFlowFaults`, passed
- flows "Starting a project's file": `TestFlows/flow-print`, passed
- attempts, review, publication, work-items, intake and terminals: the 72 `TestTend` scenarios, ported from `dev/test.sh` with the same names, passed against the rewritten engine

## Reconciliation

- flows: applied to `grove/specs/flows/spec.md`, a new maintained spec, with
  every ADDED requirement copied in word for word. Its "Observed current
  behavior" was what the work replaced, so none of it is kept.
