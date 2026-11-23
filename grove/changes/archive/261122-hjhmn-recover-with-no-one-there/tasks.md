## T-01 — A step whose agent is gone is asked once more

Result: a step recorded as asked, with no result, whose agent has no pane is
asked a second time in a new session and no more often, while an agent that
is there and not working still stops for a person, as the flows requirements
"Any step recovers without being prompted twice" and "A step is asked no
more often than its limit", the review requirement "An
uncertain repair goes to a person", the publication requirement "A request
answered with no change" and the intake requirement "Drafting prompts are
sent once" in this change say.

Files:
- `flow.go`
- `main.go`
- `flow_test.go`
- `tend_test.go`

Delegated: test layout and private helpers; how the second ask is recorded
in the run or intake folder; the wording of one-sentence messages.

Constraints: an agent still working is waited for and never prompted again;
blocked and unknown agents still stop for a person; deleting a step's record
still has it asked once more.

Acceptance:
- The flows scenarios "The agent is gone", "Gone a second time" and "A
  repair asked once more", the
  review scenario "The implementer is gone", the publication scenario "The
  implementer is gone" and the intake scenario "The drafter is gone" are
  tests that pass.
- `request-uncertain`, `repair-blocked` and the flows test for "The outcome
  of a step is unknown" still pass, their agents there and idle.

## T-02 — Watch runs again an attempt whose tend died

Result: `tend watch` runs tend again on an attempt left `working` with no
tend on it, as the watch-and-status requirement "Watch runs again an attempt
whose tend died" says.

Files:
- `main.go`
- `tend_test.go`

Delegated: test layout and private helpers; the wording of one-sentence
messages.

Constraints: an attempt that needs a person is left alone; in a repository
whose flow file sets the count, the run is within it.

Acceptance:
- Each scenario of that requirement is a test that passes.
- `watch`, `watch-unattended` and `status-interrupted` still pass.

## T-03 — The README says what recovers by itself

Result: the README says what `tend watch` and a rerun recover with no one
there, and what still waits for a person.

Files:
- `README.md`

Acceptance:
- It says watch runs again an attempt whose tend died, that a step whose
  agent is gone is asked once more in a new session, and that an agent that
  is idle with no result, blocked or unknown still needs a person.
- It says that on the first pass after upgrading, an attempt left `working`
  is run again.
