## T-01 — An attempt's notifications say what the owner is asked to do

Result: when an attempt's flow ends or tend stops for a person, the
notification names the pull request, request or work item, says what the
owner is asked to do and how, ends with the link, fits a lock screen, and is
sent once for each wait.

Files:
- `main.go`
- `attempt.go`
- `terminal.go`
- `rex.go`
- `testdata/bin/herdr`
- `testdata/bin/rex`
- `terminal_test.go`

Interfaces: one function through which every notification of an attempt or
request is sent, given the title and body, that applies the length bound
and skips a notification identical to the last one recorded in the run
folder (open questions 4 and 5).

Constraints: the hold at `land` is worded only once the protected-path hold
exists (open question 1); a stop still exits 3 and still writes `needs-you`
as before.

Acceptance:
- Each scenario of the terminals change spec's requirements "Telling a
  person", "An ended attempt says what the owner is asked to do", "Another
  stop gives its reason and the link", "One notification for each wait" and
  "Short enough for a lock screen" passes against the built binary, with
  the Herdr and rex stand-ins recording what they were asked to show.
- The scenario of "A hold at land says what held it" passes once change
  261125-6dvbs is in the target; until then the implementer records it as
  not built.

## T-02 — Intake tells the owner of a proposal

Result: when intake opens a proposal pull request, the owner is notified
once with the request it came from, how many open questions are
consequential, and to merge to approve or revise on its branch.

Files:
- `intake.go`
- `flows/prompts/follow.md`
- `intake_test.go`

Interfaces: uses T-01's notification function, with the request's run
folder as its record.

Acceptance:
- Each scenario of the intake change spec's requirement "The owner is told
  of a proposal" passes against the built binary.
- A request no proposal could be drafted from is still answered on the
  request as before, with no notification (open question 7).
