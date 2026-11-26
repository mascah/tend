# Design review, thread 3: authorization

Part of [the Tend and Grove design review](README.md): read that first. It
says what is decided and which file holds which section. Section numbers are
those of the single file these were split from.

From the third session, 2026-11-26: the evidence for thread 3, what a
sandbox on GitHub showed, and what the owner decided.

## 18. Thread 3, authorization (2026-11-26)

18.1 to 18.3 and 18.5 are evidence. 18.4 is Claude's proposal as first
put, kept for the reasoning. 18.6 is what the owner decided and what the
sandbox changed. 18.7 is what remains.

### 18.1 Who can post `tend/review`

Every agent on the runner, in one API call. Read live on 2026-11-26:

- `mascah/tend` has one ruleset, `main` (id 28050972). Its only rule is
  required status checks: `check` and `tend/review`, branch up to date,
  neither with a source pinned (no `integration_id`). No pull request or
  review rule. Bypass: the repository admin role, always.
- `tend/review` is a commit status. Tend posts it with the bot's token
  (`landing.go:79`, `publication.go:257`, `intake.go:256`), which it reads
  with `gh auth token --user <bot>` (`publication.go:28-34`). Any process of
  that OS user can run the same command; both agents run as that user.
- The bot's token is a classic OAuth token with `repo`, `workflow`, `gist`
  and `read:org`. `mascah-bot` has the write role.
- On the MacBook the same keyring holds `mascah` (admin, `repo` and
  `workflow`) as the active account. Every `gh` call of this session ran as
  the owner. An agent there can approve as the owner, merge by bypass, or
  edit the ruleset.
- The rule is bypassed in practice. All 9 merged pull requests the bot
  opened were merged by `mascah`; 7 of the merged heads carry no
  `tend/review` at all (#39, #40, #42, #44, #53, #56, #59). The two that do
  (#48, #57) show `success`, creator `mascah-bot`.
- `check` on those heads is a check run from GitHub Actions (app 15368).

So `tend/review` records what Tend said and proves nothing about who said
it. Rebuilding a record from it (14.2) rebuilds a claim, not evidence.

### 18.2 What else the bot's token allows today

- **Change CI in its own pull request.** With `workflow` scope it can push
  an edit to `.github/workflows/ci.yml`, and a `pull_request` run uses the
  pull request's copy of the workflow.
- **Approve through Actions.** The repository allows Actions to approve
  pull requests (`can_approve_pull_request_reviews: true`). A workflow
  pushed on any branch can ask for `pull-requests: write` and approve any
  open pull request as `github-actions[bot]`. Harmless today only because
  no rule asks for an approval.
- **Merge.** Write access merges any pull request whose required checks
  pass, and it can post one of the two itself.
- `ci.yml` runs the pull request's code with a read-only token and no
  secrets. `release.yml` holds write permissions and runs on pushes to
  `main` only. No workflow runs pull request code with a write token.
- `mascah/keyborg`: private, `mascah` the only collaborator, no ruleset,
  auto-merge off.

### 18.3 GitHub facts

From a delegated read of docs.github.com and the GitHub changelog on
2026-11-26, each with the quoted sentence. "Inferred" means the docs imply
it and do not say it.

| Fact | Status |
| --- | --- |
| Anyone with write access can set any commit status. A required check with no source pinned accepts it. With a source app pinned, a status from anyone else blocks the merge | Confirmed |
| "If a check and a commit status have the same name, both must pass": a posted `check` status cannot override a failed Actions run | Confirmed |
| Required checks match the job name only, whatever the workflow. A skipped job counts as success. Same-named jobs give "ambiguous" results | Confirmed; what a same-named job in an edited workflow does is not documented |
| On a personal account with Pro: rulesets with required approvals, code-owner review, approval of the most recent push, and required checks. Not available: required workflows (organization), push rules that restrict file paths (Team plan), merge queue (organization) | Confirmed; the four review settings inferred from the absence of a limit |
| Bypass actors on a personal repository: the admin role, a user, a GitHub App, a deploy key. Corrects 15.2 and 16, which said bypass lists need an organization; the live ruleset has one | Confirmed. That an empty list binds the owner is inferred |
| An author cannot approve their own pull request. Approval of the most recent push must come from someone other than the pusher, and earlier approvals are kept | Confirmed |
| An approval by `github-actions[bot]` counts when the repository setting allows it; new personal repositories default to off | Inferred from the setting's stated purpose. Whether another App's approval counts: not found |
| CODEOWNERS is read from the base branch; available on private repositories with Pro; owners are users with write access | Confirmed for review requests; inferred for the merge requirement. Deletions and renames under an owned path: not found |
| A fine-grained token cannot serve a collaborator on another account's repository. Classic `repo` includes commit statuses; no classic scope pushes to a private repository without them | Confirmed |
| Without `workflow` scope a token cannot add or update a workflow file, unless an identical file is on another branch of the repository | Confirmed as a scope rule. That a `git push` is refused, and what a deletion does: not found |
| `pull_request` and `pull_request_review` run the workflow file of the pull request's merge commit. `pull_request_target` and `issue_comment` run the default branch's. A pull request from a branch of the same repository gets secrets, and a write token if its workflow asks | Confirmed; the last sentence inferred (limits are stated for forks only) |
| From 2026-12-02 public repositories block `pull_request_target` unless an Actions event policy allows it. Private repositories are unaffected | Confirmed |
| An environment limited to `main` is closed to `pull_request` jobs and open to `pull_request_target` jobs. Environment secrets and branch limits work on private repositories with Pro; required reviewers do not | Confirmed |
| Auto-merge: available with Pro; a collaborator with write access can enable it; GitHub merges once required reviews and checks pass | Confirmed |
| A personal account can own a GitHub App with separate permissions for contents, pull requests, commit statuses, checks and workflows | Confirmed |
| Whether an approval survives a merge of the base branch into the head, under "dismiss stale approvals" and under "approval of the most recent push" | The docs contradict themselves. Settled by the sandbox (18.5): it survives. `guide/setup.md:173` says the opposite and is wrong |

### 18.4 Proposal as first put (Claude's; see 18.6 for what was decided)

**Frame.** A gate authorizes only when the credential that satisfies it
cannot be read by the agents' OS user. A credential can live in three
places: the runner, a workflow on the default branch, or the owner's own
devices. That gives each kind of gate one home:

| Gate | Home | How GitHub enforces it |
| --- | --- | --- |
| Mechanical (CI; scope unchanged; clean merge) | A workflow the bot cannot edit | Required check pinned to GitHub Actions; an approval from the workflow |
| Human (new scope, protected paths, releases) | The owner's account | Required approval; code owners |
| Judged, when it authorizes | A workflow the bot cannot edit | An approval from the workflow |
| Judged, when it only improves the work | The runner | Nothing; it is advice |

1. **Tend holds no authorization.** What must be true before a merge is a
   GitHub rule. What Tend checks before it acts is for convergence only.
2. **The runner holds one credential: the bot's, without `workflow`
   scope.** This is the linchpin on a personal account, which has neither
   required workflows nor path rules: it is what stops a pull request
   rewriting CI or adding a workflow that approves. Tend refuses to run
   when the bot's token carries `workflow`, and `tend watch` refuses when
   `gh` holds a token with admin on the repository. A machine where the
   owner is signed in beside the agents is supervised by definition
   (thread 6). No second OS user is needed for authorization.
3. **The authorizing act is an approving review, required by the ruleset,
   from someone who did not push the head.** Ruleset: pull request
   required; 1 approval; approval of the most recent push; code-owner
   review; `check` pinned to GitHub Actions; branch up to date; the owner
   the only bypass. `tend/review` stops being required. Two approvers
   exist that an agent cannot impersonate: the owner, and a workflow on
   the default branch.
4. **Protected paths become CODEOWNERS.** Decision 3a's list moves from
   the flow file's `protect` to a file GitHub reads from the base branch.
5. **Stage 1, supervised and enforced.** The owner is the only approver.
   Configuration, plus deletions in Tend.
6. **Stage 2, the hosted gate, is what makes a merge autonomous.** One
   workflow on the default branch that runs nothing from the pull request.
   It approves an implementation when: the branch is a claimed attempt;
   `grove review` says it approves no new change and the proposal and
   specs equal those at the claim commit (finding 3); and a model, given
   the approved proposal and the diff and no tools, answers that the diff
   stays inside the proposal and crosses none of decision 3b and 3c. A
   head that is a clean merge of the target into an approved head is
   approved with no model (decision 17). Anything else waits for the
   owner. The model's credential is the `claude setup-token` token (15.2)
   in an environment limited to `main`, read by a `pull_request_target`
   job. This is the part of Claude Approvals the owner needs, as one file.
7. **The runner's reviewer stays, as advice.** It runs the code, finds
   defects and drives the repair. Its verdict authorizes nothing, so
   finding 4 stops mattering and nothing rebuilds authority from it.
   Alternative not recommended: move the whole review to the hosted gate.
   One reviewer and no verdict files, but it cannot run the code, each
   repair waits for CI first, and Codex there needs an API key.
8. **`land` shrinks to two acts:** merge the target into the front
   candidate when it is behind, and enable auto-merge. GitHub waits for
   the checks and merges. The lock held across CI polling (finding 9), the
   merge call and the parsing of GitHub's refusal go. Decision 16 holds:
   Tend still orders the train, one candidate at a time.
9. **No GitHub App now.** The workflow is already an identity agents
   cannot reach, and an App's key would have to sit in the same Actions
   secret. An App in place of the bot user would make "no workflow
   permission" structural where a token scope is a choice at login; worth
   it when a second runner or an organization appears.

Goes from Tend: `hold`, `approved`, `approvedOn` and `writer`
(finding 5), the `protect` list and its parsing (`flow.go:77-185`),
`vouch` and `notDocs` (`intake.go:207-284`), the `tend/review` posts,
`carry`, and most of `landStep`. About 570 non-test lines carry
authorization or landing today. `cleanMerges` and `notDocs` move into the
hosted gate. Needs from Grove: `grove review --json`, which today prints
text only.

Open in the proposal: whether Stage 2 is built before keyborg runs
unattended or after Stage 1 has run there; and which model the hosted gate
uses.

### 18.5 Sandbox results

Run on 2026-11-26 in `mascah/tend-auth-sandbox`, a private repository made
for this, with `mascah-bot` as a collaborator with write access. The
owner allowed it. It is disposable and still exists: the token used cannot
delete a repository.

Setup. A CI workflow on `pull_request` with one job, `check`, skipped for
paths under `norun/`. A workflow on `workflow_dispatch` that approves the
pull request whose number it is given, with `pull-requests: write`.
`.github/CODEOWNERS`: `/owned/ @mascah`, then `/owned/archive/` with no
owner. Actions allowed to approve pull requests. One ruleset on `main`
with an empty bypass list, its rules changed per test: `check` required
with or without `integration_id` 15368 (GitHub Actions), and a
`pull_request` rule of 1 approval with `dismiss_stale_reviews_on_push`,
`require_last_push_approval` and `require_code_owner_review` on or off.
The bot opened every pull request and attempted every merge unless noted,
with `PUT /pulls/<n>/merge`. Its token had `workflow` scope for tests 1 to
16 and lacked it for 17 to 25 (`X-Oauth-Scopes: gist, read:org, repo`).

| # | Test | Observed |
| --- | --- | --- |
| 1 | `check` pinned to Actions; head has no check run; the bot posts a `check` status of `success` and merges | Refused: "Required status check "check" was not set by the expected GitHub app." |
| 2 | Same head, the owner merges, not on the bypass list | Refused, same message. An empty bypass list binds the owner |
| 3 | Same head, `check` required from any source | Merged on the bot's own status. This is `tend/review` on `mascah/tend` today |
| 4 | Approval rules on; the bot approves its own pull request | "Can not approve your own pull request" |
| 5 | The bot merges with no approval | Refused: "New changes require approval from someone other than the last pusher." |
| 6 | The bot dispatches the approving workflow on its own pull request, then merges | Approved by `github-actions`, decision `APPROVED`, merged. A workflow's approval counts toward the required approval and the most-recent-push rule, and a workflow anyone with write access can dispatch is self-approval |
| 7 | Workflow-approved pull request that edits a file under `/owned/` | Refused: "Waiting on code owner review from mascah." Merged once the owner approved |
| 8 | Workflow-approved, deletes a file under `/owned/` | Refused, waiting on the code owner |
| 9 | Workflow-approved, moves a file from `/owned/` to `/owned/archive/` | Refused, waiting on the code owner |
| 10 | Workflow-approved, adds a file under `/owned/archive/` | Merged. A later line with no owner un-owns a path |
| 11 | Workflow-approved, edits an owned file and removes its line from CODEOWNERS | Refused, waiting on the code owner. The base branch's copy decides |
| 12 | Owner approves; `main` moves; the bot pushes a merge of `main` into the head (a merge commit, as Tend makes). Run three times: most-recent-push rule alone, dismiss-stale alone, both | Approval kept each time, and shown on the new head; merged |
| 13 | Both rules on; owner approves; the bot pushes a commit that changes a file | Approval `DISMISSED`; refused |
| 14 | Both rules on; owner approves; the bot pushes a merge of `main` whose merge commit also changes a file | Approval `DISMISSED`; refused |
| 15 | The bot opens a pull request that adds a workflow with `pull-requests: write`, which approves a different open pull request of the bot's | It ran at once. The other pull request showed `APPROVED` by `github-actions` and the bot merged it. No person took part |
| 16 | Branch must be up to date, both review rules, auto-merge allowed. The bot enables auto-merge on two pull requests; the owner approves both | GitHub merged the first. The second waited, behind. The bot pushed a merge of `main` into it; the approval held; CI passed; GitHub merged it. `mergedBy` is the bot, who enabled auto-merge |
| 17 | Without `workflow` scope: the bot pushes a branch that edits `ci.yml` | Refused: "refusing to allow an OAuth App to create or update workflow `.github/workflows/ci.yml` without `workflow` scope" |
| 18 | Pushes a branch that adds a workflow file found on no other branch | Refused, same message |
| 19 | Pushes a branch that renames a workflow file | Refused, for the new path |
| 20 | Adds a workflow file through the contents API | 404 |
| 21 | Pushes a merge of `main` after the owner changed `ci.yml` there; the merge commit carries that change | Accepted. A further edit of `ci.yml` on that branch: refused |
| 22 | Pushes a branch that adds a workflow file identical, in path and contents, to one on another branch | Accepted. The documented exception |
| 23 | Pushes a branch that puts `ci.yml` back to an older version still held by another branch | Accepted, by the same exception |
| 24 | Pushes a branch that deletes `ci.yml` | Accepted. Such a head has no `check`, and test 1 shows the merge is then refused as "expected" |
| 25 | Pushes a branch that edits `.github/CODEOWNERS`, and one that adds `.github/actions/x/action.yml` | Accepted. The scope covers `.github/workflows/` only |

What the results change:

- **GitHub carries an approval across a clean merge of the target and
  drops it for anything else** (12 to 14). That is decision 17, done by
  the platform and checked against the diff. `cleanMerges`, `carried`,
  `carry` and the `carried-*` marker have no job left, and the hosted gate
  needs no clean-merge approver.
- **The train needs two acts from Tend** (16): enable auto-merge once, and
  merge the target into the candidate that is behind.
- **`grove/changes/` cannot be code-owned** (8, 9): an implementation
  moves its change to the archive, and that would wait for the owner every
  time. "New scope needs the owner" has to come from who approves, not
  from a path.
- **A workflow's approval is real authority** (6, 15). While the bot's
  token has `workflow` scope and Actions may approve, the bot authorizes
  itself. `mascah/tend` has both today; nothing asks for an approval there
  yet. The hosted gate of Stage 2 must take its pull request from the
  event, never from an input.

- **Without `workflow` scope the bot cannot bring new workflow content
  into the repository** (17 to 21), and Tend's merge of the target still
  goes through when the target's CI changed (21). Three limits. It can
  delete a workflow, which fails closed for a required check (24). It can
  put a workflow back to any version another branch still holds (23): a
  downgrade to something the owner once approved, visible in the pull
  request's diff, and held for the owner in Stage 2 once `.github/` is
  code-owned (tests 7 to 9). And files a workflow calls from outside
  `.github/workflows/`, such as a local action or a script, are not
  covered (25), so the hosted gate must call none.

### 18.6 Decided (owner, 2026-11-26)

1. **Yes to items 1 to 3 of 18.4.** Tend holds no authorization. The
   runner holds one credential, the bot's, without `workflow` scope. The
   authorizing act is a required approving review from someone who did not
   push the head, with `check` pinned to GitHub Actions and `tend/review`
   no longer required.
2. **Stage 2 comes after Stage 1 has run on keyborg.** The owner added
   that keyborg waits on more than this: how agents do product research
   that makes an autonomous proposal meaningful (section 17, item 9).
3. The sandbox was allowed.
4. **`land` is two acts** (item 8 of 18.4): enable auto-merge, and merge
   the target into the candidate at the front when it is behind.
5. **No GitHub App for now** (item 9 of 18.4).
6. **The owner stays on the bypass list**, as on `mascah/tend` today.

Stage 1 as it now stands, with the sandbox results applied:

- **Ruleset on the target:** pull request required; 1 approval; dismiss
  stale approvals; approval of the most recent push; `check` required from
  GitHub Actions (`integration_id` 15368); branch up to date. The owner
  bypasses, as on `mascah/tend` today. The bot does not.
- **Repository:** auto-merge allowed; squash only. Actions not allowed to
  approve pull requests where nothing needs it: off by default on a new
  personal repository such as keyborg. `mascah/tend` has it on, which
  release-please needs to open its pull request.
- **Runner:** the bot's token without `workflow`; the owner's token not in
  that OS user's keyring.
- **The owner's act is one approval per pull request.** GitHub merges.
- **CODEOWNERS waits for Stage 2.** While the owner is the only approver
  it adds nothing.
- **Tend:** `publish` enables auto-merge; `land` merges the target into
  the candidate at the front when it is behind; the approval is read as
  GitHub's review decision. `hold`, `approved`, `approvedOn`, `writer`,
  the `protect` list, `vouch`, `notDocs`, `passed` as a landing guard,
  `cleanMerges`, `carried`, `carry`, the `tend/review` posts and the rest
  of `landStep` go. This belongs to the rewrite of 14.2, not to a patch.

### 18.7 What remains of thread 3

- **The bot's token on the mini.** Only the MacBook's was changed. Each
  machine's keyring holds its own token for the bot; the runner's needs
  the same: `gh auth switch -u mascah-bot`, `gh auth refresh -h github.com
  -r workflow` with the browser signed in as the bot, `gh auth switch -u
  mascah` where the owner is signed in too. The owner's token keeps
  `workflow`.

  On the MacBook the first refresh left the scope in place and the second
  removed it; why the first failed was not established. `gh auth status`
  prints one scopes line for each account, which is easy to misread, and
  for a while the scope was believed gone when it was not. So the check
  in item 2 of 18.4 (Tend reads `X-Oauth-Scopes` and refuses to run) is
  needed, not optional.
- **Until Tend handles it:** a change that must edit CI cannot be pushed
  by the bot. Tend stops at "could not push"; the owner pushes that
  branch.
- **Stale branches hold old workflow versions** (test 23). Deleting merged
  branches keeps the set small.
- **Deferred with Stage 2:** items 4, 6 and 7 of 18.4 (code owners for
  protected paths, the hosted gate, the runner's reviewer as advice). The
  sandbox supports item 4.
- **Stage 2's shape** (items 6 and 7 of 18.4) is deferred with Stage 2.
  Carry forward: no clean-merge approver; new scope is held by the gate
  withholding its approval; the gate is never dispatchable.
- **Correct `guide/setup.md:173`** when the guide is next touched.
- **Delete `mascah/tend-auth-sandbox`** once the scope tests are run.
