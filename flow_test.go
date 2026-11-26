package main

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// The fault tend finds in a flow file's text, or "": prompts named p*.md exist.
func faultOf(text string) string {
	var ff flowFile
	if why := decode(text, &ff); why != "" {
		return why
	}
	if why := ff.Watch.fault(); why != "" {
		return why
	}
	if why := ff.protectFault(); why != "" {
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

func TestProtectFaults(t *testing.T) {
	for _, c := range []struct{ name, text, want string }{
		{"a list", "protect: [schema.sql, migrations/]\n", ""},
		{"an empty list", "protect: []\n", ""},
		{"an empty entry", "protect: [schema.sql, \"\"]\n", "gives protect an empty entry"},
		{"a blank entry", "protect:\n  - \" \"\n", "gives protect an empty entry"},
		{"the root", "protect: [/]\n", `gives protect the entry "/"`},
		{"not a list", "protect: schema.sql\n", "protect is not a list of paths"},
		{"a null entry", "protect: [null]\n", "gives protect an empty entry"},
		{"null", "protect: null\n", "gives protect no list"},
		{"no value", "protect:\n", "gives protect no list"},
		{"no protect", "watch: {at_once: 1}\n", ""},
		{"a null among others", "protect:\n  - schema.sql\n  -\n", "gives protect an empty entry"},
		{"a list in the list", "protect: [[a]]\n", "protect holds an entry that is not a path"},
		{"a doubled slash", "protect: [\"//migrations//\"]\n", `gives protect the entry "//migrations//"`},
		{"one leading slash", "protect: [/schema.sql]\n", ""},
		{"a leading dot", "protect: [./schema.sql]\n", `gives protect the entry "./schema.sql"`},
		{"a dot-dot", "protect: [db/../schema.sql]\n", `gives protect the entry "db/../schema.sql"`},
		{"a dot folder", "protect: [.claude/, .env]\n", ""},
		{"an alias", "protect: [&s schema.sql, *s]\n", ""},
		{"an alias of a list", "x: &l [a]\nprotect: [*l]\n", "protect holds an entry that is not a path"},
	} {
		got := faultOf(c.text)
		if c.want == "" && got != "" || !strings.Contains(got, c.want) {
			t.Errorf("%s: fault %q, want %q", c.name, got, c.want)
		}
	}
}

func TestProtectedPaths(t *testing.T) {
	var ff flowFile
	if ps, own := ff.protected(); own || !slices.Equal(ps, defaultProtect) {
		t.Errorf("no protect gives %v, own %v; want tend's own list", ps, own)
	}
	ff.Protect = &pathList{"schema.sql", "migrations/"}
	if ps, own := ff.protected(); !own || !slices.Equal(ps, []string{"schema.sql", "migrations/", "tend/"}) {
		t.Errorf("a project's protect gives %v, own %v; want its entries and tend/", ps, own)
	}
	ff.Protect = &pathList{"tend/"}
	if ps, _ := ff.protected(); !slices.Equal(ps, []string{"tend/"}) {
		t.Errorf("a protect naming tend/ gives %v", ps)
	}
	paths := []string{
		"main.go", "tend/flow.yaml", "tend/prompts/review.md", "tendency.md", ".github/workflows/ci.yml", ".github/CODEOWNERS",
		"CLAUDE.md", "web/CLAUDE.md", "docs/CLAUDE.md.bak", "wrangler.toml", "web/wrangler.jsonc", "go.sum", "vendor/x/go.sum",
		".claude/settings.json", "grove.yaml", "grove/specs/x/spec.md", "schema.sql", "db/schema.sql", "migrations/001.sql", "db/migrations/001.sql",
	}
	for _, c := range []struct {
		entries, want []string
	}{
		{defaultProtect, []string{"tend/flow.yaml", "tend/prompts/review.md", ".github/workflows/ci.yml", ".github/CODEOWNERS", "CLAUDE.md", "web/CLAUDE.md",
			"wrangler.toml", "web/wrangler.jsonc", "go.sum", "vendor/x/go.sum", ".claude/settings.json", "grove.yaml"}},
		{[]string{"schema.sql", "migrations/"}, []string{"schema.sql", "db/schema.sql", "migrations/001.sql"}},
		{[]string{"/schema.sql", "db/migrations/"}, []string{"schema.sql", "db/migrations/001.sql"}},
		// .github/*/ covers .github/CODEOWNERS too: a path at a folder
		// entry's own place may be a gitlink, which Git's name does not tell.
		{[]string{"db/*.sql", ".github/*/"}, []string{".github/workflows/ci.yml", ".github/CODEOWNERS", "db/schema.sql"}},
	} {
		if got := covered(c.entries, paths); !slices.Equal(got, c.want) {
			t.Errorf("%v covers %v, want %v", c.entries, got, c.want)
		}
	}
	// A folder entry covers the folder's own path, as a gitlink there is, and
	// no path that only begins with its name.
	for _, c := range []struct {
		entries, want []string
	}{
		{defaultProtect, []string{".claude", "tend", ".github/workflows"}},
		{[]string{"migrations/"}, []string{"migrations"}},
	} {
		if got := covered(c.entries, []string{".claude", ".clauderc", "tend", "tendency", ".github/workflows", ".github/workflows.md", "migrations", "db/migrations"}); !slices.Equal(got, c.want) {
			t.Errorf("%v covers %v, want %v", c.entries, got, c.want)
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

// A flow that lands: publish sends on to land, a conflict and a failed check
// go back to the implementer once each.
const landing = `
build:
  start: implement
  steps:
    implement: {agent: claude, prompt: p.md, result: commit, next: {ok: review}}
    review: {agent: codex, prompt: p.md, result: "verdict v-{head}.json", next: {success: publish, failure: publish}}
    publish: {next: {ok: land}}
    land: {next: {ok: end, conflict: resolve, failed: fix}}
    resolve: {continue: implement, prompt: p.md, result: commit, limit: 1, next: {ok: review}}
    fix: {continue: implement, prompt: p.md, result: commit, limit: 1, next: {ok: review}}
`

func TestLandFaults(t *testing.T) {
	for _, c := range []struct{ name, text, want string }{
		{"a flow that lands", landing, ""},
		{"land without publish", strings.Replace(strings.Replace(landing, "publish: {next: {ok: land}}", "", 1), "success: publish, failure: publish", "success: land, failure: land", 1), "a step land and no publish"},
		{"land given a prompt", strings.Replace(landing, "land: {next:", "land: {prompt: p.md, next:", 1), "more than next, though landing is tend's own"},
		{"land given an agent", strings.Replace(landing, "land: {next:", "land: {agent: claude, next:", 1), "landing is tend's own"},
		{"an outcome land cannot have", strings.Replace(landing, "failed: fix}", "failed: fix, spent: end}", 1), "it has ok, conflict and failed"},
		{"a conflict step with no limit", strings.Replace(landing, "limit: 1, next: {ok: review}}\n    fix", "next: {ok: review}}\n    fix", 1), "loops through review, publish, land and resolve with no limit"},
		{"a continue of land", strings.Replace(landing, "resolve: {continue: implement", "resolve: {continue: land", 1), `"land" to continue`},
		{"land in propose", "propose:\n  start: draft\n  steps:\n    draft: {agent: claude, prompt: p.md, next: {ok: publish}}\n    publish: {next: {ok: land}}\n    land: {next: {ok: end}}\n", "where tend lands nothing"},
		// Only people's approvals bring a supervised flow round to land again.
		{"a loop only through approved", strings.Replace(strings.Replace(landing, "{ok: land}", "{ok: end, approved: land}", 1), "limit: 1, ", "", -1), ""},
	} {
		got := faultOf(c.text)
		if c.want == "" && got != "" || !strings.Contains(got, c.want) {
			t.Errorf("%s: fault %q, want %q", c.name, got, c.want)
		}
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
	{"flow-land-faults", func(s *sc) {
		s.project("main", true)
		land := landFlow(s.t, false)["tend/flow.yaml"]
		for _, c := range []struct{ what, text, want string }{
			{"land without publish", strings.NewReplacer("    publish:\n      next:\n        ok: land\n        requested: request\n", "",
				"        success: publish\n        failure: publish\n", "        success: land\n        failure: land\n", "        spent: publish\n", "        spent: land\n").Replace(land),
				`^tend: tend/flow\.yaml gives flow build a step land and no publish`},
			{"land given a prompt", strings.Replace(land, "    land:\n", "    land:\n      prompt: prompts/review.md\n", 1), `^tend: tend/flow\.yaml gives step land of flow build more than next, though landing is tend's own$`},
			{"a conflict step with no limit", strings.Replace(land, "      prompt: prompts/conflict.md\n      result: commit\n      limit: 1\n", "      prompt: prompts/conflict.md\n      result: commit\n", 1), `^tend: tend/flow\.yaml loops through .* with no limit$`},
		} {
			s.commit(c.what, map[string]string{"tend/flow.yaml": c.text})
			r := s.run()
			s.check(c.what+": tend exits 3 saying why", r.code == 3 && r.said(c.want))
			s.check(c.what+": and no agent is started", s.agentStarts() == 0)
		}
		s.commit("Land", map[string]string{"tend/flow.yaml": land})
		p := s.tendIn(nil, "flow", s.repo)
		s.check("a flow naming tend's own prompts for conflict and failed, with none beside it, is valid", p.code == 0 && strings.Contains(p.out, "prompt: prompts/conflict.md") && strings.Contains(p.out, "prompt: prompts/failed.md"))
	}},
	{"flow-publish-failed", func(s *sc) {
		s.waitsForCI("test success")
		p := s.tendIn(nil, "flow", s.repo)
		s.check("a build flow naming failed on publish is valid, and tend flow prints it", p.code == 0 && regexp.MustCompile(`(?m)^ +publish:\n +next:\n( +[a-z]+: [a-z]+\n)* +failed: fix$`).MatchString(p.out))
		s.commit("Fail at publish", map[string]string{"tend/flow.yaml": strings.Replace(landFlow(s.t, true)["tend/flow.yaml"], "        approved: land\n", "        approved: land\n        failed: nowhere\n", 1)})
		r := s.run()
		s.check("one naming a step it lacks for failed is refused", r.code == 3 && r.said(`gives step publish of flow build "nowhere" after failed, a step the flow does not have`))
		s.commit("Fail without fix", map[string]string{"tend/flow.yaml": strings.Replace(buildFlow(s.t), "        ok: end\n        requested: request\n", "        ok: end\n        requested: request\n        conflict: end\n", 1)})
		r = s.run()
		s.check("and an outcome publish cannot have is refused, naming failed among those it has", r.code == 3 && r.said(`the outcome conflict, which it cannot have; it has ok, requested, approved and failed$`))
	}},
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
	{"flow-agent-gone", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Check each head", withCheck(s.t))
		s.run("CHECK_FAIL=timeout")
		s.gone("check-")
		r := s.run("CHECK_FAIL=timeout")
		s.check("a step whose agent is gone is prompted a second time", s.promptsTo("check-", "Check") == 2)
		s.check("in a new session", s.count(s.log(), `^agent start check-`) == 2)
		s.check("and a second ask that leaves nothing stops", r.code == 3)
		s.gone("check-")
		prompts := s.count(s.log(), `^agent prompt `)
		r = s.run()
		s.check("gone a second time, tend stops saying check was asked twice and left nothing", r.code == 3 && r.said(`check was asked twice at [0-9a-f]{7} and left nothing`))
		s.check("and prompts no agent", s.count(s.log(), `^agent prompt `) == prompts)
	}},
	{"flow-terminal-unreachable", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Check each head", withCheck(s.t))
		s.run("CHECK_FAIL=timeout")
		prompts := s.count(s.log(), `^agent prompt `)
		r := s.run("AGENT_GET_FAIL=1")
		s.check("a terminal that cannot be asked about the agent is not taken as the agent gone: tend stops", r.code == 3 && r.said(`check was asked at [0-9a-f]{7} and left nothing`))
		s.check("and prompts no agent", s.count(s.log(), `^agent prompt `) == prompts)
	}},
	{"flow-agent-gone-once", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.commit("Check each head", withCheck(s.t))
		s.run("CHECK_FAIL=timeout")
		s.gone("check-")
		r := s.run()
		s.check("a step whose agent is gone is asked once more, its result taken, and the flow goes on", r.passed() && s.creates() == 1)
		s.check("prompted twice in all, in a new session", s.promptsTo("check-", "Check") == 2 && s.count(s.log(), `^agent start check-`) == 2)
	}},
	{"flow-repair-once-more", func(s *sc) {
		s.project("main", false)
		r := s.run("REVIEW=actionable", "REPAIR_FAIL=timeout")
		s.check("a repair prompt that timed out stops", r.code == 3)
		s.gone("impl-")
		r = s.run("REVIEW=actionable")
		s.check("the repair is asked a second time though its limit is 1", s.prompts("An independent review") == 2)
		s.check("the head it makes is reviewed", s.reviews() == 2)
		s.check("and no third repair is asked: tend exits 0 with the failed verdict", r.code == 0 && r.conclusion() == "failure")
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
		s.check("with no watch in a repository without the setting", !strings.Contains(p.out, "watch"))
		s.commit("Start our own flow", map[string]string{"tend/flow.yaml": p.out})
		s.check("saved as the project's file, it is valid", faultOf(p.out) == "")
		r := s.run("REVIEW=repairable")
		s.check("and attempts behave as before", r.passed() && s.prompts("An independent review") == 1 && s.reviews() == 2)
		p = s.tendIn(nil, "flow", s.repo)
		s.check("tend flow then says the flows are the project's", strings.Contains(p.out, "# build: tend/flow.yaml on origin/main"))
		s.check("and still prints no watch", !strings.Contains(p.out, "watch"))
	}},
	{"flow-protect-default", func(s *sc) {
		s.project("main", true)
		p := s.tendIn(nil, "flow", s.repo)
		s.check("tend flow prints protect, saying it is tend's own", p.code == 0 && strings.Contains(p.out, "\n# protect: built in, tend's own list."))
		for _, e := range []string{"tend/", ".github/workflows/", "CODEOWNERS", "PRINCIPLES.md", "grove.yaml", "CLAUDE.md", "AGENTS.md", ".claude/", "wrangler.*", "Dockerfile", "go.sum", "package-lock.json", "Cargo.lock"} {
			s.check("with "+e+" in tend's own list", regexp.MustCompile(`(?m)^ +- '?`+regexp.QuoteMeta(e)+`'?$`).MatchString(p.out))
		}
		var ff flowFile
		s.check("which saved is a valid flow file, giving tend's own list as the project's", decode(p.out, &ff) == "" && ff.Protect != nil && slices.Equal([]string(*ff.Protect), defaultProtect))
	}},
	{"flow-protect-own", func(s *sc) {
		s.project("main", true)
		s.commit("Protect the schema", map[string]string{"tend/flow.yaml": "protect:\n  - schema.sql\n  - migrations/\n"})
		p := s.tendIn(nil, "flow", s.repo)
		s.check("tend flow prints the project's protect, saying it is the project's", p.code == 0 && strings.Contains(p.out, "\n# protect: tend/flow.yaml on origin/main, the project's own, with tend/ always protected."))
		s.check("holding its entries and tend/", strings.Contains(p.out, "\nprotect:\n    - schema.sql\n    - migrations/\n    - tend/\n"))
		s.check("and not tend's own list", !strings.Contains(p.out, "go.sum"))
		s.check("and the cycle is built as the built-in flow says", s.run().passed() && s.reviews() == 1)
	}},
	{"flow-protect-empty", func(s *sc) {
		s.project("main", true)
		s.commit("Protect nothing", map[string]string{"tend/flow.yaml": "protect:\n  - schema.sql\n  - \"\"\n"})
		r := s.run()
		s.check("an empty entry stops as not valid, naming the file and protect", r.code == 3 && r.said(`^tend: tend/flow\.yaml gives protect an empty entry, where each entry is a path from the repository's root$`))
		s.check("and starts no agent", s.agentStarts() == 0)
		s.commit("Protect null", map[string]string{"tend/flow.yaml": "protect: null\n"})
		r = s.run()
		s.check("a protect of null stops as not valid, naming the file and protect", r.code == 3 && r.said(`^tend: tend/flow\.yaml gives protect no list, where protect is a list of paths$`))
		s.check("and still starts no agent", s.agentStarts() == 0)
		s.check("and tend flow refuses it too", s.tendIn(nil, "flow", s.repo).code == 3)
	}},
	{"flow-watch-count", func(s *sc) {
		s.project("main", true)
		s.commit("Keep two going", map[string]string{"tend/flow.yaml": "watch:\n  at_once: 2\n"})
		p := s.tendIn(nil, "flow", s.repo)
		s.check("tend flow prints the count", p.code == 0 && strings.Contains(p.out, "\nwatch:\n  at_once: 2\n"))
		s.check("and both built-in flows", strings.Contains(p.out, "# build: built in") && strings.Contains(p.out, "# propose: built in"))
		s.check("which saved again is valid", faultOf(p.out) == "")
		s.check("and the cycle is built as the built-in flow says", s.run().passed() && s.reviews() == 1)
	}},
	{"flow-watch-zero", func(s *sc) {
		s.project("main", true)
		s.commit("Keep none going", map[string]string{"tend/flow.yaml": "watch:\n  at_once: 0\n"})
		r := s.run()
		s.check("a count of 0 stops as not valid, naming the file and at_once", r.code == 3 && r.said(`^tend: tend/flow\.yaml gives watch an at_once of 0, where at_once is a whole number of at least 1$`))
		s.check("and starts no agent", s.agentStarts() == 0)
		for text, want := range map[string]string{
			"watch:\n  at_once: two\n":             "an at_once of two",
			"watch:\n  at_once: 2\n  first: old\n": "field first not found in watch",
			"watch: {}\n":                          "gives watch no at_once",
		} {
			s.check("not valid: "+want, strings.Contains(faultOf(text), want))
		}
	}},
	{"flow-reviewer-writing", func(s *sc) { // review finding 1
		s.project("main", false)
		r := s.run("REV_FAIL=timeout")
		s.check("a reviewer prompt that timed out stops", r.code == 3)
		s.write(s.runDir()+"/verdict-"+s.head()+".json", `{"head": "`)
		s.agent("working", "rev-")
		r = s.run()
		s.check("the rerun waits for the reviewer writing its verdict and passes", r.passed())
		s.check("prompting it once in all", s.reviews() == 1)
		s.check("and setting nothing aside", glob(s.runDir()+"/*.rejected-*") == s.runDir()+"/*.rejected-*")
	}},
	{"flow-cached-loop", func(s *sc) { // review finding 2
		s.project("main", false)
		s.commit("Back to the implementer", map[string]string{"tend/flow.yaml": buildFlow(s.t,
			"        failure: publish\n", "        failure: implement\n",
			"      result: commit\n      next:\n        ok: review\n    review:", "      result: commit\n      limit: 1\n      next:\n        ok: review\n        spent: end\n    review:")})
		r := s.run("REVIEW=decision")
		s.check("a loop that only meets cached results stops for a person", r.code == 3 && r.said(`build reached implement again at [0-9a-f]{7} with nothing changed`))
	}},
	{"flow-continue-commit", func(s *sc) { // review finding 3
		s.project("main", false)
		s.commit("Edit, then tidy", map[string]string{
			"tend/flow.yaml": buildFlow(s.t, "      next:\n        ok: review\n    review:", `      next:
        ok: edit
    edit:
      agent: claude
      prompt: prompts/edit.md
      result: commit
      next:
        ok: tidy
    tidy:
      continue: edit
      prompt: prompts/tidy.md
      result: commit
      next:
        ok: review
    review:`),
			"tend/prompts/edit.md": "Edit it.\n",
			"tend/prompts/tidy.md": "Tidy it.\n",
		})
		r := s.run()
		s.check("a flow continuing a step that committed passes", r.passed())
		s.check("one edit agent is started", s.count(s.log(), `^agent start edit-`) == 1)
		s.check("and tidy goes to it", s.count(s.log(), `^agent prompt edit-`) == 2 && sameAgent(s.log(), `(?m)^agent prompt (edit-[0-9a-f]+) `))
	}},
	{"flow-verdict-name", func(s *sc) { // review finding 4
		s.project("main", false)
		s.commit("Name the verdict", map[string]string{"tend/flow.yaml": buildFlow(s.t, "verdict verdict-{head}.json", "verdict review-{head}.json")})
		r := s.run("REVIEW=repairable")
		s.check("the printed verdict is the repaired head's", r.passed())
	}},
	{"flow-start-limit", func(s *sc) { // review finding 5
		s.project("main", false)
		s.commit("Implement once", map[string]string{"tend/flow.yaml": buildFlow(s.t, "      result: commit\n      next:\n        ok: review\n    review:", "      result: commit\n      limit: 1\n      next:\n        ok: review\n    review:")})
		r := s.run("UNFINISHED=1")
		s.check("an unfinished implementation stops", r.code == 3)
		r = s.run()
		s.check("and at its limit is not asked again", r.code == 3 && r.said(`implement was asked 1 times, its limit`) && s.prompts("/grove-implement") == 1)
	}},
	{"flow-plan-first", func(s *sc) { // review finding 6
		s.project("main", false)
		s.commit("Plan first", map[string]string{
			"tend/flow.yaml": buildFlow(s.t, "  start: implement\n", "  start: plan\n", "  steps:\n", `  steps:
    plan:
      agent: claude
      prompt: prompts/plan.md
      result: commit
      next:
        ok: implement
`, "    implement:\n      agent: claude\n", "    implement:\n      continue: plan\n"),
			"tend/prompts/plan.md": "Plan it.\n",
		})
		r := s.run()
		s.check("a step that commits before the implementation does not stop the build", r.passed())
		s.check("plan and implement are each asked once", s.prompts("Plan it.") == 1 && s.prompts("/grove-implement") == 1)
	}},
	{"flow-plan-then-fresh-implementer", func(s *sc) { // review round 3
		s.project("main", false)
		s.commit("Plan first", map[string]string{
			"tend/flow.yaml": buildFlow(s.t, "  start: implement\n", "  start: plan\n", "  steps:\n", `  steps:
    plan:
      agent: claude
      prompt: prompts/plan.md
      result: commit
      next:
        ok: implement
`),
			"tend/prompts/plan.md": "Plan it.\n",
		})
		r := s.run("REVIEW=repairable")
		s.check("a repaired build with its own implementer passes", r.passed())
		s.check("one implementer agent is started", s.count(s.log(), `^agent start implement-`) == 1)
		s.check("and the repair goes to it", sameAgent(s.log(), `(?m)^agent prompt (implement-[0-9a-f]+) `) && s.count(s.log(), `^agent prompt implement-[0-9a-f]+ An independent review`) == 1)
	}},
	{"flow-plan-first-survivor", func(s *sc) { // review round 2
		s.project("main", false)
		s.commit("Plan first", map[string]string{
			"tend/flow.yaml": buildFlow(s.t, "  start: implement\n", "  start: plan\n", "  steps:\n", `  steps:
    plan:
      agent: claude
      prompt: prompts/plan.md
      result: commit
      next:
        ok: implement
`, "    implement:\n      agent: claude\n", "    implement:\n      continue: plan\n"),
			"tend/prompts/plan.md": "Plan it.\n",
		})
		r := s.run("IMPL_FAIL=timeout")
		s.check("an implement prompt that timed out after the plan stops", r.code == 3)
		s.agent("working", "impl")
		r = s.run()
		s.check("the rerun waits for the implementer and passes", r.passed())
		s.check("asking the plan once in all", s.prompts("Plan it.") == 1 && s.prompts("/grove-implement") == 1)
	}},
}

// Whether every match of re's first group in text names the same agent.
func sameAgent(text, re string) bool {
	m := regexp.MustCompile(re).FindAllStringSubmatch(text, -1)
	for _, x := range m {
		if x[1] != m[0][1] {
			return false
		}
	}
	return len(m) > 0
}

func TestFlows(t *testing.T) {
	t.Parallel()
	for _, x := range flowScenarios {
		t.Run(x.name, func(t *testing.T) { x.fn(newSc(t)) })
	}
}
