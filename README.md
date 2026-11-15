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
```

Grove holds what is agreed. Herdr runs the agents where you can watch and
take over. tend is currently a prototype script between them. It reconstructs
progress from Git, its own run files and Herdr; restart recovery is incomplete.

The agreed [execution plan](docs/work/261114-reliable-orchestration/spec.md)
records the intended boundaries, review findings, delivery order and session
handoff. Start there when continuing development across sessions.

## Run it

Needs `herdr`, `grove`, `git`, `jq`, `perl`, and Claude Code and Codex signed in.
Run it inside a Herdr pane:

```sh
bin/tend <repo> <cycle-id>
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

The same line attempts to carry on after exit 3 or a kill, with or without
`origin`. Recovery is still incomplete: an interrupted repair can be skipped,
and a surviving agent is not reconciled before a restart prompts again. See the plan before relying on it. `TIMEOUT_MS` sets the prompt wait timeout
(default 30 minutes); stopping the underlying agent has not been demonstrated.

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
| Was the repair used? | `runs/<attempt>/repaired` exists. |

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

The script was renamed from that prototype (`.loop/` to `.tend/`, its
messages and branch prefix) and has not been run end to end since.

Not yet shown:

- a defect found in the implementer's own work;
- a remote: a push, a pull request, a tracker, a second person;
- an agent killed in the middle of a turn;
- whether the Herdr notification reaches a phone;
- tmux without Herdr, a saved SSH machine, two cycles at once.

## Next

Follow the [delivery slices and session handoff](docs/work/261114-reliable-orchestration/spec.md#delivery).
Recoverable execution comes first. Publication, proposal intake and portability
follow its demonstrated lifecycle; broader polling is deferred.
