package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func (s *sc) build(env ...string) res { return s.tendIn(env, s.repo, "issue-7") }
func (s *sc) verdict() string         { return glob(s.ts + "/runs/*/verdict-*.json") }

// The worktree of the issue build's branch.
func (s *sc) bworktree() string { return s.worktreeOf("refs/heads/tend/build-7") }

var attemptScenarios = []scenario{
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
	{"local-resume", func(s *sc) {
		s.project("main", false)
		r := s.run("IMPL_FAIL=stub_failure")
		s.check("without origin, a failure after the claim stops", r.code == 3)
		r = s.run()
		s.check("the rerun carries on and exits 0", r.code == 0)
		s.check("one record for the cycle", s.records() == 1)
		s.check("without origin, nothing is published", s.count(s.log(), `^gh `) == 0)
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
	{"repair-gone", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=repairable", "REPAIR_FAIL=timeout")
		s.check("a repair prompt that timed out stops", r.code == 3)
		s.gone("impl-")
		r = s.run("REVIEW=repairable")
		s.check("with the implementer gone, the repair is asked once more, reviewed, and passes", r.passed() && s.reviews() == 2)
		s.check("of a new implementer session", s.prompts("An") == 2 && s.count(s.log(), `^agent start impl-`) == 2)
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
	{"fetch-target", func(s *sc) {
		s.project("main", true)
		theirs := s.pushed()
		r := s.run()
		s.check("an attempt starts from origin's target, fetched, though the local one is behind", s.gitOK(s.worktree(), "merge-base", "--is-ancestor", theirs, "HEAD"))
		s.check("and passes", r.passed())
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
}
