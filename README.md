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
take over. tend is the script between them, and keeps no state of its own.

## Run it

Needs `herdr`, `grove`, `git`, `jq`, and Claude Code and Codex signed in.
Run it inside a Herdr pane:

```sh
bin/tend <repo> <cycle-id>
```

- Exit 0: the head is reviewed, and the verdict is printed.
- Exit 3: a person is needed. A Herdr notification says why.

After exit 3, or a kill, run the same line again and it carries on.
`TIMEOUT_MS` bounds each agent turn (default 30 minutes).

To try it on a disposable repo, with a checkout of Grove beside this one:

```sh
dev/fixture.sh ~/tmp/ledger
bin/tend ~/tmp/ledger 261114-pr0t0
```

## Where it is comes from Git

| Question | Read from |
| --- | --- |
| Is there an attempt? | The newest record on `grove/runs` for the cycle, whose `checkout` still holds its `branch`. |
| Is it built? | Every change of the record is under `changes/archive/` on that branch, and the tree is clean. |
| Is the head reviewed? | `.tend/verdict-<head>.json` names that head. |
| Was the repair used? | `.tend/repaired` exists. |

A pane's state says a turn ended, never that the work is done.

## What it asks of Grove

Three things, through Grove's command line, and it never reads Grove's
Markdown:

- `grove changes --json`, for the document root;
- `grove claim <cycle-id>`, to take the cycle and record the attempt;
- `/grove-implement <cycle-id>`, as the implementer's prompt.

Grove does not know tend exists, and is useful without it.

## What it does to your agents

- Claude Code starts with `--permission-mode auto`, so routine commands ask
  nobody. Codex starts with `-s workspace-write -a never`.
- It answers one kind of dialog itself: whether to trust the folder, which
  Claude Code asks for each new worktree and Codex once for a repository.
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

1. A repo with a remote: push the branch, open the pull request, post the
   verdict on it.
2. Proposing: requests from a configured command, each drafted with
   `/grove-propose --unattended` into its own pull request, with a cap on
   how many wait unmerged. Needs
   [grove#164](https://github.com/mascah/grove/pull/164).
3. A tick over the repos you register.
