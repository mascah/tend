# Implementation: Several attempts at once

<!-- Progress on change 261122-9vbmc, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — A lock for each work item

Status: done

Each work item of a repository has its own lock, `TEND_STATE/locks/<cksum of
the repository's path>-<item>`, and an intake pass takes the item `intake`
(`takeLock`, `lockFile`, `held` in `main.go`). Exit 4 names the holder's pid,
work item and repository. A runner-wide `TEND_STATE/setup` lock is held while
a tend fetches, tidies, makes its worktree and claims, and is let go once the
run folder records the worktree (`settingUp`, `setUp`); an intake pass holds
it while it fetches and tidies, and around each request's worktree. Before
making a worktree, tend checks Grove's `claimable` for the piece and stops
with Grove's `verdict` when it is false; a claim Grove refuses anyway still
has its worktree and branch removed. `tidy` leaves the worktree of an attempt
or intake pass whose lock another tend holds, saying `a tend is on it`.
`tend status` reads the item's own lock. Usage is `usage: tend <repo> <id>`.

Tests in `tend_test.go`: `usage`, `at-once-two-items`, `at-once-intake`,
`at-once-second-intake`, `at-once-together`, `at-once-refused`,
`at-once-tidy`, with `lock`, `status-interrupted`, `tidy`,
`tidy-superseded`, `issue-build`, `remote-claim`, `remote-resume` and
`local-resume` unchanged; all pass.

Also changed: `testdata/bin/herdr` — `HOLD` holds an agent's prompt, and `WORKTREE_HOLD` a new worktree once made, until a file exists, so a test runs one tend while another is on its turn or setting up.

Check: built on main at 6bd2c29, where 261122-nh8aa (build an approved change by name) and 261122-fh9h2 (requested changes count only from people with write access) changed `main.go` and `tend_test.go` since this change's base. The claim path a change in no cycle shares is the one the `claimable` check and the cleanup cover; fh9h2's write-access filter is kept in watch's `rerun`.

Check: two tends started in the same second made the same `tend-HHMMSS` worktree, so the temporary branch is now `tend-HHMMSS-<pid>`. The maintained attempts spec names `tend-HHMMSS` in "Carry on the attempt or claim the cycle", so the change was revised on this branch with Grove's `instructions propose`: that requirement is MODIFIED in the change's attempts spec, the proposal's scope says so, T-01's acceptance names its scenarios, and the change is re-based to 6bd2c29.

Check: a lock is keyed by the repository's path as tend resolves it, so one repository reached by two spellings of its path (through a symlink) gets two locks.

Check: T-01 and T-02 are in one commit (c88ac7d): the tests of each need the other, as a second attempt during a turn is stopped without T-02 and cannot run without T-01.

### T-02 — Containment allows for tend's other attempts

Status: done

`contained` in `main.go` first waits for any tend setting up (takes and lets
go the setup lock), then treats as another's the worktree each other run or
intake folder records and the branch that worktree is on; anything else new
still stops the attempt.

Tests in `tend_test.go`: `at-once-during-turn`, `at-once-stray`,
`at-once-setting-up`, with `stray-worktree` and `stray-branch` unchanged; all
pass. With the wait removed, `at-once-setting-up` fails.

Check: tend cannot tell whether an implementer wrote into another attempt's worktree, as decision 4 says; that would show in the other attempt's own diff and review. Marked `ponytail:` in `contained`.

### T-03 — The flow file holds watch's count

Status: done

`flowFile` in `flow.go` has `watch` with `at_once`, checked by
`watchSetting.fault` wherever the flow file is read; `tend flow` prints it
first, headed by where it came from; `watchCount` gives watch the count at the
target, or none, without stopping.

Tests in `flow_test.go`: `flow-watch-count`, `flow-watch-zero`, and
`flow-print`, which now checks that a saved file without the setting prints no
`watch`; `TestFlowFaults` reads `watch` faults through `faultOf`; all pass.

Check: an `at_once` that is not a whole number (such as `two` or `1.5`) is refused as "gives watch an at_once of two, where at_once is a whole number of at least 1", and `watch: {}` as "gives watch no at_once"; a bare `watch:` with no value is read as no setting.

### T-04 — Watch keeps a count of attempts going

Status: done

In a repository whose flow file sets the count (lowered by `AT_ONCE` when that
is a whole number), `watch` in `main.go` runs `keep`: it counts the
repository's held work-item locks but `intake`, runs tend again in the
background on each attempt `rerun` picks, then starts each claimable cycle not
done and change in no cycle, oldest `created` first, saying `starting <id> in
<repo>`, while the count allows. Intake still runs in the pass, outside the
count. With `WATCH_ONCE`, watch waits for every tend it started, then exits.
A repository without the setting takes the old path unchanged.

Tests in `tend_test.go`: `watch-at-once-two`, `watch-at-once-one`,
`watch-at-once-overlap`, `watch-at-once-needs-you`, `watch-at-once-by-hand`,
`watch-at-once-cycle`, `watch-at-once-requested`, `watch-at-once-no-setting`,
`watch-at-once-lower`, `watch-at-once-higher`, `watch-at-once-zero`,
`watch-at-once-raised`, with `watch`, `watch-intake`, `watch-unattended` and
`watch-unattended-merged` unchanged; all pass.

Check: watch starts claimable pieces one at a time, waiting until each started tend's claim shows in `grove changes --json` (or that tend exits) before it chooses the next, because Grove's `claimable` read once would have watch start a piece that overlaps the one just started, to be refused and notify a person. It does not wait for the attempt itself. With the wait removed, `watch-at-once-overlap` fails.

Check: where the flow file sets a count but is not valid, watch says why and starts and runs again nothing in that repository; the change does not say what watch does then, and every tend run there would stop on the same fault.

Check: "oldest first" for attempts run again is the run folders' name order, which a Grove attempt's date-first ID makes oldest first, though attempts of the same day go by the ID's random part; an issue build's folder sorts after them. Marked `ponytail:` in `keep`.

Check: under the setting, attempts and runs again go in the background and intake runs after them in the same pass, as before, waiting for its own pass to end.

### T-05 — The README says how to run several attempts at once

Status: done

`README.md`: "Run it" gives `tend <repo> <id>`, says exit 4 means the work
item is taken and that tends on different work items run at once, setting up
one at a time; "How work reaches tend" says what watch starts with and without
the setting; the flows section shows `watch: {at_once: N}`, what the count
counts, `AT_ONCE`, that a stopped attempt frees its slot and that
interrupting an agent in its pane parks its attempt; the temporary branch is
`tend-HHMMSS-<pid>`.

Check: the README's "Not yet shown" list no longer names two cycles at once, as the task asks, but several attempts at once has been shown only by the acceptance tests with stand-in agents, not with live ones; a reviewer may want a line saying so.

Check: that interrupting an agent in its pane parks its attempt was not seen with a live agent: an interrupted turn ends with the candidate unfinished, which stops tend for a person as any unfinished turn does.

## Scenarios

- attempts "A second tend on the same work item": tend_test.go `lock`, passed
- attempts "The holder is killed": tend_test.go `lock`, passed
- attempts "Two work items at once": tend_test.go `at-once-two-items`, passed
- attempts "Intake beside a build": tend_test.go `at-once-intake`, passed
- attempts "A second intake pass": tend_test.go `at-once-second-intake`, passed
- attempts "Two started together": tend_test.go `at-once-together`, passed
- attempts "A turn ends while another attempt sets up": tend_test.go `at-once-setting-up`, passed
- attempts "A piece that overlaps one in flight": tend_test.go `at-once-refused`, passed
- attempts "A cycle reaches a verdict": tend_test.go `remote-claim`, passed
- attempts "The reviewer leaves no valid verdict": tend_test.go `new-incomplete`, passed
- attempts "No work item named": tend_test.go `usage`, passed
- attempts "A rerun after a stop": tend_test.go `remote-resume`, passed
- attempts "Without origin": tend_test.go `local-resume`, passed
- attempts "The temporary branch goes": tend_test.go `remote-claim`, passed
- attempts "The implementer adds a worktree": tend_test.go `stray-worktree`, passed
- attempts "A rerun finds it still there": tend_test.go `stray-worktree`, passed
- attempts "The implementer makes a branch with commits": tend_test.go `stray-branch`, passed
- attempts "Another attempt starts during the turn": tend_test.go `at-once-during-turn`, passed
- attempts "A stray worktree beside another attempt's": tend_test.go `at-once-stray`, passed
- flows "A count and no flow": flow_test.go `flow-watch-count`, passed
- flows "No count": flow_test.go `flow-print`, passed
- flows "A count of 0": flow_test.go `flow-watch-zero`, passed
- flows "The implementer raises the count": tend_test.go `watch-at-once-raised`, passed
- watch-and-status "Two changes that share nothing": tend_test.go `watch-at-once-two`, passed
- watch-and-status "A count of 1": tend_test.go `watch-at-once-one`, passed
- watch-and-status "Two that overlap": tend_test.go `watch-at-once-overlap`, passed
- watch-and-status "An attempt that needs a person": tend_test.go `watch-at-once-needs-you`, passed
- watch-and-status "A run by hand counts": tend_test.go `watch-at-once-by-hand`, passed
- watch-and-status "A cycle not marked": tend_test.go `watch-at-once-cycle`, passed
- watch-and-status "Changes requested go first": tend_test.go `watch-at-once-requested`, passed
- watch-and-status "No setting": tend_test.go `watch-at-once-no-setting`, passed
- watch-and-status "A pull request closed under a tend": tend_test.go `at-once-tidy`, passed (the stand-in marks the pull request merged; tend treats merged and closed alike)
- watch-and-status "A tend tidies its own finished work": tend_test.go `issue-build`, passed
- watch-and-status "A lower number on this machine": tend_test.go `watch-at-once-lower`, passed
- watch-and-status "A higher number on this machine": tend_test.go `watch-at-once-higher`, passed
- watch-and-status "Nothing new": tend_test.go `watch-at-once-zero`, passed

## Reconciliation

- attempts: applied to grove/specs/attempts/spec.md, its Boundaries now naming the lock of each work item in place of the runner lock
- flows: applied to grove/specs/flows/spec.md
- watch-and-status: applied to grove/specs/watch-and-status/spec.md, its Purpose now defining an attempt being worked on, as the change spec's preamble does
