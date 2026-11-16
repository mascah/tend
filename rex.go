package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// rex: a session per worktree, first named after its folder, and an agent per
// block, labelled with its name and running it as the block's program, so the
// name goes when the agent exits. Agent state is what the agent itself reports
// with the program status protocol (OSC 7501), which Claude Code and Codex
// send: idle, working, done, blocked or error. A running agent with no record
// is at something it does not report, such as a dialog, and is unknown.
// ponytail: polls program_status; stream rex events if polling ever costs.
type rex struct{ session, dir string }

func rexCmd(args ...string) (string, error) {
	return quiet("rex", append([]string{"--autostart=false"}, args...)...)
}

// git makes the worktree, under the state folder, then it is opened.
func (r *rex) worktree(b, base, _ string) string {
	d := state + "/worktrees/" + filepath.Base(repo) + "-" + strings.ReplaceAll(b, "/", "-")
	must(out("git", "-C", repo, "worktree", "add", "-q", "-b", b, d, base))
	r.open(d)
	return d
}

// The session recorded for the worktree in $state/rex-sessions ("<session-id>
// <path>" a line) while rex still has it, or a new one named after the
// folder. Agents in rex may rename their session, so the name finds nothing.
func (r *rex) open(d string) {
	r.dir, r.session = d, ""
	s, err := rexCmd("ls", "--json")
	if err != nil {
		stop("rex lists no sessions; is its server running?")
	}
	live := map[string]bool{}
	for _, x := range list(get(parse(s), "sessions")) {
		live[str(get(x, "session_id"))] = true
	}
	for _, line := range strings.Split(cat(state+"/rex-sessions"), "\n") {
		if id, path, _ := strings.Cut(line, " "); live[id] && sameDir(path, d) {
			r.session = id
		}
	}
	if r.session != "" {
		return
	}
	s, err = rexCmd("new", filepath.Base(d), "-c", d, "--json")
	if err != nil {
		stop("rex could not open a session for %s", d)
	}
	r.session = str(get(parse(s), "session_id"))
	f, err := os.OpenFile(state+"/rex-sessions", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err == nil {
		fmt.Fprintf(f, "%s %s\n", r.session, d)
		f.Close()
	}
}

func sameDir(a, b string) bool {
	x, err1 := os.Stat(a)
	y, err2 := os.Stat(b)
	return err1 == nil && err2 == nil && os.SameFile(x, y)
}

func (r *rex) blocks() []any {
	s, _ := rexCmd("block", "ls", "-s", r.session, "--json")
	return list(get(parse(s), "blocks"))
}

func (r *rex) pane(name string) string {
	for _, b := range r.blocks() {
		if str(get(b, "label")) == name {
			return str(get(b, "block_id"))
		}
	}
	return ""
}

// The root record, or nil when the program reports none.
func (r *rex) record(b string) any {
	s, err := rexCmd("block", "call", "-s", r.session, "-b", b, "program_status")
	if err != nil {
		return nil
	}
	for _, x := range list(get(parse(s), "records")) {
		if str(get(x, "id")) == "" {
			return x
		}
	}
	return nil
}

func state7501(rec any) string {
	switch st := str(get(rec, "state")); st {
	case "":
		return "unknown"
	case "working", "blocked", "done", "idle":
		return st
	case "error":
		return "unknown" // a person's
	default:
		return "idle" // as the protocol says of a state it does not know
	}
}

func (r *rex) status(name string) string {
	b := r.pane(name)
	if b == "" {
		return ""
	}
	return state7501(r.record(b))
}

// Every agent splits from the session's first block; fresh does not matter,
// as no agent ever takes the shell rex opened the session with.
func (r *rex) start(name, kind string, _ bool, args []string) (string, error) {
	bs := r.blocks()
	if len(bs) == 0 {
		stop("rex session %s has no block to split from", r.session)
	}
	s, err := rexCmd(append([]string{"split", "-s", r.session, "-b", str(get(bs[0], "block_id")), "--split=right", "-c", r.dir,
		"--label", name, "--focus=false", "--json", "--", kind}, args...)...)
	if err != nil {
		return "", errors.New("split_failed")
	}
	b := str(get(parse(s), "block_ids", 0))
	for deadline := time.Now().Add(60 * time.Second); time.Now().Before(deadline); time.Sleep(500 * time.Millisecond) {
		switch state7501(r.record(b)) {
		case "idle", "done":
			return b, nil
		}
		if strings.Contains(strings.ToLower(r.read(name)), "trust") {
			return b, errors.New("agent_not_ready")
		}
	}
	return b, errors.New("agent_not_ready")
}

func (r *rex) read(name string) string {
	s, _ := rexCmd("capture", "-s", r.session, "-b", r.pane(name), "--trim")
	lines := strings.Split(s, "\n")
	return strings.Join(lines[max(0, len(lines)-40):], "\n")
}

func (r *rex) keys(name string, keys ...string) {
	rexCmd(append([]string{"send-key", "-s", r.session, "-b", r.pane(name)}, keys...)...)
}

// The text goes as a bracketed paste when the agent asks for one, so its
// newlines stay text, then Enter submits it. A prompt counts as delivered once
// the agent reports something new. Codex reports nothing while it works, only
// done at the end of its turn, so the first report may come with the answer:
// it is waited for as long as the answer is, then the prompt is stalled.
func (r *rex) prompt(name, text string) (string, string) {
	b := r.pane(name)
	if b == "" {
		return "", "agent_not_found"
	}
	before := str(get(r.record(b), "updated_at"))
	if m, _ := rexCmd("block", "call", "-s", r.session, "-b", b, "modes"); slices.ContainsFunc(list(get(parse(m), "modes")), func(x any) bool {
		return str(get(x, "name")) == "bracketed_paste" && get(x, "set") == true
	}) {
		text = "\x1b[200~" + text + "\x1b[201~"
	}
	if _, err := rexCmd("send", "-s", r.session, "-b", b, text); err != nil {
		return "", "send_failed"
	}
	if _, err := rexCmd("send-key", "-s", r.session, "-b", b, "enter"); err != nil {
		return "", "send_failed"
	}
	n, _ := strconv.Atoi(timeout())
	for deadline := time.Now().Add(time.Duration(n) * time.Millisecond); ; time.Sleep(250 * time.Millisecond) {
		if rec := r.record(b); str(get(rec, "updated_at")) != before && state7501(rec) != "unknown" {
			break
		}
		if time.Now().After(deadline) {
			return "", "agent_prompt_stalled"
		}
	}
	return r.wait(name, timeout())
}

// Until the agent settles idle, done or blocked, or one of until.
func (r *rex) wait(name, ms string, until ...string) (string, string) {
	if len(until) == 0 {
		until = []string{"idle", "done", "blocked"}
	}
	n, _ := strconv.Atoi(ms)
	for deadline := time.Now().Add(time.Duration(n) * time.Millisecond); ; time.Sleep(500 * time.Millisecond) {
		b := r.pane(name)
		if b == "" {
			return "", "agent_not_found"
		}
		if st := state7501(r.record(b)); slices.Contains(until, st) {
			return st, ""
		}
		if time.Now().After(deadline) {
			return "", "timeout"
		}
	}
}

func (r *rex) closePane(b string) { rexCmd("block", "close", "-s", r.session, b) }

// Every session recorded for the worktree is killed and forgotten.
func (r *rex) closePlace(d string) {
	var keep []string
	for _, line := range strings.Split(cat(state+"/rex-sessions"), "\n") {
		if id, path, _ := strings.Cut(line, " "); line != "" && sameDir(path, d) {
			rexCmd("kill", id)
		} else if line != "" {
			keep = append(keep, line+"\n")
		}
	}
	os.WriteFile(state+"/rex-sessions", []byte(strings.Join(keep, "")), 0o644)
}

// tend reports its own state the same way, to the terminal it runs in, so its
// tab shows whether it needs you or is done.
func (r *rex) notify(title, body, sound string) {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return
	}
	defer tty.Close()
	st := "done"
	if sound == "request" {
		st = "blocked"
	}
	msg := strings.Map(func(c rune) rune {
		if c < 0x20 || c == 0x7f {
			return ' '
		}
		return c
	}, title+": "+body)
	if len(msg) > 2000 {
		msg = msg[:2000]
	}
	fmt.Fprintf(tty, "\x1b]7501;state=%s:app=tend:msg=%s\x1b\\", st, base64.StdEncoding.EncodeToString([]byte(msg)))
}
