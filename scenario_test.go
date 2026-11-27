// tend's acceptance tests: the built binary driven through its CLI with real
// Git and Grove, and stand-ins in testdata/bin for herdr, rex, gh and curl,
// whose agents act at once in tend's working directory. No live agent runs.
//
//	go test ./...                          every scenario, in parallel
//	go test -run 'TestTend/^remote-claim$' one by name
//
// TEND names another build to check instead.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

var tendBin, standIns, binDir string

func TestMain(m *testing.M) {
	standIns, _ = filepath.Abs("testdata/bin")
	// Every scenario starts hundreds of processes: by default half the cores
	// run scenarios, so the machine stays usable, and on CI (CI set) twice as
	// many scenarios as cores, as each spends much of its time waiting on the
	// processes it starts; -parallel overrides it.
	flag.Parse()
	set := false
	flag.Visit(func(f *flag.Flag) { set = set || f.Name == "test.parallel" })
	if !set {
		n := max(1, runtime.NumCPU()/2)
		if os.Getenv("CI") != "" {
			n = 2 * runtime.NumCPU()
		}
		flag.Set("test.parallel", strconv.Itoa(n))
	}
	d, err := os.MkdirTemp("", "tend-bin")
	if err != nil {
		panic(err)
	}
	binDir = d
	// grove as a mise shim costs a second process on each of thousands of
	// calls: the scenarios get the grove that the shim falls back to.
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if g := p + "/grove"; !strings.Contains(p, "mise/shims") && exists(g) {
			os.Symlink(g, d+"/grove")
			break
		}
	}
	tendBin = os.Getenv("TEND")
	if tendBin == "" {
		tendBin = d + "/tend"
		if b, err := exec.Command("go", "build", "-o", tendBin, ".").CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "go build: %v\n%s", err, b)
			os.Exit(1)
		}
	}
	code := m.Run()
	os.RemoveAll(d)
	os.Exit(code)
}

// ---- a scenario: its own folder, repository and stand-in state

type sc struct {
	t    *testing.T
	dir  string // $t in dev/test.sh: the scenario's folder
	repo string // dir/repo
	st   string // HERDR_STATE, the stand-ins' state
	ts   string // TEND_STATE
}

func newSc(t *testing.T) *sc {
	t.Parallel()
	d := t.TempDir()
	s := &sc{t: t, dir: d, repo: d + "/repo", st: d + "/state", ts: d + "/tend"}
	if err := os.MkdirAll(s.st, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { // an agent that outlived tend, as the lock scenarios leave
		if pid, err := strconv.Atoi(s.cat(s.st + "/sleeper")); err == nil {
			syscall.Kill(pid, syscall.SIGKILL)
		}
	})
	return s
}

var dropped = regexp.MustCompile(`^(TEND_|HERDR_|REX_|JIRA_|GIT_|GH_|WATCH_|NO_NOTES|REVIEW|IMPL_|REPAIR_FAIL|REV_FAIL|BLOCK_AFTER_BUILD|CREATE_LOST|STATUS_FAIL|REQUEST_|PROPOSE_FAIL|FOLLOW_FAIL|DECLINE|EDIT_FAIL|CAP|TRUST_|DRAFT_FAIL|COMMENT_FAIL|STRAY|CHECK|UNFINISHED|TIMEOUT_MS|KEEP|AT_ONCE|HOLD|WORKTREE_HOLD|CI_LATE|BEFORE_MERGE|REQUIRE_APPROVAL|MERGE_HOLD|AGENT_GET_FAIL|SILENT|PERMISSION|VISIBILITY)`)

// The environment every command of the scenario gets, then extra ("K=V").
func (s *sc) env(extra ...string) []string {
	var e []string
	for _, kv := range os.Environ() {
		if !dropped.MatchString(kv) {
			e = append(e, kv)
		}
	}
	e = append(e, "PATH="+standIns+":"+binDir+":"+os.Getenv("PATH"),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "TEND_TERMINAL=herdr", "TEND_RETRY_MS=20",
		"HERDR_STATE="+s.st, "TEND_STATE="+s.ts, "JIRA_EMAIL=owner@example.test", "JIRA_API_TOKEN=tok-secret-1", "t="+s.dir)
	return append(e, extra...)
}

func (s *sc) cmd(dir string, env []string, name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Dir, c.Env = dir, s.env(env...)
	return c
}

func (s *sc) check(what string, ok bool) {
	s.t.Helper()
	if !ok {
		s.t.Errorf("FAIL %s", what)
	}
}

// What one run of tend left: its exit status, standard output and error.
type res struct {
	code     int
	out, err string
}

func (r res) said(re string) bool { return regexp.MustCompile(`(?m)` + re).MatchString(r.err) }
func (r res) conclusion() string {
	var v struct{ Conclusion string }
	json.Unmarshal([]byte(r.out), &v)
	return v.Conclusion
}
func (r res) passed() bool { return r.code == 0 && r.conclusion() == "success" }

// tend with extra env ("K=V") and arguments, from the scenario's folder.
func (s *sc) tendIn(env []string, args ...string) res {
	c := s.cmd(s.dir, env, tendBin, args...)
	var o, e bytes.Buffer
	c.Stdout, c.Stderr = &o, &e
	err := c.Run()
	code := 0
	if x, ok := err.(*exec.ExitError); ok {
		code = x.ExitCode()
	} else if err != nil {
		s.t.Fatal(err)
	}
	return res{code, o.String(), e.String()}
}

func (s *sc) run(env ...string) res     { return s.tendIn(env, s.repo, "261114-k0000") }
func (s *sc) rexrun(env ...string) res  { return s.run(append(env, "TEND_TERMINAL=rex")...) }
func (s *sc) propose(env ...string) res { return s.tendIn(env, "propose", s.repo) }
func (s *sc) watch(env ...string) res   { return s.tendIn(append(env, "WATCH_ONCE=1"), "watch") }
func (s *sc) status() string            { return s.tendIn(nil, "status", s.repo, "261114-k0000").out }
func (s *sc) overview() string { // tend status <repo>, both streams as 2>&1
	r := s.tendIn(nil, "status", s.repo)
	return r.out + r.err
}

// tend started in the background, its output in files so an agent that
// outlives it holds no pipe of ours; returned once the implementer holds its
// first prompt, under IMPL_SLEEP, which every caller sets. The stand-in logs
// the prompt before it writes the agent's status, pending turn and sleeper,
// so the sleeper file, written last, is what is waited for.
func (s *sc) background(env ...string) *exec.Cmd {
	c := s.cmd(s.dir, env, tendBin, s.repo, "261114-k0000")
	f, err := os.Create(s.dir + "/bg")
	if err != nil {
		s.t.Fatal(err)
	}
	defer f.Close()
	c.Stdout, c.Stderr = f, f
	if err := c.Start(); err != nil {
		s.t.Fatal(err)
	}
	for range 100 {
		if exists(s.st + "/sleeper") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return c
}

// A tend in the background with extra env and arguments, its output in
// files; returned once the stand-ins' log has a line matching until.
type bgTend struct {
	s   *sc
	c   *exec.Cmd
	out string
}

func (s *sc) bg(env []string, until string, args ...string) *bgTend {
	s.t.Helper()
	f, err := os.CreateTemp(s.dir, "bg-")
	if err != nil {
		s.t.Fatal(err)
	}
	defer f.Close()
	c := s.cmd(s.dir, env, tendBin, args...)
	c.Stdout, c.Stderr = f, f
	if err := c.Start(); err != nil {
		s.t.Fatal(err)
	}
	b := &bgTend{s, c, f.Name()}
	s.t.Cleanup(func() { c.Process.Kill(); c.Wait() })
	s.waitFor(func() bool { return s.has(s.log(), until) }, until)
	return b
}

// Its exit status and output, once it exits; its JSON is the output's last
// lines from the first "{" on a line of its own.
func (b *bgTend) wait() res {
	err := b.c.Wait()
	code := 0
	if x, ok := err.(*exec.ExitError); ok {
		code = x.ExitCode()
	}
	o := b.s.cat(b.out)
	j := ""
	if i := strings.LastIndex(o, "\n{\n"); i >= 0 {
		j = o[i+1:]
	}
	return res{code, j, o}
}

// Waits up to 30 seconds for ok, failing the scenario after.
func (s *sc) waitFor(ok func() bool, what string) {
	s.t.Helper()
	for range 300 {
		if ok() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	s.t.Fatalf("waited for %s", what)
}

// The file a held stand-in waits for, touched to release it.
func (s *sc) hold() (string, func()) {
	f := s.dir + fmt.Sprintf("/hold-%d", time.Now().UnixNano())
	return f, func() { s.touch(f) }
}

// The worktrees of the repository, but its own checkout.
func (s *sc) worktrees() []string {
	var ws []string
	for _, l := range strings.Split(s.git(s.repo, "worktree", "list", "--porcelain"), "\n") {
		if p, ok := strings.CutPrefix(l, "worktree "); ok && !sameDir(p, s.repo) {
			ws = append(ws, p)
		}
	}
	return ws
}

// A bash script with the scenario's env, in dir, its stdout trimmed; a
// failure ends the scenario. The setup helpers below are defined for it;
// args are $0, $1, ...
func (s *sc) bash(dir, script string, args ...string) string {
	s.t.Helper()
	c := s.cmd(dir, nil, "bash", append([]string{"-c", "set -euo pipefail\n" + lib + script}, args...)...)
	var e bytes.Buffer
	c.Stderr = &e
	o, err := c.Output()
	if err != nil {
		s.t.Fatalf("%v: %s\n%s", err, script, e.String())
	}
	return strings.TrimRight(string(o), "\n")
}

// A command's output, both streams, trailing newlines trimmed, as "$(... 2>&1)".
func (s *sc) sh(dir, name string, args ...string) string {
	o, _ := s.cmd(dir, nil, name, args...).CombinedOutput()
	return strings.TrimRight(string(o), "\n")
}
func (s *sc) git(dir string, args ...string) string { return s.sh(dir, "git", args...) }
func (s *sc) gitOK(dir string, args ...string) bool {
	return s.cmd(dir, nil, "git", args...).Run() == nil
}

const lib = `
# A Grove project with one approved change in cycle 261114-k0000, on branch
# $2 (default main), and with origin when $3 is "remote".
project() { # dir [target] [remote]
	local d=$1 target=${2:-main}
	git init -q -b "$target" "$d" && cd "$d"
	echo x >a.txt && git add -A && git commit -qm init
	grove init >/dev/null
	sed "s/^target: .*/target: $target/" grove.yaml >g && mv g grove.yaml
	git add -A && git commit -qm grove
	local c=grove/changes/261015-c0000-a-change
	mkdir -p "$c" grove/cycles
	printf 'title: A change\ncreated: 2026-10-15T09:00:00Z\nschema: default\nbase: %s\n' "$(git rev-parse HEAD)" >"$c/.grove.yaml"
	printf '# A change\n\n## Why\n\nBecause.\n\n## What changes\n\nA file.\n' >"$c/proposal.md"
	printf '# Tasks\n\n- [ ] T-01 Add the file\n' >"$c/tasks.md"
	git add -A && git commit -qm propose
	printf -- '---\ntitle: "One"\ncreated: 2026-11-14T02:00:00Z\nchanges: [261015-c0000]\nplanned: %s\n---\n' "$(git rev-parse HEAD)" >grove/cycles/261114-k0000-one.md
	git add -A && git commit -qm plan
	if [ "${3:-}" = remote ]; then
		git clone -q --bare . "$d.git" && git remote add origin "$d.git" && git fetch -q origin && git branch -q -u "origin/$target"
	fi
	cd - >/dev/null
}
# An approved change in no cycle, created on the 16th or the day given,
# committed, and pushed when there is origin. With a file, its one task plans
# that file, and implementation.md has it done, so that it archives.
change() { # id title [depends-on] [planned file] [day]
	local c=grove/changes/$1-$(tr 'A-Z ' 'a-z-' <<<"$2")
	mkdir -p "$c"
	printf 'title: %s\ncreated: 2026-10-%sT09:00:00Z\nschema: default\nbase: %s\n' "$2" "${5:-16}" "$(git rev-parse HEAD)" >"$c/.grove.yaml"
	[ -z "${3:-}" ] || printf 'depends_on: [%s]\n' "$3" >>"$c/.grove.yaml"
	printf '# %s\n\n## Why\n\nBecause.\n' "$2" >"$c/proposal.md"
	if [ -n "${4:-}" ]; then
		printf '## T-01 — Do it\n\nResult: done.\n\nFiles:\n- \x60%s\x60\n' "$4" >"$c/tasks.md"
		printf '# Implementation\n\n## Tasks\n\n### T-01 — Do it\n\nStatus: done\n' >"$c/implementation.md"
	else
		printf '# Tasks\n\n- [ ] T-01 Do it\n' >"$c/tasks.md"
	fi
	git add -A && git commit -qm "$2"
	if git remote get-url origin >/dev/null 2>&1; then git push -q origin HEAD; fi
}
# A repository without Grove, with origin.
plain() {
	git init -q -b main "$t/repo" && (cd "$t/repo" && echo x >a.txt && git add -A && git commit -qm init &&
		git clone -q --bare . "$t/repo.git" && git remote add origin "$t/repo.git" && git fetch -q origin && git branch -q -u origin/main)
}
`

func (s *sc) project(target string, remote bool) {
	r := ""
	if remote {
		r = "remote"
	}
	s.bash(s.dir, fmt.Sprintf("project %q %q %q", s.repo, target, r))
}
func (s *sc) plain() { s.bash(s.dir, "plain") }
func (s *sc) change(id, title, dep string, fileDay ...string) {
	s.bash(s.repo, fmt.Sprintf("change %q %q %q", id, title, dep)+" "+strings.Join(fileDay, " "))
}
func (s *sc) crun(id string, env ...string) res { return s.tendIn(env, s.repo, id) } // tend <repo> <change-id>

// Files written into the repository, committed on its branch, and pushed when
// it has origin: as the target holds them, e.g. a tend/flow.yaml.
func (s *sc) commit(msg string, files map[string]string) {
	for p, body := range files {
		s.write(s.repo+"/"+p, body)
	}
	s.bash(s.repo, fmt.Sprintf("git add -A && git commit -qm %q && if git remote get-url origin >/dev/null 2>&1; then git push -q origin HEAD; fi", msg))
}

// ---- files

func (s *sc) cat(f string) string { b, _ := os.ReadFile(f); return strings.TrimRight(string(b), "\n") }
func (s *sc) write(f, body string) {
	os.MkdirAll(filepath.Dir(f), 0o755)
	if err := os.WriteFile(f, []byte(body), 0o644); err != nil {
		s.t.Fatal(err)
	}
}
func (s *sc) touch(f string) { s.write(f, "") }
func isDir(f string) bool    { st, err := os.Stat(f); return err == nil && st.IsDir() } // exists and nonempty are main's

// The first path matching the pattern, or the pattern itself, as an
// unmatched shell glob is left.
func glob(p string) string {
	if m, _ := filepath.Glob(p); len(m) > 0 {
		return m[0]
	}
	return p
}

// Lines of text matching the regular expression.
func (s *sc) count(text, re string) int {
	n, r := 0, regexp.MustCompile(re)
	for _, l := range strings.Split(text, "\n") {
		if r.MatchString(l) {
			n++
		}
	}
	return n
}
func (s *sc) has(text, re string) bool { return s.count(text, re) > 0 }

func (s *sc) readJSON(f string) any {
	b, _ := os.ReadFile(f)
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var x any
	d.Decode(&x)
	return x
}
func (s *sc) writeJSON(f string, x any) {
	b, _ := json.Marshal(x)
	s.write(f, string(b)+"\n")
}

// ---- what the stand-ins recorded

func (s *sc) log() string    { return s.cat(s.st + "/log") }
func (s *sc) rexlog() string { return s.cat(s.st + "/rexlog") }

// Prompts to agents whose name starts agent, whose text starts text.
func (s *sc) promptsTo(agent, text string) int {
	return s.count(s.log(), `^agent prompt `+regexp.QuoteMeta(agent)+`[^ ]* `+regexp.QuoteMeta(text))
}
func (s *sc) prompts(text string) int { return s.promptsTo("impl-", text) } // implementer prompts starting text
func (s *sc) reviews() int            { return s.count(s.log(), `^agent prompt rev-`) }
func (s *sc) creates() int            { return s.count(s.log(), `^gh pr create`) }
func (s *sc) drafts() int             { return s.count(s.log(), `^agent prompt prop-[^ ]* /grove-propose`) }
func (s *sc) follows() int            { return s.count(s.log(), `^agent prompt prop-[^ ]* Without asking anyone`) }

// A second repository on the same machine, with its own GitHub and tend
// state, in the same Herdr: the agents open in s are open for it too.
func (s *sc) other() *sc {
	d := s.dir + "/other"
	o := &sc{t: s.t, dir: d, repo: d + "/repo", st: d + "/state", ts: d + "/tend"}
	open, _ := filepath.Glob(s.st + "/agent-*")
	for _, a := range open {
		o.touch(o.st + "/" + filepath.Base(a))
	}
	return o
}

// An agent's state in the stand-ins, by the prefix of its name.
func (s *sc) agent(status, prefix string) {
	a := filepath.Base(glob(s.st + "/agent-" + prefix + "*"))
	s.write(s.st+"/status-"+strings.TrimPrefix(a, "agent-"), status+"\n")
}

// An agent's pane closed, by the prefix of its name: the stand-ins then know
// no agent of that name, nor any turn it had pending.
func (s *sc) gone(prefix string) {
	for _, k := range []string{"agent-", "status-", "pending-"} {
		fs, _ := filepath.Glob(s.st + "/" + k + prefix + "*")
		for _, f := range fs {
			os.Remove(f)
		}
	}
}

func (s *sc) records() int {
	x := parse(s.sh(s.repo, "grove", "changes", "--json"))
	n := 0
	for _, r := range list(get(x, "runs")) {
		if str(get(r, "cycle")) == "261114-k0000" {
			n++
		}
	}
	return n
}

// Grove's records naming change c alone and no cycle.
func (s *sc) changeRecords(c string) int {
	n := 0
	for _, r := range list(get(parse(s.sh(s.repo, "grove", "changes", "--json")), "runs")) {
		if get(r, "cycle") == nil && slices.Equal(list(get(r, "changes")), []any{c}) {
			n++
		}
	}
	return n
}
func (s *sc) worktree() string { return glob(s.st + "/wt-tend-[0-9]*") }
func (s *sc) head() string     { return s.git(s.worktree(), "rev-parse", "HEAD") }
func (s *sc) branch() string   { return s.git(s.worktree(), "branch", "--show-current") }
func (s *sc) pr() string       { return s.st + "/pr-" + strings.ReplaceAll(s.branch(), "/", "_") }
func (s *sc) runDir() string   { return glob(s.ts + "/runs/*") }

// Let tend publish to the repository's origin, as the-bot.
func (s *sc) allow() {
	f, _ := os.OpenFile(s.ts+"/publish", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if f == nil {
		os.MkdirAll(s.ts, 0o755)
		f, _ = os.OpenFile(s.ts+"/publish", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	}
	fmt.Fprintf(f, "%s the-bot\n", s.git(s.repo, "remote", "get-url", "origin"))
	f.Close()
}

func (s *sc) reviewList() []any { return list(s.readJSON(s.st + "/reviews")) }

// How many reviews name that head.
func (s *sc) reviewed(h string) int {
	n := 0
	for _, r := range s.reviewList() {
		if str(get(r, "commit_id")) == h {
			n++
		}
	}
	return n
}

// A teammate, a collaborator, requests changes on the PR's head, with a
// comment on a line.
func (s *sc) request() { s.requestAs("teammate", "COLLABORATOR", "CHANGES_REQUESTED") }

// A review in that state on the PR's head by login, whom GitHub gives that
// author association.
func (s *sc) requestAs(login, association, state string) {
	s.reviewOn(s.head(), login, association, state)
}

// A review in that state on commit h by login.
func (s *sc) reviewOn(h, login, association, state string) {
	rs := s.reviewList()
	n := 900 + len(rs)
	rs = append(rs, map[string]any{"id": n, "node_id": fmt.Sprintf("PRR_%d", n), "state": state, "commit_id": h,
		"user": map[string]any{"login": login}, "author_association": association, "body": "Say it plainly.",
		"comments": []any{map[string]any{"path": "a.txt", "line": 1, "body": "Plainer."}}})
	s.writeJSON(s.st+"/reviews", rs)
}

// ---- intake

// Issues in the stand-in's GitHub, as "number:labels,...".
func (s *sc) issues(specs ...string) {
	is := []any{}
	for _, sp := range specs {
		n, l, _ := strings.Cut(sp, ":")
		labels := []any{}
		for _, x := range strings.Split(l, ",") {
			if x != "" {
				labels = append(labels, x)
			}
		}
		num, _ := strconv.Atoi(n)
		is = append(is, map[string]any{"number": num, "title": "Request " + n, "body": "Make thing " + n + " better.",
			"url": "https://example.test/issues/" + n, "state": "OPEN", "labels": labels})
	}
	s.writeJSON(s.st+"/issues", is)
}
func (s *sc) prOf(n string) string { return s.st + "/pr-tend_issue-" + n }

// Comments on issue n.
func (s *sc) comments(n int) int {
	c := 0
	for _, x := range list(s.readJSON(s.st + "/comments")) {
		if v, ok := get(x, "issue").(json.Number); ok && v.String() == strconv.Itoa(n) {
			c++
		}
	}
	return c
}

// ---- other helpers

// Under rex: the worktree on a grove/ branch.
func (s *sc) rexwt() string { return s.worktreeOf("refs/heads/grove/") }

func (s *sc) worktreeOf(prefix string) string {
	w := ""
	for _, l := range strings.Split(s.git(s.repo, "worktree", "list", "--porcelain"), "\n") {
		if p, ok := strings.CutPrefix(l, "worktree "); ok {
			w = p
		} else if b, ok := strings.CutPrefix(l, "branch "); ok && (b == prefix || strings.HasSuffix(prefix, "/") && strings.HasPrefix(b, prefix)) {
			return w
		}
	}
	return ""
}

// A commit someone else pushed to origin's main: its hash.
func (s *sc) pushed() string {
	return s.bash(s.dir, `git clone -q "$t/repo.git" "$t/other" && (cd "$t/other" && echo theirs >b.txt && git add b.txt && git commit -qm theirs && git push -q origin main) && git -C "$t/other" rev-parse HEAD`)
}

// Header's line and the n after it, as grep -A.
func after(text, header string, n int) string {
	ls := strings.Split(text, "\n")
	for i, l := range ls {
		if strings.HasPrefix(l, header) {
			return strings.Join(ls[i:min(len(ls), i+n+1)], "\n")
		}
	}
	return ""
}

func (s *sc) cksum(p string) string {
	c := exec.Command("cksum")
	c.Stdin = strings.NewReader(p)
	o, _ := c.Output()
	f := strings.Fields(string(o))
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

func killed(c *exec.Cmd) {
	c.Process.Kill()
	c.Wait()
}

// ---- the scenarios, each list in its behavior's test file

type scenario struct {
	name string
	fn   func(s *sc)
}

func TestTend(t *testing.T) {
	t.Parallel()
	for _, x := range scenarios {
		t.Run(x.name, func(t *testing.T) { x.fn(newSc(t)) })
	}
}

var scenarios = slices.Concat(cliScenarios, attemptScenarios, worktreeScenarios, watchScenarios, statusScenarios, publicationScenarios, intakeScenarios, jiraScenarios, terminalScenarios, headlessScenarios)

var cliScenarios = []scenario{
	{"version", func(s *sc) {
		s.check("tend version names the build", strings.TrimSpace(s.tendIn(nil, "version").out) == "tend dev")
	}},
	{"usage", func(s *sc) {
		r := s.tendIn(nil, s.dir)
		s.check("tend with no work item says how to run it", r.code == 2 && r.said(`^usage: tend <repo> <id>$`))
	}},
}
