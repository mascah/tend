// tend takes one agreed work item through coding agents and brings back a
// reviewed branch: a Grove cycle or a change in no cycle, or a GitHub issue of
// a repository without Grove. Claude implements in a Herdr pane, Codex reviews the head, one repair
// if the review asks, and with a listed origin a pull request as a bot account.
//
//	tend <repo> <cycle-id> | <change-id>
//	tend <repo> issue-<n>
//	tend status <repo> [<cycle-id> | <change-id>]
//	tend watch [<repo>...]
//	tend propose <repo>
//	tend flow <repo>
//	tend version
//
// Exit 0: the head is reviewed and the verdict printed. Exit 3: a person is
// needed, and a Herdr notification says why. Exit 4: another tend is on this
// work item. Progress is read again on every run from Git, Grove, GitHub, Herdr
// and the run folders under TEND_STATE (default ~/.local/state/tend), never
// from the candidate. After exit 3, or a kill, the same line carries on: an
// agent still working is waited for, never prompted again, and the one repair
// is used once asked.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	mode, state, repo, cycle, run       string
	kind, target, base, branch, id, dir string
	impl, root, what, notes, trailer, v string
	head, pr, account, where            string
	issue, rec                          any
	item                                request  // the work item, when not a Grove cycle
	lock                                *os.File // held open: its flock is the work item's
	setup                               *os.File // held while this tend sets up
	term                                terminal
)

// ---- commands

func sh(in *string, errw io.Writer, both bool, name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	var out bytes.Buffer
	c.Stdout = &out
	c.Stderr = errw
	if both {
		c.Stderr = &out
	}
	if in != nil {
		c.Stdin = strings.NewReader(*in)
	}
	err := c.Run()
	return strings.TrimRight(out.String(), "\n"), err
}

// out runs a command as $(...) does: stdout without trailing newlines, stderr
// passed through. quiet discards stderr, both captures it with stdout.
func out(name string, args ...string) (string, error) {
	return sh(nil, os.Stderr, false, name, args...)
}
func quiet(name string, args ...string) (string, error) {
	return sh(nil, io.Discard, false, name, args...)
}
func both(name string, args ...string) (string, error) { return sh(nil, nil, true, name, args...) }
func input(in, name string, args ...string) (string, error) {
	return sh(&in, os.Stderr, false, name, args...)
}

// must is set -e: an unexpected failure ends tend.
func must[T any](x T, err error) T {
	if err != nil {
		fmt.Fprintf(os.Stderr, "tend: %v\n", err)
		os.Exit(1)
	}
	return x
}

func ok(name string, args ...string) bool { _, err := quiet(name, args...); return err == nil }
func git(args ...string) string           { return must(out("git", args...)) }
func gitQ(args ...string) string          { s, _ := quiet("git", args...); return s }

func cd(d string) {
	must("", os.Chdir(d))
	os.Setenv("PWD", d)
}

// ---- JSON, read as jq -r reads it

func parse(s string) any {
	d := json.NewDecoder(strings.NewReader(s))
	d.UseNumber()
	var x any
	if d.Decode(&x) != nil {
		return nil
	}
	return x
}

func get(x any, path ...any) any {
	for _, p := range path {
		switch k := p.(type) {
		case string:
			m, _ := x.(map[string]any)
			x = m[k]
		case int:
			a, _ := x.([]any)
			if k >= len(a) {
				return nil
			}
			x = a[k]
		}
	}
	return x
}

func str(x any) string { // jq -r, with null as nothing
	switch t := x.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	case bool:
		return strconv.FormatBool(t)
	}
	b, _ := json.Marshal(x)
	return string(b)
}

func interp(x any) string { // jq's \(...)
	if x == nil {
		return "null"
	}
	return str(x)
}

func list(x any) []any { a, _ := x.([]any); return a }

// What an issue says, on one line as tr '\n' ' ' leaves it.
func quoted(prefix string) string {
	s := fmt.Sprintf("%s. %s (%s%s)\n", interp(get(issue, "title")), interp(get(issue, "body")), prefix, interp(get(issue, "url")))
	return strings.ReplaceAll(s, "\n", " ")
}

func cat(f string) string  { b, _ := os.ReadFile(f); return strings.TrimRight(string(b), "\n") }
func exists(f string) bool { _, err := os.Stat(f); return err == nil }
func nonempty(f string) bool {
	s, err := os.Stat(f)
	return err == nil && s.Size() > 0
}
func write(f, s string) { must("", os.WriteFile(f, []byte(s), 0o644)) }

func short(h string) string { return h[:min(7, len(h))] }
func timeout() string {
	if t := os.Getenv("TIMEOUT_MS"); t != "" {
		return t
	}
	return "14400000" // four hours: a backstop for a hung agent; a three-change cycle took 80 minutes
}

// abs is cd "$1" && pwd: the logical path of a folder that exists.
func abs(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("cd: null directory")
	}
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if s, err := os.Stat(a); err != nil {
		return "", err
	} else if !s.IsDir() {
		return "", fmt.Errorf("%s: not a directory", p)
	}
	return a, nil
}

// ---- reporting

func say(f string, a ...any) { fmt.Fprintf(os.Stderr, "tend: %s\n", fmt.Sprintf(f, a...)) }

func stop(f string, a ...any) {
	msg := fmt.Sprintf(f, a...)
	say("%s", msg)
	stage("needs-you", msg, "a person deals with that, then runs tend again")
	term.notify("tend "+cycle+" needs you", msg, "request")
	os.Exit(3)
}

// What the attempt is doing, for tend status: when, a kind (working, needs-you
// or done) and what, then the next action. Written whole or not at all.
func stage(k, text, next string) {
	if run == "" {
		return
	}
	s := fmt.Sprintf("%s %s: %s\nnext: %s\n", time.Now().Format("2006-01-02T15:04:05-0700"), k, text, next)
	if os.WriteFile(run+"/status.tmp", []byte(s), 0o644) == nil {
		os.Rename(run+"/status.tmp", run+"/status")
	}
}

// Whether status file f says working, as one a tend that died leaves.
func working(f string) bool {
	return regexp.MustCompile(`^[^ ]* working:`).MatchString(strings.SplitN(cat(f), "\n", 2)[0])
}

// The release, set at build time; "dev" for a build from a checkout.
var version = "dev"

func main() {
	args := os.Args[1:]
	if len(args) == 1 && args[0] == "version" {
		fmt.Println("tend " + version)
		return
	}
	mode = "run"
	if len(args) > 0 && slices.Contains([]string{"status", "watch", "propose", "flow"}, args[0]) {
		mode, args = args[0], args[1:]
	}

	state = os.Getenv("TEND_STATE")
	if state == "" {
		x := os.Getenv("XDG_STATE_HOME")
		if x == "" {
			x = os.Getenv("HOME") + "/.local/state"
		}
		state = x + "/tend"
	}
	must("", os.MkdirAll(state+"/runs", 0o755))
	state = must(abs(state)) // tend changes into the attempt later

	// The terminal the agents run in: TEND_TERMINAL, or rex when tend itself
	// runs in rex rather than Herdr.
	switch t := os.Getenv("TEND_TERMINAL"); {
	case t == "rex", t == "" && os.Getenv("REX_SESSION") != "" && os.Getenv("HERDR_ENV") != "1":
		term = &rex{}
	default:
		term = &herdr{}
	}
	if mode == "watch" {
		watch(args)
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: tend <repo> <id>")
		os.Exit(2)
	}
	repo = must(abs(args[0]))
	if mode == "flow" {
		printFlows()
	}
	if mode == "propose" {
		cycle = "intake"
	} else if len(args) > 1 {
		cycle = args[1]
	}
	if cycle == "" && mode != "status" {
		fmt.Fprintln(os.Stderr, "usage: tend <repo> <id>")
		os.Exit(2)
	}
	if mode == "status" {
		if cycle == "" {
			overview()
		}
		status()
	}
	takeLock()
	if mode == "propose" {
		propose()
	}
	settingUp()
	tidy()
	attempt()
}

// grove changes --json in folder r.
func groveIn(r string) (any, error) {
	c := exec.Command("grove", "changes", "--json")
	c.Dir = r
	b, err := c.Output()
	return parse(string(b)), err
}

// The repository's local target fast-forwarded to origin's, fetched: Grove
// reads the local target, so without this a change or cycle merged on GitHub
// stays unseen. Only a fast-forward, and only of a clean checkout when the
// target is checked out; otherwise it is left as it is and said.
func syncTarget(r, t string) error {
	if _, err := quiet("git", "-C", r, "remote", "get-url", "origin"); err != nil || t == "" {
		return nil
	}
	if _, err := quiet("git", "-C", r, "fetch", "-q", "origin", t); err != nil {
		return err
	}
	var err error
	if gitQ("-C", r, "branch", "--show-current") != t {
		_, err = quiet("git", "-C", r, "fetch", "-q", "origin", t+":"+t)
	} else if gitQ("-C", r, "status", "--porcelain", "--untracked-files=no") != "" {
		say("left %s's %s as it is: its checkout has uncommitted changes", r, t)
		return nil
	} else {
		_, err = quiet("git", "-C", r, "merge", "-q", "--ff-only", "origin/"+t)
	}
	if err != nil {
		say("left %s's %s as it is: it cannot be fast-forwarded to origin's", r, t)
	}
	return nil
}

// grove changes --json in the current folder, or nothing without Grove.
func groveJSON() (any, error) {
	s, err := quiet("grove", "changes", "--json")
	if err != nil {
		return nil, err
	}
	return parse(s), nil
}

// The project as Grove publishes it: root, target branch and the attempt
// records, newest first, from origin/grove/runs or, without origin, the local
// grove/runs.
func project() any {
	s, err := out("grove", "changes", "--json")
	if err != nil {
		stop("grove changes --json failed in %s", os.Getenv("PWD"))
	}
	return parse(s)
}

// The newest record of an attempt at Grove work item w: a cycle's, or a
// change's in no cycle, which names the change alone and no cycle.
func itemRun(proj any, w string) any {
	for _, r := range list(get(proj, "runs")) {
		if c := str(get(r, "cycle")); c == w || c == "" && slices.Equal(list(get(r, "changes")), []any{w}) {
			return r
		}
	}
	return nil
}

// Whether Grove knows work item w, a cycle or a change, and whether it is not
// done: a cycle not done, a change not archived.
func workState(proj any, w string) (known, open bool) {
	for _, y := range list(get(proj, "cycles")) {
		if str(get(y, "id")) == w {
			return true, get(y, "done") != true
		}
	}
	for _, c := range list(get(proj, "changes")) {
		if str(get(c, "id")) == w {
			return true, str(get(c, "state")) != "archived"
		}
	}
	return false, false
}

// The run folders, as the shell's runs/*/ lists them.
func runDirs() []string {
	ds, _ := filepath.Glob(state + "/runs/*")
	return slices.DeleteFunc(ds, func(d string) bool { s, err := os.Stat(d); return err != nil || !s.IsDir() })
}

func isIssue(c string) bool { return regexp.MustCompile(`^issue-[0-9]`).MatchString(c) }
func isItem(c string) bool  { return isIssue(c) || isJira(c) }

// ---- Git and GitHub

// The target new work starts from and is compared with: origin's, fetched
// first, when the repository has an origin, so a stale local branch is never
// the base.
func fetchTarget() string {
	if _, err := quiet("git", "-C", repo, "remote", "get-url", "origin"); err != nil {
		return target
	}
	if syncTarget(repo, target) != nil {
		stop("could not fetch %s from origin", target)
	}
	return "origin/" + target
}

func commits(rng string) bool { return git("rev-list", rng) != "" }
func clean() bool             { return git("status", "--porcelain") == "" }
func prNumber() string        { return pr[strings.LastIndex(pr, "/")+1:] }
func reviewsAPI() string      { return "repos/{owner}/{repo}/pulls/" + prNumber() + "/reviews" }

func postJSON(path string, body map[string]any) error {
	b, _ := json.Marshal(body)
	_, err := input(string(b), "gh", "api", path, "--input", "-")
	return err
}

// A verdict counts only when complete and naming the head it was written
// for; a file's name is not enough.
func valid(file, h string) bool {
	b, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	x := parse(string(b))
	_, isBool := get(x, "actionable").(bool)
	_, isString := get(x, "summary").(string)
	c := get(x, "conclusion")
	return get(x, "head") == h && (c == "success" || c == "failure") && isBool && isString
}
