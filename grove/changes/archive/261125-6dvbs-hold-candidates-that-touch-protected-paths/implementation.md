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

Checks: `go test -run 'TestProtect|TestFlowFaults|TestLandFaults'` (with
`protect: [null]` and `protect: ["//migrations//"]` refused, both found
accepted by the independent review and fixed) and
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
Check: an entry ending in `/` covers the folder's own path as well as all in
it, so a gitlink added at `.claude` itself is held under `.claude/`. A path
Git gives does not say whether it is a gitlink or a file, so a folder entry
with `*` in its last segment also covers files at that level:
`.github/*/` covers `.github/CODEOWNERS`. That holds more for a person, never
less. The pull request's review found a gitlink at a protected folder's own
path landing unheld; shown by `TestProtectedPaths` and
`TestLanding/protect-submodule-folder`.
Check: a YAML alias in `protect` resolves to what it names, so
`protect: [&s schema.sql, *s]` is valid; an alias of a list or a map is
refused as any such entry is. The pull request's review found aliases
refused; shown by `TestProtectFaults`.
Check: `protect:` with no value, or `protect: null`, is not valid: the key is
there and holds no list, so tend exits 3 saying `tend/flow.yaml gives
protect no list` (`decode` notes the key, as YAML decodes it to no
`protect`; shown by `TestProtectFaults` and `TestFlows/flow-protect-empty`,
after the pull request's review found it falling back to tend's own list).
`protect: []` is the project's empty list, leaving only `tend/` protected. A null entry in the list (`protect: [null]`) is an empty
entry and not valid: `protect` decodes through its own type, `pathList`,
since YAML's decoding into a list of strings drops a null entry silently.
An entry that is a list or a map is not valid either. Shown by
`TestProtectFaults`, after an independent review found `[null]` accepted.
Check: `tend flow` prints the list under the comment `# protect: built in,
tend's own list.` or `# protect: tend/flow.yaml on <rev>, the project's own,
with tend/ always protected.`, a wording open question 6 leaves to the
implementer.

### T-02 — Hold at land a head that touches a protected path

Status: done

`landing.go` gains `changedPaths(h)` (`git diff --no-renames
--ignore-submodules=none --no-relative --name-only -z <target>...<head>`:
the paths that differ from where the head and the fetched target meet, a
rename under both names, a gitlink included whatever the repository's
config says, which an implementer's worktree shares and can set),
`cleanMerges(h)` (the head and each head it was made from by a clean merge
of the target, followed through Git alone: parents, `merge-tree`, ancestry
of the target) and `hold(h)`, which reads the protected paths afresh with
`protectedAt` on the fetched target, and where the head touches one, lands
it only if publication's `approvedOn` finds the newest approving or
change-requesting review by a person with write access, on any head of
`cleanMerges(h)`, approving. The hold never asks whether a head passed
review, so no verdict or other file in the run folder changes its outcome;
`carried`, which the review gate and the `tend/review` status still use,
is now `cleanMerges` cut at the first head that passed review. The pull
request's review found the first version following `carried`, which read
verdict files, and `git diff` honouring `diff.ignoreSubmodules`;
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
Check: just before GitHub is asked to merge, tend fetches the target again
rather than trusting the fetch at the start of the round. If the target moved
while the checks were read, tend goes back to merging it in, holding the head
first under the protected paths the moved target gives. Otherwise it holds the
head under those paths. Between that fetch and the merge request only the
reviews are read; a target that moves in that moment makes GitHub refuse the
merge if the branch rule requires an up-to-date branch, and tend fetches and
holds again before trying once more. The independent review found the first
version reused the round's earlier fetch, so a target that added protection
while the checks were read still got one merge request.
Check: publication's `approved()` is split so that `approvedOn(heads)` reads
the reviews for the heads it is given; `approved()` gives it `carried`'s
heads as before, and `hold` gives it `cleanMerges`.
Also changed: `publication.go` — `approvedOn` split out of `approved`, so the
hold can count approvals on heads followed through Git alone.
Found: the flows spec's scenario "The implementer's list is not read" says the head is held "as the landing change spec's" requirement says; folded into `grove/specs/flows/spec.md` word for word, as `grove archive` requires, it now names a change spec where it means the landing spec. Seen reconciling the specs.
Also changed: `testdata/bin/gh` — the GitHub stand-in runs `DURING_CHECKS`
once as check runs are first read, so a test can move the target while the
checks are read.
Also changed: `testdata/bin/turn.sh` — the implementer stand-in runs
`BUILD_ALSO` before committing its build and `FIX_ALSO` in a repair or fix,
so tests can have a head edit chosen files.
Also changed: `guide/flows-reference.md` — its table of top-level keys said
any key but `build`, `propose` and `watch` is refused; it now lists
`protect`, its fault, and a section on protected paths, a folder entry
covering the folder's own path.
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
- landing "The target moves in a carried head": `landing_test.go` `TestLanding/protect-carried` and `TestLanding/protect-target-own`, passed; `TestLanding/protect-run-folder` (an approval counts through two merges of the target, whatever verdict files the run folder holds), passed
- landing, changed paths from Git whatever the repository's config: `landing_test.go` `TestLanding/protect-submodule-hidden` (a gitlink under `.claude/` with `diff.ignoreSubmodules=all` set from the worktree), passed; `TestLanding/protect-submodule-folder` (a gitlink at `.codex`, the protected folder's own path), passed
- landing "A head made at fix": `landing_test.go` `TestLanding/protect-fix`, passed
- landing "An approval before the implementer's new head": `landing_test.go` `TestLanding/protect-resolved`, passed
- landing "A supervised flow": `landing_test.go` `TestLanding/protect-supervised`, passed
- landing, decided again before the merge from the target read afresh: `landing_test.go` `TestLanding/protect-read-during-checks` (the target protects the candidate's paths while its checks are read; no merge request and no merge-in), passed; `TestLanding/protect-read-again` (the target moves after the merge request, and the refused merge is held on retry), passed

## Reconciliation

- flows: applied to grove/specs/flows/spec.md
- landing: applied to grove/specs/landing/spec.md
