## T-01 — An agent's startup notice is skipped

Result: an agent that starts at a notice whose hint offers `esc skip` is sent
Esc alone, tend says so, and the start goes on as after a trust dialog
answered, as the terminals requirement "The trust dialogs are answered" in
this change says.

Files:
- `main.go`
- `testdata/bin/rex`
- `testdata/bin/turn.sh`
- `tend_test.go`
- `README.md`

Delegated: test layout and private helpers.

Constraints: the hooks review is recognised before a notice and still sends
no key; Esc is the only key sent at a notice.

Acceptance:
- The scenarios "A startup notice under rex" and "A screen tend does not
  know under rex" are tests that pass, the first named `rex-notice`.
- `rex-trust`, `rex-add-dir`, `rex-add-dir-refused` and `rex-hooks` still
  pass.
- The README's "What it does to your agents" says tend skips a startup
  notice that offers to be skipped, by Esc alone, and stops at any other
  screen.
