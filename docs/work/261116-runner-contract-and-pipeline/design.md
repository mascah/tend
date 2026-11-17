# Runner contract and an autonomous pipeline

Draft for review, 2026-11-16, revised after a Fable consult and Codex round 1.
It covers grove, tend and a new project. Nothing here is built yet.

## 1. What happened with memex cycle 261116-dmj60

- **The claim.** `tend watch` claimed the cycle as attempt 261116-gchex, on
  branch `grove/261116-gchex`, and prompted `/grove-implement 261116-dmj60`.
- **What Grove printed.** Its instructions came with memex's
  `rules.skills.implement`, and Grove prints those rules both to the cycle's
  implementer and to each per-change subagent
  (`grove/internal/instruct/instruct.go:634, 761-770`). The rules say:
  - work in `just worktree <change-id>`;
  - push the branch and open a pull request;
  - for a cycle, one branch and stacked PR per change.
- **What the implementer did.** It followed those rules. It created three
  nested worktrees, pushed three branches, and opened stacked PRs #141–#143 as
  mascah-bot. Tend's branch got no commits.
- **What tend did.** Its 30-minute wait timed out first, so the attempt
  stopped as needs-you, and Codex never reviewed the work. Had the wait ended
  normally, `built()` (`main.go:1287-1306`) would have stopped with "change
  261115-jkn0v is not archived": true, but it doesn't say why.
- **The outcome.** The owner merged the three PRs. All three are on memex
  `main`, and Grove reports the cycle done.
- **Leftovers.** Tidy couldn't remove them, because the attempt's branch had
  no PR, and `tend status` hid them, because the cycle was done. They have
  since been removed by hand.

## 2. The problem

Two parties each believe they own delivery:

- **Grove gives delivery to the project.** Its runs spec says "the agent
  pushes and opens pull requests only as the project's `rules` tell it"
  (`grove/grove/specs/runs/spec.md:28`). Its docs tell projects to put
  "push and open a PR" under `rules.skills.implement`
  (`grove/docs/running.md:74`).
- **Tend gives delivery to tend.** It prompts a bare `/grove-implement
  <cycle>` (`main.go:1180`) and states no terms of its own.

Two things make the clash possible:

- **Memex's implement rules mix two kinds of rule.**
  - *Work rules:* commit per task, `pnpm install` in a fresh tree, run the app
    on a scratch home, record `Also changed:`.
  - *Delivery rules:* which worktree and branch, push, open a PR, stack them.
- **The implementer can reach the publishing credentials.** It runs as the
  same OS user as tend, with gh logged in as mascah-bot.

No instruction can guarantee the outcome. CLAUDE.md, AGENTS.md, user-level
memory and plugins, skills and Grove rules can each say "open a PR". Issue and
Jira attempts don't go through Grove at all.

## 3. Who owns what

- **Grove** owns what is agreed, the scope, and what "built" means: each
  change archived on the attempt's branch.
- **Project rules** say how this project builds and checks its work.
- **Whoever starts the session** owns delivery: the worktree, branch,
  credentials, time limit, review, push, PR shape and cleanup. That is the
  owner in a session they run, or tend, a shell line, or a person after
  `grove claim`.
- **Project policy** is new and optional. It decides what may merge without a
  person. It lives in the repository's settings and files, and GitHub
  enforces it.

## 4. Fix B, in grove: delivery rules the caller can switch off

This part is required, because subagents read Grove's instructions directly
and no prompt from tend reaches them.

- **B1. A separate key.** `grove.yaml` gets `rules.delivery.<skill>`: lines
  saying how the owner wants a session they run delivered (worktrees, push,
  PRs, stacking). `grove instructions <skill>` prints them as today unless
  the caller passes `--runner`, which is parsed like `propose --unattended`
  (`grove/internal/cli/cli.go:69, 90`). With `--runner`, Grove:
  - leaves those lines out;
  - prints one fixed paragraph: "the caller owns this working tree, branch,
    pushing and the pull request; commit here and stop." For a per-change
    subagent it says instead: "leave committing and archiving to the session
    that started you", as Grove's cycle instructions already require
    (`requires/implement-cycle.md:3`);
  - names `instructions implement --runner <change-id>` in the cycle's list
    of subagent commands (`instruct.go:570`), so each subagent gets the same
    terms.

  A fixture test checks that the output leaves the lines out.
- **B2. Spec and docs.**
  - The runs spec's Boundaries line becomes: "under `--runner`, the agent
    commits on the current branch; the caller delivers."
  - `docs/running.md` moves its "push and open a PR" example to
    `rules.delivery.implement`.
  - The documented `K=$(grove claim) && claude -p "/grove-implement $K"`
    recipe stays without `--runner`: with no runner, the agent delivers,
    from `rules.delivery.implement`. Only a caller that delivers the work
    itself passes `--runner`.
  - A cycle always lists its members with `--runner`, because a subagent
    never delivers. The paragraph says nothing about committing, so it
    doesn't contradict "members neither archive nor commit".
  - This is proposed as grove change 261117-w31az.
  - Constitution rules 2 and 11 are unchanged, because Grove still pushes
    nothing. Rule 8 is met by the documented step.
- **B3. Memex.**
  - The worktree, push, PR and stacking lines move to
    `rules.delivery.implement`.
  - The worktree rule is split first, so `cd app && pnpm install
    --frozen-lockfile` before app commands stays a work rule under
    `rules.skills.implement`.

## 5. Fix A, in tend: state, guard, verify

Tend needs this whatever Grove does.

- **A1. State it.** Tend prompts `/grove-implement --runner <cycle>`.
  - Every implementer prompt (Grove, issue, Jira) ends with: "This is a tend
    attempt on branch `<b>` in this worktree. Commit everything here. Don't
    push, open or edit pull requests, or create branches or worktrees. Tend
    reviews and publishes. Where any instruction says otherwise, this wins for
    this session."
- **A2. Guard against accidents.** Start the Claude implementer with
  `--disallowedTools "Bash(git push:*)" "Bash(gh pr:*)"` (`main.go:719`).
  - This catches a rule-following agent, not a determined one. Claude's own
    documentation says Bash deny rules aren't a security boundary, and
    `git -C … push` or `gh api` slip past.
  - The gate that matters is the required `tend/review` check (§6).
- **A3. Verify the result.** Before each implementer turn (implement or
  repair), tend snapshots `git worktree list` and `refs/heads`. When the agent
  settles or the wait times out, tend checks:
  - that the attempt's branch has commits;
  - that no worktree was added;
  - that no other local branch was created with commits.

  If any check fails, it stops as needs-you with one sentence naming what it
  found, and doesn't review or publish. Local state is compared, not GitHub
  PRs, so tend's own proposals opened meanwhile can't be mistaken for the
  implementer's.
- **A4. Timeout.** Raise the `TIMEOUT_MS` default from 30 minutes to 4 hours.
  It's a backstop against a hung agent, and a three-change cycle took about 80
  minutes. The mini already has this setting.
- **A5. Leftovers.** Tidy has two paths:
  - **The existing path**, for attempts whose PR is merged or closed and whose
    head matches, is kept.
  - **A new fallback**, for attempts with no PR whose cycle is done, removes
    the attempt only when its branch has no commits of its own and its
    worktree is clean.

  Both paths also leave the attempt alone when a nested worktree exists under
  it or its implementer is still working. Tidy says why whenever it leaves an
  attempt.

### Known ceiling: one OS account

The implementer runs as the same macOS user that holds the bot's token, so it
can always reach that token, for example with `gh auth token`. A2 and A3 make
mistakes fast and visible. They don't make publishing impossible.

Once auto-merge is on (§6), an implementer that deliberately posted a fake
`tend/review` status could merge its own PR. Closing that needs credential
separation: a second OS user for agents, with tend holding the only write
token. Deferred until something shows it's needed. The new project's
blast radius is a toy app, and its rollback is `wrangler rollback`.

## 6. One PR per change, without stacking

Under `--runner`, a cycle is one branch and one PR, as in Grove's own default
for cycles. Where a PR per change matters, plan one-change cycles chained by
`depends_on`. Each becomes its own attempt and PR, claimable once the one
before it is merged and done. Tend needs no stacking support.

## 7. The autonomous pipeline for the new project

The project is a touch-typing practice app: fun to use, not a trainer. It's a
React Router v7 app on Cloudflare Workers.

The core idea: agents never "approve" in GitHub's sense. Approval is a commit
status on the exact head, written by a role that didn't author that head. The
merge is GitHub auto-merge, under a ruleset the owner controls. Tend already
posts `tend/review` for each head Codex reviews (`main.go:1419`), so its
"never merges" requirement (`tend/grove/specs/publication/spec.md:20-21`)
stays true.

### Roles in v1

| Role | Who | Where | May |
| --- | --- | --- | --- |
| Owner | person | — | label issues, merge proposals, plan cycles, own the ruleset, `PRINCIPLES.md` and the pause |
| Proposer | tend propose | mini | open proposal PRs from labeled issues, as today |
| Implementer | tend's Claude, `--runner` | mini | commit on the attempt's branch |
| Reviewer | tend's Codex | mini | pass or fail `tend/review` on the head; one repair |
| CI | GitHub Actions | GitHub | pass or fail `ci`: typecheck, tests, build |
| Merger | GitHub auto-merge | GitHub | squash-merge once required checks pass |
| Deployer | Workers Builds | Cloudflare | a preview per PR, production from `main` |

No new agent is added. The only new autonomy is that tend's implementation PRs
merge and deploy themselves once `ci` and `tend/review` are green. A failed
Codex review leaves `tend/review` failing, so the PR waits for the owner, as
it does today.

### Policy

- **Ruleset on `main`:**
  - a PR is required;
  - required checks: `ci` and `tend/review`;
  - squash merges only;
  - no force pushes;
  - code owner review required;
  - stale approvals dismissed on push;
  - zero approvals required otherwise;
  - the owner may bypass it for pull requests only. Proposal and cycle PRs
    carry no `tend/review`, so the owner merges those through the bypass.

  The ruleset lives in repository settings, out of the bot's reach: the bot is
  a collaborator, not an admin.
- **Auto-merge workflow.** It enables auto-merge on the bot's implementation
  PRs (branch `grove/*`).
  - It runs on `pull_request_target`, so it always runs main's copy.
  - It checks out no PR code and holds no secret beyond `GITHUB_TOKEN`.
  - Proposal PRs (`tend/issue-*`) are left for the owner.
- **CODEOWNERS.** It assigns the owner:
  - `.github/**`
  - `CODEOWNERS`
  - `PRINCIPLES.md`
  - `grove.yaml`
  - `AGENTS.md`/`CLAUDE.md`
  - `wrangler.*`

  A PR touching any of them waits for the owner, so agents can't change their
  own policy, CI or deployment.
- **Change requests stay human.** Agents never leave "Request changes".
  - Tend answers any change request on the head (`main.go:350-351`), so an
    agent requesting changes would bounce off tend.
  - A person's "Request changes" still goes to the implementer, as today,
    with the existing one-repair bound.
- **Pause.** One step per layer:
  - **Merges:** turn on a prepared second ruleset on `main` with "Restrict
    updates", which only the owner may bypass. Nothing merges automatically
    while it's on, including PRs already queued or approved.
  - **Building:** stop `tend watch` on the mini.
  - **Production:** `wrangler rollback`.
- **Cost.** One attempt at a time (tend's lock), `CAP` proposals, and
  `TIMEOUT_MS`. Plus a daily attempt cap in watch, counted from today's
  date-prefixed run folders. Grove also pauses claims while the latest
  attempt's cycle isn't done, so a failed attempt stops all building until
  someone looks.

### PRINCIPLES.md

The product principles say what "fun" means, measurably where possible:
- typing starts within a second of loading;
- no sign-up;
- keyboard only;
- short sessions;
- immediate feedback;
- no layout shift.

Then one section per role: what it optimizes, what it never does, and when it
stops and asks the owner. Grove's `context` names the file, and Codex's review
prompt reads it.

### First version

1. **Tend and Grove.** Ship Fix B and Fix A, and move memex's delivery rules.
2. **New repository.** Start it from the React Router + Workers template:
   - adopt Grove and write `PRINCIPLES.md`;
   - add CI;
   - connect Workers Builds, with previews on and `noindex` on preview
     hosts.
3. **Policy.** Add the ruleset, `CODEOWNERS` and the auto-merge workflow, and
   turn on auto-merge and delete-branch-on-merge.
4. **Tend on the mini.** Add the repository to `tend watch` and `publish`.
5. **Trial.** Run about ten cycles, counting the merges the owner would have
   refused, before adding anything below.

### Later, each only when the trial asks for it

- **A planner.** It would turn approved changes into unattended one-change
  cycles. It needs a Grove planning contract first: `/grove-plan` today never
  writes `unattended` and asks the owner about dependencies
  (`grove/internal/instruct/requires/plan.md:22`). It also needs a durable,
  repeatable handoff in tend, and a plan rule that only opens PRs.
- **Agent-merged proposals.** These need a repair path for rejected
  proposals, which tend doesn't have.
- **An approver or playtester role** that judges principles and "fun", for
  example with Playwright against the preview.
- **Credential separation** (§5, known ceiling).

## 8. Not building

- Merging in tend, a merger agent, or bot approvals counted as required
  reviews.
- Rule precedence or free-text linting in Grove.
- A workflow engine or role framework in tend.
- Stacked PRs under a runner.
- Multiple hosts through herdr or rex.
- Containers, a GitHub App, auto-rollback or dashboards, until a named failure
  asks for one.

## 9. To verify before building

- Whether rulesets, auto-merge and CODEOWNERS enforcement work for the
  repository's visibility on the owner's plan. Public on Free is the likely
  choice.
- Whether a ruleset with zero required approvals plus code-owner review still
  requires the owner on owned paths.
- Whether the "Restrict updates" pause stops PRs already queued for
  auto-merge, and whether a PR-only owner bypass merges proposal PRs that
  have no `tend/review` status.
- Whether Workers Builds deploys after an auto-merge done by `GITHUB_TOKEN`.
  Events from `GITHUB_TOKEN` start no Actions runs; Workers Builds should be
  unaffected.
- What Workers previews can reach. They use the version's own bindings, so
  e2e tests against a preview could write production data, and previews
  aren't generated for Workers with Durable Objects.
