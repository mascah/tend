package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The account a repository's origin is published as, from $state/publish
// ("<origin-url> <account>" a line), or nothing when it is not listed.
func accountOf() string {
	o := gitQ("remote", "get-url", "origin")
	for _, line := range strings.Split(cat(state+"/publish"), "\n") {
		if f := strings.Fields(line); len(f) > 1 && f[0] == o {
			return f[1]
		}
	}
	return ""
}

func token() {
	t, err := out("gh", "auth", "token", "--user", account)
	if err != nil {
		stop("gh holds no token for %s; sign it in with gh auth login, then switch back", account)
	}
	os.Setenv("GH_TOKEN", t)
}

// Push HEAD to origin's branch as the account whose token is GH_TOKEN, over
// HTTPS even when origin is an SSH URL; never forced.
func pushAs(b string) bool {
	_, err := out("git", "-c", "credential.helper=", "-c", "credential.helper=!gh auth git-credential", "-c", "url.https://github.com/.pushInsteadOf=git@github.com:",
		"push", "-q", "origin", "HEAD:refs/heads/"+b)
	return err == nil
}

// The build flow's publish: the head published, then whether a person
// requested changes on it that the flow's request step should make. Each
// request is asked of the implementer once: request-<review id> is written
// just before the prompt goes; a request on an older head was answered by a
// later commit. In a flow whose publish gives a step for failed, the checks
// on the head are waited for next, before an approval is read: a failed one
// goes to that step while it can be asked, and is said once on the pull
// request when it cannot. tend does not wait here for people: tend watch
// runs it again when one requests changes.
var (
	told        bool
	spentChecks []string // the checks that still failed after the one fix, by name
)

func publishStep() string {
	spentChecks = nil
	where, account = branch, accountOf()
	if account == "" && !told {
		told = true
		if o, err := quiet("git", "remote", "get-url", "origin"); err == nil {
			say("not publishing: %s is not listed in %s/publish", o, state)
		}
	}
	publish()
	if pr == "" {
		return "ok"
	}
	req, err := requested()
	if err != nil {
		stop("could not read the reviews of %s", pr)
	}
	if req == nil || exists(run+"/answered-"+str(get(req, "node_id"))) {
		if to := fl.Steps["publish"].Next["failed"]; to != "" {
			if failed := waitChecks(); failed != nil {
				if canAsk(to) {
					return "failed"
				}
				spentFix(to, failed)
				for _, f := range failed { // "<name> <link>"
					spentChecks = append(spentChecks, strings.TrimSpace(f[:max(0, strings.LastIndex(f, " "))]))
				}
			}
		}
		if fl.Steps["publish"].Next["approved"] != "" && approved() {
			return "approved"
		}
		return "ok"
	}
	rid := str(get(req, "id"))
	comments, err := out("gh", "api", reviewsAPI()+"/"+rid+"/comments", "--paginate", "--jq", `.[] | "\(.path):\(.line // .original_line // "?"): \(.body)"`)
	if err != nil {
		stop("could not read the comments of review %s on %s", rid, pr)
	}
	if comments != "" {
		comments = " Their comments on lines: " + comments
	}
	reviewReq = req
	vars["who"], vars["review_id"] = str(get(req, "user", "login")), rid
	vars["review"], vars["comments"] = strings.TrimRight(str(get(req, "body")), "\n"), comments
	return "requested"
}

// The checks on the published head, waited for in a flow whose publish
// gives a step for failed, before an approval is read or a person asked:
// read as land reads them, tend/review left out, so a head that failed
// review has no failed check for it. The failed checks, each its name and
// link, or nil once every check passed, decided only once every check has
// finished; with none but tend/review, or one unfinished, when the wait
// ends, tend stops for a person. A read that ends after the wait does not
// decide: the checks were not all seen finished within it. The deadline is
// checked before each sleep, which is no longer than the time left.
func waitChecks() []string {
	ms, _ := strconv.Atoi(timeout())
	wait := time.Duration(ms) * time.Millisecond
	poll := min(20*time.Second, wait/4)
	until := time.Now().Add(wait)
	var pending []string // as the last read within the wait saw them
	ci := 0
	over := func() {
		switch {
		case ci == 0:
			stop("%s at %s has no check but tend/review, so tend asks no one to approve it: a flow whose publish waits for checks needs CI on its pull requests", branch, short(head))
		case len(pending) == 0:
			stop("the checks on %s at %s were not read finished within %sms, so tend asks no one to approve it yet", branch, short(head), timeout())
		}
		stop("the checks %s on %s at %s were still unfinished after %sms, so tend asks no one to approve it yet", and(pending), branch, short(head), timeout())
	}
	for round := 0; ; round++ {
		if round > 0 {
			left := time.Until(until)
			if left <= 0 {
				over()
			}
			time.Sleep(min(poll, left))
		}
		f, p, c := readChecks(head, true)
		if time.Now().After(until) {
			if round == 0 { // no read ended within the wait: what this one saw
				pending, ci = p, c
			}
			over()
		}
		pending, ci = p, c
		switch standing(head, f, p, c) {
		case "failed":
			return f
		case "passed":
			say("checks passed on %s", short(head))
			return nil
		}
		stage("working", fmt.Sprintf("waiting for the checks on %s at %s", branch, short(head)), "tend asks a person about "+pr+" once they pass, or has a failed one fixed once")
	}
}

// Whether the step a failure leads to can still be asked: it has no limit,
// it has not reached it, or it was asked at this head and its agent is
// still working on it, or is gone and may be asked once more in a new
// session. Asked at this head and settled with no commit, it is spent.
func canAsk(name string) bool {
	s := fl.Steps[name]
	if s == nil || s.Limit == 0 {
		return true
	}
	rec := recordOf(name)
	if timesAsked(name, rec) < s.Limit {
		return true
	}
	if !slices.Contains(asks(rec), head) {
		return false
	}
	agent, _, _ := agentOf(name)
	return term.status(agent) == "working" || term.gone(agent) && !slices.Contains(asks(againOf(rec)), head)
}

// Failed checks on the head once the step for them is spent: said once on
// the pull request, found by the head's marker, so a rerun says it no more.
func spentFix(name string, failed []string) {
	mark := run + "/checks-said-" + head
	if exists(mark) {
		return
	}
	tag := "<!-- tend:checks-failed " + head + " -->"
	said, err := out("gh", "api", "repos/{owner}/{repo}/issues/"+prNumber()+"/comments", "--paginate", "--jq", fmt.Sprintf(`.[] | select(.user.login == "%s" and (.body | contains("%s"))) | .id`, account, tag))
	if err != nil {
		stop("could not read the comments on %s", pr)
	}
	if said == "" {
		body := fmt.Sprintf("**Checks failed on %s after the one fix tend asks for**\n\ntend asked %s for the fixes its flow allows, %d, so it asks no more. Failed on %s:\n\n- %s\n\nA person decides what follows: fix the branch, run the checks again, or close the pull request.\n\n%s\n",
			short(head), name, fl.Steps[name].Limit, short(head), strings.Join(failed, "\n- "), tag)
		if _, err := input(body, "gh", "pr", "comment", pr, "--body-file", "-"); err != nil {
			stop("could not comment on %s; run tend again to finish", pr)
		}
		say("said on %s that the checks on %s failed after %s was spent", pr, short(head), name)
	}
	write(mark, head+"\n")
}

// 5. Publish, when origin is listed in $state/publish with the GitHub account
// to publish as, "<origin-url> <account>" a line; otherwise the verdict stays
// here. Every write is that account's, never the owner's, so the owner can
// approve: the branch pushed, never forced; one pull request, squash-merge
// ready: its title and description the commit message the implementer
// writes, written again after a requested change until a person edits them;
// a review per reviewed head, as a comment since a PR's author may not
// approve or request changes; and the
// tend/review status on that head, so a later commit carries no passing
// result. The PR and reviews are looked up on GitHub each run, never
// remembered, so a rerun after any interruption repeats neither. Merging
// stays a person's unless the flow has a land step (landStep). A head carried
// across a clean merge of the target gets its status here too, and keeps the
// description written for the head it came from.
func publish() {
	if account == "" {
		return
	}
	head = git("rev-parse", "HEAD")
	stage("working", fmt.Sprintf("publishing %s at %s as %s", branch, short(head), account), "tend pushes it and opens its pull request or reviews the one there")
	token()
	if !pushAs(branch) {
		stop("could not push %s to origin as %s; if it has commits this checkout lacks, look at them, then run tend again", branch, account)
	}
	var err error
	if pr, err = out("gh", "pr", "list", "--head", branch, "--state", "all", "--json", "url", "--jq", ".[0].url // empty"); err != nil {
		stop("could not look for %s's pull request", branch)
	}
	// The PR is squash-merged: its title and description become the commit's,
	// so the implementer, who knows what changed and why, writes them.
	if pr == "" {
		title, body := titleBody(message(""))
		if pr, err = input(body, "gh", "pr", "create", "--base", target, "--head", branch, "--title", title, "--body-file", "-"); err != nil {
			stop("could not open a pull request for %s; run tend again, and it looks for one before opening another", branch)
		}
		write(run+"/described", head+"\n")
	} else if d := cat(run + "/described"); d != "" && d != head && !slices.Contains(carried(head), d) {
		redescribe(d)
	}
	if c := carried(head); c != nil {
		carry(head, c)
	}
	// Every verdict of the attempt on a commit of the branch, oldest first, so
	// a failure repaired within one run still shows.
	posted, err := out("gh", "api", reviewsAPI(), "--paginate", "--jq", fmt.Sprintf(`.[] | select(.user.login == "%s") | .commit_id`, account))
	if err != nil {
		stop("could not read %s's reviews", pr)
	}
	fs, _ := filepath.Glob(run + "/verdict-*.json")
	mtime := func(f string) time.Time { s, _ := os.Stat(f); return s.ModTime() }
	sort.SliceStable(fs, func(i, j int) bool { return mtime(fs[i]).Before(mtime(fs[j])) })
	for _, f := range fs {
		x := parse(cat(f))
		h := str(get(x, "head"))
		if !valid(f, h) || !ok("git", "merge-base", "--is-ancestor", h, "HEAD") || slices.Contains(strings.Split(posted, "\n"), h) {
			continue
		}
		// The status first: a rerun after it, before the review, sets it again.
		c := str(get(x, "conclusion"))
		st := "failure"
		if c == "success" {
			st = "success"
		}
		if _, err := out("gh", "api", "repos/{owner}/{repo}/statuses/"+h, "-f", "state="+st, "-f", "context=tend/review",
			"-f", "description=Independent review of "+short(h)+": "+c, "-f", "target_url="+pr); err != nil {
			stop("could not set the tend/review status on %s", short(h))
		}
		if h == head {
			markDraft(c != "success")
		}
		body := fmt.Sprintf("**Independent review of %s: %s**\n\n%s\n\n", short(h), c, str(get(x, "summary")))
		if n := strings.TrimSpace(str(get(x, "notes"))); n != "" {
			body += "**Notes that do not fail the head**, for a person to file or drop:\n\n" + n + "\n\n"
		}
		if c != "success" {
			if get(x, "actionable") == true {
				body += "The reviewer judged every finding that fails the head an in-scope fix. tend repairs once, then a person decides.\n\n"
			} else {
				body += "A finding needs a decision from a person.\n\n"
			}
		}
		if reply := cat(run + "/reply-" + h); reply != "" {
			body += "The implementer, on the changes requested:\n\n> " + strings.ReplaceAll(reply, "\n", "\n> ") + "\n\n"
		}
		if h == head && kind == "grove" {
			o, _ := both("grove", "review")
			if i := regexp.MustCompile(`(?m)^diff --git`).FindStringIndex(o); i != nil {
				o = o[:i[0]]
			}
			if o = strings.TrimRight(o, "\n"); o != "" {
				body += "<details><summary>What merging approves, from grove review</summary>\n\n```text\n" + o + "\n```\n</details>\n\n"
			}
		}
		body += "Codex reviewed this commit in a session apart from the implementer; tend posted it for attempt `" + id + "`."
		if postJSON(reviewsAPI(), map[string]any{"commit_id": h, "event": "COMMENT", "body": body}) != nil {
			stop("could not post the review of %s on %s", short(h), pr)
		}
	}
	where = pr
	write(run+"/pr", pr+"\n") // for tend watch to find; GitHub stays the record
	if item.link != nil && item.link(pr) != nil {
		stop("could not say on %s where its pull request is (%s); run tend again to finish", item.ref, pr)
	}
	say("published %s", pr)
}

// A pull request whose head failed review is made a draft, and one whose head
// passed made ready, once per reviewed head, so a person may change it after.
// The description is the squash commit's, so it does not say; and where
// GitHub offers no drafts, the review and status still do.
func markDraft(want bool) {
	d, err := out("gh", "pr", "view", pr, "--json", "isDraft", "--jq", ".isDraft")
	if err != nil || (d == "true") == want {
		return
	}
	a, as := []string{"pr", "ready", pr}, "ready for review"
	if want {
		a, as = append(a, "--undo"), "a draft"
	}
	if _, err := out("gh", a...); err != nil {
		say("could not make %s %s; its review and status say how it stands", pr, as)
	}
}

// A change request the implementer answered with no commit: its reply goes on
// the pull request once, found by the request's marker, and answered-<node id>
// keeps tend and tend watch from asking again.
func answer(nid, who, reply string) {
	tag := "<!-- tend:answered " + nid + " -->"
	said, err := out("gh", "api", "repos/{owner}/{repo}/issues/"+prNumber()+"/comments", "--paginate", "--jq", fmt.Sprintf(`.[] | select(.user.login == "%s" and (.body | contains("%s"))) | .id`, account, tag))
	if err != nil {
		stop("could not read the comments on %s", pr)
	}
	if said == "" {
		body := fmt.Sprintf("**No change made for the changes %s requested at %s**\n\nThe implementer, on why:\n\n> %s\n\ntend will not ask again. To have it tried again, request changes saying what should change; or change the branch yourself.\n\n%s\n", who, short(head), strings.ReplaceAll(strings.TrimRight(reply, "\n"), "\n", "\n> "), tag)
		if _, err := input(body, "gh", "pr", "comment", pr, "--body-file", "-"); err != nil {
			stop("could not comment on %s; run tend again to finish", pr)
		}
	}
	write(run+"/answered-"+nid, head+"\n")
	say("%s made no change for %s's request; said why on %s", impl, who, pr)
}

// The commit message for the head, which the implementer writes to
// message-<head> once: the pull request's title and description. was, if
// given, is the message on the pull request now, kept where it still holds.
func message(was string) string {
	msg := run + "/message-" + head
	if nonempty(msg) {
		return msg
	}
	prev := ""
	if was != "" {
		prev = fmt.Sprintf(" The pull request's description now is in %s, written for an earlier head: keep what still holds and correct what later commits changed.", was)
	}
	implementer()
	stage("working", fmt.Sprintf("%s writes the commit message for %s", impl, short(head)), "tend puts it on the pull request")
	ask(impl, fmt.Sprintf("Branch %s is reviewed and will be squash-merged into %s: the pull request's title becomes the commit's subject and its description the commit's body. Write that commit message to %s and nothing else. A subject of at most 72 characters in this repository's commit style (see git log %s), a blank line, then a body wrapped at 72 columns saying what the change does and why, for someone reading the history later who was not here. Describe the merged result, not how it was reached: no reviews, repairs, agents, tools, attempt IDs or file lists. End with the line '%s'.%s Commit nothing and leave the branch as it is.", branch, target, msg, target, trailer, prev), "")
	if !nonempty(msg) {
		stop("%s settled without writing %s; look at its pane, then run tend again", impl, msg)
	}
	if git("rev-parse", "HEAD") != head || !clean() {
		stop("%s changed the branch while writing the commit message", impl)
	}
	return msg
}

func titleBody(msg string) (string, string) {
	lines := strings.Split(cat(msg), "\n")
	if len(lines) > 2 {
		return lines[0], strings.Join(lines[2:], "\n")
	}
	return lines[0], ""
}

// The title and description as GitHub may give them back, to compare.
func fold(title, body string) string {
	n := func(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n")) }
	return n(title) + "\n\n" + n(body)
}

// After later commits, the description is written again for the new head
// when it is still the message tend put there, the run folder's described
// naming that head; once a person has edited it, it is theirs and tend leaves it.
func redescribe(d string) {
	s, err := out("gh", "pr", "view", pr, "--json", "state,title,body")
	if err != nil {
		stop("could not read %s", pr)
	}
	x := parse(s)
	if str(get(x, "state")) != "OPEN" {
		return
	}
	now, was := fold(str(get(x, "title")), str(get(x, "body"))), run+"/message-"+d
	if nonempty(run+"/message-"+head) && now == fold(titleBody(run+"/message-"+head)) {
		write(run+"/described", head+"\n") // a run stopped after the edit
		return
	}
	if now != fold(titleBody(was)) {
		os.Remove(run + "/described")
		say("left %s's description as it is: a person has edited it", pr)
		return
	}
	title, body := titleBody(message(was))
	if fold(title, body) != now {
		if _, err := input(body, "gh", "pr", "edit", pr, "--title", title, "--body-file", "-"); err != nil {
			stop("could not rewrite %s's description", pr)
		}
		say("rewrote %s's description for %s", pr, short(head))
	}
	write(run+"/described", head+"\n")
}

// A change request counts only from a person with write access: an author
// GitHub gives, in the review's field, the association owner, member or
// collaborator. tend and tend watch both decide by it. In a repository an
// organization owns, a member with read access only still counts.
func writer(field string) string {
	return fmt.Sprintf(`(.%s | IN("OWNER", "MEMBER", "COLLABORATOR"))`, field)
}

// Whether the newest review that approves or requests changes, by a person
// with write access, on the head or a head it is carried from, approves.
func approved() bool {
	hs := carried(head)
	if hs == nil {
		hs = []string{head}
	}
	return approvedOn(hs)
}

// Whether the newest review that approves or requests changes, by a person
// with write access, on any of the heads hs, approves: hs[0] is the head,
// the rest those it counts approvals of.
func approvedOn(hs []string) bool {
	q, _ := json.Marshal(hs)
	s, err := out("gh", "api", reviewsAPI(), "--paginate", "--jq", fmt.Sprintf(`.[] | select((.state == "APPROVED" or .state == "CHANGES_REQUESTED") and (.commit_id | IN(%s[])) and %s) | .state + " " + .user.login`, q, writer("author_association")))
	if err != nil {
		stop("could not read the reviews of %s", pr)
	}
	ls := strings.Split(s, "\n")
	last := strings.Fields(ls[len(ls)-1])
	if len(last) == 2 && last[0] == "APPROVED" {
		say("%s approved %s at %s", last[1], pr, short(hs[0]))
		return true
	}
	return false
}

// The newest change request on the open PR's head by a person with write
// access, or nothing; it says whose requests on that head it leaves.
func requested() (any, error) {
	s, err := out("gh", "pr", "view", pr, "--json", "state", "--jq", ".state")
	if err != nil || s != "OPEN" {
		return nil, nil
	}
	s, err = out("gh", "api", reviewsAPI(), "--paginate", "--jq", fmt.Sprintf(`.[] | select(.state == "CHANGES_REQUESTED" and .commit_id == "%s" and (%s | not)) | .user.login`, head, writer("author_association")))
	if who := strings.Fields(s); err == nil && len(who) > 0 {
		slices.Sort(who)
		say("left the change request of %s on %s: GitHub does not give them write access", strings.Join(slices.Compact(who), ", "), pr)
	}
	s, err = out("gh", "api", reviewsAPI(), "--paginate", "--jq", fmt.Sprintf(`[.[] | select(.state == "CHANGES_REQUESTED" and .commit_id == "%s" and %s)] | last // empty`, head, writer("author_association")))
	if err != nil || s == "" {
		return nil, err
	}
	return parse(s), nil
}
