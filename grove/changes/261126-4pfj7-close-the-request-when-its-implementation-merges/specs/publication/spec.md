# Changes to publication

<!-- The changes that change 261126-4pfj7 proposes to the requirements of
`publication`. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled. -->

Maintained spec: [publication](../../../../specs/publication/spec.md), at
7e339d9.

## ADDED Requirements

### Requirement: Close the requests a change was drafted from

When the system publishes a Grove attempt's head, the requests it closes
shall be the GitHub issues whose intake folder in this repository, not
renamed as declined, lists in `changes` one of the attempt's changes, and
the pull request's description shall be the message's body followed by a
line `Closes #<n>` for each of them that the body does not already hold as
a line of its own, in ascending issue number; an attempt whose changes no
intake folder lists shall get no such line.

#### Scenario: A change drafted from an issue

- GIVEN change `261115-p0000`, which intake drafted from issue #1, approved
  and built
- WHEN tend publishes its reviewed head
- THEN the pull request's description is the message's body followed by
  `Closes #1`

#### Scenario: A cycle drafted from two requests

- GIVEN a cycle of the changes intake drafted from issues #2 and #1
- WHEN tend publishes its reviewed head
- THEN the description ends `Closes #1` then `Closes #2`, each once

#### Scenario: The message already says it

- GIVEN the implementer's message for a change drafted from issue #1 holds
  the line `Closes #1`
- WHEN tend opens the pull request
- THEN the description holds `Closes #1` once

#### Scenario: A change no request drafted

- GIVEN an approved change a person wrote, listed in no intake folder
- WHEN tend publishes its reviewed head
- THEN the description is the message's body with no `Closes` line, and no
  Jira issue is commented on

### Requirement: Tell a Jira request where its implementation is

When the system publishes a Grove attempt's head and a Jira issue's intake
folder in this repository, not renamed as declined, lists in `changes` one
of the attempt's changes, it shall comment the pull request's URL on that
Jira issue once, found by that URL in its comments, and shall not move the
issue through its workflow.

#### Scenario: A change drafted from a Jira issue

- GIVEN change `261115-p0000`, which intake drafted from Jira issue SCRUM-1,
  approved and built
- WHEN tend publishes its reviewed head, then runs the attempt again
- THEN SCRUM-1 has one comment `tend opened <pr url> …` for the
  implementation, beside the proposal's, its status is unchanged, and the
  description has no `Closes` line for it

## MODIFIED Requirements

### Requirement: One pull request, described by the commit message

When the branch has no pull request on GitHub in any state, the system shall
have the implementer write the squash commit message for the head to
`message-<head>` in the run folder, once per head, ending with the work
item's line (`Grove: <change ids>` for a cycle), and open one pull request
onto the target, titled by the message's subject and described by its body
with the closing lines "Close the requests a change was drafted from" adds;
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
written for, other than to a head carried from it as the landing spec's "The
verdict carries across a clean merge" says, the system shall have the
implementer write the message for the new head, given the current description
to keep what still holds, and set it as the pull request's title and
description, with the closing lines "Close the requests a change was drafted
from" adds; and if a person has edited the title or description since tend
set them, it shall leave them, say so, and not rewrite them again.

#### Scenario: A new head after a requested change

- GIVEN a published pull request titled `Add the file`, and a new head made
  for a requested change
- WHEN tend publishes the new head
- THEN the title becomes `Add the file, plainer` and the description is the
  new message's body

#### Scenario: The closing line through a rewrite

- GIVEN a published pull request for a change drafted from issue #1, whose
  description tend set with `Closes #1`, and a new head made for a requested
  change
- WHEN tend publishes the new head
- THEN the description is the new message's body followed by `Closes #1`,
  once, and tend does not say that a person has edited it

#### Scenario: A description a person edited

- GIVEN a person replaced the description with their own words, then
  requested changes
- WHEN tend publishes the new head
- THEN the description is still their words, no `gh pr edit` runs, and tend
  says a person has edited it

#### Scenario: A carried head

- GIVEN a published pull request whose head `land` moved by merging the
  target in
- WHEN tend runs the attempt again
- THEN the implementer is asked for no message, and the title and
  description are as they were

## Observed current behavior

- At 7e339d9 a Grove attempt's message ends with `Grove: <change ids>` only
  (`attempt.go`, `trailer`), and `publish` in `publication.go` describes the
  pull request by the message's body as `titleBody` splits it; `redescribe`
  treats any description other than that body as a person's edit.
- At 7e339d9 an issue build's message ends with `Closes #<n>` or `Refs
  <KEY>` (`request.go`, `closes`), and `publish` comments on its Jira issue
  through `item.link`; a Grove attempt has no `item`, so it links no request.
