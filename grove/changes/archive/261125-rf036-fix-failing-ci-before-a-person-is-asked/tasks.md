## T-01 — publish waits for checks and sends a failure to the implementer once

Result: in a build flow whose `publish` names a step for `failed`, a
published head's checks are waited for before a person is asked, a failure
goes to that step while it can still be asked, and a spent one is said on the
pull request, as the publication change spec's "Checks are waited for before
a person is asked" and "A spent fix is said on the pull request" say.

Files:
- `publication.go`
- `landing.go`
- `flow.go`
- `flows/prompts/failed.md`
- `publication_test.go`
- `landing_test.go`
- `flow_test.go`
- `watch_test.go`
- `testdata/bin/gh`

Interfaces: the build flow's `publish` has the outcome `failed`, which a flow
file may name; landing's reading of checks is shared by `publish` and `land`,
with `tend/review` left out of what `publish` reads; `prompts/failed.md`
fits a failure found at either step.

Constraints: a flow whose `publish` gives nothing for `failed`, the built-in
flows among them, reads no check at `publish` and behaves as before; `land`
behaves as the landing spec says.

Acceptance:
- Each scenario of the publication change spec passes as a test driving the
  built binary.
- The watch-and-status change spec's "Waiting for checks counts" passes as a
  test.
- A build flow file naming `failed` on `publish` is valid, and `tend flow`
  prints it.
- The landing spec's scenarios still pass.

## T-02 — tend's own flow and its documentation

Result: tend's own flow file sends `publish`'s `failed` to `fix`, and the
README and guide say what `failed` on `publish` does, how it shares the one
`fix` with `land`, and why an autonomous flow does not need it.

Files:
- `tend/flow.yaml`
- `README.md`
- `guide/flows-reference.md`
- `guide/concepts/publication.md`
- `guide/concepts/landing.md`

Acceptance:
- `tend flow` on this repository prints `publish` with `failed: fix`, and
  the file is valid.
- The README's and guide's supervised flow example names `failed: fix` on
  `publish`, and their list of `publish`'s outcomes includes `failed`.
- The README says an autonomous flow may name it to keep the first CI run
  out of the landing lock, and need not, since `land` waits for checks.
