# Changes to review

The changes that change 261122-rpkqa proposes to the requirements of
`review`, the maintained spec at
[grove/specs/review/spec.md](../../../../specs/review/spec.md). This is future
work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## MODIFIED Requirements

### Requirement: A fresh reviewer for each head

When an attempt's head has no valid verdict and is not carried from a head
that passed review, as the landing spec's "The verdict carries across a clean
merge" says, the system shall start reviewer `rev-<short head>` as Codex with
`-s workspace-write -c sandbox_workspace_write.network_access=true -a never
--add-dir <run folder>`, reusing one an earlier run left, and prompt it once
with the branch, the head, the change as `git diff <base>...HEAD`, what the
work item must do and where the implementer's claims are, asking for material
findings only, no edits or commits, and the verdict written to
`verdict-<head>.json` in the run folder as `head`, `conclusion`, `actionable`
and `summary`; and once the verdict is valid it shall close the reviewer's
pane.

#### Scenario: The reviewer's one writable folder

- WHEN tend has a cycle's head reviewed
- THEN the reviewer is started with `--add-dir TEND_STATE/runs/<attempt>`,
  and the verdict file is there

#### Scenario: A reviewer left working

- GIVEN tend stopped because the reviewer's prompt timed out, and the reviewer
  is still working
- WHEN the owner runs tend again
- THEN the reviewer is started once and prompted once in all

#### Scenario: A carried head

- GIVEN an attempt whose head is a clean merge of the target into a head
  that passed review, made at `land`
- WHEN the owner runs tend again
- THEN no reviewer is started for the merge commit, and the flow goes on as
  from a passing review
