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

What the code does at 7e339d9:

- `attempt.go` ends a flow with a notification titled `tend <item>:
  <conclusion>` and the body `<pr> at <head> is ready for you`, or titled
  `tend <item>: landed`.
- `main.go` `stop()` notifies `tend <item> needs you` with the stop's
  message. The status file's `next:` line ("a person deals with that, then
  runs tend again") is not in it.
- `intake.go` sends no notification when it opens a proposal pull request,
  although merging the proposal is the owner's step.
- A head held at `land` for a protected path stops for a person
  (`landing.go`), and so its notification is the generic one above.
- When a check still fails on a published head after the one fix is spent,
  `publication.go` comments on the pull request and `publish` then gives the
  outcome it would without the wait, so the flow can end with the
  notification of a head that passed review, which does not say that a
  check failed.
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

The cases, all but the fourth named by the request, the fourth added by
the owner (decision 4):

- a published head that passed review in a flow whose `publish` acts on an
  approval: approve to land, or request changes;
- a head that failed review with its one repair spent: Codex's findings
  stand after one repair, request changes to send them to the implementer,
  or close;
- a head held at `land`, such as for a protected path: what held it, and
  approve this head to land it;
- a published head whose checks still fail after its one fix: which checks
  failed, and to fix the branch, run the checks again or close;
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

The request, issue #43, says: "Each notification says, in its first line,
which pull request or request it is about and that it waits for the owner.
Its body says what the owner is asked to do, and how to do it on GitHub
where that is how tend reads it", "The link to the pull request or request
is always there", "Notifications stay short enough to read on a phone lock
screen", and "One notification for each new wait, not one each watch pass".

On 2026-11-26 the owner answered "Yes" to each of the following, put to them
as recommendations in the session that revised this proposal:

1. **The title** is `tend: <ref> waits for you`, where `<ref>` is the pull
   request as `<owner>/<repo>#<n>`, else the request it came from in the
   same form or as its Jira key, else the cycle or change ID; and
   `tend: <ref> landed` for a landing, which waits for no one. Reason: one
   `tend watch` can serve several repositories, so the repository is named.
   Decides T-01 and T-02.
2. **Consequential open questions are counted by a fixed marker.** The
   drafter's prompt (`flows/prompts/follow.md`) asks for each consequential
   question to begin with one fixed marker, and tend counts the markers; a
   summary with none marked gives no count and says to see the open
   questions. Reason: the drafter marks them in free text today, and the
   drafted reviews on pull requests #44 and #53 already mark them in two
   different ways, so a count tend read from free text would mislead.
   Decides T-02.
3. **The rex case is checked by an in-package unit test of the message**
   tend reports under rex, not by a test-only environment variable in
   tend's code. Reason: under rex tend writes to `/dev/tty` (`rex.go`),
   which tests driving the built binary cannot read, and the package
   already has in-package tests such as `TestProtectedPaths` in
   `flow_test.go`. Decides T-01.
4. **A published head whose checks still fail after its one fix is a named
   case**: the notification says which checks failed after the one fix, and
   to fix the branch, run the checks again or close the pull request, as
   the comment tend posts on the pull request says. Reason: without it the
   flow ends with the notification of a head that passed review. Decides
   T-01.
5. **The smaller choices**, each as the draft recommended:
   - A notification identical to the last one sent for the same attempt or
     request, as recorded in its run folder's `notified`, is not sent
     again; a new head or a new reason changes the text and so is a new
     wait. Decides T-01 and T-02.
   - A title is at most 60 characters and a body at most 240, the link on
     the body's last line and whole, the text before it cut with an
     ellipsis. Decides T-01.
   - Any other stop gives its own reason, shortened to fit, then the link
     to the pull request, else to the request, adding nothing about running
     tend again. Decides T-01.
   - A request no proposal could be drafted from gets no notification,
     since its author acts on the comment intake leaves. Decides T-02.

The draft asked whether the protected-path hold (change 261125-6dvbs) had to
land first. It was merged into main in pull request #42, before this
change's base, so the hold at `land` is built and checked here with no
dependency.

## Open questions

None must be answered before review.

1. **Routine choices inside the approach**: the helper's name and the file
   that holds it, test names and placement, how the Herdr stand-in
   in `testdata/bin` records what it was asked to show, the fixed
   marker's text, and updating the README and guide lines that describe the
   notification (`README.md`, `guide/run.md`) where this change makes them
   false. Recommended: the implementer chooses and records each. Decides
   T-01 and T-02; can wait for implementation.

## Related documents

- Requirements: [terminals](specs/terminals/spec.md),
  [intake](specs/intake/spec.md)
- Tasks: [tasks.md](tasks.md)
