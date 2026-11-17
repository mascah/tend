# Implementation: The implementer leaves delivery to tend

## Tasks

### T-01 — Tend delivers; the implementer only commits

Status: done

- `main.go`:
  - `implementer()` starts Claude with `--disallowedTools "Bash(git
    push:*)" "Bash(gh pr:*)"` and `--append-system-prompt` holding tend's
    terms. It names the worktree's branch, and for `impl-` agents adds
    "Commit your work on this branch."
  - `ask()` writes `layout` just before an `impl-` prompt.
  - `checked()` wraps the build, repair and request turns, and
    `contained()` compares the layout after each and on every run before
    review.
  - The build prompt is `/grove-implement --runner <cycle>`.
  - `timeout()` defaults to 14400000.
  - `tidy()` gains `superseded()` and its two guards.
- `dev/test.sh`:
  - the stand-in implementer takes `STRAY=worktree|branch`;
  - scenarios: runner, stray-worktree, stray-branch, drafter-terms,
    tidy-superseded, tidy-unfinished.
  - `dev/test.sh`: all passed.
- Checked by hand in Herdr on the MacBook: a Claude started through
  `herdr agent start ... -- --disallowedTools "Bash(git push:*)" "Bash(gh
  pr:*)" --append-system-prompt "<text with ;, (), * and $HOME>"` quoted
  its system prompt word for word, and `git push --dry-run` was denied.

Check: tend's terms ride in the system prompt instead of the end of each prompt, as design A1 wrote, because text after `/grove-implement` is its arguments.
Check: the first try at recording the layout read the agent's status before `ask()`, and the rex stand-in caught that it prompted a working agent twice. Recording inside `ask()`, beside the mark, avoids the extra read.
Check: under rex, tidy's working-implementer guard may not see an agent before its place is opened, so it reads as not working.

## Scenarios

- attempts "The implementer adds a worktree": dev/test.sh stray-worktree, passed
- attempts "A rerun finds it still there": dev/test.sh stray-worktree, passed
- attempts "The implementer makes a branch with commits": dev/test.sh stray-branch, passed
- attempts "The documents moved on the target": dev/test.sh, unchanged, passed
- attempts "Under --runner": dev/test.sh runner, passed
- attempts "The implementer outlives a timeout", "The reviewer outlives a timeout": dev/test.sh, unchanged, passed; the 4-hour default is unverified by a test, as it is a constant
- terminals "The implementer's flags": dev/test.sh runner and publish, passed
- terminals "The drafter's flags": dev/test.sh drafter-terms, passed
- watch-and-status "A merged attempt and an open proposal", "Uncommitted files", "A merged issue build": dev/test.sh tidy and the issue scenarios, unchanged, passed
- watch-and-status "The cycle landed some other way": dev/test.sh tidy-superseded, passed
- watch-and-status "The cycle is not done": dev/test.sh tidy-unfinished, passed

## Reconciliation

- attempts: applied to grove/specs/attempts/spec.md
- terminals: applied to grove/specs/terminals/spec.md
- watch-and-status: applied to grove/specs/watch-and-status/spec.md
