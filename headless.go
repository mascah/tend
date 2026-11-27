package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// headless: agents with no terminal, the default. Each turn is one process of
// the agent's harness, claude -p or codex exec, started in a session of its
// own so it outlives a killed tend as a pane does, with its prompt on its
// standard input. Everything about an agent is in $state/agents/<name>/:
// agent.json (its harness, flags, session and the latest turn's process), and
// per turn n its n.prompt, n.out (Claude's stream-json, Codex's JSONL), n.err
// and, for Codex, n.last. The agent is working while that process runs; once
// it has ended, nothing is left that could act on the prompt, so the agent
// counts as gone and a step it left unfinished is asked once more, resuming
// the same session. Worktrees are made by Git under $state/worktrees/.
type headless struct{}

type agentRec struct {
	Harness string   `json:"harness"`
	Args    []string `json:"args"`
	Session string   `json:"session,omitempty"` // Claude's, chosen at start; Codex's, from its first turn
	Turn    int      `json:"turn"`              // the latest turn
	PID     int      `json:"pid,omitempty"`     // its process
	Began   string   `json:"began,omitempty"`   // when that process began, as ps says, so a reused pid is not taken for it
}

func agentDir(name string) string { return state + "/agents/" + name }

func loadAgent(name string) *agentRec {
	b, err := os.ReadFile(agentDir(name) + "/agent.json")
	if err != nil {
		return nil
	}
	var a agentRec
	if json.Unmarshal(b, &a) != nil {
		return nil
	}
	return &a
}

func (a *agentRec) save(name string) {
	b, _ := json.MarshalIndent(a, "", "  ")
	f := agentDir(name) + "/agent.json"
	write(f+".tmp", string(b)+"\n")
	must("", os.Rename(f+".tmp", f))
}

// The start time ps gives process pid, or "" when there is none.
func began(pid int) string {
	s, err := quiet("ps", "-o", "lstart=", "-p", strconv.Itoa(pid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

func (a *agentRec) running() bool { return a != nil && a.PID > 0 && began(a.PID) == a.Began }

func (h *headless) worktree(b, base, _ string) string {
	d := state + "/worktrees/" + filepath.Base(repo) + "-" + strings.ReplaceAll(b, "/", "-")
	must(out("git", "-C", repo, "worktree", "add", "-q", "-b", b, d, base))
	return d
}

func (h *headless) open(string) {}

func (h *headless) pane(name string) string {
	if loadAgent(name) == nil {
		return ""
	}
	return agentDir(name)
}

func (h *headless) gone(name string) bool { return !loadAgent(name).running() }

// The agent recorded, with tend's flags for its harness; no process yet. Codex's
// approval flag is the interactive one's: exec never asks.
func (h *headless) start(name, kind string, _ bool, args []string) (string, error) {
	if i := slices.Index(args, "-a"); i >= 0 && i+1 < len(args) {
		args = slices.Delete(slices.Clone(args), i, i+2)
	}
	a := &agentRec{Harness: kind, Args: args}
	if kind == "claude" {
		a.Session = uuid()
	}
	must("", os.MkdirAll(agentDir(name), 0o755))
	a.save(name)
	return agentDir(name), nil
}

func uuid() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6], b[8] = b[6]&0x0f|0x40, b[8]&0x3f|0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// What the agent's latest turn answered: Claude's result, Codex's last
// message, or the end of what it printed on error.
func (h *headless) read(name string) string {
	a := loadAgent(name)
	if a == nil || a.Turn == 0 {
		return ""
	}
	p := fmt.Sprintf("%s/%d", agentDir(name), a.Turn)
	if a.Harness == "codex" {
		if s := cat(p + ".last"); s != "" {
			return s
		}
		said := ""
		for _, l := range strings.Split(cat(p+".out"), "\n") {
			if t := str(get(parse(l), "item", "text")); get(parse(l), "item", "type") == "agent_message" && t != "" {
				said = t
			}
		}
		if said != "" {
			return said
		}
	} else if ls := strings.Split(cat(p+".out"), "\n"); len(ls) > 0 {
		if x := parse(ls[len(ls)-1]); get(x, "type") == "result" && str(get(x, "result")) != "" {
			return str(get(x, "result"))
		}
	}
	e := cat(p + ".err")
	return e[max(0, len(e)-2000):]
}

func (h *headless) keys(string, ...string) {}

func (h *headless) status(name string) string {
	switch a := loadAgent(name); {
	case a == nil:
		return ""
	case a.running():
		return "working"
	}
	return "idle"
}

// One turn: the prompt written beside the agent, its process started and
// recorded before tend waits for it. The process ending is the turn ending,
// however it ended; what it left is for the step to judge. Only a process that
// could not start at all was never asked.
func (h *headless) prompt(name, text string) (string, string) {
	a := loadAgent(name)
	if a == nil {
		return "", "agent_not_found"
	}
	a.Turn++
	p := fmt.Sprintf("%s/%d", agentDir(name), a.Turn)
	write(p+".prompt", text)
	resume := a.Turn > 1
	for try := 0; ; try++ {
		c, err := a.command(name, p, resume)
		if err == nil {
			err = c.Start()
			for _, f := range []any{c.Stdin, c.Stdout, c.Stderr} {
				f.(*os.File).Close() // the turn's own now
			}
		}
		if err != nil {
			return "", "did not start: " + err.Error()
		}
		a.PID, a.Began = c.Process.Pid, began(c.Process.Pid)
		a.save(name)
		done := make(chan error, 1)
		go func() { done <- c.Wait() }()
		ms, _ := strconv.Atoi(timeout())
		select {
		case <-time.After(time.Duration(ms) * time.Millisecond):
			return "", "timeout"
		case err = <-done:
		}
		// A Claude session is made by its first turn: a first turn killed
		// after making it, or a later one before, is sent again the other way.
		if e := cat(p + ".err"); err != nil && try == 0 && a.Harness == "claude" &&
			(strings.Contains(e, "is already in use") || strings.Contains(e, "No conversation found")) {
			resume = !resume
			continue
		}
		break
	}
	if a.Harness == "codex" && a.Session == "" {
		a.Session = str(get(parse(strings.SplitN(cat(p+".out"), "\n", 2)[0]), "thread_id"))
	}
	a.PID, a.Began = 0, ""
	a.save(name)
	return "idle", ""
}

// The turn's process: its harness with tend's flags, in this folder, in a
// session of its own, without the publishing account's token, which is
// tend's alone.
func (a *agentRec) command(name, p string, resume bool) (*exec.Cmd, error) {
	var args []string
	switch {
	case a.Harness == "claude" && resume:
		args = append([]string{"-p", "--output-format", "stream-json", "--verbose", "--resume", a.Session}, a.Args...)
	case a.Harness == "claude":
		args = append([]string{"-p", "--output-format", "stream-json", "--verbose", "--session-id", a.Session}, a.Args...)
	case a.Session != "": // codex exec resume takes the sandbox as config
		args = []string{"exec", "resume", a.Session, "--json", "-o", p + ".last"}
		for i := 0; i+1 < len(a.Args); i++ {
			switch a.Args[i] {
			case "-s":
				args = append(args, "-c", "sandbox_mode="+strconv.Quote(a.Args[i+1]))
			case "--add-dir":
				args = append(args, "-c", "sandbox_workspace_write.writable_roots=["+strconv.Quote(a.Args[i+1])+"]")
			case "-c":
				args = append(args, "-c", a.Args[i+1])
			default:
				continue
			}
			i++
		}
		args = append(args, "-")
	default:
		args = append(append([]string{"exec", "--json", "-o", p + ".last"}, a.Args...), "-")
	}
	c := exec.Command(a.Harness, args...)
	in, err := os.Open(p + ".prompt")
	if err != nil {
		return nil, err
	}
	o, err := os.Create(p + ".out")
	if err != nil {
		return nil, err
	}
	e, err := os.Create(p + ".err")
	if err != nil {
		return nil, err
	}
	c.Stdin, c.Stdout, c.Stderr = in, o, e
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GH_TOKEN=") && !strings.HasPrefix(kv, "TEND_TRIES=") {
			c.Env = append(c.Env, kv)
		}
	}
	c.Env = append(c.Env, "TEND_AGENT="+name)
	return c, nil
}

// Waits for the agent's turn, which a tend before this one started, to end.
func (h *headless) wait(name, ms string, _ ...string) (string, string) {
	n, _ := strconv.Atoi(ms)
	for deadline := time.Now().Add(time.Duration(n) * time.Millisecond); loadAgent(name).running(); time.Sleep(time.Second) {
		if time.Now().After(deadline) {
			return "", "timeout"
		}
	}
	if loadAgent(name) == nil {
		return "", "agent_not_found"
	}
	return "idle", ""
}

func (h *headless) closePane(string)              {}
func (h *headless) closePlace(string)             {}
func (h *headless) notify(string, string, string) {}
