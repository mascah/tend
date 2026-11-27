package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

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
					unfinished(why)
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
					unfinished(why)
				}
			}
		}
		at = to
	}
}

// A candidate the implementer left unfinished: a rerun asks it to finish.
func unfinished(why string) {
	stuck(impl, "", why, "Run tend again to have "+impl+" asked to finish it")
}

// Where a person sees what an agent did: its pane, or its output headless.
func lookAt(agent string) string {
	if _, ok := term.(*headless); ok {
		return "its output in " + agentDir(agent)
	}
	return "its pane"
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

// Prompt a step's agent once: head h is added to the record rec just before
// the prompt goes, and taken off again only when the terminal refused the
// prompt, as a timeout or a stall may have delivered it.
func promptStep(name, rec, h string) {
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
	setAsks(rec, append(asks(rec), h))
	prompted++
	if strings.HasPrefix(agent, "impl-") {
		write(run+"/layout", layout()+"\n") // for contained()
	}
	st, why := term.prompt(agent, text)
	if st == "" && !strings.Contains(why, "timeout") && !strings.Contains(why, "stall") && !strings.Contains(why, "after_delivery") && why != "no answer" {
		hs := asks(rec)
		if i := slices.Index(hs, h); i >= 0 {
			setAsks(rec, slices.Delete(hs, i, i+1)) // refused before it went: not asked
		}
	}
	settled(agent, st, why)
}

// The record of the asks made once more, by the head of the ask repeated: a
// name the glob of request-* does not take, so none counts toward a limit.
func againOf(rec string) string { return filepath.Dir(rec) + "/again-" + fileName(rec) }

// Ask a step once more, in a new session, when its ask at head a left nothing
// and its agent is gone, so that nothing is left to act on the earlier prompt;
// never a third time. A terminal that cannot be asked says nothing is gone,
// so a person decides. Whether it was asked.
func again(name, agent, rec, a string) bool {
	if !term.gone(agent) || slices.Contains(asks(againOf(rec)), a) {
		return false
	}
	say("%s was asked at %s and left nothing, and %s is no longer on it; asking it once more", name, short(a), agent)
	promptStep(name, againOf(rec), a)
	return true
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
	vars["earlier"] = earlier(h)
	rec := recordOf(name)
	if r.verdict != "" || (r.file != "" && !r.commit) {
		return reachAt(name, s, r, h, rec)
	}
	return reachInOrder(name, s, r, h, rec)
}

// The attempt's valid verdicts for commits of the branch before head h,
// oldest first, each its short head, conclusion and file, or none: what a
// later review is given as {earlier}. A verdict file is one a step of the
// flow names.
func earlier(h string) string {
	var fs []string
	for _, s := range fl.Steps {
		if r, _ := parseResult(s.Result); r.verdict != "" {
			m, _ := filepath.Glob(run + "/" + verdictName(r.verdict, "*"))
			fs = append(fs, m...)
		}
	}
	slices.Sort(fs)
	fs = slices.Compact(fs)
	mtime := func(f string) time.Time { s, _ := os.Stat(f); return s.ModTime() }
	slices.SortStableFunc(fs, func(a, b string) int { return mtime(a).Compare(mtime(b)) })
	var said []string
	for _, f := range fs {
		v := str(get(parse(cat(f)), "head"))
		if v == h || !valid(f, v) || !ok("git", "merge-base", "--is-ancestor", v, h) {
			continue
		}
		said = append(said, fmt.Sprintf("%s %s %s", short(v), str(get(parse(cat(f)), "conclusion")), f))
	}
	if len(said) == 0 {
		return "none"
	}
	return strings.Join(said, "; ")
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
		if o == "" && !exists(f) && again(name, agent, rec, h) { // headless, its turn over
			o = got()
		}
		if o == "" {
			if exists(f) {
				stuck(agent, rec, fmt.Sprintf("%s settled without a valid verdict for %s", agent, short(h)), "Look at "+lookAt(agent)+", then run tend again")
			}
			stuck(agent, rec, fmt.Sprintf("%s wrote no %s", name, fileName(f)), "Look at "+lookAt(agent)+", then write it and run tend again")
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
		if again(name, agent, rec, h) {
			return left()
		}
		asked := "asked"
		if slices.Contains(asks(againOf(rec)), h) {
			asked = "asked twice"
		}
		stuck(agent, rec, fmt.Sprintf("%s was %s at %s and left nothing: no %s, and %s is not working on it", name, asked, short(h), fileName(f), agent),
			fmt.Sprintf("Look at %s, then write it and run tend again, or delete %s to have it asked once more", lookAt(agent), rec))
	}
	if s.Limit > 0 && timesAsked(name, rec) >= s.Limit {
		return "spent"
	}
	promptStep(name, rec, git("rev-parse", "HEAD"))
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
				stuck(agent, rec, why, fmt.Sprintf("Look at %s, then finish it and run tend again, or delete %s to have %s asked once more", lookAt(agent), rec, name))
			}
		}
		o := got(a)
		if o == "" && seen && again(name, agent, rec, a) { // headless, its turn over
			o = got(a)
		}
		if o == "" {
			switch {
			case isReq:
				stuck(agent, rec, fmt.Sprintf("the changes %s requested on %s (review %s) were asked of %s, which committed nothing and is not working on them", vars["who"], pr, vars["review_id"], agent),
					fmt.Sprintf("Look at %s, then commit and run tend again, or delete %s to have them asked once more", lookAt(agent), rec))
			case f != "":
				stuck(agent, rec, fmt.Sprintf("the %s asked of %s at %s has no commit and no %s, and %s is not working on it", name, agent, short(a), fileName(f), agent),
					fmt.Sprintf("Look at %s, then commit or write %s and run tend again, or delete %s to have it asked once more", lookAt(agent), f, rec))
			}
			stuck(agent, rec, fmt.Sprintf("the %s asked of %s at %s has no commit and %s is not working on it", name, agent, short(a), agent),
				fmt.Sprintf("Look at %s, then commit and run tend again, or delete %s to have it asked once more", lookAt(agent), rec))
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
		} else if got(a) == "" || implementing && unbuilt() != "" {
			seen = again(name, agent, rec, a)
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
	promptStep(name, rec, h)
	used[rec]++
	return left(h, true)
}
