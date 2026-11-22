## T-01 — tend builds, reports and tidies a change in no cycle

Result: `tend <repo> <change-id>`, `tend status <repo> <change-id>`, the
overview and the tidying of finished worktrees treat an approved change that
no cycle lists as they treat a cycle, as the attempts requirement "A change
in no cycle is a work item" and the watch-and-status requirement "What waits
for whom" in this change say.

Files:
- `main.go`
- `tend_test.go`

Delegated: test layout and private helpers; the wording of one-sentence
messages.

Constraints: a cycle is found, claimed and built as before; tend checks
nothing about Grove's version, and a claim Grove refuses stops with Grove's
sentence.

Acceptance:
- Each scenario of "A change in no cycle is a work item" is a test that
  passes, run against Grove 0.14.0.
- The "What waits for whom" scenarios "A repository with work at every
  stage", "A change that waits" and "An attempt at a change" are tests that
  pass, and `overview` passes with the new heading.
- The review tend posts on a proposal pull request no longer says the merged
  change waits for a cycle. It says the change can be built by its ID, or
  planned into a cycle with others.
- `go test ./...` passes.

## T-02 — Grove 0.14.0 is the Grove tend needs

Result: CI runs the tests against Grove 0.14.0, and the README says tend
needs it and how a change in no cycle is built.

Files:
- `.github/workflows/ci.yml`
- `README.md`

Acceptance:
- CI installs Grove `v0.14.0`.
- The README's "What it asks of Grove" names Grove 0.14.0 or later and
  `grove claim <change-id>`.
- The README's "Run it" and "How work reaches tend" say an approved change
  is built with `tend <repo> <change-id>`, with a cycle needed only to build
  several changes together.
- `tend watch` is still described as starting only cycles marked
  `unattended`.
