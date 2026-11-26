# Changes to review

<!-- The changes that change 261126-5xkrh proposes to the requirements of
`review`. This is future work: nothing here describes the target branch
until the change is implemented and the maintained spec is reconciled. -->

Maintained spec: [review](../../../../specs/review/spec.md), as at e80da30.

## ADDED Requirements

### Requirement: Fail a head only for findings that matter

The system's built-in review prompt shall ask the reviewer to conclude
`failure` only for a defect a realistic use would hit, a requirement missed
in its main case, a security or data hole that can be reached, or a claim in
the implementer's notes the code does not support; to write every other
finding, such as a corner that needs contrived setup to reach, which of two
reasons a stop gives, wording or polish, in the verdict's `notes` and
conclude `success`; and to judge `actionable` on the findings that fail the
head alone; and when a verdict concludes `success`, the system shall go on
as from a passing review whatever its `notes` hold, passing no note to the
repair.

#### Scenario: The bar in the prompt

- WHEN tend prompts a reviewer with its built-in review prompt
- THEN the prompt names the four kinds of finding that fail a head, says
  that contrived corners, the reason a stop gives, wording and polish do
  not, and asks for those in `notes`

#### Scenario: A passing head with notes

- GIVEN a reviewer that writes a verdict concluding `success` with `notes`
  naming a minor finding
- WHEN the owner runs tend on the cycle
- THEN no repair is asked, `status` says `done: success`, and tend exits 0

#### Scenario: A failure with notes

- GIVEN a reviewer that fails the first head as actionable, with `notes`
  naming a minor finding, and passes the next
- WHEN tend asks the repair
- THEN the repair prompt holds the verdict's `summary` and not its `notes`

#### Scenario: A project's own bar

- GIVEN a project whose flow names `prompts/review.md` and whose
  `tend/prompts/review.md` on its target branch states another bar
- WHEN tend has a head reviewed
- THEN the reviewer is prompted with the project's text, not the built-in
  bar

### Requirement: A later review judges what changed

When tend prompts a review of a head and the attempt holds a valid verdict
for an earlier commit of the branch, the system shall fill `{earlier}` with
each such verdict's head, conclusion and file, and the built-in review prompt
shall ask the reviewer to judge whether the findings that failed an earlier
head are fixed and what the commits since the latest earlier head changed,
and not to fail the head for code unchanged since a head an earlier review
saw, unless it is a security or data hole that can be reached; and when the
attempt holds no such verdict, `{earlier}` shall say there is none.

#### Scenario: The review after a repair

- GIVEN a reviewer that fails the first head as actionable, and a repair
  that commits
- WHEN tend prompts the review of the new head
- THEN the prompt names the first head, its conclusion `failure` and its
  verdict file, and asks for the commits since that head to be judged

#### Scenario: The first review

- WHEN tend prompts the first review of an attempt
- THEN the prompt says there is no earlier verdict

#### Scenario: A verdict for a commit no longer on the branch

- GIVEN an attempt whose run folder holds a verdict for a commit the branch
  no longer contains
- WHEN tend prompts the review of the head
- THEN that verdict is not named in the prompt

## MODIFIED Requirements

### Requirement: A fresh reviewer for each head

When an attempt's head has no valid verdict and is not carried from a head
that passed review, as the landing spec's "The verdict carries across a clean
merge" says, the system shall start reviewer `rev-<short head>` as Codex with
`-s workspace-write -c sandbox_workspace_write.network_access=true -a never
--add-dir <run folder>`, reusing one an earlier run left, and prompt it once
with the branch, the head, the change as `git diff <base>...HEAD`, what the
work item must do, where the implementer's claims are and the attempt's
earlier verdicts, asking it to fail the head only as "Fail a head only for
findings that matter" says, for no edits or commits, and for the verdict
written to `verdict-<head>.json` in the run folder as `head`, `conclusion`,
`actionable`, `summary` and `notes`; and once the verdict is valid it shall
close the reviewer's pane.

#### Scenario: The reviewer's one writable folder

- WHEN tend has a cycle's head reviewed
- THEN the reviewer is started with `--add-dir TEND_STATE/runs/<attempt>`,
  and the verdict file is there

#### Scenario: A reviewer left working

- GIVEN tend stopped because the reviewer's prompt timed out, and the reviewer
  is still working
- WHEN the owner runs tend again
- THEN the reviewer is started once and prompted once in all

#### Scenario: A carried head

- GIVEN an attempt whose head is a clean merge of the target into a head
  that passed review, made at `land`
- WHEN the owner runs tend again
- THEN no reviewer is started for the merge commit, and the flow goes on as
  from a passing review

### Requirement: Only a valid verdict counts

The system shall count a verdict, new or cached, only when its file is valid
JSON naming the head it is for, with `conclusion` `success` or `failure`,
`actionable` a boolean, `summary` a string, and `notes`, where present, a
string; if a cached verdict is not valid, it shall keep it as
`verdict-<head>.json.rejected-<time>` and review the head again; and if the
reviewer settles without writing a valid verdict, it shall stop for a person,
naming the reviewer's pane.

#### Scenario: A cached verdict for another head

- GIVEN a reviewed attempt whose verdict file now names head `0000…0000`
- WHEN the owner runs tend again
- THEN the head is reviewed a second time and the verdict names the head

#### Scenario: A cached verdict without a conclusion

- GIVEN a reviewed attempt whose verdict file has lost its `conclusion`
- WHEN the owner runs tend again
- THEN the head is reviewed again and the set-aside file is kept as
  `verdict-<head>.json.rejected-<time>`

#### Scenario: A new verdict without a conclusion

- GIVEN a reviewer that writes a verdict with no `conclusion`
- WHEN the owner runs tend
- THEN tend exits 3, saying the reviewer settled without a valid verdict

#### Scenario: A verdict without notes

- GIVEN a reviewer that writes a passing verdict with no `notes`
- WHEN the owner runs tend
- THEN the verdict counts, and tend exits 0

#### Scenario: Notes that are not text

- GIVEN a reviewer that writes a verdict whose `notes` is a list
- WHEN the owner runs tend
- THEN tend exits 3, saying the reviewer settled without a valid verdict
