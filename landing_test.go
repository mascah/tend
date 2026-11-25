package main

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

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

// As lands, with more files on the target, a flow file among them taking
// the place of landFlow's.
func (s *sc) landsWith(supervised bool, ci string, files map[string]string) {
	s.project("main", true)
	s.allow()
	all := landFlow(s.t, supervised)
	for p, b := range files {
		all[p] = b
	}
	s.commit("Land", all)
	s.ci(ci)
}

// Whether tend held the head, naming exactly the protected paths given.
func (r res) held(paths string) bool {
	return r.code == 3 && r.said(`^tend: [^ ]+ at [0-9a-f]+ changes the protected paths? `+regexp.QuoteMeta(paths)+`, so tend lands it only once a person with write access approves that head on https://`)
}

// The approving reviews on the pull request.
func (s *sc) approvals() int {
	n := 0
	for _, r := range s.reviewList() {
		if get(r, "state") == "APPROVED" {
			n++
		}
	}
	return n
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
		b := s.bg([]string{"TIMEOUT_MS=4000"}, ``, s.repo, "261114-k0000")
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
	{"protect-flow-file", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BUILD_ALSO=echo "# ours" >>tend/flow.yaml && echo "package main" >main.go`)
		s.check("an autonomous flow holds a head that edits the flow file, naming it and not main.go", r.held("tend/flow.yaml"))
		s.check("the pull request stays open and nothing is merged", !s.merged() && s.merges() == 0 && s.creates() == 1)
		reviewed, prompts := s.head(), s.count(s.log(), `^agent prompt `)
		s.pushed() // the target moves, so a landing would merge it in
		s.reviewOn(reviewed, "stranger", "NONE", "APPROVED")
		r = s.run()
		s.check("a stranger's approval releases nothing: tend exits 3 naming the flow file", r.held("tend/flow.yaml"))
		s.check("merging nothing", !s.merged() && s.merges() == 0)
		s.check("pushing no merge commit to the branch", s.git(s.repo+".git", "rev-parse", s.branch()) == reviewed && s.head() == reviewed)
		s.check("and prompting no agent", s.count(s.log(), `^agent prompt `) == prompts)
		s.approve(reviewed)
		r = s.run()
		s.check("once a teammate who is a collaborator approves the head, the next run lands it", r.passed() && s.merged())
		s.check("after merging the target in", len(s.parents(s.mergedAt()[0])) == 3)
	}},
	{"protect-nested", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BUILD_ALSO=mkdir -p web && echo "Be brief." >web/CLAUDE.md`)
		s.check("a nested agent instruction file is held by tend's own list", r.held("web/CLAUDE.md"))
		s.check("and nothing is merged", !s.merged() && s.merges() == 0)
	}},
	{"protect-renamed", func(s *sc) {
		s.landsWith(false, "check success", map[string]string{".github/workflows/ci.yml": "on: pull_request\n"})
		r := s.run(`BUILD_ALSO=git mv .github/workflows/ci.yml ci.yml.bak`)
		s.check("a protected file moved away is held under its old name", r.held(".github/workflows/ci.yml"))
		s.check("and nothing is merged", !s.merged() && s.merges() == 0)
	}},
	{"protect-unprotected", func(s *sc) {
		s.landsWith(false, "check success", map[string]string{
			"tend/flow.yaml": landFlow(s.t, false)["tend/flow.yaml"] + "protect:\n  - schema.sql\n",
			"schema.sql":     "create table a (b int);\n",
		})
		r := s.run(`BUILD_ALSO=grep -v -e "^protect:" -e "^  - schema.sql" tend/flow.yaml >f && mv f tend/flow.yaml && echo "create table c (d int);" >>schema.sql`)
		s.check("a candidate that drops schema.sql from its own protect is held by the target's, naming both", r.held("schema.sql and tend/flow.yaml"))
		s.check("its own flow file holds no protect", !strings.Contains(s.git(s.worktree(), "show", "HEAD:tend/flow.yaml"), "protect"))
		s.check("and nothing is merged", !s.merged() && s.merges() == 0)
	}},
	{"protect-carried", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BUILD_ALSO=echo "Be brief." >>CLAUDE.md`)
		s.check("a head that edits CLAUDE.md is held", r.held("CLAUDE.md"))
		reviewed := s.head()
		s.approve(reviewed)
		theirs := s.pushed()
		r = s.run()
		m := s.head()
		s.check("once approved, the head carried across a clean merge of the target lands", r.passed() && s.merged())
		s.check("at the merge commit, approved only at the head it is carried from", slices.Equal(s.mergedAt(), []string{m}) && slices.Equal(s.parents(m), []string{m, reviewed, theirs}) && s.approvals() == 1)
	}},
	{"protect-run-folder", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BUILD_ALSO=echo "Be brief." >>CLAUDE.md`)
		s.check("a head that edits CLAUDE.md is held", r.held("CLAUDE.md"))
		reviewed := s.head()
		s.approve(reviewed)
		s.pushed()
		s.ci("check pending")
		r = s.run("TIMEOUT_MS=400")
		m1 := s.head()
		s.check("approved, the target is merged in, and tend stops waiting for its checks", r.code == 3 && r.said(`still unfinished`) && s.parents(m1)[1] == reviewed)
		// A verdict for the merge commit, as anything that writes the run
		// folder could leave there, and the target moved again.
		s.write(s.runDir()+"/verdict-"+m1+".json", `{"head": "`+m1+`", "conclusion": "success", "actionable": false, "summary": "Fine."}`)
		s.bash(s.dir, `git clone -q "$t/repo.git" "$t/third" && cd "$t/third" && echo again >c.txt && git add c.txt && git commit -qm again && git push -q origin main`)
		s.ci("check success")
		r = s.run()
		m2 := s.head()
		s.check("the approval of the reviewed head still counts through both merges of the target: it lands", r.passed() && s.merged() && slices.Equal(s.mergedAt(), []string{m2}) && s.parents(m2)[1] == m1)
	}},
	{"protect-submodule-hidden", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BUILD_ALSO=git config diff.ignoreSubmodules all && mkdir -p .claude/vendor && git update-index --add --cacheinfo "160000,$(git rev-parse HEAD),.claude/vendor"`)
		s.check("a gitlink added under .claude/ is held, though the repository's config ignores submodules", r.held(".claude/vendor"))
		s.check("the config was set where the worktree reads it", s.git(s.worktree(), "config", "diff.ignoreSubmodules") == "all")
		s.check("and nothing is merged", !s.merged() && s.merges() == 0)
	}},
	{"protect-target-own", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BEFORE_MERGE=git clone -q "$t/repo.git" "$t/other" && cd "$t/other" && echo "FROM scratch" >Dockerfile && git add Dockerfile && git commit -qm theirs && git push -q origin main`)
		s.check("a protected path the target's own commit touches is not the candidate's change: it lands", r.passed() && s.merged() && s.merges() == 2)
	}},
	{"protect-read-again", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BEFORE_MERGE=git clone -q "$t/repo.git" "$t/other" && cd "$t/other" && printf "protect:\n  - a.txt\n  - grove/\n" >>tend/flow.yaml && git commit -qam theirs && git push -q origin main`)
		s.check("a target that protects the candidate's paths before the merge holds it, read afresh", r.code == 3 && r.said(`changes the protected paths grove/`))
		s.check("GitHub was asked once, and nothing is merged", s.merges() == 1 && !s.merged())
	}},
	{"protect-read-during-checks", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`DURING_CHECKS=git clone -q "$t/repo.git" "$t/other" && cd "$t/other" && printf "protect:\n  - a.txt\n  - grove/\n" >>tend/flow.yaml && git commit -qam theirs && git push -q origin main`)
		s.check("a target that protects the candidate's paths while its checks are read holds it", r.code == 3 && r.said(`changes the protected paths grove/`))
		s.check("before GitHub is asked to merge", s.merges() == 0 && !s.merged())
		s.check("and before the target is merged in", s.git(s.repo+".git", "rev-parse", s.branch()) == s.head() && len(s.parents(s.head())) == 2)
	}},
	{"protect-fix", func(s *sc) {
		s.lands(false, "check failure")
		r := s.run(`FIX_ALSO=mkdir -p .github/workflows && echo "on: push" >.github/workflows/ci.yml`)
		s.check("the fix for a failed check is reviewed", s.prompts("tend could not land") == 1 && s.reviews() == 2 && s.reviewed(s.head()) == 1)
		s.check("and its head, which edits a workflow, is held", r.held(".github/workflows/ci.yml"))
		s.check("and nothing is merged", !s.merged() && s.merges() == 0)
	}},
	{"protect-resolved", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BUILD_ALSO=echo mine >a.txt && echo "# ours" >>grove.yaml`)
		s.check("a head that edits grove.yaml is held", r.held("grove.yaml"))
		reviewed := s.head()
		s.approve(reviewed)
		s.bash(s.dir, `git clone -q "$t/repo.git" "$t/other" && cd "$t/other" && echo theirs >a.txt && git commit -qam theirs && git push -q origin main`)
		r = s.run()
		s.check("approved, it meets a conflict the implementer resolves in a new head that is reviewed", s.prompts("tend could not land") == 1 && s.reviewed(s.head()) == 1 && s.head() != reviewed)
		s.check("which is held, its approval being of the earlier head", r.held("grove.yaml"))
		s.check("and nothing is merged", !s.merged() && s.merges() == 0)
		s.approve(s.head())
		r = s.run()
		s.check("until a person with write access approves the new head", r.passed() && s.merged())
	}},
	{"protect-supervised", func(s *sc) {
		s.lands(true, "check success")
		r := s.run(`BUILD_ALSO=echo "# ours" >>tend/flow.yaml`)
		s.check("a supervised flow publishes a head that edits the flow file, merging nothing", r.passed() && !s.merged())
		s.approve(s.head())
		r = s.run()
		s.check("and lands it once a teammate approves, as before", r.passed() && s.merged())
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
	t.Parallel()
	for _, x := range landScenarios {
		t.Run(x.name, func(t *testing.T) { x.fn(newSc(t)) })
	}
}
