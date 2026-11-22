# Changes to watch-and-status

The changes that change 261122-rpkqa proposes to the requirements of
`watch-and-status`, the maintained spec at
[grove/specs/watch-and-status/spec.md](../../../../specs/watch-and-status/spec.md).
This is future work: nothing here describes the target branch until the change
is implemented and the maintained spec is reconciled.

## ADDED Requirements

### Requirement: Watch runs an approved attempt again

While `tend watch` runs, the system shall run tend again on each attempt
whose flow's `publish` gives a next step for `approved` and whose open pull
request's head has an approval that step would act on, as the landing spec's
"A person's approval is an outcome of publishing" says, unless the attempt
needs a person, saying which pull request it acts on; and in a repository
whose flow file sets the count, it shall do so within the count and before
starting new pieces.

#### Scenario: A teammate approves

- GIVEN a flow whose `publish` sends `approved` to `land`, and a published
  candidate that passed review
- WHEN one pass of `tend watch` runs, then a teammate who is a collaborator
  approves the head and another pass runs
- THEN the first runs nothing, and the second says the pull request was
  approved and lands it

#### Scenario: The built-in flow

- GIVEN a repository with no flow file, and a teammate approves a published
  head
- WHEN a pass of `tend watch` runs
- THEN it runs nothing

#### Scenario: An attempt that needs a person

- GIVEN an approval on the head of an attempt whose `status` says
  `needs-you`
- WHEN a pass of `tend watch` runs
- THEN it runs nothing
