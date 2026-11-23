# Changes to watch-and-status

The changes that change 261122-hjhmn proposes to the requirements of
`watch-and-status`, the maintained spec at
[grove/specs/watch-and-status/spec.md](../../../../specs/watch-and-status/spec.md).
This is future work: nothing here describes the target branch until the change
is implemented and the maintained spec is reconciled.

## ADDED Requirements

### Requirement: Watch runs again an attempt whose tend died

While `tend watch` runs, the system shall run tend again on each attempt of a
repository it was given whose `status` says `working`, that no tend is on
and, for a Grove cycle or change, that Grove shows as its attempt in flight,
saying which it runs again; and in a repository whose flow file sets the
count, it shall do so within the count and before starting new pieces.

#### Scenario: After a kill

- GIVEN a tend that was killed with `kill -9` while its implementer was on
  its turn, the implementer still working
- WHEN a pass of `tend watch <repo>` runs
- THEN it says it runs the attempt again, tend waits for the implementer,
  the implementer was prompted once in all, and the head is reviewed

#### Scenario: After the agents went too

- GIVEN that tend was killed and the implementer's pane is closed
- WHEN a pass of `tend watch <repo>` runs
- THEN the attempt is run again, the implementer is asked once more in a new
  session, and the head is reviewed

#### Scenario: A tend is on it

- GIVEN a tend whose implementer is on its turn
- WHEN a pass of `tend watch <repo>` runs
- THEN it runs nothing for that attempt

#### Scenario: An attempt that needs a person

- GIVEN an attempt whose `status` says `needs-you`
- WHEN a pass of `tend watch <repo>` runs
- THEN it runs nothing for that attempt

#### Scenario: An attempt retried since

- GIVEN an attempt left with `status` saying `working`, whose change was
  claimed again with `grove claim --retry` from another checkout
- WHEN a pass of `tend watch <repo>` runs
- THEN it runs nothing for the earlier attempt
