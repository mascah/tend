package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// tend status <repo>: what waits for whom in the repository, read from Grove's
// JSON, GitHub and this runner's run folders. Reads only.
func overview() {
	cd(repo)
	proj, err := groveJSON()
	if err != nil { // a repository without Grove
		proj = map[string]any{}
	}
	fmt.Println("Approved changes in no cycle:")
	for _, c := range list(get(proj, "changes")) {
		if str(get(c, "state")) == "approved" && get(c, "cycle") == nil {
			fmt.Printf("  %s %s: %s\n", interp(get(c, "id")), interp(get(c, "title")), interp(get(c, "verdict")))
		}
	}
	fmt.Println("Cycles not done:")
	for _, y := range list(get(proj, "cycles")) {
		if d := get(y, "done"); d == nil || d == false {
			fmt.Printf("  %s %s: %s\n", interp(get(y, "id")), interp(get(y, "title")), interp(get(y, "verdict")))
		}
	}
	fmt.Println("Attempts on this runner, for cycles and changes not done and issues:")
	ds := runDirs()
	for _, d := range ds {
		a := strings.Split(cat(d+"/attempt"), "\n")
		if a[0] != repo || !exists(d+"/status") || len(a) < 2 {
			continue
		}
		if _, open := workState(proj, a[1]); !isItem(a[1]) && !open {
			continue
		}
		line := strings.SplitN(cat(d+"/status"), "\n", 2)[0]
		if _, rest, found := strings.Cut(line, " "); found {
			line = rest
		}
		fmt.Printf("  %s (%s): %s\n", filepath.Base(d), a[1], line)
	}
	s, err := quiet("gh", "pr", "list", "--state", "open", "--json", "number,title,headRefName,url,reviewDecision")
	if err != nil {
		fmt.Println("GitHub: not readable with gh here")
		os.Exit(0)
	}
	prs := list(parse(s))
	pull := func(p any) string {
		return fmt.Sprintf("  #%s %s %s", interp(get(p, "number")), interp(get(p, "title")), interp(get(p, "url")))
	}
	fmt.Println("Proposals to review (answer their questions, revise or merge):")
	for _, p := range prs {
		if h := str(get(p, "headRefName")); strings.HasPrefix(h, "tend/issue-") || strings.HasPrefix(h, "tend/jira-") {
			fmt.Println(pull(p))
		}
	}
	fmt.Println("Implementations to review (review and merge):")
	for _, p := range prs {
		if h := str(get(p, "headRefName")); strings.HasPrefix(h, "grove/") || strings.HasPrefix(h, "tend/build-") {
			asked := ""
			if str(get(p, "reviewDecision")) == "CHANGES_REQUESTED" {
				asked = ", changes requested: tend makes them on its next run"
			}
			fmt.Println(pull(p) + asked)
		}
	}
	fmt.Println("Requests waiting for tend propose:")
	handled, _ := quiet("gh", "pr", "list", "--state", "all", "--limit", "1000", "--json", "headRefName", "--jq", `[.[].headRefName | select(startswith("tend/issue-") or startswith("tend/jira-"))]`)
	done := list(parse(handled))
	lines, _ := quiet("gh", "issue", "list", "--label", "tend", "--state", "open", "--json", "number,title,url", "--jq", `.[] | "  #\(.number) \(.title) \(.url)"`)
	for _, line := range strings.Split(lines, "\n") {
		if line == "" {
			continue
		}
		num, _, _ := strings.Cut(strings.TrimPrefix(line, "  #"), " ")
		if !slices.Contains(done, any("tend/issue-"+num)) {
			fmt.Println(line)
		}
	}
	origin := gitQ("remote", "get-url", "origin")
	for _, m := range jiraMaps() {
		if m.origin != origin {
			continue
		}
		if os.Getenv("JIRA_EMAIL") == "" || os.Getenv("JIRA_API_TOKEN") == "" {
			fmt.Printf("  Jira %s: not readable without JIRA_EMAIL and JIRA_API_TOKEN\n", m.project)
			continue
		}
		for _, x := range jiraSearch(m, fmt.Sprintf(`project = "%s" AND labels = "tend" AND statusCategory != Done ORDER BY created ASC`, m.project)) {
			if k := str(get(x, "key")); !slices.Contains(done, any("tend/jira-"+k)) {
				fmt.Printf("  %s %s %s/browse/%s\n", k, str(get(x, "fields", "summary")), m.site, k)
			}
		}
	}
	os.Exit(0)
}

// tend status <repo> <cycle or change>: the newest attempt at it, what it last recorded
// and whether a tend is on it now. Reads only; takes no lock, prompts nobody.
func status() {
	if isItem(cycle) {
		run = issueRun(cycle)
		if !exists(run) {
			fmt.Printf("no attempt at %s\n", cycle)
			os.Exit(0)
		}
		fmt.Printf("attempt %s of %s: tend/build-%s\n", filepath.Base(run), cycle, strings.TrimPrefix(cycle, "issue-"))
	} else {
		cd(repo)
		s := must(out("grove", "changes", "--json"))
		r := itemRun(parse(s), cycle)
		if r == nil {
			fmt.Printf("no attempt at %s\n", cycle)
			os.Exit(0)
		}
		run = state + "/runs/" + str(get(r, "id"))
		fmt.Printf("attempt %s of %s: %s in %s\n", interp(get(r, "id")), cycle, interp(get(r, "branch")), interp(get(r, "checkout")))
	}
	if held(repo, cycle) {
		fmt.Printf("tend is on it: %s\n", cat(lockFile(repo, cycle)))
	} else if working(run + "/status") {
		fmt.Println(strings.SplitN(cat(run+"/status"), "\n", 2)[0])
		fmt.Println("next: tend stopped before this finished; run it again, and it waits for an agent still working rather than prompting it again")
		os.Exit(0)
	}
	if b, err := os.ReadFile(run + "/status"); err == nil {
		os.Stdout.Write(b)
	} else {
		fmt.Println("nothing recorded yet")
	}
	os.Exit(0)
}
