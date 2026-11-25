# Changes to flows

<!-- The changes that change 261125-6dvbs proposes to the requirements of
`flows`. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled. -->

Maintained spec: [flows](../../../../specs/flows/spec.md), as of 2a8ae80.

## ADDED Requirements

### Requirement: The flow file names the protected paths

The system shall take the protected paths in effect from the flow file on
the target branch, read as the flows are read: where the file holds
`protect`, its entries and `tend/`; otherwise tend's default list. An entry
is a path from the repository's root. An entry ending in `/` covers that
folder and everything in it, an entry with no `/` matches a file of that
name in any folder, and `*` matches within one path segment. `tend flow`
shall print the protected paths in effect and say whether they are the
project's or tend's own. If `protect` is not a list of entries that are not
empty, the system shall stop as "A flow that is not valid" says.

#### Scenario: The default list

- GIVEN a repository with no flow file
- WHEN the owner runs `tend flow <repo>`
- THEN it prints `protect` with tend's default list, holding `tend/`,
  `.github/workflows/`, `CODEOWNERS`, `PRINCIPLES.md`, `grove.yaml`,
  `CLAUDE.md`, `AGENTS.md`, `.claude/`, deploy config and lockfiles, and
  says it is tend's own

#### Scenario: A project's own list

- GIVEN a flow file whose `protect` holds `schema.sql` and `migrations/`
- WHEN the owner runs `tend flow <repo>`
- THEN it prints `protect` with `schema.sql`, `migrations/` and `tend/`,
  says it is the project's, and does not hold `go.sum`

#### Scenario: An empty entry

- GIVEN a flow file whose `protect` holds an empty entry
- WHEN the owner runs tend
- THEN tend exits 3 with one sentence naming `tend/flow.yaml` and
  `protect`, and no agent is started

#### Scenario: The implementer's list is not read

- GIVEN a target whose flow file protects `schema.sql`, and an attempt whose
  implementer commits a flow file with no `protect`
- WHEN tend reaches `land` with a head that edits `schema.sql`
- THEN the head is held as the landing change spec's "A head that touches a
  protected path waits for a person" says
