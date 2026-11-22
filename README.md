# tend

A small loop that takes agreed work through coding agents and brings back a
reviewed branch. Today it does one pass for one [Grove](https://github.com/mascah/grove)
cycle, or one approved change no cycle lists:

```text
grove claim
    -> Claude Code implements, in a Herdr worktree pane
    -> Codex reviews the head commit
    -> one repair in the implementer's conversation, if the review asks
    -> Codex reviews the new head
    -> a branch and a verdict, waiting for you
    -> with a listed origin, a pull request and its reviews, as a bot account
```

Grove holds what is agreed. Herdr runs the agents where you can watch and
take over. tend is a small Go program between them. It reconstructs
progress from Git, its own run files and Herdr on every run.

The agreed [execution plan](docs/work/261114-reliable-orchestration/spec.md)
records the intended boundaries, review findings, delivery order and session
handoff. Start there when continuing development across sessions.

## Run it

Install a release into `~/.local/bin` (`TEND_INSTALL_DIR` to change it), or a
given version, and run it again to update; `tend version` says which you have:

```sh
curl -fsSL https://raw.githubusercontent.com/mascah/tend/main/install.sh | sh
curl -fsSL https://raw.githubusercontent.com/mascah/tend/main/install.sh | sh -s -- v0.1.0
```

Or build it with Go 1.26 (`go build -o bin/tend .`). It needs `grove` (0.14.0
or later), `git` and `gh` on PATH, Herdr or rex, and Claude Code and Codex
signed in. Run it inside a Herdr pane or a rex terminal:

```sh
bin/tend <repo> <change-id>
bin/tend <repo> <cycle-id>
bin/tend <repo> issue-<n>
bin/tend status <repo> [<change-id> | <cycle-id>]
bin/tend watch [<repo>...]
bin/tend propose <repo>
```

An approved change is built with `tend <repo> <change-id>`, as a cycle of
that one change: tend claims it by name (`grove claim <change-id>`), and
carries on, reports and publishes it as it does a cycle. A cycle is needed
only to build several changes together. A change a cycle lists is built with
its cycle, and a change that waits on another is refused; either way tend
stops with Grove's sentence.

The agents run in Herdr, or in rex when tend runs in rex (not in Herdr);
`TEND_TERMINAL=herdr` or `rex` chooses. In rex each worktree gets a session,
first named after its folder and found again by its ID in
`TEND_STATE/rex-sessions` (agents may rename it), and each agent a block named
after it. Its state is
what the agent reports with rex's program status protocol, as Claude Code and
Codex do, so a dialog the agent does not report shows as unknown and stops
tend rather than being prompted into. rex worktrees are made by Git under
`TEND_STATE/worktrees/`, and tend reports its own state to its tab.

- Exit 0: the head is reviewed, and the verdict is printed.
- Exit 2: no work item was given; a usage line says how to give one.
- Exit 3: a person is needed. A Herdr notification says why, or under rex
  tend's own tab.
- Exit 4: another tend is running with the same `TEND_STATE`, and nothing was
  done. One tend runs at a time per state folder; the lock goes when it exits
  or is killed.
- Exit 1: a command tend relies on failed unexpectedly.

Its own files live under `TEND_STATE` (default `$XDG_STATE_HOME/tend`, or
`~/.local/state/tend` without it): the
lock, and a folder per attempt under `runs/` holding its verdicts and repair
mark. They never enter the candidate, so a repository needs no Git excludes.
An attempt from before this, with a `.tend/` folder in its checkout, has
those files moved to its folder on the next run, and a name clash stops for
you rather than overwriting either.

After exit 3, or a kill, the same line carries on, with or without `origin`:

- An agent still working, as after tend was killed mid-turn, is waited for
  and never prompted again. A blocked or unknown agent stops for you.
- A timed-out or stalled prompt is not assumed undelivered, so tend never
  sends it twice.
- The one repair is used once asked: `repair-asked` records the reviewed head
  just before the prompt goes. If that head has no commit since and the
  implementer is not working, the outcome is unknown, and tend stops for you
  on every run until you commit a fix or delete `repair-asked` to have it asked
  once more. A repair refused before sending, such as by a blocked
  implementer, is not used.

`tend status` says which attempt the change or cycle has, whether a tend is
on it now, what it last recorded and the next action, including "run it
again" when tend was killed. It reads only, so it works while tend runs.

With `origin`, tend publishes only if a line of `TEND_STATE/publish` names that
URL and the GitHub account to publish as, `<origin-url> <account>`; otherwise
it says so and keeps the verdict local. The account is a separate one, such
as a bot user with write access, signed in to `gh` beside yours (`gh auth
login`, then `gh auth switch` back); tend takes its token for its own calls
only. Every write is that account's, which also lets you approve its PRs:

- the branch is pushed over HTTPS, never forced; a branch someone else pushed
  to stops for you;
- one pull request onto the target, ready to squash-merge: its title and
  description are the commit message the implementer writes for the reviewed
  head (what the change does and why, for a later reader of the history), set
  when the PR opens and written again for a later head, until a person edits
  them;
- each verdict of the attempt is a comment review on its own commit, the
  head's with `grove review`'s summary folded in, and each reviewed commit
  gets the `tend/review` status, so a later commit has no pass until it is
  reviewed.

The PR and the reviews are looked up on GitHub each run, never remembered, so
a rerun after any interruption repeats neither. tend never merges.

When a person requests changes (GitHub's "Request changes") on the PR's head,
the next run gives their review and line comments to the implementer's own
session once, has the new head reviewed and publishes it, with the
implementer's reply in the review. `request-<review id>` in the run folder
records the ask just before it goes; asked with no commit since and nobody
working, it stops for you, as the repair does. When the implementer changes
nothing (a request it judges out of scope, say), its reply goes on the PR as a
comment and the request is not asked again; request changes again to have it
tried. After the new head, the implementer rewrites the commit message and
tend puts it on the PR, unless a person has edited the title or description
since tend set them. Plain comments are not acted on, nor is a request from
anyone without write access: tend acts only on a review whose author GitHub
gives as owner, member or collaborator (in an organization's repository, a
member with read access only still counts), and says whose request it left.

`tend watch` runs in a pane of its own and does that without you: every
`WATCH_SECONDS` (default 300) it runs tend on each attempt whose PR is open
with a change request on its head from a person with write access, and `tend propose` on each repository it
was given. In each of those repositories it then starts the oldest cycle that
Grove would claim and whose plan marks it `unattended`: marking a cycle so when
you plan it is how you let it be built with no one there. An attempt or request
waiting for you is left alone until you have run tend on it.

`tend status <repo>`, without a work item, says what waits for whom: approved
changes in no cycle and cycles not done, each with Grove's verdict,
this runner's attempts for them, proposals and implementations to review
(and which have changes requested), and labelled requests not yet drafted.

`tend <repo> issue-<n>` builds a GitHub issue instead of a Grove cycle, in a
repository without Grove; where Grove holds the scope, it refuses and points
to the `tend` label, so a request becomes a proposal first. The `tend-build` label approves it (only
people with triage access can label); an open, labelled issue may start, and a
closed one is done. Its title and description are the scope. It is built on
`tend/build-<n>`, done when the branch has a commit and the implementer's notes
for the reviewer are in the run folder; the reviewer checks the diff against
the issue and those notes. Everything after is the same: one repair, the
lock, recovery, publication (the commit message ends `Closes #<n>`), requested
changes and watch.

`tend propose <repo>` turns requests into proposals, once per run. Each open
GitHub issue labelled `tend` in a listed repository, oldest first, is drafted
by Claude with Grove's `/grove-propose --unattended`, on branch
`tend/issue-<n>`. A draft is committed with a squash-ready message ending
`Refs #<n>` and opened as a pull request by the publishing account, whose
review lists the open questions, consequential ones first, and what merging
with them open means. A request no change can be drafted from gets a comment
saying why, and its label is taken off; putting it back tries again. A pull
request from the issue's branch, in any state, means the issue is handled;
at most `CAP` (default 5) proposal pull requests stay open. Labelling an issue
hands its text to an agent running in auto mode, so only people who may label
(triage access) can send requests. Approving the proposal (merging it) and
starting its build stay with people.

New work branches from the target as origin has it, fetched first. Once a pull
request is merged or closed, the next `tend propose` or run in that repository
removes the worktree tend made for it, with its local branch and the agents'
place, when its head is the pull request's and it holds no uncommitted files;
otherwise it is left and tend says why. A declined request's worktree goes at
once.

Jira is a second source of requests and work items; code, pull requests and
reviews stay on GitHub. A line of `TEND_STATE/jira` maps a project to the
repository it feeds, `<site> <project-key> <origin-url>` (the origin must also
be listed in `publish`), and tend authenticates with `JIRA_EMAIL` and
`JIRA_API_TOKEN` from the environment, through `curl` with the token on its
standard input. `tend propose` then also drafts the project's issues labelled
`tend` and not done, on `tend/jira-<KEY>`, and comments the pull request's URL
on the Jira issue; `tend <repo> <KEY>` builds one labelled `tend-build` in a
repository without Grove, as `issue-<n>` does, ending its commit message `Refs
<KEY>` and commenting the pull request's URL on it once published. Anyone who can create or edit issues in a mapped project can start an
agent this way, so map only projects whose members you trust with that. A
handoff stopped between the pull request and the Jira comment is finished by
the next `tend propose`, which completes every intake folder with a `pr` and no
`linked`. tend never moves Jira issues through their workflow.

To cancel, stop tend (Ctrl-C or a kill). The agents keep their work and
their turns. Stop an agent in its pane if you want it stopped too; running
tend again carries on from what Git and the run folder hold.
`TIMEOUT_MS` sets the prompt wait timeout (default 4 hours, a backstop for a
hung agent); a timeout does not stop the agent.

To try it on a disposable repo, with a checkout of Grove beside this one:

```sh
dev/fixture.sh ~/tmp/ledger
bin/tend ~/tmp/ledger 261114-pr0t0
```

`go test ./...` builds tend and runs its acceptance tests, in parallel:
`tend_test.go` drives the binary with real Git and Grove and stand-ins in
`testdata/bin` for Herdr, rex, `gh` and `curl`, whose agents act at once. It
starts no agent and needs no Grove checkout, but needs `grove` and `jq` on
`PATH`. Run one scenario by name with `go test -run 'TestTend/^publish$'`.

## How work reaches tend

With `tend watch <repo>...` running, each pass (every `WATCH_SECONDS`, default
300) picks up four kinds of work in the repositories it was given.

1. **A request becomes a proposal.** Label an open GitHub issue `tend`, or an
   issue of a mapped Jira project that is not done. tend drafts a Grove change
   and opens it as a pull request by the publishing account, on
   `tend/issue-<n>` or `tend/jira-<KEY>`, with a review listing the open
   questions, consequential ones first; a Jira issue also gets a comment with
   the pull request's URL. When no change can be drafted, tend comments why
   and takes the label off; label it again after editing it to have it tried
   again. At most `CAP` (default 5) proposals stay open per repository, and a
   request with a proposal pull request in any state is not drafted again.
2. **An approved change becomes a build.** Merging the proposal approves the
   change. Run `tend <repo> <change-id>`, and tend claims it, has Claude
   implement it and Codex review it, repairs once if the review asks, and
   opens the implementation pull request. A cycle is needed only to build
   several changes together, or for watch to start the build without you:
   plan one (`/grove-plan`) and mark it `unattended: true`. A cycle not
   marked `unattended` is built only when someone runs
   `tend <repo> <cycle-id>`.
3. **Requested changes.** On an implementation pull request, GitHub's
   "Request changes" (not a plain comment) by an owner, member or
   collaborator goes to the implementer; anyone else's is left. Either it
   commits, the new head is reviewed, and the description is written again
   unless a person has edited it; or it makes no change and its reply goes on
   the pull request, and that review is not asked again. Request changes again
   to have it tried.
4. **Cleanup.** Once a pull request is merged or closed, a later pass removes
   its worktree, branch and terminal workspace.

Not picked up: plain comments, on any pull request, and comments on proposals
(revising a proposal is a person's); issues without the label; approved
changes in no cycle, and cycles not marked `unattended`; a proposal closed
without merging, which is never drafted again. In a repository without Grove, an issue labelled `tend-build` is built
only by `tend <repo> issue-<n>` (or `<KEY>`).

To see where things stand: `tend status <repo>` says what waits for whom;
tend watch's own output says what it ran; the agents work in the terminal's
workspaces, one per worktree; and when tend needs a person it stops and
notifies.

## Where the prototype reads progress

| Question | Read from |
| --- | --- |
| Is there an attempt? | The newest record for the cycle, or naming the change alone and no cycle, in `grove changes --json` (from `origin/grove/runs`, or the local `grove/runs` without `origin`), whose `checkout` still holds its `branch`. |
| Is it built? | Every change of the record is under `changes/archive/` on that branch, and the tree is clean. |
| Is the head reviewed? | `runs/<attempt>/verdict-<head>.json` names that head and has a `success` or `failure` conclusion, a boolean `actionable` and a `summary`. A cached file that does not is kept as `.rejected-<time>` and the head is reviewed again. |
| Was the repair used? | `runs/<attempt>/repair-asked` exists (a prototype's `repaired` is read as asked). |
| What is it doing? | `runs/<attempt>/status`: the last stage and next action. |
| Is an agent still on a turn? | Herdr's status for its name: `working` is waited for. |
| Is there a pull request, and its reviews? | `gh pr list --head <branch>` and the account's reviews by commit, on GitHub, every run. |

A pane's state says a turn ended, never that the work is done.

## Flows: the steps an attempt runs

What an attempt does is a flow of steps. tend has two built in: `build`, the
loop above (implement, review, one repair, publish, act on requested
changes), and `propose`, which drafts a request. A project changes them with
`tend/flow.yaml`, read from the target branch as origin holds it, never from
the attempt's own branch, so an agent cannot change the flow it runs in. A
flow named in the file replaces the built-in one of that name; one it does not
name stays built in.

`tend flow <repo>` prints the flows in effect, each headed by where it comes
from. Saved as `tend/flow.yaml`, that text is a valid flow file that behaves
the same, and a place to start editing.

```yaml
build:
  start: implement
  steps:
    implement:
      agent: claude
      prompt: prompts/implement-{kind}.md
      result: commit
      next: {ok: review}
    review:
      agent: codex
      prompt: prompts/review.md
      result: verdict verdict-{head}.json
      next: {success: check, failure: publish, actionable: repair}
    check: # the project's own, its prompt in tend/prompts/check.md
      agent: codex
      prompt: prompts/check.md
      result: verdict check-{head}.json
      next: {success: publish}
    repair:
      continue: implement
      prompt: prompts/repair.md
      result: commit
      limit: 1
      next: {ok: review, spent: publish}
    publish:
      next: {ok: end, requested: request}
    request:
      continue: implement
      prompt: prompts/request.md
      result: commit or file reply-{review_id}
      next: {ok: review}
```

A step can say:

- `agent: claude` or `agent: codex`, for a fresh agent with tend's own
  flags for that harness, or `continue: <step>`, for that step's agent and
  conversation. A flow names no flag.
- `prompt`: a file under `tend/` on the target branch, or else one of tend's
  built-in prompts. Its `{branch}`, `{head}`, `{base}`, `{run}`, `{result}`,
  `{notes}`, `{findings}` and the like are filled in.
- `result`: `commit`, `file <name>`, `verdict <name>` (written in the run
  folder, in the reviewer's JSON shape) or `commit or file <name>`. With none,
  the turn ending is the result. Only verdicts named `verdict-{head}.json`
  are published as reviews of the pull request.
- `limit`: how many times the step is asked in an attempt.
- `next`: where each outcome goes, to a step, `end` or `person`. Outcomes are
  `ok`, a verdict's `success`, `failure` and `actionable`, and `spent` once
  the limit is used. An outcome `next` does not name stops for a person.

`publish` is tend's own: a flow places it and says only what follows it. A
flow without it publishes nothing. tend refuses, exiting 3 before any agent
starts, a file that is not YAML, holds a field it does not know, names a
step, harness, prompt or outcome that does not exist, or loops with no
limited step.

Each step asked is recorded in the run folder (`<step>-asked`, a head a
line), so a rerun after a kill waits for an agent still working, takes a
result already there, and never asks a step twice for the same thing.

## What it asks of Grove

Its command line and skills, and it never reads Grove's Markdown:

- `grove changes --json`, for the document root, the target branch and the
  attempt records, and as the sign that a repository uses Grove;
- `grove claim <cycle-id>` or `grove claim <change-id>`, to take the cycle,
  or an approved change no cycle lists, and record the attempt (a change by
  name needs Grove 0.14.0 or later, which tend needs);
- `/grove-implement --runner <cycle-id>` or `<change-id>`, as the
  implementer's prompt, so
  the project's `rules.delivery` lines are left out: tend delivers the work
  (Grove 0.12.0 or later);
- `grove review`, whose summary goes in the head's published review;
- `/grove-propose --unattended`, as the drafter's prompt in `tend propose`.

Grove does not know tend exists, and is useful without it.

## What it does to your agents

- Claude Code starts with `--permission-mode auto`, so routine commands ask
  nobody. Codex starts with `-s workspace-write -a never`, with network
  access (`-c sandbox_workspace_write.network_access=true`) so the tests it
  runs may serve on localhost, which its sandbox otherwise denies.
- tend delivers the work, never the agent. The implementer and the drafter
  are told so in their system prompt (`--append-system-prompt`), above any
  skill or project rule: work on this branch in this worktree, push nothing,
  open or edit no pull request, create no branch or worktree; the
  implementer is also told to commit, the drafter to commit only when tend
  asks. `git push` and `gh pr` are denied to
  them (`--disallowedTools`), which catches a mistake, not a determined
  agent. After each implementer turn tend compares the repository's
  worktrees and local branches with a record made just before the prompt.
  A worktree added, a new branch holding commits, or the attempt's worktree
  left on another branch stops for you, and nothing is reviewed or
  published. The agents run as your OS user, so they can reach the
  publishing account's token; a separate user for agents would close that.
- It answers the trust dialogs itself: whether to trust the folder, which
  Claude Code asks for each new worktree and Codex once for a repository,
  and Claude Code's whether to add the run folder (`--add-dir`). To skip
  that last one, trust tend's state folder once: run `claude` in
  `~/.local/state/tend` (or `$TEND_STATE`) and choose "Yes, I trust this
  folder"; every run folder below it is then added without asking. It never
  trusts Codex hooks: when Codex asks to review new or changed hooks, tend
  stops for you, since a trusted hook runs outside the sandbox. A startup
  notice that offers to be skipped (its hint says `esc skip`), such as
  Codex's "Update available", it skips by Esc alone, so it never updates an
  agent or picks a model, and says so. At any other screen an agent starts
  at, tend stops for you, saying the agent did not start.
- The reviewer may write in the attempt's run folder besides its checkout
  (`--add-dir`), where its verdict goes. So may the implementer, which writes
  its notes, the commit message and its replies there; the repair prompt
  carries the findings themselves.
  It answers for whatever repo you pass it. Any other dialog stops the loop
  and notifies you.
- Each attempt gets a Herdr worktree workspace. `herdr worktree create`
  needs a branch name and `grove claim` then makes its own, so tend deletes
  the `tend-HHMMSS` branch it made after the claim.

## What has been seen, and what has not

On 2026-11-14, as a prototype on Grove's ledger fixture (Herdr 0.9.5, Claude
Code 2.1.325, codex-cli 0.169.0, Grove 0.10.0), three runs of the same line
took one cycle through: two commits and an archived change in about 2 min
40 s; a Codex review of the real candidate, `success`, in 52 s; then a
defect planted by hand, which Codex failed as actionable, Claude repaired
and Codex passed, in 135 s together. With Codex at its usage limit, Herdr
reported `done`; the missing verdict file stopped the loop in 8 s, and the
next run carried on.

On 2026-11-15 the current script took a fresh attempt through with tend
killed (`kill -9`) during implementation, review and repair, and Claude killed
mid-repair. Each rerun waited for the agent still working without prompting it
again. The killed repair stopped as uncertain on every run until the mark was
deleted, then was asked once more and passed review. A planted defect (message
and test reworded together) was failed by Codex and repaired. Details are in
the plan's session handoff.

Not yet shown:

- a defect found in the implementer's own work;
- a dialog Herdr does not recognize (it reports `idle`, so tend prompts into
  it; the missing verdict then stops tend);
- whether the Herdr notification reaches a phone;
- tmux without Herdr, a saved SSH machine, two cycles at once.

## Next

Follow the [delivery slices and session handoff](docs/work/261114-reliable-orchestration/spec.md#delivery).
Recoverable execution comes first. Publication, proposal intake and portability
follow its demonstrated lifecycle; broader polling is deferred.
