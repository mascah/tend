# Requested changes count only from people with write access

## Observed

tend acts on a change request whoever wrote it.

- `requested` in `main.go` at 40ca101 takes the newest review on the pull
  request's head whose state is `CHANGES_REQUESTED`. It reads nothing about
  the review's author.
- `publishStep` in the same file then fills the review's body and its line
  comments into the request prompt (`flows/prompts/request.md` at 40ca101,
  "Their review: {review}{comments} Make the changes they ask for"), and the
  flow sends it to the implementer.
- The implementer is Claude Code started with `--permission-mode auto`
  (`startAgent` in `flow.go` at 40ca101), so it acts on that text with no one
  asked.
- `watch` in `main.go` at 40ca101 runs tend again on any attempt whose pull
  request has a `CHANGES_REQUESTED` review on its head, again whoever wrote
  it, so with `tend watch` running no person starts any of this.

`mascah/tend` is a public repository. GitHub lets any signed-in account
submit a review on a public repository's pull request, including one that
requests changes. That was not tried against GitHub for this proposal; the
code above was read.

## Expected

Only a person with write access to the repository can have tend's
implementer act on their review.

The owner decided this on 2026-11-21. Asked how to fix that "tend acts today
on a change request from any GitHub account, so on a public repository a
stranger's review text reaches the implementer as a prompt", the owner chose
"A fifth change, a bugfix": "requested changes count only from accounts
GitHub reports as owner, member or collaborator", drafted with the changes
for running several attempts at once and built first. The owner then
answered "Yes" to a round of decisions that included these:

1. Write access is told from the review's author association: owner, member
   or collaborator. Reason: GitHub gives it on every review, so it costs no
   further call.
2. tend and `tend watch` ignore anyone else's change request, and tend says
   whose request it left.
3. One limit is accepted: in a repository an organization owns, a member of
   the organization with read access only still counts.
4. Test layout, private helpers and the wording of the sentence tend says
   are the implementer's to choose.

## Reproduction

From a clean checkout at 40ca101, with Go and Grove installed:

1. Run `go test -run 'TestTend/^request$'`. It passes. The scenario's change
   request (`request` in `tend_test.go`) is written by the login `teammate`
   and names no author association, and the stand-in `testdata/bin/gh`
   returns none. tend prompts the implementer with it all the same.
2. Search `main.go`, `flow.go`, `testdata/bin/gh` and `tend_test.go` for
   `author_association` or `authorAssociation`. Nothing is found: tend asks
   GitHub for no fact about a review's author but the login it quotes in the
   prompt, so a review by an account with no access is acted on as the
   teammate's is.

## The requirement it breaks

None. The code does what the maintained specs say, and the specs are what is
wrong:

- publication, "Requested changes are acted on once" (40ca101), acts "When
  the open pull request's newest review on its head requests changes", with
  no condition on who wrote the review.
- watch-and-status, "Watch acts on a person's behalf" (40ca101), runs tend
  again on "a change request on its head that tend has not answered", again
  with no condition.

The owner agreed to change both. The change specs below give each
requirement in full.

## Open questions

None.

## Related documents

- Tasks: [tasks.md](tasks.md)
- Requirements: [specs/publication/spec.md](specs/publication/spec.md),
  [specs/watch-and-status/spec.md](specs/watch-and-status/spec.md)
