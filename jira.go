package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
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

// ---- Jira

// A mapped Jira project: TEND_STATE/jira lists "<site> <project-key>
// <origin-url>" a line.
type jiraMap struct{ site, project, origin string }

func jiraMaps() []jiraMap {
	var ms []jiraMap
	seen := map[string]bool{}
	for _, line := range strings.Split(cat(state+"/jira"), "\n") {
		f := strings.Fields(line)
		if len(f) != 3 {
			continue
		}
		if seen[f[1]] {
			stop("Jira project %s is mapped twice in %s/jira; one project maps to one repository", f[1], state)
		}
		seen[f[1]] = true
		ms = append(ms, jiraMap{strings.TrimRight(f[0], "/"), f[1], f[2]})
	}
	return ms
}

var jiraKey = regexp.MustCompile(`^[A-Z][A-Z0-9_]*-[0-9]+$`)

func isJira(c string) bool { return jiraKey.MatchString(c) }

// The project of key, mapped to this repository's origin.
func jiraMapFor(key string) jiraMap {
	p := key[:strings.LastIndex(key, "-")]
	o := gitQ("remote", "get-url", "origin")
	for _, m := range jiraMaps() {
		if m.project == p {
			if m.origin != o {
				stop("Jira project %s is mapped to %s in %s/jira, not to this repository's origin %s", p, m.origin, state, o)
			}
			return m
		}
	}
	stop("Jira project %s is not mapped to a repository in %s/jira", p, state)
	return jiraMap{}
}

// A Jira REST call through curl, Basic authentication from JIRA_EMAIL and
// JIRA_API_TOKEN given on curl's standard input, never in its arguments.
func jira(m jiraMap, method, path string, body any) (any, error) {
	email, tok := os.Getenv("JIRA_EMAIL"), os.Getenv("JIRA_API_TOKEN")
	if email == "" || tok == "" {
		stop("JIRA_EMAIL and JIRA_API_TOKEN must be set to reach %s", m.site)
	}
	esc := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	cfg := fmt.Sprintf("user = \"%s:%s\"\n", esc.Replace(email), esc.Replace(tok))
	args := []string{"-sS", "--fail-with-body", "-K", "-", "-X", method, "-H", "Accept: application/json"}
	if body != nil {
		f, err := os.CreateTemp("", "tend-jira-*.json")
		if err != nil {
			return nil, err
		}
		defer os.Remove(f.Name())
		json.NewEncoder(f).Encode(body)
		f.Close()
		args = append(args, "-H", "Content-Type: application/json", "--data-binary", "@"+f.Name())
	}
	o, err := input(cfg, "curl", append(args, m.site+path)...)
	return parse(o), err
}

const jiraFields = "summary,description,labels,status"

// Every issue the JQL finds, page by page.
func jiraSearch(m jiraMap, jql string) []any {
	var all []any
	next := ""
	for {
		q := url.Values{"jql": {jql}, "fields": {jiraFields}, "maxResults": {"50"}}
		if next != "" {
			q.Set("nextPageToken", next)
		}
		x, err := jira(m, "GET", "/rest/api/3/search/jql?"+q.Encode(), nil)
		if err != nil {
			stop("could not search Jira project %s at %s", m.project, m.site)
		}
		all = append(all, list(get(x, "issues"))...)
		if next = str(get(x, "nextPageToken")); next == "" || get(x, "isLast") == true {
			return all
		}
	}
}

// Whether a comment on the issue contains text, reading every page.
func jiraSaid(m jiraMap, key, text string) (bool, error) {
	for start := 0; ; {
		x, err := jira(m, "GET", fmt.Sprintf("/rest/api/3/issue/%s/comment?startAt=%d&maxResults=50", key, start), nil)
		if err != nil {
			return false, err
		}
		cs := list(get(x, "comments"))
		for _, c := range cs {
			if strings.Contains(adfText(get(c, "body")), text) {
				return true, nil
			}
		}
		start += len(cs)
		total, _ := strconv.Atoi(str(get(x, "total")))
		if len(cs) == 0 || start >= total {
			return false, nil
		}
	}
}

// A comment, a paragraph a line, once: not if a comment already holds once.
func jiraComment(m jiraMap, key, text, once string) error {
	if said, err := jiraSaid(m, key, once); err != nil || said {
		return err
	}
	var paras []any
	for _, line := range strings.Split(text, "\n") {
		if line != "" {
			paras = append(paras, map[string]any{"type": "paragraph", "content": []any{map[string]any{"type": "text", "text": line}}})
		}
	}
	_, err := jira(m, "POST", "/rest/api/3/issue/"+key+"/comment", map[string]any{"body": map[string]any{"type": "doc", "version": 1, "content": paras}})
	return err
}

// The issue shaped as gh gives one, its description as text.
func jiraIssue(m jiraMap, x any) any {
	state := "OPEN"
	if str(get(x, "fields", "status", "statusCategory", "key")) == "done" {
		state = "CLOSED"
	}
	var labels []any
	for _, l := range list(get(x, "fields", "labels")) {
		labels = append(labels, map[string]any{"name": str(l)})
	}
	return map[string]any{
		"title": str(get(x, "fields", "summary")), "body": strings.TrimSpace(adfText(get(x, "fields", "description"))),
		"url": m.site + "/browse/" + str(get(x, "key")), "state": state, "labels": labels,
	}
}

func jiraRequest(m jiraMap, x any) request {
	key := str(get(x, "key"))
	return request{
		named: "Jira issue " + key, ref: "issue " + key, from: key, slug: "jira-" + key, folder: "jira-" + key, key: key, agent: strings.ToLower(key),
		refs: "Refs " + key, closes: "Refs " + key, source: "Jira issue ", fields: jiraIssue(m, x),
		decline: func(text, id string) error {
			tag := "[tend:declined " + id + "]"
			return jiraComment(m, key, "No proposal could be drafted from this request:\n"+text+"\nEdit the request and put the tend label back to have it tried again.\n"+tag, tag)
		},
		unlabel: func() error {
			_, err := jira(m, "PUT", "/rest/api/3/issue/"+key, map[string]any{"update": map[string]any{"labels": []any{map[string]any{"remove": "tend"}}}})
			return err
		},
		link: func(pr string) error {
			return jiraComment(m, key, "tend opened "+pr+" for this issue. A person reviews it and decides whether to merge it.", pr)
		},
	}
}

func jiraGet(key string) request {
	m := jiraMapFor(key)
	x, err := jira(m, "GET", "/rest/api/3/issue/"+key+"?fields="+jiraFields, nil)
	if err != nil {
		stop("could not read Jira issue %s", key)
	}
	return jiraRequest(m, x)
}

// Atlassian Document Format as text that keeps what carries scope: a line per
// block, code as written, and the destination of every link and smart link.
func adfText(n any) string {
	var b strings.Builder
	var walk func(n any, prefix string)
	walk = func(n any, prefix string) {
		switch t := str(get(n, "type")); t {
		case "text":
			b.WriteString(str(get(n, "text")))
			for _, mk := range list(get(n, "marks")) {
				if str(get(mk, "type")) == "link" {
					if href := str(get(mk, "attrs", "href")); href != "" && href != str(get(n, "text")) {
						b.WriteString(" (" + href + ")")
					}
				}
			}
		case "inlineCard", "blockCard", "embedCard":
			b.WriteString(str(get(n, "attrs", "url")))
		case "hardBreak":
			b.WriteString("\n")
		case "mention", "emoji", "status", "date":
			b.WriteString(str(get(n, "attrs", "text")))
		case "listItem":
			b.WriteString(prefix + "- ")
			for _, c := range list(get(n, "content")) {
				walk(c, prefix+"  ")
			}
		case "bulletList", "orderedList":
			for _, c := range list(get(n, "content")) {
				walk(c, prefix)
			}
		default: // doc, paragraph, heading, codeBlock, blockquote, panel, table...
			for _, c := range list(get(n, "content")) {
				walk(c, prefix)
			}
		}
		switch str(get(n, "type")) {
		case "paragraph", "heading", "codeBlock", "blockCard", "embedCard", "rule":
			b.WriteString("\n")
		}
	}
	walk(n, "")
	return regexp.MustCompile(`\n{3,}`).ReplaceAllString(b.String(), "\n\n")
}
