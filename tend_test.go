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
	// run scenarios, so the machine stays usable; -parallel overrides it.
	flag.Parse()
	set := false
	flag.Visit(func(f *flag.Flag) { set = set || f.Name == "test.parallel" })
	if !set {
		flag.Set("test.parallel", strconv.Itoa(max(1, runtime.NumCPU()/2)))
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

var dropped = regexp.MustCompile(`^(TEND_|HERDR_|REX_|JIRA_|GIT_|GH_|WATCH_|NO_NOTES|REVIEW|IMPL_|REPAIR_FAIL|REV_FAIL|BLOCK_AFTER_BUILD|CREATE_LOST|REQUEST_|PROPOSE_FAIL|FOLLOW_FAIL|DECLINE|EDIT_FAIL|CAP|TRUST_|DRAFT_FAIL|COMMENT_FAIL|STRAY|CHECK|UNFINISHED|TIMEOUT_MS|KEEP|AT_ONCE|HOLD|WORKTREE_HOLD|CI_LATE|BEFORE_MERGE|REQUIRE_APPROVAL|MERGE_HOLD)`)

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
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "TEND_TERMINAL=herdr",
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
func (s *sc) build(env ...string) res   { return s.tendIn(env, s.repo, "issue-7") }
func (s *sc) jbuild(env ...string) res  { return s.tendIn(env, s.repo, "SCRUM-7") }
func (s *sc) propose(env ...string) res { return s.tendIn(env, "propose", s.repo) }
func (s *sc) watch(env ...string) res   { return s.tendIn(append(env, "WATCH_ONCE=1"), "watch") }
func (s *sc) watchrepo(env ...string) res {
	return s.tendIn(append(env, "WATCH_ONCE=1"), "watch", s.repo)
}
func (s *sc) status() string { return s.tendIn(nil, "status", s.repo, "261114-k0000").out }
func (s *sc) overview() string { // tend status <repo>, both streams as 2>&1
	r := s.tendIn(nil, "status", s.repo)
	return r.out + r.err
}

// tend started in the background, its output in files so an agent that
// outlives it holds no pipe of ours; with the first implementer prompt sent.
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
		if s.count(s.log(), `^agent prompt impl-`) > 0 {
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

// A project whose target's flow file sets watch.at_once to n, listed for
// publication, with no requests.
func (s *sc) atOnce(n int) {
	s.project("main", true)
	s.allow()
	s.issues()
	s.commit("Keep some going", map[string]string{"tend/flow.yaml": fmt.Sprintf("watch:\n  at_once: %d\n", n)})
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

// An agent's state in the stand-ins, by the prefix of its name.
func (s *sc) agent(status, prefix string) {
	a := filepath.Base(glob(s.st + "/agent-" + prefix + "*"))
	s.write(s.st+"/status-"+strings.TrimPrefix(a, "agent-"), status+"\n")
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
func (s *sc) verdict() string  { return glob(s.ts + "/runs/*/verdict-*.json") }
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
func (s *sc) labelled(n int) bool {
	for _, i := range list(s.readJSON(s.st + "/issues")) {
		if str(get(i, "number")) == strconv.Itoa(n) {
			return slices.Contains(list(get(i, "labels")), any("tend"))
		}
	}
	return false
}

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

// ---- Jira, through the curl stand-in

// Map SCRUM to the repository's origin.
func (s *sc) jmap() {
	os.MkdirAll(s.ts, 0o755)
	f, _ := os.OpenFile(s.ts+"/jira", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	fmt.Fprintf(f, "https://jira.test SCRUM %s\n", s.git(s.repo, "remote", "get-url", "origin"))
	f.Close()
}

func adf(text string) any {
	return map[string]any{"type": "doc", "version": 1, "content": []any{map[string]any{"type": "paragraph", "content": []any{map[string]any{"type": "text", "text": text}}}}}
}

// An issue: key, labels, status category, a description in ADF ("" for a
// plain one) and how many comments it has.
func (s *sc) jissue(key, labels, status, desc string, comments int) {
	js := list(s.readJSON(s.st + "/jira"))
	var d any = adf("Make " + key + " better.")
	if desc != "" {
		d = parse(desc)
	}
	ls := []any{}
	for _, x := range strings.Split(labels, ",") {
		if x != "" {
			ls = append(ls, x)
		}
	}
	cs := []any{}
	for i := range comments {
		cs = append(cs, map[string]any{"body": adf(fmt.Sprintf("Earlier comment %d", i))})
	}
	js = append(js, map[string]any{"key": key, "comments": cs, "fields": map[string]any{"summary": "Request " + key, "description": d,
		"labels": ls, "status": map[string]any{"statusCategory": map[string]any{"key": status}}}})
	s.writeJSON(s.st+"/jira", js)
}

// The text of an issue's comments, one paragraph a line.
func (s *sc) jcomments(key string) string {
	var ls []string
	for _, x := range list(s.readJSON(s.st + "/jira")) {
		if str(get(x, "key")) != key {
			continue
		}
		for _, c := range list(get(x, "comments")) {
			for _, p := range list(get(c, "body", "content")) {
				for _, t := range list(get(p, "content")) {
					if v, ok := get(t, "text").(string); ok {
						ls = append(ls, v)
					}
				}
			}
		}
	}
	return strings.Join(ls, "\n")
}

const rich = `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Implement "},{"type":"text","text":"the spec","marks":[{"type":"link","attrs":{"href":"https://docs.test/spec"}}]},{"type":"text","text":" and "},{"type":"inlineCard","attrs":{"url":"https://jira.test/browse/SCRUM-9"}}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"keep totals"}]}]}]},{"type":"codeBlock","content":[{"type":"text","text":"ledger list --json"}]}]}`

// ---- other helpers

// The worktree of the issue build's branch.
func (s *sc) bworktree() string { return s.worktreeOf("refs/heads/tend/build-7") }

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

// The cycle marked unattended in its plan, in the checkout dir, pushed.
func (s *sc) unattend(dir string) {
	f := dir + "/grove/cycles/261114-k0000-one.md"
	s.write(f, regexp.MustCompile(`(?m)^planned: .*\n`).ReplaceAllString(s.cat(f)+"\n", "${0}unattended: true\n"))
	s.bash(dir, "git commit -qam unattend && git push -q origin main")
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

// ---- the scenarios, as dev/test.sh had them, in its order

type scenario struct {
	name string
	fn   func(s *sc)
}

func TestTend(t *testing.T) {
	for _, x := range scenarios {
		t.Run(x.name, func(t *testing.T) { x.fn(newSc(t)) })
	}
}

var scenarios = []scenario{
	{"version", func(s *sc) {
		s.check("tend version names the build", strings.TrimSpace(s.tendIn(nil, "version").out) == "tend dev")
	}},
	{"remote-claim", func(s *sc) {
		s.project("main", true)
		r := s.run()
		s.check("claims with origin and reaches a verdict", r.code == 0)
		s.check("the verdict says success", r.conclusion() == "success")
		s.check("with no Git excludes, the candidate stays clean", s.git(s.worktree(), "status", "--porcelain") == "")
		s.check("the verdict is kept outside the candidate", exists(s.verdict()))
		s.check("the reviewer may write only there besides its checkout", strings.Contains(s.log(), "--add-dir "+s.ts+"/runs/"))
		s.check("the branch made for the worktree goes after the claim", s.git(s.repo, "branch", "--list", "tend-*") == "")
	}},
	{"remote-resume", func(s *sc) {
		s.project("main", true)
		r := s.run("IMPL_FAIL=stub_failure")
		s.check("an implementer failure after the claim stops for a person", r.code == 3)
		id := str(get(parse(s.sh(s.repo, "grove", "changes", "--json")), "runs", 0, "id"))
		r = s.run()
		s.check("the rerun exits 0", r.code == 0)
		s.check("the rerun carries on the same attempt", r.said("carrying on attempt "+id))
		s.check("one record for the cycle", s.records() == 1)
		s.check("an origin not listed for publication gets no push or pull request", s.count(s.log(), `^gh `) == 0)
		s.check("and tend says why", r.said("is not listed in"))
		s.check("and its candidate branch is not pushed", s.git(s.repo+".git", "branch", "--list", "grove/2*") == "")
	}},
	{"runner", func(s *sc) {
		s.project("main", true)
		s.run()
		s.check("the implementer is asked for the cycle under --runner", s.has(s.log(), `^agent prompt impl-[^ ]* /grove-implement --runner 261114-k0000 --wait`))
		start := strings.Join(slices.DeleteFunc(strings.Split(s.log(), "\n"), func(l string) bool { return !strings.HasPrefix(l, "agent start impl-") }), "\n")
		s.check("pushing and gh pr are denied to it", strings.Contains(start, "--disallowedTools Bash(git push:*) Bash(gh pr:*)"))
		s.check("it is told tend delivers, on its branch", strings.Contains(start, "--append-system-prompt This session works for tend on branch "+s.branch()+" in this worktree. Commit your work on this branch. Push nothing"))
	}},
	{"stray-worktree", func(s *sc) {
		s.project("main", true)
		r := s.run("STRAY=worktree")
		s.check("an implementer that adds a worktree stops for a person", r.code == 3)
		s.check("naming it", r.said("it added the worktree .*stray-impl-"))
		s.check("and nothing is reviewed", s.reviews() == 0)
		r = s.run()
		s.check("a rerun stops again, from the same record", r.code == 3)
		s.check("still with nothing reviewed", s.reviews() == 0)
	}},
	{"stray-branch", func(s *sc) {
		s.project("main", true)
		r := s.run("STRAY=branch")
		s.check("an implementer that makes a branch with commits stops for a person", r.code == 3)
		s.check("naming it", r.said("it made the branch stray-impl-.*, with commits"))
		s.check("and nothing is reviewed", s.reviews() == 0)
	}},
	{"local-resume", func(s *sc) {
		s.project("main", false)
		r := s.run("IMPL_FAIL=stub_failure")
		s.check("without origin, a failure after the claim stops", r.code == 3)
		r = s.run()
		s.check("the rerun carries on and exits 0", r.code == 0)
		s.check("one record for the cycle", s.records() == 1)
		s.check("without origin, nothing is published", s.count(s.log(), `^gh `) == 0)
	}},
	{"publish", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run()
		s.check("with origin listed, the run reaches a verdict", r.passed())
		s.check("the reviewed head is pushed to origin", s.git(s.repo+".git", "rev-parse", s.branch()) == s.head())
		s.check("one pull request is opened, onto the target", s.count(s.log(), `^gh pr create --base main `) == 1)
		s.check("by the listed account, not the owner", s.cat(s.pr()+".by") == "the-bot")
		s.check("the implementer was asked for the commit message once", s.prompts("Branch") == 1)
		s.check("titled with its subject", s.has(s.log(), `^gh pr create .*--title Add the file `))
		s.check("described by its body alone", s.cat(s.pr()) == "The project needed a file.\n\nGrove: 261015-c0000")
		s.check("the implementer may write in the run folder", s.has(s.log(), `^agent start impl-.*--add-dir `+regexp.QuoteMeta(s.ts+"/runs/")))
		s.check("the reviewer has the network, to serve on localhost", s.has(s.log(), `^agent start rev-.*-c sandbox_workspace_write\.network_access=true -a never`))
		rv := s.cat(s.st + "/reviews")
		s.check("the head's review shows what merging approves", strings.Contains(rv, "What merging approves"))
		n := 0
		for _, x := range s.reviewList() {
			if str(get(x, "commit_id")) == s.head() && str(get(x, "event")) == "COMMENT" && str(get(x, "user", "login")) == "the-bot" {
				n++
			}
		}
		s.check("the head gets one review, a comment by the account", n == 1)
		s.check("saying it passed", s.has(rv, `review of .*: success`))
		s.check("the head carries the tend/review status, set by the account", s.has(s.cat(s.st+"/statuses"), `^`+s.head()+` success the-bot$`))
		s.check("status points at the pull request", strings.Contains(s.status(), "a person reviews https://example.test/pr/"))
		s.check("a passing head leaves it ready", !exists(s.pr()+".draft") && s.count(s.log(), `^gh pr ready`) == 0)
		s.write(s.pr(), "A person's description.\n")
		s.run()
		s.check("a repeated run opens no second pull request", s.creates() == 1)
		s.check("leaves the description to people", s.cat(s.pr()) == "A person's description.")
		s.check("and asks for no second commit message", s.prompts("Branch") == 1)
		s.check("and posts no second review of the head", s.reviewed(s.head()) == 1)
	}},
	{"publish-lost", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run("CREATE_LOST=1")
		s.check("a pull request opened on GitHub but not acknowledged stops", r.code == 3)
		s.check("and says a rerun looks first", r.said("looks for one before opening another"))
		r = s.run()
		s.check("the rerun finishes", r.passed())
		s.check("without opening another", s.creates() == 1)
		s.check("and posts the review the first run did not", s.reviewed(s.head()) == 1)
		s.check("the commit message was asked for once", s.prompts("Branch") == 1)
	}},
	{"publish-repair", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run("REVIEW=repairable")
		s.check("a run with a repair passes", r.passed())
		n := 0
		for _, x := range s.reviewList() {
			if strings.Contains(str(get(x, "body")), "failure") {
				n++
			}
		}
		s.check("the failed head's review is posted on that commit", n == 1)
		s.check("and says tend repairs once", strings.Contains(s.cat(s.st+"/reviews"), "tend repairs once"))
		s.check("as is the repaired head's", s.reviewed(s.head()) == 1)
		st := s.cat(s.st + "/statuses")
		s.check("each head has its own status", s.count(st, ` failure the-bot$`) == 1 && s.count(st, `^`+s.head()+` success`) == 1)
	}},
	{"publish-new-head", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		first := s.head()
		s.bash(s.worktree(), `echo more >>a.txt && git commit -qam "A later change"`)
		r := s.run("REVIEW=decision")
		st := s.cat(s.st + "/statuses")
		s.check("a later commit is reviewed and published", r.code == 0)
		s.check("its failure goes on the new head", s.has(st, `^`+s.head()+` failure the-bot$`))
		s.check("the old head's pass is not the new head's", s.count(st, `^`+s.head()+` success`) == 0)
		s.check("its review says it needs a decision", strings.Contains(s.cat(s.st+"/reviews"), "needs a decision"))
		s.check("the earlier head keeps its one review", s.reviewed(first) == 1)
		s.check("still one pull request", s.creates() == 1)
		s.check("a failed head makes the pull request a draft", exists(s.pr()+".draft"))
		s.bash(s.worktree(), `echo fixed >>a.txt && git commit -qam "Fix it"`)
		s.run()
		s.check("a later head that passes makes it ready", !exists(s.pr()+".draft"))
	}},
	{"publish-failed", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run("REVIEW=decision")
		s.check("a first publish whose head failed review", r.code == 0)
		s.check("opens the pull request as a draft", exists(s.pr()+".draft"))
		s.check("described by the commit message alone", s.cat(s.pr()) == "The project needed a file.\n\nGrove: 261015-c0000")
	}},
	{"publish-no-drafts", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run("REVIEW=decision", "DRAFT_FAIL=1")
		s.check("where drafts are refused, the run still publishes", r.code == 0)
		s.check("and says the review and status say how it stands", r.said("its review and status say how it stands"))
	}},
	{"request", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		first := s.head()
		s.request()
		r := s.run()
		s.check("a run after a teammate requests changes passes", r.passed())
		s.check("the implementer is asked once", s.prompts("teammate") == 1)
		s.check("with their review", s.has(s.log(), `^agent prompt impl-.* Their review: Say it plainly\.`))
		s.check("and their comment on a line", strings.Contains(s.log(), "Their comments on lines: a.txt:1: Plainer."))
		s.check("the new head is reviewed", s.reviews() == 2 && s.head() != first)
		body := ""
		for _, x := range s.reviewList() {
			if str(get(x, "commit_id")) == s.head() && str(get(x, "user", "login")) == "the-bot" {
				body += str(get(x, "body")) + "\n"
			}
		}
		s.check("and published: its review carries the implementer's reply", s.count(body, `^> Made it plainer, as asked\.`) == 1)
		s.check("its own status", s.has(s.cat(s.st+"/statuses"), `^`+s.head()+` success the-bot$`))
		s.check("still one pull request", s.creates() == 1)
		s.run()
		s.check("the answered request is not asked again", s.prompts("teammate") == 1)
		s.check("the description is written again for the new head", s.cat(s.pr()+".title") == "Add the file, plainer")
		s.check("its body too", strings.Contains(s.cat(s.pr()), "needed a plain file"))
	}},
	{"request-edited", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.write(s.pr(), "Their own words.\n")
		s.request()
		r := s.run()
		s.check("a description a person edited is not written again", s.cat(s.pr()) == "Their own words." && s.count(s.log(), `^gh pr edit`) == 0)
		s.check("and tend says it left it", r.said("a person has edited it"))
	}},
	{"request-declined", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		first := s.head()
		s.request()
		r := s.run("REQUEST_DECLINE=1")
		s.check("a request answered with no change exits 0", r.code == 0)
		s.check("the implementer's reply goes on the pull request, as the account", str(get(s.readJSON(s.st+"/comments"), 0, "user", "login")) == "the-bot")
		s.check("in its words", strings.Contains(s.cat(s.st+"/comments"), "Out of scope: the issue asks"))
		s.check("the unchanged head is not reviewed again", s.reviews() == 1 && s.head() == first)
		s.run("REQUEST_DECLINE=1")
		s.check("a rerun neither asks again nor comments again", s.prompts("teammate") == 1 && len(list(s.readJSON(s.st+"/comments"))) == 1)
		w := s.watch()
		s.check("watch leaves the answered request", s.count(w.err, `running tend`) == 0)
		s.request()
		w = s.watch("REQUEST_DECLINE=1")
		s.check("but runs tend for a new request on the same head", w.said("changes requested on"))
		s.check("which is asked", s.prompts("teammate") == 2)
	}},
	{"request-uncertain", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.request()
		r := s.run("REQUEST_FAIL=agent_prompt_stalled")
		s.check("a stalled request prompt stops", r.code == 3)
		r = s.run()
		s.check("the rerun stops, as nothing was committed", r.code == 3)
		s.check("and says so", r.said("committed nothing and is not working on them"))
		s.check("without asking again", s.prompts("teammate") == 1)
		m, _ := filepath.Glob(s.ts + "/runs/*/request-*")
		for _, f := range m {
			os.Remove(f)
		}
		r = s.run()
		s.check("with the mark deleted by a person, it is asked again and passes", r.passed())
		s.check("twice in all", s.prompts("teammate") == 2)
	}},
	{"request-merged", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.request()
		s.touch(s.pr() + ".merged")
		s.run()
		s.check("a request on a merged pull request is not acted on", s.prompts("teammate") == 0)
	}},
	{"request-stranger", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.requestAs("stranger", "NONE", "CHANGES_REQUESTED")
		w := s.watch()
		s.check("watch runs nothing for a stranger's change request", s.count(w.err, `running tend`) == 0 && s.prompts("stranger") == 0)
		r := s.run()
		s.check("a stranger's change request exits 0", r.code == 0)
		s.check("is not asked", s.prompts("stranger") == 0)
		m, _ := filepath.Glob(s.ts + "/runs/*/request-*")
		s.check("and no request mark is written", len(m) == 0)
		s.check("tend says whose request it left", r.said("left the change request of stranger on https://example.test/pr/"))
	}},
	{"request-stranger-after", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.request()
		s.requestAs("stranger", "CONTRIBUTOR", "CHANGES_REQUESTED")
		s.requestAs("passer-by", "NONE", "COMMENTED")
		s.run()
		s.check("a teammate's request before a stranger's later review is asked once", s.prompts("teammate") == 1)
		s.check("the stranger's is not", s.prompts("stranger") == 0)
	}},
	{"watch", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		w := s.watch()
		s.check("watch runs nothing when no one requested changes", s.count(w.err, `running tend`) == 0)
		s.request()
		w = s.watch()
		s.check("watch runs tend when a teammate requests changes", w.said("changes requested on https://example.test/pr/"))
		s.check("and the request is asked", s.prompts("teammate") == 1)
		s.request()
		s.write(s.runDir()+"/status", "x needs-you: something\nnext: a person\n")
		s.watch()
		s.check("watch leaves an attempt that needs a person", s.prompts("teammate") == 1)
	}},
	{"propose", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend", "2:", "3:tend,bug")
		r := s.propose()
		s.check("tend propose exits 0", r.code == 0)
		s.check("each labelled issue is drafted once", s.drafts() == 2)
		s.check("unattended, from the issue's words", s.has(s.log(), `^`+regexp.QuoteMeta("agent prompt prop-1 /grove-propose --unattended Request 1. Make thing 1 better. (GitHub issue https://example.test/issues/1)")))
		s.check("an unlabelled issue is left alone", !exists(s.prOf("2")))
		s.check("a pull request per draft, by the account", s.cat(s.prOf("1")+".by")+" "+s.cat(s.prOf("3")+".by") == "the-bot the-bot")
		s.check("from the issue's branch, onto the target", s.has(s.log(), `^gh pr create --base main --head tend/issue-1 `))
		s.check("titled and described by the drafter's commit", strings.Contains(s.log(), "--title Propose a thing --body-file"))
		s.check("described by its body, which refers to the issue", s.has(s.cat(s.prOf("1")), `^Refs #1$`))
		s.check("the branch holds the change", s.count(s.git(s.repo+".git", "ls-tree", "-r", "--name-only", "tend/issue-1"), `^grove/changes/261115-p0000`) == 1)
		rv := s.cat(s.st + "/reviews")
		s.check("a review lists the open questions", strings.Contains(rv, "Which thing?"))
		s.check("and says what merging with them open means", strings.Contains(rv, "its implementation may then choose those answers"))
		s.check("naming the change to revise", strings.Contains(rv, "/grove-propose 261115-p0000"))
		s.check("and that once merged it can be built by its ID, or planned with others", strings.Contains(rv, "can be built by its ID, `tend <repo> <change-id>`, or planned into a cycle with others") && !strings.Contains(rv, "waits for a cycle"))
		s.propose()
		s.check("a second pass drafts nothing again", s.drafts() == 2)
		s.check("and opens nothing again", s.creates() == 2)
		s.check("nor reviews again", len(s.reviewList()) == 2)
	}},
	{"propose-declined", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("DECLINE=1")
		s.check("a request that cannot be drafted exits 0", r.code == 0)
		s.check("opens no pull request", !exists(s.prOf("1")))
		s.check("says why on the issue, as the account", str(get(s.readJSON(s.st+"/comments"), 0, "user", "login")) == "the-bot")
		s.check("in the drafter's words", strings.Contains(s.cat(s.st+"/comments"), "It names no outcome."))
		s.check("and takes the label off", !s.labelled(1))
		s.check("its worktree is removed, with its branch", !exists(s.st+"/wt-tend/issue-1") && s.git(s.repo, "branch", "--list", "tend/issue-1") == "")
		s.propose("DECLINE=1")
		s.check("so a second pass leaves it", s.drafts() == 1 && s.comments(1) == 1)
		s.issues("1:tend")
		s.propose()
		s.check("labelled again, it is drafted afresh", s.drafts() == 2 && exists(s.prOf("1")))
	}},
	{"propose-decline-lost", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("DECLINE=1", "EDIT_FAIL=1")
		s.check("a decline whose label removal fails stops", r.code == 3)
		r = s.propose("DECLINE=1")
		s.check("the rerun finishes it", r.code == 0 && !s.labelled(1))
		s.check("without commenting twice or drafting again", s.comments(1) == 1 && s.drafts() == 1)
	}},
	{"propose-cap", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend", "2:tend", "3:tend")
		r := s.propose("CAP=2")
		s.check("the cap holds open proposals", s.drafts() == 2 && !exists(s.prOf("3")))
		s.check("and says so", r.said("the cap; the rest wait"))
		s.touch(s.prOf("1") + ".merged")
		s.propose("CAP=2")
		s.check("a merged proposal frees a place", exists(s.prOf("3")))
	}},
	{"propose-follow-lost", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("FOLLOW_FAIL=agent_prompt_stalled")
		s.check("a stalled follow-up stops", r.code == 3)
		r = s.propose()
		s.check("the rerun stops too, not asking again", r.code == 3 && s.follows() == 1)
		s.check("and says what to do", r.said("delete .*/followed"))
		m, _ := filepath.Glob(s.ts + "/intake/*/followed")
		for _, f := range m {
			os.Remove(f)
		}
		r = s.propose()
		s.check("with the mark deleted, it is asked again and published", r.code == 0 && s.follows() == 2 && exists(s.prOf("1")))
		s.check("the draft was asked for once", s.drafts() == 1)
	}},
	{"propose-survivor", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("PROPOSE_FAIL=timeout")
		s.check("a drafting prompt that timed out stops", r.code == 3)
		s.agent("working", "prop")
		r = s.propose()
		s.check("the rerun waits for the drafter, then has it commit", r.code == 0 && exists(s.prOf("1")))
		s.check("without asking for the draft again", s.drafts() == 1)
	}},
	{"propose-unlisted", func(s *sc) {
		s.project("main", true)
		s.issues("1:tend")
		r := s.propose()
		s.check("a repository not listed for publication gives no requests", r.code == 3 && s.drafts() == 0)
		s.check("and tend says why", r.said("tend takes no requests from it"))
	}},
	{"overview", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.change("261016-c0001", "Another", "")
		s.issues("1:tend", "2:tend")
		s.propose("CAP=1")
		s.run()
		s.request()
		o := s.overview()
		s.check("the overview names an approved change in no cycle, with Grove's verdict", s.has(after(o, "Approved changes in no cycle", 1), `^  261016-c0001 Another: grove claim 261016-c0001 would claim it`))
		s.check("a cycle not done, with Grove's verdict", s.has(o, `^  261114-k0000 One: 261114-k0000 is in flight in attempt`))
		s.check("the attempt on this runner and its stage", s.has(o, `^  [0-9]{6}-[a-z0-9]{5} \(261114-k0000\): done: success`))
		s.check("the proposal to review", strings.Contains(after(o, "Proposals to review", 1), "Propose a thing"))
		s.check("the implementation to review, with the change requested", s.has(after(o, "Implementations to review", 1), `Add the file .*changes requested`))
		s.check("and the request still waiting for intake", strings.Contains(after(o, "Requests waiting", 1), "#2 Request 2"))
		s.check("but not the one already drafted", !strings.Contains(o, "#1 Request 1"))
		s.touch(s.prOf("1") + ".merged")
		s.check("nor once its proposal is merged", !strings.Contains(s.overview(), "#1 Request 1"))
		s.bash(s.repo, `mkdir -p grove/changes/archive && git mv grove/changes/261015-c0000-a-change grove/changes/archive/261015-c0000-a-change && git commit -qm landed && git push -q origin main`)
		s.check("an attempt whose cycle is done is left out", !strings.Contains(s.overview(), "(261114-k0000)"))
	}},
	{"change-claim", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.change("261016-c0001", "Another", "")
		r := s.crun("261016-c0001")
		s.check("a change in no cycle is claimed by name and passes review", r.passed())
		s.check("Grove holds one record, naming it alone and no cycle", s.changeRecords("261016-c0001") == 1 && s.records() == 0)
		s.check("the implementer is asked for the change under --runner", s.has(s.log(), `^agent prompt impl-[^ ]* /grove-implement --runner 261016-c0001 --wait`))
		s.check("one pull request is opened, onto the target", s.count(s.log(), `^gh pr create --base main `) == 1)
		s.check("its description ends with the change", strings.HasSuffix(s.cat(s.pr()), "\nGrove: 261016-c0001"))
		st := s.tendIn(nil, "status", s.repo, "261016-c0001").out
		s.check("status finds the attempt", s.has(st, `^attempt [0-9]{6}-[a-z0-9]{5} of 261016-c0001: grove/[0-9]{6}-[a-z0-9]{5} in `))
		s.check("and what it recorded", s.has(st, `done: success`))
	}},
	{"change-resume", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		r := s.crun("261016-c0001", "IMPL_FAIL=stub_failure")
		s.check("an implementer failure after the claim stops for a person", r.code == 3)
		id := str(get(parse(s.sh(s.repo, "grove", "changes", "--json")), "runs", 0, "id"))
		r = s.crun("261016-c0001")
		s.check("the rerun exits 0", r.code == 0)
		s.check("the rerun carries on the same attempt", r.said("carrying on attempt "+id))
		s.check("one record for the change", s.changeRecords("261016-c0001") == 1)
	}},
	{"change-waits", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "261016-c0001")
		r := s.crun("261017-c0002")
		s.check("a change that waits stops for a person", r.code == 3)
		s.check("with Grove's sentence", r.said("261017-c0002 waits on 261016-c0001"))
		s.check("and no implementer is prompted", s.count(s.log(), `^agent prompt`) == 0)
		o := s.overview()
		s.check("the overview gives its verdict", s.has(after(o, "Approved changes in no cycle", 2), `^  261017-c0002 Third: 261017-c0002 waits on 261016-c0001`))
	}},
	{"change-in-cycle", func(s *sc) {
		s.project("main", true)
		r := s.crun("261015-c0000")
		s.check("a change a cycle lists stops for a person", r.code == 3)
		s.check("with Grove's sentence", r.said("261015-c0000 is in cycle 261114-k0000"))
		s.check("and no implementer is prompted", s.count(s.log(), `^agent prompt`) == 0)
		s.check("the worktree made for the refused claim is removed", s.worktree() == s.st+"/wt-tend-[0-9]*" && s.git(s.repo, "branch", "--list", "tend-*") == "")
		s.check("the cycle is still claimed as before", s.run().passed() && s.records() == 1)
	}},
	{"change-overview", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		s.crun("261016-c0001")
		s.check("the overview lists the attempt at the change", s.has(s.overview(), `^  [0-9]{6}-[a-z0-9]{5} \(261016-c0001\): done: success`))
		s.bash(s.repo, `grove archive 261016-c0001 >/dev/null && git add -A && git commit -qm landed && git push -q origin main`)
		s.check("and no longer once the target holds it archived", !strings.Contains(s.overview(), "(261016-c0001)"))
	}},
	{"change-tidy", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.change("261016-c0001", "Another", "")
		s.crun("261016-c0001", "IMPL_FAIL=stub_failure")
		wt, b := s.worktree(), s.branch()
		s.issues()
		s.propose()
		s.check("an attempt at a change not archived stays", isDir(wt))
		s.bash(s.repo, `grove archive 261016-c0001 >/dev/null && git add -A && git commit -qm "built elsewhere" && git push -q origin main`)
		s.propose()
		s.check("an attempt with no pull request whose change is archived elsewhere is removed", !exists(wt))
		s.check("with its local branch", s.git(s.repo, "branch", "--list", b) == "")
	}},
	{"watch-intake", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		s.watchrepo()
		s.check("watch runs intake for a repository it is given", exists(s.prOf("1")))
		s.issues("1:tend", "2:tend")
		s.watchrepo("FOLLOW_FAIL=agent_prompt_stalled")
		s.watchrepo()
		s.check("and leaves a request that needs a person", s.follows() == 2)
	}},
	{"issue-build", func(s *sc) {
		s.plain()
		s.allow()
		s.issues("7:tend-build")
		r := s.build()
		s.check("an issue labelled tend-build is built, reviewed and published without Grove", r.passed())
		s.check("the implementer is given the issue as its scope", s.has(s.log(), `^agent prompt impl-issue-7-[0-9]* Implement GitHub issue #7 .* Request 7. Make thing 7 better.`))
		s.check("the reviewer is given the issue and the implementer's notes", s.has(s.log(), `^agent prompt rev-.*what it must do is GitHub issue #7, quoted here: Request 7.*notes are in `+regexp.QuoteMeta(s.ts)+`/runs/issue-7-`))
		s.check("on branch tend/build-7, pushed", s.git(s.repo+".git", "rev-parse", "tend/build-7") == s.git(s.bworktree(), "rev-parse", "HEAD"))
		s.check("the commit message is asked to close the issue", s.has(s.log(), `^agent prompt impl-issue-7-.*End with the line 'Closes #7'`))
		s.check("one pull request by the account", s.cat(s.st+"/pr-tend_build-7.by") == "the-bot")
		s.check("its review shows no grove review", !strings.Contains(s.cat(s.st+"/reviews"), "What merging approves"))
		s.check("its attempt is named by cksum of the repository's path, as before", isDir(s.ts+"/runs/issue-7-"+s.cksum(s.repo)))
		s.check("status finds the attempt", s.has(s.tendIn(nil, "status", s.repo, "issue-7").out, `^attempt issue-7-.* of issue-7: tend/build-7`))
		s.check("the overview lists it to review", strings.Contains(after(s.tendIn(nil, "status", s.repo).out, "Implementations to review", 1), "Add the file"))
		s.build()
		s.check("a rerun neither builds nor reviews again", s.count(s.log(), `^agent prompt impl-issue-7-[0-9]* Implement`) == 1 && s.reviews() == 1)
		s.touch(s.st + "/pr-tend_build-7.merged")
		r = s.build()
		s.check("once its pull request is merged, a rerun does nothing", r.code == 0 && s.reviews() == 1)
		s.check("and says so", r.said("is merged; nothing to do"))
		s.check("its worktree is removed", s.bworktree() == "")
	}},
	{"issue-ineligible", func(s *sc) {
		s.plain()
		s.allow()
		s.issues("7:tend")
		r := s.build()
		s.check("an issue without tend-build stops, not approved", r.code == 3 && s.count(s.log(), `^agent prompt`) == 0)
		s.check("and says so", r.said("not labelled tend-build"))
		is := list(s.readJSON(s.st + "/issues"))
		for _, i := range is {
			i.(map[string]any)["state"] = "CLOSED"
		}
		s.writeJSON(s.st+"/issues", is)
		r = s.build()
		s.check("a closed issue is done: exit 0, nothing run", r.code == 0 && s.count(s.log(), `^agent prompt`) == 0)
	}},
	{"issue-in-grove", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("7:tend-build")
		r := s.build()
		s.check("an issue in a repository with Grove is not built directly", r.code == 3 && s.count(s.log(), `^agent prompt`) == 0)
		s.check("and tend points to a proposal", r.said("label it tend for tend propose"))
	}},
	{"issue-no-notes", func(s *sc) {
		s.plain()
		s.allow()
		s.issues("7:tend-build")
		r := s.build("NO_NOTES=1")
		s.check("a build without notes for the reviewer stops", r.code == 3)
		s.check("and says so", r.said("wrote no notes for the reviewer"))
	}},
	{"issue-survivor", func(s *sc) {
		s.plain()
		s.allow()
		s.issues("7:tend-build")
		r := s.build("IMPL_FAIL=timeout")
		s.check("an issue build whose prompt timed out stops", r.code == 3)
		s.agent("working", "impl")
		r = s.build()
		s.check("the rerun waits for the implementer on the same branch and passes", r.code == 0 && r.conclusion() == "success")
		s.check("without prompting it again", s.count(s.log(), `^agent prompt impl-issue-7-[0-9]* Implement`) == 1)
	}},
	{"issue-repair", func(s *sc) {
		s.plain()
		s.allow()
		s.issues("7:tend-build")
		r := s.build("REVIEW=repairable")
		s.check("an issue build is repaired once and passes", r.passed())
		s.check("the repair is told to update the notes", s.has(s.log(), `^agent prompt impl-issue-7-.*update `+regexp.QuoteMeta(s.ts)+`/runs/issue-7-[0-9]*/implemented where a claim changes`))
	}},
	{"publish-diverged", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.bash(s.dir, fmt.Sprintf(`git clone -q "$t/repo.git" "$t/other" && cd "$t/other" && git checkout -q %q && echo theirs >>a.txt && git commit -qam theirs && git push -q origin HEAD`, s.branch()))
		s.bash(s.worktree(), `echo ours >>a.txt && git commit -qam ours`)
		r := s.run()
		s.check("a branch someone else pushed to is not overwritten", r.code == 3)
		s.check("and tend says why", r.said("could not push"))
		s.check("their commit is still on origin", s.git(s.repo+".git", "log", "-1", "--format=%s", s.branch()) == "theirs")
	}},
	{"cached-wrong-head", func(s *sc) {
		s.project("main", false)
		s.run()
		v := s.verdict()
		x := s.readJSON(v).(map[string]any)
		x["head"] = "0000000000000000000000000000000000000000"
		s.writeJSON(v, x)
		r := s.run()
		s.check("rerun exits 0", r.code == 0)
		s.check("a cached verdict naming another head is reviewed again", s.reviews() == 2)
		s.check("the new verdict names the head", str(get(s.readJSON(v), "head")) == s.head())
	}},
	{"cached-incomplete", func(s *sc) {
		s.project("main", false)
		s.run()
		v := s.verdict()
		x := s.readJSON(v).(map[string]any)
		delete(x, "conclusion")
		s.writeJSON(v, x)
		s.run()
		s.check("a cached verdict without a conclusion is reviewed again", s.reviews() == 2)
		m, _ := filepath.Glob(v + ".rejected-*")
		s.check("the set-aside verdict is kept", len(m) > 0)
	}},
	{"new-incomplete", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=noconclusion")
		s.check("a new verdict without a conclusion stops for a person", r.code == 3)
		s.check("and says so", r.said("without a valid verdict"))
	}},
	{"target-branch", func(s *sc) {
		s.project("trunk", false)
		r := s.run()
		s.check("a project targeting trunk reaches a verdict", r.code == 0)
		s.check("the worktree is based on trunk", strings.Contains(s.log(), "--base trunk"))
		s.check("the reviewer is shown the diff from trunk", strings.Contains(s.log(), "git diff trunk...HEAD"))
	}},
	{"relative-state", func(s *sc) {
		s.project("main", false)
		r := s.tendIn([]string{"TEND_STATE=rel"}, s.repo, "261114-k0000")
		s.check("a relative TEND_STATE reaches a verdict", r.code == 0)
		s.check("its files stay outside the candidate", s.git(s.worktree(), "status", "--porcelain") == "")
	}},
	{"root-moved", func(s *sc) {
		s.project("main", false)
		s.run()
		s.bash(s.repo, `git mv grove agreed && sed 's/^root: .*/root: agreed/' grove.yaml >g && mv g grove.yaml && git commit -qam "Move the documents"`)
		r := s.run()
		s.check("after the target moves its documents, the attempt carries on", r.code == 0)
		s.check("its finished candidate is not implemented again", s.count(s.log(), `^agent prompt impl-`) == 1)
	}},
	{"legacy", func(s *sc) {
		s.project("main", false)
		s.run()
		// Put the attempt back as the prototype left it: its files in .tend/ in the checkout.
		runDir := filepath.Dir(s.verdict())
		if strings.HasPrefix(runDir, s.ts+"/runs/") {
			s.touch(runDir + "/repaired")
			s.bash(runDir, `mkdir "$0/.tend" && mv ./* "$0/.tend/"`+"\n", s.worktree())
		}
		r := s.run()
		s.check("an attempt with a legacy .tend/ carries on", r.code == 0)
		s.check("its verdict is reused, not discarded", s.reviews() == 1)
		s.check("its consumed repair stays consumed, as asked", exists(runDir+"/repair-asked"))
		s.check("and .tend/ leaves the candidate", !exists(s.worktree()+"/.tend"))
	}},
	{"lock", func(s *sc) {
		s.project("main", false)
		bg := s.background("IMPL_SLEEP=30")
		r := s.run()
		s.check("a second tend while one runs is refused", r.code == 4)
		s.check("and names the holder", r.said(fmt.Sprintf("pid %d", bg.Process.Pid)))
		s.check("the implementer was prompted once", s.count(s.log(), `^agent prompt impl-`) == 1)
		killed(bg)
		r = s.run()
		s.check("after the holder is killed, tend takes over", r.code == 0)
		s.check("carrying on its attempt", r.said("carrying on attempt"))
		s.check("waiting for the implementer, prompted once in all", s.count(s.log(), `^agent prompt impl-`) == 1)
	}},
	{"repair", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=repairable")
		s.check("a failed actionable review is repaired once and passes", r.passed())
		s.check("the repair was asked once", s.prompts("An") == 1)
		s.check("status says done", strings.Contains(s.status(), "done: success"))
	}},
	{"repair-undelivered", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=repairable", "REPAIR_FAIL=agent_prompt_stalled")
		s.check("a stalled repair prompt stops for a person", r.code == 3)
		r = s.run("REVIEW=repairable")
		s.check("the rerun neither skips the repair nor exits 0 with the failed verdict", r.code == 3)
		s.check("nor asks the repair again", s.prompts("An") == 1)
		s.check("it says the outcome is unknown and what to do", r.said("has no commit and impl-.* is not working on it"))
		s.check("status shows that", strings.Contains(s.status(), "needs-you: the repair asked"))
		os.Remove(s.runDir() + "/repair-asked")
		r = s.run("REVIEW=repairable")
		s.check("with repair-asked deleted by a person, the repair is asked again and passes", r.passed())
		s.check("twice in all", s.prompts("An") == 2)
	}},
	{"repair-blocked", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=repairable", "BLOCK_AFTER_BUILD=1")
		s.check("an implementer blocked before the repair stops for a person", r.code == 3)
		s.check("and the repair is not used", !exists(s.runDir()+"/repair-asked"))
		s.agent("idle", "impl")
		r = s.run("REVIEW=repairable")
		s.check("once unblocked, the repair is asked and passes", r.passed())
	}},
	{"unfinished", func(s *sc) {
		s.project("main", false)
		r := s.run("UNFINISHED=1")
		s.check("an implementer turn that leaves the cycle unbuilt stops for a person", r.code == 3)
		s.check("naming what is missing, with no record to delete", r.said(`261015-c0000`) && !r.said(`delete`))
		r = s.run()
		s.check("the rerun asks the implementation again and passes", r.passed())
		s.check("prompting it twice in all", s.prompts("/grove-implement") == 2)
	}},
	{"survivor-implementing", func(s *sc) {
		s.project("main", false)
		r := s.run("IMPL_FAIL=timeout")
		s.check("an implementer prompt that timed out stops", r.code == 3)
		s.check("status shows the stop", s.has(s.status(), `needs-you: impl-.*: timeout`))
		s.agent("working", "impl")
		r = s.run()
		s.check("the rerun waits for the implementer still working", r.code == 0)
		s.check("and does not prompt it again", s.prompts("/grove-implement") == 1)
		s.check("and says so", r.said("still working on an earlier prompt"))
	}},
	{"survivor-reviewing", func(s *sc) {
		s.project("main", false)
		r := s.run("REV_FAIL=timeout")
		s.check("a reviewer prompt that timed out stops", r.code == 3)
		s.agent("working", "rev")
		r = s.run()
		s.check("the rerun waits for the reviewer still working and passes", r.passed())
		s.check("it is prompted once", s.reviews() == 1)
		s.check("and started once", s.count(s.log(), `^agent start rev-`) == 1)
	}},
	{"survivor-repairing", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=repairable", "REPAIR_FAIL=timeout")
		s.check("a repair prompt that timed out stops", r.code == 3)
		s.agent("working", "impl")
		r = s.run("REVIEW=repairable")
		s.check("the rerun waits for the repair still working, reviews it and passes", r.passed())
		s.check("the repair was asked once", s.prompts("An") == 1)
	}},
	{"status-interrupted", func(s *sc) {
		s.project("main", false)
		bg := s.background("IMPL_SLEEP=30")
		st := s.status()
		s.check("status names the running tend", strings.Contains(st, fmt.Sprintf("tend is on it: pid %d", bg.Process.Pid)))
		s.check("and what it is waiting for", s.has(st, `working: impl-.* implements`))
		killed(bg)
		if pid, err := strconv.Atoi(s.cat(s.st + "/sleeper")); err == nil {
			syscall.Kill(pid, syscall.SIGTERM)
		}
		time.Sleep(500 * time.Millisecond)
		s.check("after a kill, status says to run tend again", strings.Contains(s.status(), "tend stopped before this finished"))
	}},
	{"jira-propose", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.jmap()
		s.issues()
		s.jissue("SCRUM-1", "tend", "new", rich, 3)
		s.jissue("SCRUM-2", "tend", "done", "", 0)
		s.jissue("SCRUM-3", "", "new", "", 0)
		s.jissue("SCRUM-4", "tend,bug", "indeterminate", "", 0)
		s.jissue("SCRUM-5", "tend", "new", "", 0)
		r := s.propose()
		s.check("tend propose takes Jira requests and exits 0", r.code == 0)
		s.check("each labelled, unfinished one is drafted once, past the first page", s.count(s.log(), `^agent prompt prop-scrum-[0-9]* /grove-propose`) == 3)
		s.check("the drafter gets the description with its links, list and code", s.has(s.log(), `^`+regexp.QuoteMeta("agent prompt prop-scrum-1 /grove-propose --unattended Request SCRUM-1. Implement the spec (https://docs.test/spec) and https://jira.test/browse/SCRUM-9 - keep totals ledger list --json (Jira issue https://jira.test/browse/SCRUM-1)")))
		s.check("a done or unlabelled request is left alone", !exists(s.st+"/pr-tend_jira-SCRUM-2") && !exists(s.st+"/pr-tend_jira-SCRUM-3"))
		s.check("a pull request from its branch, by the account", s.cat(s.st+"/pr-tend_jira-SCRUM-1.by") == "the-bot")
		s.check("the drafter is asked to end the commit with the key", s.has(s.log(), `^agent prompt prop-scrum-1 .*ends with the line 'Refs SCRUM-1'`))
		s.check("the review says where it came from", strings.Contains(s.cat(s.st+"/reviews"), "Proposal drafted from SCRUM-1 "))
		s.check("the Jira issue is told where its proposal is", strings.Contains(s.jcomments("SCRUM-1"), "tend opened https://example.test/pr/tend_jira-SCRUM-1 "))
		s.check("the handoff is recorded finished", nonempty(s.ts+"/intake/o_r-jira-SCRUM-1/linked"))
		s.check("the token is never in curl's arguments", !strings.Contains(s.log(), "tok-secret-1"))
		s.propose()
		s.check("a second pass drafts and opens nothing again", s.count(s.log(), `^agent prompt prop-scrum-[0-9]* /grove-propose`) == 3 && s.creates() == 3)
		s.check("nor comments again, though its comment is on a later page", s.count(s.jcomments("SCRUM-1"), `tend opened`) == 1)
		s.jissue("SCRUM-6", "tend", "new", "", 0)
		o := s.tendIn(nil, "status", s.repo).out
		waiting := o[max(0, strings.Index(o, "Requests waiting")):]
		s.check("the overview lists a Jira request waiting, not the handled ones", s.count(waiting, `SCRUM-`) == 1 && strings.Contains(o, "  SCRUM-6 Request SCRUM-6 https://jira.test/browse/SCRUM-6"))
		s.check("and its proposals to review", strings.Contains(after(o, "Proposals to review", 3), "example.test/pr/tend_jira"))
	}},
	{"jira-handoff", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.jmap()
		s.issues()
		s.jissue("SCRUM-1", "tend", "new", "", 0)
		r := s.propose("COMMENT_FAIL=1")
		s.check("a Jira comment that fails after the pull request stops", r.code == 3)
		s.check("and says a rerun finishes", r.said("run tend propose again to finish"))
		js := list(s.readJSON(s.st + "/jira"))
		for _, j := range js {
			j.(map[string]any)["fields"].(map[string]any)["labels"] = []any{}
		}
		s.writeJSON(s.st+"/jira", js)
		s.propose()
		s.check("the next pass finishes it, though the label is gone", s.count(s.jcomments("SCRUM-1"), `tend opened`) == 1)
		s.check("without a second pull request", s.creates() == 1)
		s.check("and records it finished", nonempty(s.ts+"/intake/o_r-jira-SCRUM-1/linked"))
	}},
	{"jira-declined", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.jmap()
		s.issues()
		s.jissue("SCRUM-1", "tend", "new", "", 0)
		s.propose("DECLINE=1")
		c := s.jcomments("SCRUM-1")
		s.check("a Jira request no proposal can be drafted from is answered", strings.Contains(c, "No proposal could be drafted from this request"))
		s.check("with the drafter's reason", s.has(c, `^It names no outcome`))
		s.check("and a marker to find it by", s.has(c, `^\[tend:declined [0-9]*\]$`))
		s.check("and loses its tend label", len(list(get(s.readJSON(s.st+"/jira"), 0, "fields", "labels"))) == 0)
		m, _ := filepath.Glob(s.ts + "/intake/o_r-jira-SCRUM-1.declined-*")
		s.check("its run folder is set aside", len(m) > 0)
	}},
	{"jira-build", func(s *sc) {
		s.plain()
		s.allow()
		s.jmap()
		s.jissue("SCRUM-7", "tend-build", "new", rich, 0)
		r := s.jbuild()
		s.check("a Jira issue labelled tend-build is built, reviewed and published", r.passed())
		s.check("the implementer is given the Jira issue as its scope", s.has(s.log(), `^agent prompt impl-jira-scrum-7-[0-9]* Implement Jira issue SCRUM-7 .* Implement the spec \(https://docs.test/spec\)`))
		s.check("on branch tend/build-SCRUM-7, pushed", s.gitOK(s.repo+".git", "rev-parse", "--verify", "-q", "tend/build-SCRUM-7"))
		s.check("the commit message is asked to refer to the key", s.has(s.log(), `^agent prompt impl-jira-scrum-7-.*End with the line 'Refs SCRUM-7'`))
		s.check("the Jira issue is told where its pull request is", strings.Contains(s.jcomments("SCRUM-7"), "tend opened https://example.test/pr/tend_build-SCRUM-7 "))
		s.check("its attempt is named for the key and repository", isDir(s.ts+"/runs/jira-scrum-7-"+s.cksum(s.repo)))
		s.check("status finds it", s.has(s.tendIn(nil, "status", s.repo, "SCRUM-7").out, `^attempt jira-scrum-7-.* of SCRUM-7: tend/build-SCRUM-7`))
		s.jbuild()
		s.check("a rerun builds, reviews and comments nothing again", s.count(s.log(), `^agent prompt impl-jira`) == 2 && s.reviews() == 1 && s.count(s.jcomments("SCRUM-7"), `tend opened`) == 1)
	}},
	{"jira-ineligible", func(s *sc) {
		s.plain()
		s.allow()
		s.jmap()
		s.jissue("SCRUM-7", "", "new", "", 0)
		s.jissue("SCRUM-8", "tend-build", "done", "", 0)
		r := s.jbuild()
		s.check("an unlabelled Jira issue is not built", r.code == 3 && r.said("issue SCRUM-7 is not labelled tend-build"))
		r = s.tendIn(nil, s.repo, "SCRUM-8")
		s.check("a done one has nothing to do", r.code == 0 && r.said("issue SCRUM-8 is closed"))
		r = s.tendIn(nil, s.repo, "OTHER-1")
		s.check("an unmapped project is refused", r.code == 3 && r.said("Jira project OTHER is not mapped"))
		s.write(s.ts+"/jira", regexp.MustCompile(`(?m) [^ ]*$`).ReplaceAllString(s.cat(s.ts+"/jira"), " git@github.com:someone/else.git")+"\n")
		r = s.jbuild()
		s.check("a project mapped to another repository is refused", r.code == 3 && r.said("not to this repository's origin"))
	}},
	{"rex", func(s *sc) {
		s.project("main", false)
		r := s.rexrun()
		s.check("under rex, a cycle is claimed, built and reviewed", r.passed())
		s.check("its worktree is made by Git under the state folder", sameDir(filepath.Dir(s.rexwt()), s.ts+"/worktrees"))
		sessions := strings.Fields(s.cat(s.st + "/rex-sessions"))
		s.check("and opened as a session of its own", len(sessions) == 2 && sessions[1] == filepath.Base(s.rexwt()))
		s.check("each agent runs in a block named for it", strings.Contains(s.cat(s.st+"/rex-blocks"), " impl-"))
		s.check("the reviewer's block is closed after its verdict", s.count(s.cat(s.st+"/rex-blocks"), ` rev-`) == 0)
		s.check("prompts go as a paste", s.has(s.rexlog(), "^rex send .*\x1b\\[200~/grove-implement"))
		s.check("nothing was asked of Herdr", s.count(s.log(), `^worktree|^pane|^agent get`) == 0)
		s.write(s.st+"/rex-sessions", regexp.MustCompile(`(?m) .*$`).ReplaceAllString(s.cat(s.st+"/rex-sessions"), " renamed-by-an-agent")+"\n")
		s.rexrun()
		s.check("a rerun finds its session, though an agent renamed it, and reviews nothing again", len(strings.Split(s.cat(s.st+"/rex-sessions"), "\n")) == 1 && s.reviews() == 1)
	}},
	{"rex-repair", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("REVIEW=repairable")
		s.check("under rex, a failed review is repaired once and passes", r.passed())
		s.check("the repair was asked once", s.prompts("An") == 1)
	}},
	{"rex-stalled", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("REX_STALL=1", "TIMEOUT_MS=1500")
		s.check("under rex, a prompt the agent never takes up stops, after the prompt's timeout", r.code == 3)
		s.check("as stalled", r.said("agent_prompt_stalled"))
		s.agent("working", "impl")
		r = s.rexrun()
		s.check("the rerun waits for the agent once it is working", r.code == 0 && r.conclusion() == "success")
		s.check("and does not prompt it again", s.prompts("/grove-implement") == 1)
		s.check("and says so", r.said("still working on an earlier prompt"))
	}},
	{"rex-dialog", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("REVIEW=repairable", "BLOCK_AFTER_BUILD=1")
		s.check("under rex, an agent reporting blocked stops before the repair", r.code == 3)
		s.check("saying it is blocked", r.said("impl-.* is blocked"))
		s.agent("none", "impl")
		r = s.rexrun("REVIEW=repairable")
		s.check("an agent reporting nothing, as at a dialog, is not prompted", r.code == 3)
		s.check("it is unknown, a person's", r.said("is unknown; look at it"))
		s.check("and the repair is not used", !exists(s.runDir()+"/repair-asked"))
		s.agent("idle", "impl")
		r = s.rexrun("REVIEW=repairable")
		s.check("once it reports idle, the repair is asked and passes", r.passed())
	}},
	{"rex-trust", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1")
		s.check("under rex, the trust dialog of a new folder is answered", r.passed())
		s.check("by choosing to trust it", s.has(s.rexlog(), `^rex send-key .* down enter`))
	}},
	{"rex-add-dir", func(s *sc) {
		s.project("main", false)
		began := time.Now()
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Yes, add it")
		s.check("Claude's dialog for adding the run folder is answered", r.passed())
		s.check("by choosing to add it", s.has(s.rexlog(), `^rex send-key .* down enter`))
		s.check("without waiting out the agents' first reports", time.Since(began) < 60*time.Second)
	}},
	{"rex-add-dir-refused", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Yes, add it", "TRUST_REFUSE=1")
		s.check("a dialog that refuses tend's first answer is answered again", r.passed())
		seen, twice := map[string]bool{}, false
		re := regexp.MustCompile(`^rex send-key .* down enter`)
		for _, l := range strings.Split(s.rexlog(), "\n") {
			if re.MatchString(l) {
				twice = twice || seen[l]
				seen[l] = true
			}
		}
		s.check("twice", twice)
	}},
	{"rex-hooks", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Hooks can run outside the sandbox after you trust them.")
		s.check("an agent asking to review new hooks stops for a person", r.code == 3)
		s.check("saying so", r.said("asks to review new or changed hooks"))
		s.check("and is sent no key", !s.has(s.rexlog(), `^rex send-key`))
	}},
	{"rex-hooks-wrapped", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Hooks can run outside the\nsandbox after you trust them.\n  enter continue · esc\nskip")
		s.check("a hooks review wrapped in a narrow pane still stops for a person", r.code == 3)
		s.check("saying so", r.said("asks to review new or changed hooks"))
		s.check("and is sent no key, not even Esc", !s.has(s.rexlog(), `^rex send-key`))
	}},
	{"rex-notice", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("NOTICE=1", "TRUST_DIALOG=1")
		s.check("under rex, a startup notice offering esc skip is skipped", r.passed())
		first := map[string]string{} // each block's first key
		re := regexp.MustCompile(`^rex send-key -s \S+ -b (\S+) (.*)$`)
		for _, l := range strings.Split(s.rexlog(), "\n") {
			if m := re.FindStringSubmatch(l); m != nil && first[m[1]] == "" {
				first[m[1]] = m[2]
			}
		}
		esc := 0
		for _, k := range first {
			if k == "esc" {
				esc++
			}
		}
		s.check("by Esc alone, the first key each agent is sent", esc > 0 && esc == len(first))
		s.check("saying so for each agent", s.count(r.err, `skipped .*'s startup notice`) == esc)
	}},
	{"rex-unknown-screen", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Choose a theme")
		s.check("under rex, an agent at a screen tend does not know stops for a person", r.code == 3)
		s.check("saying it did not start", r.said("did not start"))
		s.check("and is sent no key", !s.has(s.rexlog(), `^rex send-key`))
	}},
	{"watch-unattended", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues()
		s.watchrepo()
		s.check("watch starts no cycle that is not marked unattended", s.records() == 0)
		s.unattend(s.repo)
		w := s.watchrepo()
		s.check("watch starts the cycle once its plan marks it unattended", w.said("starting unattended cycle 261114-k0000"))
		s.check("which is claimed, built and reviewed", s.records() == 1 && s.reviews() == 1)
		w = s.watchrepo()
		s.check("a claimed cycle is not started again", s.records() == 1 && s.count(w.err, `starting`) == 0)
	}},
	{"watch-unattended-merged", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues()
		s.bash(s.dir, `git clone -q "$t/repo.git" "$t/other"`)
		s.unattend(s.dir + "/other")
		w := s.watchrepo()
		s.check("a cycle marked unattended on GitHub, not in the local checkout, is started", w.said("starting unattended cycle 261114-k0000"))
		s.check("the local target is fast-forwarded to origin's", s.git(s.repo, "rev-parse", "main") == s.git(s.repo+".git", "rev-parse", "main"))
		s.write(s.repo+"/a.txt", "local\n")
		s.bash(s.dir+"/other", `echo b >b.txt && git add b.txt && git commit -qm b && git push -q origin main`)
		w = s.watchrepo()
		s.check("a checkout with uncommitted changes is left as it is", w.said("has uncommitted changes"))
	}},
	{"drafter-terms", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		s.propose()
		start := strings.Join(slices.DeleteFunc(strings.Split(s.log(), "\n"), func(l string) bool { return !strings.HasPrefix(l, "agent start prop-") }), "\n")
		s.check("the drafter is denied pushing and gh pr", strings.Contains(start, "--disallowedTools Bash(git push:*) Bash(gh pr:*)"))
		s.check("and told tend publishes", strings.Contains(start, "Push nothing, open or edit no pull request"))
		s.check("but to commit only when tend asks", strings.Contains(start, "Commit only when tend asks you to."))
	}},
	{"fetch-target", func(s *sc) {
		s.project("main", true)
		theirs := s.pushed()
		r := s.run()
		s.check("an attempt starts from origin's target, fetched, though the local one is behind", s.gitOK(s.worktree(), "merge-base", "--is-ancestor", theirs, "HEAD"))
		s.check("and passes", r.passed())
	}},
	{"fetch-propose", func(s *sc) {
		s.project("main", true)
		s.allow()
		theirs := s.pushed()
		s.issues("1:tend")
		s.propose()
		s.check("a proposal starts from origin's target, fetched", s.gitOK(s.st+"/wt-tend/issue-1", "merge-base", "--is-ancestor", theirs, "HEAD"))
		s.check("and is opened", exists(s.prOf("1")))
	}},
	{"tidy", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		s.run()
		s.propose()
		wt, b, p, pw := s.worktree(), s.branch(), s.pr(), s.st+"/wt-tend/issue-1"
		s.touch(p + ".merged")
		s.propose()
		s.check("a merged attempt's worktree is removed", !exists(wt))
		s.check("with its local branch", s.git(s.repo, "branch", "--list", b) == "")
		s.check("an open proposal's worktree stays", isDir(pw))
		s.touch(s.prOf("1") + ".merged")
		s.write(pw+"/scratch.txt", "scratch\n")
		r := s.propose()
		s.check("a merged proposal's worktree with uncommitted files is left", isDir(pw))
		s.check("and tend says so", r.said("it has uncommitted files"))
		os.Remove(pw + "/scratch.txt")
		s.propose()
		s.check("once clean, it is removed", !exists(pw))
		m, _ := filepath.Glob(s.ts + "/intake/*-1/worktree")
		s.check("and not looked at again", len(m) == 0)
	}},
	{"tidy-superseded", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run("IMPL_FAIL=stub_failure")
		wt, b := s.worktree(), s.branch()
		s.bash(s.repo, `grove archive 261015-c0000 >/dev/null && git add -A && git commit -qm "built elsewhere" && git push -q origin main`)
		s.bash(s.repo, `git worktree add -q "$0/nested" -b nested`+"\n", wt)
		s.issues()
		r := s.propose()
		s.check("an attempt with a worktree inside it is left", isDir(wt))
		s.check("and tend says so", r.said("a worktree is inside it"))
		s.bash(s.repo, `git worktree remove "$0/nested"`+"\n", wt)
		s.agent("working", "impl-")
		r = s.propose()
		s.check("an attempt whose implementer is working is left", isDir(wt))
		s.check("and tend says so", r.said("its implementer is working"))
		s.agent("idle", "impl-")
		s.propose()
		s.check("an attempt with no pull request whose cycle is done elsewhere is removed", !exists(wt))
		s.check("with its local branch", s.git(s.repo, "branch", "--list", b) == "")
	}},
	{"tidy-unfinished", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run("IMPL_FAIL=stub_failure")
		wt := s.worktree()
		s.issues()
		s.propose()
		s.check("an attempt with no pull request whose cycle is not done stays", isDir(wt))
	}},
	{"rex-tidy", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues()
		s.rexrun()
		wt := s.rexwt()
		s.touch(s.st + "/pr-" + strings.ReplaceAll(s.git(wt, "branch", "--show-current"), "/", "_") + ".merged")
		s.propose("TEND_TERMINAL=rex")
		s.check("under rex, a merged attempt's worktree is removed", !exists(wt))
		s.check("its session killed", s.has(s.rexlog(), `^rex kill session:1`))
		s.check("and forgotten", !nonempty(s.ts+"/rex-sessions"))
	}},
	{"usage", func(s *sc) {
		r := s.tendIn(nil, s.dir)
		s.check("tend with no work item says how to run it", r.code == 2 && r.said(`^usage: tend <repo> <id>$`))
	}},
	{"at-once-two-items", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		f, release := s.hold()
		k := s.bg([]string{"HOLD=" + f}, `^agent prompt impl-`, s.repo, "261114-k0000")
		r := s.crun("261016-c0001")
		s.check("a tend on another work item runs while one is on its turn, and passes", r.passed())
		s.check("claiming it in a worktree of its own", len(s.worktrees()) == 2 && s.changeRecords("261016-c0001") == 1)
		release()
		s.check("the first reaches its verdict too", k.wait().passed())
		s.check("each implementer was prompted once", s.prompts("/grove-implement --runner 261114-k0000") == 1 && s.prompts("/grove-implement --runner 261016-c0001") == 1)
	}},
	{"at-once-intake", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		f, release := s.hold()
		k := s.bg([]string{"HOLD=" + f}, `^agent prompt impl-`, s.repo, "261114-k0000")
		r := s.propose()
		s.check("an intake pass runs beside a build", r.code == 0 && s.drafts() == 1 && exists(s.prOf("1")))
		release()
		s.check("and the build reaches its verdict", k.wait().passed())
	}},
	{"at-once-second-intake", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		f, release := s.hold()
		p := s.bg([]string{"HOLD=" + f}, `^agent prompt prop-`, "propose", s.repo)
		r := s.propose()
		s.check("a second intake pass for the repository is refused", r.code == 4)
		s.check("naming the first pass's pid", r.said(fmt.Sprintf("pid %d", p.c.Process.Pid)))
		release()
		s.check("the first finishes", p.wait().code == 0 && s.drafts() == 1)
	}},
	{"at-once-together", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "")
		var rs [2]res
		done := make(chan bool)
		for i, id := range []string{"261016-c0001", "261017-c0002"} {
			go func() { rs[i] = s.crun(id); done <- true }()
		}
		<-done
		<-done
		s.check("two tends started together both reach a verdict", rs[0].passed() && rs[1].passed())
		s.check("Grove holds one record for each", s.changeRecords("261016-c0001") == 1 && s.changeRecords("261017-c0002") == 1)
		s.check("each in a worktree of its own", len(s.worktrees()) == 2)
		s.check("with no tend-* branch left", s.git(s.repo, "branch", "--list", "tend-*") == "")
	}},
	{"at-once-refused", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "", "a.txt")
		s.change("261017-c0002", "Third", "", "a.txt")
		s.crun("261016-c0001", "IMPL_FAIL=stub_failure")
		before := s.git(s.repo, "worktree", "list") + s.git(s.repo, "branch", "--list")
		r := s.crun("261017-c0002")
		s.check("a piece that overlaps one in flight stops for a person", r.code == 3)
		s.check("with Grove's sentence", r.said(`261017-c0002 shares a\.txt with 261016-c0001`))
		s.check("leaving the worktrees and branches as they were", s.git(s.repo, "worktree", "list")+s.git(s.repo, "branch", "--list") == before)
	}},
	{"at-once-tidy", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.change("261016-c0001", "Another", "")
		s.run()
		wt, b, p := s.worktree(), s.branch(), s.pr()
		s.request()
		f, release := s.hold()
		k := s.bg([]string{"HOLD=" + f}, `^agent prompt impl-.*requested changes at commit`, s.repo, "261114-k0000")
		s.touch(p + ".merged")
		r := s.crun("261016-c0001")
		s.check("a closed pull request's worktree stays while a tend is on it", isDir(wt) && s.git(s.repo, "branch", "--list", b) != "")
		s.check("and tend says why", r.said(`left .*: a tend is on it`))
		release()
		k.wait()
		s.crun("261016-c0001")
		s.check("once that tend has exited, the next run removes them", !exists(wt) && s.git(s.repo, "branch", "--list", b) == "")
	}},
	{"at-once-during-turn", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		f, release := s.hold()
		k := s.bg([]string{"HOLD=" + f}, `^agent prompt impl-`, s.repo, "261114-k0000")
		s.crun("261016-c0001")
		release()
		r := k.wait()
		s.check("another attempt started during the turn does not stop this one", r.passed() && !r.said("left work outside"))
		s.check("whose head is reviewed", s.count(s.log(), `^agent prompt rev-`) == 2)
	}},
	{"at-once-stray", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		f, release := s.hold()
		k := s.bg([]string{"HOLD=" + f, "STRAY=worktree"}, `^agent prompt impl-`, s.repo, "261114-k0000")
		s.crun("261016-c0001")
		release()
		r := k.wait()
		s.check("a stray worktree beside another attempt's still stops the attempt", r.code == 3)
		s.check("naming the implementer's worktree", r.said(`it added the worktree .*stray-impl-`))
		s.check("not the other attempt's", !r.said(`it added the worktree .*wt-tend-`))
	}},
	{"at-once-setting-up", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		f, release := s.hold()
		k := s.bg([]string{"HOLD=" + f}, `^agent prompt impl-`, s.repo, "261114-k0000")
		g, letGo := s.hold()
		c := s.bg([]string{"WORKTREE_HOLD=" + g}, `^worktree create .*tend: 261016-c0001`, s.repo, "261016-c0001")
		s.waitFor(func() bool { return exists(g + ".made") }, "the second worktree")
		release()
		time.Sleep(1500 * time.Millisecond)
		s.check("a turn that ends while another attempt sets up waits for it", s.reviews() == 0)
		letGo()
		r := k.wait()
		s.check("then its head is reviewed", r.passed() && !r.said("left work outside"))
		s.check("and the other attempt passes", c.wait().passed())
	}},
	{"watch-at-once-two", func(s *sc) {
		s.atOnce(2)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "", "", "17")
		w := s.watchrepo()
		s.check("watch starts each of two claimable changes", w.said(`starting 261016-c0001 in `) && w.said(`starting 261017-c0002 in `))
		s.check("each claimed, built and reviewed", s.changeRecords("261016-c0001") == 1 && s.changeRecords("261017-c0002") == 1 && s.reviews() == 2)
		s.check("each in its own worktree", len(s.worktrees()) == 2 && s.records() == 0)
	}},
	{"watch-at-once-one", func(s *sc) {
		s.atOnce(1)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "", "", "17")
		w := s.watchrepo()
		s.check("a count of 1 starts the oldest only", w.said(`starting 261016-c0001`) && !w.said(`starting 261017-c0002`))
		s.check("which is reviewed", s.changeRecords("261016-c0001") == 1 && s.changeRecords("261017-c0002") == 0 && s.reviews() == 1)
		w = s.watchrepo()
		s.check("and the next pass starts the next", w.said(`starting 261017-c0002`) && s.changeRecords("261017-c0002") == 1)
	}},
	{"watch-at-once-overlap", func(s *sc) {
		s.atOnce(2)
		s.change("261016-c0001", "Another", "", "a.txt")
		s.change("261017-c0002", "Third", "", "a.txt", "17")
		w := s.watchrepo()
		s.check("of two that overlap, the first is built", s.changeRecords("261016-c0001") == 1 && w.said(`starting 261016-c0001`))
		s.check("and no worktree is made for the second", !w.said(`starting 261017-c0002`) && !s.has(s.log(), `^worktree create .*tend: 261017-c0002`) && s.changeRecords("261017-c0002") == 0)
		s.bash(s.repo, `grove archive 261016-c0001 >/dev/null && git add -A && git commit -qm landed && git push -q origin main`)
		w = s.watchrepo()
		s.check("once the target holds the first archived, the second is started", w.said(`starting 261017-c0002`) && s.changeRecords("261017-c0002") == 1)
	}},
	{"watch-at-once-needs-you", func(s *sc) {
		s.atOnce(1)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "", "", "17")
		s.crun("261016-c0001", "IMPL_FAIL=stub_failure")
		w := s.watchrepo()
		s.check("an attempt that needs a person frees its slot", w.said(`starting 261017-c0002`) && s.changeRecords("261017-c0002") == 1)
		s.check("and is left alone", !w.said(`261016-c0001`) && s.prompts("/grove-implement --runner 261016-c0001") == 1)
	}},
	{"watch-at-once-by-hand", func(s *sc) {
		s.atOnce(1)
		s.change("261016-c0001", "Another", "")
		f, release := s.hold()
		a := s.bg([]string{"HOLD=" + f}, `^agent prompt impl-`, s.repo, "261016-c0001")
		w := s.watchrepo()
		s.check("a tend run by hand counts", s.count(w.err, `starting`) == 0 && s.records() == 0)
		release()
		a.wait()
	}},
	{"watch-at-once-cycle", func(s *sc) {
		s.atOnce(1)
		w := s.watchrepo()
		s.check("a claimable cycle not marked unattended is started", w.said(`starting 261114-k0000 in `) && s.records() == 1 && s.reviews() == 1)
	}},
	{"watch-at-once-requested", func(s *sc) {
		s.atOnce(1)
		s.run()
		s.request()
		s.change("261016-c0001", "Another", "")
		w := s.watchrepo()
		s.check("an attempt with changes requested goes first", w.said(`changes requested on .*; running tend on 261114-k0000`) && s.prompts("teammate") == 1)
		s.check("and nothing else starts in that pass", !w.said(`starting`) && s.changeRecords("261016-c0001") == 0)
	}},
	{"watch-at-once-no-setting", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues()
		s.change("261016-c0001", "Another", "")
		w := s.watchrepo()
		s.check("without the setting, watch starts no change and no cycle not marked", s.count(w.err, `starting`) == 0 && s.records() == 0 && s.changeRecords("261016-c0001") == 0)
	}},
	{"watch-at-once-lower", func(s *sc) {
		s.atOnce(2)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "", "", "17")
		w := s.watchrepo("AT_ONCE=1")
		s.check("a lower number on this machine starts one", s.count(w.err, `starting`) == 1)
	}},
	{"watch-at-once-higher", func(s *sc) {
		s.atOnce(1)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "", "", "17")
		w := s.watchrepo("AT_ONCE=5")
		s.check("a higher number on this machine still starts one", s.count(w.err, `starting`) == 1)
	}},
	{"watch-at-once-zero", func(s *sc) {
		s.atOnce(2)
		s.change("261016-c0001", "Another", "")
		w := s.watchrepo("AT_ONCE=0")
		s.check("AT_ONCE=0 starts nothing", s.count(w.err, `starting`) == 0 && s.changeRecords("261016-c0001") == 0)
	}},
	{"watch-at-once-raised", func(s *sc) {
		s.atOnce(1)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "", "", "17")
		s.crun("261016-c0001", "IMPL_FAIL=stub_failure")
		s.bash(s.worktree(), `mkdir -p tend && printf 'watch:\n  at_once: 5\n' >tend/flow.yaml && git add -A && git commit -qm "Raise the count"`)
		w := s.watchrepo()
		s.check("a count an implementer commits is not watch's", s.count(w.err, `starting`) == 1)
	}},
	{"at-once-alias", func(s *sc) {
		s.project("main", true)
		f, release := s.hold()
		k := s.bg([]string{"HOLD=" + f}, `^agent prompt impl-`, s.repo, "261114-k0000")
		os.Symlink(s.repo, s.dir+"/alias")
		r := s.tendIn(nil, s.dir+"/alias", "261114-k0000")
		s.check("the same work item through another spelling of the path is refused", r.code == 4 && r.said(fmt.Sprintf("pid %d", k.c.Process.Pid)))
		for _, p := range []string{s.repo + "/grove", s.worktree()} {
			r = s.tendIn(nil, p, "261114-k0000")
			s.check("and through a subfolder or a linked worktree: "+p, r.code == 4)
		}
		release()
		k.wait()
	}},
	{"at-once-dependency-landed", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		s.change("261017-c0002", "Third", "261016-c0001")
		s.bash(s.dir, `git clone -q "$t/repo.git" "$t/other" && cd "$t/other" && grove archive 261016-c0001 >/dev/null && git add -A && git commit -qm landed && git push -q origin main`)
		r := s.crun("261017-c0002")
		s.check("a change whose dependency landed on origin is claimed, though the local target was behind", r.passed() && s.changeRecords("261017-c0002") == 1)
	}},
	{"watch-at-once-unnamed", func(s *sc) {
		s.atOnce(1)
		s.run()
		s.request()
		w := s.watch("AT_ONCE=0")
		s.check("watch given no repository still holds an attempt's rerun to its count", !w.said(`running tend`) && s.prompts("teammate") == 0)
		w = s.watch()
		s.check("and runs it within the count", w.said(`running tend on 261114-k0000`) && s.prompts("teammate") == 1)
	}},
	{"watch-at-once-subfolder", func(s *sc) {
		s.atOnce(1)
		s.tendIn(nil, s.repo+"/grove", "261114-k0000")
		s.request()
		s.change("261016-c0001", "Another", "")
		w := s.watchrepo()
		s.check("an attempt started through a subfolder is its repository's, and its rerun goes first", w.said(`running tend on 261114-k0000`) && s.prompts("teammate") == 1)
		s.check("so nothing new starts in that pass", !w.said(`starting`))
	}},
}

// ---- landing

// The built-in build flow with land: publish sends ok to it, or approved
// for a supervised flow, and a conflict and failed checks go back to the
// implementer once each with tend's own prompts.
func landFlow(t *testing.T, supervised bool) map[string]string {
	on := "        ok: land\n"
	if supervised {
		on = "        ok: end\n        approved: land\n"
	}
	return map[string]string{"tend/flow.yaml": buildFlow(t, "        ok: end\n        requested: request\n", on+`        requested: request
    land:
      next:
        ok: end
        conflict: resolve
        failed: fix
    resolve:
      continue: implement
      prompt: prompts/conflict.md
      result: commit
      limit: 1
      next:
        ok: review
    fix:
      continue: implement
      prompt: prompts/failed.md
      result: commit
      limit: 1
      next:
        ok: review
`)}
}

// A project listed for publication whose target's flow lands, and the checks
// every commit gets on GitHub, "<name> <conclusion>" a line.
func (s *sc) lands(supervised bool, ci string) {
	s.project("main", true)
	s.allow()
	s.commit("Land", landFlow(s.t, supervised))
	s.ci(ci)
}

func (s *sc) ci(checks string) { s.write(s.st+"/ci", checks+"\n") }

// What origin's target holds: its newest commit's subject and body.
func (s *sc) landedMsg() string { return s.git(s.repo+".git", "log", "-1", "--format=%B", "main") }
func (s *sc) merges() int       { return s.count(s.log(), `^gh api -X PUT repos/.*/pulls/1/merge `) }

// The heads GitHub was asked to merge pull requests at, oldest first.
func (s *sc) mergedAt() []string { return strings.Fields(s.cat(s.st + "/merge-requests")) }
func (s *sc) merged() bool       { return exists(s.pr() + ".merged") }
func (s *sc) approve(h string)   { s.reviewOn(h, "teammate", "COLLABORATOR", "APPROVED") }

// The parents of commit c in the worktree, c first.
func (s *sc) parents(c string) []string {
	return strings.Fields(s.git(s.worktree(), "rev-list", "--parents", "-n", "1", c))
}

var landScenarios = []scenario{
	{"land-landed", func(s *sc) {
		s.lands(false, "check success")
		before := s.git(s.repo+".git", "rev-parse", "main")
		r := s.run()
		s.check("a candidate whose checks pass lands, and tend exits 0", r.passed())
		s.check("the target gains one commit", s.git(s.repo+".git", "rev-parse", "main^") == before)
		s.check("the target gains its squash commit, titled with the pull request's title and number", strings.HasPrefix(s.landedMsg(), "Add the file (#1)\n\nThe project needed a file."))
		s.check("whose body ends with the trailer", strings.HasSuffix(s.landedMsg(), "Grove: 261015-c0000"))
		s.check("merged by the publishing account", s.cat(s.pr()+".merged-by") == "the-bot")
		s.check("at exactly the reviewed head", slices.Equal(s.mergedAt(), []string{s.head()}))
		s.check("the target did not move, so no merge commit is made", s.git(s.worktree(), "rev-list", "--merges", "origin/main..HEAD") == "")
		s.check("status says done and that it landed", s.has(s.status(), `^[^ ]* done: landed https://example.test/pr/`))
		s.check("no agent was asked to merge it: the implementer, its message and one reviewer", s.count(s.log(), `^agent prompt `) == 3)
	}},
	{"land-none", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.ci("check success")
		r := s.run()
		s.check("without a flow file, a passing candidate is published", r.passed() && s.creates() == 1)
		s.check("and tend merges nothing", !s.merged() && s.merges() == 0)
	}},
	{"land-failed-head", func(s *sc) {
		s.lands(false, "check success")
		r := s.run("REVIEW=decision")
		s.check("a head that failed review is published", s.creates() == 1)
		s.check("then tend stops at land, saying it did not pass review", r.code == 3 && r.said(`did not pass review, so tend lands nothing`))
		s.check("and nothing is merged", !s.merged() && s.merges() == 0)
	}},
	{"land-unlisted", func(s *sc) {
		s.project("main", true)
		s.commit("Land", landFlow(s.t, false))
		r := s.run()
		s.check("with origin not listed, land stops saying there is no pull request to land", r.code == 3 && r.said(`there is no pull request to land`))
		s.check("and runs no gh command", s.count(s.log(), `^gh `) == 0)
	}},
	{"land-no-ci", func(s *sc) {
		s.lands(false, "")
		r := s.run("TIMEOUT_MS=400")
		s.check("a head with no check but tend/review stops for a person", r.code == 3 && r.said(`has no check but tend/review`))
		s.check("and nothing is merged", !s.merged() && s.merges() == 0)
		// Run again, a person merges the pull request while tend waits for a check.
		s.ci("check pending")
		reads := s.count(s.log(), `^gh api repos/.*/check-runs`)
		b := s.bg([]string{"TIMEOUT_MS=20000"}, ``, s.repo, "261114-k0000")
		s.waitFor(func() bool { return s.count(s.log(), `^gh api repos/.*/check-runs`) > reads }, "tend to wait for a check")
		s.touch(s.pr() + ".merged")
		r = b.wait()
		s.check("a pull request a person merged before tend merges it gives ok, and tend exits 0", r.code == 0 && r.said(`is merged`))
		s.check("merging nothing more", s.merges() == 0)
	}},
	{"land-target-moved", func(s *sc) {
		s.lands(true, "check success")
		r := s.run()
		s.check("a supervised flow's first run merges nothing and exits 0", r.passed() && !s.merged() && s.merges() == 0)
		reviewed, theirs := s.head(), s.pushed()
		s.approve(reviewed)
		r = s.run()
		s.check("once a teammate approves, the next run lands it", r.passed() && s.merged())
		m := s.head()
		s.check("the branch's head is a merge of the reviewed head and the target's tip", slices.Equal(s.parents(m), []string{m, reviewed, theirs}))
		s.check("origin's branch was at it", slices.Equal(s.mergedAt(), []string{m}))
		s.check("pushed by no force", !s.has(s.log(), `push.*(--force|-f |\+HEAD)`))
		s.check("the merge commit has tend/review success, set by the account", s.has(s.cat(s.st+"/statuses"), `^`+m+` success the-bot$`))
		s.check("described as carried from the reviewed head", s.has(s.cat(s.st+"/descriptions"), `^`+m+` Carried from `+reviewed[:7]))
		s.check("with no review of the merge commit", s.reviewed(m) == 0)
		s.check("and one reviewer in the attempt", s.count(s.log(), `^agent start rev-`) == 1)
		s.check("the target holds both", s.gitOK(s.repo+".git", "merge-base", "--is-ancestor", theirs, "main") && strings.HasPrefix(s.landedMsg(), "Add the file (#1)"))
	}},
	{"land-rerun-after-kill", func(s *sc) {
		s.lands(true, "check pending")
		s.run()
		reviewed := s.head()
		s.pushed()
		s.approve(reviewed)
		b := s.bg(nil, `^gh api repos/.*/commits/.*/check-runs`, s.repo, "261114-k0000")
		killed(b.c)
		m := s.git(s.repo+".git", "rev-parse", s.branch())
		s.check("killed while waiting, the merge commit is pushed", m != reviewed && s.parents(m)[1] == reviewed)
		s.ci("check success")
		w := s.watch()
		s.check("watch runs it again on the approval of the head it was carried from", w.said(`was approved; running tend on 261114-k0000`))
		s.check("and the rerun lands it", s.merged())
		s.check("prompting no reviewer", s.reviews() == 1)
		s.check("asking the implementer for no message", s.prompts("Branch") == 1)
		s.check("and leaving the title and description as they were", s.cat(s.pr()+".title") == "Add the file" && s.count(s.log(), `^gh pr edit`) == 0)
	}},
	{"land-merge-edit", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.pushed()
		s.bash(s.worktree(), `git fetch -q origin && git merge -q --no-commit origin/main && echo mine >>a.txt && git add a.txt && git commit -qm "Merge, with an edit"`)
		r := s.run()
		s.check("a merge with an edit of its own is not carried: a reviewer is prompted for it", r.passed() && s.reviews() == 2 && s.reviewed(s.head()) == 1)
	}},
	{"land-conflict", func(s *sc) {
		s.lands(true, "check success")
		s.run()
		s.bash(s.worktree(), `echo mine >a.txt && git commit -qam mine`)
		s.run()
		reviewed := s.head()
		s.bash(s.dir, `git clone -q "$t/repo.git" "$t/other" && cd "$t/other" && echo theirs >a.txt && git commit -qam theirs && git push -q origin main`)
		s.approve(reviewed)
		r := s.run()
		at := strings.Split(s.cat(s.st+"/at-conflict"), "\n")
		s.check("on a conflict, the implementer is asked once with tend's own prompt", s.prompts("tend could not land") == 1 && s.has(s.log(), `^agent prompt impl-.* does not merge into it cleanly`))
		s.check("finding the branch at the reviewed head, with no merge under way and nothing uncommitted", len(at) == 1 && at[0] == reviewed)
		s.check("the head it makes is reviewed", r.code == 0 && s.reviews() == 3 && s.reviewed(s.head()) == 1 && s.parents(s.head())[1] == reviewed)
		s.check("and nothing lands, its approval being of the earlier head", !s.merged() && s.merges() == 0)
	}},
	{"land-check-fails", func(s *sc) {
		s.lands(true, "check failure")
		s.run()
		reviewed := s.head()
		s.pushed()
		s.approve(reviewed)
		r := s.run()
		s.check("a check failing on the merged head merges nothing", r.code == 0 && !s.merged() && s.merges() == 0)
		s.check("the implementer is asked with tend's own prompt, naming the failed check and its link", s.has(s.log(), `^agent prompt impl-[^ ]* tend could not land .*check https://example\.test/checks/check`))
		s.check("the head it makes is reviewed", s.reviews() == 2 && s.reviewed(s.head()) == 1)
		s.ci("check success")
		s.run()
		s.check("nothing lands until a person approves the new head", !s.merged())
		s.approve(s.head())
		r = s.run()
		s.check("then it lands", r.passed() && s.merged())
	}},
	{"land-cancelled", func(s *sc) {
		s.lands(false, "build success\ncheck cancelled")
		r := s.run()
		s.check("a cancelled check merges nothing", !s.merged() && s.merges() == 0)
		s.check("and the prompt names the cancelled check, not the one that passed", s.prompts("tend could not land") >= 1 && s.has(s.log(), `^agent prompt impl-[^ ]* tend could not land .*: check https://`) && !s.has(s.log(), `^agent prompt impl-[^ ]* tend could not land .*build https://`))
		s.check("tend stops once the step is spent", r.code == 3 && r.said(`fix was asked 1 times, its limit`))
	}},
	{"land-skipped", func(s *sc) {
		s.lands(false, "build success\nlint skipped\nstyle neutral")
		s.check("with one check passed, one skipped and one neutral, it lands", s.run().passed() && s.merged())
	}},
	{"land-moved-before-merge", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BEFORE_MERGE=git clone -q "$t/repo.git" "$t/other" && cd "$t/other" && echo theirs >b.txt && git add b.txt && git commit -qm theirs && git push -q origin main`)
		s.check("a target that moves before the merge is merged in again, and the candidate lands", r.passed() && s.merged())
		s.check("GitHub was asked twice", s.merges() == 2)
		m := s.head()
		s.check("the second time at the new merge commit", len(s.parents(m)) == 3 && len(s.mergedAt()) == 2 && s.mergedAt()[1] == m)
		s.check("whose checks were waited for", s.has(s.log(), `^gh api repos/.*/commits/`+m+`/check-runs`))
		s.check("and the target holds both", s.git(s.repo+".git", "show", "main:b.txt") == "theirs" && strings.HasPrefix(s.landedMsg(), "Add the file (#1)"))
	}},
	{"land-late-check", func(s *sc) {
		s.lands(false, "title success")
		r := s.run("CI_LATE=check", "TIMEOUT_MS=4000")
		s.check("a merge refused for a required check not yet reported waits, then lands", r.passed() && s.merged())
		s.check("stopping for no one, and asking GitHub twice", s.merges() == 2 && r.said(`for a check it still waits for`))
	}},
	{"land-edited-description", func(s *sc) {
		s.lands(true, "check success")
		s.run()
		s.write(s.pr(), "Their own words.\n")
		s.approve(s.head())
		s.run()
		s.check("a description a person edited is the commit's body", s.landedMsg() == "Add the file (#1)\n\nTheir own words.")
	}},
	{"land-retargeted", func(s *sc) {
		s.lands(true, "check success")
		s.run()
		s.write(s.pr()+".base", "release\n")
		s.approve(s.head())
		r := s.run()
		s.check("a pull request a person retargeted stops for a person", r.code == 3 && r.said(`merges into release, not main, so tend lands nothing`))
		s.check("merging nothing", !s.merged() && s.merges() == 0)
	}},
	{"land-refused", func(s *sc) {
		s.lands(false, "check success")
		r := s.run("REQUIRE_APPROVAL=1")
		s.check("a merge GitHub refuses for a rule stops, quoting GitHub", r.code == 3 && r.said(`GitHub refused to merge .*At least 1 approving review is required`))
		s.check("and the pull request stays open", !s.merged())
	}},
	{"land-two", func(s *sc) {
		s.lands(false, "check success")
		s.change("261016-c0001", "Another", "")
		f, release := s.hold()
		a := s.bg([]string{"MERGE_HOLD=" + f}, `^gh api -X PUT repos/.*/merge `, s.repo, "261114-k0000")
		b := s.bg(nil, `^agent prompt impl-[^ ]* /grove-implement --runner 261016-c0001`, s.repo, "261016-c0001")
		s.waitFor(func() bool {
			return strings.Contains(s.tendIn(nil, "status", s.repo, "261016-c0001").out, "waits to land")
		}, "the second to wait its turn")
		release()
		s.check("the first lands", a.wait().code == 0)
		first := s.git(s.repo+".git", "rev-parse", "main")
		s.check("the second lands after it", b.wait().code == 0)
		bb := s.git(s.repo+".git", "rev-parse", "main^")
		s.check("the target holds both, the first's below the second's", bb == first && s.count(s.git(s.repo+".git", "log", "--format=%s", "main"), `^Add the file \(#1\)$`) == 2)
		h := s.git(s.repo+".git", "rev-parse", "main^{tree}")
		s.check("the second's branch gained a merge of the target holding the first's commit", s.merges() == 2 && s.has(s.log(), `^gh api repos/.*/statuses/`) && h != "")
		var merge string // the second's, which merges last
		if hs := s.mergedAt(); len(hs) == 2 {
			merge = hs[1]
		}
		s.check("whose second parent holds the first", merge != "" && s.gitOK(s.repo+".git", "merge-base", "--is-ancestor", first, merge) && len(strings.Fields(s.git(s.repo+".git", "rev-list", "--parents", "-n", "1", merge))) == 3)
	}},
	{"approve-stranger", func(s *sc) {
		s.lands(true, "check success")
		s.run()
		s.reviewOn(s.head(), "stranger", "NONE", "APPROVED")
		r := s.run()
		s.check("a stranger's approval merges nothing, and tend exits 0", r.code == 0 && !s.merged() && s.merges() == 0)
		w := s.watch()
		s.check("nor does watch run tend for it", s.count(w.err, `running tend`) == 0)
	}},
	{"approve-built-in", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.ci("check success")
		s.run()
		s.approve(s.head())
		w := s.watch()
		s.check("with no flow file, watch runs nothing for an approval", s.count(w.err, `running tend`) == 0)
		r := s.run()
		s.check("and a run merges nothing and exits 0 as before", r.passed() && !s.merged() && s.merges() == 0)
	}},
	{"approve-watch", func(s *sc) {
		s.lands(true, "check success")
		s.run()
		w := s.watch()
		s.check("watch runs nothing before an approval", s.count(w.err, `running tend`) == 0)
		s.approve(s.head())
		w = s.watch()
		s.check("once a teammate approves, watch says the pull request was approved", w.said(`https://example.test/pr/[^ ]* was approved; running tend on 261114-k0000`))
		s.check("and lands it", s.merged())
	}},
	{"approve-needs-you", func(s *sc) {
		s.lands(true, "check success")
		s.run()
		s.approve(s.head())
		s.write(s.runDir()+"/status", "x needs-you: something\nnext: a person\n")
		w := s.watch()
		s.check("watch leaves an approved attempt that needs a person", s.count(w.err, `running tend`) == 0 && !s.merged())
	}},
}

func TestLanding(t *testing.T) {
	for _, x := range landScenarios {
		t.Run(x.name, func(t *testing.T) { x.fn(newSc(t)) })
	}
}
