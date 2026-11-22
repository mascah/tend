# Implementation: Build an approved change by name

<!-- Progress on change 261122-nh8aa, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — tend builds, reports and tidies a change in no cycle

Status: done

Evidence: in `main.go`, `itemRun` finds the newest record for a work item,
either a cycle's or one naming the change alone and no cycle. `attempt` and
`status` use it in place of `firstRun(proj, "cycle", …)`, and `status` prints
the work item it was given. `workState` says whether Grove knows a cycle or a
change and whether it is done (a cycle done, a change archived). The
overview's attempt filter and `superseded` (tidying) use it in place of
`cycleOpen`. The overview's heading is now "Approved changes in no cycle:",
and each change shows Grove's `verdict`. The heading over attempts reads "for
cycles and changes not done and issues". The proposal review now ends "Once
merged, the change can be built by its ID, `tend <repo> <change-id>`, or
planned into a cycle with others: `/grove-plan`." Claiming, carrying on, the
finished check and the `Grove:` trailer already worked from the record's
`changes`, so they are unchanged.

New scenarios in `tend_test.go`, run against Grove 0.14.0: `change-claim`,
`change-resume`, `change-waits`, `change-in-cycle`, `change-overview` and
`change-tidy`. `overview` gets the new heading and a verdict check; `propose`
checks the new closing sentence. Mutation check: with `itemRun` back to
cycle-only and `workState` blind to changes, `change-claim`, `change-resume`,
`change-overview` and `change-tidy` fail. `go test ./...` passes.

Also changed: `testdata/bin/turn.sh` — the stand-in implementer always wrote
`Grove: 261015-c0000` as the commit message's last line. It now writes the
trailer the prompt asks for, so a change's pull request can be checked to end
`Grove: 261016-c0001`.

Check: when Grove refuses a claim, tend now removes the empty worktree and its
`tend-HHMMSS` branch it had just made, then stops with Grove's sentence.
Before, the worktree stayed behind. A rerun in the same second then hit the
same name and failed (exit 1, "exit status 255") instead of claiming. Neither
the change nor the attempts spec says what happens to that worktree, and
refusals are now an everyday case (a change that waits, a change in a cycle),
so I chose to remove it. `change-in-cycle` checks that it is gone.

Check: Grove 0.14.0 gives a claimed cycle the verdict "<cycle> is in flight
in attempt …", not "claimed by attempt …". That is why `overview` failed on
main before this change. The revision on this branch rewrites the "A
repository with work at every stage" scenario in
`specs/watch-and-status/spec.md` to quote it, and re-bases the change on
b9b2d31. That commit (261122-fh9h2) changed `main.go`, `tend_test.go` and
`README.md` only to filter change requests by write access, which this task
neither touches nor depends on.

Check: the usage line still reads `usage: tend <repo> <cycle-id>`, as the
proposal's out-of-scope list says. Only the comment at the top of `main.go`
names `<change-id>`.

### T-02 — Grove 0.14.0 is the Grove tend needs

Status: done

Evidence: `.github/workflows/ci.yml` installs Grove `v0.14.0`. `gh release
view v0.14.0 -R mascah/grove` shows the release, with 5 assets. The tests ran
locally against `grove 0.14.0`. In `README.md`, "What it asks of Grove" names
`grove claim <change-id>` and Grove 0.14.0 or later. "Run it" lists
`tend <repo> <change-id>` and says a cycle is needed only to build several
changes together. "How work reaches tend" item 2 says the same. `tend watch`
is still described as starting only cycles marked `unattended`, and "Not
picked up" adds approved changes in no cycle. The README's intro, its status
paragraphs and its "Is there an attempt?" row now name changes too, because
they described cycles only.

Check: CI on Grove v0.14.0 is not seen until this branch's pull request runs.

## Scenarios

- attempts "A change claimed by name": tend_test.go TestTend/change-claim, passed
- attempts "A rerun after a stop": tend_test.go TestTend/change-resume, passed
- attempts "A change that waits": tend_test.go TestTend/change-waits, passed
- attempts "A member of a cycle": tend_test.go TestTend/change-in-cycle, passed
- attempts "Its status": tend_test.go TestTend/change-claim ("status finds the attempt", "and what it recorded"), passed
- attempts "Its pull request": tend_test.go TestTend/change-claim ("one pull request is opened, onto the target", "its description ends with the change"), passed
- attempts "The change landed some other way": tend_test.go TestTend/change-tidy, passed
- watch-and-status "A repository with work at every stage": tend_test.go TestTend/overview, passed
- watch-and-status "A change that waits": tend_test.go TestTend/change-waits ("the overview gives its verdict"), passed
- watch-and-status "An attempt at a change": tend_test.go TestTend/change-overview, passed
- watch-and-status "A cycle done": tend_test.go TestTend/overview ("an attempt whose cycle is done is left out"), passed
- watch-and-status "Jira requests": tend_test.go TestTend/jira-propose ("the overview lists a Jira request waiting, not the handled ones"), passed (unchanged by this change)

## Reconciliation

- attempts: applied to grove/specs/attempts/spec.md
- watch-and-status: applied to grove/specs/watch-and-status/spec.md
