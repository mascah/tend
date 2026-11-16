# tend

A small loop that takes agreed work through coding agents and brings back a
reviewed branch. Today it does one pass for one [Grove](https://github.com/mascah/grove)
cycle:

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

Build it with Go 1.26 (`go build -o bin/tend .`). It needs `grove`, `git` and
`gh` on PATH, Herdr or rex, and Claude Code and Codex signed in. Run it inside
a Herdr pane or a rex terminal:

```sh
bin/tend <repo> <cycle-id>
bin/tend <repo> issue-<n>
bin/tend status <repo> [<cycle-id>]
bin/tend watch [<repo>...]
bin/tend propose <repo>
```

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
- Exit 3: a person is needed. A Herdr notification says why.
- Exit 4: another tend is running on this machine, and nothing was done.
  One tend runs at a time; the lock goes when it exits or is killed.

Its own files live under `TEND_STATE` (default `~/.local/state/tend`): the
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

`tend status` says which attempt the cycle has, whether a tend is on it now,
what it last recorded and the next action, including "run it again" when
tend was killed. It reads only, so it works while tend runs.

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
since tend set them. Plain comments are not acted on.

`tend watch` runs in a pane of its own and does that without you: every
`WATCH_SECONDS` (default 300) it runs tend on each attempt whose PR is open
with a change request on its head, and `tend propose` on each repository it
was given. An attempt or request waiting for you is left alone until you have
run tend on it.

`tend status <repo>`, without a cycle, says what waits for whom: approved
changes without a cycle (plan one), cycles not done with Grove's verdict,
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
planning its cycle stay with people.

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
<KEY>`. Anyone who can create or edit issues in a mapped project can start an
agent this way, so map only projects whose members you trust with that. A
handoff stopped between the pull request and the Jira comment is finished by
the next `tend propose`, which completes every intake folder with a `pr` and no
`linked`. tend never moves Jira issues through their workflow.

To cancel, stop tend (Ctrl-C or a kill). The agents keep their work and
their turns. Stop an agent in its pane if you want it stopped too; running
tend again carries on from what Git and the run folder hold.
`TIMEOUT_MS` sets the prompt wait timeout (default 30 minutes); a timeout
does not stop the agent.

To try it on a disposable repo, with a checkout of Grove beside this one:

```sh
dev/fixture.sh ~/tmp/ledger
bin/tend ~/tmp/ledger 261114-pr0t0
```

`dev/test.sh` builds tend and runs the regressions: real Git and Grove, a
stand-in for Herdr whose agents act at once. It starts no agent and needs no
Grove checkout.

## Where the prototype reads progress

| Question | Read from |
| --- | --- |
| Is there an attempt? | The newest record for the cycle in `grove changes --json` (from `origin/grove/runs`, or the local `grove/runs` without `origin`), whose `checkout` still holds its `branch`. |
| Is it built? | Every change of the record is under `changes/archive/` on that branch, and the tree is clean. |
| Is the head reviewed? | `runs/<attempt>/verdict-<head>.json` names that head and has a `success` or `failure` conclusion, a boolean `actionable` and a `summary`. A cached file that does not is kept as `.rejected-<time>` and the head is reviewed again. |
| Was the repair used? | `runs/<attempt>/repair-asked` exists (a prototype's `repaired` is read as asked). |
| What is it doing? | `runs/<attempt>/status`: the last stage and next action. |
| Is an agent still on a turn? | Herdr's status for its name: `working` is waited for. |
| Is there a pull request, and its reviews? | `gh pr list --head <branch>` and the account's reviews by commit, on GitHub, every run. |

A pane's state says a turn ended, never that the work is done.

## What it asks of Grove

Three things, through Grove's command line, and it never reads Grove's
Markdown:

- `grove changes --json`, for the document root, the target branch and the
  attempt records;
- `grove claim <cycle-id>`, to take the cycle and record the attempt;
- `/grove-implement <cycle-id>`, as the implementer's prompt.

Grove does not know tend exists, and is useful without it.

## What it does to your agents

- Claude Code starts with `--permission-mode auto`, so routine commands ask
  nobody. Codex starts with `-s workspace-write -a never`.
- It answers one kind of dialog itself: whether to trust the folder, which
  Claude Code asks for each new worktree and Codex once for a repository.
- The reviewer may write in the attempt's run folder besides its checkout
  (`--add-dir`), where its verdict goes. The repair prompt carries the
  findings themselves, so the implementer reads nothing outside its worktree.
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
- a tracker;
- a dialog Herdr does not recognize (it reports `idle`, so tend prompts into
  it; the missing verdict then stops tend);
- whether the Herdr notification reaches a phone;
- tmux without Herdr, a saved SSH machine, two cycles at once.

## Next

Follow the [delivery slices and session handoff](docs/work/261114-reliable-orchestration/spec.md#delivery).
Recoverable execution comes first. Publication, proposal intake and portability
follow its demonstrated lifecycle; broader polling is deferred.
