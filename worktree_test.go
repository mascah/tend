package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

var worktreeScenarios = []scenario{
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
}
