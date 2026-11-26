# Changes to publication

<!-- The changes that change 261125-rf036 proposes to the requirements of
`publication`. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled. -->

Starts from the maintained [publication spec](../../../../specs/publication/spec.md)
at 2a8ae80. A check, failed, passed and the wait's bound are as the landing
spec's "Checks decide" says.

## ADDED Requirements

### Requirement: Checks are waited for before a person is asked

Where a build flow's `publish` gives a next step for `failed`, the system
shall, after publishing a head and finding no change a person requested on
it, wait for every check GitHub reports on that head other than
`tend/review`, as the landing spec's "Checks decide" reads them, before it
reads an approval or gives `ok`; if one failed and the step `failed` leads
to can still be asked, it shall give `failed`, giving that step the failed
checks' names and links as `{checks}`; and if no check but `tend/review` is
reported, or one is unfinished, when the wait ends, it shall stop for a
person saying so. In a flow whose `publish` gives nothing for `failed`, the
system shall read no check at `publish`.

#### Scenario: A red pull request in a supervised flow

- GIVEN a flow whose `publish` sends `ok` to the end, `approved` to `land`
  and `failed` to `fix`, a `fix` step continuing the implementer with
  `prompts/failed.md` and limit 1, and a head that passed review on which the
  check `test` fails once the pull request opens
- WHEN tend runs the attempt
- THEN `fix` is prompted naming `test` and its link, the head it makes is
  reviewed and published, its checks are waited for, and once they pass tend
  exits 0 with the pull request open and nothing merged

#### Scenario: The checks pass

- GIVEN that flow and a head on which every check passes
- WHEN tend runs the attempt
- THEN no `fix` is prompted, and tend exits 0 with the pull request open

#### Scenario: Failed review is not a failed check

- GIVEN that flow, a head that failed review with `actionable` false, so its
  `tend/review` status is `failure`, and every other check on it passes
- WHEN tend publishes it
- THEN no `fix` is prompted, and tend exits 0 with the failed verdict

#### Scenario: A change requested first

- GIVEN that flow, a published head whose checks have not finished, and a
  change a collaborator requested on it
- WHEN tend runs the attempt again
- THEN the request is asked of the implementer without waiting for the
  checks

#### Scenario: Approved but red

- GIVEN that flow, and a teammate's approval of a head on which a check
  fails
- WHEN tend runs the attempt again
- THEN `fix` is prompted, nothing is merged, and the new head needs a new
  approval

#### Scenario: No CI

- GIVEN that flow, and a head on which GitHub reports no check but
  `tend/review` until the wait ends
- WHEN tend publishes it
- THEN tend exits 3 saying the head has no check but `tend/review`

#### Scenario: A flow without failed

- GIVEN a repository with no flow file, and a published head on which a
  check fails
- WHEN tend runs the attempt
- THEN tend reads no check and exits 0 as before

### Requirement: A spent fix is said on the pull request

When a check on the published head failed and the step `publish` gives for
`failed` is at its limit, the system shall post one comment on the pull
request by the account saying the head's checks failed after the one fix
tend asks for, naming the failed checks and their links, post no second such
comment for that head, and give the outcome `publish` would give without the
wait.

#### Scenario: Still red after the fix

- GIVEN the supervised flow above, and a check that fails on the head and
  again on the head `fix` makes
- WHEN tend runs the attempt
- THEN `fix` is prompted once, the pull request has one comment naming the
  failed check on the second head, and tend exits 0 waiting for a person

#### Scenario: One fix in all

- GIVEN that attempt, after which a teammate approves the second head
- WHEN tend runs the attempt again and `land` finds the check still failed
- THEN no `fix` is prompted, nothing is merged, and tend exits 3 saying `fix`
  was asked once, its limit

#### Scenario: A rerun

- GIVEN that attempt, run again with no new commit
- WHEN tend publishes the second head again
- THEN the pull request still has one such comment
