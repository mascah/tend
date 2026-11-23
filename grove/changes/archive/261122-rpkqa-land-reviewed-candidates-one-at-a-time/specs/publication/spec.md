# Changes to publication

The changes that change 261122-rpkqa proposes to the requirements of
`publication`, the maintained spec at
[grove/specs/publication/spec.md](../../../../specs/publication/spec.md). This
is future work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

The maintained spec's Boundaries ends "tend never merges a pull request;
merging stays with people". Reconciling this change makes that sentence say
that tend merges a pull request only at a `land` step, as the landing spec
says, and that in a flow without one merging stays with people.

## MODIFIED Requirements

### Requirement: The description follows later heads until a person edits it

When the open pull request's head has moved past the head its description was
written for, other than to a head carried from it as the landing spec's "The
verdict carries across a clean merge" says, the system shall have the
implementer write the message for the new head, given the current description
to keep what still holds, and set it as the pull request's title and
description; and if a person has edited the title or description since tend
set them, it shall leave them, say so, and not rewrite them again.

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

#### Scenario: A carried head

- GIVEN a published pull request whose head `land` moved by merging the
  target in
- WHEN tend runs the attempt again
- THEN the implementer is asked for no message, and the title and
  description are as they were
