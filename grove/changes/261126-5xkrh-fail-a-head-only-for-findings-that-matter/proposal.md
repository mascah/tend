# Fail a head in review only for findings that matter

## Problem

GitHub issue [mascah/tend#45](https://github.com/mascah/tend/issues/45)
reports that the reviewer fails a head for every edge case it can construct,
so a change does not converge. Pull request #42 (change 261125-6dvbs,
protected paths) went through five Codex reviews on 2026-11-25; each passed
the previous findings and raised new ones, each marked actionable. The issue
says most findings needed a temporary test the reviewer wrote to construct
the case, several were corners no one will hit, and one only changed which
reason a stop gives. Each round cost a full implementer turn, a full review
and about 15 minutes, and a failing verdict blocks `land` and the ruleset's
`tend/review` status, so the owner could only ask for another round or merge
with admin bypass.

The issue names three causes, which the code at e80da30 bears out:

- The built-in prompt, `flows/prompts/review.md` at e80da30, asks for
  "material findings: a defect, an unmet requirement, or a claim in {notes}
  the code does not support". The specs state requirements absolutely, so a
  constructed corner that departs from one reads as an unmet requirement.
- `actionable` says whether a finding needs a decision from the owner, not
  whether it matters (`grove/specs/review/spec.md`, Purpose, at e80da30).
- Each review is prompted with the whole change, `git diff {base}...HEAD`,
  and nothing of the earlier verdicts, so a finding on code unchanged since
  the previous round can first appear in round five.

Commit eecd7e6 (#52) added a stopgap for tend's own repository only:
`tend/prompts/review.md` states the bar, and tend reads it before the
built-in prompt. Every other project still gets the old bar.

## Intended outcome

- The built-in review prompt and the review spec state the bar: a head fails
  only for a defect a realistic use would hit, a requirement missed in its
  main case, a security or data hole that can be reached, or a claim in the
  implementer's notes the code does not support. Corners that need contrived
  setup, which of two reasons a stop gives, wording and polish do not fail a
  head. A project can tune the bar in its own `tend/prompts/review.md`, as it
  can replace any built-in prompt today.
- Findings that do not fail the head are written as notes in the verdict and
  shown on the pull request, apart from the findings that fail it, so the
  owner can file or drop them. They never reach the repair prompt.
- A later review of the same attempt is given the earlier verdicts. It
  judges whether the findings that failed an earlier head are fixed and what
  the commits since then changed, and does not fail the head over code an
  earlier review saw and passed over.

## Scope

### In scope

- The built-in review prompt's bar, its notes, and what it says of earlier
  rounds.
- An optional `notes` string in the verdict, and the verdict's validity with
  and without it.
- A `{earlier}` placeholder filled for every step, naming the attempt's
  earlier verdicts on commits of its branch.
- The review that publication posts: notes under their own heading, which
  do not change the `tend/review` status.
- tend's own `tend/prompts/review.md` stopgap (see open question 6).

### Out of scope

- A cap on review rounds before a person decides. The issue asks for one
  only "if the above is not enough"; see open question 3.
- A separate step that verifies each failing finding before the repair is
  spent. A project's flow file can add one today, as the handoff notes say;
  see open question 4.
- Severity per finding, or a structured list of findings in the verdict.
- Overriding a failing verdict from GitHub, or changing what `land` or the
  ruleset require of `tend/review`.
- Whether the reviewer's judgement is right, which tend does not check.

## Decisions

No decision was made beyond the request. The request (GitHub issue #45) sets
these constraints:

- "A head fails review only for findings that matter: a defect a realistic
  use would hit, a requirement missed in its main case, a security or data
  hole that is reachable, or a claim in the notes the code does not support.
  Corners that need contrived setup, wording, and polish do not fail a head."
- "Findings that do not fail the head are still reported, on the pull
  request, so the owner can file them or drop them."
- "A later round does not move the goalposts: it judges whether the earlier
  findings are fixed and what the new commits changed, and does not fail the
  head over code the previous review saw and passed over."
- "The bar is stated in the built-in review prompt and the review spec, so a
  project can tune it in its own `tend/prompts/review.md`."

## Open questions

None must be answered before review; each has a recommended answer that
`/grove-implement` builds and records.

1. **A prompt change only, or a verdict field tend reads?** Recommended: keep
   `conclusion` as the one thing that fails a head, with no severity per
   finding, and add an optional `notes` string for findings that do not fail
   it. Reason: `conclusion` already is the blocking signal, so a severity
   would duplicate it; but notes kept apart from `summary` let tend show them
   under their own heading and keep them out of `{findings}`, so the one
   repair is not spent on polish. Optional, so a cached verdict or a
   project's own prompt without it stays valid. Decides T-01 and T-02.
2. **Can a later round fail a head over unchanged code at all?** The request
   says it does not. Recommended: follow the request, with one exception: a
   security or data hole that can be reached still fails the head wherever
   it is. Reason: such a hole costs more to land than one more round, and it
   is the narrowest of the four kinds of finding. Without the exception, the
   specs drop that sentence. Decides T-01.
3. **A cap on rounds?** Recommended: none in this change. Reason: the repair
   is already limited to one, every later round is one a person asked for,
   and the bar and the earlier-rounds rule should be seen working before a
   cap is added; a cap would be its own change. Decides nothing here.
4. **A step that verifies failing findings before the repair?** Recommended:
   not in the built-in flow. Reason: it costs a second Codex turn on every
   failure, and a project can add it in its flow file with no change to
   tend, as item 5 of the Claude Ship notes in
   `docs/work/261121-autonomous-direction/handoff-261124.md` says. Decides
   nothing here.
5. **How notes reach the pull request.** Recommended: in the comment review
   tend already posts for the head, after the summary, under a heading
   saying they do not fail the head; without publication they stay in the
   verdict file. Reason: it is the one place the owner already reads each
   verdict, and it adds no new kind of GitHub post. Decides T-02.
6. **tend's own `tend/prompts/review.md`.** Recommended: delete it, so tend's
   flow falls back to the built-in prompt, as the flows spec's scenario "A
   flow names tend's own prompts" says. Reason: commit eecd7e6 says it can go
   once the built-in prompt holds the bar, and left in place it would lack
   `{earlier}` and `notes`. It is a protected path, so the owner merges it.
   Decides T-03.
7. **Routine choices.** Recommended: the wording of the prompt and of the
   notes heading, the form of `{earlier}` (one line per earlier verdict,
   giving its short head, conclusion and file path, or `none`), the helper
   that builds it, and test names and layout are the implementer's. Reason:
   they do not change behavior the specs fix. Decides T-01 and T-02.
8. **Does approved change 261125-rf036 (fix failing CI before a person is
   asked) need to land first?** It also changes publication and
   `publication.go`, though not the review a head gets. Recommended: no
   dependency. Reason: the two touch different parts of publishing, so
   whichever lands second merges the other's code without changing what it
   does. Decides T-02.

The Purpose of the maintained review spec defines actionable as "every
finding is an in-scope code fix"; at archive it should read "every finding
that fails the head".

## Related documents

- Requirements: [specs/review/spec.md](specs/review/spec.md),
  [specs/publication/spec.md](specs/publication/spec.md)
- Tasks: [tasks.md](tasks.md)
