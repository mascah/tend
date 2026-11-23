package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed flows
var builtin embed.FS

const flowPath = "tend/flow.yaml"

// A prompt file: the target's under tend/ when rev is given and holds it,
// else tend's own.
func promptText(rev, p string) (string, bool) {
	if rev != "" {
		if s, err := quiet("git", "-C", repo, "show", rev+":tend/"+p); err == nil {
			return s, true
		}
	}
	b, err := builtin.ReadFile("flows/" + p)
	return strings.TrimRight(string(b), "\n"), err == nil
}

// The flows in effect at rev, the target as origin holds it: each one the
// project's tend/flow.yaml holds, and tend's own for the rest, with where
// each came from. A flow that cannot run stops tend, exit 3, before any agent
// starts.
func flowsAt(rev string) (flowFile, map[string]string) {
	var ff flowFile
	from := map[string]string{}
	if rev != "" && ok("git", "-C", repo, "cat-file", "-e", rev+":"+flowPath) {
		s, _ := quiet("git", "-C", repo, "show", rev+":"+flowPath)
		if why := decode(s, &ff); why != "" {
			stop("%s %s", flowPath, why)
		}
		if why := ff.Watch.fault(); why != "" {
			stop("%s %s", flowPath, why)
		}
		has := func(p string) bool { _, ok := promptText(rev, p); return ok }
		for name, f := range ff.named() {
			if *f == nil {
				continue
			}
			if why := (*f).fault(name, has); why != "" {
				stop("%s %s", flowPath, why)
			}
			from[name] = flowPath + " on " + rev
		}
	}
	for name, f := range ff.named() {
		if *f != nil {
			continue
		}
		var own flowFile
		b, _ := builtin.ReadFile("flows/" + name + ".yaml")
		if why := decode(string(b), &own); why != "" {
			stop("flows/%s.yaml %s", name, why)
		}
		*f = *own.named()[name]
		if why := (*f).fault(name, func(p string) bool { _, ok := promptText("", p); return ok }); why != "" {
			stop("flows/%s.yaml %s", name, why)
		}
		from[name] = "built in"
	}
	return ff, from
}

// The count tend watch keeps going at rev, the target as origin holds it, or
// -1 for none; and why the flow file is not valid, or "". Stops nothing.
func watchCount(rev string) (int, string) {
	if !ok("git", "-C", repo, "cat-file", "-e", rev+":"+flowPath) {
		return -1, ""
	}
	s, _ := quiet("git", "-C", repo, "show", rev+":"+flowPath)
	var ff flowFile
	if why := decode(s, &ff); why != "" {
		return -1, flowPath + " " + why
	}
	if why := ff.Watch.fault(); why != "" {
		return -1, flowPath + " " + why
	}
	return ff.Watch.count(), ""
}

// tend flow <repo>: the flows the repository's attempts would run, as a flow
// file, each said to be the project's or tend's own. Reads only.
func printFlows() {
	cd(repo)
	t := ""
	if g, err := groveJSON(); err == nil {
		t = str(get(g, "target"))
	} else if s, err := quiet("gh", "repo", "view", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name"); err == nil {
		t = s
	} else {
		t = strings.TrimPrefix(gitQ("symbolic-ref", "--short", "refs/remotes/origin/HEAD"), "origin/")
	}
	if t == "" {
		t = gitQ("branch", "--show-current")
	}
	rev := t
	if _, err := quiet("git", "remote", "get-url", "origin"); err == nil {
		quiet("git", "fetch", "-q", "origin", t)
		rev = "origin/" + t
	}
	ff, from := flowsAt(rev)
	fmt.Printf("# The flows tend runs in %s, read from %s. Saved as %s, with\n# the prompt files it names under tend/ where yours differ from tend's own,\n# it is what attempts run.\n", repo, rev, flowPath)
	if ff.Watch != nil {
		fmt.Printf("\n# watch: %s on %s\nwatch:\n  at_once: %d\n", flowPath, rev, ff.Watch.count())
	}
	for _, name := range []string{"build", "propose"} {
		fmt.Printf("\n# %s: %s\n", name, from[name])
		if from[name] == "built in" {
			b, _ := builtin.ReadFile("flows/" + name + ".yaml")
			os.Stdout.Write(b)
			continue
		}
		b, _ := yaml.Marshal(map[string]*flow{name: *ff.named()[name]})
		os.Stdout.Write(b)
	}
	os.Exit(0)
}

// The flow of the name in effect at rev, ready to walk.
func loadFlow(name, rev string) {
	ff, from := flowsAt(rev)
	fl, flowName, flowFrom = *ff.named()[name], name, from[name]
	promptRev = ""
	if from[name] != "built in" {
		promptRev = rev
	}
	reqStep = ""
	if p := fl.Steps["publish"]; p != nil {
		reqStep = p.Next["requested"]
	}
}
