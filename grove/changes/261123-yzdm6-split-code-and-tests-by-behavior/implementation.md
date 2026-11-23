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

## Scenarios

## Reconciliation
