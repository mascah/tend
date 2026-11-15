package main

import (
	"errors"
	"time"
)

// The terminal the agents run in. A place is the workspace of one worktree,
// where its agents' panes open; an agent is found by its name.
type terminal interface {
	worktree(branch, base, label string) string // a new worktree from base, opened as the place: its path
	open(dir string)                            // an existing worktree, opened as the place
	pane(name string) string                    // the pane holding the agent, or ""
	// The agent started in a free pane of the place, or in a fresh one: its
	// pane, or why not (a dialog may be waiting; see start).
	start(name, kind string, fresh bool, args []string) (string, error)
	read(name string) string // its visible screen
	keys(name string, keys ...string)
	status(name string) string                              // "" when there is no such agent
	prompt(name, text string) (string, string)              // the state it settled in, or "" and why
	wait(name, ms string, until ...string) (string, string) // likewise
	closePane(pane string)
	notify(title, body, sound string)
}

// Herdr: agents by name through its CLI, worktrees as its workspaces.
type herdr struct{ ws, dir string }

func (h *herdr) worktree(b, base, label string) string {
	wt := parse(must(out("herdr", "worktree", "create", "--cwd", repo, "--branch", b, "--base", base, "--label", label, "--no-focus")))
	h.dir, h.ws = str(get(wt, "result", "worktree", "path")), str(get(wt, "result", "workspace", "workspace_id"))
	return h.dir
}

func (h *herdr) open(path string) {
	h.dir, h.ws = path, ""
	s := must(out("herdr", "worktree", "list", "--cwd", repo))
	for _, w := range list(get(parse(s), "result", "worktrees")) {
		if str(get(w, "path")) == path && h.ws == "" {
			h.ws = str(get(w, "open_workspace_id"))
		}
	}
	if h.ws == "" {
		s = must(out("herdr", "worktree", "open", "--cwd", repo, "--path", path, "--no-focus"))
		h.ws = str(get(parse(s), "result", "workspace", "workspace_id"))
	}
}

func (h *herdr) pane(name string) string {
	s, err := quiet("herdr", "agent", "get", name)
	if err != nil {
		return ""
	}
	return str(get(parse(s), "result", "agent", "pane_id"))
}

// A new pane, split from the workspace's first.
func (h *herdr) split() string {
	panes := must(out("herdr", "pane", "list", "--workspace", h.ws))
	s := must(out("herdr", "pane", "split", str(get(parse(panes), "result", "panes", 0, "pane_id")), "--direction", "right", "--cwd", h.dir, "--no-focus"))
	return str(get(parse(s), "result", "pane", "pane_id"))
}

// A pane just split is not yet at its prompt, so a busy pane is retried.
func (h *herdr) start(name, kind string, fresh bool, args []string) (string, error) {
	p := ""
	if !fresh {
		for _, x := range list(get(parse(must(out("herdr", "pane", "list", "--workspace", h.ws))), "result", "panes")) {
			if get(x, "agent") == nil {
				p = str(get(x, "pane_id"))
				break
			}
		}
	}
	if p == "" {
		p = h.split()
	}
	var o string
	for range 10 {
		var err error
		o, err = both("herdr", append([]string{"agent", "start", name, "--kind", kind, "--pane", p, "--timeout", "60000", "--"}, args...)...)
		if err == nil {
			return p, nil
		}
		if str(get(parse(o), "error", "code")) != "agent_pane_busy" {
			break
		}
		time.Sleep(time.Second)
	}
	return p, errors.New(str(get(parse(o), "error", "code")))
}

func (h *herdr) read(name string) string {
	s, _ := quiet("herdr", "agent", "read", name, "--source", "visible", "--lines", "40")
	return s
}

func (h *herdr) keys(name string, keys ...string) {
	quiet("herdr", append([]string{"agent", "send-keys", name}, keys...)...)
}

func (h *herdr) status(name string) string {
	s, _ := quiet("herdr", "agent", "get", name)
	return str(get(parse(s), "result", "agent", "agent_status"))
}

// Herdr's answer: the state the agent settled in, or its error code.
func settledAs(o string, _ error) (string, string) {
	if st := str(get(parse(o), "result", "agent", "agent_status")); st != "" {
		return st, ""
	}
	code := str(get(parse(o), "error", "code"))
	if code == "" {
		code = "no answer"
	}
	return "", code
}

func (h *herdr) prompt(name, text string) (string, string) {
	return settledAs(both("herdr", "agent", "prompt", name, text, "--wait", "--timeout", timeout()))
}

func (h *herdr) wait(name, ms string, until ...string) (string, string) {
	args := []string{"agent", "wait", name}
	for _, u := range until {
		args = append(args, "--until", u)
	}
	o, err := both("herdr", append(args, "--timeout", ms)...)
	if err != nil && len(until) > 0 {
		return "", "did not settle"
	}
	return settledAs(o, err)
}

func (h *herdr) closePane(pane string) { quiet("herdr", "pane", "close", pane) }

func (h *herdr) notify(title, body, sound string) {
	quiet("herdr", "notification", "show", title, "--body", body, "--sound", sound)
}
