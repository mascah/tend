# Work items

<!-- A maintained spec states the agreed behavior expected of the target branch.
It is not a report of what the code does today or of what is deployed: where
they differ, keep the rule and record the difference under "Known discrepancies". -->

## Purpose

`tend <repo> issue-<n>` and `tend <repo> <KEY>` build a GitHub issue or a Jira issue as the work item in a repository without Grove, approved by the `tend-build` label, and Jira projects reach tend through a mapping to repositories and credentials from the environment. A Jira key is a project key and a number, such as `SCRUM-7`. Owners of repositories without Grove, and anyone who may label issues there or edit issues in a mapped Jira project, rely on this; labelling or editing such an issue can start an agent.

## Boundaries

This covers what is particular to an issue as the work item: when it may be
built, its scope, branch and attempt name, when it is built, the line that
ends its commit message and the link back to a Jira issue; and, for every use
of Jira, the mapping, the credentials and how a description becomes text.
Everything after the build is as for any attempt: recovery, the lock and
exit statuses are [attempts](../attempts/spec.md); the review and the one
repair are [review](../review/spec.md); publication and requested changes are
[publication](../publication/spec.md). Turning issues labelled `tend` into
proposals is [intake](../intake/spec.md). tend never moves a Jira issue
through its workflow.

## Requirements

### Requirement: Not in a repository with Grove

If the repository uses Grove, that is `grove changes --json` succeeds in it,
then the system shall refuse to build an issue directly, saying to label it
`tend` for `tend propose` to draft a proposal, prompt no agent and exit 3.

#### Scenario: An issue in a Grove project

- GIVEN a repository with Grove and issue #7 labelled `tend-build`
- WHEN the owner runs `tend <repo> issue-7`
- THEN it exits 3 with no agent prompted, saying to label it tend for tend
  propose

### Requirement: Approved by tend-build, done when closed or merged

When asked to build an issue, the system shall do nothing and exit 0 if the
issue is closed (for Jira, its status category is Done) or the pull request
from its build branch is merged; and if the issue is open and lacks the
`tend-build` label, the system shall stop for a person, saying it is not
approved for building, and prompt no agent.

#### Scenario: Not labelled

- GIVEN open issue #7 labelled only `tend`
- WHEN the owner runs `tend <repo> issue-7`
- THEN it exits 3 saying it is not labelled tend-build, and no agent is
  prompted

#### Scenario: Closed

- GIVEN issue #7 is closed
- WHEN the owner runs `tend <repo> issue-7`
- THEN it exits 0 with no agent prompted

#### Scenario: Jira issues

- GIVEN mapped Jira issues SCRUM-7, open and unlabelled, and SCRUM-8,
  labelled `tend-build` and done
- WHEN the owner runs `tend <repo> SCRUM-7`, then `tend <repo> SCRUM-8`
- THEN the first exits 3 saying issue SCRUM-7 is not labelled tend-build, and
  the second exits 0 saying issue SCRUM-8 is closed

#### Scenario: Merged

- GIVEN issue #7's pull request from `tend/build-7` is merged
- WHEN the owner runs `tend <repo> issue-7`
- THEN it exits 0 saying it is merged and there is nothing to do, with no
  further review

### Requirement: The issue is the scope

When building an open issue labelled `tend-build`, the system shall work on
`tend/build-<n>` or `tend/build-<KEY>` in its worktree, wherever that is,
started from the repository's default branch; prompt the implementer with the
issue's title, description and URL as the agreed scope, asking it to commit on
that branch and write notes for the reviewer to `implemented` in the run
folder; count the build finished only when the branch has a commit and those
notes are not empty, stopping for a person otherwise; and tell the reviewer
the issue's words and where the notes are.

#### Scenario: A GitHub issue built

- GIVEN a repository without Grove, with origin listed for publication, and
  issue #7 "Request 7", "Make thing 7 better.", labelled `tend-build`
- WHEN the owner runs `tend <repo> issue-7`
- THEN the implementer is prompted `Implement GitHub issue #7 …` with
  "Request 7. Make thing 7 better.", the reviewer is told what it must do is
  `GitHub issue #7, quoted here: Request 7…` and where the notes are, branch
  `tend/build-7` is pushed, one pull request is opened by the account, and
  tend exits 0 with a passing verdict

#### Scenario: No notes for the reviewer

- GIVEN an implementer that commits but writes no notes
- WHEN the owner runs `tend <repo> issue-7`
- THEN it exits 3 saying the implementer wrote no notes for the reviewer

#### Scenario: Repaired

- GIVEN a reviewer that fails the first head as actionable
- WHEN the owner runs `tend <repo> issue-7`
- THEN the repair prompt says to update `TEND_STATE/runs/issue-7-<n>/implemented`
  where a claim changes, and tend exits 0 with a passing verdict

### Requirement: An issue's attempt is named by it and its repository

The system shall name an issue's attempt `issue-<n>-<cksum>`, or for a Jira
issue `jira-<key in lower case>-<cksum>`, where cksum is POSIX `cksum` of the
repository's path, so a rerun finds the attempt, its branch and its verdicts.

#### Scenario: A rerun of a GitHub issue

- GIVEN issue #7 built and reviewed
- WHEN the owner runs `tend <repo> issue-7` again
- THEN the run folder is `TEND_STATE/runs/issue-7-<cksum of the repository
  path>`, and the issue is neither implemented nor reviewed again

#### Scenario: A Jira issue

- WHEN the owner builds SCRUM-7
- THEN its run folder is `TEND_STATE/runs/jira-scrum-7-<cksum>`, and `tend
  status <repo> SCRUM-7` prints `attempt jira-scrum-7-<cksum> of SCRUM-7:
  tend/build-SCRUM-7`

### Requirement: The commit message names the issue

When an issue build is published, the system shall have its commit message
end with `Closes #<n>` for a GitHub issue or `Refs <KEY>` for a Jira issue,
and shall comment the pull request's URL on a Jira issue once, found by that
URL in its comments; and its reviews shall carry no `grove review` summary.

#### Scenario: GitHub

- WHEN issue #7's build is published
- THEN the implementer is asked to end the message with `Closes #7`, and no
  review says what merging approves

#### Scenario: Jira

- GIVEN Jira issue SCRUM-7 labelled `tend-build`
- WHEN the owner builds it, then runs tend on it again
- THEN the implementer is asked to end the message with `Refs SCRUM-7`, branch
  `tend/build-SCRUM-7` is pushed, SCRUM-7 has one comment `tend opened <pr
  url> …`, and the rerun builds, reviews and comments nothing again

### Requirement: Jira projects are mapped to repositories

When tend reads Jira, the system shall take a project only from a line
`<site> <project-key> <origin-url>` of `TEND_STATE/jira`; refuse, stopping for
a person, a Jira work item whose project is not mapped or is mapped to another
origin than the repository's; and stop if one project is mapped twice.

#### Scenario: Unmapped

- WHEN the owner runs `tend <repo> OTHER-1` with no line for `OTHER`
- THEN it exits 3 saying Jira project OTHER is not mapped

#### Scenario: Mapped to another repository

- GIVEN SCRUM is mapped to `git@github.com:someone/else.git`
- WHEN the owner runs `tend <repo> SCRUM-7`
- THEN it exits 3 saying it is not mapped to this repository's origin

### Requirement: Jira credentials stay off the command line

The system shall reach Jira's REST API v3 through `curl` with Basic
authentication from `JIRA_EMAIL` and `JIRA_API_TOKEN`, given on curl's
standard input and never in its arguments, and shall stop for a person when
either is unset.

#### Scenario: The token

- GIVEN `JIRA_API_TOKEN=tok-secret-1`
- WHEN `tend propose` reads and comments on Jira issues
- THEN no logged `curl` argument contains `tok-secret-1`

### Requirement: A Jira description as text

When tend gives a Jira issue's description to an agent, the system shall turn
its Atlassian Document Format into text keeping what carries scope: a line per
paragraph, heading and list item, code as written, every link's destination
after its text, and the URL of every smart link.

#### Scenario: Links, a list and code

- GIVEN SCRUM-1's description: "Implement" then "the spec" linked to
  `https://docs.test/spec`, "and" a smart link to
  `https://jira.test/browse/SCRUM-9`, a list item "keep totals" and a code
  block `ledger list --json`
- WHEN `tend propose` drafts it
- THEN the drafter's prompt reads `Request SCRUM-1. Implement the spec
  (https://docs.test/spec) and https://jira.test/browse/SCRUM-9 - keep totals
  ledger list --json (Jira issue https://jira.test/browse/SCRUM-1)`

## Known discrepancies

None recorded.

## Open questions

None.

## History

- Written when tend adopted Grove (2026-11-16).
