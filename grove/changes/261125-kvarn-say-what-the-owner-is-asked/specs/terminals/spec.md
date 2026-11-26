# Changes to terminals

The changes that change 261125-kvarn proposes to the requirements of
[terminals](../../../../specs/terminals/spec.md), as that spec stands at
7e339d9. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled.

The examples use the title wording, length bound and once-per-wait record
that the proposal's decisions 1 and 5 fix.

## MODIFIED Requirements

### Requirement: Telling a person

When tend stops for a person or finishes an attempt, the system shall tell
them through the terminal: under Herdr a notification whose title names the
pull request, else the request, else the work item, and says it waits for
them (`tend: <ref> waits for you`), or that it landed (`tend: <ref>
landed`), with a body saying what the person is asked to do and ending with
the link to that pull request or request where there is one; under rex by reporting tend's own state, `blocked` for a stop or
`done` for an end, with that title and body as its message, to the terminal
tend runs in.

#### Scenario: A stop under Herdr

- GIVEN tend under Herdr stops because issue #7 of `o/r` lacks `tend-build`
- WHEN it exits 3
- THEN Herdr shows a notification `tend: o/r#7 waits for you` whose body
  says the issue is not labelled tend-build and ends with the issue's URL

#### Scenario: A stop under rex

- GIVEN the same stop under rex
- WHEN it exits 3
- THEN tend reports the state `blocked` to its terminal with a message that
  begins `tend: o/r#7 waits for you` and holds the issue's URL

#### Scenario: A landing

- GIVEN a flow that lands pull request #42 of `o/r`
- WHEN tend lands it
- THEN the notification is titled `tend: o/r#42 landed` and does not say it
  waits for anyone

## ADDED Requirements

### Requirement: An ended attempt says what the owner is asked to do

When an attempt's flow ends with a published head waiting for a person,
the system shall say in the notification's body what that head's review
concluded and the choices tend acts on: for a head that passed review,
approve to land or request changes where the flow's `publish` acts on an
approval, and merge or request changes where it does not; for a head that
failed review, that Codex's findings stand, after one repair where the
repair was spent, and to request changes to send them to the implementer
or close the pull request.

#### Scenario: A supervised head that passed review

- GIVEN a flow whose `publish` sends `approved` to `land`, and head 4df1bd8
  of pull request #42 of `o/r` that passed review
- WHEN the flow ends
- THEN the notification is titled `tend: o/r#42 waits for you` and its body
  says review passed at 4df1bd8, to approve to land or request changes, and
  ends with the pull request's URL

#### Scenario: A head that passed review where approval lands nothing

- GIVEN the built-in build flow and a published head that passed review
- WHEN the flow ends
- THEN the body says to merge it or request changes, and says nothing of
  approving

#### Scenario: Findings that stand after the repair

- GIVEN a published head whose review failed after the attempt's one repair
  was spent
- WHEN the flow ends
- THEN the body says Codex's findings stand after one repair, to request
  changes to send them to the implementer or close, and ends with the pull
  request's URL

#### Scenario: A failure the reviewer did not judge repairable

- GIVEN a published head whose review failed and was not marked actionable,
  so no repair was asked
- WHEN the flow ends
- THEN the body says Codex's findings stand, to request changes to send
  them to the implementer or close, and does not mention a repair

#### Scenario: No pull request

- GIVEN a repository whose origin is not listed in `TEND_STATE/publish`,
  and a head of branch `tend/issue-7` that passed review
- WHEN the flow ends
- THEN the body names the branch and its head, and ends with the URL of
  issue #7, the request the work came from

### Requirement: A hold at land says what held it

When tend stops at `land` because the head may land only with a person's
approval, the system shall say in the notification's body what held it and
to approve this head to land it, ending with the pull request's URL.

#### Scenario: A protected path

- GIVEN a head of pull request #42 of `o/r` that passed review and changes
  `tend/flow.yaml`, a protected path, with no approval
- WHEN tend reaches `land` and stops
- THEN the notification is titled `tend: o/r#42 waits for you`, and its
  body names `tend/flow.yaml`, says to approve this head to land it, and
  ends with the pull request's URL

### Requirement: Failed checks after the one fix are said

When an attempt's flow ends at a published head on which a check still
failed after the step `publish` gives for `failed` was spent, as the
publication spec's "A spent fix is said on the pull request" says, the
system shall say in the notification's body which checks failed after the
one fix, and to fix the branch, run the checks again or close the pull
request, ending with the pull request's URL, and shall not say the head is
ready to approve or merge.

#### Scenario: Still red after the fix

- GIVEN a flow whose `publish` sends `failed` to a `fix` step with limit 1,
  and pull request #42 of `o/r` whose check `test` fails on its head and
  again on the head `fix` makes
- WHEN the flow ends
- THEN the notification is titled `tend: o/r#42 waits for you`, and its
  body names `test`, says it failed after the one fix, says to fix the
  branch, run the checks again or close, ends with the pull request's URL,
  and says nothing of approving

### Requirement: Another stop gives its reason and the link

When tend stops for a person for any other reason, the system shall give
the stop's reason as the notification's body, followed by the link to the
attempt's pull request where it has one and otherwise to its request where
it has one.

#### Scenario: A stop after publishing

- GIVEN an attempt that published pull request #42 of `o/r`, whose
  implementer then committed nothing for a requested change and is not
  working on it
- WHEN tend stops
- THEN the notification is titled `tend: o/r#42 waits for you`, and its
  body gives that reason and ends with the pull request's URL

### Requirement: One notification for each wait

If a notification is the same as the last one sent for the same attempt or
request, then the system shall not send it again.

#### Scenario: A died tend run again by watch

- GIVEN an attempt whose flow ended at a published head that passed review,
  notified once, and whose status was then left `working` by a tend that
  died before writing `done`
- WHEN `tend watch` runs tend on it again and the flow ends at the same head
- THEN no second notification is shown

#### Scenario: A new head

- GIVEN that attempt notified for head 4df1bd8, and a change a person
  requested that the implementer commits as a new head that passes review
- WHEN the flow ends again
- THEN a notification for the new head is shown

### Requirement: Short enough for a lock screen

The system shall keep a notification's title to at most 60 characters and
its body to at most 240, with the link on the body's last line and whole,
shortening the text before the link and marking the cut with an ellipsis.

#### Scenario: A long stop reason

- GIVEN a stop whose reason is 600 characters long, on an attempt with a
  pull request
- WHEN tend notifies
- THEN the body is at most 240 characters, ends with the pull request's
  whole URL, and the reason is cut with an ellipsis before it

## Observed current behavior

- `attempt.go` at 7e339d9 notifies `tend <item>: <conclusion>` with `<pr or
  branch> at <head> is ready for you` when a flow ends, or `tend <item>:
  landed` with `<pr> is merged`.
- `main.go` `stop()` at 7e339d9 notifies `tend <item> needs you` with the
  stop's message, which is sometimes several hundred characters long (for
  example the stops in `flow_run.go` that tell a person to look at a pane).
- `rex.go` at 7e339d9 joins title and body as `<title>: <body>`, replaces
  control characters with spaces and cuts the message at 2000 bytes.
- Nothing at 7e339d9 records a notification already sent.
