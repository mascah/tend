# Implementation: Hold a candidate that touches a protected path for the owner

<!-- Progress on change 261125-6dvbs, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — Protected paths read from the target's flow file

Status: done

The flow file gains `protect`, a list of paths (`flow.go`: `flowFile.Protect`,
`defaultProtect`, `protectFault`, `protected`, `entryRE`, `covered`).
`flowsAt` and `watchCount` reject a `protect` with an empty entry, naming
`tend/flow.yaml` and `protect`; `protectedAt(rev)` gives the protected paths
in effect on the target with where they came from (`flow_load.go`), read
only through `flowsAt`, so never from the attempt's branch; `tend flow`
prints `protect` after `watch`, saying whether it is tend's own list or the
project's.

Checks: `go test -run 'TestProtect|TestFlowFaults|TestLandFaults'` and
`go test -run 'TestFlows/flow-(protect|print|watch)'`, passed. `tend flow .`
on this repository prints tend's own list.

Check: a project's list replaces the default and `tend/` is added unless the
list names `tend/` itself, as open question 2 recommends; the default list is
open question 4's, without `grove/cycles/`.
Check: matching follows open question 3. Beyond it: a leading `/` changes
nothing (`/schema.sql` means the root's `schema.sql` only, as any entry with
a `/` in it does), and an entry that is only `/`, or holds `//`, is not
valid, since it names no path. `*` is the only wildcard; any other character
matches itself.
Check: `protect:` with no value (YAML null) counts as no `protect`, so tend's
own list applies; `protect: []` is the project's empty list, leaving only
`tend/` protected.
Check: `tend flow` prints the list under the comment `# protect: built in,
tend's own list.` or `# protect: tend/flow.yaml on <rev>, the project's own,
with tend/ always protected.`, a wording open question 6 leaves to the
implementer.

### T-02 — Hold at land a head that touches a protected path

Status: done

`landing.go` gains `changedPaths(h)` (`git diff --no-renames --name-only -z
<target>...<head>`: the paths that differ from where the head and the
fetched target meet, a rename under both names) and `hold(h)`, which reads
the protected paths afresh with `protectedAt` on the fetched target, and
where the head touches one, lands it only if publication's `approved()`
finds the newest approving or change-requesting review by a person with
write access, on that head or a head it is carried from, approving;
otherwise it stops (exit 3) naming the paths. `landStep` calls it in the
first round after fetching the target, before any merge-in, in every round
where the target moved, before merging it in, and just before asking
GitHub to merge.

Checks: `go test ./...`, passed (the existing landing tests unchanged). The
new `TestLanding/protect-*` scenarios fail with the hold disabled, except
`protect-supervised` and `protect-target-own`, which show it does not hold
what it should not; `protect-target-own` fails when the changed paths are a
two-dot diff against the target, which would count the target's own commit.

Check: open question 5 is built as recommended: a held head is a stop for a
person (exit 3, status needs-you); `tend watch` leaves it alone, so a person
who approves it runs tend again. The watch-and-status spec is unchanged.
Check: the stop reads `<branch> at <head> changes the protected path(s) <paths>,
so tend lands it only once a person with write access approves that head on
<pull request>`, a wording open question 6 leaves to the implementer.
Check: the check before the merge reuses the target the round fetched at its
start, not a second fetch; between the two only the check reads run. A
target that moves after that makes GitHub refuse the merge, and the next
round fetches it and holds again before merging it in.
Check: `hold` sets the global `head` to the landing head before calling
`approved()`, which reads it; `head` is the branch's HEAD at that point
either way.
Also changed: `testdata/bin/turn.sh` — the implementer stand-in runs
`BUILD_ALSO` before committing its build and `FIX_ALSO` in a repair or fix,
so tests can have a head edit chosen files.
Also changed: `guide/flows-reference.md` — its table of top-level keys said
any key but `build`, `propose` and `watch` is refused; it now lists
`protect`, its fault, and a section on protected paths.
Also changed: `guide/concepts/landing.md` — its steps and diagram said
every reviewed head with passing checks lands; they now show the hold.
Also changed: `guide/setup.md` — it said every head that passed review lands
in the autonomous shape, and that tend reads an approval only at `publish`.
Also changed: `guide/built-vs-planned.md` — it listed protected paths as
wholly planned; it now says the landing hold is built.

## Scenarios

- flows "The default list": `flow_test.go` `TestFlows/flow-protect-default`, passed
- flows "A project's own list": `flow_test.go` `TestFlows/flow-protect-own`, passed
- flows "An empty entry": `flow_test.go` `TestFlows/flow-protect-empty`, passed
- flows "The implementer's list is not read": `landing_test.go` `TestLanding/protect-unprotected`, passed
- landing "An autonomous flow and an edit to the flow file": `landing_test.go` `TestLanding/protect-flow-file`, passed
- landing "A person approves the held head": `landing_test.go` `TestLanding/protect-flow-file` (its last run), passed
- landing "A stranger approves the held head": `landing_test.go` `TestLanding/protect-flow-file` (its second run), passed
- landing "No protected path": `landing_test.go` `TestLanding/land-landed` (an autonomous flow, a head touching no protected path), passed
- landing "A nested agent instruction file": `landing_test.go` `TestLanding/protect-nested`, passed
- landing "A protected file moved away": `landing_test.go` `TestLanding/protect-renamed`, passed
- landing "The candidate unprotects a path": `landing_test.go` `TestLanding/protect-unprotected`, passed
- landing "The target moves in a carried head": `landing_test.go` `TestLanding/protect-carried` and `TestLanding/protect-target-own`, passed
- landing "A head made at fix": `landing_test.go` `TestLanding/protect-fix`, passed
- landing "An approval before the implementer's new head": `landing_test.go` `TestLanding/protect-resolved`, passed
- landing "A supervised flow": `landing_test.go` `TestLanding/protect-supervised`, passed
- landing, decided again before the merge from the target read afresh: `landing_test.go` `TestLanding/protect-read-again`, passed

## Reconciliation
