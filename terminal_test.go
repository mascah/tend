package main

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var terminalScenarios = []scenario{
	{"rex", func(s *sc) {
		s.project("main", false)
		r := s.rexrun()
		s.check("under rex, a cycle is claimed, built and reviewed", r.passed())
		s.check("its worktree is made by Git under the state folder", sameDir(filepath.Dir(s.rexwt()), s.ts+"/worktrees"))
		sessions := strings.Fields(s.cat(s.st + "/rex-sessions"))
		s.check("and opened as a session of its own", len(sessions) == 2 && sessions[1] == filepath.Base(s.rexwt()))
		s.check("each agent runs in a block named for it", strings.Contains(s.cat(s.st+"/rex-blocks"), " impl-"))
		s.check("the reviewer's block is closed after its verdict", s.count(s.cat(s.st+"/rex-blocks"), ` rev-`) == 0)
		s.check("prompts go as a paste", s.has(s.rexlog(), "^rex send .*\x1b\\[200~/grove-implement"))
		s.check("nothing was asked of Herdr", s.count(s.log(), `^worktree|^pane|^agent get`) == 0)
		s.write(s.st+"/rex-sessions", regexp.MustCompile(`(?m) .*$`).ReplaceAllString(s.cat(s.st+"/rex-sessions"), " renamed-by-an-agent")+"\n")
		s.rexrun()
		s.check("a rerun finds its session, though an agent renamed it, and reviews nothing again", len(strings.Split(s.cat(s.st+"/rex-sessions"), "\n")) == 1 && s.reviews() == 1)
	}},
	{"rex-repair", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("REVIEW=repairable")
		s.check("under rex, a failed review is repaired once and passes", r.passed())
		s.check("the repair was asked once", s.prompts("An") == 1)
	}},
	{"rex-stalled", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("REX_STALL=1", "TIMEOUT_MS=1500")
		s.check("under rex, a prompt the agent never takes up stops, after the prompt's timeout", r.code == 3)
		s.check("as stalled", r.said("agent_prompt_stalled"))
		s.agent("working", "impl")
		r = s.rexrun()
		s.check("the rerun waits for the agent once it is working", r.code == 0 && r.conclusion() == "success")
		s.check("and does not prompt it again", s.prompts("/grove-implement") == 1)
		s.check("and says so", r.said("still working on an earlier prompt"))
	}},
	{"rex-gone-mid-turn", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("REX_VANISH=impl")
		s.check("under rex, an implementer whose block closes after its prompt went stops", r.code == 3 && r.said("agent_gone_after_delivery"))
		s.check("its ask stays recorded, as delivered", s.cat(s.runDir()+"/implement-asked") != "")
		r = s.rexrun("REX_VANISH=impl")
		s.check("the rerun asks it once more in a new session, which goes too", r.code == 3 && s.prompts("/grove-implement") == 2)
		s.check("and that second ask stays recorded", s.cat(s.runDir()+"/again-implement-asked") != "")
		r = s.rexrun()
		s.check("gone a second time, tend stops and asks no third time", r.code == 3 && s.prompts("/grove-implement") == 2)
	}},
	{"rex-dialog", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("REVIEW=repairable", "BLOCK_AFTER_BUILD=1")
		s.check("under rex, an agent reporting blocked stops before the repair", r.code == 3)
		s.check("saying it is blocked", r.said("impl-.* is blocked"))
		s.agent("none", "impl")
		r = s.rexrun("REVIEW=repairable")
		s.check("an agent reporting nothing, as at a dialog, is not prompted", r.code == 3)
		s.check("it is unknown, a person's", r.said("is unknown; look at it"))
		s.check("and the repair is not used", !exists(s.runDir()+"/repair-asked"))
		s.agent("idle", "impl")
		r = s.rexrun("REVIEW=repairable")
		s.check("once it reports idle, the repair is asked and passes", r.passed())
	}},
	{"rex-trust", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1")
		s.check("under rex, the trust dialog of a new folder is answered", r.passed())
		s.check("by choosing to trust it", s.has(s.rexlog(), `^rex send-key .* down enter`))
	}},
	{"rex-add-dir", func(s *sc) {
		s.project("main", false)
		began := time.Now()
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Yes, add it")
		s.check("Claude's dialog for adding the run folder is answered", r.passed())
		s.check("by choosing to add it", s.has(s.rexlog(), `^rex send-key .* down enter`))
		s.check("without waiting out the agents' first reports", time.Since(began) < 60*time.Second)
	}},
	{"rex-add-dir-refused", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Yes, add it", "TRUST_REFUSE=1")
		s.check("a dialog that refuses tend's first answer is answered again", r.passed())
		seen, twice := map[string]bool{}, false
		re := regexp.MustCompile(`^rex send-key .* down enter`)
		for _, l := range strings.Split(s.rexlog(), "\n") {
			if re.MatchString(l) {
				twice = twice || seen[l]
				seen[l] = true
			}
		}
		s.check("twice", twice)
	}},
	{"rex-hooks", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Hooks can run outside the sandbox after you trust them.")
		s.check("an agent asking to review new hooks stops for a person", r.code == 3)
		s.check("saying so", r.said("asks to review new or changed hooks"))
		s.check("and is sent no key", !s.has(s.rexlog(), `^rex send-key`))
	}},
	{"rex-hooks-wrapped", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Hooks can run outside the\nsandbox after you trust them.\n  enter continue · esc\nskip")
		s.check("a hooks review wrapped in a narrow pane still stops for a person", r.code == 3)
		s.check("saying so", r.said("asks to review new or changed hooks"))
		s.check("and is sent no key, not even Esc", !s.has(s.rexlog(), `^rex send-key`))
	}},
	{"rex-notice", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("NOTICE=1", "TRUST_DIALOG=1")
		s.check("under rex, a startup notice offering esc skip is skipped", r.passed())
		first := map[string]string{} // each block's first key
		re := regexp.MustCompile(`^rex send-key -s \S+ -b (\S+) (.*)$`)
		for _, l := range strings.Split(s.rexlog(), "\n") {
			if m := re.FindStringSubmatch(l); m != nil && first[m[1]] == "" {
				first[m[1]] = m[2]
			}
		}
		esc := 0
		for _, k := range first {
			if k == "esc" {
				esc++
			}
		}
		s.check("by Esc alone, the first key each agent is sent", esc > 0 && esc == len(first))
		s.check("saying so for each agent", s.count(r.err, `skipped .*'s startup notice`) == esc)
	}},
	{"rex-unknown-screen", func(s *sc) {
		s.project("main", false)
		r := s.rexrun("TRUST_DIALOG=1", "TRUST_TEXT=Choose a theme", "TIMEOUT_MS=1000") // its first report waited for that long
		s.check("under rex, an agent at a screen tend does not know stops for a person", r.code == 3)
		s.check("saying it did not start", r.said("did not start"))
		s.check("and is sent no key", !s.has(s.rexlog(), `^rex send-key`))
	}},
}
