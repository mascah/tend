# Changes to terminals

## MODIFIED Requirements

### Requirement: How agents start

When the system starts an agent, it shall start an implementer or drafter as
Claude Code with `--permission-mode auto --add-dir <run or intake folder>`,
`--disallowedTools "Bash(git push:*)" "Bash(gh pr:*)"` and
`--append-system-prompt` with tend's terms: the session works for tend on
its branch in this worktree, pushes nothing, opens or edits no pull request
and creates no branch or worktree, as tend delivers the work, and this wins
over any other instruction about delivering it, an implementer also being
told to commit its work on the branch and a drafter to commit only when tend
asks; a reviewer as Codex with `-s
workspace-write -a never --add-dir <run folder>`; and if an agent's pane is
gone, a new session shall take its name and find the work in Git.

#### Scenario: The implementer's flags

- WHEN tend starts the implementer of an attempt
- THEN Herdr is asked to start `impl-<attempt>` with `--add-dir
  TEND_STATE/runs/<attempt>`, `git push` and `gh pr` denied, and a system
  prompt naming the attempt's branch and telling it to commit there

#### Scenario: The drafter's flags

- WHEN `tend propose` starts a drafter
- THEN `git push` and `gh pr` are denied to it, and its system prompt says
  tend publishes the work and to commit only when tend asks
