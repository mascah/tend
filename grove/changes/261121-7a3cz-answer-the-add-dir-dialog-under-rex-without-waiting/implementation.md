# Implementation: Answer the add-dir dialog under rex without waiting a minute

## Tasks

### T-01 — The add-dir dialog is seen as soon as it shows

Status: done

`start` in `rex.go` now returns from its wait for the agent's first report
when the screen holds "yes, add it" as well as "trust", so `start` in
`main.go` answers the dialog at once. `dev/test.sh` gained the check
"rex-add-dir: without waiting out the agents' first reports", which fails if
the scenario's run takes a minute or more.

Checks run on 2026-11-21: `gofmt -l .` and `go vet ./...` clean; `dev/test.sh`
all passed, 327 checks. `rex-add-dir` took 5 seconds where it took 124, and
`rex-add-dir-refused` 6 where it took 127; the whole run took 196 seconds
where it took 398.

## Scenarios

The change has no change spec, so no scenario of its own. The terminals
scenarios it touches still pass:

- terminals "Adding the run folder under rex": dev/test.sh rex-add-dir, passed
- terminals "A dialog that refuses the first answer under rex": dev/test.sh rex-add-dir-refused, passed
- terminals "A new folder under rex": dev/test.sh rex-trust, passed
- terminals "New hooks under rex": dev/test.sh rex-hooks, passed

## Reconciliation

No change spec to apply.
