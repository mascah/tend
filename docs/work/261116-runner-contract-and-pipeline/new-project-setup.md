# Setting up the typing app's pipeline

How to stand up the touch-typing practice app with the autonomous pipeline
from design §7. Steps run on the MacBook unless they say "on the mini".
`APP` is the repository's name; pick one and set it once:

```sh
APP=<name>
```

The pipeline in one line: tend implements and Codex reviews on the mini; GitHub
auto-merges the bot's implementation PRs once `ci` and `tend/review` pass; and
Workers Builds deploys. You merge proposals and plan cycles.

## 0. Before you start

- Grove 0.12.0 or later and tend 0.2.0 or later are on the mini (`grove
  version`, `tend version` as `tend`). Tend needs `--runner`.
- Decide visibility. Rulesets and required code-owner review work on public
  repositories on GitHub Free; a private repository needs Pro or Team for
  both. The guide assumes public.

## 1. The app

```sh
cd ~/GitHub/mascah
npm create cloudflare@latest -- "$APP" --framework=react-router
cd "$APP"
```

Choose TypeScript, git, and no deploy at the end. Then:

- Switch to pnpm (`rm package-lock.json && pnpm install`) and add
  `"packageManager"` to `package.json`, or keep npm and change the commands
  below.
- Add a test runner (`pnpm add -D vitest`) and a `"test": "vitest run"`
  script, with one passing test, so `ci` has something to run from day one.
- Keep v1 free of bindings (no D1, KV or Durable Objects). Previews use the
  version's own bindings, so they would write production data, and Workers
  with Durable Objects get no previews. Practice stats live in
  `localStorage` until a trial shows they need a server.
- Mark previews `noindex`. In the Worker's entry (`workers/app.ts`), on
  every response whose host isn't the production host:

  ```ts
  const res = await requestHandler(request, { cloudflare: { env, ctx } });
  if (new URL(request.url).hostname !== env.PRODUCTION_HOST) {
    res.headers.set("X-Robots-Tag", "noindex");
  }
  return res;
  ```

  Then set `"vars": { "PRODUCTION_HOST": "<app>.<subdomain>.workers.dev" }`
  in `wrangler.jsonc` and run `pnpm cf-typegen`. If the handler's response
  headers are immutable, build a new `Response` with copied headers.

## 2. Principles and agent instructions

`PRINCIPLES.md` holds what "fun" means, and what each role may do:

```markdown
# Principles

## Product
- Typing starts within a second of the page loading.
- No sign-up, ever, for practice.
- Keyboard only: every action has a key; nothing needs the mouse.
- Sessions are short: a round is under a minute.
- Feedback is immediate: every keystroke shows right or wrong at once.
- No layout shift while typing.
- When a change trades one of these away, the proposal says so.

## Roles
### Implementer (tend's Claude)
Optimizes: the agreed change, small and tested. Never: changes CI, policy,
deployment or this file; adds a dependency the change doesn't name. Stops:
when the change can't be built as agreed; it blocks the task with a question.

### Reviewer (tend's Codex)
Optimizes: the head does what the change says and keeps these principles.
Never: requests changes in GitHub; passes a head it didn't run. Stops: fails
the review, saying whether the failure is repairable.

### Proposer (tend propose)
Optimizes: a proposal a person can approve in one read. Never: decides an
open question itself. Stops: declines a request that names no outcome.
```

`AGENTS.md`, which Codex reads on its own:

```markdown
Read PRINCIPLES.md before any work and judge your work against it. Commands:
`pnpm typecheck`, `pnpm test`, `pnpm build`. Run all three before saying a
task is done.
```

`CLAUDE.md` holds one line, `@AGENTS.md`, so Claude reads the same.

## 3. Grove

```sh
grove init          # claude and codex, as skills
```

In `grove.yaml`:

```yaml
context: "React Router v7 on Cloudflare Workers, TypeScript, pnpm. PRINCIPLES.md says what the app must be; read it first. Before a task is done: pnpm typecheck && pnpm test && pnpm build."
rules:
  skills:
    implement:
      - Commit after each task, naming its task ID.
  delivery:
    propose:
      - Commit the change on a branch other than main, push it to origin and open a pull request with gh pr create.
    plan:
      - Commit the cycle on a plan/<cycle-id> branch, push it and open a pull request with gh pr create.
    implement:
      - When you are done, push the branch to origin and open a pull request with gh pr create, listing first every Check line and each blocked task with its question.
```

Under tend, `--runner` leaves out the delivery lines and tend delivers. In
your own sessions they end with a pull request, so nothing is left for you
to finish.

## 4. CI

`.github/workflows/ci.yml`, whose one job is named `ci` (the ruleset
requires that name):

```yaml
name: ci
on:
  pull_request:
  push:
    branches: [main]
permissions:
  contents: read
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: false
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm typecheck
      - run: pnpm test
      - run: pnpm build
```

Pin the actions to commit SHAs if you run zizmor, as grove and tend do.

## 5. Publish the repository

```sh
git add -A && git commit -m "chore: start the app"
gh repo create "mascah/$APP" --public --source . --push
gh repo edit "mascah/$APP" --enable-auto-merge --delete-branch-on-merge \
  --enable-squash-merge --enable-merge-commit=false --enable-rebase-merge=false
gh api -X PATCH "repos/mascah/$APP" \
  -f squash_merge_commit_title=PR_TITLE -f squash_merge_commit_message=PR_BODY
gh api -X PUT "repos/mascah/$APP/collaborators/mascah-bot" -f permission=push
```

Tend writes each PR's title and description as the squash commit, so the
squash settings make the merged commit match. On the mini, accept the
invitation as the bot:

```sh
gh api user/repository_invitations --jq '.[] | "\(.id) \(.repository.full_name)"'
gh api -X PATCH user/repository_invitations/<id>
```

## 6. Policy

`.github/CODEOWNERS`, so a PR touching agent policy, CI, deployment or the
plan waits for you:

```
/.github/        @mascah
/PRINCIPLES.md   @mascah
/AGENTS.md       @mascah
/CLAUDE.md       @mascah
/grove.yaml      @mascah
/grove/cycles/   @mascah
/.claude/        @mascah
/.agents/        @mascah
/wrangler.*      @mascah
```

The memo's list, plus `grove/cycles/`, so no agent can mark its own work
`unattended`, and the skill folders, so no agent can rewrite Grove's skills.

`.github/workflows/auto-merge.yml` turns on auto-merge for the bot's
implementation PRs only. It never checks out PR code:

```yaml
name: auto-merge
on:
  pull_request_target:
    types: [opened, reopened, ready_for_review]
permissions: {}
jobs:
  enable:
    if: >-
      github.event.pull_request.user.login == 'mascah-bot' &&
      startsWith(github.event.pull_request.head.ref, 'grove/') &&
      github.event.pull_request.head.repo.full_name == github.repository
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
    steps:
      - run: gh pr merge --auto --squash "$PR" --repo "$REPO"
        env:
          GH_TOKEN: ${{ github.token }}
          PR: ${{ github.event.pull_request.number }}
          REPO: ${{ github.repository }}
```

Proposals (`tend/issue-*`) and plans (`plan/*`) are left for you.

The ruleset on `main`. Save it as `main.json` and create it with `gh api
repos/mascah/$APP/rulesets --input main.json`:

```json
{
  "name": "main",
  "target": "branch",
  "enforcement": "active",
  "conditions": { "ref_name": { "include": ["~DEFAULT_BRANCH"], "exclude": [] } },
  "bypass_actors": [
    { "actor_id": 5, "actor_type": "RepositoryRole", "bypass_mode": "pull_request" }
  ],
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    { "type": "pull_request", "parameters": {
        "required_approving_review_count": 0,
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": true,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false,
        "allowed_merge_methods": ["squash"] } },
    { "type": "required_status_checks", "parameters": {
        "strict_required_status_checks_policy": false,
        "required_status_checks": [ { "context": "ci" }, { "context": "tend/review" } ] } }
  ]
}
```

Role 5 is admin, which is you; the bot has write, so it can't bypass.
"Pull requests only" lets you merge proposals and plans that carry no
`tend/review`, and never lets anyone push to `main` directly.

The pause ruleset is created disabled. Save it as `pause.json`:

```json
{
  "name": "pause",
  "target": "branch",
  "enforcement": "disabled",
  "conditions": { "ref_name": { "include": ["~DEFAULT_BRANCH"], "exclude": [] } },
  "bypass_actors": [
    { "actor_id": 5, "actor_type": "RepositoryRole", "bypass_mode": "always" }
  ],
  "rules": [ { "type": "update" } ]
}
```

```sh
gh api "repos/mascah/$APP/rulesets" --input pause.json --jq .id   # note the id
# pause:   gh api -X PUT repos/mascah/$APP/rulesets/<id> -f enforcement=active
# resume:  gh api -X PUT repos/mascah/$APP/rulesets/<id> -f enforcement=disabled
```

Commit CODEOWNERS and the workflow through a PR you merge with the bypass,
as you will every policy change from now on.

## 7. Workers Builds

In the Cloudflare dashboard, go to Workers & Pages, then Create, then Import
a repository, and pick `mascah/$APP`:

- Build command `pnpm run build`; deploy command `npx wrangler deploy`.
- Builds for non-production branches on, with `npx wrangler versions
  upload`, so each PR gets a preview URL that Cloudflare comments on the PR.
- Production branch `main`.

Workers Builds deploys from its GitHub App, so a merge made by auto-merge
still deploys. Production rolls back with `npx wrangler rollback`.

## 8. Tend on the mini

As `tend` on the mini (`ssh mini-tend`):

```sh
git clone "https://github.com/mascah/$APP.git" ~/repos/$APP
echo "https://github.com/mascah/$APP.git mascah-bot" >> ~/.local/state/tend/publish
```

Add `~/repos/$APP` to the `exec tend watch ...` line in `~/.local/bin/tend-watch`.
Then restart the watch: `pkill -f "tend watch"`. launchd's KeepAlive starts it
again within 30 seconds.

## 9. Check the policy before trusting it

These are design §9's open points. Each takes a minute with a throwaway PR:

1. **The bot can't merge or push.** As the bot, `git push origin HEAD:main`
   is refused, and so is `gh pr merge` on a PR without both checks.
2. **Code owners hold with zero approvals.** A bot PR touching
   `PRINCIPLES.md` with both checks green stays blocked until you approve.
3. **The bypass merges a proposal.** You can merge a `tend/issue-*` PR that
   has `ci` but no `tend/review`.
4. **Auto-merge fires.** A bot `grove/*` PR merges by itself once `ci` and a
   `tend/review` success status are on its head, and production deploys.
5. **The pause stops a queued merge.** Activate `pause` while a PR waits on
   auto-merge, add the last check, and it doesn't merge; disable it and it
   does.
6. **CI runs on the bot's PRs.** If a bot PR's run shows "approval
   required", allow it under Settings, Actions, "Fork pull request workflows
   from outside collaborators". The bot is a collaborator, so this is
   unlikely.

## 10. The trial

- **Start:** label an issue `tend`. Tend drafts a proposal PR; read it,
  answer its open questions on the PR, and merge it.
- **Plan:** run `/grove-plan` for one change, mark the cycle `unattended`,
  and merge the plan. Keep cycles to one change, chained by `depends_on`
  when order matters, so each PR is one change.
- **Watch:** tend claims it on its next pass. It implements, reviews and
  publishes; auto-merge merges, and Workers Builds deploys.
- **Count:** for about ten cycles, note in each merged PR whether you would
  have refused it, and why. One refusal is a finding: pause, read it, then
  fix the cause in principles, Grove rules or the review, not in the PR.

### Ceilings to remember

- The agents run as the mini's `tend` user, which holds the bot's token. An
  agent determined to could post its own `tend/review` and merge. The guard
  and the layout check stop mistakes; credential separation (design §5) is
  the fix if the trial ever shows a need.
- No daily attempt cap. You plan and mark every cycle, which already bounds
  the work; add one when a planner role exists.

Later steps, such as a planner, agent-merged proposals or a playtester, wait
for the trial to ask for them (design §7, "Later").
