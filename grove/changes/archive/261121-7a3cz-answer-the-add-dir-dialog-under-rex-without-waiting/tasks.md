## T-01 — The add-dir dialog is seen as soon as it shows

Result: under rex, an agent that starts at Claude Code's dialog for adding
the run folder is answered without tend waiting out the time it allows for
the agent's first report.

Files:
- `rex.go`
- `dev/test.sh`

Acceptance:
- The `rex-add-dir` scenario takes under a minute, where it took about two,
  and a check in it fails if it does not.
- `dev/test.sh` passes.
