## T-01 — A change request counts only from a person with write access

Result: tend and `tend watch` act on a change request only when GitHub gives
the review's author as owner, member or collaborator, as the publication
requirement "Requested changes are acted on once" and the watch-and-status
requirement "Watch acts on a person's behalf" in this change say.

Files:
- `main.go`
- `testdata/bin/gh`
- `tend_test.go`
- `README.md`

Delegated: test layout and private helpers; the wording of the sentence tend
says when it leaves a request.

Constraints: the same rule decides in `tend <repo> <item>` and in `tend
watch`, so watch never runs tend for a request tend would then leave.

Acceptance:
- The publication scenarios "A stranger requests changes" and "A stranger's
  review after a teammate's", and the watch-and-status scenario "A stranger's
  request", are tests that pass.
- `request`, `request-edited`, `request-declined`, `request-uncertain`,
  `request-merged` and `watch` still pass, their reviews written by a
  collaborator.
- The README's account of requested changes says whose requests tend acts
  on, and that it leaves anyone else's.
