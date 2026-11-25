# Handoff, 2026-11-24: trying step 3 live, step 4, and Claude Ship

Notes from one working session, for reading later. [direction.md](direction.md)
holds the decisions and [steps.md](steps.md) the order of work; this file adds
what was checked that day and what was recommended. Recommendations are marked
as such and are not decisions.

## Where things stand

`main` is at `9f3c79b`. Steps 1 to 3 of steps.md are built and merged; steps 4
to 6 have not started.

- tend 0.4.0 is released. It holds the flow file, building a change by name,
  several attempts at once, landing, and recovery with no one there.
- The drafter-name collision (issue #26) was fixed in pull request #32, after
  0.4.0. The release pull request for 0.4.1 (#33) is open and holds it.
- Grove has no approved change waiting.

Loose ends, none urgent:

- steps.md still says of step 3 "proposed as five changes, none approved or
  built yet". All five are merged (#21, #22, #23, #27, #29).
- Issue #28 is still open. Pull request #31 built it, except its third part,
  an `internal/flow` package, which the proposal left for a later change.
- `built()` in `attempt.go` has no caller. The refactor recorded this and
  left it.
- About 30 merged local branches remain, and one agent worktree under
  `.claude/worktrees/` whose work went in as pull request #15.
- Open issues: #25 (a cheaper test suite) and #24 (a dashboard of what waits
  for a person across repositories).

## Trying step 3 in a real repository

Nothing in step 3 has run against real GitHub or a real terminal. The test
suite drives the binary through stand-ins for `gh`, Herdr and rex. The
README's "What has been seen" ends on 2026-11-15, before any of step 3
existed. A live run is where real merge refusals and CI timing would show.

### What this MacBook lacked on 2026-11-24

- The installed tend was 0.3.0, which has none of step 3.
- `~/.local/state/tend` had no `publish` file, so tend here publishes
  nothing and could land nothing. The runner is presumably the mini, which
  was not looked at. *Corrected 2026-11-25: the runner is the separate OS
  user `tend` on the mini, not the owner's own user there.*
- `mascah/tend` had no ruleset on `main`.

### Setup, once per repository

1. Update tend on the runner to 0.4.1 or later (0.4.0 lacks the #26 fix,
   which matters when watch covers two repositories under Herdr).
2. List the repository's origin and the publishing account in
   `TEND_STATE/publish`.
3. Commit `tend/flow.yaml` to the target branch. Start from the output of
   `tend flow <repo>`, then add:
   - `watch: {at_once: 2}`;
   - the `land`, `resolve` and `fix` steps from the README's "Landing"
     section;
   - for a supervised flow, `publish: {ok: end, approved: land, requested:
     request}`; for an autonomous one, `publish: {ok: land, requested:
     request}`.
4. Add a branch rule on the target: require the project's CI and
   `tend/review`, require branches to be up to date, let the publishing
   account merge, and allow no bypass. *Corrected 2026-11-25: the publishing
   account gets no bypass, but the owner keeps the admin bypass. Proposal,
   release and hand-made pull requests never get `tend/review`, so the owner
   merges those with `gh pr merge N --squash --admin`. mascah/tend's ruleset
   28050972 is set this way.* Without CI on pull requests tend
   cannot land: a head whose only check is `tend/review` stops for a person.
5. Run `tend watch <repo>` in a Herdr pane or a rex terminal.

### What one change then looks like

1. You label an issue `tend`. A proposal pull request appears, with its open
   questions in a review.
2. You merge the proposal. That approves the change.
3. Within one watch pass (300 seconds by default) tend claims it and starts
   a worktree with Claude in a pane. Up to `at_once` attempts run side by
   side, where Grove says their planned files and requirements do not
   overlap. Two changes that overlap run one after the other.
4. Codex reviews each head, with one repair if the review asks. The
   publishing account opens the implementation pull request with a
   `tend/review` status.
5. Supervised: you approve the pull request on GitHub, and the next watch
   pass runs the attempt again and sends it to `land`. Autonomous: a head
   that passed review goes to `land` with no one asked.
6. At `land`, tend merges the target into the branch if the target moved,
   waits for every check on that head, then squash-merges. One candidate
   lands for each CI run; the next waits behind it.
7. A conflict or a failed check goes back to the implementer once, and the
   new head is reviewed again. Anything else stops, notifies you and frees
   the slot.
8. If tend dies, watch runs the attempt again. An agent still working is
   waited for; a step whose agent is gone is asked once more.

The person's steps are then three: label the issue, merge the proposal,
approve the pull request. In the autonomous flow they are two.

### Recommended for the first live run

- Start with the supervised flow, so every merge still has your approval.
- Use two small approved changes that do not overlap, or nothing runs in
  parallel.
- Watch for: how tend reads a real refusal from GitHub's merge API, how long
  a real CI run holds the train, whether a rule that dismisses approvals on
  new commits undoes your approval when tend merges the target in (the
  README warns of this), and whether the squash commit's subject passes this
  repository's pull-request-title check.

## Step 4: the autonomous flow

Step 4 takes the person out of approving each proposal, and adds a step
above requests: the owner states a goal and agents break it into requests.

### Its four parts

| Part | Kind | State |
|---|---|---|
| A daily spend cap, counting every prompt tend sends unattended, checked before each | Mechanical | Nearly settled |
| Protected paths: a candidate that touches the flow file, CI, the ruleset, CODEOWNERS, the principles file, `grove.yaml`, agent instructions, skills or deploy config waits for the owner | Mechanical | Settled (direction decision 3a) |
| A check step that judges a proposal, and the built change, for product identity, money and data | An agent's verdict | The flow file already runs verdict steps; who judges is open |
| Outcomes: work for an outcome the owner accepted flows; an outcome an agent adds is a pitch and waits | A new Grove concept | Mostly open |

### Why outcomes are the hard part

Until step 3, every merge to the target was a person's, so "merged" meant
"the owner approved". With a `land` step the publishing account merges too.
tend therefore needs another way to know that an outcome's text is the
owner's, and that a proposal stays inside it. Direction decision 14 settles
that an outcome has a permanent Grove ID and owner-approved text in Git. It
does not settle how a runner verifies that approval.

### What must be settled first (from steps.md)

1. How the owner's approval of an outcome is recorded and verified, how a
   change refers to its outcome, how a request records where it came from,
   and where broad goals and incoming requests are kept.
2. What a built pitch is: a throwaway to try, or the real change parked on
   an unmerged branch.
3. When no person approves a proposal, who answers its open questions, and
   which step makes the judgments in the table above.
4. Whether a project can ask for a person's approval before production. The
   direction document assumes so; the owner has not confirmed it.

### Recommended order

1. Build the spend cap and the protected-path check first. Neither waits on
   the four questions, and both make step 3 safer to leave running.
2. Explore outcomes (`/grove-explore`). It begins as Grove work, as step 2
   did.
3. Then the check step, with a report-only period first (see the next
   section), and only then letting tend approve a proposal.

## Claude Ship: what to borrow

Anthropic announced Claude Ship on 2026-11-23
([article](https://claude.com/resources/articles/introducing-claude-ship),
[docs](https://code.claude.com/docs/en/claude-ship)). It is three agents for
GitHub pull requests:

- **Code Review** reviews a pull request through several lenses, verifies
  its own findings adversarially, and posts them as inline comments ranked
  by severity.
- **PR Steward** fixes failing checks, merge conflicts and review feedback,
  and comments after each push with what it changed and how to undo it. It
  is designed not to approve or merge.
- **Claude Approvals** decides from a policy file in the repository,
  `.claude/approvals.yml`, which changes Claude may approve and how many
  people must approve changes to a path. It can use classifiers written in
  plain language. In observe mode it reports what the policy would decide
  and blocks nothing. It blocks a merge only where its check is made
  required in GitHub.

`/ship` in Claude Code opens the pull request, reviews it, then watches it
from the local session until it is merged or closed. The person merges.

How it sits beside tend: Ship starts where a pull request exists. tend also
covers request, proposal, approval and build, and reviews with a different
harness (Codex) from the one that implements. The two overlap from the pull
request onward.

Limits of this reading: it came from a summary of the article and two docs
pages. The policy-file and classifier pages were not read. The article says
Ship is for Team and Enterprise plans, and `mascah/tend` is on a personal
account, so treat these as designs to borrow, not a service to depend on.

### Recommended, in order of value

1. **A report-only mode for step 4's judge.** The check step posts its
   verdict on proposals the owner still merges. Compare its calls with
   yours before it gets authority. This makes the direction document's
   "later, with trust" measurable.
2. **A policy file for what still needs the owner.** Path rules for the
   protected files, plain-language rules for product identity, money and
   data, and the result as a GitHub check that the branch rule requires.
   tend already does this for review with the `tend/review` status. It
   would partly answer which step judges.
3. **Fix failing CI before a person is asked.** tend reads checks only at
   `land` (`landing.go`). In a supervised flow a pull request with failing
   CI waits for approval while nobody fixes it. A small change on its own,
   and it could be proposed now.
4. **Stop an attempt from GitHub.** Ship takes `@claude stop` on the pull
   request. tend parks an attempt only when someone interrupts the agent in
   its pane, which a phone cannot do.
5. **Verify findings before spending the repair.** The build flow allows
   one repair, so a false finding uses it up. This needs no change to tend:
   a project's flow file can add a verdict step that checks the reviewer's
   findings.

### Recommended to leave as it is

- **Plain comments.** Ship answers them. tend acts only on "Request
  changes" from a person with write access, a rule chosen on purpose
  (change 261122-fh9h2) so a stranger's text never reaches the implementer.
- **Where the work runs.** A `/ship` watch lives in a Claude Code session
  and ends with it. tend runs unattended on the runner.

Items 1 and 2 belong in the step 4 explore. Item 3 is independent.
