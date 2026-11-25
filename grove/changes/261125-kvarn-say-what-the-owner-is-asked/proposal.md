# Say what the owner is asked to do in each notification

## Problem

When tend waits for the owner, its notification does not say what is
wanted, so the owner cannot act from a phone without opening the pull
request, its reviews or a terminal. GitHub issue
[#43](https://github.com/mascah/tend/issues/43) records the first live run:
pull request #42 (change 261125-6dvbs) used its one repair, Codex still
asked for changes, and tend published the pull request as a draft and
notified

> tend 261125-6dvbs: failure
> https://github.com/mascah/tend/pull/42 at 4df1bd8 is ready for you

which does not say that anything is wrong, what the owner is asked to do,
or what each choice leads to.

What the code does at f3f9f87:

- `attempt.go` ends a flow with a notification titled `tend <item>:
  <conclusion>` and the body `<pr> at <head> is ready for you`, or titled
  `tend <item>: landed`.
- `main.go` `stop()` notifies `tend <item> needs you` with the stop's
  message. The status file's `next:` line ("a person deals with that, then
  runs tend again") is not in it.
- `intake.go` sends no notification when it opens a proposal pull request,
  although merging the proposal is the owner's step.
- Nothing records that a wait was already notified, so a run of tend that
  reaches the same wait again, such as `tend watch` running again an
  attempt whose tend died, notifies again.

## Intended outcome

Each notification that waits for the owner names, in its title, the pull
request or request it is about and says it waits for them. Its body says
what the owner is asked to do and, where tend reads the answer from GitHub,
how to give it there: approve, request changes, close, or merge. The link to
the pull request or request is always in it. It is short enough to read on
a phone's lock screen, and each new wait is notified once, not on each pass
of `tend watch`.

The cases the request names:

- a published head that passed review in a flow whose `publish` acts on an
  approval: approve to land, or request changes;
- a head that failed review with its one repair spent: Codex's findings
  stand after one repair, request changes to send them to the implementer,
  or close;
- a head held at `land`, such as for a protected path: what held it, and
  approve this head to land it;
- a proposal pull request: merge to approve, or revise on its branch, with
  how many open questions are consequential.

## Scope

### In scope

- The title and body of the notifications an attempt sends when its flow
  ends and when it stops for a person (`attempt.go`, `main.go`), under
  Herdr and under rex.
- A notification when intake opens a proposal pull request.
- Recording, in the run folder, the wait last notified, so the same wait is
  not notified twice.
- A length bound for the title and the body, keeping the link whole.

### Out of scope

- How notifications reach a phone: tend calls `herdr notification show` or
  rex's equivalent, and delivery is Herdr's or rex's.
- A standing agent to answer follow-up questions (step 5 of
  `docs/work/261121-autonomous-direction/steps.md`).
- A dashboard of everything waiting (issue #24).
- What a stop itself decides, and whether `tend watch` acts after the owner
  answers one; this change only words what the notification says.
- The status file's `next:` line.

## Decisions

None was made beyond the request in issue #43, which says: "Each
notification says, in its first line, which pull request or request it is
about and that it waits for the owner. Its body says what the owner is
asked to do, and how to do it on GitHub where that is how tend reads it",
"The link to the pull request or request is always there", "Notifications
stay short enough to read on a phone lock screen", and "One notification
for each new wait, not one each watch pass".

## Open questions

Q1 to Q3 are consequential and should be answered before merging, and Q9
must be answered before review; Q4 to Q8 can wait for the implementer, who builds the recommended answer and records
it for the reviewer.

1. **Land with the protected-path hold first?** The hold at `land` that the
   request names is change 261125-6dvbs (approved, its pull request #42
   open), which does not exist at f3f9f87. Recommended: yes, add
   261125-6dvbs to `depends_on`, because the hold's scenario in the
   terminals change spec cannot be built or checked without it; the other
   cases do not need it. If not, T-01 builds every case but the hold, and
   the hold's wording goes into 261125-6dvbs. Decides T-01. The same
   applies, less strongly, to change 261125-rf036 (fix failing CI before a
   person is asked), which adds a published head whose checks failed after
   its fix was spent: recommended to word that case here only if it lands
   first, and otherwise leave it to whichever lands second.
2. **Title wording.** Recommended `tend: <ref> waits for you`, where `<ref>`
   is the pull request as `<owner>/<repo>#<n>`, else the request it came
   from in the same form or as its Jira key, else the cycle or change ID;
   and `tend: <ref> landed` for a landing, which waits for no one. The
   repository is in it because one `tend watch` can serve several. Decides
   T-01 and T-02.
3. **How tend counts consequential open questions.** The drafter's
   `summary` marks them in free text today (`flows/prompts/follow.md`).
   Recommended: the follow prompt asks for each consequential question to
   begin with one fixed marker that tend counts, and a summary with none
   marked gives a notification that says "see its open questions" with no
   count, because a count tend guesses at would mislead. Decides T-02.
4. **How a wait is known to be already notified.** Recommended: the run
   folder's `notified` holds the title and body last sent, and a
   notification identical to it is not sent again; a new head or a new
   reason changes the text and so is a new wait. Simple and needs no list
   of wait kinds. Decides T-01 and T-02.
5. **The length bound.** Recommended: a title of at most 60 characters and
   a body of at most 240, the link on the body's last line and kept whole,
   the text before it cut with an ellipsis where it is longer. Decides T-01.
6. **A stop that is not one of the named cases** (an agent that settled
   without a result, a flow that loops, a refused push). Recommended: the
   body is the stop's own reason, shortened to fit, then the link to the
   pull request if there is one, else to the request, adding nothing about
   running tend again, since the reason already says what to do and the
   request notes the person often need not. Decides T-01.
7. **A request no proposal could be drafted from.** Intake comments on the
   request and takes its label off; the request does not name this case.
   Recommended: no notification in this change, because it is the request's
   author who acts on it. Decides nothing unless answered otherwise, when it
   becomes part of T-02.
8. **Routine choices inside the approach**: the helper's name and the file
   that holds it, test names and placement, how the Herdr and rex stand-ins
   in `testdata/bin` record what they were asked to show, and updating the
   README and guide lines that describe the notification (`README.md`,
   `guide/run.md`) where this change makes them false. Recommended: the
   implementer chooses and records each.
9. **Checking the rex scenario** (raised by checking the change; must be
   answered before review). Under rex tend reports its state by writing to
   `/dev/tty` (`rex.go`), which the acceptance tests, driving the built
   binary, cannot read, so T-01's acceptance for the terminals scenario "A
   stop under rex" cannot be checked as written. Recommended: tend writes
   that report to a file named by an environment variable when it is set,
   used only by the tests, so the scenario is checked through the binary
   like the others; the alternative is to check it with a unit test of the
   message tend builds and drop the stand-in from T-01's acceptance for it.
   Decides T-01.

## Related documents

- Requirements: [terminals](specs/terminals/spec.md),
  [intake](specs/intake/spec.md)
- Tasks: [tasks.md](tasks.md)
