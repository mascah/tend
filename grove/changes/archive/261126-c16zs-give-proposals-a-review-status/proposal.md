# Give proposal pull requests a tend/review status

## Problem

The branch rule on `main` requires `tend/review`, so that the publishing
account can only merge a head that passed review (guide/setup.md, section 6,
at c6505b1). Only implementation pull requests ever get that status: it is set
by the publish step for each verdict and by `land` for a carried head
(publication.go and landing.go at c6505b1); intake opens a proposal and posts a
summary review but sets no status (intake.go, `proposal`, at c6505b1).

So every proposal, every release pull request and every hand-made pull
request shows as blocked, and the owner merges each with admin bypass.
Issue #50 reports that in the first live run (2026-11-25) this was #38, #39,
#40, #41, #42, #47 and #49, with #44 and #46 waiting the same way.

As the issue puts it, `tend/review` means "a head the bot may merge was
reviewed". tend merges only its own implementation pull requests, at `land`;
any other merge is the owner's, and the owner's merge is itself the approval.
The requirement guards the bot, and the rest pays for it.

## Intended outcome

When `tend propose` opens or finds an open proposal pull request, it sets
`tend/review` on the pull request's head: `success` when every path the head
changes lies in one change's Grove documents (`grove/changes/<id>-<slug>/`),
decided from Git alone, and no success otherwise. A new head pushed to a
proposal's branch, such as a person's revision, is checked on the next pass
and gets its own status. The owner then merges a proposal with the ordinary
Merge button. The ruleset is unchanged, and the bot still cannot merge an
implementation head that did not pass review.

## Scope

### In scope

- A `tend/review` status on the head of each open proposal pull request the
  publishing account opened from a `tend/issue-` or `tend/jira-` branch, set
  when intake opens it and on each later pass for a head that has none.
- The mechanical check of a head's changed paths against the target as
  fetched.

### Out of scope

- Letting the bot merge proposals or release pull requests (the issue's own
  out-of-scope line).
- Hand-made pull requests: they keep admin bypass.
- Release pull requests (release-please). The release workflow could post
  `tend/review` success on its own pull request, but that is an owner's edit
  to `.github/workflows/`, a protected path, outside tend's code. Noted here
  so it is not forgotten.
- Whether the bot may merge a proposal in the autonomous flow. That stays a
  step 4 decision (docs/work/261121-autonomous-direction/steps.md); this
  change only gives such a proposal's head a status.
- Any change to the ruleset.

## Decisions

None was made beyond the request (GitHub issue #50), which settles:

- The status is posted "when intake opens or finds a proposal pull request",
  `success` "when the head changes only that change's Grove documents
  (`grove/changes/<id>/`)", with a description such as "Grove documents only:
  merging approves this change".
- "The check is mechanical and reads Git alone. A head that touches anything
  else gets no success."
- "A new head on a proposal, such as a person's revision pushed to its
  branch, is checked again on the next pass and gets its own status."
- "The ruleset is unchanged."

## Open questions

1. **What the publishing account can now merge. Must be answered before
   review.** With this change a docs-only proposal head carries
   `tend/review` success, so the ruleset no longer stops the publishing
   account from merging it, and merging a proposal approves the change.
   tend's code never merges a proposal, but a drafter runs in auto mode as
   the OS user holding the account's token. Recommended: accept it. The
   parked ceiling in steps.md already says an agent holding that token could
   post its own `tend/review`, so the rule never stopped a determined agent;
   what this adds is that tend itself vouches for docs-only heads. Recording
   that in the ceiling is the owner's choice. Decides whether T-01 is built
   at all.
2. **What a head that touches anything else gets.** Recommended: a
   `failure` status whose description names a path outside the change's
   documents, so the owner sees why Merge is blocked; not leaving the head
   with no status, which looks the same as a pass not yet run. Can wait;
   decides T-01.
3. **Which paths count as "that change's Grove documents".** Recommended:
   every changed path lies under one folder directly below
   `<root>/changes/` (not `<root>/changes/archive/`), `<root>` being the
   Grove root the project's `grove.yaml` names; a head that changes nothing,
   touches two change folders, a cycle, or a maintained spec gets no
   success. Reason: the request names one change's folder and a check that
   reads Git alone, so the folder is read from the diff, not from what the
   intake folder recorded. Can wait; decides T-01.
4. **The guide.** guide/glossary.md says `tend/review` is set "on each
   reviewed commit", and guide/setup.md section 6 implies only reviewed heads
   pass the rule; both become incomplete. Recommended: the implementer
   updates them and records each with an `Also changed:` line, as Grove's
   rules allow for a file a task makes false. Can wait.
5. **Routine choices inside the approach.** Recommended: each pass checks
   the open proposals before taking new requests, whatever the cap, and the
   publish step checks a pull request it just opened; the head is the one
   GitHub reports for the pull request, fetched from origin; a head is
   checked once, skipped when the account already set `tend/review` on it;
   the status links the pull request; a failure to set it stops for a person
   (exit 3), as the other status writes do; test names and layout are the
   implementer's. Can wait; decides T-01.

## Related documents

- Requirements: [specs/intake/spec.md](specs/intake/spec.md)
- Tasks: [tasks.md](tasks.md)
