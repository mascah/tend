// tend takes one agreed work item through coding agents and brings back a
// reviewed branch: a Grove cycle, or a GitHub issue of a repository without
// Grove. Claude implements in a Herdr pane, Codex reviews the head, one repair
// if the review asks, and with a listed origin a pull request as a bot account.
//
//	tend <repo> <cycle-id>
//	tend <repo> issue-<n>
//	tend status <repo> [<cycle-id>]
//	tend watch [<repo>...]
//	tend propose <repo>
//	tend flow <repo>
//	tend version
//
// Exit 0: the head is reviewed and the verdict printed. Exit 3: a person is
// needed, and a Herdr notification says why. Exit 4: another tend holds this
// runner. Progress is read again on every run from Git, Grove, GitHub, Herdr
// and the run folders under TEND_STATE (default ~/.local/state/tend), never
// from the candidate. After exit 3, or a kill, the same line carries on: an
// agent still working is waited for, never prompted again, and the one repair
// is used once asked.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	mode, state, repo, cycle, run       string
	kind, target, base, branch, id, dir string
	impl, root, what, notes, trailer, v string
	head, pr, account, where            string
	issue, rec                          any
	item                                request  // the work item, when not a Grove cycle
	lock                                *os.File // held open: its flock is the runner's
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

// POSIX cksum, which names an issue's attempt after its repository.
func cksum(b []byte) uint32 {
	var crc uint32
	add := func(c byte) {
		crc ^= uint32(c) << 24
		for range 8 {
			if crc&0x80000000 != 0 {
				crc = crc<<1 ^ 0x04c11db7
			} else {
				crc <<= 1
			}
		}
	}
	for _, c := range b {
		add(c)
	}
	for l := len(b); l > 0; l >>= 8 {
		add(byte(l))
	}
	return ^crc
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

func needsYou(f string) bool {
	return regexp.MustCompile(`(?m)^[^ ]* needs-you:`).MatchString(cat(f))
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

	// ponytail: one lock for the runner, so one attempt at a time; lock per
	// attempt when capacity grows.
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
		fmt.Fprintln(os.Stderr, "usage: tend <repo> <cycle-id>")
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
		fmt.Fprintln(os.Stderr, "usage: tend <repo> <cycle-id>")
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
	tidy()
	attempt()
}

// tend watch: every WATCH_SECONDS (default 300), each attempt whose pull
// request is open with a change request on its head is run again, and for
// each repository named tend propose runs and then the oldest cycle Grove
// would claim that its plan marks unattended is started: marking it is the
// person's approval to build it with no one there. An attempt or request
// waiting for a person (needs-you) is left alone until a person has run tend
// on it; a claimed cycle is Grove's to refuse.
func watch(repos []string) {
	self := must(os.Executable())
	again := func(a ...string) {
		c := exec.Command(self, a...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		c.Run()
	}
	for {
		for _, r := range repos {
			r, err := abs(r)
			if err != nil {
				continue
			}
			c := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
			c.Dir = r
			b, err := c.Output()
			if err != nil {
				continue
			}
			if g, err := groveIn(r); err == nil {
				syncTarget(r, str(get(g, "target")))
			}
			s := strings.ReplaceAll(strings.TrimSpace(string(b)), "/", "_")
			waits, _ := filepath.Glob(state + "/intake/" + s + "-*/status")
			if !slices.ContainsFunc(waits, needsYou) {
				again("propose", r)
			}
			if c := unattended(r); c != "" {
				say("starting unattended cycle %s in %s", c, r)
				again(r, c)
			}
		}
		ds := runDirs()
		for _, d := range ds {
			if !exists(d+"/pr") || !exists(d+"/attempt") || needsYou(d+"/status") {
				continue
			}
			s, _ := quiet("gh", "pr", "view", cat(d+"/pr"), "--json", "state,headRefOid,reviews", "--jq",
				`select(.state == "OPEN") | .headRefOid as $h | [.reviews[] | select(.state == "CHANGES_REQUESTED" and .commit.oid == $h)] | last | .id // empty`)
			if s == "" || exists(d+"/answered-"+s) {
				continue
			}
			a := strings.SplitN(cat(d+"/attempt"), "\n", 3)
			if len(a) < 2 {
				continue
			}
			say("changes requested on %s; running tend on %s", cat(d+"/pr"), a[1])
			again(a[0], a[1])
		}
		if os.Getenv("WATCH_ONCE") != "" {
			os.Exit(0)
		}
		secs, err := strconv.Atoi(os.Getenv("WATCH_SECONDS"))
		if err != nil {
			secs = 300
		}
		time.Sleep(time.Duration(secs) * time.Second)
	}
}

// The oldest cycle of the repository that Grove would claim now (eligible, not
// claimed, the project not paused) and whose plan marks it unattended, or "".
func unattended(r string) string {
	g, err := groveIn(r)
	if err != nil {
		return ""
	}
	best, when := "", ""
	for _, x := range list(get(g, "cycles")) {
		if get(x, "unattended") == true && get(x, "runs") == true && get(x, "done") != true && (best == "" || str(get(x, "created")) < when) {
			best, when = str(get(x, "id")), str(get(x, "created"))
		}
	}
	return best
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

func firstRun(proj any, key, val string) any {
	for _, r := range list(get(proj, "runs")) {
		if str(get(r, key)) == val {
			return r
		}
	}
	return nil
}

func cycleOpen(proj any, c string) bool {
	for _, y := range list(get(proj, "cycles")) {
		if str(get(y, "id")) == c && get(y, "done") != true {
			return true
		}
	}
	return false
}

// tend status <repo>: what waits for whom in the repository, read from Grove's
// JSON, GitHub and this runner's run folders. Reads only.
func overview() {
	cd(repo)
	proj, err := groveJSON()
	if err != nil { // a repository without Grove
		proj = map[string]any{}
	}
	fmt.Println("Approved changes without a cycle (plan one: /grove-plan):")
	for _, c := range list(get(proj, "changes")) {
		if str(get(c, "state")) == "approved" && get(c, "cycle") == nil {
			ready := ""
			if r := get(c, "ready"); r == nil || r == false {
				ready = ", not ready"
			}
			fmt.Printf("  %s %s%s\n", interp(get(c, "id")), interp(get(c, "title")), ready)
		}
	}
	fmt.Println("Cycles not done:")
	for _, y := range list(get(proj, "cycles")) {
		if d := get(y, "done"); d == nil || d == false {
			fmt.Printf("  %s %s: %s\n", interp(get(y, "id")), interp(get(y, "title")), interp(get(y, "verdict")))
		}
	}
	fmt.Println("Attempts on this runner, for cycles not done and issues:")
	ds := runDirs()
	for _, d := range ds {
		a := strings.Split(cat(d+"/attempt"), "\n")
		if a[0] != repo || !exists(d+"/status") || len(a) < 2 {
			continue
		}
		if !isItem(a[1]) && !cycleOpen(proj, a[1]) {
			continue
		}
		line := strings.SplitN(cat(d+"/status"), "\n", 2)[0]
		if _, rest, found := strings.Cut(line, " "); found {
			line = rest
		}
		fmt.Printf("  %s (%s): %s\n", filepath.Base(d), a[1], line)
	}
	s, err := quiet("gh", "pr", "list", "--state", "open", "--json", "number,title,headRefName,url,reviewDecision")
	if err != nil {
		fmt.Println("GitHub: not readable with gh here")
		os.Exit(0)
	}
	prs := list(parse(s))
	pull := func(p any) string {
		return fmt.Sprintf("  #%s %s %s", interp(get(p, "number")), interp(get(p, "title")), interp(get(p, "url")))
	}
	fmt.Println("Proposals to review (answer their questions, revise or merge):")
	for _, p := range prs {
		if h := str(get(p, "headRefName")); strings.HasPrefix(h, "tend/issue-") || strings.HasPrefix(h, "tend/jira-") {
			fmt.Println(pull(p))
		}
	}
	fmt.Println("Implementations to review (review and merge):")
	for _, p := range prs {
		if h := str(get(p, "headRefName")); strings.HasPrefix(h, "grove/") || strings.HasPrefix(h, "tend/build-") {
			asked := ""
			if str(get(p, "reviewDecision")) == "CHANGES_REQUESTED" {
				asked = ", changes requested: tend makes them on its next run"
			}
			fmt.Println(pull(p) + asked)
		}
	}
	fmt.Println("Requests waiting for tend propose:")
	handled, _ := quiet("gh", "pr", "list", "--state", "all", "--limit", "1000", "--json", "headRefName", "--jq", `[.[].headRefName | select(startswith("tend/issue-") or startswith("tend/jira-"))]`)
	done := list(parse(handled))
	lines, _ := quiet("gh", "issue", "list", "--label", "tend", "--state", "open", "--json", "number,title,url", "--jq", `.[] | "  #\(.number) \(.title) \(.url)"`)
	for _, line := range strings.Split(lines, "\n") {
		if line == "" {
			continue
		}
		num, _, _ := strings.Cut(strings.TrimPrefix(line, "  #"), " ")
		if !slices.Contains(done, any("tend/issue-"+num)) {
			fmt.Println(line)
		}
	}
	origin := gitQ("remote", "get-url", "origin")
	for _, m := range jiraMaps() {
		if m.origin != origin {
			continue
		}
		if os.Getenv("JIRA_EMAIL") == "" || os.Getenv("JIRA_API_TOKEN") == "" {
			fmt.Printf("  Jira %s: not readable without JIRA_EMAIL and JIRA_API_TOKEN\n", m.project)
			continue
		}
		for _, x := range jiraSearch(m, fmt.Sprintf(`project = "%s" AND labels = "tend" AND statusCategory != Done ORDER BY created ASC`, m.project)) {
			if k := str(get(x, "key")); !slices.Contains(done, any("tend/jira-"+k)) {
				fmt.Printf("  %s %s %s/browse/%s\n", k, str(get(x, "fields", "summary")), m.site, k)
			}
		}
	}
	os.Exit(0)
}

// The run folders, as the shell's runs/*/ lists them.
func runDirs() []string {
	ds, _ := filepath.Glob(state + "/runs/*")
	return slices.DeleteFunc(ds, func(d string) bool { s, err := os.Stat(d); return err != nil || !s.IsDir() })
}

func isIssue(c string) bool { return regexp.MustCompile(`^issue-[0-9]`).MatchString(c) }
func isItem(c string) bool  { return isIssue(c) || isJira(c) }

// An issue's attempt is named after it and its repository: issue-7-<cksum>,
// jira-scrum-5-<cksum>.
func issueRun(c string) string {
	if isJira(c) {
		c = "jira-" + strings.ToLower(c)
	}
	return fmt.Sprintf("%s/runs/%s-%d", state, c, cksum([]byte(repo)))
}

// tend status <repo> <cycle>: the newest attempt at it, what it last recorded
// and whether a tend is on it now. Reads only; takes no lock, prompts nobody.
func status() {
	if isItem(cycle) {
		run = issueRun(cycle)
		if !exists(run) {
			fmt.Printf("no attempt at %s\n", cycle)
			os.Exit(0)
		}
		fmt.Printf("attempt %s of %s: tend/build-%s\n", filepath.Base(run), cycle, strings.TrimPrefix(cycle, "issue-"))
	} else {
		cd(repo)
		s := must(out("grove", "changes", "--json"))
		r := firstRun(parse(s), "cycle", cycle)
		if r == nil {
			fmt.Printf("no attempt at %s\n", cycle)
			os.Exit(0)
		}
		run = state + "/runs/" + str(get(r, "id"))
		fmt.Printf("attempt %s of %s: %s in %s\n", interp(get(r, "id")), interp(get(r, "cycle")), interp(get(r, "branch")), interp(get(r, "checkout")))
	}
	held := false
	if f, err := os.OpenFile(state+"/lock", os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
		held = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil
		f.Close()
	}
	if held && strings.Contains(cat(state+"/lock"), "cycle "+cycle+" in "+repo) {
		fmt.Printf("tend is on it: %s\n", cat(state+"/lock"))
	} else if first := strings.SplitN(cat(run+"/status"), "\n", 2)[0]; regexp.MustCompile(`^[^ ]* working:`).MatchString(first) {
		fmt.Println(first)
		fmt.Println("next: tend stopped before this finished; run it again, and it waits for an agent still working rather than prompting it again")
		os.Exit(0)
	}
	if b, err := os.ReadFile(run + "/status"); err == nil {
		os.Stdout.Write(b)
	} else {
		fmt.Println("nothing recorded yet")
	}
	os.Exit(0)
}

// One controller on this runner. The lock is an flock on a descriptor no child
// inherits, so it goes when tend exits or is killed, whatever children (an
// agent's wait, a shell) outlive it. A lock just dropped is waited for briefly.
func takeLock() {
	f, err := os.OpenFile(state+"/lock", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		stop("could not take the lock in %s", state)
	}
	for i := 0; ; i++ {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK {
			stop("could not take the lock in %s", state)
		}
		if i == 10 {
			say("another tend holds this runner: %s", cat(state+"/lock"))
			os.Exit(4)
		}
		time.Sleep(200 * time.Millisecond)
	}
	lock = f
	write(state+"/lock", fmt.Sprintf("pid %d, cycle %s in %s\n", os.Getpid(), cycle, repo))
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
func start(name, kind string, fresh bool, args ...string) string {
	p, err := term.start(name, kind, fresh, args)
	for range 4 {
		screen := strings.Join(strings.Fields(term.read(name)), " ") // a hint or warning may wrap
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
		// Codex comes out of its dialog as done, not idle.
		err = nil
		if _, why := term.wait(name, "60000", "idle", "done"); why != "" {
			err = errors.New(why)
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

// The repository's worktrees and local branches, one a line.
func layout() string {
	var ls []string
	for _, line := range strings.Split(git("worktree", "list", "--porcelain"), "\n") {
		if p, found := strings.CutPrefix(line, "worktree "); found {
			ls = append(ls, "worktree "+p)
		}
	}
	for _, b := range strings.Split(git("for-each-ref", "--format=%(refname:short)", "refs/heads"), "\n") {
		ls = append(ls, "branch "+b)
	}
	return strings.Join(ls, "\n")
}

// Work left anywhere but the attempt's branch in its worktree is a person's.
// The repository's layout is recorded just before each implementer prompt
// goes, and a rerun compares with the same record.
func contained() {
	if !exists(run + "/layout") {
		return
	}
	why := ""
	if b := git("branch", "--show-current"); b != branch {
		why = "its worktree is on " + b + ", not " + branch
	}
	was := strings.Split(cat(run+"/layout"), "\n")
	for _, l := range strings.Split(layout(), "\n") {
		if why != "" {
			break
		}
		if slices.Contains(was, l) {
			continue
		}
		if p, found := strings.CutPrefix(l, "worktree "); found {
			why = "it added the worktree " + p
		} else if b := strings.TrimPrefix(l, "branch "); commits(base + ".." + b) {
			why = "it made the branch " + b + ", with commits"
		}
	}
	if why != "" {
		stop("%s left work outside %s: %s. tend reviews and publishes nothing; move the work onto %s and remove the rest, then run tend again", impl, branch, why, branch)
	}
}

// ---- Git and GitHub

// The account a repository's origin is published as, from $state/publish
// ("<origin-url> <account>" a line), or nothing when it is not listed.
func accountOf() string {
	o := gitQ("remote", "get-url", "origin")
	for _, line := range strings.Split(cat(state+"/publish"), "\n") {
		if f := strings.Fields(line); len(f) > 1 && f[0] == o {
			return f[1]
		}
	}
	return ""
}

func token() {
	t, err := out("gh", "auth", "token", "--user", account)
	if err != nil {
		stop("gh holds no token for %s; sign it in with gh auth login, then switch back", account)
	}
	os.Setenv("GH_TOKEN", t)
}

// Push HEAD to origin's branch as the account whose token is GH_TOKEN, over
// HTTPS even when origin is an SSH URL; never forced.
func pushAs(b string) bool {
	_, err := out("git", "-c", "credential.helper=", "-c", "credential.helper=!gh auth git-credential", "-c", "url.https://github.com/.pushInsteadOf=git@github.com:",
		"push", "-q", "origin", "HEAD:refs/heads/"+b)
	return err == nil
}

// The worktree holding branch b, opened as the place, or a new one from base:
// sets dir. A branch is found wherever its worktree is, so a rerun finds its
// work.
func worktreeFor(b, base, label string) {
	dir = ""
	w := ""
	for _, line := range strings.Split(git("-C", repo, "worktree", "list", "--porcelain"), "\n") {
		if p, found := strings.CutPrefix(line, "worktree "); found {
			w = p
		} else if line == "branch refs/heads/"+b {
			dir = w
		}
	}
	if dir != "" {
		term.open(dir)
		return
	}
	dir = term.worktree(b, base, label)
}

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

// The worktree and its local branch removed, with the place the agents ran
// in; left, and said, while it holds uncommitted files.
func removeWorktree(d, b string) {
	if !exists(d) {
		return
	}
	if s := gitQ("-C", d, "status", "--porcelain"); s != "" {
		say("left %s: it has uncommitted files", d)
		return
	}
	term.closePlace(d)
	if exists(d) {
		quiet("git", "-C", repo, "worktree", "remove", d)
	}
	if !exists(d) {
		quiet("git", "-C", repo, "branch", "-D", b)
		say("removed %s", d)
	}
}

// Worktrees tend made for this repository whose pull request is merged or
// closed, removed when their head is the pull request's, so nothing is lost;
// and those of attempts with no pull request whose cycle is done, when their
// branch has no commit the target lacks. Never one with a worktree inside it
// or whose implementer is working. A run folder names its worktree in
// worktree, deleted once it is gone.
func tidy() {
	var listed []string
	for _, line := range strings.Split(git("-C", repo, "worktree", "list", "--porcelain"), "\n") {
		if p, found := strings.CutPrefix(line, "worktree "); found {
			listed = append(listed, p)
		}
	}
	fs, _ := filepath.Glob(state + "/runs/*/worktree")
	more, _ := filepath.Glob(state + "/intake/*/worktree")
	for _, f := range append(fs, more...) {
		d, r := cat(f), filepath.Dir(f)
		if !slices.ContainsFunc(listed, func(p string) bool { return sameDir(p, d) }) {
			continue
		}
		b := ""
		if nonempty(r + "/pr") {
			s, err := out("gh", "pr", "view", cat(r+"/pr"), "--json", "state,headRefOid,headRefName", "--jq", `select(.state != "OPEN") | .headRefOid + " " + .headRefName`)
			h, pb, _ := strings.Cut(s, " ")
			if err != nil || h == "" {
				continue
			}
			if gitQ("-C", d, "rev-parse", "HEAD") != h {
				say("left %s: its head is not its pull request's", d)
				continue
			}
			b = pb
		} else if b = superseded(d, r); b == "" {
			continue
		}
		real, _ := filepath.EvalSymlinks(d)
		if slices.ContainsFunc(listed, func(p string) bool { return strings.HasPrefix(p, real+"/") }) {
			say("left %s: a worktree is inside it", d)
			continue
		}
		if filepath.Dir(r) == state+"/runs" && term.status("impl-"+filepath.Base(r)) == "working" {
			say("left %s: its implementer is working", d)
			continue
		}
		removeWorktree(d, b)
		if !exists(d) {
			os.Remove(f)
		}
	}
}

// The branch of an attempt with no pull request whose Grove cycle is done
// and whose branch the local target already holds, as when its work reached
// the target some other way; or nothing.
func superseded(d, r string) string {
	a := strings.Split(cat(r+"/attempt"), "\n")
	if filepath.Dir(r) != state+"/runs" || len(a) < 2 || !sameDir(a[0], repo) || isItem(a[1]) {
		return ""
	}
	g, err := groveIn(repo)
	if err != nil || cycleOpen(g, a[1]) || !slices.ContainsFunc(list(get(g, "cycles")), func(y any) bool { return str(get(y, "id")) == a[1] }) {
		return ""
	}
	if !ok("git", "-C", d, "merge-base", "--is-ancestor", "HEAD", str(get(g, "target"))) {
		return ""
	}
	return gitQ("-C", d, "branch", "--show-current")
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

// tend propose <repo>: one pass of intake from the repository's GitHub issues
// and the Jira projects mapped to it. Each open request labelled tend, oldest
// first, becomes a Grove change drafted by Claude with no one in the session,
// on branch tend/issue-<n> or tend/jira-<KEY>, opened as a pull request by the
// publishing account; or, when no change can be drafted from it, a comment on
// it saying why, and the label taken off, so it is not tried again until a
// person puts it back. The request's number or key is its identity: a pull
// request from its branch, in any state, means it is not drafted again. A
// handoff is finished when the run folder records linked: a folder with a pull
// request and no linked is completed on every pass, whatever the cap or the
// request's label. At most CAP (default 5) proposal pull requests stay open
// per repository. Each prompt is marked just before it goes, as in an attempt,
// so none is sent twice or skipped.
func propose() {
	cd(repo)
	account = accountOf()
	if account == "" {
		o, err := quiet("git", "remote", "get-url", "origin")
		if err != nil {
			o = repo + " has no origin; it"
		}
		stop("%s is not listed in %s/publish, so tend takes no requests from it", o, state)
	}
	token()
	proj := project()
	target, root = str(get(proj, "target")), str(get(proj, "root"))
	base = fetchTarget()
	loadFlow("propose", base)
	tidy()
	slug, err := out("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	if err != nil {
		stop("gh cannot see %s's repository", repo)
	}
	prefix := state + "/intake/" + strings.ReplaceAll(slug, "/", "_") + "-"
	limit, err := strconv.Atoi(os.Getenv("CAP"))
	if err != nil {
		limit = 5
	}
	origin := gitQ("remote", "get-url", "origin")

	// Handoffs a stopped pass left unfinished.
	folders, _ := filepath.Glob(prefix + "*")
	for _, f := range folders {
		rest := strings.TrimPrefix(f, prefix)
		if strings.Contains(rest, ".declined-") || !exists(f+"/pr") || exists(f+"/linked") {
			continue
		}
		if k, isJ := strings.CutPrefix(rest, "jira-"); isJ {
			draft(prefix, jiraGet(k))
		} else {
			draft(prefix, ghRequest(rest, "title,body,url"))
		}
	}

	type pending struct {
		branch string
		get    func() request
	}
	var todo []pending
	nums, _ := out("gh", "issue", "list", "--label", "tend", "--state", "open", "--json", "number", "--jq", "sort_by(.number) | .[].number")
	for _, num := range strings.Fields(nums) {
		todo = append(todo, pending{"tend/issue-" + num, func() request { return ghRequest(num, "title,body,url") }})
	}
	for _, m := range jiraMaps() {
		if m.origin != origin {
			continue
		}
		for _, x := range jiraSearch(m, fmt.Sprintf(`project = "%s" AND labels = "tend" AND statusCategory != Done ORDER BY created ASC`, m.project)) {
			todo = append(todo, pending{"tend/jira-" + str(get(x, "key")), func() request { return jiraRequest(m, x) }})
		}
	}
	for _, t := range todo {
		if s, _ := out("gh", "pr", "list", "--head", t.branch, "--state", "all", "--json", "url", "--jq", ".[0].url // empty"); s != "" {
			continue
		}
		o, _ := out("gh", "pr", "list", "--author", account, "--state", "open", "--json", "headRefName", "--jq", `[.[] | select(.headRefName | startswith("tend/issue-") or startswith("tend/jira-"))] | length`)
		if open, _ := strconv.Atoi(o); open >= limit {
			say("%d proposal pull requests are open, the cap; the rest wait", open)
			break
		}
		draft(prefix, t.get())
	}
	os.Exit(0)
}

const marker = "<!-- tend:declined"

func draft(prefix string, r request) {
	b := "tend/" + r.slug
	run = prefix + r.folder
	must("", os.MkdirAll(run, 0o755))
	if !nonempty(run + "/id") {
		write(run+"/id", strconv.FormatInt(time.Now().Unix(), 10)+"\n")
	}
	id = cat(run + "/id")
	issue = r.fields
	worktreeFor(b, base, "tend: "+r.ref)
	write(run+"/worktree", dir+"\n")
	cd(dir)
	impl = "prop-" + r.agent

	// The propose flow: the drafter drafts, then commits the draft or says why
	// there is none, each asked once, as the built-in flow says or the
	// project's.
	vars = map[string]string{"branch": b, "base": base, "run": run, "item": r.ref, "target": target,
		"request": quoted(r.source), "refs": r.refs, "ref": r.ref}
	walkFlow(func() string { proposal(r, b); return "ok" })
	cd(repo)
}

// The propose flow's publish: the proposal opened as a pull request by the
// account, or the request told why there is none.
func proposal(r request, b string) {
	switch {
	case commits(base + "..HEAD"):
		if !clean() {
			stop("%s: %s left uncommitted files in %s", r.ref, impl, dir)
		}
		changed := "\n" + git("diff", "--name-only", base+"...HEAD")
		if !strings.Contains(changed, "\n"+root+"/changes/") {
			stop("%s: %s has commits but no change under %s/changes/", r.ref, b, root)
		}
		stage("working", "publishing the proposal from "+r.ref, "a person reviews it")
		var err error
		pr, err = out("gh", "pr", "list", "--head", b, "--state", "all", "--json", "url", "--jq", ".[0].url // empty")
		if err != nil {
			stop("could not look for %s's pull request", b)
		}
		// Pushed only to open the pull request: once it exists the branch is
		// people's, and may be gone after a merge.
		if pr == "" {
			if !pushAs(b) {
				stop("could not push %s as %s", b, account)
			}
			pr, err = input(git("log", "-1", "--format=%b")+"\n", "gh", "pr", "create", "--base", target, "--head", b, "--title", git("log", "-1", "--format=%s"), "--body-file", "-")
			if err != nil {
				stop("could not open a pull request for %s; run tend propose again, and it looks for one before opening another", b)
			}
		}
		write(run+"/pr", pr+"\n")
		head = git("rev-parse", "HEAD")
		posted, _ := out("gh", "api", reviewsAPI(), "--paginate", "--jq", fmt.Sprintf(`.[] | select(.user.login == "%s" and .commit_id == "%s") | .id`, account, head))
		if posted == "" {
			re := regexp.MustCompile("^" + regexp.QuoteMeta(root) + `/changes/([0-9]*-[a-z0-9]*)-`)
			var ids []string
			for _, f := range strings.Split(git("diff", "--name-only", "--diff-filter=A", base+"...HEAD"), "\n") {
				if m := re.FindStringSubmatch(f); m != nil && !slices.Contains(ids, m[1]) {
					ids = append(ids, m[1])
				}
			}
			sort.Strings(ids)
			q := cat(run + "/summary")
			if q == "" {
				q = "The drafter wrote no summary of open questions; see the proposal."
			}
			body := fmt.Sprintf("**Proposal drafted from %s with no one in the session**\n\n%s\n\nMerging approves the change with any question still open, and its implementation may then choose those answers within Grove's rules. Answer the consequential ones before merging, by revising the proposal on `%s` (`/grove-propose %s`). Once merged, the change waits for a cycle: `/grove-plan`.", r.from, q, b, strings.Join(ids, " "))
			if postJSON(reviewsAPI(), map[string]any{"commit_id": head, "event": "COMMENT", "body": body}) != nil {
				stop("could not post the summary on %s", pr)
			}
		}
		if r.link != nil && r.link(pr) != nil {
			stop("could not say on %s where its proposal is (%s); run tend propose again to finish", r.ref, pr)
		}
		write(run+"/linked", pr+"\n")
		stage("done", "proposal from "+r.ref+" at "+pr, "a person answers its open questions, revises or merges it")
		say("%s: %s", r.ref, pr)
	case nonempty(run + "/declined"):
		// Said on the request once, found by the attempt's marker, then the label off.
		if r.decline(cat(run+"/declined"), id) != nil {
			stop("could not comment on %s", r.ref)
		}
		if r.unlabel() != nil {
			stop("could not take the tend label off %s", r.ref)
		}
		stage("done", "no proposal from "+r.ref+"; said why on it", "a person edits the request and labels it again")
		must("", os.Rename(run, run+".declined-"+id)) // a fresh attempt when the label comes back
		cd(repo)
		removeWorktree(dir, b)
		say("%s: no proposal, said why", r.ref)
	default:
		stop("%s: %s committed no change and wrote no reason. Look at its pane, then commit the draft or write %s/declined, or delete %s/followed to have it asked again", r.ref, impl, run, run)
	}
	cd(repo)
}

// ---- an attempt

// 1. The attempt, and what the work item says it must do. Two kinds:
//
// A Grove cycle. The newest record for it, if its checkout still holds its
// branch; otherwise a new worktree and a claim made inside it, so the
// attempt's branch is the worktree's. Grove holds approval, eligibility and
// the scope; the build is done when each change is archived on the branch.
//
// issue-<n>, a GitHub issue of a repository without Grove: where Grove holds
// the scope, a request goes through a proposal (tend propose) instead, so the
// maintained specs keep up with what is built. It is approved by the
// tend-build label, which only people who may label can put on; it may start
// while it is open and labelled, and a closed one is done. Its title and
// description are the scope. Its branch is tend/build-<n>, found wherever its
// worktree is. The build is done when the branch has a commit and the
// implementer's notes for the reviewer are in the run folder.
func attempt() {
	if isItem(cycle) {
		kind = "issue"
		cd(repo)
		ref := "issue #" + strings.TrimPrefix(cycle, "issue-")
		if isJira(cycle) {
			ref = "issue " + cycle
		}
		if ok("grove", "changes", "--json") {
			stop("%s uses Grove, so %s is not built directly: label it tend for tend propose to draft a proposal, then approve and plan it", repo, ref)
		}
		if isJira(cycle) {
			item = jiraGet(cycle)
		} else {
			item = ghRequest(strings.TrimPrefix(cycle, "issue-"), "title,body,url,state,labels")
		}
		issue = item.fields
		if str(get(issue, "state")) != "OPEN" {
			say("%s is closed; nothing to do", item.ref)
			os.Exit(0)
		}
		labelled := false
		for _, l := range list(get(issue, "labels")) {
			labelled = labelled || str(get(l, "name")) == "tend-build"
		}
		if !labelled {
			stop("%s is not labelled tend-build, so it is not approved for building", item.ref)
		}
		var err error
		if target, err = out("gh", "repo", "view", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name"); err != nil {
			stop("gh cannot see %s's repository", repo)
		}
		branch = "tend/build-" + item.key
		if s, _ := out("gh", "pr", "list", "--head", branch, "--state", "all", "--json", "state", "--jq", ".[0].state // empty"); s == "MERGED" {
			say("%s's pull request from %s is merged; nothing to do", item.ref, branch)
			os.Exit(0)
		}
		id = filepath.Base(issueRun(cycle))
		base = fetchTarget()
		loadFlow("build", base)
		worktreeFor(branch, base, "tend: "+item.ref)
		cd(dir)
		say("%s on %s in %s", item.ref, branch, dir)
		what = item.named + ", quoted here: " + quoted("")
	} else {
		kind = "grove"
		cd(repo)
		proj := project()
		target = str(get(proj, "target"))
		base = fetchTarget()
		loadFlow("build", base)
		rec = firstRun(proj, "cycle", cycle)
		dir = str(get(rec, "checkout"))
		if dir != "" && gitQ("-C", dir, "branch", "--show-current") == str(get(rec, "branch")) {
			term.open(dir)
			cd(dir)
			say("carrying on attempt %s in %s", str(get(rec, "id")), dir)
		} else {
			tmp := "tend-" + time.Now().Format("150405")
			dir = term.worktree(tmp, base, "tend: "+cycle)
			cd(dir)
			if _, err := sh(nil, os.Stderr, false, "grove", "claim", cycle); err != nil {
				stop("grove claim refused %s", cycle)
			}
			b := git("branch", "--show-current")
			if b != tmp {
				quiet("git", "branch", "-D", tmp) // Herdr needs a branch; Grove's claim makes its own
			}
			rec = firstRun(project(), "branch", b)
			if rec == nil {
				stop("claimed %s, but Grove shows no record for %s", cycle, b)
			}
			say("claimed %s as attempt %s in %s", cycle, str(get(rec, "id")), dir)
		}
		branch, id = str(get(rec, "branch")), str(get(rec, "id"))
		root = str(get(project(), "root")) // the attempt's, not the caller's
	}
	impl = "impl-" + id
	run = state + "/runs/" + id
	must("", os.MkdirAll(run, 0o755))
	// What the reviewer is told the change must do, where the implementer's
	// claims are, and the line that ends the squash commit message.
	if kind == "grove" {
		what = "in the change folders that diff moves into " + root + "/changes/archive/ (proposal.md, specs, tasks.md, implementation.md); reproduce each scenario the change spec gives"
		notes = "implementation.md"
		var cs []string
		for _, c := range list(get(rec, "changes")) {
			cs = append(cs, str(c))
		}
		trailer = "Grove: " + strings.Join(cs, ", ")
	} else {
		what += "; reproduce what it describes. The implementer's notes are in " + run + "/implemented"
		notes = run + "/implemented"
		trailer = item.closes
	}
	// The prototype kept an attempt's verdicts and repair mark in .tend/ in the
	// checkout. They move to the attempt's folder, never over a file there.
	if exists(".tend") {
		fs, _ := filepath.Glob(".tend/*")
		for _, f := range fs {
			if t := run + "/" + filepath.Base(f); !exists(t) {
				os.Rename(f, t)
			}
		}
		if os.Remove(".tend") == nil {
			say("moved .tend/ to %s", run)
		} else {
			left, _ := os.ReadDir(".tend")
			var names []string
			for _, e := range left {
				names = append(names, e.Name())
			}
			stop("left in .tend/, as %s holds the same: %s", run, strings.Join(names, "\n"))
		}
	}
	// The prototype wrote repaired before sending the repair, so it means asked.
	if exists(run+"/repaired") && !exists(run+"/repair-asked") {
		os.Rename(run+"/repaired", run+"/repair-asked")
	}
	write(run+"/attempt", repo+"\n"+cycle+"\n") // for tend watch
	write(run+"/worktree", dir+"\n")            // for tidy

	// 2. The build flow: implement, review, repair, publish and requested
	//    changes as the built-in flow says, or as the project's does.
	vars = map[string]string{"branch": branch, "base": base, "run": run, "item": cycle, "target": target,
		"what": what, "notes": notes, "kind": kind, "named": item.named, "scope": quoted("")}
	walkFlow(publishStep)

	if v == "" {
		stage("done", "the flow ended for "+branch+" with no verdict", "a person reviews "+where+" and decides whether to merge it")
		say("done: %s, with no verdict", branch)
		os.Exit(0)
	}
	verdict := parse(cat(v))
	conclusion, h := str(get(verdict, "conclusion")), short(str(get(verdict, "head")))
	term.notify("tend "+cycle+": "+conclusion, where+" at "+h+" is ready for you", "done")
	stage("done", fmt.Sprintf("%s for %s at %s, verdict %s", conclusion, branch, h, v), "a person reviews "+where+" and decides whether to merge it")
	say("done: %s, verdict %s", branch, v)
	var pretty bytes.Buffer
	if json.Indent(&pretty, []byte(cat(v)), "", "  ") == nil {
		fmt.Println(pretty.String())
	}
	os.Exit(0)
}

// Done is what Git holds, never what the pane's state says: why the branch is
// not a finished candidate, or nothing.
func unbuilt() string {
	if !clean() {
		return branch + " has uncommitted files"
	}
	if kind == "issue" {
		if !commits(base + "..HEAD") {
			return branch + " has no commit for " + item.ref
		}
		if !nonempty(run + "/implemented") {
			return impl + " wrote no notes for the reviewer in " + run + "/implemented"
		}
		return ""
	}
	for _, c := range list(get(rec, "changes")) {
		if m, _ := filepath.Glob(root + "/changes/archive/" + str(c) + "-*"); len(m) == 0 {
			return "change " + str(c) + " is not archived on " + branch
		}
	}
	return ""
}

func built() {
	if why := unbuilt(); why != "" {
		stop("%s", why)
	}
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

// The build flow's publish: the head published, then whether a person
// requested changes on it that the flow's request step should make. Each
// request is asked of the implementer once: request-<review id> is written
// just before the prompt goes; a request on an older head was answered by a
// later commit. tend does not wait here for people: tend watch runs it again
// when one requests changes.
var told bool

func publishStep() string {
	where, account = branch, accountOf()
	if account == "" && !told {
		told = true
		if o, err := quiet("git", "remote", "get-url", "origin"); err == nil {
			say("not publishing: %s is not listed in %s/publish", o, state)
		}
	}
	publish()
	if pr == "" {
		return "ok"
	}
	req, err := requested()
	if err != nil {
		stop("could not read the reviews of %s", pr)
	}
	if req == nil || exists(run+"/answered-"+str(get(req, "node_id"))) {
		return "ok"
	}
	rid := str(get(req, "id"))
	comments, err := out("gh", "api", reviewsAPI()+"/"+rid+"/comments", "--paginate", "--jq", `.[] | "\(.path):\(.line // .original_line // "?"): \(.body)"`)
	if err != nil {
		stop("could not read the comments of review %s on %s", rid, pr)
	}
	if comments != "" {
		comments = " Their comments on lines: " + comments
	}
	reviewReq = req
	vars["who"], vars["review_id"] = str(get(req, "user", "login")), rid
	vars["review"], vars["comments"] = strings.TrimRight(str(get(req, "body")), "\n"), comments
	return "requested"
}

// 5. Publish, when origin is listed in $state/publish with the GitHub account
// to publish as, "<origin-url> <account>" a line; otherwise the verdict stays
// here. Every write is that account's, never the owner's, so the owner can
// approve: the branch pushed, never forced; one pull request, squash-merge
// ready: its title and description the commit message the implementer
// writes, written again after a requested change until a person edits them;
// a review per reviewed head, as a comment since a PR's author may not
// approve or request changes; and the
// tend/review status on that head, so a later commit carries no passing
// result. The PR and reviews are looked up on GitHub each run, never
// remembered, so a rerun after any interruption repeats neither. Merging
// stays a person's.
func publish() {
	if account == "" {
		return
	}
	head = git("rev-parse", "HEAD")
	stage("working", fmt.Sprintf("publishing %s at %s as %s", branch, short(head), account), "tend pushes it and opens its pull request or reviews the one there")
	token()
	if !pushAs(branch) {
		stop("could not push %s to origin as %s; if it has commits this checkout lacks, look at them, then run tend again", branch, account)
	}
	var err error
	if pr, err = out("gh", "pr", "list", "--head", branch, "--state", "all", "--json", "url", "--jq", ".[0].url // empty"); err != nil {
		stop("could not look for %s's pull request", branch)
	}
	// The PR is squash-merged: its title and description become the commit's,
	// so the implementer, who knows what changed and why, writes them.
	if pr == "" {
		title, body := titleBody(message(""))
		if pr, err = input(body, "gh", "pr", "create", "--base", target, "--head", branch, "--title", title, "--body-file", "-"); err != nil {
			stop("could not open a pull request for %s; run tend again, and it looks for one before opening another", branch)
		}
		write(run+"/described", head+"\n")
	} else if d := cat(run + "/described"); d != "" && d != head {
		redescribe(d)
	}
	// Every verdict of the attempt on a commit of the branch, oldest first, so
	// a failure repaired within one run still shows.
	posted, err := out("gh", "api", reviewsAPI(), "--paginate", "--jq", fmt.Sprintf(`.[] | select(.user.login == "%s") | .commit_id`, account))
	if err != nil {
		stop("could not read %s's reviews", pr)
	}
	fs, _ := filepath.Glob(run + "/verdict-*.json")
	mtime := func(f string) time.Time { s, _ := os.Stat(f); return s.ModTime() }
	sort.SliceStable(fs, func(i, j int) bool { return mtime(fs[i]).Before(mtime(fs[j])) })
	for _, f := range fs {
		x := parse(cat(f))
		h := str(get(x, "head"))
		if !valid(f, h) || !ok("git", "merge-base", "--is-ancestor", h, "HEAD") || slices.Contains(strings.Split(posted, "\n"), h) {
			continue
		}
		// The status first: a rerun after it, before the review, sets it again.
		c := str(get(x, "conclusion"))
		st := "failure"
		if c == "success" {
			st = "success"
		}
		if _, err := out("gh", "api", "repos/{owner}/{repo}/statuses/"+h, "-f", "state="+st, "-f", "context=tend/review",
			"-f", "description=Independent review of "+short(h)+": "+c, "-f", "target_url="+pr); err != nil {
			stop("could not set the tend/review status on %s", short(h))
		}
		if h == head {
			markDraft(c != "success")
		}
		body := fmt.Sprintf("**Independent review of %s: %s**\n\n%s\n\n", short(h), c, str(get(x, "summary")))
		if c != "success" {
			if get(x, "actionable") == true {
				body += "The reviewer judged every finding an in-scope fix. tend repairs once, then a person decides.\n\n"
			} else {
				body += "A finding needs a decision from a person.\n\n"
			}
		}
		if reply := cat(run + "/reply-" + h); reply != "" {
			body += "The implementer, on the changes requested:\n\n> " + strings.ReplaceAll(reply, "\n", "\n> ") + "\n\n"
		}
		if h == head && kind == "grove" {
			o, _ := both("grove", "review")
			if i := regexp.MustCompile(`(?m)^diff --git`).FindStringIndex(o); i != nil {
				o = o[:i[0]]
			}
			if o = strings.TrimRight(o, "\n"); o != "" {
				body += "<details><summary>What merging approves, from grove review</summary>\n\n```text\n" + o + "\n```\n</details>\n\n"
			}
		}
		body += "Codex reviewed this commit in a session apart from the implementer; tend posted it for attempt `" + id + "`."
		if postJSON(reviewsAPI(), map[string]any{"commit_id": h, "event": "COMMENT", "body": body}) != nil {
			stop("could not post the review of %s on %s", short(h), pr)
		}
	}
	where = pr
	write(run+"/pr", pr+"\n") // for tend watch to find; GitHub stays the record
	if item.link != nil && item.link(pr) != nil {
		stop("could not say on %s where its pull request is (%s); run tend again to finish", item.ref, pr)
	}
	say("published %s", pr)
}

// A pull request whose head failed review is made a draft, and one whose head
// passed made ready, once per reviewed head, so a person may change it after.
// The description is the squash commit's, so it does not say; and where
// GitHub offers no drafts, the review and status still do.
func markDraft(want bool) {
	d, err := out("gh", "pr", "view", pr, "--json", "isDraft", "--jq", ".isDraft")
	if err != nil || (d == "true") == want {
		return
	}
	a, as := []string{"pr", "ready", pr}, "ready for review"
	if want {
		a, as = append(a, "--undo"), "a draft"
	}
	if _, err := out("gh", a...); err != nil {
		say("could not make %s %s; its review and status say how it stands", pr, as)
	}
}

// A change request the implementer answered with no commit: its reply goes on
// the pull request once, found by the request's marker, and answered-<node id>
// keeps tend and tend watch from asking again.
func answer(nid, who, reply string) {
	tag := "<!-- tend:answered " + nid + " -->"
	said, err := out("gh", "api", "repos/{owner}/{repo}/issues/"+prNumber()+"/comments", "--paginate", "--jq", fmt.Sprintf(`.[] | select(.user.login == "%s" and (.body | contains("%s"))) | .id`, account, tag))
	if err != nil {
		stop("could not read the comments on %s", pr)
	}
	if said == "" {
		body := fmt.Sprintf("**No change made for the changes %s requested at %s**\n\nThe implementer, on why:\n\n> %s\n\ntend will not ask again. To have it tried again, request changes saying what should change; or change the branch yourself.\n\n%s\n", who, short(head), strings.ReplaceAll(strings.TrimRight(reply, "\n"), "\n", "\n> "), tag)
		if _, err := input(body, "gh", "pr", "comment", pr, "--body-file", "-"); err != nil {
			stop("could not comment on %s; run tend again to finish", pr)
		}
	}
	write(run+"/answered-"+nid, head+"\n")
	say("%s made no change for %s's request; said why on %s", impl, who, pr)
}

// The commit message for the head, which the implementer writes to
// message-<head> once: the pull request's title and description. was, if
// given, is the message on the pull request now, kept where it still holds.
func message(was string) string {
	msg := run + "/message-" + head
	if nonempty(msg) {
		return msg
	}
	prev := ""
	if was != "" {
		prev = fmt.Sprintf(" The pull request's description now is in %s, written for an earlier head: keep what still holds and correct what later commits changed.", was)
	}
	implementer()
	stage("working", fmt.Sprintf("%s writes the commit message for %s", impl, short(head)), "tend puts it on the pull request")
	ask(impl, fmt.Sprintf("Branch %s is reviewed and will be squash-merged into %s: the pull request's title becomes the commit's subject and its description the commit's body. Write that commit message to %s and nothing else. A subject of at most 72 characters in this repository's commit style (see git log %s), a blank line, then a body wrapped at 72 columns saying what the change does and why, for someone reading the history later who was not here. Describe the merged result, not how it was reached: no reviews, repairs, agents, tools, attempt IDs or file lists. End with the line '%s'.%s Commit nothing and leave the branch as it is.", branch, target, msg, target, trailer, prev), "")
	if !nonempty(msg) {
		stop("%s settled without writing %s; look at its pane, then run tend again", impl, msg)
	}
	if git("rev-parse", "HEAD") != head || !clean() {
		stop("%s changed the branch while writing the commit message", impl)
	}
	return msg
}

func titleBody(msg string) (string, string) {
	lines := strings.Split(cat(msg), "\n")
	if len(lines) > 2 {
		return lines[0], strings.Join(lines[2:], "\n")
	}
	return lines[0], ""
}

// The title and description as GitHub may give them back, to compare.
func fold(title, body string) string {
	n := func(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n")) }
	return n(title) + "\n\n" + n(body)
}

// After later commits, the description is written again for the new head
// when it is still the message tend put there, the run folder's described
// naming that head; once a person has edited it, it is theirs and tend leaves it.
func redescribe(d string) {
	s, err := out("gh", "pr", "view", pr, "--json", "state,title,body")
	if err != nil {
		stop("could not read %s", pr)
	}
	x := parse(s)
	if str(get(x, "state")) != "OPEN" {
		return
	}
	now, was := fold(str(get(x, "title")), str(get(x, "body"))), run+"/message-"+d
	if nonempty(run+"/message-"+head) && now == fold(titleBody(run+"/message-"+head)) {
		write(run+"/described", head+"\n") // a run stopped after the edit
		return
	}
	if now != fold(titleBody(was)) {
		os.Remove(run + "/described")
		say("left %s's description as it is: a person has edited it", pr)
		return
	}
	title, body := titleBody(message(was))
	if fold(title, body) != now {
		if _, err := input(body, "gh", "pr", "edit", pr, "--title", title, "--body-file", "-"); err != nil {
			stop("could not rewrite %s's description", pr)
		}
		say("rewrote %s's description for %s", pr, short(head))
	}
	write(run+"/described", head+"\n")
}

// The newest change request on the open PR's head, or nothing.
func requested() (any, error) {
	s, err := out("gh", "pr", "view", pr, "--json", "state", "--jq", ".state")
	if err != nil || s != "OPEN" {
		return nil, nil
	}
	s, err = out("gh", "api", reviewsAPI(), "--paginate", "--jq", fmt.Sprintf(`[.[] | select(.state == "CHANGES_REQUESTED" and .commit_id == "%s")] | last // empty`, head))
	if err != nil || s == "" {
		return nil, err
	}
	return parse(s), nil
}
