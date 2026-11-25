package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var landed string // the pull request this run found merged at land

// The verdict file a step names for head h, in the run folder.
func verdictName(name, h string) string { return fill(strings.ReplaceAll(name, "{head}", h)) }

// Whether head h passed review: each verdict the flow's steps wrote for it is
// valid and a success, and there is one.
func passed(h string) bool {
	n := 0
	for _, s := range fl.Steps {
		r, _ := parseResult(s.Result)
		if f := run + "/" + verdictName(r.verdict, h); r.verdict != "" && exists(f) {
			if !valid(f, h) || get(parse(cat(f)), "conclusion") != "success" {
				return false
			}
			n++
		}
	}
	return n > 0
}

// The heads h is carried through, h first and the head that passed review
// last, when h is a merge commit whose first parent passed review or is
// itself carried, whose second parent is a commit of the target, and whose
// tree is exactly Git's merge of the two; or nil. Read from Git, never from a
// mark an agent could write.
func carried(h string) []string {
	p := strings.Fields(gitQ("rev-list", "--parents", "-n", "1", h))
	if len(p) != 3 || !ok("git", "merge-base", "--is-ancestor", p[2], base) {
		return nil
	}
	t, err := quiet("git", "merge-tree", "--write-tree", p[1], p[2])
	if err != nil || strings.SplitN(t, "\n", 2)[0] != gitQ("rev-parse", h+"^{tree}") {
		return nil
	}
	if passed(p[1]) {
		return []string{h, p[1]}
	}
	if c := carried(p[1]); c != nil {
		return append([]string{h}, c...)
	}
	return nil
}

// The tend/review status on carried head h, set once: success, carried from
// the head that passed review. No review is posted and no reviewer asked.
func carry(h string, c []string) {
	mark := run + "/carried-" + h
	if exists(mark) {
		return
	}
	from := c[len(c)-1]
	if _, err := out("gh", "api", "repos/{owner}/{repo}/statuses/"+h, "-f", "state=success", "-f", "context=tend/review",
		"-f", "description=Carried from "+short(from)+", which passed independent review, by a clean merge of "+target, "-f", "target_url="+pr); err != nil {
		stop("could not set the tend/review status on %s", short(h))
	}
	write(mark, from+"\n")
}

// The paths head h changes: those that differ between h and the commit
// where h and the target's tip as fetched meet, a renamed file under both
// its names. Read from Git alone.
func changedPaths(h string) []string {
	s := git("diff", "--no-renames", "--name-only", "-z", base+"..."+h)
	return strings.FieldsFunc(s, func(r rune) bool { return r == 0 })
}

// A head whose change touches a protected path of the target as now fetched
// lands only once a person with write access approves it, or a head it is
// carried from; otherwise tend stops for one, naming the paths. Decided from
// Git and GitHub's reviews alone, never from a verdict or a file an agent writes.
func hold(h string) {
	ps, _ := protectedAt(base)
	touched := covered(ps, changedPaths(h))
	if len(touched) == 0 {
		return
	}
	head = h
	if approved() {
		return
	}
	what := "the protected path "
	if len(touched) > 1 {
		what = "the protected paths "
	}
	stop("%s at %s changes %s%s, so tend lands it only once a person with write access approves that head on %s", branch, short(h), what, and(touched), pr)
}

// One landing at a time for each repository on this runner: the lock, held
// until it is closed or tend exits.
func landingLock() *os.File {
	f := lockFile(repo, "land")
	l, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		stop("could not take the landing lock %s", f)
	}
	if syscall.Flock(int(l.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil {
		stage("working", fmt.Sprintf("%s waits to land: another tend lands a candidate of %s", pr, repo), "tend lands it once that one is done")
		say("waiting to land %s: another tend lands a candidate of %s", pr, repo)
		for err = syscall.EINTR; err == syscall.EINTR; {
			err = syscall.Flock(int(l.Fd()), syscall.LOCK_EX)
		}
		if err != nil {
			stop("could not take the landing lock %s", f)
		}
	}
	return l
}

// The checks GitHub reports on commit h, check runs and commit statuses
// alike: each its name, pass, fail or pending, and its link.
func checksOn(h string) [][]string {
	runs, err := out("gh", "api", "repos/{owner}/{repo}/commits/"+h+"/check-runs", "--paginate", "--jq",
		`.check_runs[] | [.name, (if .status != "completed" then "pending" elif (.conclusion | IN("success", "neutral", "skipped")) then "pass" else "fail" end), .html_url // ""] | @tsv`)
	if err != nil {
		stop("could not read the check runs on %s", short(h))
	}
	sts, err := out("gh", "api", "repos/{owner}/{repo}/commits/"+h+"/status", "--paginate", "--jq",
		`.statuses[] | [.context, (if .state == "pending" then "pending" elif .state == "success" then "pass" else "fail" end), .target_url // ""] | @tsv`)
	if err != nil {
		stop("could not read the statuses on %s", short(h))
	}
	var cs [][]string
	for _, l := range strings.Split(runs+"\n"+sts, "\n") {
		if f := strings.Split(l, "\t"); len(f) == 3 {
			cs = append(cs, f)
		}
	}
	return cs
}

// How the checks on head h stand: failed, with the failed checks' names and
// links in {checks}; passed; or "" while one is unfinished or none but
// tend/review is reported, which stops for a person once the wait is over.
func checked(h string, over bool) string {
	var failed, pending []string
	ci := 0
	for _, c := range checksOn(h) {
		if c[0] != "tend/review" {
			ci++
		}
		switch c[1] {
		case "fail":
			failed = append(failed, c[0]+" "+c[2])
		case "pending":
			pending = append(pending, c[0])
		}
	}
	switch {
	case len(failed) > 0:
		vars["checks"] = "The failed checks, each with its link: " + strings.Join(failed, "; ") + "."
		say("checks failed on %s: %s", short(h), strings.Join(failed, "; "))
		return "failed"
	case ci > 0 && len(pending) == 0:
		return "passed"
	case over && ci == 0:
		stop("%s at %s has no check but tend/review, so tend lands nothing: a project lands only with CI on its pull requests", branch, short(h))
	case over:
		stop("the checks %s on %s at %s were still unfinished after %sms, so tend landed nothing", and(pending), branch, short(h), timeout())
	}
	return ""
}

// The build flow's land: a head that passed review, or is carried from one
// that did, with its pull request open, landed one at a time, held for a
// person's approval while its change touches a protected path: the target
// merged in when it moved, never by a forced push, every check on the head
// waited for, then the pull request squash-merged at exactly that head by
// the publishing account, its title and number the subject and its
// description the body.
func landStep() string {
	if pr == "" {
		stop("there is no pull request to land: the repository's origin is not listed in %s/publish", state)
	}
	h := git("rev-parse", "HEAD")
	if !passed(h) && carried(h) == nil {
		stop("%s at %s did not pass review, so tend lands nothing", branch, short(h))
	}
	token()
	l := landingLock()
	defer l.Close()
	ms, _ := strconv.Atoi(timeout())
	wait := time.Duration(ms) * time.Millisecond
	poll := min(20*time.Second, wait/4)
	until := time.Now().Add(wait)
	held := false // whether the head was held to its protected paths yet
	// Each round reads the pull request, the head and the target afresh, so a
	// merge by a person or a target that moved is seen while checks run.
	for round := 0; ; round++ {
		if round > 0 {
			time.Sleep(poll)
		}
		s, err := out("gh", "pr", "view", pr, "--json", "state,headRefOid,baseRefName,number,title,body")
		if err != nil {
			stop("could not read %s", pr)
		}
		x := parse(s)
		switch str(get(x, "state")) {
		case "MERGED":
			landed = pr
			say("%s is merged", pr)
			return "ok"
		case "OPEN":
		default:
			stop("%s is closed, so tend lands nothing", pr)
		}
		if b := str(get(x, "baseRefName")); b != target {
			stop("%s merges into %s, not %s, so tend lands nothing", pr, b, target)
		}
		h = git("rev-parse", "HEAD")
		if o := str(get(x, "headRefOid")); o != h {
			stop("%s's head is %s, not %s's %s, so tend lands nothing", pr, short(o), branch, short(h))
		}
		fetchTarget()
		moved := !ok("git", "merge-base", "--is-ancestor", base, "HEAD")
		if moved || !held {
			hold(h)
			held = true
		}
		if moved {
			stage("working", fmt.Sprintf("merging %s into %s", base, branch), "tend waits for the merged head's checks, then lands it")
			if _, err := both("git", "merge", "--no-ff", "--no-edit", "-m", "Merge "+target+" into "+branch, base); err != nil {
				quiet("git", "merge", "--abort")
				say("%s does not merge cleanly into %s; merged nothing", base, branch)
				return "conflict"
			}
			if !pushAs(branch) {
				stop("could not push %s to origin as %s; if it has commits this checkout lacks, look at them, then run tend again", branch, account)
			}
			h = git("rev-parse", "HEAD")
			if c := carried(h); c != nil {
				carry(h, c)
			}
			say("merged %s into %s at %s", base, branch, short(h))
			until = time.Now().Add(wait)
			round = -1 // the new head's checks are read at once
			continue
		}
		switch checked(h, time.Now().After(until)) {
		case "failed":
			return "failed"
		case "":
			stage("working", fmt.Sprintf("waiting for the checks on %s at %s", branch, short(h)), "tend squash-merges "+pr+" once every check passes")
			continue
		}
		hold(h)
		stage("working", fmt.Sprintf("landing %s at %s", pr, short(h)), "GitHub squash-merges it")
		// GitHub's merge itself, never gh pr merge, which may queue the pull
		// request or defer the merge and refuses a blocked one in words of its own.
		req, _ := json.Marshal(map[string]any{"sha": h, "merge_method": "squash",
			"commit_title": fmt.Sprintf("%s (#%s)", str(get(x, "title")), str(get(x, "number"))), "commit_message": str(get(x, "body"))})
		var refused bytes.Buffer
		in := string(req)
		m, err := sh(&in, &refused, false, "gh", "api", "-X", "PUT", "repos/{owner}/{repo}/pulls/"+str(get(x, "number"))+"/merge", "--input", "-", "--jq", ".merged")
		if err == nil && m == "true" {
			landed = pr
			say("landed %s at %s", pr, short(h))
			return "ok"
		}
		why := strings.Join(strings.Fields(refused.String()), " ")
		if err == nil {
			stop("GitHub did not merge %s: it answered merged %s", pr, m)
		}
		fetchTarget()
		switch {
		case !ok("git", "merge-base", "--is-ancestor", base, "HEAD"):
			say("GitHub refused to merge %s, as %s moved: %s", pr, target, why)
			round = -1 // merged in at once
		case regexp.MustCompile(`(?i)required status check.*(expected|pending|in progress)`).MatchString(why) && time.Now().Before(until):
			say("GitHub refused to merge %s for a check it still waits for: %s", pr, why)
		default:
			stop("GitHub refused to merge %s: %s", pr, why)
		}
	}
}
