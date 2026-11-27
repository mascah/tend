package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

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

func firstRun(proj any, key, val string) any {
	for _, r := range list(get(proj, "runs")) {
		if str(get(r, key)) == val {
			return r
		}
	}
	return nil
}

// Grove's sentence on why it would not claim work item w now, or "" when it
// would or does not say.
func unclaimable(proj any, w string) string {
	for _, x := range append(list(get(proj, "cycles")), list(get(proj, "changes"))...) {
		if str(get(x, "id")) == w && get(x, "claimable") == false {
			return str(get(x, "verdict"))
		}
	}
	return ""
}

// An issue's attempt is named after it and its repository: issue-7-<cksum>,
// jira-scrum-5-<cksum>.
func issueRun(c string) string {
	if isJira(c) {
		c = "jira-" + strings.ToLower(c)
	}
	return fmt.Sprintf("%s/runs/%s-%d", state, c, cksum([]byte(repo)))
}

// One tend for each work item of a repository, and one intake pass ("intake")
// for each repository. A lock is an flock on a descriptor no child inherits,
// so it goes when tend exits or is killed, whatever children (an agent's wait,
// a shell) outlive it. A lock just dropped is waited for briefly.
func takeLock() {
	f := lockFile(repo, cycle)
	os.MkdirAll(filepath.Dir(f), 0o755)
	l, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		stop("could not take the lock %s", f)
	}
	for i := 0; ; i++ {
		err = syscall.Flock(int(l.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK {
			stop("could not take the lock %s", f)
		}
		if i == 10 {
			say("another tend is on %s in %s: %s", cycle, repo, cat(f))
			os.Exit(4)
		}
		time.Sleep(200 * time.Millisecond)
	}
	lock = l
	write(f, fmt.Sprintf("pid %d, %s in %s\n", os.Getpid(), cycle, repo))
}

// The lock file of work item w of repository r.
func lockFile(r, w string) string {
	// One lock however the repository is reached: a symlink, a subfolder or a
	// linked worktree all share its common Git directory.
	if g, err := quiet("git", "-C", r, "rev-parse", "--path-format=absolute", "--git-common-dir"); err == nil {
		r = g
	}
	if p, err := filepath.EvalSymlinks(r); err == nil {
		r = p
	}
	return fmt.Sprintf("%s/locks/%d-%s", state, cksum([]byte(r)), strings.ReplaceAll(w, "/", "_"))
}

// Whether a tend other than this one holds the lock of work item w of
// repository r; this tend's own reads as held too. Reads only.
func held(r, w string) bool {
	f, err := os.Open(lockFile(r, w))
	if err != nil {
		return false
	}
	defer f.Close()
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil
}

// While one tend fetches, tidies, makes its worktree and claims, other tends
// on this TEND_STATE wait: settingUp until setUp. A tend records its worktree
// before it lets the next go.
func settingUp() {
	f, err := os.OpenFile(state+"/setup", os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		stop("could not take the setup lock in %s", state)
	}
	for err = syscall.EINTR; err == syscall.EINTR; {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	}
	if err != nil {
		stop("could not take the setup lock in %s", state)
	}
	setup = f
}

func setUp() {
	if setup != nil {
		setup.Close()
		setup = nil
	}
}

// ---- an attempt

// 1. The attempt, and what the work item says it must do. Two kinds:
//
// A Grove cycle, or an approved change no cycle lists, built as a cycle of that
// one change. The newest record for it, if its checkout still holds its
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
		was := gitQ("-C", repo, "rev-parse", target)
		base = fetchTarget()
		if gitQ("-C", repo, "rev-parse", target) != was {
			proj = project() // as the target now is, which the fetch moved
		}
		loadFlow("build", base)
		rec = itemRun(proj, cycle)
		dir = str(get(rec, "checkout"))
		if dir != "" && gitQ("-C", dir, "branch", "--show-current") == str(get(rec, "branch")) {
			term.open(dir)
			cd(dir)
			proj = project() // the attempt's root, not the caller's
			say("carrying on attempt %s in %s", str(get(rec, "id")), dir)
		} else {
			if why := unclaimable(proj, cycle); why != "" {
				stop("%s", why)
			}
			tmp := fmt.Sprintf("tend-%s-%d", time.Now().Format("150405"), os.Getpid())
			dir = term.worktree(tmp, base, "tend: "+cycle)
			cd(dir)
			if _, err := sh(nil, os.Stderr, false, "grove", "claim", cycle); err != nil {
				cd(repo)
				removeWorktree(dir, tmp) // nothing in it yet; a rerun makes its own
				stop("grove claim refused %s", cycle)
			}
			b := git("branch", "--show-current")
			if b != tmp {
				quiet("git", "branch", "-D", tmp) // Herdr needs a branch; Grove's claim makes its own
			}
			proj = project()
			rec = firstRun(proj, "branch", b)
			if rec == nil {
				stop("claimed %s, but Grove shows no record for %s", cycle, b)
			}
			say("claimed %s as attempt %s in %s", cycle, str(get(rec, "id")), dir)
		}
		branch, id = str(get(rec, "branch")), str(get(rec, "id"))
		root = str(get(proj, "root"))
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
	if fl.Steps["publish"] != nil && fl.Steps["publish"].Next["approved"] != "" {
		write(run+"/approves", "publish acts on an approval\n") // for tend watch
	} else {
		os.Remove(run + "/approves")
	}
	write(run+"/worktree", dir+"\n") // for tidy and contained
	setUp()

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
	settledDecision(fmt.Sprintf("%s is reviewed, %s", branch, conclusion))
	if landed != "" {
		term.notify("tend "+cycle+": landed", landed+" is merged", "done")
		stage("done", fmt.Sprintf("landed %s, %s at %s, verdict %s", landed, branch, short(git("rev-parse", "HEAD")), v), "nothing: it is merged")
		say("done: landed %s, verdict %s", landed, v)
	} else {
		term.notify("tend "+cycle+": "+conclusion, where+" at "+h+" is ready for you", "done")
		stage("done", fmt.Sprintf("%s for %s at %s, verdict %s", conclusion, branch, h, v), "a person reviews "+where+" and decides whether to merge it")
		say("done: %s, verdict %s", branch, v)
	}
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
