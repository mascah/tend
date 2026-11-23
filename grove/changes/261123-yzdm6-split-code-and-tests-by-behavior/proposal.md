# Split the CLI implementation and tests by behavior

## Problem

[Issue #28](https://github.com/mascah/tend/issues/28) asks for tend's code to
be laid out by behavior, so that "a contributor working on landing should be
able to find its merge decisions and acceptance tests directly".

- At `43758c3`, `main.go` has 2,220 lines, `tend_test.go` 2,406 and `flow.go`
  882. `main.go` holds dispatch, watch, status, locks, agents, Git and GitHub
  helpers, intake, the attempt, publication and landing, under section
  comments such as `// ---- landing` (`main.go:1981`).
- The issue observes that, at `0da41b8`, the six most recent commits touching
  Go code all edited both `main.go` and `tend_test.go`, for unrelated changes
  to startup notices, review permissions, scheduling and landing.
- The issue asked to wait until the change "Recover with no one there"
  (`261122-hjhmn`) had merged and its scenarios passed, as a behavioral
  baseline. It is archived at `43758c3`
  (`grove/changes/archive/261122-hjhmn-recover-with-no-one-there/`), and
  `go test ./...` passed on `43758c3` when this proposal was written.

## Intended outcome

Each behavior of tend has a source file and a test file named for it, in the
same `package main`, so the next change to a behavior has an obvious place
for its code and its acceptance test. tend behaves as it did: the same
commands, output, exit statuses and run-folder records, and the same test
and scenario names.

The files, from the issue's table:

| Area | Files |
| --- | --- |
| CLI and attempts | `main.go` for dispatch and startup, `attempt.go` for starting and resuming builds, `worktree.go` for the checkout's lifecycle |
| Watch and status | `watch.go` for scheduling and reruns, `status.go` for reporting |
| Delivery | `publication.go` for publishing a pull request and answering it, `landing.go` for landing locks, reviewed heads, checks and merge decisions |
| Intake and terminals | `intake.go` and `request.go` for proposal intake and what trackers share, `jira.go` for Jira, `terminal.go` for the terminal contract beside `herdr.go` and `rex.go` |
| Flows | `flow.go` for definitions and validation, `flow_load.go` for revision lookup and the built-in defaults, `flow_run.go` for execution and recovery |
| Tests | `scenario_test.go` for the shared CLI fixtures, and a file for each behavior, such as `landing_test.go` and `watch_test.go` |

## Scope

### In scope

- Moving the declarations of `main.go` and `flow.go` into the files above.
- Moving the fixtures and scenarios of `tend_test.go` into
  `scenario_test.go` and a test file for each behavior.
- Edits to the code where the split needs them, with behavior unchanged.
- The `context` line of `grove.yaml` and the README paragraph that name
  `tend_test.go`.

### Out of scope

- Any change in what tend does. No maintained spec changes, so this change
  has no change specs.
- An `internal/flow` package, the issue's third step. It is a later proposal,
  once the split shows whether the boundary holds.
- New packages or folders of any kind: the root stays buildable with
  `go build -o bin/tend .`, as the release configuration (`.goreleaser.yaml`,
  `main: .`, `-X main.version`) and the test harness expect.
- Renaming tests or scenarios, or adding per-area tests such as `TestWatch`.
- Changing the path or format of any record in a run folder or in
  `TEND_STATE`.

## Decisions

The owner answered the round before writing with: "yes to all, but we can
make changes to the code if needed". Each decision below is a
recommendation that answer accepted, the second as the owner changed it.

1. **One change, a task for each area; `internal/flow` left out.** Owner:
   "yes to all". The moves all cut the same two files, `main.go` and
   `tend_test.go`, so separate changes would conflict with each other, and
   whether flow validation warrants a package is a question for after the
   split.
2. **Code moves as it is by default, and may be edited where needed.** The
   round recommended moves only, with shared reader and writer helpers, such
   as for the `approves` marker, waiting for a follow-up. Owner: "yes to all,
   but we can make changes to the code if needed". So an edit is allowed,
   and behavior still may not change. How far "if needed" reaches is open
   question 1.
3. **`TestTend`, `TestLanding` and `TestFlows` and every scenario name are
   kept.** Owner: "yes to all". Every documented selector keeps working, in
   the README, in `grove.yaml` and in the archived changes'
   `implementation.md` files, at the cost of per-area test names such as
   `TestWatch`.
4. **The issue's file table is used as written, and which file each
   declaration goes in is the implementer's choice.** Owner: "yes to all".
   The table names the files; placing about 150 declarations and every
   scenario one by one is not worth deciding here.
5. **The `grove.yaml` context line and the README are updated where they
   name `tend_test.go`.** Owner: "yes to all". Both become false once that
   file is gone.

## Open questions

None must be answered before review. Each has a recommended answer that
`/grove-implement` builds unless the owner says otherwise.

1. **How far does "if needed" reach?** Decides T-01 to T-06. Recommended:
   an edit is needed when a declaration cannot move as it is, or reads wrongly
   where it lands: a function that spans two areas is split, a comment that
   points at its old neighbours is reworded, an import list follows its
   code. Sharing one implementation between a record's writer and its
   readers, as the issue suggests for the `approves` marker (written at
   `main.go:1600`, read at `main.go:467`), is allowed where the two now sit
   in different files and the edit is a few lines, and is otherwise left for
   a follow-up. Reason: review of this change rests on most of the diff
   being moved lines, and every other edit is recorded with its reason.
2. **Where do the helpers every area uses live?** Decides T-02. The table
   gives no file to the command, JSON and reporting helpers (`main.go:54` to
   `main.go:248` at `43758c3`). Recommended: those that more than one area
   uses stay in `main.go`, and one that a single area uses moves with that
   area, as `needsYou` does with watch, its only caller (`main.go:448`).
   Reason: the shared ones are the vocabulary of every other file and change
   rarely, and a file made only to hold them adds a place to look without
   adding a behavior.
3. **Which test files, beyond the two the issue names?** Decides T-07.
   Recommended: one for each source file whose behavior has scenarios:
   `attempt_test.go`, `worktree_test.go`, `watch_test.go`, `status_test.go`,
   `publication_test.go`, `intake_test.go`, `jira_test.go` and
   `terminal_test.go`, with `landing_test.go`, and `flow_test.go` left one
   file. Reason: a contributor looks for the test file beside the source
   file, and `flow_test.go` has 530 lines.
4. **Routine choices.** Decide no task's result: how `TestTend` gathers its
   scenarios from several files, the names of the scenario lists, the order
   of declarations and scenarios within a file, and where the section
   comments go. Recommended: whatever keeps each file reading top to bottom,
   with each scenario list declared in its own file.
5. **Do flow definition and validation warrant an `internal/flow` package?**
   Blocks no task; it can wait until this change is built. The issue
   suggests a boundary that accepts YAML and a prompt-existence lookup and
   returns a validated definition or a diagnostic, and notes that YAML error
   formatting refers to types in `main`. Recommended: decide it in a
   proposal of its own, from how `flow.go` reads once `flow_load.go` and
   `flow_run.go` are apart from it.

## Related documents

- Requirements: none. This change alters no requirement of a maintained
  spec.
- Tasks: [tasks.md](tasks.md)
