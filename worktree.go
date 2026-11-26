package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

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
// goes, and a rerun compares with the same record. What tend's other attempts
// and intake passes made is theirs: the worktree each run or intake folder
// records and the branch it is on, once any tend setting up has recorded its
// own.
// ponytail: tend cannot tell whether an implementer wrote into another
// attempt's worktree; that attempt's own diff and review would show it.
func contained() {
	if !exists(run + "/layout") {
		return
	}
	settingUp() // until the layout is compared, so no tend makes a worktree meanwhile
	var dirs, branches []string
	fs, _ := filepath.Glob(state + "/runs/*/worktree")
	more, _ := filepath.Glob(state + "/intake/*/worktree")
	for _, f := range append(fs, more...) {
		if d := cat(f); f != run+"/worktree" && exists(d) {
			dirs = append(dirs, d)
			branches = append(branches, gitQ("-C", d, "branch", "--show-current"))
		}
	}
	theirs := func(l string) bool {
		if p, found := strings.CutPrefix(l, "worktree "); found {
			return slices.ContainsFunc(dirs, func(d string) bool { return sameDir(d, p) })
		}
		return slices.Contains(branches, strings.TrimPrefix(l, "branch "))
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
		if slices.Contains(was, l) || theirs(l) {
			continue
		}
		if p, found := strings.CutPrefix(l, "worktree "); found {
			why = "it added the worktree " + p
		} else if b := strings.TrimPrefix(l, "branch "); commits(base + ".." + b) {
			why = "it made the branch " + b + ", with commits"
		}
	}
	setUp()
	if why != "" {
		stop("%s left work outside %s: %s. tend reviews and publishes nothing; move the work onto %s and remove the rest, then run tend again", impl, branch, why, branch)
	}
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
// closed, removed when the pull request's head holds theirs, so nothing is lost;
// and those of attempts with no pull request whose cycle or change is done, when their
// branch has no commit the target lacks. Never one with a worktree inside it
// or whose implementer is working. A run folder names its worktree in
// worktree, deleted once it is gone.
func tidy() {
	g := sync.OnceValues(func() (any, error) { return groveIn(repo) }) // one read for every folder
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
			// The pull request's head holds every commit of the worktree's, as
			// when a person updated the branch on GitHub before merging.
			if hd := gitQ("-C", d, "rev-parse", "HEAD"); hd != h {
				quiet("git", "-C", repo, "fetch", "-q", "origin", h)
				if !ok("git", "-C", repo, "merge-base", "--is-ancestor", hd, h) {
					say("left %s: its head is not in its pull request's", d)
					continue
				}
			}
			b = pb
		} else if b = superseded(d, r, g); b == "" {
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
		on, w := repo, "intake" // an intake folder's worktree listed here is this repository's
		if filepath.Dir(r) == state+"/runs" {
			a := strings.Split(cat(r+"/attempt"), "\n")
			on, w = a[0], a[min(1, len(a)-1)]
		}
		if !(sameDir(on, repo) && w == cycle) && held(on, w) {
			say("left %s: a tend is on it", d)
			continue
		}
		removeWorktree(d, b)
		if !exists(d) {
			os.Remove(f)
		}
	}
}

// The branch of an attempt with no pull request whose Grove cycle or change is done
// and whose branch the local target already holds, as when its work reached
// the target some other way; or nothing.
func superseded(d, r string, grove func() (any, error)) string {
	a := strings.Split(cat(r+"/attempt"), "\n")
	if filepath.Dir(r) != state+"/runs" || len(a) < 2 || !sameDir(a[0], repo) || isItem(a[1]) {
		return ""
	}
	g, err := grove()
	if known, open := workState(g, a[1]); err != nil || !known || open {
		return ""
	}
	if !ok("git", "-C", d, "merge-base", "--is-ancestor", "HEAD", str(get(g, "target"))) {
		return ""
	}
	return gitQ("-C", d, "branch", "--show-current")
}
