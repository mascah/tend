package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func (s *sc) labelled(n int) bool {
	for _, i := range list(s.readJSON(s.st + "/issues")) {
		if str(get(i, "number")) == strconv.Itoa(n) {
			return slices.Contains(list(get(i, "labels")), any("tend"))
		}
	}
	return false
}

// A commit on origin's branch b, started from origin's branch from when b is
// not there yet, writing each path given; its sha.
func (s *sc) pushTo(b, from string, paths ...string) string {
	return s.bash(s.dir, `w=$(mktemp -d "$t/push.XXXX") && git clone -q "$t/repo.git" "$w" && cd "$w" &&
		{ git checkout -q "$1" 2>/dev/null || git checkout -q -b "$1" "origin/$2"; } && b=$1 && shift 2 &&
		for p in "$@"; do mkdir -p "$(dirname "$p")" && echo "$RANDOM" >>"$p"; done &&
		git add -A && git commit -qm push && git push -q origin "HEAD:refs/heads/$b" && git rev-parse HEAD`, append([]string{"_", b, from}, paths...)...)
}

// An open pull request from origin's branch b, opened by account by, as the
// gh stand-in keeps one.
func (s *sc) openPR(b, by string) {
	k := s.st + "/pr-" + strings.ReplaceAll(b, "/", "_")
	s.write(k, "A body\n")
	s.write(k+".by", by+"\n")
	s.write(k+".where", b+"\n"+s.git(s.repo, "remote", "get-url", "origin")+"\n")
}

// The tend/review statuses set on head h, as "<state> <account>" lines.
func (s *sc) statusesOn(h string) []string {
	var r []string
	for _, l := range strings.Split(s.cat(s.st+"/statuses"), "\n") {
		if a, ok := strings.CutPrefix(l, h+" "); ok {
			r = append(r, a)
		}
	}
	return r
}

var intakeScenarios = []scenario{
	{"propose", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend", "2:", "3:tend,bug")
		r := s.propose()
		s.check("tend propose exits 0", r.code == 0)
		s.check("each labelled issue is drafted once", s.drafts() == 2)
		s.check("unattended, from the issue's words", s.has(s.log(), `^`+regexp.QuoteMeta("agent prompt prop-1-"+s.cksum(s.repo)+" /grove-propose --unattended Request 1. Make thing 1 better. (GitHub issue https://example.test/issues/1)")))
		s.check("an unlabelled issue is left alone", !exists(s.prOf("2")))
		s.check("a pull request per draft, by the account", s.cat(s.prOf("1")+".by")+" "+s.cat(s.prOf("3")+".by") == "the-bot the-bot")
		s.check("from the issue's branch, onto the target", s.has(s.log(), `^gh pr create --base main --head tend/issue-1 `))
		s.check("titled and described by the drafter's commit", strings.Contains(s.log(), "--title Propose a thing --body-file"))
		s.check("described by its body, which refers to the issue", s.has(s.cat(s.prOf("1")), `^Refs #1$`))
		s.check("the branch holds the change", s.count(s.git(s.repo+".git", "ls-tree", "-r", "--name-only", "tend/issue-1"), `^grove/changes/261115-p0000`) == 1)
		rv := s.cat(s.st + "/reviews")
		s.check("a review lists the open questions", strings.Contains(rv, "Which thing?"))
		s.check("and says what merging with them open means", strings.Contains(rv, "its implementation may then choose those answers"))
		s.check("naming the change to revise", strings.Contains(rv, "/grove-propose 261115-p0000"))
		s.check("and that once merged it can be built by its ID, or planned with others", strings.Contains(rv, "can be built by its ID, `tend <repo> <change-id>`, or planned into a cycle with others") && !strings.Contains(rv, "waits for a cycle"))
		s.propose()
		s.check("a second pass drafts nothing again", s.drafts() == 2)
		s.check("and opens nothing again", s.creates() == 2)
		s.check("nor reviews again", len(s.reviewList()) == 2)
	}},
	{"propose-notify", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("43:tend")
		r := s.propose(`SUMMARY=1. [consequential] Which thing?\n2. [consequential] Which place?\n3. Which name?`)
		n := s.lastNote()
		s.check("a proposal pull request is notified as waiting for the person", r.code == 0 && len(s.notes()) == 1 && n.Title == "tend: o/r#1 waits for you")
		s.check("naming the request it came from and the consequential questions marked", strings.Contains(n.Body, "Proposal from #43: 2 open questions are consequential"))
		s.check("saying to merge to approve or revise it on its branch", strings.Contains(n.Body, "Merge to approve, or revise it on tend/issue-43"))
		s.check("ending with the pull request's URL", n.endsWith("https://example.test/pr/tend_issue-43"))
		s.propose()
		s.check("the next intake pass notifies it no more", len(s.notes()) == 1)
	}},
	{"propose-notify-unmarked", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		s.propose()
		n := s.lastNote()
		s.check("a summary with no question marked gives no count", !strings.Contains(n.Body, "consequential"))
		s.check("and says to see its open questions", strings.Contains(n.Body, "Proposal from #1: see its open questions."))
	}},
	{"propose-declined", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("DECLINE=1")
		s.check("a request that cannot be drafted exits 0", r.code == 0)
		s.check("opens no pull request", !exists(s.prOf("1")))
		s.check("says why on the issue, as the account", str(get(s.readJSON(s.st+"/comments"), 0, "user", "login")) == "the-bot")
		s.check("in the drafter's words", strings.Contains(s.cat(s.st+"/comments"), "It names no outcome."))
		s.check("and takes the label off", !s.labelled(1))
		s.check("notifying no one: its author acts on the comment", len(s.notes()) == 0)
		s.check("its worktree is removed, with its branch", !exists(s.st+"/wt-tend/issue-1") && s.git(s.repo, "branch", "--list", "tend/issue-1") == "")
		s.propose("DECLINE=1")
		s.check("so a second pass leaves it", s.drafts() == 1 && s.comments(1) == 1)
		s.issues("1:tend")
		s.propose()
		s.check("labelled again, it is drafted afresh", s.drafts() == 2 && exists(s.prOf("1")))
	}},
	{"propose-decline-lost", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("DECLINE=1", "EDIT_FAIL=1")
		s.check("a decline whose label removal fails stops", r.code == 3)
		r = s.propose("DECLINE=1")
		s.check("the rerun finishes it", r.code == 0 && !s.labelled(1))
		s.check("without commenting twice or drafting again", s.comments(1) == 1 && s.drafts() == 1)
	}},
	{"propose-cap", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend", "2:tend", "3:tend")
		r := s.propose("CAP=2")
		s.check("the cap holds open proposals", s.drafts() == 2 && !exists(s.prOf("3")))
		s.check("and says so", r.said("the cap; the rest wait"))
		s.touch(s.prOf("1") + ".merged")
		s.propose("CAP=2")
		s.check("a merged proposal frees a place", exists(s.prOf("3")))
	}},
	{"propose-follow-lost", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("FOLLOW_FAIL=agent_prompt_stalled")
		s.check("a stalled follow-up stops", r.code == 3)
		r = s.propose()
		s.check("the rerun stops too, not asking again", r.code == 3 && s.follows() == 1)
		s.check("and says what to do", r.said("delete .*/followed"))
		m, _ := filepath.Glob(s.ts + "/intake/*/followed")
		for _, f := range m {
			os.Remove(f)
		}
		r = s.propose()
		s.check("with the mark deleted, it is asked again and published", r.code == 0 && s.follows() == 2 && exists(s.prOf("1")))
		s.check("the draft was asked for once", s.drafts() == 1)
	}},
	{"propose-follow-gone", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("FOLLOW_FAIL=timeout")
		s.check("a follow-up that timed out stops", r.code == 3)
		s.gone("prop-")
		r = s.propose()
		s.check("with the drafter gone, the rerun asks the follow-up once more and publishes", r.code == 0 && s.follows() == 2 && exists(s.prOf("1")))
		s.check("of a new drafter session", s.count(s.log(), `^agent start prop-`) == 2)
		s.check("the draft was asked for once", s.drafts() == 1)
	}},
	{"propose-survivor", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("PROPOSE_FAIL=timeout")
		s.check("a drafting prompt that timed out stops", r.code == 3)
		s.agent("working", "prop")
		r = s.propose()
		s.check("the rerun waits for the drafter, then has it commit", r.code == 0 && exists(s.prOf("1")))
		s.check("without asking for the draft again", s.drafts() == 1)
	}},
	{"propose-unlisted", func(s *sc) {
		s.project("main", true)
		s.issues("1:tend")
		r := s.propose()
		s.check("a repository not listed for publication gives no requests", r.code == 3 && s.drafts() == 0)
		s.check("and tend says why", r.said("tend takes no requests from it"))
	}},
	{"drafter-terms", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		s.propose()
		start := strings.Join(slices.DeleteFunc(strings.Split(s.log(), "\n"), func(l string) bool { return !strings.HasPrefix(l, "agent start prop-") }), "\n")
		s.check("the drafter is denied pushing and gh pr", strings.Contains(start, "--disallowedTools Bash(git push:*) Bash(gh pr:*)"))
		s.check("and told tend publishes", strings.Contains(start, "Push nothing, open or edit no pull request"))
		s.check("but to commit only when tend asks", strings.Contains(start, "Commit only when tend asks you to."))
	}},
	{"fetch-propose", func(s *sc) {
		s.project("main", true)
		s.allow()
		theirs := s.pushed()
		s.issues("1:tend")
		s.propose()
		s.check("a proposal starts from origin's target, fetched", s.gitOK(s.st+"/wt-tend/issue-1", "merge-base", "--is-ancestor", theirs, "HEAD"))
		s.check("and is opened", exists(s.prOf("1")))
	}},
	{"at-once-intake", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		f, release := s.hold()
		k := s.bg([]string{"HOLD=" + f}, `^agent prompt impl-`, s.repo, "261114-k0000")
		r := s.propose()
		s.check("an intake pass runs beside a build", r.code == 0 && s.drafts() == 1 && exists(s.prOf("1")))
		release()
		s.check("and the build reaches its verdict", k.wait().passed())
	}},
	{"at-once-second-intake", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		f, release := s.hold()
		p := s.bg([]string{"HOLD=" + f}, `^agent prompt prop-`, "propose", s.repo)
		r := s.propose()
		s.check("a second intake pass for the repository is refused", r.code == 4)
		s.check("naming the first pass's pid", r.said(fmt.Sprintf("pid %d", p.c.Process.Pid)))
		release()
		s.check("the first finishes", p.wait().code == 0 && s.drafts() == 1)
	}},
	{"propose-same-number", func(s *sc) {
		// Two repositories with an issue #1 each, in one Herdr: the second is
		// drafted while the first's drafter is still open.
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		s.check("the first repository's issue is drafted", s.propose().code == 0 && s.drafts() == 1)
		o := s.other()
		o.project("main", true)
		o.allow()
		o.issues("1:tend")
		r := o.propose()
		a, b := "prop-1-"+s.cksum(s.repo), "prop-1-"+o.cksum(o.repo)
		s.check("each repository's drafter has its own name", a != b && s.has(s.log(), "^agent start "+a+" "))
		s.check("the second repository starts its own drafter", o.has(o.log(), "^agent start "+b+" "))
		s.check("and prompts it, not the first's", o.count(o.log(), "^agent prompt "+b+" /grove-propose") == 1 && !o.has(o.log(), "^agent prompt "+a+" "))
		s.check("its proposal is opened", r.code == 0 && exists(o.prOf("1")))
	}},
	{"propose-vouch", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		s.propose()
		h := s.git(s.repo+".git", "rev-parse", "tend/issue-1")
		s.check("a proposal just opened has tend/review success, set by the account", slices.Equal(s.statusesOn(h), []string{"success the-bot"}))
		s.check("described as Grove documents only", s.has(s.cat(s.st+"/descriptions"), `^`+h+` Grove documents only: merging approves this change$`))
		s.check("linking the pull request", s.has(s.log(), `^gh api repos/\{owner\}/\{repo\}/statuses/`+h+` .*target_url=https://example.test/pr/tend_issue-1$`))
		s.propose()
		s.check("a second pass with no new head sets no status again", len(s.statusesOn(h)) == 1)
		n := s.pushTo("tend/issue-1", "main", "grove/changes/261115-p0000-a-proposal/proposal.md")
		r := s.propose()
		s.check("a person's revision of the proposal gets its own success", r.code == 0 && slices.Equal(s.statusesOn(n), []string{"success the-bot"}))
		s.check("and the earlier head's is not set again", len(s.statusesOn(h)) == 1)
		c := s.pushTo("tend/issue-1", "main", "intake.go")
		s.propose()
		s.check("a head that also touches code gets no success", slices.Equal(s.statusesOn(c), []string{"failure the-bot"}))
		s.check("its description names the path", s.has(s.cat(s.st+"/descriptions"), `^`+c+` Changes intake.go, outside`))
	}},
	{"propose-vouch-resumed", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:tend")
		r := s.propose("STATUS_FAIL=1")
		drafted := s.git(s.repo+".git", "rev-parse", "tend/issue-1")
		s.check("a pass that could not set the status stops", r.code != 0 && len(s.statusesOn(drafted)) == 0)
		c := s.pushTo("tend/issue-1", "main", "intake.go")
		r = s.propose()
		s.check("resuming, the drafter's commit, no longer the head, gets no status", r.code == 0 && len(s.statusesOn(drafted)) == 0)
		s.check("the head a person pushed gets failure", slices.Equal(s.statusesOn(c), []string{"failure the-bot"}))
	}},
	{"propose-vouch-open", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.issues("1:", "2:tend")
		old := s.pushTo("tend/issue-1", "main", "grove/changes/261115-p0000-thing/proposal.md")
		s.openPR("tend/issue-1", "the-bot")
		two := s.pushTo("tend/issue-3", "main", "grove/changes/261115-p0000-thing/proposal.md", "grove/changes/261115-p0001-other/proposal.md")
		s.openPR("tend/issue-3", "the-bot")
		merged := s.pushTo("tend/issue-4", "main", "grove/changes/261115-p0002-done/proposal.md")
		s.openPR("tend/issue-4", "the-bot")
		s.touch(s.st + "/pr-tend_issue-4.merged")
		theirs := s.pushTo("tend/issue-5", "main", "grove/changes/261115-p0003-theirs/proposal.md")
		s.openPR("tend/issue-5", "owner")
		arch := s.pushTo("tend/jira-AB-1", "main", "grove/changes/archive/261115-p0004-old/proposal.md")
		s.openPR("tend/jira-AB-1", "the-bot")
		r := s.propose("CAP=3")
		s.check("with the cap reached, no request is drafted", r.code == 0 && s.drafts() == 0 && r.said("the cap"))
		s.check("but a proposal opened before has tend/review success, set by the account", slices.Equal(s.statusesOn(old), []string{"success the-bot"}))
		s.check("two changes in one head get no success", slices.Equal(s.statusesOn(two), []string{"failure the-bot"}))
		s.check("nor does an archived change", slices.Equal(s.statusesOn(arch), []string{"failure the-bot"}))
		s.check("a merged proposal gets no status", len(s.statusesOn(merged)) == 0)
		s.check("nor one a person opened from a tend/issue- branch", len(s.statusesOn(theirs)) == 0)
	}},
}
