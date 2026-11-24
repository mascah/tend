# Changes to intake

The changes that change 261124-60tny proposes to the requirements of
`intake`, the maintained spec at
[grove/specs/intake/spec.md](../../../../specs/intake/spec.md). This is future
work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## MODIFIED Requirements

### Requirement: Draft, commit and open the proposal

When drafting a request, the system shall work on its proposal branch in a
worktree started from the target as origin has it, fetched first; prompt the
drafter `prop-<n>-<cksum>` or `prop-<key in lower case>-<cksum>`, where cksum
is POSIX `cksum` of the repository's path, with `/grove-propose --unattended`
followed by the request's title, description and URL; then ask it once to
commit the draft as one commit whose body ends `Refs #<n>` or `Refs <KEY>`
and to write the open questions, consequential ones first, to `summary`; push
the branch as the account and open a pull request onto the target titled and
described by that commit; and post a comment review on its head listing the
open questions and saying that merging approves the change with them open and
how to revise it. If the drafter leaves uncommitted files, or commits nothing
under `<root>/changes/`, then the system shall stop for a person.

#### Scenario: A proposal from an issue

- GIVEN issue #1, "Request 1", "Make thing 1 better.", labelled `tend`
- WHEN the owner runs `tend propose <repo>`
- THEN `prop-1-<cksum of the repository's path>` is prompted `/grove-propose
  --unattended Request 1. Make thing 1 better. (GitHub issue <url>)`, branch
  `tend/issue-1` on origin holds the change, the account opens a pull request
  onto `main` from it titled `Propose a thing` whose description ends `Refs
  #1`, and its review lists "Which thing?", names `/grove-propose
  261115-p0000` and says the implementation may then choose those answers

#### Scenario: Origin's target moved

- GIVEN someone else pushed to origin's `main` after the local clone
- WHEN `tend propose` drafts issue #1
- THEN the proposal's branch contains their commit and its pull request opens

#### Scenario: A Jira request

- GIVEN Jira issue SCRUM-1 labelled `tend`
- WHEN `tend propose` drafts it
- THEN the drafter `prop-scrum-1-<cksum>` is asked to end the commit with
  `Refs SCRUM-1`, the account opens a pull request from `tend/jira-SCRUM-1`,
  and its review says `Proposal drafted from SCRUM-1`

#### Scenario: Two repositories with the same issue number

- GIVEN two repositories that each have an issue #1 labelled `tend`, in one
  Herdr, and the first repository's drafter still open after its pass
- WHEN the owner runs `tend propose` on the second repository
- THEN tend starts a drafter named for the second repository, prompts only
  it, and the second repository's proposal is opened as a pull request
