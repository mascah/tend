package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Agents with no terminal: the claude and codex stand-ins take the turns
// turn.sh gives, as the Herdr stand-in's agents do.
func (s *sc) hrun(env ...string) res { return s.run(append(env, "TEND_TERMINAL=headless")...) }

func (s *sc) issue(n int) string { return s.cat(s.st + "/issue-" + strconv.Itoa(n)) }

var headlessScenarios = []scenario{
	{"headless", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.hrun()
		s.check("headless, a cycle is claimed, built, reviewed and published", r.passed() && s.creates() == 1)
		s.check("with no terminal asked anything", s.count(s.log(), `^(worktree|pane|agent get|agent wait)`) == 0)
		s.check("its worktree made by Git under the state folder", sameDir(filepath.Dir(s.worktreeOf("refs/heads/grove/")), s.ts+"/worktrees"))
		impl := "impl-" + strings.TrimPrefix(s.runDir(), s.ts+"/runs/")
		s.check("the prompt sent is kept beside the agent", strings.HasPrefix(s.cat(s.ts+"/agents/"+impl+"/1.prompt"), "/grove-implement --runner 261114-k0000"))
		s.check("the implementer's later turns resume its session", s.has(s.log(), `^agent start `+impl) && s.count(s.log(), `^agent start `+impl) == 1 && s.prompts("Branch ") == 1)
		rv := str(get(s.reviewList()[0], "body"))
		s.check("the review on the pull request says what was approved", strings.Contains(rv, "**What you approved:** [261015-c0000](o/r/blob/") && strings.Contains(rv, "grove/changes/261015-c0000-a-change/proposal.md)"))
		r = s.hrun()
		s.check("a rerun repeats nothing", r.passed() && s.reviews() == 1 && s.creates() == 1 && s.prompts("/grove-implement") == 1)
	}},
	{"headless-killed-mid-turn", func(s *sc) {
		s.project("main", true)
		s.allow()
		c := s.background("TEND_TERMINAL=headless", "IMPL_SLEEP=3")
		killed(c)
		s.check("killed mid-turn, the attempt reads as working", strings.Contains(s.status(), "tend stopped before this finished"))
		r := s.hrun()
		s.check("the rerun waits for the turn that outlived tend, then carries on to a pull request", r.passed() && s.creates() == 1)
		s.check("and says so", r.said("still working on an earlier prompt"))
		s.check("prompting it nothing again", s.prompts("/grove-implement") == 1)
		s.check("and claiming nothing again", s.records() == 1)
	}},
	{"headless-reboot", func(s *sc) {
		s.project("main", true)
		s.allow()
		c := s.background("TEND_TERMINAL=headless", "IMPL_SLEEP=30")
		killed(c)
		pid, _ := strconv.Atoi(s.cat(s.st + "/sleeper"))
		syscall.Kill(pid, syscall.SIGKILL) // the agent's turn goes too, as in a reboot
		r := s.hrun()
		s.check("with tend and its agent's turn both gone, the rerun asks it once more and carries on", r.passed() && s.creates() == 1)
		s.check("once more only", s.prompts("/grove-implement") == 2)
		s.check("on the same claim", s.records() == 1)
	}},
	{"headless-blip", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.hrun("GH_FAIL=pr list --head", "TEND_RETRY_MS=100")
		s.check("a failed read of GitHub is tried again, not a person's", r.passed() && r.said("could not look for grove/.*'s pull request; trying again in 100ms"))
		s.check("repeating nothing done before it", s.creates() == 1 && s.reviews() == 1 && s.prompts("/grove-implement") == 1)
	}},
	{"decision", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.hrun("SILENT=impl")
		s.check("an implementer that leaves nothing is asked once more, then tend stops", r.code == 3 && s.prompts("/grove-implement") == 2)
		q := s.issue(100)
		s.check("asking the owner on GitHub, as the publishing account", s.cat(s.st+"/issue-100.by") == "the-bot" && strings.HasPrefix(s.cat(s.st+"/issue-100.title"), "tend needs you on 261114-k0000: the implement asked of impl-"))
		s.check("with what the agent said last", strings.Contains(q, "the proposal names two files for one table"))
		s.check("and how to answer", strings.Contains(q, "as **completed** for tend to carry on, asking impl-") && strings.Contains(q, "as **not planned**"))
		s.check("its status names the issue", strings.Contains(s.status(), "a person answers https://github.com/o/r/issues/100"))
		r = s.hrun("SILENT=impl")
		s.check("a rerun by hand that stops the same way asks nothing new while the issue is open", r.code == 3 && s.prompts("/grove-implement") == 3 && !exists(s.st+"/issue-101.title"))
		s.watch("TEND_TERMINAL=headless")
		s.check("tend watch leaves it while it is open", s.prompts("/grove-implement") == 3)
		s.write(s.st+"/issue-100.state", "closed completed the-bot\n")
		w := s.watch("TEND_TERMINAL=headless")
		s.check("closed by the account the agents share, it is no answer", s.prompts("/grove-implement") == 3 && w.said("takes no answer from it"))
		s.write(s.st+"/issue-100.state", "closed completed owner\n")
		w = s.watch("TEND_TERMINAL=headless")
		s.check("closed as completed by the owner, tend carries on to a pull request", w.said("owner answered .*issues/100: carry on") && s.creates() == 1 && s.prompts("/grove-implement") == 4)
		s.check("and the question is settled", !exists(s.runDir()+"/decision") && strings.Contains(s.status(), "done:"))
	}},
	{"decision-public", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.hrun("SILENT=impl", "VISIBILITY=PUBLIC")
		q := s.issue(100)
		s.check("on a public repository the issue asks without the agent's words", !strings.Contains(q, "the proposal names two files for one table") && strings.Contains(q, "is not posted to this public repository"))
		s.check("and names no path or machine of the runner's", !strings.Contains(q, s.ts) && !strings.Contains(q, s.dir) && !strings.Contains(q, hostname()))
	}},
	{"decision-left", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.hrun("SILENT=impl")
		s.write(s.st+"/issue-100.state", "closed not_planned owner\n")
		s.watch("TEND_TERMINAL=headless")
		s.check("closed as not planned, tend leaves the attempt", s.prompts("/grove-implement") == 2 && strings.Contains(s.status(), "as not planned, so tend leaves the attempt"))
		s.watch("TEND_TERMINAL=headless")
		s.check("and does not come back to it", s.prompts("/grove-implement") == 2)
	}},
	{"workflow-scope", func(s *sc) {
		s.project("main", true)
		s.allow()
		os.WriteFile(s.st+"/scopes", []byte("repo, workflow\n"), 0o644)
		r := s.hrun()
		s.check("a publishing token with the workflow scope publishes nothing", r.code == 3 && r.said("the-bot's token has the workflow scope") && s.creates() == 0)
	}},
}

func hostname() string { h, _ := os.Hostname(); return h }
