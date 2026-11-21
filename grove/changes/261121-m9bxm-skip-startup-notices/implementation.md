# Implementation: Skip startup notices

<!-- Progress on change 261121-m9bxm, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — An agent's startup notice is skipped

Status: done

Evidence: `start` in `main.go` presses Esc alone at a screen whose text holds
`esc skip`, after the hooks review and the trust dialogs are checked, says
"tend: skipped <agent>'s startup notice" on standard error, and waits for the
agent as after a trust dialog. It answers up to four screens in turn, not
three, so a notice, a trust dialog and a refused answer still fit. The rex
stand-in (`testdata/bin/rex`) gains `NOTICE`, which starts agents at Codex's
"Update available" notice that only Esc clears, before any trust dialog. New
tests in `tend_test.go`: `rex-notice` (notice then trust dialog; each agent's
first key is `esc`, one "skipped" line per agent, passing verdict) and
`rex-unknown-screen` (exit 3, "did not start", no key sent). `rex-notice`
fails with the `main.go` and `rex.go` changes stashed. The README's "What it
does to your agents" says tend skips such a notice by Esc alone and stops at
any other screen. `go vet ./...` and `go test ./...` pass.

Also changed: `rex.go` — `rex.start` stops waiting for a first report at a screen tend answers. Without `esc skip` there, it would wait out the full 60 seconds at a notice before `start` saw it.

Check: `main.go`, `testdata/bin/rex`, `testdata/bin/turn.sh`, `tend_test.go` and `README.md` changed on main after base 0a4e821, when 261121-gapde (Go tests and stand-ins in `testdata/bin`), #13 (`rex.start` returns early at a dialog) and #15 (flows) landed. The tests are built on `tend_test.go` and the stand-ins as they are now. #13's early return needed `esc skip` added (see `rex.go` above). `turn.sh` needed no change.
Check: after Esc, tend waits up to 60 seconds for the agent to report ready before it looks at the screen again. A trust dialog that follows the notice, as in the owner's run, is answered only once that wait runs out. This is how `start` already treats one dialog after another, and the change does not ask to shorten it. As a result, `rex-notice` takes about two minutes. Shortening the wait would change how every dialog is answered, including the refused-answer case, so I left it alone.
Check: under Herdr, the notice is caught by the same screen check in `start` once Herdr reports `agent_not_ready`, but no test drives it: the change's scenarios are under rex, and the Herdr stand-in shows no startup screen.

## Scenarios

- terminals "A new folder under rex": tend_test.go TestTend/rex-trust, passed
- terminals "Adding the run folder under rex": tend_test.go TestTend/rex-add-dir, passed
- terminals "A dialog that refuses the first answer under rex": tend_test.go TestTend/rex-add-dir-refused, passed
- terminals "New hooks under rex": tend_test.go TestTend/rex-hooks, passed
- terminals "A startup notice under rex": tend_test.go TestTend/rex-notice, passed
- terminals "A screen tend does not know under rex": tend_test.go TestTend/rex-unknown-screen, passed

## Reconciliation

<!-- One line per change spec, once applied to its maintained spec:

    - record-editing: applied to specs/record-editing/spec.md
-->
