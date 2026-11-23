package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (s *sc) jbuild(env ...string) res { return s.tendIn(env, s.repo, "SCRUM-7") }

// ---- Jira, through the curl stand-in

// Map SCRUM to the repository's origin.
func (s *sc) jmap() {
	os.MkdirAll(s.ts, 0o755)
	f, _ := os.OpenFile(s.ts+"/jira", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	fmt.Fprintf(f, "https://jira.test SCRUM %s\n", s.git(s.repo, "remote", "get-url", "origin"))
	f.Close()
}

func adf(text string) any {
	return map[string]any{"type": "doc", "version": 1, "content": []any{map[string]any{"type": "paragraph", "content": []any{map[string]any{"type": "text", "text": text}}}}}
}

// An issue: key, labels, status category, a description in ADF ("" for a
// plain one) and how many comments it has.
func (s *sc) jissue(key, labels, status, desc string, comments int) {
	js := list(s.readJSON(s.st + "/jira"))
	var d any = adf("Make " + key + " better.")
	if desc != "" {
		d = parse(desc)
	}
	ls := []any{}
	for _, x := range strings.Split(labels, ",") {
		if x != "" {
			ls = append(ls, x)
		}
	}
	cs := []any{}
	for i := range comments {
		cs = append(cs, map[string]any{"body": adf(fmt.Sprintf("Earlier comment %d", i))})
	}
	js = append(js, map[string]any{"key": key, "comments": cs, "fields": map[string]any{"summary": "Request " + key, "description": d,
		"labels": ls, "status": map[string]any{"statusCategory": map[string]any{"key": status}}}})
	s.writeJSON(s.st+"/jira", js)
}

// The text of an issue's comments, one paragraph a line.
func (s *sc) jcomments(key string) string {
	var ls []string
	for _, x := range list(s.readJSON(s.st + "/jira")) {
		if str(get(x, "key")) != key {
			continue
		}
		for _, c := range list(get(x, "comments")) {
			for _, p := range list(get(c, "body", "content")) {
				for _, t := range list(get(p, "content")) {
					if v, ok := get(t, "text").(string); ok {
						ls = append(ls, v)
					}
				}
			}
		}
	}
	return strings.Join(ls, "\n")
}

const rich = `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Implement "},{"type":"text","text":"the spec","marks":[{"type":"link","attrs":{"href":"https://docs.test/spec"}}]},{"type":"text","text":" and "},{"type":"inlineCard","attrs":{"url":"https://jira.test/browse/SCRUM-9"}}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"keep totals"}]}]}]},{"type":"codeBlock","content":[{"type":"text","text":"ledger list --json"}]}]}`

var jiraScenarios = []scenario{
	{"jira-propose", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.jmap()
		s.issues()
		s.jissue("SCRUM-1", "tend", "new", rich, 3)
		s.jissue("SCRUM-2", "tend", "done", "", 0)
		s.jissue("SCRUM-3", "", "new", "", 0)
		s.jissue("SCRUM-4", "tend,bug", "indeterminate", "", 0)
		s.jissue("SCRUM-5", "tend", "new", "", 0)
		r := s.propose()
		s.check("tend propose takes Jira requests and exits 0", r.code == 0)
		s.check("each labelled, unfinished one is drafted once, past the first page", s.count(s.log(), `^agent prompt prop-scrum-[0-9]* /grove-propose`) == 3)
		s.check("the drafter gets the description with its links, list and code", s.has(s.log(), `^`+regexp.QuoteMeta("agent prompt prop-scrum-1 /grove-propose --unattended Request SCRUM-1. Implement the spec (https://docs.test/spec) and https://jira.test/browse/SCRUM-9 - keep totals ledger list --json (Jira issue https://jira.test/browse/SCRUM-1)")))
		s.check("a done or unlabelled request is left alone", !exists(s.st+"/pr-tend_jira-SCRUM-2") && !exists(s.st+"/pr-tend_jira-SCRUM-3"))
		s.check("a pull request from its branch, by the account", s.cat(s.st+"/pr-tend_jira-SCRUM-1.by") == "the-bot")
		s.check("the drafter is asked to end the commit with the key", s.has(s.log(), `^agent prompt prop-scrum-1 .*ends with the line 'Refs SCRUM-1'`))
		s.check("the review says where it came from", strings.Contains(s.cat(s.st+"/reviews"), "Proposal drafted from SCRUM-1 "))
		s.check("the Jira issue is told where its proposal is", strings.Contains(s.jcomments("SCRUM-1"), "tend opened https://example.test/pr/tend_jira-SCRUM-1 "))
		s.check("the handoff is recorded finished", nonempty(s.ts+"/intake/o_r-jira-SCRUM-1/linked"))
		s.check("the token is never in curl's arguments", !strings.Contains(s.log(), "tok-secret-1"))
		s.propose()
		s.check("a second pass drafts and opens nothing again", s.count(s.log(), `^agent prompt prop-scrum-[0-9]* /grove-propose`) == 3 && s.creates() == 3)
		s.check("nor comments again, though its comment is on a later page", s.count(s.jcomments("SCRUM-1"), `tend opened`) == 1)
		s.jissue("SCRUM-6", "tend", "new", "", 0)
		o := s.tendIn(nil, "status", s.repo).out
		waiting := o[max(0, strings.Index(o, "Requests waiting")):]
		s.check("the overview lists a Jira request waiting, not the handled ones", s.count(waiting, `SCRUM-`) == 1 && strings.Contains(o, "  SCRUM-6 Request SCRUM-6 https://jira.test/browse/SCRUM-6"))
		s.check("and its proposals to review", strings.Contains(after(o, "Proposals to review", 3), "example.test/pr/tend_jira"))
	}},
	{"jira-handoff", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.jmap()
		s.issues()
		s.jissue("SCRUM-1", "tend", "new", "", 0)
		r := s.propose("COMMENT_FAIL=1")
		s.check("a Jira comment that fails after the pull request stops", r.code == 3)
		s.check("and says a rerun finishes", r.said("run tend propose again to finish"))
		js := list(s.readJSON(s.st + "/jira"))
		for _, j := range js {
			j.(map[string]any)["fields"].(map[string]any)["labels"] = []any{}
		}
		s.writeJSON(s.st+"/jira", js)
		s.propose()
		s.check("the next pass finishes it, though the label is gone", s.count(s.jcomments("SCRUM-1"), `tend opened`) == 1)
		s.check("without a second pull request", s.creates() == 1)
		s.check("and records it finished", nonempty(s.ts+"/intake/o_r-jira-SCRUM-1/linked"))
	}},
	{"jira-declined", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.jmap()
		s.issues()
		s.jissue("SCRUM-1", "tend", "new", "", 0)
		s.propose("DECLINE=1")
		c := s.jcomments("SCRUM-1")
		s.check("a Jira request no proposal can be drafted from is answered", strings.Contains(c, "No proposal could be drafted from this request"))
		s.check("with the drafter's reason", s.has(c, `^It names no outcome`))
		s.check("and a marker to find it by", s.has(c, `^\[tend:declined [0-9]*\]$`))
		s.check("and loses its tend label", len(list(get(s.readJSON(s.st+"/jira"), 0, "fields", "labels"))) == 0)
		m, _ := filepath.Glob(s.ts + "/intake/o_r-jira-SCRUM-1.declined-*")
		s.check("its run folder is set aside", len(m) > 0)
	}},
	{"jira-build", func(s *sc) {
		s.plain()
		s.allow()
		s.jmap()
		s.jissue("SCRUM-7", "tend-build", "new", rich, 0)
		r := s.jbuild()
		s.check("a Jira issue labelled tend-build is built, reviewed and published", r.passed())
		s.check("the implementer is given the Jira issue as its scope", s.has(s.log(), `^agent prompt impl-jira-scrum-7-[0-9]* Implement Jira issue SCRUM-7 .* Implement the spec \(https://docs.test/spec\)`))
		s.check("on branch tend/build-SCRUM-7, pushed", s.gitOK(s.repo+".git", "rev-parse", "--verify", "-q", "tend/build-SCRUM-7"))
		s.check("the commit message is asked to refer to the key", s.has(s.log(), `^agent prompt impl-jira-scrum-7-.*End with the line 'Refs SCRUM-7'`))
		s.check("the Jira issue is told where its pull request is", strings.Contains(s.jcomments("SCRUM-7"), "tend opened https://example.test/pr/tend_build-SCRUM-7 "))
		s.check("its attempt is named for the key and repository", isDir(s.ts+"/runs/jira-scrum-7-"+s.cksum(s.repo)))
		s.check("status finds it", s.has(s.tendIn(nil, "status", s.repo, "SCRUM-7").out, `^attempt jira-scrum-7-.* of SCRUM-7: tend/build-SCRUM-7`))
		s.jbuild()
		s.check("a rerun builds, reviews and comments nothing again", s.count(s.log(), `^agent prompt impl-jira`) == 2 && s.reviews() == 1 && s.count(s.jcomments("SCRUM-7"), `tend opened`) == 1)
	}},
	{"jira-ineligible", func(s *sc) {
		s.plain()
		s.allow()
		s.jmap()
		s.jissue("SCRUM-7", "", "new", "", 0)
		s.jissue("SCRUM-8", "tend-build", "done", "", 0)
		r := s.jbuild()
		s.check("an unlabelled Jira issue is not built", r.code == 3 && r.said("issue SCRUM-7 is not labelled tend-build"))
		r = s.tendIn(nil, s.repo, "SCRUM-8")
		s.check("a done one has nothing to do", r.code == 0 && r.said("issue SCRUM-8 is closed"))
		r = s.tendIn(nil, s.repo, "OTHER-1")
		s.check("an unmapped project is refused", r.code == 3 && r.said("Jira project OTHER is not mapped"))
		s.write(s.ts+"/jira", regexp.MustCompile(`(?m) [^ ]*$`).ReplaceAllString(s.cat(s.ts+"/jira"), " git@github.com:someone/else.git")+"\n")
		r = s.jbuild()
		s.check("a project mapped to another repository is refused", r.code == 3 && r.said("not to this repository's origin"))
	}},
}
