# Close the request when its change's implementation is merged

## Problem

A request that tend turned into a change stays open after that change's
implementation is merged
([issue #54](https://github.com/mascah/tend/issues/54)).

- The issue reports that #36 and #37 were labelled `tend`, proposed (#39,
  #40), built and merged (#42, #48), and stayed open and labelled until the
  owner closed them by hand. None of those pull requests links its issue
  (`closingIssuesReferences` is empty); the link exists only in the
  proposal's prose.
- At 7e339d9, an attempt on a Grove cycle ends its squash commit message with
  `Grove: <change ids>` and nothing else (`attempt.go`, where `trailer` is
  set for `kind == "grove"`); only an issue build, which has no proposal,
  ends with `Closes #<n>` (`item.closes`, `request.go`).
- At 7e339d9, intake's run folder for a request
  (`TEND_STATE/intake/<owner>_<repo>-<n>` or `…-jira-<KEY>`) records `id`,
  `worktree`, `pr` and `linked`, but not which changes the proposal drafted:
  `proposal` in `intake.go` computes the change IDs the branch adds only to
  name them in the summary review.
- Intake skips a request that already has a pull request from its branch
  (`propose` in `intake.go`), so the open issue does no harm, but it looks
  unfinished and nobody closes it.

## Intended outcome

When tend publishes the implementation of a change that intake drafted from
a GitHub issue, the pull request's description has the line `Closes #<n>`,
so GitHub closes the issue when the pull request is merged, whoever merges
it: tend at `land`, or the owner with bypass. tend finds the request from
its own records: intake writes the IDs of the changes a proposal adds to the
request's intake folder when it publishes the proposal, and publication
looks the attempt's changes up there. The line stays through tend rewriting
the description and is never there twice. A cycle or change drafted from
several requests closes each. For a Jira request tend comments the
implementation pull request on the Jira issue once and does not transition
it. A change no request drafted gets no line and no comment.

## Scope

### In scope

- Intake recording, in the request's intake folder, the IDs of the changes
  the proposal's branch adds, when it publishes the proposal.
- Publication of a Grove attempt: a `Closes #<n>` line in the pull request's
  description for each GitHub request that drafted one of its changes, when
  the pull request opens and each time tend rewrites the description, added
  only where the description lacks it.
- A comment, once, on each Jira request that drafted one of its changes,
  with the implementation pull request's URL.
- The README and guide pages that list the intake folder's files and say how
  a pull request is described.

### Out of scope

- Closing a request when its proposal is merged: approval is not done (the
  request's words).
- Closing requests from before this change: the owner closed #36 and #37 by
  hand. Their intake folders hold no change IDs, so they get no line.
- Transitioning a Jira issue through its workflow.
- Issue builds (`tend <repo> issue-<n>`), which already end their message
  with `Closes #<n>` or `Refs <KEY>` as the work-items spec says.
- Parsing the request from the proposal's prose or its `Refs #<n>` line.

## Decisions

None was made beyond the request. The request asks that the implementation
pull request of a change intake drafted from a GitHub issue "says `Closes
#<n>`" so that "GitHub then closes the issue on merge, whoever merges it";
that "tend finds the request from its own records … The link is not parsed
from the proposal's prose"; that "the line survives tend rewriting the
description, and is not added twice"; that "a cycle or change built from
several requests closes each of them"; that for a Jira request tend "does
not transition the issue. It comments on it with the implementation pull
request"; and that "a change no request drafted, such as one written by a
person, gets no such line".

## Open questions

This draft is written to the recommended answers. None must be answered
before review; each can wait, and `/grove-implement` builds the recommended
answer and records it for the reviewer.

1. **How publication finds the requests.** Recommended: when intake
   publishes a proposal, it writes the IDs of the changes the branch adds
   under `<root>/changes/` to `changes` in the request's intake folder, the
   same IDs the summary review names; publication of a Grove attempt reads
   the intake folders of the repository (declined ones left out) and takes
   each whose `changes` lists one of the attempt's changes. Reason: the
   request asks for tend's own records rather than the proposal's prose, the
   IDs are already computed there, and a folder written before this change
   has no `changes`, which keeps #36 and #37 out as the scope says. The
   alternative, reading the proposal pull request's files from GitHub at
   publish time, costs a call per request and still starts from the intake
   folder. Decides T-01 and T-02.
2. **Who writes the line.** Recommended: tend adds each missing `Closes
   #<n>` line, one per line, after the implementer's body in the description
   it sets, rather than asking the implementer to end the message with it.
   Reason: a line tend adds cannot be dropped or reworded by the agent, a
   check for the exact line keeps it from appearing twice, and since the pull
   request is squash-merged its description becomes the commit body, so the
   merged commit carries the line too. The implementer is still asked to end
   with `Grove: <change ids>`. Decides T-02.
3. **A description a person edited.** Recommended: tend leaves it as it is,
   and adds no missing line to it, as "The description follows later heads
   until a person edits it" already says. Reason: the line is there from the
   moment the pull request opens, so it is missing only if a person took it
   out, and their edit is theirs. Decides T-02.
4. **A change added to the proposal after intake published it.**
   Recommended: `changes` records what the branch adds when intake publishes
   the proposal (including when a later pass finishes an interrupted
   handoff), and is not read again; a change a person adds to the branch
   afterwards is not linked to the request. Reason: intake does not revisit
   a request whose folder records `linked`, and adding a change to a drafted
   proposal by hand is rare; the request's own branch can still be revised
   with `/grove-propose <change-id>`, which keeps the ID. Decides T-01.
5. **A target that is not the repository's default branch.** Recommended:
   accept that GitHub honours `Closes #<n>` only on pull requests into the
   default branch, and say so in the guide; tend does not close the issue
   itself. Reason: the request relies on GitHub closing the issue "whoever
   merges it", and closing from tend would cover only merges tend makes.
   Decides T-03.
6. **Routine choices inside this approach.** Recommended, each the
   implementer's to make and record: the closing lines in ascending issue
   number; a Jira request getting no `Refs <KEY>` line in the description,
   only the comment, since the request asks only for the comment; the Jira
   comment's wording, found by the pull request's URL in its comments as the
   existing Jira links are; the name and layout of tests and private
   helpers. Decides T-01 and T-02.

## Related documents

- Requirements: [specs/publication/spec.md](specs/publication/spec.md),
  [specs/intake/spec.md](specs/intake/spec.md)
- Tasks: [tasks.md](tasks.md)
