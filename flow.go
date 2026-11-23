package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ---- flows
//
// A flow is a named set of steps an attempt goes through once it has its
// branch and worktree. A step is an agent given a prompt, a result tend
// checks, and what each outcome of that result allows next: another step, end
// (the flow is done) or person (tend stops for one). tend runs two flows,
// build and propose; a project changes them with tend/flow.yaml on its
// target branch, each prompt a file beside it, and tend ships its own in
// flows/ for every flow that file does not hold.

type step struct {
	Agent    string            `yaml:"agent,omitempty"`    // claude or codex: a fresh agent of that harness
	Continue string            `yaml:"continue,omitempty"` // or an earlier step, whose agent is prompted again
	Prompt   string            `yaml:"prompt,omitempty"`   // a file beside the flow file
	Result   string            `yaml:"result,omitempty"`   // what tend checks once the turn ends
	Limit    int               `yaml:"limit,omitempty"`    // asked at most this often in an attempt
	Next     map[string]string `yaml:"next,omitempty"`     // outcome: step, end or person
}

type flow struct {
	Start string           `yaml:"start"`
	Steps map[string]*step `yaml:"steps"`
}

type flowFile struct {
	Watch   *watchSetting `yaml:"watch,omitempty"`
	Build   *flow         `yaml:"build,omitempty"`
	Propose *flow         `yaml:"propose,omitempty"`
}

// How many attempts tend watch keeps going in the repository: with it, watch
// starts what Grove calls claimable; without, it starts only cycles marked
// unattended.
type watchSetting struct {
	AtOnce any `yaml:"at_once"` // a whole number of at least 1, checked by fault
}

// Why the setting is not valid, said after the file's name, or "".
func (w *watchSetting) fault() string {
	if w == nil {
		return ""
	}
	if n, isInt := w.AtOnce.(int); !isInt || n < 1 {
		if w.AtOnce == nil {
			return "gives watch no at_once, where at_once is a whole number of at least 1"
		}
		return fmt.Sprintf("gives watch an at_once of %v, where at_once is a whole number of at least 1", w.AtOnce)
	}
	return ""
}

// The count, or -1 for none.
func (w *watchSetting) count() int {
	if w != nil {
		if n, isInt := w.AtOnce.(int); isInt {
			return n
		}
	}
	return -1
}

func (ff *flowFile) named() map[string]**flow {
	return map[string]**flow{"build": &ff.Build, "propose": &ff.Propose}
}

//go:embed flows
var builtin embed.FS

const flowPath = "tend/flow.yaml"

// A step's result, as the flow says it: commit (a new commit on the branch),
// file <name> (a file in the run folder that is not empty), verdict <name> (a
// verdict file as the review spec defines one), or commit or file <name>.
// None at all is a turn that ends.
type result struct {
	commit        bool
	file, verdict string // names in the run folder, placeholders not yet filled
}

func parseResult(s string) (result, bool) {
	var r result
	f := strings.Fields(s)
	switch {
	case len(f) == 0:
	case len(f) == 1 && f[0] == "commit":
		r.commit = true
	case len(f) == 2 && f[0] == "file":
		r.file = f[1]
	case len(f) == 2 && f[0] == "verdict":
		r.verdict = f[1]
	case len(f) == 4 && f[0] == "commit" && f[1] == "or" && f[2] == "file":
		r.commit, r.file = true, f[3]
	default:
		return r, false
	}
	return r, !strings.Contains(r.file+r.verdict, "/")
}

// The outcomes a step can have, which are what its next may name.
func (s *step) outcomes(name, flowName string) []string {
	var o []string
	r, _ := parseResult(s.Result)
	switch {
	case name == "publish" && flowName == "build":
		o = []string{"ok", "requested", "approved"}
	case name == "land":
		o = []string{"ok", "conflict", "failed"}
	case r.verdict != "":
		o = []string{"success", "actionable", "failure"}
	default:
		o = []string{"ok"}
	}
	if s.Limit > 0 {
		o = append(o, "spent")
	}
	return o
}

func sorted[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func and(xs []string) string {
	if len(xs) < 2 {
		return strings.Join(xs, "")
	}
	return strings.Join(xs[:len(xs)-1], ", ") + " and " + xs[len(xs)-1]
}

// Why the flow cannot run, said after the file's name, or "". has says
// whether a prompt file exists.
func (f *flow) fault(name string, has func(string) bool) string {
	if len(f.Steps) == 0 {
		return fmt.Sprintf("gives flow %s no steps", name)
	}
	if f.Steps[f.Start] == nil {
		return fmt.Sprintf("starts flow %s at %q, a step it does not have", name, f.Start)
	}
	kinds := []string{""}
	if name == "build" {
		kinds = []string{"grove", "issue"} // {kind} in a prompt's name
	}
	for _, n := range sorted(f.Steps) {
		s := f.Steps[n]
		at := fmt.Sprintf("gives step %s of flow %s", n, name)
		if s == nil {
			return at + " nothing"
		}
		if n == "end" || n == "person" {
			return fmt.Sprintf("names a step %s in flow %s, a word kept for what follows a step", n, name)
		}
		if n == "publish" || n == "land" {
			if s.Agent != "" || s.Continue != "" || s.Prompt != "" || s.Result != "" || s.Limit != 0 {
				return fmt.Sprintf("gives step %s of flow %s more than next, though %s is tend's own", n, name, map[string]string{"publish": "publishing", "land": "landing"}[n])
			}
			if n == "land" && name != "build" {
				return fmt.Sprintf("names a step land in flow %s, where tend lands nothing", name)
			}
			if n == "land" && f.Steps["publish"] == nil {
				return fmt.Sprintf("gives flow %s a step land and no publish, where land lands the pull request publish opens", name)
			}
		} else {
			if (s.Agent == "") == (s.Continue == "") {
				return at + " both or neither of agent and continue, where it needs one"
			}
			if s.Agent != "" && s.Agent != "claude" && s.Agent != "codex" {
				return fmt.Sprintf("%s the harness %s, where an agent is claude or codex", at, s.Agent)
			}
			if s.Continue != "" {
				root, seen := n, 0
				for f.Steps[root] != nil && f.Steps[root].Continue != "" && seen <= len(f.Steps) {
					root, seen = f.Steps[root].Continue, seen+1
				}
				if c := f.Steps[root]; c == nil || root == "publish" || root == "land" || c.Agent == "" {
					return fmt.Sprintf("%s %q to continue, which is no step with an agent", at, s.Continue)
				}
			}
			if s.Prompt == "" {
				return at + " no prompt"
			}
			for _, k := range kinds {
				if p := strings.ReplaceAll(s.Prompt, "{kind}", k); !has(p) {
					return fmt.Sprintf("%s the prompt file %s, which does not exist", at, p)
				}
			}
			if _, ok := parseResult(s.Result); !ok {
				return fmt.Sprintf("%s the result %q, where a result is commit, file <name>, verdict <name> or commit or file <name>", at, s.Result)
			}
			if s.Limit < 0 {
				return at + " a limit below 0"
			}
		}
		outs := s.outcomes(n, name)
		for _, o := range sorted(s.Next) {
			if !slices.Contains(outs, o) {
				return fmt.Sprintf("%s the outcome %s, which it cannot have; it has %s", at, o, and(outs))
			}
			if to := s.Next[o]; to != "end" && to != "person" && f.Steps[to] == nil {
				return fmt.Sprintf("%s %q after %s, a step the flow does not have", at, to, o)
			}
		}
	}
	if l := f.loop(); l != nil {
		return fmt.Sprintf("loops through %s with no limit", and(l))
	}
	return ""
}

// The steps of a loop no step of which has a limit, or nil. A loop through
// publish's requested or approved comes round only when a person requests
// changes or approves, so people bound it.
func (f *flow) loop() []string {
	state := map[string]int{} // 1 on the path, 2 done
	var path []string
	var visit func(n string) []string
	visit = func(n string) []string {
		s := f.Steps[n]
		if s == nil || s.Limit > 0 {
			return nil
		}
		if state[n] == 1 {
			return slices.Clone(path[slices.Index(path, n):])
		}
		if state[n] == 2 {
			return nil
		}
		state[n] = 1
		path = append(path, n)
		for _, o := range sorted(s.Next) {
			if n == "publish" && (o == "requested" || o == "approved") {
				continue
			}
			if l := visit(s.Next[o]); l != nil {
				return l
			}
		}
		path = path[:len(path)-1]
		state[n] = 2
		return nil
	}
	for _, n := range append([]string{f.Start}, sorted(f.Steps)...) {
		if l := visit(n); l != nil {
			return l
		}
	}
	return nil
}

// Why the text is no flow file, or "".
func decode(s string, ff *flowFile) string {
	d := yaml.NewDecoder(strings.NewReader(s))
	d.KnownFields(true)
	if err := d.Decode(ff); err == io.EOF {
		return "is empty"
	} else if err != nil {
		msg := strings.Join(strings.Fields(strings.TrimPrefix(err.Error(), "yaml: ")), " ")
		return "is not a flow file tend can read: " + strings.NewReplacer("unmarshal errors: ", "",
			"in type main.step", "in a step", "in type main.flowFile", "at the top, where a flow is build or propose",
			"in type main.flow", "in a flow", "in type main.watchSetting", "in watch").Replace(msg)
	}
	return ""
}

// A prompt file: the target's under tend/ when rev is given and holds it,
// else tend's own.
func promptText(rev, p string) (string, bool) {
	if rev != "" {
		if s, err := quiet("git", "-C", repo, "show", rev+":tend/"+p); err == nil {
			return s, true
		}
	}
	b, err := builtin.ReadFile("flows/" + p)
	return strings.TrimRight(string(b), "\n"), err == nil
}

// The flows in effect at rev, the target as origin holds it: each one the
// project's tend/flow.yaml holds, and tend's own for the rest, with where
// each came from. A flow that cannot run stops tend, exit 3, before any agent
// starts.
func flowsAt(rev string) (flowFile, map[string]string) {
	var ff flowFile
	from := map[string]string{}
	if rev != "" && ok("git", "-C", repo, "cat-file", "-e", rev+":"+flowPath) {
		s, _ := quiet("git", "-C", repo, "show", rev+":"+flowPath)
		if why := decode(s, &ff); why != "" {
			stop("%s %s", flowPath, why)
		}
		if why := ff.Watch.fault(); why != "" {
			stop("%s %s", flowPath, why)
		}
		has := func(p string) bool { _, ok := promptText(rev, p); return ok }
		for name, f := range ff.named() {
			if *f == nil {
				continue
			}
			if why := (*f).fault(name, has); why != "" {
				stop("%s %s", flowPath, why)
			}
			from[name] = flowPath + " on " + rev
		}
	}
	for name, f := range ff.named() {
		if *f != nil {
			continue
		}
		var own flowFile
		b, _ := builtin.ReadFile("flows/" + name + ".yaml")
		if why := decode(string(b), &own); why != "" {
			stop("flows/%s.yaml %s", name, why)
		}
		*f = *own.named()[name]
		if why := (*f).fault(name, func(p string) bool { _, ok := promptText("", p); return ok }); why != "" {
			stop("flows/%s.yaml %s", name, why)
		}
		from[name] = "built in"
	}
	return ff, from
}

// The count tend watch keeps going at rev, the target as origin holds it, or
// -1 for none; and why the flow file is not valid, or "". Stops nothing.
func watchCount(rev string) (int, string) {
	if !ok("git", "-C", repo, "cat-file", "-e", rev+":"+flowPath) {
		return -1, ""
	}
	s, _ := quiet("git", "-C", repo, "show", rev+":"+flowPath)
	var ff flowFile
	if why := decode(s, &ff); why != "" {
		return -1, flowPath + " " + why
	}
	if why := ff.Watch.fault(); why != "" {
		return -1, flowPath + " " + why
	}
	return ff.Watch.count(), ""
}

// tend flow <repo>: the flows the repository's attempts would run, as a flow
// file, each said to be the project's or tend's own. Reads only.
func printFlows() {
	cd(repo)
	t := ""
	if g, err := groveJSON(); err == nil {
		t = str(get(g, "target"))
	} else if s, err := quiet("gh", "repo", "view", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name"); err == nil {
		t = s
	} else {
		t = strings.TrimPrefix(gitQ("symbolic-ref", "--short", "refs/remotes/origin/HEAD"), "origin/")
	}
	if t == "" {
		t = gitQ("branch", "--show-current")
	}
	rev := t
	if _, err := quiet("git", "remote", "get-url", "origin"); err == nil {
		quiet("git", "fetch", "-q", "origin", t)
		rev = "origin/" + t
	}
	ff, from := flowsAt(rev)
	fmt.Printf("# The flows tend runs in %s, read from %s. Saved as %s, with\n# the prompt files it names under tend/ where yours differ from tend's own,\n# it is what attempts run.\n", repo, rev, flowPath)
	if ff.Watch != nil {
		fmt.Printf("\n# watch: %s on %s\nwatch:\n  at_once: %d\n", flowPath, rev, ff.Watch.count())
	}
	for _, name := range []string{"build", "propose"} {
		fmt.Printf("\n# %s: %s\n", name, from[name])
		if from[name] == "built in" {
			b, _ := builtin.ReadFile("flows/" + name + ".yaml")
			os.Stdout.Write(b)
			continue
		}
		b, _ := yaml.Marshal(map[string]*flow{name: *ff.named()[name]})
		os.Stdout.Write(b)
	}
	os.Exit(0)
}

// ---- the engine

var (
	fl        *flow
	flowName  string            // build or propose
	flowFrom  string            // where the flow was read
	promptRev string            // the target, for a project's flow, whose prompt files come first
	vars      map[string]string // what prompts are filled in with
	used      map[string]int    // the asks in each record this walk has answered
	reviewReq any               // the change request publish found, for the step requested leads to
	reqStep   string
	lastV     any // the verdict last reached, whose summary is {findings}
)

// The flow of the name in effect at rev, ready to walk.
func loadFlow(name, rev string) {
	ff, from := flowsAt(rev)
	fl, flowName, flowFrom = *ff.named()[name], name, from[name]
	promptRev = ""
	if from[name] != "built in" {
		promptRev = rev
	}
	reqStep = ""
	if p := fl.Steps["publish"]; p != nil {
		reqStep = p.Next["requested"]
	}
}

func fill(s string) string {
	var kv []string
	for k, v := range vars {
		kv = append(kv, "{"+k+"}", v)
	}
	return strings.NewReplacer(kv...).Replace(s)
}

// Go through the flow from its start. Every step is reached afresh on each
// run, its record and result in the run folder saying whether it is done, so
// a rerun after any stop carries on where the last one was. publish is tend's
// own, given here.
var prompted int // prompts sent in this run

func walkFlow(publish func() string) {
	used, reviewReq, lastV = map[string]int{}, nil, nil
	at := fl.Start
	been := map[string]bool{}
	for {
		s := fl.Steps[at]
		// A step reached again at the same head with nothing asked or taken
		// since would only give the same outcome again, for ever.
		n := prompted
		for _, u := range used {
			n += u
		}
		state := fmt.Sprintf("%s %s %d", at, git("rev-parse", "HEAD"), n)
		if been[state] {
			stop("%s reached %s again at %s with nothing changed, so %s loops; a person decides", flowName, at, short(git("rev-parse", "HEAD")), flowFrom)
		}
		been[state] = true
		var o string
		if at == "publish" {
			o = publish()
		} else if at == "land" {
			o = landStep()
		} else {
			o = reach(at, s)
		}
		to := s.Next[o]
		switch to {
		case "end":
			if flowName == "build" {
				if why := unbuilt(); why != "" {
					stop("%s", why)
				}
			}
			return
		case "", "person":
			why := at + " gave " + o
			if o == "spent" {
				why = fmt.Sprintf("%s was asked %d times, its limit", at, s.Limit)
			} else if r, _ := parseResult(s.Result); r.verdict != "" {
				why += ": " + str(get(lastV, "summary"))
			}
			if to == "" {
				stop("%s, and %s gives it nothing to go to, so a person decides", why, flowFrom)
			}
			stop("%s, and %s sends that to a person", why, flowFrom)
		}
		// The candidate is finished before anything but a commit step follows.
		if flowName == "build" {
			if r, _ := parseResult(fl.Steps[to].Result); to == "publish" || !r.commit {
				if why := unbuilt(); why != "" {
					stop("%s", why)
				}
			}
		}
		at = to
	}
}

// The record of the step's asks in the run folder, a head a line: the names
// the specs give tend's own steps, and <step>-asked for the rest.
func recordOf(name string) string {
	switch {
	case reviewReq != nil && name == reqStep:
		return run + "/request-" + str(get(reviewReq, "id"))
	case flowName == "propose" && name == "draft":
		return run + "/asked"
	case flowName == "propose" && name == "follow":
		return run + "/followed"
	}
	return run + "/" + name + "-asked"
}

func asks(rec string) []string { return strings.Fields(cat(rec)) }

// How often a step was asked in the attempt, for its limit: its record's
// asks, those of every review's record for requested changes, and the
// finished asks of the implementation, which leave its record.
func timesAsked(name, rec string) int {
	n := len(asks(run + "/" + name + "-done"))
	if reviewReq != nil && name == reqStep {
		fs, _ := filepath.Glob(run + "/request-*")
		for _, f := range fs {
			if !strings.HasSuffix(f, ".tmp") {
				n += len(asks(f))
			}
		}
		return n
	}
	return n + len(asks(rec))
}

func setAsks(rec string, hs []string) {
	if len(hs) == 0 {
		os.Remove(rec)
		return
	}
	write(rec+".tmp", strings.Join(hs, "\n")+"\n")
	must("", os.Rename(rec+".tmp", rec))
}

// The agent a step prompts, its harness, and the step that started it: the
// first step's is the implementer or drafter, kept for the attempt; any other
// fresh agent is new for each head, as a reviewer is.
func agentOf(name string) (string, string, string) {
	root := name
	for fl.Steps[root].Continue != "" {
		root = fl.Steps[root].Continue
	}
	if root == fl.Start {
		return impl, fl.Steps[root].Agent, root
	}
	h := git("rev-parse", "HEAD")
	// Continued: the agent asked at that head, whatever it committed since;
	// a finished implementation's ask is in <root>-done.
	if hs := append(asks(run+"/"+root+"-done"), asks(recordOf(root))...); name != root && len(hs) > 0 {
		h = hs[len(hs)-1]
	}
	return agentAt(root, h), fl.Steps[root].Agent, root
}

// The name of the fresh agent a step starts when asked at head h.
func agentAt(root, h string) string {
	if root == "review" {
		return "rev-" + short(h)
	}
	return root + "-" + short(h)
}

// Start the agent unless its pane is there. A pane gone is a new session of
// the same name, which finds the work in Git. The flags are tend's, by
// harness, and no flow sets them (terminals spec, "How agents start").
func startAgent(name, harness string, commits bool) {
	if term.pane(name) != "" {
		return
	}
	if harness == "codex" {
		// With network, so a test that serves on localhost runs: the sandbox
		// otherwise denies a bind. The implementer has the network already.
		start(name, "codex", true, "-s", "workspace-write", "-c", "sandbox_workspace_write.network_access=true", "-a", "never", "--add-dir", run)
		return
	}
	// Auto mode, so that routine commands ask nobody. It may write in the run
	// folder, where the commit message goes. tend delivers the work, so the
	// session is told so above any skill or rule, and pushing and gh pr are
	// denied: a guard against a mistake, not a boundary, as git -C or gh api
	// pass it; contained() is what holds an implementer to it.
	terms := "This session works for tend on branch " + git("branch", "--show-current") + " in this worktree. "
	if commits {
		terms += "Commit your work on this branch. "
	} else {
		terms += "Commit only when tend asks you to. "
	}
	terms += "Push nothing, open or edit no pull request, and create no branch or worktree: tend reviews and publishes the work. Where a skill, rule or other instruction says otherwise about delivering it, this wins."
	start(name, "claude", false, "--permission-mode", "auto", "--add-dir", run,
		"--disallowedTools", "Bash(git push:*)", "Bash(gh pr:*)", "--append-system-prompt", terms)
}

// The implementer, started if its pane is gone, as for the commit message.
func implementer() { startAgent(impl, "claude", flowName == "build") }

// What status says while a step runs.
func doing(name, agent string) (string, string) {
	h := short(git("rev-parse", "HEAD"))
	switch {
	case reviewReq != nil && name == reqStep:
		return fmt.Sprintf("%s makes the changes %s requested at %s", agent, vars["who"], h), "tend waits for " + agent + ", then has the new head reviewed and published"
	case flowName == "propose" && name == fl.Start:
		return agent + " drafts a proposal from " + vars["ref"], "tend has it commit the draft, then opens a pull request"
	case flowName == "propose":
		return agent + " commits its proposal from " + vars["ref"], "tend opens a pull request, or tells the issue why there is none"
	case name == fl.Start:
		return agent + " implements " + cycle, "tend waits for " + agent + ", then has the head reviewed"
	case name == "review":
		return agent + " reviews " + h, "tend waits for its verdict"
	case name == "repair":
		return agent + " repairs " + h, "tend waits for " + agent + ", then has the repair reviewed"
	}
	return fmt.Sprintf("%s works on step %s at %s", agent, name, h), "tend waits for " + agent + ", then checks its result"
}

// Prompt a step's agent once: the head is added to the step's record just
// before the prompt goes, and taken off again only when the terminal refused
// the prompt, as a timeout or a stall may have delivered it.
func promptStep(name, rec string) {
	s := fl.Steps[name]
	agent, harness, root := agentOf(name)
	rr, _ := parseResult(fl.Steps[root].Result)
	if flowName == "build" && root == fl.Start {
		rr.commit = true // the implementer commits, whatever else continues it
	}
	startAgent(agent, harness, rr.commit)
	text, _ := promptText(promptRev, strings.ReplaceAll(s.Prompt, "{kind}", kind))
	text = fill(text)
	text1, next1 := doing(name, agent)
	stage("working", text1, next1)
	switch st := term.status(agent); st {
	case "working":
		say("%s is still working on an earlier prompt; waiting for it before prompting it", agent)
		await(agent)
	case "blocked", "unknown":
		stop("%s is %s; look at it, then run tend again", agent, st)
	}
	h := git("rev-parse", "HEAD")
	setAsks(rec, append(asks(rec), h))
	prompted++
	if strings.HasPrefix(agent, "impl-") {
		write(run+"/layout", layout()+"\n") // for contained()
	}
	st, why := term.prompt(agent, text)
	if st == "" && !strings.Contains(why, "timeout") && !strings.Contains(why, "stall") && why != "no answer" {
		hs := asks(rec)
		if i := slices.Index(hs, h); i >= 0 {
			setAsks(rec, slices.Delete(hs, i, i+1)) // refused before it went: not asked
		}
	}
	settled(agent, st, why)
}

// Reach a step at the branch's head: take its result if it is there, wait for
// its agent if it is still working on its prompt, stop for a person if it was
// asked and left nothing, go to spent at its limit, or prompt it. Its outcome.
func reach(name string, s *step) string {
	if flowName == "build" {
		contained() // on a rerun too, which finds the candidate built
	}
	r, _ := parseResult(s.Result)
	h := git("rev-parse", "HEAD")
	// A head carried across a clean merge of the target is judged by the
	// verdict of the head that passed review, and no agent is asked.
	if r.verdict != "" && !exists(run+"/"+verdictName(r.verdict, h)) {
		if c := carried(h); c != nil {
			h = c[len(c)-1]
		}
	}
	vars["head"] = h
	vars["findings"] = str(get(lastV, "summary"))
	rec := recordOf(name)
	if r.verdict != "" || (r.file != "" && !r.commit) {
		return reachAt(name, s, r, h, rec)
	}
	return reachInOrder(name, s, r, h, rec)
}

// A step whose result is a file named for the head is judged at the head.
func reachAt(name string, s *step, r result, h, rec string) string {
	f := run + "/" + fill(r.file+r.verdict)
	vars["result"] = f
	got := func() string {
		if r.file != "" {
			if nonempty(f) {
				return "ok"
			}
			return ""
		}
		if !valid(f, h) {
			return ""
		}
		x := parse(cat(f))
		lastV = x
		if !strings.HasPrefix(fileName(v), "verdict-") || strings.HasPrefix(r.verdict, "verdict-") {
			v = f // the newest, the reviewer's verdict-* first
		}
		switch {
		case get(x, "conclusion") == "success":
			return "success"
		case get(x, "actionable") == true:
			return "actionable"
		}
		return "failure"
	}
	agent, _, _ := agentOf(name)
	left := func() string { // the turn ended: the result, which leaves the branch as it was
		o := got()
		if o == "" {
			if exists(f) {
				stop("%s settled without a valid verdict for %s; look at pane %s", agent, short(h), term.pane(agent))
			}
			stop("%s wrote no %s; look at pane %s", name, fileName(f), term.pane(agent))
		}
		if _, _, root := agentOf(name); root != fl.Start && !continued(root) {
			term.closePane(term.pane(agent))
		}
		if flowName == "build" && (git("rev-parse", "HEAD") != h || !clean()) {
			stop("%s changed the checkout", agent)
		}
		return o
	}
	// An agent still on its turn is waited for before its file is judged,
	// which it may be writing.
	if slices.Contains(asks(rec), h) && term.status(agent) == "working" {
		say("%s is still working on an earlier prompt; waiting for it, not prompting again", agent)
		await(agent)
		return left()
	}
	if r.verdict != "" && exists(f) && !valid(f, h) {
		os.Rename(f, f+".rejected-"+strconv.FormatInt(time.Now().Unix(), 10))
		say("set aside an invalid verdict %s for %s", f, short(h))
		if hs := asks(rec); slices.Contains(hs, h) {
			setAsks(rec, slices.DeleteFunc(hs, func(x string) bool { return x == h })) // asked again
		}
	}
	if o := got(); o != "" {
		return o
	}
	if slices.Contains(asks(rec), h) {
		stop("%s was asked at %s and left nothing: no %s, and %s is not working on it. Look at its pane, then write it and run tend again, or delete %s to have it asked once more", name, short(h), fileName(f), agent, rec)
	}
	if s.Limit > 0 && timesAsked(name, rec) >= s.Limit {
		return "spent"
	}
	promptStep(name, rec)
	return left()
}

// The build's implementation step: implement, or the first step in a flow
// without one.
func implStep() string {
	if _, ok := fl.Steps["implement"]; ok {
		return "implement"
	}
	return fl.Start
}

func fileName(f string) string { return f[strings.LastIndex(f, "/")+1:] }

func continued(root string) bool {
	for _, s := range fl.Steps {
		if s.Continue == root {
			return true
		}
	}
	return false
}

// A step whose result is a commit, or a turn that ends, is judged by its asks
// in order: the k-th time a walk reaches it answers its k-th ask.
func reachInOrder(name string, s *step, r result, h, rec string) string {
	agent, _, _ := agentOf(name)
	f := ""
	if r.file != "" {
		f = run + "/" + fill(r.file)
		vars["result"] = f
	}
	got := func(a string) string {
		if !r.commit && r.file == "" || r.commit && git("rev-parse", "HEAD") != a || f != "" && nonempty(f) {
			return "ok"
		}
		return ""
	}
	isReq := reviewReq != nil && name == reqStep
	// The build's implementation is asked whenever Git shows the branch
	// unfinished, so an ask whose turn tend saw end leaves the record, which
	// then holds only one that may still be in flight (killed, timed out).
	implementing := flowName == "build" && name == implStep() && r.commit
	left := func(a string, seen bool) string {
		if implementing && seen {
			hs := asks(rec)
			done := run + "/" + name + "-done"
			setAsks(done, append(asks(done), hs[used[rec]-1]))
			setAsks(rec, slices.Delete(hs, used[rec]-1, used[rec]))
			used[rec]--
		}
		if implementing && !seen { // an ask that may still be in flight is not sent again
			if why := unbuilt(); why != "" {
				stop("%s. Look at %s's pane, then finish it and run tend again, or delete %s to have %s asked once more", why, agent, rec, name)
			}
		}
		o := got(a)
		if o == "" {
			switch {
			case isReq:
				stop("the changes %s requested on %s (review %s) were asked of %s, which committed nothing and is not working on them. Look at its pane, then commit and run tend again, or delete %s to have them asked once more", vars["who"], pr, vars["review_id"], agent, rec)
			case f != "":
				stop("the %s asked of %s at %s has no commit and no %s, and %s is not working on it. Look at its pane, then commit or write %s and run tend again, or delete %s to have it asked once more", name, agent, short(a), fileName(f), agent, f, rec)
			}
			stop("the %s asked of %s at %s has no commit and %s is not working on it. Look at its pane, then commit and run tend again, or delete %s to have it asked once more", name, agent, short(a), agent, rec)
		}
		if flowName == "build" {
			contained()
		}
		if isReq {
			if now := git("rev-parse", "HEAD"); now == a {
				answer(str(get(reviewReq, "node_id")), vars["who"], cat(f))
			} else if nonempty(f) {
				os.Rename(f, run+"/reply-"+now) // quoted in the new head's review
			}
		}
		return o
	}
	if hs := asks(rec); used[rec] < len(hs) {
		a := hs[used[rec]]
		used[rec]++
		if _, _, root := agentOf(name); root == name && root != fl.Start {
			agent = agentAt(root, a) // asked at a, whatever it committed since
		}
		seen := term.status(agent) == "working"
		if seen {
			text1, next1 := doing(name, agent)
			stage("working", text1, next1)
			say("%s is still working on an earlier prompt; waiting for it, not prompting again", agent)
			await(agent)
		}
		return left(a, seen)
	}
	if f != "" && nonempty(f) {
		return left(h, false)
	}
	// The implementation of a branch that already holds a finished candidate
	// is never asked, judged from Git.
	if implementing && unbuilt() == "" {
		return "ok"
	}
	if s.Limit > 0 && timesAsked(name, rec) >= s.Limit {
		return "spent"
	}
	promptStep(name, rec)
	used[rec]++
	return left(h, true)
}
