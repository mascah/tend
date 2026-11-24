package main

import (
	"errors"
	"os"
	"strings"
	"time"
)

// The terminal the agents run in. A place is the workspace of one worktree,
// where its agents' panes open; an agent is found by its name.
type terminal interface {
	worktree(branch, base, label string) string // a new worktree from base, opened as the place: its path
	open(dir string)                            // an existing worktree, opened as the place
	pane(name string) string                    // the pane holding the agent, or ""
	gone(name string) bool                      // the terminal answered that it holds no such agent; a failed call is not gone
	// The agent started in a free pane of the place, or in a fresh one: its
	// pane, or why not (a dialog may be waiting; see start).
	start(name, kind string, fresh bool, args []string) (string, error)
	read(name string) string // its visible screen
	keys(name string, keys ...string)
	status(name string) string                              // "" when there is no such agent
	prompt(name, text string) (string, string)              // the state it settled in, or "" and why
	wait(name, ms string, until ...string) (string, string) // likewise
	closePane(pane string)
	closePlace(dir string) // the worktree's place closed, with its agents; Git removes what is left
	notify(title, body, sound string)
}

// ---- agents

// Start an agent. A folder the harness has never seen gets its trust dialog,
// Claude's per worktree and Codex's per repository, answered here, as is
// Claude's dialog for adding tend's run folder with --add-dir. One may follow
// another, so each is answered in turn. Claude refuses keys that reach a dialog
// before it takes focus, as typeahead, and puts it back at "No"; so tend pauses
// before answering, and a dialog still on screen after an answer, even behind
// an agent called ready, is answered again. Codex's review of new or changed
// hooks is never answered: a hook trusted runs outside the sandbox, and one in
// the candidate is the implementer's to write. A notice whose hint offers
// "esc skip", as Codex's "Update available", is skipped by Esc alone, which can
// choose nothing: never an update, never a model.
// ponytail: trusts whatever repo the caller passed; the real tool should
// answer only for repos the owner registered.
var dialogs = []string{"Hooks can run outside the sandbox", "Yes, I trust this folder", "Yes, add it", "Trust and continue", "esc skip"}

// The first of the dialogs tend knows that is on the agent's screen, or "".
func dialogOn(name string) (string, string) {
	screen := strings.Join(strings.Fields(term.read(name)), " ") // a hint or warning may wrap
	for _, d := range dialogs {
		if strings.Contains(screen, d) {
			return d, screen
		}
	}
	return "", screen
}

func start(name, kind string, fresh bool, args ...string) string {
	p, err := term.start(name, kind, fresh, args)
	for range 4 {
		answered, screen := dialogOn(name)
		switch {
		case strings.Contains(screen, "Hooks can run outside the sandbox"):
			stop("%s asks to review new or changed hooks, which run outside its sandbox once trusted; tend trusts none. Answer it in its pane, then run tend again", name)
		case strings.Contains(screen, "Yes, I trust this folder"), strings.Contains(screen, "Yes, add it"):
			time.Sleep(time.Second)
			term.keys(name, "down", "enter")
		case strings.Contains(screen, "Trust and continue"):
			time.Sleep(time.Second)
			term.keys(name, "enter")
		case strings.Contains(screen, "esc skip"):
			time.Sleep(time.Second)
			term.keys(name, "esc")
			say("skipped %s's startup notice", name)
		case err == nil:
			return p
		default:
			code := err.Error()
			if code == "" {
				code = "unknown"
			}
			stop("%s did not start: %s", name, code)
		}
		// Codex comes out of its dialog as done, not idle. Another dialog may
		// follow at once, which no ready state comes before: it is answered
		// without waiting for one.
		err = errors.New("timeout")
		for deadline := time.Now().Add(time.Minute); time.Now().Before(deadline); time.Sleep(500 * time.Millisecond) {
			if st := term.status(name); st == "idle" || st == "done" {
				err = nil
				break
			}
			if d, _ := dialogOn(name); d != "" && d != answered {
				err = nil
				break
			}
		}
	}
	stop("%s did not start: its startup dialog would not take tend's answer", name)
	return p
}

// Carry on only if a prompt or wait settled the agent idle or done.
func settled(name, st, why string) {
	switch st {
	case "idle", "done":
	case "":
		stop("%s: %s", name, why)
	default:
		stop("%s is %s", name, st)
	}
}

func await(name string) {
	st, why := term.wait(name, timeout())
	settled(name, st, why)
}

// Prompt an agent and wait for it to settle. An agent still working, as after
// tend was killed, is waited for and not prompted again. Blocked or unknown is
// a person's; tend answers no dialog but those in start. A timeout or
// stall does not prove a prompt undelivered, so nothing here sends one twice.
// mark, if given, records HEAD just before the prompt is sent.
func ask(name, text, mark string) {
	switch st := term.status(name); st {
	case "working":
		say("%s is still working on an earlier prompt; waiting for it, not prompting again", name)
		await(name)
		return
	case "blocked", "unknown":
		stop("%s is %s; look at it, then run tend again", name, st)
	}
	if mark != "" {
		write(mark+".tmp", git("rev-parse", "HEAD")+"\n")
		must("", os.Rename(mark+".tmp", mark))
	}
	if strings.HasPrefix(name, "impl-") {
		write(run+"/layout", layout()+"\n") // for contained()
	}
	st, why := term.prompt(name, text)
	settled(name, st, why)
}
