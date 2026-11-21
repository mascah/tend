# Changes to terminals

The changes that change 261121-m9bxm proposes to the requirements of
`terminals`, the maintained spec at
[grove/specs/terminals/spec.md](../../../../specs/terminals/spec.md). This is
future work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## MODIFIED Requirements

### Requirement: The trust dialogs are answered

If an agent starts at the dialog asking whether to trust its folder, at
Claude Code's dialog asking whether to add tend's run folder with
`--add-dir`, or at a notice whose hint offers `esc skip`, such as Codex's
"Update available", then the system shall answer it, by choosing to trust or
add the folder, Down then Enter at "Yes, I trust this folder" or "Yes, add
it", or Enter at "Trust and continue", or by pressing Esc alone at the notice
and saying on standard error that it skipped that agent's startup notice, and
wait for the agent to be ready, answering each in turn when one follows
another; it shall pause before answering, as
Claude Code refuses keys that reach a dialog before it takes focus, and answer
again a dialog still on screen after an answer, even when the terminal calls
the agent ready; and if an agent fails to start for any other reason, the
system shall stop for a person, saying it did not start.

If an agent starts at Codex's dialog asking to review new or changed hooks,
then the system shall stop for a person, saying so, and send it no key, even
when the terminal calls the agent ready or the screen offers `esc skip`: a trusted hook runs outside the
sandbox, and a hook in the candidate is the implementer's to write.

#### Scenario: A new folder under rex

- GIVEN under rex agents that start at the trust dialog
- WHEN tend takes a cycle to a verdict
- THEN rex is sent the keys `down enter`, and tend exits 0 with a passing
  verdict

#### Scenario: Adding the run folder under rex

- GIVEN under rex agents that start at the dialog for adding a directory
- WHEN tend takes a cycle to a verdict
- THEN rex is sent the keys `down enter`, and tend exits 0 with a passing
  verdict

#### Scenario: A dialog that refuses the first answer under rex

- GIVEN under rex agents at the dialog for adding a directory, which refuses
  the first answer and is called idle behind it
- WHEN tend takes a cycle to a verdict
- THEN rex is sent `down enter` twice for an agent, and tend exits 0 with a
  passing verdict

#### Scenario: New hooks under rex

- GIVEN under rex an agent that starts at the dialog asking to review hooks
- WHEN tend starts it
- THEN tend exits 3 saying it asks to review new or changed hooks, and rex
  is sent no key

#### Scenario: A startup notice under rex

- GIVEN under rex agents that start at Codex's "Update available" notice,
  whose hint is "enter continue · esc skip"
- WHEN tend takes a cycle to a verdict
- THEN rex is sent the key `esc` for each agent and never `enter` at the
  notice, tend says it skipped each agent's startup notice, and it exits 0
  with a passing verdict

#### Scenario: A screen tend does not know under rex

- GIVEN under rex an agent that starts at a screen that is neither a trust
  dialog nor a notice offering `esc skip`
- WHEN tend starts it
- THEN tend exits 3 saying the agent did not start, and rex is sent no key
