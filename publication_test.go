package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var publicationScenarios = []scenario{
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
		s.check("a verdict without notes has no notes heading", !strings.Contains(rv, "do not fail the head"))
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
	{"publish-notes", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run("REVIEW=notes")
		s.check("a passing head with notes publishes", r.passed())
		s.check("with status success", s.has(s.cat(s.st+"/statuses"), `^`+s.head()+` success the-bot$`))
		s.check("and one review", s.reviewed(s.head()) == 1)
		b := ""
		for _, x := range s.reviewList() {
			b = str(get(x, "body"))
		}
		s.check("showing the note under a heading saying it does not fail the head",
			strings.Contains(b, "**Notes that do not fail the head**, for a person to file or drop:\n\na.txt:1 could read better."))
	}},
	{"publish-failure-notes", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run("REVIEW=repairable-notes")
		s.check("a repaired head with notes on the failed one passes", r.passed())
		b := ""
		for _, x := range s.reviewList() {
			if strings.Contains(str(get(x, "body")), ": failure") {
				b = str(get(x, "body"))
			}
		}
		sum, note, rep := strings.Index(b, "a.txt:1 lacks the fix."), strings.Index(b, "a.txt:2 could read better."), strings.Index(b, "tend repairs once")
		s.check("the failed head's review shows its notes after its summary", sum >= 0 && note > sum)
		s.check("and still says the failure was judged repairable", rep > note)
		st := s.cat(s.st + "/statuses")
		s.check("the notes change neither status", s.count(st, ` failure the-bot$`) == 1 && s.count(st, `^`+s.head()+` success`) == 1)
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
	{"request-gone", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.request()
		r := s.run("REQUEST_FAIL=timeout")
		s.check("a request prompt that timed out stops", r.code == 3)
		s.gone("impl-")
		r = s.run()
		s.check("with the implementer gone, the rerun asks the request once more and passes", r.passed())
		s.check("of a new implementer session", s.prompts("teammate") == 2 && s.count(s.log(), `^agent start impl-`) == 2)
		s.check("and the new head is reviewed and published", s.reviews() == 2 && s.has(s.cat(s.st+"/statuses"), `^`+s.head()+` success the-bot$`))
		r = s.run()
		s.check("a later run asks nothing again", r.code == 0 && s.prompts("teammate") == 2)
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
	{"ci-red", func(s *sc) {
		s.waitsForCI("test failure")
		r := s.run(fixesCI)
		s.check("a failed check sends the head to fix, naming the check and its link", s.prompts("Checks GitHub ran on") == 1 && s.has(s.log(), `^agent prompt impl-[^ ]* Checks GitHub ran on .*test https://example\.test/checks/test`))
		s.check("the head fix makes is reviewed and published", s.reviews() == 2 && s.reviewed(s.head()) == 1 && s.git(s.repo+".git", "rev-parse", s.branch()) == s.head())
		s.check("its checks are waited for", s.has(s.log(), `^gh api repos/.*/commits/`+s.head()+`/check-runs`))
		s.check("and once they pass tend exits 0 with the pull request open", r.passed() && s.creates() == 1 && !s.merged() && s.merges() == 0)
		s.check("saying nothing on it of a spent fix", s.ciComments() == 0)
	}},
	{"ci-pass", func(s *sc) {
		s.waitsForCI("test success")
		r := s.run()
		s.check("with every check passed no fix is prompted", s.prompts("Checks GitHub ran on") == 0)
		s.check("and tend exits 0 with the pull request open", r.passed() && s.creates() == 1 && !s.merged())
		s.check("having read the head's checks", s.has(s.log(), `^gh api repos/.*/commits/`+s.head()+`/check-runs`))
	}},
	{"ci-failed-review", func(s *sc) {
		s.waitsForCI("test success")
		r := s.run("REVIEW=decision")
		s.check("the head's tend/review status is failure", s.has(s.cat(s.st+"/statuses"), `^`+s.head()+` failure the-bot$`))
		s.check("a failed review is not a failed check: no fix is prompted", s.prompts("Checks GitHub ran on") == 0)
		s.check("and tend exits 0 with the failed verdict", r.code == 0 && r.conclusion() == "failure")
	}},
	{"ci-requested-first", func(s *sc) {
		s.waitsForCI("test success")
		s.run()
		first := s.head()
		reads := s.count(s.log(), `^gh api repos/.*/commits/`+first+`/check-runs`)
		s.ci("test pending")
		s.request()
		s.run("TIMEOUT_MS=400")
		s.check("a change requested on a head whose checks run is asked of the implementer", s.prompts("teammate") == 1 && s.head() != first)
		s.check("without waiting for that head's checks", s.count(s.log(), `^gh api repos/.*/commits/`+first+`/check-runs`) == reads)
	}},
	{"ci-approved-red", func(s *sc) {
		s.waitsForCI("test success")
		s.run()
		first := s.head()
		s.ci("test failure")
		s.approve(first)
		r := s.run(fixesCI)
		s.check("an approved head whose check fails goes to fix", s.prompts("Checks GitHub ran on") == 1 && s.head() != first)
		s.check("nothing is merged", !s.merged() && s.merges() == 0)
		s.check("and the new head, its checks passed, waits for a new approval", r.passed() && s.reviewed(s.head()) == 1)
	}},
	{"ci-none", func(s *sc) {
		s.waitsForCI("")
		r := s.run("TIMEOUT_MS=400")
		s.check("with no check but tend/review, tend stops for a person saying so", r.code == 3 && r.said(`has no check but tend/review, so tend asks no one to approve it`))
		s.check("and prompts no fix", s.prompts("Checks GitHub ran on") == 0)
	}},
	{"ci-unfinished", func(s *sc) {
		s.waitsForCI("test pending")
		r := s.run("TIMEOUT_MS=400")
		s.check("with a check unfinished when the wait ends, tend stops for a person naming it", r.code == 3 && r.said(`the checks test on .* were still unfinished after 400ms`))
	}},
	{"ci-mixed", func(s *sc) {
		s.waitsForCI("test failure\nslow pending")
		r := s.run("TIMEOUT_MS=400")
		s.check("with one check failed and another unfinished, no fix is prompted before every check finishes", s.prompts("Checks GitHub ran on") == 0)
		s.check("and when the wait ends tend stops for a person naming the unfinished check", r.code == 3 && r.said(`the checks slow on .* were still unfinished after 400ms`))
	}},
	{"ci-past-deadline", func(s *sc) {
		s.waitsForCI("test success")
		r := s.run("TIMEOUT_MS=800", "CHECKS_SLEEP=2")
		s.check("a pass read only after the wait ended is not taken: tend stops for a person", r.code == 3 && r.said(`the checks on .* were not read finished within 800ms`))
		s.check("and the head is not reported done", r.conclusion() == "")
	}},
	{"ci-wait-bound", func(s *sc) {
		s.waitsForCI("test pending")
		r := s.run("TIMEOUT_MS=4000", "CHECKS_SLEEP=0.1")
		var ts []float64
		for _, l := range strings.Fields(s.cat(s.st + "/check-reads")) {
			f, _ := strconv.ParseFloat(l, 64)
			ts = append(ts, f)
		}
		s.check("with a check unfinished, tend stops for a person", r.code == 3 && r.said(`were still unfinished after 4000ms`))
		s.check("its last read starting no later than the wait's end, the sleep before it no longer than the time left", len(ts) > 1 && ts[len(ts)-1]-ts[0] < 4.35)
	}},
	{"ci-fix-no-commit", func(s *sc) {
		s.waitsForCI("test failure")
		r := s.run(`FIX_ALSO=git checkout -q a.txt; false`)
		s.check("a fix that settles with no commit stops for a person", r.code == 3 && s.prompts("Checks GitHub ran on") == 1)
		r = s.run()
		s.check("run again, the fix is spent: it is not asked again", s.prompts("Checks GitHub ran on") == 1)
		s.check("the pull request has one comment saying the checks failed after it", s.ciComments() == 1 && s.has(s.ciComment(), `^\*\*Checks failed on `+short(s.head())+` after the one fix`))
		s.check("and tend exits 0 waiting for a person", r.code == 0 && !s.merged())
	}},
	{"ci-no-flow", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.ci("test failure")
		r := s.run()
		s.check("without a flow file tend reads no check", s.count(s.log(), `/check-runs`) == 0)
		s.check("and exits 0 as before", r.passed() && s.prompts("Checks GitHub ran on") == 0)
	}},
	{"ci-still-red", func(s *sc) {
		first := s.stillRed()
		s.check("fix is prompted once", s.prompts("Checks GitHub ran on") == 1 && s.head() != first)
		s.check("the pull request has one comment naming the failed check on the second head", s.ciComments() == 1 && s.has(s.ciComment(), `^\*\*Checks failed on `+short(s.head())+` after the one fix`) && s.has(s.ciComment(), `^- test https://example\.test/checks/test$`))
	}},
	{"ci-one-fix", func(s *sc) {
		s.stillRed()
		s.approve(s.head())
		r := s.run()
		s.check("land finding the check still failed prompts no second fix", s.prompts("Checks GitHub ran on") == 1)
		s.check("merges nothing", !s.merged() && s.merges() == 0)
		s.check("and stops saying fix was asked once, its limit", r.code == 3 && r.said(`fix was asked 1 times, its limit`))
	}},
	{"ci-rerun", func(s *sc) {
		s.stillRed()
		at := s.head()
		r := s.run()
		s.check("a rerun with no new commit publishes the same head", r.code == 0 && s.head() == at)
		s.check("and the pull request still has one such comment", s.ciComments() == 1)
		os.Remove(s.runDir() + "/checks-said-" + at)
		s.run()
		s.check("found on GitHub when the run folder's mark is gone", s.ciComments() == 1)
	}},
}

// The supervised flow with land that also waits for checks at publish,
// sending a failure there to the same fix land uses, and the checks every
// commit gets.
func (s *sc) waitsForCI(ci string) {
	s.project("main", true)
	s.allow()
	s.commit("Wait for CI", map[string]string{"tend/flow.yaml": ciFlow(s.t)})
	s.ci(ci)
}

func ciFlow(t *testing.T) string {
	return strings.Replace(landFlow(t, true)["tend/flow.yaml"], "        approved: land\n", "        approved: land\n        failed: fix\n", 1)
}

// A fix that makes every check pass.
const fixesCI = `FIX_ALSO=echo "test success" >"$HERDR_STATE/ci"`

// A check that fails on the head and on the head fix makes, run once; the
// first head.
func (s *sc) stillRed() string {
	s.waitsForCI("test failure")
	r := s.run()
	s.check("still red after the fix, tend exits 0 waiting for a person", r.code == 0 && !s.merged())
	return s.parents(s.head())[1]
}

// tend's comments on the pull request that say a fix was spent, and the last.
func (s *sc) ciComments() int { return len(s.ciCommentList()) }
func (s *sc) ciComment() string {
	l := s.ciCommentList()
	if len(l) == 0 {
		return ""
	}
	return l[len(l)-1]
}
func (s *sc) ciCommentList() []string {
	var bs []string
	for _, x := range list(s.readJSON(s.st + "/comments")) {
		if b := str(get(x, "body")); str(get(x, "issue")) == filepath.Base(s.pr())[len("pr-"):] && strings.Contains(b, "<!-- tend:checks-failed ") {
			bs = append(bs, b)
		}
	}
	return bs
}
