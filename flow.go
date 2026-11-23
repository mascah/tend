package main

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ---- flows
//
// A flow is a named set of steps an attempt goes through once it has its
// branch and worktree. A step is an agent given a prompt, a result tend
// checks, and what each outcome of that result allows next: another step, end
// (the flow is done) or person (tend stops for one). tend runs two flows,
// build and propose; a project changes them with tend/flow.yaml on its
// target branch, each prompt a file beside it, and tend ships its own in
// flows/ for every flow that file does not hold.

type step struct {
	Agent    string            `yaml:"agent,omitempty"`    // claude or codex: a fresh agent of that harness
	Continue string            `yaml:"continue,omitempty"` // or an earlier step, whose agent is prompted again
	Prompt   string            `yaml:"prompt,omitempty"`   // a file beside the flow file
	Result   string            `yaml:"result,omitempty"`   // what tend checks once the turn ends
	Limit    int               `yaml:"limit,omitempty"`    // asked at most this often in an attempt
	Next     map[string]string `yaml:"next,omitempty"`     // outcome: step, end or person
}

type flow struct {
	Start string           `yaml:"start"`
	Steps map[string]*step `yaml:"steps"`
}

type flowFile struct {
	Watch   *watchSetting `yaml:"watch,omitempty"`
	Build   *flow         `yaml:"build,omitempty"`
	Propose *flow         `yaml:"propose,omitempty"`
}

// How many attempts tend watch keeps going in the repository: with it, watch
// starts what Grove calls claimable; without, it starts only cycles marked
// unattended.
type watchSetting struct {
	AtOnce any `yaml:"at_once"` // a whole number of at least 1, checked by fault
}

// Why the setting is not valid, said after the file's name, or "".
func (w *watchSetting) fault() string {
	if w == nil {
		return ""
	}
	if n, isInt := w.AtOnce.(int); !isInt || n < 1 {
		if w.AtOnce == nil {
			return "gives watch no at_once, where at_once is a whole number of at least 1"
		}
		return fmt.Sprintf("gives watch an at_once of %v, where at_once is a whole number of at least 1", w.AtOnce)
	}
	return ""
}

// The count, or -1 for none.
func (w *watchSetting) count() int {
	if w != nil {
		if n, isInt := w.AtOnce.(int); isInt {
			return n
		}
	}
	return -1
}

func (ff *flowFile) named() map[string]**flow {
	return map[string]**flow{"build": &ff.Build, "propose": &ff.Propose}
}

// A step's result, as the flow says it: commit (a new commit on the branch),
// file <name> (a file in the run folder that is not empty), verdict <name> (a
// verdict file as the review spec defines one), or commit or file <name>.
// None at all is a turn that ends.
type result struct {
	commit        bool
	file, verdict string // names in the run folder, placeholders not yet filled
}

func parseResult(s string) (result, bool) {
	var r result
	f := strings.Fields(s)
	switch {
	case len(f) == 0:
	case len(f) == 1 && f[0] == "commit":
		r.commit = true
	case len(f) == 2 && f[0] == "file":
		r.file = f[1]
	case len(f) == 2 && f[0] == "verdict":
		r.verdict = f[1]
	case len(f) == 4 && f[0] == "commit" && f[1] == "or" && f[2] == "file":
		r.commit, r.file = true, f[3]
	default:
		return r, false
	}
	return r, !strings.Contains(r.file+r.verdict, "/")
}

// The outcomes a step can have, which are what its next may name.
func (s *step) outcomes(name, flowName string) []string {
	var o []string
	r, _ := parseResult(s.Result)
	switch {
	case name == "publish" && flowName == "build":
		o = []string{"ok", "requested", "approved"}
	case name == "land":
		o = []string{"ok", "conflict", "failed"}
	case r.verdict != "":
		o = []string{"success", "actionable", "failure"}
	default:
		o = []string{"ok"}
	}
	if s.Limit > 0 {
		o = append(o, "spent")
	}
	return o
}

func sorted[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func and(xs []string) string {
	if len(xs) < 2 {
		return strings.Join(xs, "")
	}
	return strings.Join(xs[:len(xs)-1], ", ") + " and " + xs[len(xs)-1]
}

// Why the flow cannot run, said after the file's name, or "". has says
// whether a prompt file exists.
func (f *flow) fault(name string, has func(string) bool) string {
	if len(f.Steps) == 0 {
		return fmt.Sprintf("gives flow %s no steps", name)
	}
	if f.Steps[f.Start] == nil {
		return fmt.Sprintf("starts flow %s at %q, a step it does not have", name, f.Start)
	}
	kinds := []string{""}
	if name == "build" {
		kinds = []string{"grove", "issue"} // {kind} in a prompt's name
	}
	for _, n := range sorted(f.Steps) {
		s := f.Steps[n]
		at := fmt.Sprintf("gives step %s of flow %s", n, name)
		if s == nil {
			return at + " nothing"
		}
		if n == "end" || n == "person" {
			return fmt.Sprintf("names a step %s in flow %s, a word kept for what follows a step", n, name)
		}
		if n == "publish" || n == "land" {
			if s.Agent != "" || s.Continue != "" || s.Prompt != "" || s.Result != "" || s.Limit != 0 {
				return fmt.Sprintf("gives step %s of flow %s more than next, though %s is tend's own", n, name, map[string]string{"publish": "publishing", "land": "landing"}[n])
			}
			if n == "land" && name != "build" {
				return fmt.Sprintf("names a step land in flow %s, where tend lands nothing", name)
			}
			if n == "land" && f.Steps["publish"] == nil {
				return fmt.Sprintf("gives flow %s a step land and no publish, where land lands the pull request publish opens", name)
			}
		} else {
			if (s.Agent == "") == (s.Continue == "") {
				return at + " both or neither of agent and continue, where it needs one"
			}
			if s.Agent != "" && s.Agent != "claude" && s.Agent != "codex" {
				return fmt.Sprintf("%s the harness %s, where an agent is claude or codex", at, s.Agent)
			}
			if s.Continue != "" {
				root, seen := n, 0
				for f.Steps[root] != nil && f.Steps[root].Continue != "" && seen <= len(f.Steps) {
					root, seen = f.Steps[root].Continue, seen+1
				}
				if c := f.Steps[root]; c == nil || root == "publish" || root == "land" || c.Agent == "" {
					return fmt.Sprintf("%s %q to continue, which is no step with an agent", at, s.Continue)
				}
			}
			if s.Prompt == "" {
				return at + " no prompt"
			}
			for _, k := range kinds {
				if p := strings.ReplaceAll(s.Prompt, "{kind}", k); !has(p) {
					return fmt.Sprintf("%s the prompt file %s, which does not exist", at, p)
				}
			}
			if _, ok := parseResult(s.Result); !ok {
				return fmt.Sprintf("%s the result %q, where a result is commit, file <name>, verdict <name> or commit or file <name>", at, s.Result)
			}
			if s.Limit < 0 {
				return at + " a limit below 0"
			}
		}
		outs := s.outcomes(n, name)
		for _, o := range sorted(s.Next) {
			if !slices.Contains(outs, o) {
				return fmt.Sprintf("%s the outcome %s, which it cannot have; it has %s", at, o, and(outs))
			}
			if to := s.Next[o]; to != "end" && to != "person" && f.Steps[to] == nil {
				return fmt.Sprintf("%s %q after %s, a step the flow does not have", at, to, o)
			}
		}
	}
	if l := f.loop(); l != nil {
		return fmt.Sprintf("loops through %s with no limit", and(l))
	}
	return ""
}

// The steps of a loop no step of which has a limit, or nil. A loop through
// publish's requested or approved comes round only when a person requests
// changes or approves, so people bound it.
func (f *flow) loop() []string {
	state := map[string]int{} // 1 on the path, 2 done
	var path []string
	var visit func(n string) []string
	visit = func(n string) []string {
		s := f.Steps[n]
		if s == nil || s.Limit > 0 {
			return nil
		}
		if state[n] == 1 {
			return slices.Clone(path[slices.Index(path, n):])
		}
		if state[n] == 2 {
			return nil
		}
		state[n] = 1
		path = append(path, n)
		for _, o := range sorted(s.Next) {
			if n == "publish" && (o == "requested" || o == "approved") {
				continue
			}
			if l := visit(s.Next[o]); l != nil {
				return l
			}
		}
		path = path[:len(path)-1]
		state[n] = 2
		return nil
	}
	for _, n := range append([]string{f.Start}, sorted(f.Steps)...) {
		if l := visit(n); l != nil {
			return l
		}
	}
	return nil
}

// Why the text is no flow file, or "".
func decode(s string, ff *flowFile) string {
	d := yaml.NewDecoder(strings.NewReader(s))
	d.KnownFields(true)
	if err := d.Decode(ff); err == io.EOF {
		return "is empty"
	} else if err != nil {
		msg := strings.Join(strings.Fields(strings.TrimPrefix(err.Error(), "yaml: ")), " ")
		return "is not a flow file tend can read: " + strings.NewReplacer("unmarshal errors: ", "",
			"in type main.step", "in a step", "in type main.flowFile", "at the top, where a flow is build or propose",
			"in type main.flow", "in a flow", "in type main.watchSetting", "in watch").Replace(msg)
	}
	return ""
}
