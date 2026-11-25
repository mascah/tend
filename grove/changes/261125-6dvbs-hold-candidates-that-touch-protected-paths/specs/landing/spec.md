# Changes to landing

<!-- The changes that change 261125-6dvbs proposes to the requirements of
`landing`. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled. -->

Maintained spec: [landing](../../../../specs/landing/spec.md), as of 2a8ae80.

The protected paths in effect are what the flows change spec's "The flow
file names the protected paths" says. A head's changed paths are those that
differ between the head and the commit where the head and the target's tip,
as origin holds it, fetched, meet. A renamed file counts under both its old
and its new name. A person with write access is who the publication spec's
"Requested changes are acted on once" says.

## ADDED Requirements

### Requirement: A head that touches a protected path waits for a person

When a flow reaches `land` with a head whose changed paths include a
protected path, the system shall land it only if the newest review that
approves or requests changes, by a person with write access, on that head or
on a head it is carried from, approves it; and otherwise it shall merge
nothing, push nothing, prompt no agent, and stop for a person, naming each
protected path the head changes and saying that a person with write access
must approve the head. The system shall decide this from Git and GitHub's
reviews alone, never from a verdict or any file an agent writes, and shall
decide it again, with the protected paths read afresh from the target,
before it has GitHub merge the pull request.

#### Scenario: An autonomous flow and an edit to the flow file

- GIVEN a flow whose `publish` sends `ok` to `land`, and a candidate that
  passed review, whose change edits `tend/flow.yaml` and `main.go`, with no
  approving review
- WHEN tend reaches `land`
- THEN tend exits 3 naming `tend/flow.yaml` and not `main.go`, the pull
  request stays open, and no merge commit is pushed to the branch

#### Scenario: A person approves the held head

- GIVEN that held candidate, whose head a teammate who is a collaborator
  then approves
- WHEN the owner runs tend again
- THEN the pull request is landed

#### Scenario: A stranger approves the held head

- GIVEN that held candidate, whose head an account with no access to the
  repository approves
- WHEN the owner runs tend again
- THEN tend exits 3 naming `tend/flow.yaml`, and nothing is merged

#### Scenario: No protected path

- GIVEN a flow whose `publish` sends `ok` to `land`, and a candidate that
  passed review, whose change touches only `main.go`, with no approving
  review
- WHEN tend reaches `land` and its checks pass
- THEN the pull request is landed

#### Scenario: A nested agent instruction file

- GIVEN a flow whose `publish` sends `ok` to `land`, no `protect` in the
  target's flow file, and a candidate that passed review, whose change adds
  `web/CLAUDE.md`
- WHEN tend reaches `land`
- THEN tend exits 3 naming `web/CLAUDE.md`, and nothing is merged

#### Scenario: A protected file moved away

- GIVEN a flow whose `publish` sends `ok` to `land`, and a candidate that
  passed review, whose change renames `.github/workflows/ci.yml` to
  `ci.yml.bak`
- WHEN tend reaches `land`
- THEN tend exits 3 naming `.github/workflows/ci.yml`, and nothing is merged

#### Scenario: The candidate unprotects a path

- GIVEN a target whose flow file sends `publish`'s `ok` to `land` and
  protects `schema.sql`, and a candidate that passed review, whose change
  edits `schema.sql` and drops it from `protect` in its own flow file
- WHEN tend reaches `land`
- THEN tend exits 3 naming `schema.sql` and `tend/flow.yaml`, and nothing is
  merged

#### Scenario: The target moves in a carried head

- GIVEN a flow whose `publish` sends `ok` to `land`, a candidate that passed
  review and edits `CLAUDE.md`, a teammate's approval of that head, and a
  commit that touches another file reaching the target
- WHEN tend merges the target in and the checks on the merge commit pass
- THEN the carried head is landed without a second approval, and the
  target's own commit is not counted as the candidate's change

#### Scenario: A head made at fix

- GIVEN a flow whose `publish` sends `ok` to `land` and whose `land` sends
  `failed` to `fix`, a candidate touching only `main.go` whose checks fail,
  and an implementer whose commit at `fix` edits `.github/workflows/ci.yml`
  and passes review
- WHEN tend reaches `land` with that head
- THEN tend exits 3 naming `.github/workflows/ci.yml`, and nothing is merged

#### Scenario: An approval before the implementer's new head

- GIVEN a flow whose `publish` sends `ok` to `land` and whose `land` sends
  `conflict` to `resolve`, a candidate that edits `grove.yaml` with a
  teammate's approval, and a conflict the implementer resolves at `resolve`
  in a new head that passes review
- WHEN tend reaches `land` with that head
- THEN tend exits 3 naming `grove.yaml`, and nothing is merged until a
  person with write access approves the new head

#### Scenario: A supervised flow

- GIVEN a flow whose `publish` sends `approved` to `land` and `ok` to the
  end, and a candidate that edits `tend/flow.yaml`, which a teammate
  approves
- WHEN the owner runs tend again
- THEN the pull request is landed as before
