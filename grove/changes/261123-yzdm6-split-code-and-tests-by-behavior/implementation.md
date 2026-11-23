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

## Scenarios

## Reconciliation
