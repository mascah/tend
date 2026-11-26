package main

import (
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

// tend propose <repo>: one pass of intake from the repository's GitHub issues
// and the Jira projects mapped to it. Each open request labelled tend, oldest
// first, becomes a Grove change drafted by Claude with no one in the session,
// on branch tend/issue-<n> or tend/jira-<KEY>, opened as a pull request by the
// publishing account; or, when no change can be drafted from it, a comment on
// it saying why, and the label taken off, so it is not tried again until a
// person puts it back. The request's number or key is its identity: a pull
// request from its branch, in any state, means it is not drafted again. A
// handoff is finished when the run folder records linked: a folder with a pull
// request and no linked is completed on every pass, whatever the cap or the
// request's label. At most CAP (default 5) proposal pull requests stay open
// per repository. Each prompt is marked just before it goes, as in an attempt,
// so none is sent twice or skipped.
func propose() {
	cd(repo)
	account = accountOf()
	if account == "" {
		o, err := quiet("git", "remote", "get-url", "origin")
		if err != nil {
			o = repo + " has no origin; it"
		}
		stop("%s is not listed in %s/publish, so tend takes no requests from it", o, state)
	}
	token()
	proj := project()
	target, root = str(get(proj, "target")), str(get(proj, "root"))
	settingUp()
	base = fetchTarget()
	loadFlow("propose", base)
	tidy()
	setUp()
	slug, err := out("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	if err != nil {
		stop("gh cannot see %s's repository", repo)
	}
	prefix := state + "/intake/" + strings.ReplaceAll(slug, "/", "_") + "-"
	limit, err := strconv.Atoi(os.Getenv("CAP"))
	if err != nil {
		limit = 5
	}
	origin := gitQ("remote", "get-url", "origin")

	// Handoffs a stopped pass left unfinished.
	folders, _ := filepath.Glob(prefix + "*")
	for _, f := range folders {
		rest := strings.TrimPrefix(f, prefix)
		if strings.Contains(rest, ".declined-") || !exists(f+"/pr") || exists(f+"/linked") {
			continue
		}
		if k, isJ := strings.CutPrefix(rest, "jira-"); isJ {
			draft(prefix, jiraGet(k))
		} else {
			draft(prefix, ghRequest(rest, "title,body,url"))
		}
	}

	// Open proposals first, whatever the cap: a head with no tend/review from
	// the account, such as a person's revision, gets its own.
	vouchOpen()

	type pending struct {
		branch string
		get    func() request
	}
	var todo []pending
	nums, _ := out("gh", "issue", "list", "--label", "tend", "--state", "open", "--json", "number", "--jq", "sort_by(.number) | .[].number")
	for _, num := range strings.Fields(nums) {
		todo = append(todo, pending{"tend/issue-" + num, func() request { return ghRequest(num, "title,body,url") }})
	}
	for _, m := range jiraMaps() {
		if m.origin != origin {
			continue
		}
		for _, x := range jiraSearch(m, fmt.Sprintf(`project = "%s" AND labels = "tend" AND statusCategory != Done ORDER BY created ASC`, m.project)) {
			todo = append(todo, pending{"tend/jira-" + str(get(x, "key")), func() request { return jiraRequest(m, x) }})
		}
	}
	for _, t := range todo {
		if s, _ := out("gh", "pr", "list", "--head", t.branch, "--state", "all", "--json", "url", "--jq", ".[0].url // empty"); s != "" {
			continue
		}
		ps, _ := openProposals()
		if open := len(ps); open >= limit {
			say("%d proposal pull requests are open, the cap; the rest wait", open)
			break
		}
		draft(prefix, t.get())
	}
	os.Exit(0)
}

func draft(prefix string, r request) {
	b := "tend/" + r.slug
	run = prefix + r.folder
	must("", os.MkdirAll(run, 0o755))
	if !nonempty(run + "/id") {
		write(run+"/id", strconv.FormatInt(time.Now().Unix(), 10)+"\n")
	}
	id = cat(run + "/id")
	issue = r.fields
	settingUp()
	worktreeFor(b, base, "tend: "+r.ref)
	write(run+"/worktree", dir+"\n")
	setUp()
	cd(dir)
	// Named after the repository too, as an issue's attempt is: Herdr finds an
	// agent by its name alone, and two repositories can have an issue #1.
	impl = fmt.Sprintf("prop-%s-%d", r.agent, cksum([]byte(repo)))

	// The propose flow: the drafter drafts, then commits the draft or says why
	// there is none, each asked once, as the built-in flow says or the
	// project's.
	vars = map[string]string{"branch": b, "base": base, "run": run, "item": r.ref, "target": target,
		"request": quoted(r.source), "refs": r.refs, "ref": r.ref}
	walkFlow(func() string { proposal(r, b); return "ok" })
	cd(repo)
}

// The propose flow's publish: the proposal opened as a pull request by the
// account, or the request told why there is none.
func proposal(r request, b string) {
	switch {
	case commits(base + "..HEAD"):
		if !clean() {
			stop("%s: %s left uncommitted files in %s", r.ref, impl, dir)
		}
		changed := "\n" + git("diff", "--name-only", base+"...HEAD")
		if !strings.Contains(changed, "\n"+root+"/changes/") {
			stop("%s: %s has commits but no change under %s/changes/", r.ref, b, root)
		}
		stage("working", "publishing the proposal from "+r.ref, "a person reviews it")
		var err error
		pr, err = out("gh", "pr", "list", "--head", b, "--state", "all", "--json", "url", "--jq", ".[0].url // empty")
		if err != nil {
			stop("could not look for %s's pull request", b)
		}
		// Pushed only to open the pull request: once it exists the branch is
		// people's, and may be gone after a merge.
		created := pr == ""
		if created {
			if !pushAs(b) {
				stop("could not push %s as %s", b, account)
			}
			pr, err = input(git("log", "-1", "--format=%b")+"\n", "gh", "pr", "create", "--base", target, "--head", b, "--title", git("log", "-1", "--format=%s"), "--body-file", "-")
			if err != nil {
				stop("could not open a pull request for %s; run tend propose again, and it looks for one before opening another", b)
			}
		}
		write(run+"/pr", pr+"\n")
		head = git("rev-parse", "HEAD")
		if created {
			// Just pushed, so HEAD is the pull request's head. One found
			// instead may have moved on; vouchOpen checks its real head.
			vouch(head, pr)
		}
		posted, _ := out("gh", "api", reviewsAPI(), "--paginate", "--jq", fmt.Sprintf(`.[] | select(.user.login == "%s" and .commit_id == "%s") | .id`, account, head))
		if posted == "" {
			re := regexp.MustCompile("^" + regexp.QuoteMeta(root) + `/changes/([0-9]*-[a-z0-9]*)-`)
			var ids []string
			for _, f := range strings.Split(git("diff", "--name-only", "--diff-filter=A", base+"...HEAD"), "\n") {
				if m := re.FindStringSubmatch(f); m != nil && !slices.Contains(ids, m[1]) {
					ids = append(ids, m[1])
				}
			}
			sort.Strings(ids)
			q := cat(run + "/summary")
			if q == "" {
				q = "The drafter wrote no summary of open questions; see the proposal."
			}
			body := fmt.Sprintf("**Proposal drafted from %s with no one in the session**\n\n%s\n\nMerging approves the change with any question still open, and its implementation may then choose those answers within Grove's rules. Answer the consequential ones before merging, by revising the proposal on `%s` (`/grove-propose %s`). Once merged, the change can be built by its ID, `tend <repo> <change-id>`, or planned into a cycle with others: `/grove-plan`.", r.from, q, b, strings.Join(ids, " "))
			if postJSON(reviewsAPI(), map[string]any{"commit_id": head, "event": "COMMENT", "body": body}) != nil {
				stop("could not post the summary on %s", pr)
			}
		}
		if r.link != nil && r.link(pr) != nil {
			stop("could not say on %s where its proposal is (%s); run tend propose again to finish", r.ref, pr)
		}
		write(run+"/linked", pr+"\n")
		stage("done", "proposal from "+r.ref+" at "+pr, "a person answers its open questions, revises or merges it")
		say("%s: %s", r.ref, pr)
	case nonempty(run + "/declined"):
		// Said on the request once, found by the attempt's marker, then the label off.
		if r.decline(cat(run+"/declined"), id) != nil {
			stop("could not comment on %s", r.ref)
		}
		if r.unlabel() != nil {
			stop("could not take the tend label off %s", r.ref)
		}
		stage("done", "no proposal from "+r.ref+"; said why on it", "a person edits the request and labels it again")
		must("", os.Rename(run, run+".declined-"+id)) // a fresh attempt when the label comes back
		cd(repo)
		removeWorktree(dir, b)
		say("%s: no proposal, said why", r.ref)
	default:
		stop("%s: %s committed no change and wrote no reason. Look at its pane, then commit the draft or write %s/declined, or delete %s/followed to have it asked again", r.ref, impl, run, run)
	}
	cd(repo)
}

// Each open pull request the account opened from a tend/issue- or tend/jira-
// branch, as "<branch>\t<head>\t<url>". Every page is read: gh pr list stops
// at --limit, and open proposals have no bound.
func openProposals() ([]string, error) {
	o, err := out("gh", "api", "repos/{owner}/{repo}/pulls?state=open&per_page=100", "--paginate", "--jq",
		fmt.Sprintf(`.[] | select(.user.login == "%s" and (.head.ref | startswith("tend/issue-") or startswith("tend/jira-"))) | [.head.ref, .head.sha, .html_url] | @tsv`, account))
	if err != nil || o == "" {
		return nil, err
	}
	return strings.Split(o, "\n"), nil
}

// Each open proposal, its head as GitHub reports it fetched from origin and
// given tend/review once.
func vouchOpen() {
	ps, err := openProposals()
	if err != nil {
		stop("could not list the open proposal pull requests")
	}
	for _, l := range ps {
		f := strings.Split(l, "\t")
		if len(f) != 3 || f[1] == "" {
			continue
		}
		if !ok("git", "cat-file", "-e", f[1]+"^{commit}") {
			// The branch fetched into no ref of ours, so a pass beside this
			// one has nothing of its to race.
			quiet("git", "fetch", "-q", "origin", "refs/heads/"+f[0])
		}
		if !ok("git", "cat-file", "-e", f[1]+"^{commit}") {
			say("could not fetch %s's head %s; the next pass checks it", f[0], short(f[1]))
			continue
		}
		vouch(f[1], f[2])
	}
}

// The tend/review status on proposal head h, set as the account once:
// success when every path h changes against the target as fetched lies in
// one change's folder under <root>/changes/, read from Git alone; failure,
// naming why, otherwise.
func vouch(h, url string) {
	set, err := out("gh", "api", "repos/{owner}/{repo}/commits/"+h+"/statuses", "--paginate", "--jq",
		fmt.Sprintf(`.[] | select(.context == "tend/review" and .creator.login == "%s") | .state`, account))
	if err != nil {
		stop("could not read the statuses on %s", short(h))
	}
	if set != "" {
		return
	}
	st, d := "success", "Grove documents only: merging approves this change"
	if why := notDocs(h); why != "" {
		st, d = "failure", why
	}
	if r := []rune(d); len(r) > 140 { // GitHub's limit on a description
		d = string(r[:139]) + "…"
	}
	if _, err := out("gh", "api", "repos/{owner}/{repo}/statuses/"+h, "-f", "state="+st, "-f", "context=tend/review",
		"-f", "description="+d, "-f", "target_url="+url); err != nil {
		stop("could not set the tend/review status on %s", short(h))
	}
}

// Why head h is not one change's Grove documents only, or "" when it is:
// a folder directly under <root>/changes/ named <id>-<slug>, which leaves
// out the archive.
func notDocs(h string) string {
	re := regexp.MustCompile("^" + regexp.QuoteMeta(root) + `/changes/([0-9]+-[a-z0-9]+-[^/]+)/`)
	ps := changedPaths(h)
	if len(ps) == 0 {
		return "Changes nothing against " + target
	}
	folder := ""
	for _, p := range ps {
		m := re.FindStringSubmatch(p)
		switch {
		case m == nil:
			return "Changes " + p + ", outside one change's Grove documents"
		case folder == "":
			folder = m[1]
		case m[1] != folder:
			return "Changes two changes' documents: " + folder + " and " + m[1]
		}
	}
	return ""
}
