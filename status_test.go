package main

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var statusScenarios = []scenario{
	{"overview", func(s *sc) {
		s.project("main", true)
		s.allow()
		s.change("261016-c0001", "Another", "")
		s.issues("1:tend", "2:tend")
		s.propose("CAP=1")
		s.run()
		s.request()
		o := s.overview()
		s.check("the overview names an approved change in no cycle, with Grove's verdict", s.has(after(o, "Approved changes in no cycle", 1), `^  261016-c0001 Another: grove claim 261016-c0001 would claim it`))
		s.check("a cycle not done, with Grove's verdict", s.has(o, `^  261114-k0000 One: 261114-k0000 is in flight in attempt`))
		s.check("the attempt on this runner and its stage", s.has(o, `^  [0-9]{6}-[a-z0-9]{5} \(261114-k0000\): done: success`))
		s.check("the proposal to review", strings.Contains(after(o, "Proposals to review", 1), "Propose a thing"))
		s.check("the implementation to review, with the change requested", s.has(after(o, "Implementations to review", 1), `Add the file .*changes requested`))
		s.check("and the request still waiting for intake", strings.Contains(after(o, "Requests waiting", 1), "#2 Request 2"))
		s.check("but not the one already drafted", !strings.Contains(o, "#1 Request 1"))
		s.touch(s.prOf("1") + ".merged")
		s.check("nor once its proposal is merged", !strings.Contains(s.overview(), "#1 Request 1"))
		s.bash(s.repo, `mkdir -p grove/changes/archive && git mv grove/changes/261015-c0000-a-change grove/changes/archive/261015-c0000-a-change && git commit -qm landed && git push -q origin main`)
		s.check("an attempt whose cycle is done is left out", !strings.Contains(s.overview(), "(261114-k0000)"))
	}},
	{"change-overview", func(s *sc) {
		s.project("main", true)
		s.change("261016-c0001", "Another", "")
		s.crun("261016-c0001")
		s.check("the overview lists the attempt at the change", s.has(s.overview(), `^  [0-9]{6}-[a-z0-9]{5} \(261016-c0001\): done: success`))
		s.bash(s.repo, `grove archive 261016-c0001 >/dev/null && git add -A && git commit -qm landed && git push -q origin main`)
		s.check("and no longer once the target holds it archived", !strings.Contains(s.overview(), "(261016-c0001)"))
	}},
	{"status-interrupted", func(s *sc) {
		s.project("main", false)
		bg := s.background("IMPL_SLEEP=30")
		st := s.status()
		s.check("status names the running tend", strings.Contains(st, fmt.Sprintf("tend is on it: pid %d", bg.Process.Pid)))
		s.check("and what it is waiting for", s.has(st, `working: impl-.* implements`))
		killed(bg)
		if pid, err := strconv.Atoi(s.cat(s.st + "/sleeper")); err == nil {
			syscall.Kill(pid, syscall.SIGTERM)
		}
		time.Sleep(500 * time.Millisecond)
		s.check("after a kill, status says to run tend again", strings.Contains(s.status(), "tend stopped before this finished"))
	}},
}
