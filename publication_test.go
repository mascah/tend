package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
}
