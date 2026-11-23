package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func needsYou(f string) bool {
	return regexp.MustCompile(`(?m)^[^ ]* needs-you:`).MatchString(cat(f))
}

// tend watch: every WATCH_SECONDS (default 300), each attempt whose tend died
// on it or whose pull request is open with a change request on its head is
// run again, and for
// each repository named tend propose runs and then the oldest cycle Grove
// would claim that its plan marks unattended is started: marking it is the
// person's approval to build it with no one there. An attempt or request
// waiting for a person (needs-you) is left alone until a person has run tend
// on it; a claimed cycle is Grove's to refuse. A repository whose flow file
// sets watch.at_once is kept going under that count instead (keep).
func watch(repos []string) {
	self := must(os.Executable())
	again := func(a ...string) {
		c := exec.Command(self, a...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		c.Run()
	}
	var started []chan struct{}
	begin := func(a ...string) chan struct{} {
		c := exec.Command(self, a...)
		c.Stdout, c.Stderr = os.Stdout, os.Stderr
		done := make(chan struct{})
		if c.Start() != nil {
			close(done)
			return done
		}
		go func() { c.Wait(); close(done) }()
		started = append(started, done)
		return done
	}
	for {
		counted := map[string]bool{} // repositories kept under a count
		for _, r := range repos {
			r, err := abs(r)
			if err != nil {
				continue
			}
			c := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
			c.Dir = r
			b, err := c.Output()
			if err != nil {
				continue
			}
			n := countIn(r, true)
			if n >= 0 {
				counted[lockFile(r, "")] = true // by the repository's identity, however a run folder spells its path
				keep(r, n, begin, true)
			}
			s := strings.ReplaceAll(strings.TrimSpace(string(b)), "/", "_")
			waits, _ := filepath.Glob(state + "/intake/" + s + "-*/status")
			if !slices.ContainsFunc(waits, needsYou) {
				again("propose", r)
			}
			if c := unattended(r); n < 0 && c != "" {
				say("starting unattended cycle %s in %s", c, r)
				again(r, c)
			}
		}
		for _, d := range runDirs() {
			a := strings.SplitN(cat(d+"/attempt"), "\n", 3)
			if a[0] == "" {
				continue
			}
			k := lockFile(a[0], "")
			if _, seen := counted[k]; !seen { // a repository not named: its count, if any, holds its reruns
				if n := countIn(a[0], false); n >= 0 {
					counted[k] = true
					keep(a[0], n, begin, false)
				} else {
					counted[k] = false
				}
			}
			why := ""
			if !counted[k] {
				why = rerun(d)
			}
			if why == "" {
				continue
			}
			say("%s; running tend on %s", why, a[1])
			again(a[0], a[1])
		}
		if os.Getenv("WATCH_ONCE") != "" {
			for _, d := range started {
				<-d
			}
			os.Exit(0)
		}
		secs, err := strconv.Atoi(os.Getenv("WATCH_SECONDS"))
		if err != nil {
			secs = 300
		}
		time.Sleep(time.Duration(secs) * time.Second)
	}
}

// The count tend watch keeps going in repository r, AT_ONCE lowering it, or -1
// for none; 0 where the flow file is not valid, saying why. sync first
// fast-forwards its target to origin's.
func countIn(r string, sync bool) int {
	g, err := groveIn(r)
	if err != nil {
		return -1
	}
	t := str(get(g, "target"))
	if sync {
		syncTarget(r, t)
	}
	repo = r
	rev := t
	if _, err := quiet("git", "-C", r, "remote", "get-url", "origin"); err == nil {
		rev = "origin/" + t
	}
	n, why := watchCount(rev)
	if why != "" {
		say("starting nothing in %s: %s", r, why)
		n = 0
	}
	if a, err := strconv.Atoi(os.Getenv("AT_ONCE")); err == nil && a >= 0 && n >= 0 {
		n = min(n, a)
	}
	return n
}

// Why tend watch runs tend again on the attempt of run folder d, or "": its
// pull request is open with a change request on its head, from a person with
// write access, not yet answered; or, where its flow's publish acts on an
// approval (approves, written by its last run), the newest review on its head
// by such a person that approves or requests changes approves; or the tend on
// it died (stopped); and it waits for no person.
func rerun(d string) string {
	a := strings.Split(cat(d+"/attempt"), "\n")
	if len(a) < 2 || needsYou(d+"/status") {
		return ""
	}
	if p := cat(d + "/pr"); p != "" {
		// An approval holds for a head carried from the approved one: carried-<head>
		// names the head it came from, written once tend found it so in Git.
		from := map[string]string{}
		ms, _ := filepath.Glob(d + "/carried-*")
		for _, m := range ms {
			from[strings.TrimPrefix(filepath.Base(m), "carried-")] = cat(m)
		}
		c, _ := json.Marshal(from)
		s, _ := quiet("gh", "pr", "view", p, "--json", "state,headRefOid,reviews", "--jq",
			string(c)+` as $c | select(.state == "OPEN") | .headRefOid as $h | [.reviews[] | select((.state == "CHANGES_REQUESTED" or .state == "APPROVED") and `+writer("authorAssociation")+`)] |
			(map(select(.state == "CHANGES_REQUESTED" and .commit.oid == $h)) | last | .id // ""), (map(select(.commit.oid == $h or .commit.oid == $c[$h])) | last | .state // "")`)
		ls := strings.Split(s, "\n")
		switch {
		case ls[0] != "" && !exists(d+"/answered-"+ls[0]):
			return "changes requested on " + p
		case len(ls) > 1 && ls[1] == "APPROVED" && exists(d+"/approves"):
			return p + " was approved"
		}
	}
	if stopped(d, a[0], a[1]) {
		return "tend stopped on attempt " + filepath.Base(d) + " before it finished"
	}
	return ""
}

// Whether the tend on the attempt of run folder d, at work item w of
// repository r, died: its status says working, no tend holds its lock, and,
// for a Grove cycle or change, Grove shows d's attempt still in flight, not
// retried since from another checkout.
func stopped(d, r, w string) bool {
	if !working(d+"/status") || held(r, w) {
		return false
	}
	if isItem(w) {
		return true
	}
	g, err := groveIn(r)
	if err != nil {
		return false
	}
	for _, x := range list(get(g, "runs")) {
		if str(get(x, "id")) == filepath.Base(d) {
			return get(x, "in_flight") == true
		}
	}
	return false
}

// tend watch under a count: of repository r's attempts on this runner, at most
// n being worked on, a tend holding its lock, whoever started it. Attempts to
// run again go first, then each cycle not done and change in no cycle that
// Grove calls claimable, oldest created first, each started in the background
// and its claim seen to land before the next is chosen, so that one which
// overlaps it is not started to be refused. Without starts, as for a
// repository watch was not given, it only runs attempts again.
// ponytail: run folders go oldest first by name, which a Grove attempt's
// date-first ID makes so; an issue build's folder sorts after them.
func keep(r string, n int, begin func(...string) chan struct{}, starts bool) {
	busy := 0
	ls, _ := filepath.Glob(lockFile(r, "*"))
	for _, l := range ls {
		if w := strings.TrimPrefix(l, lockFile(r, "")); w != "intake" && w != "land" && held(r, w) {
			busy++
		}
	}
	free := n - busy
	for _, d := range runDirs() {
		a := strings.SplitN(cat(d+"/attempt"), "\n", 3)
		if free <= 0 {
			return
		}
		if len(a) < 2 || lockFile(a[0], "") != lockFile(r, "") || held(a[0], a[1]) {
			continue
		}
		why := rerun(d)
		if why == "" {
			continue
		}
		say("%s; running tend on %s", why, a[1])
		begin(a[0], a[1])
		free--
	}
	tried := map[string]bool{}
	for ; starts && free > 0; free-- {
		g, _ := groveIn(r)
		next, when := "", ""
		for _, x := range append(list(get(g, "cycles")), list(get(g, "changes"))...) {
			w := str(get(x, "id"))
			if get(x, "claimable") == true && get(x, "done") != true && get(x, "cycle") == nil && !tried[w] && !held(r, w) &&
				(next == "" || str(get(x, "created")) < when) {
				next, when = w, str(get(x, "created"))
			}
		}
		if next == "" {
			return
		}
		tried[next] = true
		say("starting %s in %s", next, r)
		before := runDirs()
		done := begin(r, next)
		// Its claim has landed once its run folder names it: tend makes that
		// folder only after Grove shows the claim.
		for landed := false; !landed; {
			select {
			case <-done:
				landed = true
			case <-time.After(200 * time.Millisecond):
				landed = slices.ContainsFunc(runDirs(), func(d string) bool {
					a := strings.Split(cat(d+"/attempt"), "\n")
					return !slices.Contains(before, d) && len(a) > 1 && a[1] == next
				})
			}
		}
	}
}

// The oldest cycle of the repository that Grove would claim now (eligible, not
// claimed, the project not paused) and whose plan marks it unattended, or "".
func unattended(r string) string {
	g, err := groveIn(r)
	if err != nil {
		return ""
	}
	best, when := "", ""
	for _, x := range list(get(g, "cycles")) {
		if get(x, "unattended") == true && get(x, "runs") == true && get(x, "done") != true && (best == "" || str(get(x, "created")) < when) {
			best, when = str(get(x, "id")), str(get(x, "created"))
		}
	}
	return best
}
