## T-01 — A lock for each work item

Result: tends on different work items, and an intake pass beside them, run at
the same time; tends that start together set up one at a time; and a refused
claim leaves nothing behind, as the attempts requirements "One tend for each
work item", "Attempts set up one at a time", "A refused claim leaves nothing
behind" and "Exit statuses" in this change say; and cleanup leaves the
worktree of an attempt a tend is on, as the watch-and-status requirement
"Tidying leaves what a tend is on" says.

Files:
- `main.go`
- `tend_test.go`

Delegated: test layout and private helpers; the names of the lock files under
TEND_STATE; the wording of one-sentence messages.

Interfaces: whether a tend holds a work item's lock, which `tend status` and
`tend watch` read.

Constraints: `tend status <repo> <item>` still says `tend is on it` only for
the item asked about; a lock goes when its holder exits or is killed.

Acceptance:
- Each scenario of the three ADDED attempts requirements but "A turn ends
  while another attempt sets up", and the watch-and-status scenario "A pull
  request closed under a tend", is a test that passes; "A tend tidies its
  own finished work" is the existing test of a merged issue build.
- `tidy` and `tidy-superseded` still pass: a tend still removes its own
  finished worktree, and an intake pass a merged proposal's.
- `lock` and `status-interrupted` still pass, and the usage test expects
  `usage: tend <repo> <id>`.
- The scenarios of the MODIFIED attempts requirement "Carry on the attempt
  or claim the cycle" still pass (`remote-resume`, `local-resume`,
  `remote-claim`), with the temporary branch named `tend-HHMMSS-<pid>`.

## T-02 — Containment allows for tend's other attempts

Result: an attempt is not stopped for a worktree or branch that another
attempt or intake pass of this runner made, and is still stopped for any
other, as the attempts requirement "Work stays on the attempt's branch" in
this change says.

Files:
- `main.go`
- `tend_test.go`

Delegated: test layout and private helpers.

Acceptance:
- The scenarios "Another attempt starts during the turn", "A stray
  worktree beside another attempt's" and "A turn ends while another attempt
  sets up" are tests that pass.
- `stray-worktree` and `stray-branch` still pass.

## T-03 — The flow file holds watch's count

Result: `watch.at_once` is read from the target's flow file, refused when it
is not valid, and printed by `tend flow`, as the flows requirement "The flow
file sets how many attempts watch keeps going" says.

Files:
- `flow.go`
- `flow_test.go`

Delegated: test layout and private helpers.

Interfaces: the count in effect for a repository's target, or none, which
`tend watch` reads.

Acceptance:
- The scenarios "A count and no flow", "No count" and "A count of 0" are
  tests that pass.
- A flow file saved from `tend flow`'s output in a repository without the
  setting is still valid and holds no `watch`.

## T-04 — Watch keeps a count of attempts going

Result: in a repository whose flow file sets the count, `tend watch` starts
what Grove calls claimable and keeps at most the count being worked on, as
the watch-and-status requirements "Watch keeps a count of attempts going" and
"The runner may lower the count" say; and the flows scenario "The implementer
raises the count" holds.

Files:
- `main.go`
- `tend_test.go`

Delegated: test layout and private helpers; the wording of one-sentence
messages.

Constraints: a repository without the setting is watched exactly as before.

Acceptance:
- Each scenario of the two watch-and-status requirements, and the flows
  scenario "The implementer raises the count", is a test that passes.
- `watch`, `watch-intake`, `watch-unattended` and `watch-unattended-merged`
  still pass unchanged.

## T-05 — The README says how to run several at once

Result: the README describes the lock, the `watch` setting and `AT_ONCE`.

Files:
- `README.md`

Acceptance:
- "Run it" says tends on different work items run at once, that exit 4 means
  the work item is taken, and gives the usage line as `tend <repo> <id>`.
- The flows section shows `watch: {at_once: N}`, says what watch starts with
  and without it, what the count counts, and that `AT_ONCE` lowers it.
- It says an attempt that stops for a person frees its slot, and that
  interrupting an agent in its pane parks its attempt.
- The "Next" list no longer names two cycles at once as untried.
