## T-01 — Tend delivers; the implementer only commits

Result: the implementer is prompted under `--runner`, both Claude agents
start with tend's terms and the push guard, each implementer turn is
checked against the layout recorded before its prompt, the timeout is 4
hours, and tidy removes superseded attempts, as the attempts, terminals and
watch-and-status change specs say.

Files:
- `main.go`
- `dev/test.sh`
- `README.md`

Interfaces:
- `TEND_STATE/runs/<attempt>/layout`: the repository's worktrees and local
  branches, written just before each implementer prompt.

Constraints:
- No status read of the implementer beside `ask`'s own before a prompt, so
  a working agent is never prompted twice.

Acceptance:
- `dev/test.sh` passes, with scenarios for: the `--runner` prompt; the
  implementer's and drafter's start flags; a worktree added, also on a
  rerun; a branch made with commits; a superseded attempt left for a
  worktree inside it and for a working implementer, then removed; an
  unfinished cycle's attempt left.
- The README names `--runner`, the terms, the guard, the check, the
  ceiling of one OS user and the 4-hour default.
