# Run steps from a flow file

## Problem

How work flows through tend is written into its code. `main.go` at ed385ad
holds the whole loop for a build (implement, review, one repair, publish, a
person's requested changes) and for intake (draft, commit, open the proposal)
as Go, one function after another. Adding a step, dropping an approval, or
running a different flow in another project is a change to tend, a release,
and an update of every runner.

The owner wants projects to opt into more autonomy than tend gives today:
agents that decompose goals, check proposals and plan, with different rules
per project (`docs/work/261121-autonomous-direction/direction.md`, decisions
1 to 7). Each of those is one more loop written the way review is written
today. Only review works as a checked step now: the reviewer must write a
verdict file, tend validates it, and tend decides what follows
(`grove/specs/review/spec.md` at ed385ad, "Only a valid verdict counts").

The acceptance suite grew the same way. `dev/test.sh` at ed385ad is one Bash
file of 1,329 lines with 72 scenarios that run one after another, in 6 min
38 s on the owner's machine, and none can be run alone.

## Intended outcome

tend runs steps. A step is an agent given a prompt, a result tend checks, and
a rule for what that result allows next. A flow is a named set of steps. tend
ships the flows it runs today, so a project that holds no flow file sees no
change in behaviour. A project that holds one changes how its work flows by
editing that file, in a pull request, with no change to tend.

The tests are rewritten in Go with the code, so one scenario can be run alone
and all of them in parallel.

## Scope

### In scope

- The step engine: prompting a step's agent once, checking its result,
  choosing the next step, bounding repeats, and recovering after a stop or a
  kill without prompting twice.
- The flow file: its place in the project, what a step can say, and reading it
  from the target branch.
- Today's build and intake loops as built-in flows.
- `tend flow`, which prints the flows in effect.
- Go tests in place of `dev/test.sh`.

### Out of scope

- Any new step for the autonomous mode (decomposing a goal, checking a
  proposal, planning), the standing agent, the spend cap, several pieces of
  work at once and the merge train. Each is a later change that this one makes
  possible (direction document, implied order).
- Any change to Grove.
- What the seven maintained specs say. The built-in flows keep their
  behaviour; where a sentence of theirs cannot hold, the implementer says so
  and stops, as decision 13 says.
- How agents are started, their flags, and the dialogs tend answers
  (`grove/specs/terminals/spec.md`). A flow names a harness and nothing else
  about it.
- The 60-second wait under rex before Claude's dialog for adding the run
  folder is answered. It is wrong today and independent of this change;
  bugfix change 261121-7a3cz fixes it.

## Decisions

The owner decided these in the conversation of 2026-11-21 that the direction
document records.

1. How work flows is a file the project holds, not code in tend. The owner,
   asked whether changing how work flows should be a change to tend or to a
   file in the project: "B, a file in the project". Reason: autonomy is a mode
   a project opts into, and what needs a person differs by project.
2. tend drives and agents do steps; no standing agent holds authority. The
   owner on Gas City, where an orchestrator runs workflows held as data and
   gates are checks: "I don't want to use gas city itself, but I want to be
   influenced by it."
3. tend is rewritten, and its tests with it; the existing suite is not a bar
   to pass first. The owner: "rewrite them all! `tend` is a project I
   literally started a week ago, we can rewrite the whole thing."
4. A step either starts a fresh agent or continues the agent of an earlier
   step, and tend reads a project's rules from the target branch, never from a
   candidate. Both came from the review of the direction document, whose
   fixes the owner accepted ("ok with the rest", then "ok").

5. The flow file is YAML, at `tend/flow.yaml`, with each prompt a file beside
   it; tend takes the dependency that needs. The draft proposed JSON to keep
   tend on the standard library alone. The owner: "No, we can use
   dependencies, I never said tend was std lib only." Reason: the file is what
   people read, review and protect, and YAML takes comments.

The owner accepted the draft's other choices as written: "Agree with the rest
of the calls".

6. This is one change. The engine cannot be accepted without a flow to run,
   and a tend that runs builds through the engine and intake through the old
   code has two ways of doing everything.
7. Both built-in flows move in this change, `build` and `propose`. A build is
   a Grove cycle or a work item, with a person's requested changes.
8. Claiming, the worktree, the lock, tidying, `watch` and `status` stay tend's
   own and are not steps. A flow starts once the attempt has its branch and
   worktree. Publishing is tend's own too, and a flow places it by name.
9. A flow names an agent's harness, `claude` or `codex`, and cannot set its
   flags. The flags are what keep an agent from pushing or opening a pull
   request (`grove/specs/terminals/spec.md`, "How agents start"), so a flow
   file must not be able to loosen them.
10. A flow that can loop for ever is refused. Every loop must pass through a
    step with a limit. Today's "one repair" is the repair step's limit of 1.
11. A flow that is not valid exits 3, a person being needed, and starts or
    prompts no agent.
12. `dev/test.sh` is deleted in this change, once each of its scenarios has a
    Go test of the same name or a reason in `implementation.md` for having
    none. The four stand-ins (Herdr, rex, `gh`, `curl`) are kept as scripts
    under `testdata/`, since they hold what a week of live runs taught; only
    the scenarios are rewritten.
13. The seven maintained specs are not modified. They describe the built-in
    flows, including the names of run files such as `repair-asked`. If the
    implementer finds a sentence the engine cannot keep, it blocks the task
    with the sentence and the reason, and does not rewrite the spec.
14. `tend flow` exists. Without it a project has nothing to start its file
    from, and nobody can see which flow is in effect.

## Open questions

None.

## Related documents

- Requirements: [specs/flows/spec.md](specs/flows/spec.md)
- Tasks: [tasks.md](tasks.md)
- Direction: `docs/work/261121-autonomous-direction/direction.md`
