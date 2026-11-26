# Codex second opinion on thread 4

Asked 2026-11-26 through the Codex CLI 0.172.0, model `gpt-6.2-astra`,
reasoning effort high, read-only, thread
`01a219ca-895d-7851-b4fc-24de8b8c1e91`. It was given
[the research record](261126-design-review-research.md) with Claude's
three thread-4 proposals cut out, the owner's three objections, and both
codebases to read, and asked what should be fundamentally different in
how a request becomes approved, built, reviewed and landed, with every
existing constraint open. The reply is below as given, with local paths in its links rewritten; its link for the
Beads project was a placeholder. Section 19.9 of the research record
checks it.

## 1. The problem is when the owner has to exercise judgment

I think batching is the wrong frame. I would give Tend a bounded amount of time to find out what a request entails, and let the owner make a decision when there is evidence worth deciding on. For some requests, the first such evidence will be a working change. For others, an example of two plausible behaviors will be enough to settle the question before anyone builds either one.

The recorded PR data shows six requests produced 42 questions. The drafters marked 29 consequential, including two about which change would land first. On the build side, only one of five changes passed its first independent review. Three still failed after the allowed repair. #42 accumulated five failures and three owner change requests. #59 merged with a failure standing. Two approved changes waited 6 h 40 min and 3 h 36 min to be claimed, largely while earlier PRs remained open. Those are elapsed waits. The record does not measure hours of owner labor or establish that every review finding was sound. [Proposal questions and claim times](261126-design-review-research.md#L1056) · [Implementation reviews](261126-design-review-research.md#L1397)

Memex suggests what the owner values. They considered six changes together, kept six implementation PRs, and merged the stack in one six-second sitting. The timestamps establish when they merged, not when they read the code. My inference is that the owner wants to choose when to concentrate their attention and still be able to reject one piece independently. Neither a larger proposal nor a larger build delivers that property by itself. [Memex sequence](261126-design-review-research.md#L1044)

## 2. Give permission to investigate, then approve a concrete result

I would build around three units. A **request** records who wants what outcome and why. A **decision** records a choice only the owner can settle, the evidence for it, and the owner’s answer. A **candidate** is the smallest change the owner could usefully accept or decline on its own. One feature request can produce several candidates. One decision, such as whether keyborg may retain user input, can govern many requests. Tasks belong to the agent doing the work and can change as it learns.

The owner sets a short list of current priorities, a daily effort budget, and boundaries such as “ask before retaining user input or incurring a recurring cost.” Within those limits Tend can investigate, reproduce, prototype, or build on a branch. An outside issue supplies evidence. An owner-selected priority supplies permission to spend effort. When two plausible interpretations would produce materially different user experiences, the agent shows the difference and asks at the last point where the answer is still cheap to act on. A reversible implementation choice is the agent’s to make.

For a small fix like memex’s hand-off emitting a path its own CLI refuses, the owner reports the symptom once. The agent reproduces the failure, fixes it, and presents the same hand-off succeeding through the real CLI, with a regression test. The owner’s next touch is approval of that PR. For a feature like summarizing a recorded note, the agent might discover that a note without meeting turns admits two sensible outcomes. It brings the owner one real example and shows what each outcome would look like. The owner chooses, the agent records the answer, and the first candidate implements one usable interaction end to end. Both requests follow the same rule. The evidence determines when a question is worth asking. [Memex requests](261126-design-review-research.md#L390)

A fresh reviewer exercises the candidate against the request, the owner’s decisions, and the existing product promises. A reproducible failure goes back to the builder while it is making progress within budget. A repeated failure with no new evidence parks the attempt. A genuine product disagreement goes to the owner with the disputed example. Tend presents ready candidates and open decisions in one sitting, initially capped at two outstanding owner items. GitHub holds the attributable owner answers and PR approvals. Required CI and the owner’s approval authorize landing. Tend keeps attempt history in local SQLite on one persistent runner and reconciles it against GitHub after interruption. The runner has no owner credential. [Recorded authorization tests](261126-design-review-research.md#L892)

## 3. Retire the proposal merge and the advance reservation

I verified in code that Tend derives a proposal branch from each labelled issue, Grove refuses a claim when planned files overlap, and Tend’s own flow gives repair one turn. I would remove all three requirements from keyborg. The plans were actually quite good forecasts. Implementations touched 42 of 44 named file entries, but 17 of 59 actual touches were unlisted. That makes a file list useful to a builder and unreliable as an exclusive reservation. Tasks also occupied only 9–16% of proposal lines. Moving `tasks.md` later while keeping the owner’s reading assignment would save little. I would start with one active builder on keyborg and allow another when ready work is regularly waiting on build capacity. [Tend intake](../../intake.go#L80) · [Grove claim](https://github.com/mascah/grove/blob/3cfd2ef/internal/cli/claim.go#L223) · [Tend flow](../../tend/flow.yaml#L18) · [Plan measurements](261126-design-review-research.md#L1316)

I would ship one tool on keyborg and keep Grove’s most valuable idea inside the practice of building. A durable product promise deserves plain language, history, and, where practical, an executable example. It can land alongside the first code that honors it. An expensive or irreversible choice can still deserve a written decision before implementation. I would retain thread 1’s explicit decision records and separation of deciding from acting, thread 2’s headless persistent runner, and thread 3’s external merge authority. Those buy recovery, a simple runtime, and an approval the builder cannot grant itself. [Grove’s product boundaries](https://github.com/mascah/grove/blob/3cfd2ef/CONSTITUTION.md#L15)

## 4. SQLite now. Beads or Dolt when there is a concrete coordination problem

Gas City offers a pattern worth borrowing immediately. Its controller reconciles observed state, and an infrastructure retry consumes a different budget from a failed attempt. Gas Town illustrates the cost of distributing one job across many records and human surfaces. Neither reference establishes that this owner needs a distributed work store. I would try Beads if independent workers repeatedly duplicated work or lost dependencies, and its ready-work and claim operations could replace code Tend would otherwise own. I would choose Dolt specifically if independently edited work records needed versioned SQL history and branch-and-merge reconciliation across machines. Several agents on one laptop establish neither need. [Reference-system research](261126-design-review-research.md#L292) · [Beads project](/PLACEHOLDER) · [Dolt documentation](https://www.dolthub.com/docs/introduction/what-is-dolt/)

## 5. The strongest objection is that a working demo can anchor the owner

On a new product, the most important decision may be whether a feature belongs at all. A polished implementation makes “yes” easier, and its tests can faithfully encode the wrong premise. Grove’s early prose gives the owner a cheap opportunity to reject that premise. My recommendation depends on agents being able to surface meaningful alternatives before they become expensive. If the owner repeatedly discovers those alternatives only while reviewing finished code, I would restore an explicit early conversation for that class of decision.

## 6. Test the decision timing before building the tool

Run one real keyborg request by hand over two days. Pick the first interaction the owner actually wants to try. Have the owner privately write down what they care about and what would make them reject it. Give the agent the ordinary request, one hour, and the project boundaries. Allow one mid-course question with a concrete example. Have a second agent exercise the result, then give the owner the candidate through a normal PR. If first use exposes a small defect, run that through the same procedure.

Measure active owner minutes, separate sittings, rework after the owner sees the result, and which private acceptance points the reviewer missed. Compare the experience with the owner’s last Tend proposal and implementation. I would continue if the owner can decide from the example and candidate in one sitting, or two where an early answer changed the work. A shorter queue accompanied by a missed product constraint or a feature the owner feels pressured to accept would count against the design.