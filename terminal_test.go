package main

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
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

// ---- telling a person

type note struct{ Title, Body, Sound string }

// The notifications the Herdr stand-in was asked to show, oldest first.
func (s *sc) notes() []note {
	var ns []note
	for _, l := range strings.Split(s.cat(s.st+"/notifications"), "\n") {
		var n note
		if json.Unmarshal([]byte(l), &n) == nil {
			ns = append(ns, n)
		}
	}
	return ns
}

func (s *sc) lastNote() note {
	ns := s.notes()
	if len(ns) == 0 {
		return note{}
	}
	return ns[len(ns)-1]
}

// The pull request's URL as the gh stand-in gives it.
func (s *sc) prURL() string { return "https://example.test/pr/" + strings.ReplaceAll(s.branch(), "/", "_") }

// Whether n's body ends with link alone on its last line, and both fit a
// lock screen.
func (n note) endsWith(link string) bool {
	return strings.HasSuffix(n.Body, "\n"+link) && utf8.RuneCountInString(n.Title) <= 60 && utf8.RuneCountInString(n.Body) <= 240
}

var notifyScenarios = []scenario{
	{"notify-stop", func(s *sc) {
		s.plain()
		s.allow()
		s.issues("7:tend")
		r := s.build()
		n := s.lastNote()
		s.check("a stop still exits 3", r.code == 3)
		s.check("titled with the issue and that it waits", n.Title == "tend: o/r#7 waits for you")
		s.check("its body says why and ends with the issue's URL", strings.Contains(n.Body, "not labelled tend-build") && n.endsWith("https://example.test/issues/7"))
		s.check("sounded as a request", n.Sound == "request")
	}},
	{"notify-landed", func(s *sc) {
		s.lands(false, "check success")
		r := s.run()
		n := s.lastNote()
		s.check("a landing is notified as landed", r.passed() && n.Title == "tend: o/r#1 landed")
		s.check("waiting for no one, with the pull request's URL", !strings.Contains(n.Title+n.Body, "waits for you") && n.endsWith(s.prURL()))
	}},
	{"notify-supervised", func(s *sc) {
		s.lands(true, "check success")
		r := s.run()
		n := s.lastNote()
		s.check("a supervised head that passed review waits for the person", r.passed() && n.Title == "tend: o/r#1 waits for you")
		s.check("its body says review passed at the head, to approve to land or request changes", strings.Contains(n.Body, "Review passed at "+short(s.head())) && strings.Contains(n.Body, "Approve it to land it, or request changes"))
		s.check("ending with the pull request's URL", n.endsWith(s.prURL()))
	}},
	{"notify-built-in", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run()
		n := s.lastNote()
		s.check("under the built-in flow, a head that passed review is to merge or request changes", r.passed() && strings.Contains(n.Body, "Merge it, or request changes"))
		s.check("saying nothing of approving", !strings.Contains(strings.ToLower(n.Body), "approv") && n.endsWith(s.prURL()))
	}},
	{"notify-repair-spent", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run("REVIEW=actionable")
		n := s.lastNote()
		s.check("a head still failing after its one repair is published and waits", r.code == 0 && r.conclusion() == "failure" && n.Title == "tend: o/r#1 waits for you")
		s.check("its body says Codex's findings stand after one repair", strings.Contains(n.Body, "Codex's findings stand at "+short(s.head())+" after one repair"))
		s.check("to request changes to send them to the implementer, or close", strings.Contains(n.Body, "Request changes to send them to the implementer, or close") && n.endsWith(s.prURL()))
	}},
	{"notify-decision", func(s *sc) {
		s.project("main", true)
		s.allow()
		r := s.run("REVIEW=decision")
		n := s.lastNote()
		s.check("a failure not judged repairable says Codex's findings stand", r.code == 0 && strings.Contains(n.Body, "Codex's findings stand at "+short(s.head())+"."))
		s.check("to request changes or close, with no repair named", strings.Contains(n.Body, "Request changes to send them to the implementer, or close") && !strings.Contains(n.Body, "repair") && n.endsWith(s.prURL()))
	}},
	{"notify-unpublished", func(s *sc) {
		s.plain()
		s.issues("7:tend-build")
		r := s.build()
		n := s.lastNote()
		h := s.git(s.bworktree(), "rev-parse", "--short=7", "HEAD")
		s.check("with no pull request the notification names the issue", r.passed() && n.Title == "tend: o/r#7 waits for you")
		s.check("its body names the branch and its head", strings.Contains(n.Body, "tend/build-7 at "+h))
		s.check("and ends with the issue's URL", n.endsWith("https://example.test/issues/7"))
	}},
	{"notify-held", func(s *sc) {
		s.lands(false, "check success")
		r := s.run(`BUILD_ALSO=echo "# ours" >>tend/flow.yaml`)
		n := s.lastNote()
		s.check("a head held at land waits for the person", r.held("tend/flow.yaml") && n.Title == "tend: o/r#1 waits for you")
		s.check("its body names the protected path and to approve this head", strings.Contains(n.Body, "tend/flow.yaml") && strings.Contains(n.Body, "Approve this head to land it"))
		s.check("ending with the pull request's URL, given once", n.endsWith(s.prURL()) && strings.Count(n.Body, s.prURL()) == 1)
	}},
	{"notify-still-red", func(s *sc) {
		s.stillRed()
		n := s.lastNote()
		s.check("checks still failing after the one fix wait for the person", n.Title == "tend: o/r#1 waits for you")
		s.check("naming the check and that it failed after the one fix", strings.Contains(n.Body, "Checks test failed at "+short(s.head())+" after the one fix"))
		s.check("to fix the branch, run the checks again or close", strings.Contains(n.Body, "Fix the branch, run the checks again, or close the pull request"))
		s.check("saying nothing of approving", !strings.Contains(strings.ToLower(n.Body), "approv") && n.endsWith(s.prURL()))
	}},
	{"notify-other-stop", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.request()
		s.run("REQUEST_FAIL=agent_prompt_stalled")
		r := s.run()
		n := s.lastNote()
		s.check("a stop after publishing waits for the person on the pull request", r.code == 3 && n.Title == "tend: o/r#1 waits for you")
		s.check("its body gives the reason and ends with the pull request's URL", strings.Contains(n.Body, "committed nothing") && n.endsWith(s.prURL()))
		s.check("and still records needs-you", s.has(s.status(), `^[^ ]* needs-you: `))
		s.check("adding nothing about running tend again but the reason's own words", !strings.Contains(n.Body, "a person deals with that"))
	}},
	{"notify-once", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.check("an ended attempt is notified once", len(s.notes()) == 1)
		s.write(s.runDir()+"/status", "2026-11-25T00:00:00+0000 working: reviewing\nnext: tend goes on\n")
		w := s.watch()
		s.check("watch runs again an attempt a dead tend left working", s.count(w.err, `running tend`) == 1)
		s.check("and the same end is not notified again", len(s.notes()) == 1)
		first := s.head()
		s.request()
		r := s.run()
		n := s.lastNote()
		s.check("a new head that passes review is notified", r.passed() && s.head() != first && len(s.notes()) == 2 && strings.Contains(n.Body, short(s.head())))
	}},
	{"notify-long", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.run()
		s.request()
		long := strings.Repeat("why ", 150)
		r := s.run("REQUEST_FAIL=" + long)
		n := s.lastNote()
		s.check("a stop with a 600-character reason exits 3", r.code == 3 && len(long) == 600)
		s.check("its body is at most 240 characters and ends with the pull request's whole URL", n.endsWith(s.prURL()))
		s.check("the reason cut with an ellipsis before the link", strings.HasSuffix(n.Body, "…\n"+s.prURL()))
	}},
}

// Under rex tend writes its state to /dev/tty, which a test of the built
// binary cannot read: the message it reports there is checked here.
func TestRexStopMessage(t *testing.T) {
	pr, cycle, run, repo = "", "issue-7", "", ""
	issue = parse(`{"url": "https://github.com/o/r/issues/7"}`)
	item = request{from: "#7", key: "7"}
	title, body := waits("issue #7 is not labelled tend-build, so it is not approved for building")
	st, msg := rexMessage(title, body, "request")
	if st != "blocked" {
		t.Errorf("state %q, want blocked", st)
	}
	if !strings.HasPrefix(msg, "tend: o/r#7 waits for you") {
		t.Errorf("message %q does not begin with the title", msg)
	}
	if !strings.Contains(msg, "https://github.com/o/r/issues/7") || strings.ContainsAny(msg, "\n") {
		t.Errorf("message %q lacks the issue's URL or holds a newline", msg)
	}
}
