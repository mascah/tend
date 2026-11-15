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
    -> with a listed origin, the branch pushed and one pull request
```

Grove holds what is agreed. Herdr runs the agents where you can watch and
take over. tend is currently a prototype script between them. It reconstructs
progress from Git, its own run files and Herdr on every run.

The agreed [execution plan](docs/work/261114-reliable-orchestration/spec.md)
records the intended boundaries, review findings, delivery order and session
handoff. Start there when continuing development across sessions.

## Run it

Needs `herdr`, `grove`, `git`, `jq`, `perl`, and Claude Code and Codex signed in.
Run it inside a Herdr pane:

```sh
bin/tend <repo> <cycle-id>
bin/tend status <repo> <cycle-id>
```

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

With `origin`, tend publishes only if that URL is a line of `TEND_STATE/publish`;
otherwise it says so and keeps the verdict local. When listed, at the end of a
run it pushes the branch (never forced: a branch someone else pushed to stops
for you), opens or updates one pull request onto the target, and sets the
`tend/review` commit status on the reviewed head. The PR body carries the
verdict, earlier reviews of the attempt and `grove review`'s summary. The PR is
found on GitHub by its branch each run, so a rerun after any interruption
updates it rather than opening another, and a later commit has no passing
status until it is reviewed. tend never merges. Needs `gh` signed in.

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

`dev/test.sh` runs the regressions: real Git and Grove, a stand-in for Herdr
whose agents act at once. It starts no agent and needs no Grove checkout.

## Where the prototype reads progress

| Question | Read from |
| --- | --- |
| Is there an attempt? | The newest record for the cycle in `grove changes --json` (from `origin/grove/runs`, or the local `grove/runs` without `origin`), whose `checkout` still holds its `branch`. |
| Is it built? | Every change of the record is under `changes/archive/` on that branch, and the tree is clean. |
| Is the head reviewed? | `runs/<attempt>/verdict-<head>.json` names that head and has a `success` or `failure` conclusion, a boolean `actionable` and a `summary`. A cached file that does not is kept as `.rejected-<time>` and the head is reviewed again. |
| Was the repair used? | `runs/<attempt>/repair-asked` exists (a prototype's `repaired` is read as asked). |
| What is it doing? | `runs/<attempt>/status`: the last stage and next action. |
| Is an agent still on a turn? | Herdr's status for its name: `working` is waited for. |
| Is there a pull request? | `gh pr list --head <branch>` on GitHub, every run. |

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
  needs a branch name and `grove claim` then makes its own, so each attempt
  leaves a stray `tend-HHMMSS` branch.

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
- a tracker, or a second person acting through the pull request;
- a dialog Herdr does not recognize (it reports `idle`, so tend prompts into
  it; the missing verdict then stops tend);
- whether the Herdr notification reaches a phone;
- tmux without Herdr, a saved SSH machine, two cycles at once.

## Next

Follow the [delivery slices and session handoff](docs/work/261114-reliable-orchestration/spec.md#delivery).
Recoverable execution comes first. Publication, proposal intake and portability
follow its demonstrated lifecycle; broader polling is deferred.
