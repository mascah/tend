package main

import (
	"fmt"
)

// A request or work item from a tracker, a GitHub issue or a Jira issue, with
// the names tend gives it. fields is shaped as gh gives an issue: title, body,
// url, state (OPEN or not) and labels ({name}).
type request struct {
	named   string // in prompts: "GitHub issue #7", "Jira issue SCRUM-5"
	ref     string // in messages: "issue #7", "issue SCRUM-5"
	from    string // "#7", "SCRUM-5"
	slug    string // its proposal branch is tend/<slug>: "issue-7", "jira-SCRUM-5"
	folder  string // its intake folder ends -<folder>: "7", "jira-SCRUM-5"
	key     string // its build branch is tend/build-<key>: "7", "SCRUM-5"
	agent   string // "7", "scrum-5", as agent names are lower case
	refs    string // the line ending a proposal's commit: "Refs #7"
	closes  string // the line ending a build's commit: "Closes #7", "Refs SCRUM-5"
	source  string // before its URL in a proposal prompt: "GitHub issue "
	fields  any
	decline func(text, id string) error // comment why, once, found by id
	unlabel func() error                // take the tend label off
	link    func(pr string) error       // say on it where its pull request is, once
}

func ghRequest(num, fieldList string) request {
	s, err := out("gh", "issue", "view", num, "--json", fieldList)
	if err != nil {
		stop("could not read issue #%s", num)
	}
	return request{
		named: "GitHub issue #" + num, ref: "issue #" + num, from: "#" + num, slug: "issue-" + num, folder: num, key: num, agent: num,
		refs: "Refs #" + num, closes: "Closes #" + num, source: "GitHub issue ", fields: parse(s),
		decline: func(text, id string) error {
			said, _ := out("gh", "api", "repos/{owner}/{repo}/issues/"+num+"/comments", "--paginate", "--jq", fmt.Sprintf(`.[] | select(.user.login == "%s" and (.body | contains("%s %s"))) | .id`, account, marker, id))
			if said != "" {
				return nil
			}
			_, err := input(fmt.Sprintf("No proposal could be drafted from this request:\n\n%s\n\nEdit the request and put the `tend` label back to have it tried again.\n\n%s %s -->\n", text, marker, id), "gh", "issue", "comment", num, "--body-file", "-")
			return err
		},
		unlabel: func() error { _, err := out("gh", "issue", "edit", num, "--remove-label", "tend"); return err },
	}
}

const marker = "<!-- tend:declined"
