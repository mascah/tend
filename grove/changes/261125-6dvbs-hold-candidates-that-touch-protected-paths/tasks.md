## T-01 — Protected paths read from the target's flow file

Result: tend knows the protected paths in effect for a repository, either
the project's `protect` from the target's flow file with `tend/` added, or
tend's default list. It validates them as part of the flow file, and
`tend flow` prints them.

Files:
- `flow.go`
- `flow_load.go`
- `flow_test.go`

Interfaces: the flow file's `protect` field; a function giving the
protected paths in effect at a revision of the target, with where they came
from; and a function saying which of a list of repository paths an entry
list covers.

Constraints: the protected paths are never read from the attempt's branch or
worktree. The default list is the one open question 4 of the proposal
recommends, unless the owner answers otherwise.

Acceptance:
- The flows change spec's "The flow file names the protected paths" holds,
  each of its scenarios shown by a test.
- `tend flow` on a repository with no flow file still prints a file that is
  valid, and attempts behave as before it was added.

## T-02 — Hold at land a head that touches a protected path

Result: at `land`, a head whose changed paths include a protected path lands
only with a person's approval of that head, or of the head it is carried
from. Otherwise tend stops for a person and names the paths.

Files:
- `landing.go`
- `landing_test.go`

Interfaces: none beyond T-01's.

Constraints: the changed paths come from Git against the target's tip, and
the approval comes from GitHub's reviews, by the same rule as `publish`'s
`approved`. No verdict, run-folder file or agent's output changes the
outcome. The check runs when `land` begins, before any merge commit is
pushed, and again before GitHub is asked to merge. Both times it uses the
protected paths of the target as then fetched.

Acceptance:
- The landing change spec's "A head that touches a protected path waits for
  a person" holds, each of its scenarios shown by a test.
- The existing landing tests pass unchanged.
