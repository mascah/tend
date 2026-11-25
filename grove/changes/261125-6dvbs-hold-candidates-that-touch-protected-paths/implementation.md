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

Status: not started

## Scenarios

## Reconciliation
