# tend

How tend works, with diagrams: the [documentation site](https://mascah.github.io/tend/) (source in [`guide/`](guide/index.md)).

A small loop that takes agreed work through coding agents and brings back a
reviewed branch. Today it does one pass for one [Grove](https://github.com/mascah/grove)
cycle, or one approved change no cycle lists:

```text
grove claim
    -> Claude Code implements, headless (claude -p), in a worktree of its own
    -> Codex reviews the head commit
    -> one repair in the implementer's conversation, if the review asks
    -> Codex reviews the new head
    -> a branch and a verdict, waiting for you
    -> with a listed origin, a pull request and its reviews, as a bot account
```

Grove holds what is agreed. The agents run headless, so tend and they need
no terminal: close it and the work goes on. tend is a small Go program
between them. It reconstructs progress from Git, GitHub and its own run
files on every run, and asks you on GitHub when it needs a decision.

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
or later), `git` and `gh` on PATH, and Claude Code and Codex signed in. It
needs no terminal of its own:

```sh
bin/tend <repo> <id>          # a change, a cycle, issue-<n> or a Jira <KEY>
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

The agents run headless. Each turn is one process of its harness, `claude -p`
(`--session-id`, then `--resume`) or `codex exec`, in a session of its own so
it outlives a killed tend, with its prompt on its standard input. Everything
about an agent is under `TEND_STATE/agents/<name>/`: `agent.json` (its
harness, flags, session and the running turn's process), and for turn n
`n.prompt`, `n.out` (Claude's stream-json, which `tail -f` follows; Codex's
JSONL), `n.err`, and for Codex `n.last`. An agent is working while its turn's
process runs; once the turn has ended nothing is left that could act on the
prompt, so a step it left with nothing to show is asked once more, resuming
the same session, and never a third time. To look at a session, `claude
--resume <session>` in its worktree once its turn is over. Worktrees are made
by Git under `TEND_STATE/worktrees/`.

To watch the agents in panes instead, `TEND_TERMINAL=herdr` or `rex` runs them
there. In rex each worktree gets a session,
first named after its folder and found again by its ID in
`TEND_STATE/rex-sessions` (agents may rename it), and each agent a block named
after it. Its state is
what the agent reports with rex's program status protocol, as Claude Code and
Codex do, so a dialog the agent does not report shows as unknown and stops
tend rather than being prompted into. rex worktrees are made by Git under
`TEND_STATE/worktrees/`, and tend reports its own state to its tab.

- Exit 0: the head is reviewed, and the verdict is printed.
- Exit 2: no work item was given; tend prints `usage: tend <repo> <id>`.
- Exit 3: a person is needed. With a listed origin, the publishing account
  opens an issue on the repository: what happened, what the agent said last,
  where the work is, and how to answer (see [Decisions](#decisions-what-tend-asks-you)).
  Under Herdr or rex a notification says why too.
- Exit 4: the work item is taken: another tend with the same `TEND_STATE` is
  on it in that repository (or, for `tend propose`, another intake pass of
  the repository), and nothing was done. The lock goes when its holder exits
  or is killed.
- Exit 1: a command tend relies on failed unexpectedly.

Tends on different work items run at once, each with its own lock, worktree
and agents, and one intake pass per repository runs beside them. Tends that
start together set up one at a time: each fetches, tidies, makes its worktree
and claims while the others wait, and a claim Grove refuses, such as for a
piece that shares a planned file with one in flight, stops with Grove's
sentence and leaves no worktree or branch. A worktree or branch another
attempt or intake pass made is not held against an implementer; anything
else new outside its branch still stops the attempt.

Its own files live under `TEND_STATE` (default `$XDG_STATE_HOME/tend`, or
`~/.local/state/tend` without it): the
locks under `locks/`, and a folder per attempt under `runs/` holding its verdicts and repair
mark. They never enter the candidate, so a repository needs no Git excludes.
An attempt from before this, with a `.tend/` folder in its checkout, has
those files moved to its folder on the next run, and a name clash stops for
you rather than overwriting either.

After exit 3, or a kill, the same line carries on, with or without `origin`:

- An agent still working, as after tend was killed mid-turn, is waited for
  and never prompted again. A blocked or unknown agent stops for you.
- A timed-out or stalled prompt is not assumed undelivered, so tend never
  sends it twice to an agent that is still there.
- A step that was asked and left nothing, whose agent is gone (the terminal
  has no pane for it, as after a reboot or a restart of Herdr or rex), is
  asked once more in a new session, which finds the work in Git: nothing is
  left that could act on the earlier prompt. This applies to any step of the
  flow, the repair, requested changes and intake's follow-up. The second ask
  is recorded in `again-<record>` beside the step's record and does not count
  toward the step's limit. If it also leaves nothing, tend stops for you.
- An agent that is there but idle with nothing to show still stops for you,
  since it may have acted on the prompt.
- The one repair is used once asked: `repair-asked` records the reviewed head
  just before the prompt goes. If that head has no commit since and the
  implementer is there and not working, the outcome is unknown, and tend stops
  for you on every run until you commit a fix or delete `repair-asked` to have
  it asked once more. A repair refused before sending, such as by a blocked
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
a rerun after any interruption repeats neither. tend merges only where the
project's flow has a `land` step (see [Landing](#landing-merging-one-candidate-at-a-time));
otherwise merging stays with people.

When a person requests changes (GitHub's "Request changes") on the PR's head,
the next run gives their review and line comments to the implementer's own
session once, has the new head reviewed and publishes it, with the
implementer's reply in the review. `request-<review id>` in the run folder
records the ask just before it goes; asked with no commit since and nobody
working, it stops for you, as the repair does, unless the implementer is gone
and the request is asked once more. When the implementer changes
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
with a change request on its head from a person with write access, and on
each attempt whose tend died: its status says `working`, no tend is on it,
and, for Grove work, Grove still shows it in flight. It also runs
`tend propose` on each repository it
was given. In each of those repositories it then starts the oldest cycle that
Grove would claim and whose plan marks it `unattended`: marking a cycle so when
you plan it is how you let it be built with no one there. An attempt or request
waiting for you is left alone until you have answered its issue (see
[Decisions](#decisions-what-tend-asks-you)) or run tend on it. Started, watch
first checks each repository it was given can tell you things once no one is
at the terminal: a listed origin whose publishing account can push to it and
holds no token with the `workflow` scope; it refuses to start otherwise. It
also runs tend on any attempt Grove shows in flight on this machine, in a
worktree tend made, that has no run folder, as when tend was killed while
claiming, so that claim is not stranded.

With watch running, an attempt carries on with no one there after its tend is
killed, and after a reboot that took its agents too. The rerun waits for a
surviving agent, asks a step whose agent is gone once more, and stops for you
only where an agent is idle with nothing to show, blocked or in an unknown
state, or where a second ask also left nothing. On the first pass after you
upgrade, any attempt left `working` from before is run again.

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
place, when the pull request's head holds its head (as after GitHub's "Update
branch") and it holds no uncommitted files;
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

`go test ./...` builds tend and runs its acceptance tests, on half the cores
unless `-parallel` says otherwise:
`scenario_test.go` holds their fixtures, and the test file of each behavior,
such as `landing_test.go`, its scenarios. They drive the binary with real Git
and Grove and stand-ins in `testdata/bin` for Herdr, rex, `gh` and `curl`,
whose agents act at once. They start no agent and need no Grove checkout, but
need `grove` and `jq` on `PATH` (past a mise shim, the `grove` it falls back
to). Run one scenario by name with `go test -run 'TestTend/^publish$'`.

## How work reaches tend

With `tend watch <repo>...` running, each pass (every `WATCH_SECONDS`, default
300) picks up five kinds of work in the repositories it was given.

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
   several changes together. Without you, watch starts a build in one of two
   ways: with `watch: {at_once: N}` in the flow file it starts every change
   and cycle Grove calls claimable (see [Flows](#flows-the-steps-an-attempt-runs));
   without it, only a cycle you plan (`/grove-plan`) and mark
   `unattended: true`, and anything else only when someone runs
   `tend <repo> <id>`.
3. **Requested changes.** On an implementation pull request, GitHub's
   "Request changes" (not a plain comment) by an owner, member or
   collaborator goes to the implementer; anyone else's is left. Either it
   commits, the new head is reviewed, and the description is written again
   unless a person has edited it; or it makes no change and its reply goes on
   the pull request, and that review is not asked again. Request changes again
   to have it tried.
4. **Cleanup.** Once a pull request is merged or closed, a later pass removes
   its worktree, branch and terminal workspace, but not while a tend is on
   that attempt.
5. **A tend that died.** An attempt whose `status` says `working` with no
   tend on it, after a kill or a reboot, is run again, within the count where
   the flow file sets one. An attempt Grove no longer shows in flight, such as
   one retried with `grove claim --retry` from another checkout, is left
   alone.

Not picked up: plain comments, on any pull request, and comments on proposals
(revising a proposal is a person's); issues without the label; approved
changes in no cycle, and cycles not marked `unattended`, unless the flow file
sets `watch.at_once`; a proposal closed
without merging, which is never drafted again. In a repository without Grove, an issue labelled `tend-build` is built
only by `tend <repo> issue-<n>` (or `<KEY>`).

To see where things stand: `tend status <repo>` says what waits for whom;
tend watch's own output says what it ran; the agents work in the terminal's
workspaces, one per worktree; and when tend needs a person it stops and
notifies.

## Decisions: what tend asks you

When an attempt stops for a person and its origin is listed in `publish`, the
publishing account opens one issue on the repository, titled `tend needs you
on <work item>: <what happened>`. It says what happened, what the agent said
last, where the work is (the worktree on the runner, its commits and diff
size since the target) and the pull request if there is one. You answer by
closing it:

- **as completed**: tend carries on. tend watch runs the attempt again on its
  next pass, asking the agent that left nothing once more. If the question
  needs something done first, such as a branch you push yourself, do it,
  then close the issue.
- **as not planned**: tend leaves the attempt as it is, and says so in its
  status.

Only a close by someone other than the publishing account counts, since
every agent on the runner can act as that account. A run that stops the same
way while the issue is open opens no second one; an attempt that carries on
by other means, as when you run tend on it by hand, closes it saying how it
stands. `run/decision` in the attempt's folder holds the issue, the record
of the ask to make once more, and the question. The answer is read from
GitHub, never from a file you edit.

A failed read or repeatable write of GitHub, Git or Grove, such as a 502 from
GitHub or a fetch that times out, is not a question: tend says so, waits (15
seconds, doubling) and runs itself again from the start, which repeats
nothing done, with its status `working` meanwhile. After 6 tries in a row it
asks you. A push GitHub refuses (someone else's commits, a workflow file the
account may not push, no access) is a question at once.

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

`publish` gives `ok`, `requested` when a person with write access requested
changes on the head, and, in a flow that names it, `approved` when the newest
review on the head by such a person approves. In a flow that names it,
`publish` also gives `failed`: after publishing, with no change requested,
it waits for the checks on the head as `land` does (below), `tend/review`
left out, before it reads an approval. A failed check gives `failed`, the
failed checks' names and links going to the next prompt as `{checks}`. A
head with no check but `tend/review`, or one still unfinished, after
`TIMEOUT_MS` stops for a person. A flow that does not name `failed` reads no
check at `publish`.

### Landing: merging one candidate at a time

`land` is tend's own too, and no built-in flow has it, so a project that adds
none sees tend merge nothing. At `land`, tend takes a head that passed review
whose pull request is open and, one landing at a time for each repository on
this runner:

1. merges the target into the branch when the head lacks the target's tip,
   as a merge commit pushed without force. A clean merge keeps the verdict:
   the merge commit gets `tend/review` as carried from the reviewed head,
   with no new review and no new description. A conflict is aborted and
   gives `conflict`;
2. waits, for at most `TIMEOUT_MS`, for every check GitHub reports on the
   head, check runs and commit statuses alike. Success, neutral and skipped
   pass; any other finish gives `failed`, and the next step's prompt is
   given the failed checks' names and links as `{checks}`. A head with no
   check but `tend/review` stops for a person, so a project with no CI on
   pull requests cannot land;
3. squash-merges the pull request at exactly that head, as the publishing
   account, its title and number the commit's subject and its description
   the body. If the target moved it starts again from 1; if GitHub still
   expects a required check, it goes back to waiting; any other refusal
   stops for a person with GitHub's reason.

One candidate lands for each CI run. tend ships `prompts/conflict.md` and
`prompts/failed.md` for the steps that follow `conflict` and `failed`. A
supervised flow sends `publish`'s `approved` to `land`, so a person with
write access approves each head before it lands, and `tend watch` runs the
attempt again when it sees the approval. It sends `publish`'s `failed` to
the same `fix` as `land`'s, so a red pull request is fixed before anyone is
asked to approve it. `fix`'s limit counts asks from both: an attempt gets one
fix for failed checks in all. When `publish` finds failed checks and `fix` is
spent, it says so once on the pull request, naming the checks, and asks a
person as it would have. If `land` later finds `fix` spent, it stops for a
person as before:

```yaml
    publish:
      next: {ok: end, approved: land, requested: request, failed: fix}
    land:
      next: {ok: end, conflict: resolve, failed: fix}
    resolve:
      continue: implement
      prompt: prompts/conflict.md
      result: commit
      limit: 1
      next: {ok: review}
    fix:
      continue: implement
      prompt: prompts/failed.md
      result: commit
      limit: 1
      next: {ok: review}
```

An autonomous flow lands every head that passed review, with the same
`land`, `resolve` and `fix` steps:

```yaml
    publish:
      next: {ok: land, requested: request}
```

It need not name `failed` on `publish`, since `land` waits for the checks and
sends a failure to `fix`. It may, to keep a candidate's first CI run out of
the landing lock, so that run does not hold up another candidate's landing.

Set a branch rule on the target that requires the project's CI and
`tend/review` and requires branches to be up to date before merging: tend
checks the head itself, and the rule closes the moment between its check and
GitHub's merge. The publishing account must be allowed to merge and must not
be able to bypass the rule; for an account that administers the repository,
turn on "Do not allow bypassing the above settings". A rule that dismisses
approvals on new commits undoes an approval whenever tend merges the target
in. tend reads an approval once, at `publish`: a supervised project that wants
an approval withdrawn, or changes requested, while the candidate waits to
hold the merge sets the rule to require an approving review, so that GitHub
refuses and tend stops with its reason.

### Watch's count

The flow file also says how many attempts `tend watch` keeps going:

```yaml
watch:
  at_once: 2
```

Without `watch`, watch starts only cycles marked `unattended`, one a pass.
With it, each pass first runs tend again on attempts with changes requested,
then starts every cycle not done and change in no cycle that Grove calls
claimable, marked or not, oldest first, until that many attempts on this
runner in that repository are being worked on: a tend holds its lock,
whoever started it, building, in review or repairing. Watch waits for none of
them before its next pass, and intake runs beside them, outside the count.
`AT_ONCE` in watch's environment lowers the count for this machine, never
raises it; `AT_ONCE=0` starts nothing new and lets running work finish. The
count is read from the target, so an attempt that commits a different one
changes nothing. Each runner counts its own; Grove's claim still refuses an
overlap whichever machine asks.

An attempt that stops for a person frees its slot, so work that does not
overlap it goes on: it keeps its change in flight in Grove, so nothing claims
it again, and watch leaves it until a person runs tend on it. To park an
attempt that is running, interrupt its agent in its pane; tend stops for you,
and the slot is free.

`at_once` must be a whole number of at least 1; anything else in `watch`
stops tend as a flow file that is not valid.

Each step asked is recorded in the run folder (`<step>-asked`, a head a
line), so a rerun after a kill waits for an agent still working, takes a
result already there, and asks a step a second time only when its agent is
gone, never a third.

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
  the `tend-HHMMSS-<pid>` branch it made after the claim.

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
- tmux without Herdr, a saved SSH machine.

## Next

Follow the [delivery slices and session handoff](docs/work/261114-reliable-orchestration/spec.md#delivery).
Recoverable execution comes first. Publication, proposal intake and portability
follow its demonstrated lifecycle; broader polling is deferred.
