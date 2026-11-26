# Implementation: Give proposal pull requests a tend/review status

<!-- Progress on change 261126-c16zs, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — Set tend/review on proposal heads

Status: done

Evidence: `intake.go` gains `vouchOpen` and `vouch`. `vouchOpen` runs on each
pass after unfinished handoffs and before new requests, whatever the cap. It
covers every open pull request the account opened from a `tend/issue-` or
`tend/jira-` branch, read from `repos/{owner}/{repo}/pulls` with
`--paginate` so no page limit hides one, taking each head's `head.sha` and
fetching it from origin when it is missing. The cap's count of open
proposals reads the same list, through `openProposals`. `vouch` reads the
commit's statuses and skips a head the account already set `tend/review` on.
Otherwise it sets `success`, "Grove documents only: merging approves this
change", or `failure` naming why, linking the pull request. The publish step
calls `vouch` on the head of a pull request it just opened. One it found
instead, as when a stopped pass resumes, may have moved on, so its real head
is left to `vouchOpen`. The decision, `notDocs`,
uses landing's `changedPaths` (Git alone, against the target as fetched):
every path must lie under one `<root>/changes/<id>-<slug>/` folder. The gh
stand-in now reports `headRefOid` in `pr list` and serves
`commits/<sha>/statuses` with each status's creator. New tests
`propose-vouch`, `propose-vouch-open` and `propose-vouch-resumed` are in
`intake_test.go`; the last covers a resumed pass whose branch a person moved
on, with the stand-in's new `STATUS_FAIL`.
`go test ./...` passed (119s), including every intake, publication and
landing scenario.

Check: the proposal's first open question was marked "must be answered before
review": with this change the publishing account can merge a docs-only
proposal head, because the rule no longer stops it. The change does not answer
it. I built its recommendation (accept it), since the change is approved as
written and is easy to revert. The owner accepted it in review of #59.
Whether to add this to the known ceilings in `guide/built-vs-planned.md` is
left to them.
Check: a head that touches anything else gets `failure`, as the second open
question recommends. Its description names the outside path, or the two
change folders, or says "Changes nothing", cut to GitHub's 140 characters.
Check: "one change's folder" means a folder directly under `<root>/changes/`
named `<digits>-<id>-<slug>`, as the third open question recommends. So
`changes/archive/` and loose files under `changes/` fail; a test covers an
archived change.
Check: a head that already has a `tend/review` from the account, success or
failure, is never checked again. A failure to read or set the status stops
for a person (exit 3), as other status writes do. A head that `gh` reports
but tend cannot fetch is skipped with a message and tried on the next pass.
Check: no real GitHub call was made. The status endpoints are used as GitHub
documents them (`commits/{ref}/statuses` with `creator.login`), but were seen
only through the stand-in.
Also changed: `guide/glossary.md` — `tend/review` was described as set only on reviewed commits
Also changed: `guide/setup.md` — section 6 implied only reviewed heads pass the rule
Also changed: `guide/concepts/intake.md` — intake now sets a status on each proposal head

## Scenarios

- intake "A proposal just opened": intake_test.go propose-vouch ("a proposal just opened has tend/review success…", "described as Grove documents only"), passed
- intake "A proposal opened before this change": intake_test.go propose-vouch-open ("but a proposal opened before has tend/review success…", with CAP=3 reached), passed
- intake "A person revises the proposal": intake_test.go propose-vouch ("a person's revision … gets its own success", "and the earlier head's is not set again"), passed
- intake "A head that touches code": intake_test.go propose-vouch ("a head that also touches code gets no success"), passed
- intake "Two changes in one head": intake_test.go propose-vouch-open ("two changes in one head get no success"), passed
- intake "Pull requests that are not proposals": intake_test.go propose-vouch-open ("a merged proposal gets no status", "nor one a person opened from a tend/issue- branch"), passed

## Reconciliation

- intake: applied to grove/specs/intake/spec.md
