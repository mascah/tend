# Changes to publication

The changes that change 261122-fh9h2 proposes to the requirements of
`publication`, the maintained spec at
[grove/specs/publication/spec.md](../../../../specs/publication/spec.md). This
is future work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## MODIFIED Requirements

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
