package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

func (s *sc) watchrepo(env ...string) res {
	return s.tendIn(append(env, "WATCH_ONCE=1"), "watch", s.repo)
}

// A project whose target's flow file sets watch.at_once to n, listed for
// publication, with no requests.
func (s *sc) atOnce(n int) {
	s.project("main", true)
	s.allow()
	s.issues()
	s.commit("Keep some going", map[string]string{"tend/flow.yaml": fmt.Sprintf("watch:\n  at_once: %d\n", n)})
}

// The cycle marked unattended in its plan, in the checkout dir, pushed.
func (s *sc) unattend(dir string) {
	f := dir + "/grove/cycles/261114-k0000-one.md"
	s.write(f, regexp.MustCompile(`(?m)^planned: .*\n`).ReplaceAllString(s.cat(f)+"\n", "${0}unattended: true\n"))
	s.bash(dir, "git commit -qam unattend && git push -q origin main")
}

var watchScenarios = []scenario{
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
	{"watch-killed", func(s *sc) {
		s.project("main", false)
		bg := s.background("IMPL_SLEEP=30")
		w := s.watchrepo()
		s.check("watch runs nothing for an attempt a tend is on", s.count(w.err, `running tend`) == 0)
		killed(bg)
		st := s.cat(s.runDir() + "/status")
		s.write(s.runDir()+"/status", "x needs-you: something\nnext: a person\n")
		w = s.watchrepo()
		s.check("nor for one that needs a person", s.count(w.err, `running tend`) == 0)
		s.write(s.runDir()+"/status", st+"\n")
		w = s.watchrepo()
		s.check("after a kill, watch says it runs the attempt again", w.said(`tend stopped on attempt [^ ]* before it finished; running tend on 261114-k0000`))
		s.check("tend waits for the implementer, prompted once in all", s.prompts("/grove-implement") == 1)
		s.check("and the head is reviewed", s.reviews() == 1 && strings.Contains(s.status(), "done: success"))
		w = s.watchrepo()
		s.check("a finished attempt is not run again", s.count(w.err, `running tend`) == 0)
	}},
	{"watch-killed-agents-gone", func(s *sc) {
		s.project("main", false)
		killed(s.background("IMPL_SLEEP=30"))
		if pid, err := strconv.Atoi(s.cat(s.st + "/sleeper")); err == nil {
			syscall.Kill(pid, syscall.SIGKILL)
		}
		s.gone("impl-")
		w := s.watchrepo()
		s.check("with the agents gone too, watch runs the attempt again", w.said(`running tend on 261114-k0000`))
		s.check("the implementer is asked once more in a new session", s.prompts("/grove-implement") == 2 && s.count(s.log(), `^agent start impl-`) == 2)
		s.check("and the head is reviewed", s.reviews() == 1 && strings.Contains(s.status(), "done: success"))
	}},
	{"watch-killed-retried", func(s *sc) {
		s.project("main", false)
		killed(s.background("IMPL_SLEEP=30"))
		s.git(s.repo, "worktree", "add", "-q", s.dir+"/other", "-b", "other")
		s.sh(s.dir+"/other", "grove", "claim", "--retry", "261114-k0000")
		s.check("the cycle is claimed again from another checkout", s.records() == 2)
		w := s.watchrepo()
		s.check("watch runs nothing for the earlier attempt", s.count(w.err, `running tend`) == 0 && s.prompts("/grove-implement") == 1)
	}},
	{"watch-at-once-killed", func(s *sc) {
		s.atOnce(1)
		s.change("261016-c0001", "Another", "")
		killed(s.background("IMPL_SLEEP=30"))
		w := s.watchrepo()
		s.check("under a count, watch runs the attempt whose tend died", w.said(`running tend on 261114-k0000`) && s.reviews() == 1)
		s.check("within the count, before starting anything new", !w.said(`starting 261016-c0001`))
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
	{"watch-at-once-checks", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues()
		s.commit("Wait for CI", map[string]string{"tend/flow.yaml": "watch:\n  at_once: 1\n" + ciFlow(s.t)})
		s.change("261016-c0001", "Another", "")
		s.ci("test pending")
		a := s.bg([]string{"TIMEOUT_MS=8000"}, `^gh api repos/.*/check-runs`, s.repo, "261114-k0000")
		s.waitFor(func() bool { return strings.Contains(s.status(), "waiting for the checks on ") }, "status to say it waits for the checks")
		w := s.watchrepo()
		s.check("an attempt whose tend waits for the checks on its head holds the slot", s.count(w.err, `starting`) == 0 && s.changeRecords("261016-c0001") == 0)
		s.ci("test success")
		r := a.wait()
		s.check("its checks passed, its tend exits waiting for a person's approval", r.code == 0 && strings.Contains(s.status(), "a person reviews https://example.test/pr/"))
		w = s.watchrepo()
		s.check("which frees the slot for the next", w.said(`starting 261016-c0001`) && s.changeRecords("261016-c0001") == 1)
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
