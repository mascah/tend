# Changes to publication

The changes that change 261122-hjhmn proposes to the requirements of
`publication`, the maintained spec at
[grove/specs/publication/spec.md](../../../../specs/publication/spec.md). This
is future work: nothing here describes the target branch until the change is
implemented and the maintained spec is reconciled.

## MODIFIED Requirements

### Requirement: A request answered with no change

If the implementer commits nothing for a requested change but writes its
reply, then the system shall post the reply once as a comment on the pull
request by the account, saying tend will not ask again, record
`answered-<review node id>`, review nothing again and exit 0; if it neither
commits nor replies and is there and not working, the system shall stop for a
person on every run until a person commits or deletes `request-<review id>`,
which has the request asked once more; and if it neither commits nor replies
and is gone, the system shall ask the request once more of a new session, as
the flows spec's "Any step recovers without being prompted twice" says.

#### Scenario: Out of scope

- GIVEN changes requested on the head, and an implementer that judges them out
  of scope and commits nothing
- WHEN the owner runs tend, then runs it again
- THEN both exit 0, the pull request has one comment by the account with the
  implementer's reply, the request was asked once and the head was reviewed
  once

#### Scenario: A request prompt that stalled

- GIVEN tend stopped because the request's prompt stalled, and the implementer
  committed nothing and is idle
- WHEN the owner runs tend again
- THEN it exits 3, saying the implementer committed nothing and is not working
  on them, without asking again; once a person deletes `request-<review id>`,
  the next run asks a second time and passes

#### Scenario: The implementer is gone

- GIVEN tend was killed after asking a requested change, the implementer
  committed nothing and wrote no reply, and its pane is closed
- WHEN the owner runs tend again
- THEN the request is asked once more of a new implementer session, and the
  new head is reviewed and published
