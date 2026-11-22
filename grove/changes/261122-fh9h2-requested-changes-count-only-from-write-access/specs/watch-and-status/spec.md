# Changes to watch-and-status

The changes that change 261122-fh9h2 proposes to the requirements of
`watch-and-status`, the maintained spec at
[grove/specs/watch-and-status/spec.md](../../../../specs/watch-and-status/spec.md).
This is future work: nothing here describes the target branch until the change
is implemented and the maintained spec is reconciled.

## MODIFIED Requirements

### Requirement: Watch acts on a person's behalf

While `tend watch [<repo>...]` runs, the system shall, every `WATCH_SECONDS`
(default 300), first fast-forward each repository's local target to origin's,
fetched (only a fast-forward, and only of a clean checkout), since Grove reads
the local target; run `tend propose` on each repository given unless one of its
intake folders needs a person; then start on each repository given the oldest
cycle that Grove would claim (`runs` in `grove changes --json`) and whose plan
marks it `unattended`, saying so; and run tend again on each attempt whose
pull request is open with a change request on its head, from a person with
write access as the publication spec's "Requested changes are acted on once"
says, that tend has not answered, unless the attempt needs a person, saying
which pull request it acts on.

#### Scenario: Changes requested

- GIVEN a published attempt and no change requested
- WHEN one pass of `tend watch` runs (`WATCH_ONCE` set), then a teammate who
  is a collaborator requests changes and another pass runs
- THEN the first runs nothing, and the second says changes were requested on
  the pull request and the request is asked of the implementer

#### Scenario: An attempt that needs a person

- GIVEN a new change request on an attempt whose `status` says `needs-you`
- WHEN a pass of `tend watch` runs
- THEN the request is not asked

#### Scenario: A request already answered

- GIVEN a request the implementer answered with no change
- WHEN a pass of `tend watch` runs
- THEN it runs nothing; once a teammate requests changes again on the same
  head, the next pass runs tend and the new request is asked

#### Scenario: A stranger's request

- GIVEN a published attempt, and an account with no access to the repository
  requests changes on its head
- WHEN a pass of `tend watch` runs
- THEN it runs nothing

#### Scenario: Intake

- GIVEN issue #1 labelled `tend`
- WHEN a pass of `tend watch <repo>` runs
- THEN its proposal pull request opens; and once a request's follow-up has
  stalled and needs a person, a further pass asks nothing more

#### Scenario: A cycle marked unattended on GitHub

- GIVEN a cycle marked `unattended` on origin's target, which the repository's
  local target does not yet hold
- WHEN a pass of `tend watch <repo>` runs
- THEN the local target is fast-forwarded to origin's and the cycle is started

#### Scenario: A cycle marked unattended

- GIVEN a repository whose only cycle is not marked `unattended`
- WHEN one pass of `tend watch <repo>` runs, then the plan is marked
  `unattended` on the target and another pass runs, then a third
- THEN the first starts nothing; the second says it is starting the cycle,
  which is claimed, built and reviewed; the third starts nothing, as the
  cycle is claimed
