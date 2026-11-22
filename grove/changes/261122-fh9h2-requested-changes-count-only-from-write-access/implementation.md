# Implementation: Requested changes count only from people with write access

<!-- Progress on change 261122-fh9h2, written by whoever implements it. It is not
part of what was approved and does not amend it. If the owner drops the change,
write "Status: dropped" here with their reason. -->

## Tasks

### T-01 — A change request counts only from a person with write access

Status: done

- `main.go`: `writer(field)` is the one rule, a jq test that the review's
  author association is `OWNER`, `MEMBER` or `COLLABORATOR`. `requested`
  (tend) applies it to the REST field `author_association` when picking the
  newest change request on the head, and says on standard error "left the
  change request of <logins> on <pr>: GitHub does not give them write access"
  for each other account's change request on that head. `watch` applies the
  same rule to `authorAssociation` from `gh pr view --json reviews`, so watch
  never runs tend for a request tend would leave.
- `testdata/bin/gh`: `pr view` gives each review's `authorAssociation` from
  its `author_association`.
- `tend_test.go`: `request()` now writes the teammate as a `COLLABORATOR`
  through `requestAs(login, association, state)`; new scenarios
  `request-stranger` and `request-stranger-after`.
- `README.md`: the requested-changes paragraph, the `tend watch` paragraph
  and step 3 of the loop say only owners, members and collaborators are acted
  on, that anyone else's request is left and tend says whose, and the
  organization-member limit.
- Checks: `go test -run 'TestTend/^(request.*|watch)$'` passes. With the
  association test replaced by `true` in `main.go`, `request-stranger` and
  `request-stranger-after` fail; with the filter dropped from `watch` only,
  `request-stranger` fails on its watch check. The full `go test ./...`
  passes except `overview`, which fails the same way on main at 4cbdc71
  ("a cycle not done, with Grove's verdict"); it depends on the installed
  Grove's output, not on this change.

Check: tend says whose request it left on every run while that request is the
head's, not once; the wording is delegated and repeating it costs nothing.
Check: the real `gh pr view --json reviews` field `authorAssociation` and the
REST field `author_association` were not tried against GitHub; the stand-in
gives both names as GitHub documents them.

## Scenarios

- publication "A teammate requests changes": tend_test.go `TestTend/request`, passed
- publication "The pull request is merged": tend_test.go `TestTend/request-merged`, passed
- publication "A stranger requests changes": tend_test.go `TestTend/request-stranger`, passed
- publication "A stranger's review after a teammate's": tend_test.go `TestTend/request-stranger-after`, passed
- watch-and-status "Changes requested": tend_test.go `TestTend/watch`, passed
- watch-and-status "An attempt that needs a person": tend_test.go `TestTend/watch`, passed
- watch-and-status "A request already answered": tend_test.go `TestTend/request-declined`, passed
- watch-and-status "A stranger's request": tend_test.go `TestTend/request-stranger` (its watch check), passed
- watch-and-status "Intake": tend_test.go `TestTend/watch-intake`, passed
- watch-and-status "A cycle marked unattended on GitHub": tend_test.go `TestTend/watch-unattended-merged`, passed
- watch-and-status "A cycle marked unattended": tend_test.go `TestTend/watch-unattended`, passed

## Reconciliation

<!-- One line per change spec, once applied to its maintained spec:

    - record-editing: applied to specs/record-editing/spec.md
-->
