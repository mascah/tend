package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ---- decisions: what tend asks the owner, on GitHub
//
// When a build attempt stops for a person and its origin is published, the
// publishing account opens an issue on the repository saying what happened,
// what the agent said last, where the work is and how to answer. Closing the
// issue is the answer, which tend watch reads with who gave it: closed as
// completed, carry on, asking the agent that left nothing once more; closed
// as not planned, leave the attempt as it is. The account the agents share
// cannot answer. A rerun that stops with the same question while its issue is
// open asks nothing new. decision in the run folder holds the issue, the
// record of the ask to make once more, and the question.

// The ask that stopped, where an agent was asked and left nothing: its agent,
// the record holding the ask, and what happened.
var stuckOn struct{ agent, rec, what string }

// Stop where an agent was asked and left nothing: what happened, then what a
// person at this machine can do. Carrying on asks it once more.
func stuck(agent, rec, what, here string) {
	stuckOn.agent, stuckOn.rec, stuckOn.what = agent, rec, what
	stop("%s. %s", what, here)
}

// The issue that asks the owner, opened unless the same question is open
// already: its URL, or "" where nothing was asked on GitHub.
func tell(msg string) string {
	if run == "" || filepath.Dir(run) != state+"/runs" {
		return ""
	}
	if account == "" {
		account = accountOf()
	}
	if account == "" {
		return ""
	}
	q := msg
	if stuckOn.what != "" {
		q = stuckOn.what
	}
	f := run + "/decision"
	if d := strings.SplitN(cat(f), "\n", 3); len(d) == 3 && d[2] == q {
		if s, err := quiet("gh", "api", issueAPI(d[0]), "--jq", ".state"); err == nil && s == "open" {
			say("asked already on %s", d[0])
			return d[0]
		}
	}
	token()
	u, err := input(scrub(question(q)), "gh", "issue", "create", "--title", "tend needs you on "+cycle+": "+clip(scrub(q), 90), "--body-file", "-")
	if ls := strings.Split(u, "\n"); err == nil {
		u = ls[len(ls)-1]
	}
	if err != nil || u == "" {
		// ponytail: not asked again until the next stop; a pass of watch could retry it.
		say("could not open an issue asking about this; tend status has it")
		return ""
	}
	write(f, u+"\n"+stuckOn.rec+"\n"+q)
	say("asked on %s", u)
	return u
}

// The issue's body: the question and what it takes to answer it.
func question(q string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "tend stopped building `%s` (attempt `%s`, branch `%s`) and needs a decision from you.\n\n%s\n\n", cycle, id, branch, quote(q))
	if a := stuckOn.agent; a != "" {
		if s := strings.TrimSpace(term.read(a)); s == "" {
		} else if v, err := quiet("gh", "repo", "view", "--json", "visibility", "--jq", ".visibility"); err == nil && v == "PRIVATE" {
			fmt.Fprintf(&b, "**What %s said last:**\n\n%s\n\n", a, quote(clip(s, 4000)))
		} else { // a public repository's issues are anyone's: the agent's words stay on the runner
			fmt.Fprintf(&b, "What %s said last is on the runner, in `$TEND_STATE/agents/%s/`, and is not posted to this public repository.\n\n", a, a)
		}
	}
	if dir != "" && exists(dir) {
		log := gitQ("-C", dir, "log", "--oneline", base+"..HEAD")
		if log == "" {
			log = "no commits"
		}
		fmt.Fprintf(&b, "**Where the work is:** branch `%s` on the runner (`tend status` there names its worktree), at `%s`. Since `%s`:\n\n```text\n%s\n%s\n```\n\n",
			branch, gitQ("-C", dir, "rev-parse", "--short", "HEAD"), base, log, gitQ("-C", dir, "diff", "--shortstat", base+"...HEAD"))
	}
	if p := cat(run + "/pr"); p != "" {
		fmt.Fprintf(&b, "Its pull request is %s.\n\n", p)
	}
	again := ""
	if stuckOn.rec != "" {
		again = ", asking " + stuckOn.agent + " once more"
	}
	fmt.Fprintf(&b, "**To answer, close this issue**\n\n- as **completed** for tend to carry on%s. Where the question needs something done first, do it in the branch's worktree on the runner, then close it.\n- as **not planned** for tend to leave the attempt as it is.\n\ntend watch reads the answer on its next pass.\n\n<!-- tend:decision %s -->\n", again, id)
	return b.String()
}

// The answer to run folder d's decision, in repository r, once its issue is
// closed by anyone but the account the agents share: carrying on removes the
// record of the ask to make once more and gives why tend runs again; leaving
// it says so in the status and gives "". "" while it is not answered.
func answered(d, r string) string {
	ls := strings.SplitN(cat(d+"/decision"), "\n", 3)
	if ls[0] == "" {
		return ""
	}
	s, err := quiet("gh", "api", issueAPI(ls[0]), "--jq", `[.state, .state_reason // "", .closed_by.login // ""] | join(" ")`)
	f := strings.Split(s, " ")
	if err != nil || len(f) != 3 || f[0] != "closed" {
		return ""
	}
	if f[2] == accountIn(r) {
		say("%s was closed by %s, the account the agents share, so tend takes no answer from it", ls[0], f[2])
		return ""
	}
	os.Remove(d + "/decision")
	if f[1] == "not_planned" {
		stageIn(d, "done", f[2]+" closed "+ls[0]+" as not planned, so tend leaves the attempt as it is", "nothing, unless a person runs tend on it")
		return ""
	}
	if len(ls) > 1 && ls[1] != "" {
		os.Remove(ls[1])
	}
	return f[2] + " answered " + ls[0] + ": carry on"
}

// An attempt that carried on past its decision, as when a person ran tend on
// it by hand, closes the issue still asking, saying how it stands.
func settledDecision(how string) {
	ls := strings.SplitN(cat(run+"/decision"), "\n", 3)
	if ls[0] == "" || account == "" {
		return
	}
	if _, err := quiet("gh", "issue", "close", ls[0], "--comment", "tend carried on: "+how+"."); err == nil {
		os.Remove(run + "/decision")
	}
}

// https://github.com/<owner>/<repo>/issues/<n> as GitHub's API names it.
func issueAPI(u string) string {
	p := strings.Split(strings.TrimSuffix(u, "/"), "/")
	if len(p) < 4 {
		return u
	}
	return "repos/" + strings.Join(p[len(p)-4:], "/")
}

func quote(s string) string { return "> " + strings.ReplaceAll(s, "\n", "\n> ") }

func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

// Text for an issue, with the runner's own paths named as TEND_STATE and ~
// rather than spelled out.
func scrub(s string) string {
	s = strings.ReplaceAll(s, state, "$TEND_STATE")
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		s = strings.ReplaceAll(s, h, "~")
	}
	return s
}
