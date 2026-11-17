# Changes to watch-and-status

## MODIFIED Requirements

### Requirement: Finished worktrees are tidied

When `tend <repo> <item>` or `tend propose <repo>` runs, the system shall
remove each worktree tend made for that repository whose pull request is
merged or closed, when its head is the pull request's head, and each
attempt's worktree with no pull request whose Grove cycle is done and whose
head the target holds, each with its local branch and its agents' place,
when it holds no uncommitted files, no worktree is inside it and its
implementer is not working, and otherwise leave it and say why; and a
declined request's worktree shall go at once.

#### Scenario: A merged attempt and an open proposal

- GIVEN a published attempt whose pull request is merged and an open proposal
- WHEN `tend propose` runs
- THEN the attempt's worktree and local branch are gone, and the proposal's
  worktree stays

#### Scenario: Uncommitted files

- GIVEN the proposal's pull request is merged and its worktree holds an
  uncommitted file
- WHEN `tend propose` runs
- THEN the worktree stays and tend says it has uncommitted files; once the
  file is gone, the next pass removes it

#### Scenario: A merged issue build

- GIVEN issue #7's pull request is merged
- WHEN the owner runs `tend <repo> issue-7`
- THEN the worktree of `tend/build-7` is removed

#### Scenario: The cycle landed some other way

- GIVEN an attempt with no pull request and no commits, whose cycle the
  target then shows done, with a worktree inside the attempt's
- WHEN `tend propose` runs, then runs again once that worktree is gone
  while the implementer is working, then again once it is idle
- THEN the first two leave the worktree, saying a worktree is inside it and
  that its implementer is working, and the third removes it with its local
  branch

#### Scenario: The cycle is not done

- GIVEN an attempt with no pull request whose cycle is not done
- WHEN `tend propose` runs
- THEN its worktree stays
