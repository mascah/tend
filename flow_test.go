package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The fault tend finds in a flow file's text, or "": prompts named p*.md exist.
func faultOf(text string) string {
	var ff flowFile
	if why := decode(text, &ff); why != "" {
		return why
	}
	for name, f := range ff.named() {
		if *f != nil {
			if why := (*f).fault(name, func(p string) bool { return strings.HasPrefix(p, "p") }); why != "" {
				return why
			}
		}
	}
	return ""
}

func TestBuiltInFlowsAreValid(t *testing.T) {
	for _, name := range []string{"build", "propose"} {
		b, err := builtin.ReadFile("flows/" + name + ".yaml")
		if err != nil {
			t.Fatal(err)
		}
		var ff flowFile
		if why := decode(string(b), &ff); why != "" {
			t.Fatalf("%s %s", name, why)
		}
		if why := (*ff.named()[name]).fault(name, func(p string) bool { _, ok := promptText("", p); return ok }); why != "" {
			t.Errorf("flows/%s.yaml %s", name, why)
		}
	}
}

const loopy = `
build:
  start: review
  steps:
    review:
      agent: codex
      prompt: p.md
      result: verdict v-{head}.json
      next: {actionable: repair, success: end}
    repair:
      continue: review
      prompt: p.md
      result: commit
      next: {ok: review}
`

func TestFlowFaults(t *testing.T) {
	for _, c := range []struct{ name, text, want string }{
		{"a loop with no limit", loopy, "loops through review and repair with no limit"},
		{"a limit breaks the loop", strings.Replace(loopy, "result: commit", "result: commit\n      limit: 2", 1), ""},
		{"a flag", strings.Replace(loopy, "prompt: p.md\n      result: commit", "prompt: p.md\n      dangerously-skip-permissions: true\n      result: commit", 1), "field dangerously-skip-permissions not found in a step"},
		{"not YAML", "build: [", "is not a flow file tend can read"},
		{"empty", "", "is empty"},
		{"a flow tend does not run", "deploy:\n  start: x\n", "field deploy not found at the top"},
		{"no such harness", strings.Replace(loopy, "agent: codex", "agent: gpt", 1), "the harness gpt, where an agent is claude or codex"},
		{"no such step", strings.Replace(loopy, "success: end", "success: check", 1), `"check" after success, a step the flow does not have`},
		{"no such outcome", strings.Replace(loopy, "{ok: review}", "{ok: review, spent: end}", 1), "the outcome spent, which it cannot have"},
		{"no such prompt", strings.Replace(loopy, "prompt: p.md\n      result: verdict", "prompt: q.md\n      result: verdict", 1), "the prompt file q.md, which does not exist"},
		{"no such result", strings.Replace(loopy, "result: commit", "result: a pony", 1), `the result "a pony"`},
		{"a continue of no agent", strings.Replace(loopy, "continue: review", "continue: publish", 1), `"publish" to continue`},
		{"publish with an agent", loopy + "    publish:\n      agent: claude\n", "publishing is tend's own"},
		{"no start", strings.Replace(loopy, "start: review", "start: nothing", 1), `at "nothing"`},
	} {
		got := faultOf(c.text)
		if c.want == "" && got != "" || !strings.Contains(got, c.want) {
			t.Errorf("%s: fault %q, want %q", c.name, got, c.want)
		}
	}
}

// A person requesting changes bounds the loop through publish.
func TestRequestedLoopNeedsNoLimit(t *testing.T) {
	text := `
build:
  start: review
  steps:
    review: {agent: codex, prompt: p.md, result: verdict v.json, next: {success: publish, failure: publish, actionable: publish}}
    publish: {next: {ok: end, requested: request}}
    request: {continue: review, prompt: p.md, result: commit, next: {ok: review}}
`
	if got := faultOf(text); got != "" {
		t.Errorf("fault %q", got)
	}
}

// ---- the flows spec's scenarios, through the binary and the stand-ins

// The built-in build flow as text, with each pair of old, new replaced once.
func buildFlow(t *testing.T, edits ...string) string {
	b, _ := builtin.ReadFile("flows/build.yaml")
	s := string(b)
	for i := 0; i < len(edits); i += 2 {
		if !strings.Contains(s, edits[i]) {
			t.Fatalf("the built-in build flow has no %q", edits[i])
		}
		s = strings.Replace(s, edits[i], edits[i+1], 1)
	}
	return s
}

// A step check after a passing review, whose result is a verdict.
const checkStep = `    check:
      agent: codex
      prompt: prompts/check.md
      result: verdict check-{head}.json
      next:
        success: publish
    publish:`

const checkPrompt = "Check branch {branch} at {head}; write only the verdict {result}.\n"

func withCheck(t *testing.T) map[string]string {
	return map[string]string{
		"tend/flow.yaml":        buildFlow(t, "        success: publish\n", "        success: check\n", "    publish:", checkStep),
		"tend/prompts/check.md": checkPrompt,
	}
}

// The line number of the first log line matching re, or -1.
func (s *sc) at(re string) int {
	for i, l := range strings.Split(s.log(), "\n") {
		if regexp.MustCompile(re).MatchString(l) {
			return i
		}
	}
	return -1
}

func (s *sc) agentStarts() int { return s.count(s.log(), `^agent start `) }

var flowScenarios = []scenario{
	{"flow-one-more-step", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Check each head", withCheck(s.t))
		r := s.run()
		s.check("a build with a step of the project's own passes", r.passed())
		rev, chk, pub := s.at(`^agent prompt rev-`), s.at(`^agent prompt check-`), s.at(`^gh pr create`)
		s.check("check is prompted after the passing review", rev >= 0 && chk > rev)
		s.check("and before publishing", pub > chk)
		s.check("with the project's prompt, filled in", s.has(s.log(), `^agent prompt check-[0-9a-f]{7} Check branch grove/.* at [0-9a-f]{40}; write only the verdict .*/check-[0-9a-f]{40}\.json`))
		s.check("its agent is started with the reviewer's flags", s.has(s.log(), `^agent start check-.* -s workspace-write .* -a never --add-dir `))
		s.issues("1:tend")
		p := s.propose()
		s.check("a request is drafted by the built-in propose flow", p.code == 0 && s.drafts() == 1 && exists(s.prOf("1")))
	}},
	{"flow-from-target", func(s *sc) {
		s.project("main", true)
		s.run()
		dropped := buildFlow(s.t, "        ok: review\n    review:", "        ok: publish\n    review:")
		s.bash(s.worktree(), `mkdir -p tend && printf '%s' "$0" >tend/flow.yaml && git add -A && git commit -qm "Drop the review"`, dropped)
		r := s.run()
		s.check("a flow the implementer commits does not change the attempt", r.passed())
		s.check("its head is reviewed as the target's flow says", s.reviews() == 2)
		s.check("and the edit is part of the candidate", s.git(s.worktree(), "ls-tree", "--name-only", "HEAD", "tend/flow.yaml") == "tend/flow.yaml")
	}},
	{"flow-no-result", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Check each head", withCheck(s.t))
		r := s.run("CHECK=none")
		s.check("a step that leaves no result stops for a person", r.code == 3)
		s.check("saying check wrote no check-<head>.json", r.said(`check wrote no check-[0-9a-f]{40}\.json`))
		s.check("and no later step runs", s.creates() == 0 && s.count(s.log(), `^gh `) == 0)
	}},
	{"flow-repair-continues", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=repairable")
		s.check("a repaired build passes", r.passed())
		s.check("the repair goes to the implementer's own agent", s.prompts("An independent review") == 1)
		s.check("one implementer is started", s.count(s.log(), `^agent start impl-`) == 1)
		s.check("and one reviewer for each head", s.count(s.log(), `^agent start rev-`) == 2)
	}},
	{"flow-flag", func(s *sc) {
		s.project("main", true)
		s.commit("Loosen the implementer", map[string]string{"tend/flow.yaml": buildFlow(s.t, "      agent: claude\n", "      agent: claude\n      flags: --dangerously-skip-permissions\n")})
		r := s.run()
		s.check("a flow that sets a flag stops as not valid", r.code == 3 && r.said(`tend/flow.yaml is not a flow file tend can read: line [0-9]+: field flags not found in a step`))
		s.check("and starts no agent", s.agentStarts() == 0)
	}},
	{"flow-check-fails", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Check each head", withCheck(s.t))
		r := s.run("CHECK=failure")
		s.check("a failing check with nothing to go to stops for a person", r.code == 3)
		s.check("naming check and the verdict's summary", r.said(`check gave failure: a\.txt:1 is not checked\.`))
		s.check("and nothing is published", s.creates() == 0)
	}},
	{"flow-second-repair", func(s *sc) {
		s.project("main", false)
		s.commit("Two repairs", map[string]string{"tend/flow.yaml": buildFlow(s.t, "      limit: 1\n      next:\n        ok: review\n        spent: publish\n", "      limit: 2\n      next:\n        ok: review\n")})
		r := s.run("REVIEW=actionable")
		s.check("the repair is prompted twice", s.prompts("An independent review") == 2)
		s.check("three heads are reviewed", s.reviews() == 3)
		s.check("and tend stops at the third failure", r.code == 3 && r.said(`repair was asked 2 times, its limit`))
	}},
	{"flow-one-repair", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=actionable")
		s.check("the built-in flow asks one repair", s.prompts("An independent review") == 1)
		s.check("and exits 0 with the failed verdict", r.code == 0 && r.conclusion() == "failure")
	}},
	{"flow-killed-in-step", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Check each head", withCheck(s.t))
		r := s.run("CHECK_FAIL=timeout")
		s.check("a check whose prompt timed out stops", r.code == 3)
		s.agent("working", "check-")
		r = s.run()
		s.check("the rerun waits for the check still working", r.passed() && r.said("still working on an earlier prompt"))
		s.check("check was prompted once in all", s.promptsTo("check-", "Check") == 1)
		s.check("and the flow goes on to publish", s.creates() == 1)
	}},
	{"flow-unknown-outcome", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Check each head", withCheck(s.t))
		s.run("CHECK_FAIL=timeout")
		prompts := s.count(s.log(), `^agent prompt `)
		r := s.run()
		s.check("a step asked with no result and no agent working stops", r.code == 3)
		s.check("saying check was asked and left nothing", r.said(`check was asked at [0-9a-f]{7} and left nothing`))
		s.check("and prompts no agent", s.count(s.log(), `^agent prompt `) == prompts)
		r = s.run()
		s.check("on every run", r.code == 3 && s.count(s.log(), `^agent prompt `) == prompts)
		os.Remove(glob(s.ts + "/runs/*/check-asked"))
		r = s.run()
		s.check("until a person deletes its record, and it is asked once more", r.passed() && s.promptsTo("check-", "Check") == 2)
	}},
	{"flow-without-publish", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Keep it here", map[string]string{"tend/flow.yaml": buildFlow(s.t, "        success: publish\n        failure: publish\n", "        success: end\n        failure: end\n")})
		r := s.run()
		s.check("a flow that ends after review prints the verdict", r.passed())
		s.check("pushes no branch", s.branch() != "" && !s.gitOK(s.repo+".git", "rev-parse", "--verify", "-q", "refs/heads/"+s.branch()))
		s.check("and opens no pull request", s.creates() == 0)
	}},
	{"flow-loop", func(s *sc) {
		s.project("main", true)
		s.commit("Repair for ever", map[string]string{"tend/flow.yaml": buildFlow(s.t, "      limit: 1\n", "", "        spent: publish\n", "")})
		r := s.run()
		s.check("a loop with no limit exits 3", r.code == 3)
		s.check("saying so", r.said(`tend/flow.yaml loops through review and repair with no limit`))
		s.check("and no agent is started", s.agentStarts() == 0)
	}},
	{"flow-print", func(s *sc) {
		s.project("main", true)
		p := s.tendIn(nil, "flow", s.repo)
		s.check("tend flow prints the flows in effect", p.code == 0 && strings.Contains(p.out, "# build: built in") && strings.Contains(p.out, "# propose: built in"))
		s.commit("Start our own flow", map[string]string{"tend/flow.yaml": p.out})
		s.check("saved as the project's file, it is valid", faultOf(p.out) == "")
		r := s.run("REVIEW=repairable")
		s.check("and attempts behave as before", r.passed() && s.prompts("An independent review") == 1 && s.reviews() == 2)
		p = s.tendIn(nil, "flow", s.repo)
		s.check("tend flow then says the flows are the project's", strings.Contains(p.out, "# build: tend/flow.yaml on origin/main"))
	}},
}

func TestFlows(t *testing.T) {
	for _, x := range flowScenarios {
		t.Run(x.name, func(t *testing.T) { x.fn(newSc(t)) })
	}
}
