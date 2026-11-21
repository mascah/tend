# Answer the add-dir dialog under rex without waiting a minute

## Observed

Under rex, an agent that starts at Claude Code's dialog for adding the run
folder is answered only after 60 seconds.

- `start` in `rex.go` at ed385ad (lines 137-146) waits up to 60 seconds for
  the agent's first report, and returns early only when the screen holds the
  word "trust".
- The dialog for adding the run folder reads "Yes, add it", which holds no
  "trust". So the wait runs out before `start` in `main.go` at ed385ad (line
  663), which answers that dialog, sees it.
- `dev/test.sh` at ed385ad, timed on 2026-11-21: `rex-add-dir` took 124
  seconds and `rex-add-dir-refused` 127, two agents waiting 60 seconds each.
  The other 70 scenarios took 147 seconds together.

The trust dialogs of a new folder ("Yes, I trust this folder", "Trust and
continue") and Codex's review of hooks all hold "trust", and are seen at once.

## Expected

The dialog is answered as soon as it shows, as the trust dialog is. The owner,
on 2026-11-21: "We should fix the rex hang up as well".

## Reproduction

1. At ed385ad, run `dev/test.sh` and note the time between the last check of
   `rex-trust` and the first of `rex-add-dir`: about two minutes.
2. Or, under rex, run tend on a cycle whose run folder Claude Code has not
   added: the implementer sits at the dialog for a minute before tend
   answers it.

## The requirement it breaks

None. "The trust dialogs are answered" in `grove/specs/terminals/spec.md`
requires the dialog to be answered and says nothing of how soon. No new
requirement was agreed; this fixes the wait and leaves the spec as it is.

## Open questions

None.

## Related documents

- Tasks: [tasks.md](tasks.md)
