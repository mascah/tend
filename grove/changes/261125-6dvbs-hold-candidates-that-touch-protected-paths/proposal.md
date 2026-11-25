# Hold a candidate that touches a protected path for the owner

## Problem

tend reads the flow file and its prompts from the target branch, never from
the candidate (`grove/specs/flows/spec.md` at 2a8ae80, "The flow is read from
the target branch"), so a candidate cannot change the flow it runs in. Once
it lands, though, its edit governs every attempt after it.

In a supervised flow a head lands only after a person with write access
approves it (`grove/specs/landing/spec.md` at 2a8ae80, "A person's approval
is an outcome of publishing"). In an autonomous flow, whose `publish` sends
`ok` to `land`, `landStep` in `landing.go` at 2a8ae80 lands any head that
passed review, or is carried from one that did, once its checks pass. No one
is asked, even when the head edits `tend/flow.yaml`, a CI workflow or an
agent's instructions.

Direction decision 3a (`docs/work/261121-autonomous-direction/direction.md`)
names the rules agents work under as protected paths that still need the
owner. The request (GitHub issue
[#36](https://github.com/mascah/tend/issues/36)) calls this "the first check
that lets a project switch to the autonomous flow safely."

## Intended outcome

When a candidate's change against the target touches a protected path, it
lands only once a person with write access has approved that head. Without
that approval tend merges nothing and stops for a person, naming the
protected paths the change touches. The check is mechanical: tend reads the
changed paths from Git and asks no agent. The protected paths are read from
the target branch, never from the candidate. A project can name its own in
its flow file, which is itself always protected. Without that, a default
list applies, covering the files decision 3a names that a repository can
hold. The check applies to every head that reaches `land`: a head that
passed review, a head carried across a clean merge of the target, and a head
the implementer made at `resolve` or `fix` and that was reviewed again.

## Scope

### In scope

- At `land`, finding the paths the head changes against the target, and
  matching them against the protected paths in effect.
- Holding a head that touches one: landing it only with a person's approval
  of that head, or of the head it is carried from, and otherwise stopping
  for a person with the paths named.
- The protected paths in effect: the project's, read from the flow file on
  the target, or the default list. `tend flow` prints them.

### Out of scope

- The repository ruleset. It lives in GitHub, not in the repository, so a
  candidate cannot change it.
- Judging product identity, money or data (decisions 3b and 3c). Those need
  an agent's verdict and come later.
- Proposal pull requests, which hold only Grove documents.
- `tend watch` running a held attempt again after a person approves it (see
  open question 5).

## Decisions

These come from the request (issue #36). No decision was made beyond it.

- **The check is mechanical.** "It reads the changed paths from Git and asks
  no agent, so an agent's verdict cannot waive it."
- **Only a person's approval releases a held head.** "When a candidate's
  change against the target touches a protected path, it does not land
  unless a person with write access has approved that head. Otherwise it
  stops for a person and says which paths it touched."
- **The protected paths come from the target.** "The protected paths are
  read from the target branch and never from the candidate. A project can
  name its own, and the file that names them is protected too."
- **The default covers decision 3a's file-backed rules.** "the flow file and
  `tend/` prompts, CI workflows, CODEOWNERS, the principles file,
  `grove.yaml`, agent instructions (`CLAUDE.md`, `AGENTS.md`), skills and
  agent settings (`.claude/`), deploy config, and lockfiles (decision 3c)."
- **It holds for every head that reaches land.** "It holds for heads carried
  by a clean merge of the target, and for heads made by `resolve` or `fix`."

## Open questions

No one attended this draft, so every choice below has a recommended answer
that the change specs and tasks follow. None of them must be answered before
review. The owner can confirm or change each one when reviewing. Questions 1
to 4 decide how T-01 is built, and question 5 decides T-02.

1. **Where does a project name its protected paths?** Recommended: a
   `protect` list in `tend/flow.yaml`. tend already reads that file from the
   target and validates it, `tend flow` already prints it, and it is
   protected whatever it says, which meets "the file that names them is
   protected too" with no second file. The alternative is a separate file,
   such as `tend/protected`. That file would sit under `tend/` and so be
   protected too, but it would need its own reading and validation.
2. **Does a project's list replace the default or add to it?** Recommended:
   it replaces the default, and `tend/` stays protected whatever the list
   says. A project can then drop an entry that does not fit it, such as a
   lockfile, but cannot unprotect the file that names the list. Adding to
   the default is safer, but it gives a project no way to drop an entry.
3. **How is an entry matched?** Recommended: an entry is a path from the
   repository's root. An entry ending in `/` covers that folder and
   everything in it. An entry with no `/` at all matches a file of that name
   in any folder, so `CLAUDE.md` covers a nested `CLAUDE.md` and
   `CODEOWNERS` covers `.github/CODEOWNERS`. `*` matches within one path
   segment, as in `wrangler.*`. An empty entry makes the flow file not
   valid. This is close to the CODEOWNERS file that
   `docs/work/261116-runner-contract-and-pipeline/new-project-setup.md`
   proposes, without its full syntax.
4. **What is the default list?** Recommended: `tend/`, `.github/workflows/`,
   `CODEOWNERS`, `PRINCIPLES.md`, `grove.yaml`, `CLAUDE.md`, `AGENTS.md`,
   `.claude/`, `.agents/` (the Codex skills Grove writes), `.codex/`;
   deploy config `wrangler.*`, `fly.toml`, `vercel.json`, `netlify.toml`,
   `render.yaml`, `Procfile`, `Dockerfile`; lockfiles `go.sum`,
   `package-lock.json`, `npm-shrinkwrap.json`, `yarn.lock`,
   `pnpm-lock.yaml`, `bun.lock`, `bun.lockb`, `Cargo.lock`, `Gemfile.lock`,
   `poetry.lock`, `uv.lock`, `Pipfile.lock` and `composer.lock`. The
   principles file is `PRINCIPLES.md`, as the new-project setup names it.
   The same setup also protects `grove/cycles/`, so that no agent can mark
   its own work `unattended`. That folder is not in the request's list, so
   it is left out; the owner may want it added.
5. **What happens to a held attempt once a person approves it?**
   Recommended: the hold is a stop for a person, so tend exits 3 like
   `land`'s other refusals. A person who approves the head then runs tend
   again, and it lands. `tend watch` leaves the attempt alone, as it does
   any attempt that needs a person. Having watch run it again on an
   approval would change the watch-and-status spec, and is better as its
   own change once the autonomous flow is in use. The alternative is to
   build that rerun here.
6. **Routine choices inside the approach:** the names and placement of
   tests and helpers, where the default list is kept in the code, and the
   wording of the stop. Recommended: leave these to the implementer, who
   records each choice. They decide parts of T-01 and T-02.

## Related documents

- Requirements: [landing](specs/landing/spec.md), [flows](specs/flows/spec.md)
- Tasks: [tasks.md](tasks.md)
