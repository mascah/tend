package main

import (
	"cmp"
	"errors"
	"os"
	"path/filepath"
	"regexp"
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

// ---- telling a person

// What a notification is about, as its title names it, and the link it ends
// with: the pull request, else the request the work came from, else the work
// item by its ID with no link.
func subject() (ref, link string) {
	if pr != "" {
		ref = refOf(pr, func() string {
			if n := prNumber(); regexp.MustCompile(`^[0-9]+$`).MatchString(n) {
				return n
			}
			n, _ := quiet("gh", "pr", "view", pr, "--json", "number", "--jq", ".number")
			return n
		})
		return cmp.Or(ref, pr), pr
	}
	u := str(get(issue, "url"))
	switch {
	case item.from != "" && !strings.HasPrefix(item.from, "#"):
		return item.from, u // a Jira key
	case u != "" || item.key != "":
		if ref = cmp.Or(refOf(u, func() string { return item.key }), u); ref != "" {
			return ref, u
		}
	}
	return cmp.Or(cycle, "tend"), ""
}

var ghURL = regexp.MustCompile(`^https://github\.com/([^/]+/[^/]+)/(?:pull|issues)/([0-9]+)$`)

// A GitHub pull request or issue as <owner>/<repo>#<n>, from its URL, else
// from the repository gh sees and the number num gives; "" when neither says.
var nwo string

func refOf(url string, num func() string) string {
	if m := ghURL.FindStringSubmatch(url); m != nil {
		return m[1] + "#" + m[2]
	}
	n := num()
	if n == "" {
		return ""
	}
	if nwo == "" {
		nwo, _ = quiet("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	}
	if nwo == "" {
		return ""
	}
	return nwo + "#" + n
}

// The notification of a wait: titled with what it is about and that it waits
// for the person, the body text then the link.
func waits(text string) (string, string) {
	ref, link := subject()
	if link != "" {
		text = strings.ReplaceAll(text, link, ref) // the link comes once, at the end
	}
	return fit(titled(ref, " waits for you"), text, link)
}

// A title of at most 60 characters that keeps what it says of the work,
// suffix, whole: a long ref loses its start, so its number stays.
func titled(ref, suffix string) string {
	ref = strings.Join(strings.Fields(ref), " ")
	room := 60 - len("tend: ") - len([]rune(suffix))
	if r := []rune(ref); len(r) > room {
		ref = "…" + string(r[len(r)-room+1:])
	}
	return "tend: " + ref + suffix
}

// A title of at most 60 characters and a body of at most 240, the link whole
// on the body's last line, the text before it cut with an ellipsis.
func fit(title, text, link string) (string, string) {
	title = cut(strings.Join(strings.Fields(title), " "), 60)
	text = strings.Join(strings.Fields(text), " ")
	if link == "" {
		return title, cut(text, 240)
	}
	room := 240 - len([]rune(link)) - 1
	if room < 2 || text == "" {
		return title, link
	}
	return title, cut(text, room) + "\n" + link
}

func cut(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimRight(string(r[:n-1]), " ") + "…"
}

// Tell the person, once for each wait: a notification the same as the last
// one recorded in file rec is not sent again. With no file it is always sent.
func tell(rec, title, body, sound string) {
	text := title + "\n" + body + "\n"
	if rec != "" && cat(rec)+"\n" == text {
		say("already told: %s", title)
		return
	}
	term.notify(title, body, sound)
	if rec != "" && os.MkdirAll(filepath.Dir(rec), 0o755) == nil {
		write(rec+".tmp", text)
		os.Rename(rec+".tmp", rec)
	}
}

// The file that records what tend last told about this work: in the
// attempt's run folder, or, before there is one, under TEND_STATE/notified,
// named as the work item's lock is.
func toldIn() string {
	if run != "" {
		return run + "/notified"
	}
	if state != "" && repo != "" && cycle != "" {
		return state + "/notified/" + filepath.Base(lockFile(repo, cycle))
	}
	return ""
}
