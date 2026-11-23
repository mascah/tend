# Implementation: Split the CLI implementation and tests by behavior

<!-- Progress on change 261123-yzdm6, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

Every move was made by a small throwaway tool that cuts each top-level
declaration with its comments out of its file, verbatim, appends it to the
file the task names, rewrites each touched file's import list to what its code
uses and runs gofmt. So, unless a task says otherwise below, the diff is moved
lines and import lists. Baseline: `go test -v ./...` on `6ae5060` (the code is
that of `43758c3`) passed, reporting 175 test and scenario names.

## Tasks

### T-01 — Landing has its own files

Status: done

- `landing.go`: the whole `// ---- landing` section of `main.go`: `landed`,
  `verdictName`, `passed`, `carried`, `carry`, `landingLock`, `checksOn`,
  `checked`, `landStep`.
- `landing_test.go`: the `// ---- landing` section of `tend_test.go`:
  `landFlow`, the landing fixtures (`lands`, `ci`, `landedMsg`, `merges`,
  `mergedAt`, `merged`, `approve`, `parents`), `landScenarios` and
  `TestLanding`.
- Edit beyond moves: the `// ---- landing` section comment is dropped from the
  top of both new files, where the file's name says it.
- Checks: `go build -o bin/tend .` succeeded; `go test ./...` passed;
  `go test -v -run TestLanding` reported the same 23 names as the baseline.
  `main.go` keeps no landing lock, reviewed head, check or merge decision
  (`grep` for them finds only `watch`'s own lock and claim code).

### T-02 — The CLI and attempts are apart

Status: done

- `attempt.go`: starting and resuming a build with its locks and its claim:
  `cksum`, `firstRun`, `unclaimable`, `issueRun`, `takeLock`, `lockFile`,
  `held`, `settingUp`, `setUp`, and the `// ---- an attempt` section
  (`attempt`, `unbuilt`, `built`).
- `worktree.go`: `layout`, `contained`, `worktreeFor`, `removeWorktree`,
  `tidy`, `superseded`.
- `main.go` keeps dispatch and startup (`version`, `main`, the globals) and
  the helpers more than one area uses, as open question 2 recommends: the
  command, JSON and reporting helpers, and the Grove and Git readers several
  areas call (`groveIn`, `syncTarget`, `groveJSON`, `project`, `itemRun`,
  `workState`, `runDirs`, `isIssue`, `isItem`, `fetchTarget`, `commits`,
  `clean`, `prNumber`, `reviewsAPI`, `postJSON`, `valid`).
- Edit beyond moves: none.
- Checks: `go build -o bin/tend .` succeeded; `bin/tend version` printed
  `tend dev`, as at `43758c3`; `go test -count=1 ./...` passed. A reference
  scan (each `main.go` declaration and the files that use it) found none used
  only by `attempt.go` or only by `worktree.go`.

Check: `lockFile` and `issueRun` are also read by watch, status and landing;
they went to `attempt.go` because the task names the attempt's locks there
and an issue attempt's run folder is named by `issueRun`. `contained` and
`layout`, the guard that work stays in the attempt's worktree, went to
`worktree.go` as part of the checkout's lifecycle; `valid` (whether a verdict
file is complete) stayed in `main.go`, being used by publication, landing and
flows alike.

### T-03 — Watch and status have their own files

Status: done

- `watch.go`: `needsYou`, `watch`, `countIn`, `rerun`, `stopped`, `keep`,
  `unattended`.
- `status.go`: `overview`, `status`.
- Shared and kept in `main.go`: `working` (watch's `stopped` and `status`),
  `runDirs` (watch, `keep` and `overview`), `syncTarget` (`countIn` and
  `fetchTarget`), `groveIn`, `groveJSON`, `project`, `itemRun`, `workState`.
- Edit beyond moves: none.
- Checks: `go build -o bin/tend .` succeeded; `go test -count=1 ./...`
  passed; the reference scan found no `main.go` declaration used only by
  `watch.go` or only by `status.go`.

### T-04 — Publication has its own file

Status: done

- `publication.go`: the GitHub account and token helpers `accountOf`, `token`
  and `pushAs`, and publishing and answering a pull request: `told`,
  `publishStep`, `publish`, `markDraft`, `answer`, `message`, `titleBody`,
  `fold`, `redescribe`, `writer`, `approved`, `requested`.
- Shared and kept in `main.go`: `prNumber`, `reviewsAPI` and `postJSON`, which
  intake's `proposal` uses as well.
- Edit beyond moves: the `// ---- Git and GitHub` section comment, which sat
  above `accountOf`, stays in `main.go` above the Git and GitHub helpers that
  remain there (`fetchTarget` on), instead of moving with `accountOf`.
- Checks: `go build -o bin/tend .` succeeded; `go test -count=1 ./...`
  passed; the reference scan found no `main.go` declaration used only by
  `publication.go` but `ask`, which is the terminal's and moves with T-05.

Check: `accountOf`, `token` and `pushAs` are also used by intake and landing;
they are in `publication.go` because the task names the GitHub account and
token helpers there.

### T-05 — Intake and terminals have their own files

Status: done

- `intake.go`: `propose`, `draft`, `proposal`.
- `request.go`: the `request` type and `ghRequest` from `jira.go`, with
  `marker` from `main.go`, which only `ghRequest` uses. `jira.go` keeps Jira
  alone: nothing in it is used by `ghRequest`.
- `terminal.go`: the `terminal` interface from `herdr.go`, and the
  `// ---- agents` section of `main.go`: `dialogs`, `dialogOn`, `start`,
  `settled`, `await`, `ask`. `herdr.go` keeps the `herdr` type, its methods
  and `settledAs`, which only they use; rex uses nothing in it.
- Edit beyond moves: the `// ---- Jira` section comment is dropped from the
  top of `jira.go`, now the file's whole subject.
- Checks: `go build -o bin/tend .` succeeded; `go test -count=1 ./...`
  passed; the reference scan found no `main.go` declaration used only by
  intake or only by the terminal code.

Check: `ghRequest`, the GitHub request, went to `request.go` with the type
it builds: the table gives GitHub requests no file of their own, and
`jira.go` may hold nothing a GitHub request uses.

### T-06 — Flows are three files

Status: done

- `flow.go` keeps a flow's definition and validation: `step`, `flow`,
  `flowFile`, `watchSetting` and their methods, `result`, `parseResult`,
  `sorted`, `and`, `decode`.
- `flow_load.go`: `builtin` (still `//go:embed flows`), `flowPath`,
  `promptText`, `flowsAt`, `watchCount`, `printFlows`, `loadFlow`.
- `flow_run.go`: the `// ---- the engine` section: its variables, `fill`,
  `prompted`, `walkFlow`, the ask records (`recordOf`, `asks`, `timesAsked`,
  `setAsks`), agents (`agentOf`, `agentAt`, `startAgent`, `implementer`,
  `doing`), `promptStep`, recovery (`againOf`, `again`, `continued`) and
  reaching steps (`reach`, `reachAt`, `implStep`, `fileName`,
  `reachInOrder`).
- Edit beyond moves: none.
- Checks: `go build -o bin/tend .` succeeded; `bin/tend flow .` in this
  repository printed byte for byte what a build of `43758c3` printed;
  `go test -count=1 ./...` passed. `flow.go` calls no `git`, `quiet`, `ok`,
  `builtin` or terminal method.

Check: `loadFlow` sets the engine's variables but went to `flow_load.go`, as
it is finding the flow in effect at a revision; the variables it sets stay
with the engine in `flow_run.go`.

### T-07 — Tests are filed by behavior

Status: done

- `tend_test.go` is renamed `scenario_test.go` with `git mv` and keeps the fixtures every scenario shares, `TestMain`, `TestTend`
  and the `version` and `usage` scenarios.
- The one `scenarios` list became one list for each test file, each scenario
  kept in its order within its list: `attemptScenarios` (36),
  `worktreeScenarios` (11), `watchScenarios` (22), `statusScenarios` (3),
  `publicationScenarios` (15), `intakeScenarios` (12), `jiraScenarios` (5),
  `terminalScenarios` (12), and `cliScenarios` (2) in `scenario_test.go`.
  118 scenarios before and after.
- Fixtures only one file's scenarios use moved with them: `jbuild` and the
  Jira section (`jmap`, `adf`, `jissue`, `jcomments`, `rich`) to
  `jira_test.go`; `build`, `bworktree` and `verdict` to `attempt_test.go`;
  `labelled` to `intake_test.go`; `watchrepo`, `atOnce` and `unattend` to
  `watch_test.go`.
- Edits beyond moves:
  - `var scenarios = slices.Concat(cliScenarios, attemptScenarios, …)` in
    `scenario_test.go` gathers the lists, so `TestTend` itself is unchanged
    and every scenario stays under it.
  - The section comment over `TestTend`, "the scenarios, as dev/test.sh had
    them, in its order", became "the scenarios, each list in its behavior's
    test file": the order is no longer dev/test.sh's.
  - The `// ---- other helpers` section comment stays in `scenario_test.go`
    above `rexwt` instead of moving with `bworktree`.
- `README.md` and the `context` line of `grove.yaml` name `scenario_test.go`
  and the behavior's test files, such as `landing_test.go`, instead of
  `tend_test.go`.
- Checks: `go test -count=1 -v ./...` passed and reported the same 175 test
  and scenario names as the baseline (sorted lists compared with `diff`);
  `go test -run 'TestTend/^publish$'` ran that one scenario and passed;
  `git grep tend_test -- ':!grove/changes'` finds nothing.

Check: test files follow open question 3's recommendation, plus one choice
per scenario. Recovery and repair scenarios (`repair*`, `survivor-*`,
`unfinished`, `cached-*`, `legacy`) are in `attempt_test.go`, as the attempts
spec covers recovering an attempt and no `flow_run_test.go` was asked for;
the containment scenarios (`stray-*`, `at-once-stray`,
`at-once-during-turn`, `at-once-setting-up`) and every tidy scenario,
`rex-tidy` included, are in `worktree_test.go`; `at-once-intake` and
`at-once-second-intake` are in `intake_test.go`; `version` and `usage`, the
CLI's dispatch, stay in `scenario_test.go`, as no test file was asked for
`main.go`.
Found: `built()` in `attempt.go` (in `main.go` at `43758c3`) has no caller;
`walkFlow` and `reachInOrder` call `unbuilt()` directly. Seen in the
reference scan made to place declarations.

## Scenarios

None: the change has no change specs, as it changes no behavior. That
behavior did not change is shown by `go test -count=1 -v ./...` on the final
tree: passed, reporting the same 175 test and scenario names as `6ae5060`.

## Reconciliation

None: the change has no change specs, so no maintained spec changes.
