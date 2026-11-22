# Land reviewed candidates one at a time

## Problem

Once several attempts run at once, each reviewed candidate was tested against
the target as it stood when the candidate was pushed, and nothing brings it up
to date or merges it but a person.

- tend merges nothing. The publication spec at 40ca101 ends its Boundaries
  with "tend never merges a pull request; merging stays with people", and
  `main.go` at 40ca101 holds no call that merges one.
- After one candidate merges, the checks on every other open candidate
  describe a target that no longer exists. A person has to update each
  branch, wait for its checks and merge it, one after another.
- GitHub's merge queue would do this, but GitHub's documentation offers it
  "in any public repository owned by an organization, or in private
  repositories owned by organizations using GitHub Enterprise Cloud"
  ([Managing a branch protection rule](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/managing-a-branch-protection-rule),
  read 2026-11-21). `mascah/tend` and `mascah/grove` are owned by a personal
  account. The documentation does not say the queue is unavailable there;
  that is the reading of its one sentence on availability.
- Branch protection is available on `mascah/tend`: on 2026-11-21 GitHub's API
  answered "Branch not protected" for `main`, and the repository had no
  ruleset. A CI run there took about 3 minutes.

[direction.md](https://github.com/mascah/tend/blob/main/docs/work/261121-autonomous-direction/direction.md)
asks for this. Decision 1 has agents merge. Decision 2 keeps human approval
and the number of simultaneous builds as separate choices. Decision 9 has
finished candidates land through a merge train, each "tested by CI combined
with the target and the candidates ahead of it", a failed one going back to
its implementer with its new head reviewed again.

## Intended outcome

A project's flow can hold a `land` step. There tend takes one reviewed
candidate at a time: it merges the target into the candidate when the target
has moved, waits for the checks on that head, and has GitHub squash-merge
exactly that head. A candidate that does not merge cleanly, or whose checks
fail, goes back to its implementer through the flow, and its new head is
reviewed again. A head tend made by merging the target cleanly keeps the
verdict of the head it came from.

Which candidates reach `land` is the flow's to say. A supervised flow sends
there a head a person with write access approved; an autonomous flow sends
every head that passed review. The built-in flows have no `land`, so a
project that adds none sees no change, and people merge as before.

## Scope

### In scope

- The step `land` in a flow: that it is tend's own, its outcomes, and what
  makes a flow with it not valid.
- Landing a candidate: which head may land, one landing at a time, merging
  the target in, waiting for checks, the squash merge, and what each failure
  leads to.
- A head carried by a clean merge of the target: its `tend/review` status,
  no review and no reviewer for it, and its pull request's description left
  as it is.
- The outcome `approved` of `publish`, and `tend watch` running an approved
  attempt again.
- Prompts tend ships for the steps a conflict and a failed check lead to,
  which a project's flow can name.
- A maintained spec `landing`, the publication spec's sentence on merging,
  and the README: a supervised and an autonomous flow file, and the branch
  rule to set.

### Out of scope

- Several candidates tested at once. The owner chose one at a time; the
  ceiling is one landing for each CI run.
- GitHub's merge queue.
- A `land` step in the built-in flows.
- Setting or reading a repository's branch rule. tend checks the head itself
  before it merges, and the rule "require branches to be up to date", which
  the README tells the owner to set, is what closes the moment between tend's
  check and GitHub's merge.
- Reading an approval again before the merge. tend reads it once, at
  `publish`. If the approver withdraws it or requests changes while the
  candidate waits for its turn or its checks, tend still merges, unless the
  branch rule requires an approving review, which the README tells a
  supervised project to set; GitHub then refuses and tend stops with its
  reason.
- An order of landing across machines. Each runner lands one candidate at a
  time; between runners GitHub's rule decides.
- Running a failed check again, and deleting a merged branch.
- Judging whether merged work crosses the owner's boundaries of direction
  decisions 3b and 3c: step 4.

## Decisions

The owner settled the first three in an explore on 2026-11-21, each by
choosing the option quoted, and answered "Yes" to the round of decisions
before writing, which holds the rest.

1. Who runs the train: "tend, one at a time". "A landing step a flow opts
   into. tend merges the target into the front candidate, waits for CI on
   that head, then squash-merges. GitHub's 'branch must be up to date' rule
   is the guarantee. Works on personal repositories. Ceiling: one landing per
   CI run, about 3 minutes, so not 'several tested at once' from decision 9."
   The question said that any option where tend lands work "ends 'tend never
   merges', only in a project whose flow file has the landing step."
2. The verdict: "Carried if the merge is clean". "tend checks the new head is
   exactly Git's clean merge of the reviewed head and the target, then sets
   tend/review on it as carried from the reviewed head. Only CI judges the
   combination. Any commit the implementer makes, to resolve a conflict or
   fix a CI failure, is a new head and is reviewed again." Reason: reviewing
   every head the train makes would slow each landing to a reviewer's pace.
3. A person's approval: "An approving review". "The publish step gains an
   outcome 'approved', counted only from people with write access, on the
   current head. A supervised flow sends 'approved' to the landing step; an
   autonomous flow sends every passing head there. tend watch runs the
   attempt again when it sees the approval, as it does for requested
   changes."
4. `land` is tend's own, like `publish`, with outcomes `ok`, `conflict` and
   `failed`. A flow with `land` and no `publish` is not valid. The built-in
   flows stay without it.
5. What `land` does:
   - It lands only a head that passed review and has an open pull request;
     any other head stops for a person.
   - One landing runs at a time per repository on this runner.
   - If the head lacks the target's tip, tend merges the target in, as a
     merge commit and never by a forced push. A conflict is aborted and gives
     `conflict`.
   - It waits for every check GitHub reports on the head, check runs and
     commit statuses alike; success, neutral and skipped pass, and any other
     finished check gives `failed`.
   - It squash-merges exactly that head, with the pull request's title and
     description as the commit message.
   - If the target moved meanwhile, it starts again from the merge. If a
     required check is pending or not yet reported, it goes back to waiting.
     Any other refusal stops for a person with GitHub's reason.
6. A head with no check other than `tend/review` stops for a person. Reason:
   decision 9's train is a test by CI, and merging with none would be a train
   in name only.
7. A carried head is a merge commit whose first parent passed review or is
   itself carried, whose second parent is on the target, and whose tree is
   exactly Git's merge of the two. tend sets `tend/review` on it as carried,
   with no review posted and no reviewer prompted.
8. `publish` gives `approved` only in a flow that names it, when the newest
   review on the head from a person with write access approves. An approval
   holds for a head carried from the approved one.
9. A new maintained spec `landing` holds the train, and the publication
   spec's sentence becomes that tend merges only at a `land` step.
10. This change depends on 261122-9vbmc, for several attempts at once, and on
    261122-fh9h2, which says who a person with write access is.
11. Routine: the wait for checks is bounded by `TIMEOUT_MS`; tend ships
    prompts for the conflict and failed-check steps; the README shows a
    supervised and an autonomous flow file and the branch rule to set. Test
    layout, private helpers, file names under TEND_STATE that no spec names,
    and the wording of one-sentence messages are the implementer's to choose.

12. From a review on 2026-11-22, each fix approved by the owner:
    - Which finished checks pass, that commit statuses count, and that a
      merge refused for a required check still pending or not yet reported
      goes back to waiting. Reason: a second workflow can register after the
      first finishes, as `check` does after `title` in tend's own repository.
    - The README says the branch rule must bind the publishing account, and
      that a supervised project requires an approving review in the rule.
      Reason: an account that can bypass the rule loses the up-to-date
      guarantee, and tend reads an approval only once.

## Open questions

Can wait: how long `land` waits for a first check before it says the project
has no CI on pull requests. Decision 11 bounds every wait for checks by
`TIMEOUT_MS`, four hours by default, which is long for a project that simply
has none. It decides how 261122-rpkqa:T-02 is built: `/grove-implement`
builds the `TIMEOUT_MS` bound and records it for the reviewer.

## Related documents

- Requirements: [specs/landing/spec.md](specs/landing/spec.md),
  [specs/flows/spec.md](specs/flows/spec.md),
  [specs/publication/spec.md](specs/publication/spec.md),
  [specs/review/spec.md](specs/review/spec.md),
  [specs/watch-and-status/spec.md](specs/watch-and-status/spec.md)
- Tasks: [tasks.md](tasks.md)
