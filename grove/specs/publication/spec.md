# Publication

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

With its origin listed in `TEND_STATE/publish`, an attempt's reviewed head is published by a separate GitHub account: the branch pushed, one pull request titled and described by the implementer's commit message, each verdict a comment review with a `tend/review` status, and changes a person requests on the pull request's head acted on once. The publishing account is the one a line `<origin-url> <account>` of `TEND_STATE/publish` names, such as a bot user signed in to `gh` beside the owner, so the owner can approve its pull requests. People reviewing and merging tend's pull requests rely on this.

## Boundaries

This covers what an attempt writes to GitHub after its head is reviewed: the
push, the pull request and its title and description, the reviews and
statuses, and requested changes with the implementer's reply. Reviewing a
head is [review](../review/spec.md); opening proposal pull requests from
requests is [intake](../intake/spec.md); the closing line and Jira link of an
issue build are [work-items](../work-items/spec.md); `tend watch` running an
attempt again when changes are requested is
[watch-and-status](../watch-and-status/spec.md). tend never merges a pull
request; merging stays with people.

## Requirements

### Requirement: Publish only to listed origins, as their account

When an attempt's head is reviewed, the system shall publish only if a line
of `TEND_STATE/publish` reads `<origin-url> <account>` for the repository's
origin, making every GitHub write with that account's token from `gh auth
token --user <account>`; with an origin that is not listed it shall say so,
call GitHub not at all and keep the verdict local; and if `gh` holds no token
for the account it shall stop for a person.

#### Scenario: An origin not listed

- GIVEN a repository with origin and no `TEND_STATE/publish`
- WHEN tend takes a cycle to a verdict
- THEN it runs no `gh` command, origin has no `grove/` branch, standard error
  says the origin is not listed in `TEND_STATE/publish`, and tend exits 0

#### Scenario: Without origin

- GIVEN a repository with no remote
- WHEN tend takes a cycle to a verdict
- THEN it runs no `gh` command

#### Scenario: A listed origin

- GIVEN `TEND_STATE/publish` lists the origin with account `the-bot`
- WHEN tend takes a cycle to a passing verdict
- THEN the pull request, its review and the head's status are each
  `the-bot`'s

### Requirement: Push, never forced

When publishing, the system shall push the head to origin's branch of the
same name over HTTPS with the account's token, even where origin is an SSH
URL, never forced; and if the push is refused, as when someone else pushed to
the branch, it shall stop for a person and leave origin's branch as it is.

#### Scenario: The reviewed head is pushed

- WHEN tend publishes a cycle's reviewed head
- THEN origin's branch is at that head

#### Scenario: Someone else pushed to the branch

- GIVEN another person pushed a commit to the attempt's branch on origin, and
  the attempt has a commit of its own since
- WHEN the owner runs tend again
- THEN it exits 3, saying it could not push, and origin's branch still ends
  with the other person's commit

### Requirement: One pull request, described by the commit message

When the branch has no pull request on GitHub in any state, the system shall
have the implementer write the squash commit message for the head to
`message-<head>` in the run folder, once per head, ending with the work
item's line (`Grove: <change ids>` for a cycle), and open one pull request
onto the target, titled by the message's subject and described by its body;
it shall look the pull request up on GitHub on every run, never from memory,
and never open a second.

#### Scenario: The pull request opens

- WHEN tend publishes a cycle's reviewed head
- THEN one pull request is opened onto `main` by the account, titled `Add the
  file` and described by `The project needed a file.` and `Grove:
  261015-c0000`, and the implementer was asked for the message once

#### Scenario: GitHub opened it but tend never heard back

- GIVEN `gh pr create` opened the pull request and then failed
- WHEN tend stops and the owner runs it again
- THEN the first run exited 3 saying a rerun looks for one before opening
  another, and the rerun opens no second pull request, posts the head's review
  and asks for no second message

#### Scenario: A repeated run

- GIVEN a published attempt whose description a person has replaced
- WHEN the owner runs tend again with no new commit
- THEN no second pull request is opened, the description stays the person's,
  no second message is asked for and the head gets no second review

### Requirement: The description follows later heads until a person edits it

When the open pull request's head has moved past the head its description was
written for, the system shall have the implementer write the message for the
new head, given the current description to keep what still holds, and set it
as the pull request's title and description; and if a person has edited the
title or description since tend set them, it shall leave them, say so, and
not rewrite them again.

#### Scenario: A new head after a requested change

- GIVEN a published pull request titled `Add the file`, and a new head made
  for a requested change
- WHEN tend publishes the new head
- THEN the title becomes `Add the file, plainer` and the description is the
  new message's body

#### Scenario: A description a person edited

- GIVEN a person replaced the description with their own words, then
  requested changes
- WHEN tend publishes the new head
- THEN the description is still their words, no `gh pr edit` runs, and tend
  says a person has edited it

### Requirement: A review and a status for each reviewed head

For each valid verdict of the attempt, oldest first, whose head is on the
branch and that the account has not reviewed, the system shall set the
`tend/review` commit status on that head, `success` for a passing verdict and
`failure` otherwise, then post a comment review on that commit by the account
giving the conclusion and summary, whether a failure was judged repairable or
needs a person's decision, the implementer's reply to a requested change, and,
on the head of a Grove cycle, what `grove review` says merging approves; and
it shall post no review twice.

#### Scenario: A passing head

- WHEN tend publishes a cycle's head that passed review
- THEN the head has one comment review by the account saying the review
  passed and showing what merging approves, and the status `tend/review`
  `success`

#### Scenario: A repair within one run

- GIVEN a head failed as actionable and its repair passed in the same run
- WHEN tend publishes
- THEN the failed head has its own review, saying tend repairs once, and
  status `failure`, and the repaired head has one review and status `success`

#### Scenario: A later commit that needs a decision

- GIVEN a published attempt, then a new commit whose review fails with
  `actionable` false
- WHEN the owner runs tend again
- THEN the new head gets status `failure` and a review saying a finding needs
  a decision, the earlier head keeps its one review, and there is still one
  pull request

### Requirement: Requested changes are acted on once

When the newest review on the open pull request's head by a person with write
access requests changes, the system shall write `request-<review id>` in the
run folder just before giving the implementer the review's body and line
comments once, then have the new head reviewed and published, the head's
review quoting the implementer's reply; and it shall not act on a request on
an older head, on a merged or closed pull request, on plain comments, or on a
review by anyone else, saying on standard error whose request it left. A
person with write access is an account whose review GitHub gives the author
association owner, member or collaborator.

#### Scenario: A teammate requests changes

- GIVEN a published attempt, and a teammate who is a collaborator requests
  changes on its head saying "Say it plainly." with the line comment
  `a.txt:1` "Plainer."
- WHEN the owner runs tend again
- THEN the implementer is asked once, with the review and `a.txt:1:
  Plainer.`, the new head is reviewed, its review quotes the implementer's
  reply, there is still one pull request, and a further run asks nothing again

#### Scenario: The pull request is merged

- GIVEN changes requested on the head of a pull request that is then merged
- WHEN the owner runs tend again
- THEN the implementer is not asked to make them

#### Scenario: A stranger requests changes

- GIVEN a published attempt, and an account with no access to the repository
  requests changes on its head
- WHEN the owner runs tend again
- THEN the implementer is not asked, no `request-<review id>` is written,
  standard error names the account whose request tend left, and tend exits 0

#### Scenario: A stranger's review after a teammate's

- GIVEN a teammate who is a collaborator requests changes on the head, and
  an account with no access then submits a later review on the same head
- WHEN the owner runs tend again
- THEN the implementer is asked once, with the teammate's review

### Requirement: A request answered with no change

If the implementer commits nothing for a requested change but writes its
reply, then the system shall post the reply once as a comment on the pull
request by the account, saying tend will not ask again, record
`answered-<review node id>`, review nothing again and exit 0; and if it
neither commits nor replies and is not working, the system shall stop for a
person on every run until a person commits or deletes `request-<review id>`,
which has the request asked once more.

#### Scenario: Out of scope

- GIVEN changes requested on the head, and an implementer that judges them out
  of scope and commits nothing
- WHEN the owner runs tend, then runs it again
- THEN both exit 0, the pull request has one comment by the account with the
  implementer's reply, the request was asked once and the head was reviewed
  once

#### Scenario: A request prompt that stalled

- GIVEN tend stopped because the request's prompt stalled, and the implementer
  committed nothing and is idle
- WHEN the owner runs tend again
- THEN it exits 3, saying the implementer committed nothing and is not working
  on them, without asking again; once a person deletes `request-<review id>`,
  the next run asks a second time and passes

## Known discrepancies

None recorded.

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
- [261122-fh9h2](../../changes/archive/261122-fh9h2-requested-changes-count-only-from-write-access/) Requested changes count only from people with write access
