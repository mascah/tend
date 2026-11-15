#!/usr/bin/env bash
# Regressions for bin/tend: real Git and Grove, a stand-in for Herdr whose agents
# act at once in tend's working directory. No live agent runs.
#
#   dev/test.sh
set -uo pipefail
tend=$(cd "$(dirname "$0")/.." && pwd)/bin/tend
top=$(mktemp -d)
[ -n "${KEEP:-}" ] && echo "kept $top" || trap 'rm -rf "$top"' EXIT
export GIT_AUTHOR_NAME=Test GIT_AUTHOR_EMAIL=test@example.com GIT_COMMITTER_NAME=Test GIT_COMMITTER_EMAIL=test@example.com
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1

mkdir "$top/bin"
cat >"$top/bin/herdr" <<'EOS'
#!/usr/bin/env bash
# Herdr stand-in. An agent's turn happens at once in tend's working directory:
# the implementer archives every change and commits, or for a repair commits a
# fix; the reviewer writes the verdict its prompt names, shaped by $REVIEW.
# IMPL_FAIL, REPAIR_FAIL and REV_FAIL fail that prompt with the code they hold,
# leaving its turn pending; an agent whose status file says working then takes
# that turn when waited for, as one that outlived tend would. IMPL_SLEEP holds
# the implementer's prompt that long. BLOCK_AFTER_BUILD leaves it blocked once
# its build turn has settled. As Herdr does, a blocked agent refuses a prompt.
set -euo pipefail
s=$HERDR_STATE
printf '%s\n' "$*" >>"$s/log"
said() { jq -n --arg st "$1" '{result: {agent: {agent_status: $st, pane_id: "p9"}}}'; }
turn() { # name text
	case $1 in
	impl-*)
		case $2 in
		/grove-implement*)
			for d in grove/changes/*/; do
				[ "$d" = grove/changes/archive/ ] || grove archive "$(basename "$d" | cut -d- -f1-2)" >/dev/null 2>&1
			done
			git add -A && git commit -qm build
			[ -z "${BLOCK_AFTER_BUILD:-}" ] || block=$1 ;;
		*) echo fix >>a.txt && git commit -qam repair ;;
		esac ;;
	rev-*)
		local v h
		v=$(grep -oE '[^ ]*verdict-[0-9a-f]{40}\.json' <<<"$2" | head -1)
		h=${v##*verdict-} h=${h%.json}
		case ${REVIEW:-success} in
		success) jq -n --arg h "$h" '{head: $h, conclusion: "success", actionable: false, summary: "ok"}' ;;
		noconclusion) jq -n --arg h "$h" '{head: $h, actionable: false, summary: "ok"}' ;;
		decision) jq -n --arg h "$h" '{head: $h, conclusion: "failure", actionable: false, summary: "a.txt:1 needs an owner decision."}' ;;
		repairable) # fails the first head it sees, passes the rest
			if [ -e "$s/failed-once" ]; then
				jq -n --arg h "$h" '{head: $h, conclusion: "success", actionable: false, summary: "ok"}'
			else
				touch "$s/failed-once"
				jq -n --arg h "$h" '{head: $h, conclusion: "failure", actionable: true, summary: "a.txt:1 lacks the fix."}'
			fi ;;
		esac >"$v" ;;
	esac
}
case "$1 $2" in
"worktree create")
	shift 2
	while [ $# -gt 0 ]; do
		case $1 in --cwd) repo=$2 ;; --branch) b=$2 ;; --base) base=$2 ;; esac
		shift
	done
	git -C "$repo" worktree add -q -b "$b" "$s/wt-$b" "$base" >/dev/null 2>&1
	jq -n --arg p "$s/wt-$b" '{result: {worktree: {path: $p}, workspace: {workspace_id: "w1"}}}' ;;
"worktree list") echo '{"result":{"worktrees":[]}}' ;;
"worktree open") echo '{"result":{"workspace":{"workspace_id":"w1"}}}' ;;
"pane list") echo '{"result":{"panes":[{"pane_id":"p0","agent":null}]}}' ;;
"pane split") echo '{"result":{"pane":{"pane_id":"p1"}}}' ;;
"agent start") touch "$s/agent-$3" ;;
"agent get")
	[ -f "$s/agent-$3" ] || { echo '{"error":{"code":"agent_not_found"}}'; exit 1; }
	said "$(cat "$s/status-$3" 2>/dev/null || echo idle)" ;;
"agent wait")
	if [ "$(cat "$s/status-$3" 2>/dev/null)" = working ] && [ -f "$s/pending-$3" ]; then
		turn "$3" "$(cat "$s/pending-$3")"
		rm "$s/pending-$3" "$s/status-$3"
	fi
	said "$(cat "$s/status-$3" 2>/dev/null || echo idle)" ;;
"agent prompt")
	case $3 in
	impl-*) case $4 in /grove-implement*) fail=${IMPL_FAIL:-} ;; *) fail=${REPAIR_FAIL:-} ;; esac ;;
	*) fail=${REV_FAIL:-} ;;
	esac
	if [ "$(cat "$s/status-$3" 2>/dev/null)" = blocked ]; then
		echo '{"error":{"code":"agent_blocked"}}'
		exit 1
	fi
	if [ -n "$fail" ]; then
		printf '%s' "$4" >"$s/pending-$3"
		jq -n --arg c "$fail" '{error: {code: $c}}'
		exit 1
	fi
	case $3 in impl-*) [ -z "${IMPL_SLEEP:-}" ] || { echo $$ >"$s/sleeper"; exec sleep "$IMPL_SLEEP"; } ;; esac
	turn "$3" "$4"
	said "$(cat "$s/status-$3" 2>/dev/null || echo idle)"
	[ -z "${block:-}" ] || echo blocked >"$s/status-$block" ;;
*) echo '{"result":{}}' ;;
esac
EOS
chmod +x "$top/bin/herdr"
cat >"$top/bin/gh" <<'EOS'
#!/usr/bin/env bash
# GitHub stand-in: one PR per head branch, its body in pr-<branch>, statuses
# appended to statuses as "<sha> <state>". CREATE_LOST opens the PR and then
# fails, as when GitHub acted but tend never heard back.
set -euo pipefail
s=$HERDR_STATE
printf 'gh %s\n' "$*" >>"$s/log"
arg() { while [ $# -gt 1 ]; do [ "$1" != "$flag" ] || { echo "$2"; return; }; shift; done; }
case "$1 $2" in
"pr list") flag=--head b=$(arg "$@"); [ ! -f "$s/pr-${b//\//_}" ] || echo "https://example.test/pr/${b//\//_}" ;;
"pr create")
	flag=--head b=$(arg "$@")
	cat >"$s/pr-${b//\//_}"
	[ -z "${CREATE_LOST:-}" ] || exit 1
	echo "https://example.test/pr/${b//\//_}" ;;
"pr edit") cat >"$s/pr-${3##*/}" ;;
api*)
	sha=${2##*/} flag=-f st=$(printf '%s\n' "$@" | sed -n 's/^state=//p')
	echo "$sha $st" >>"$s/statuses" ;;
esac
EOS
chmod +x "$top/bin/gh"
export PATH="$top/bin:$PATH"

fails=0
check() { # description, then a command that must succeed
	local what=$1
	shift
	if "$@"; then echo "ok   $name: $what"; else echo "FAIL $name: $what"; fails=$((fails + 1)); fi
}

# A Grove project with one approved change in cycle 261114-k0000, on branch
# $2 (default main), and with origin when $3 is "remote". Runs in $t.
project() { # dir [target] [remote]
	local d=$1 target=${2:-main}
	git init -q -b "$target" "$d" && cd "$d" || return
	echo x >a.txt && git add -A && git commit -qm init
	grove init >/dev/null
	sed "s/^target: .*/target: $target/" grove.yaml >g && mv g grove.yaml
	git add -A && git commit -qm grove
	local c=grove/changes/261015-c0000-a-change
	mkdir -p "$c" grove/cycles
	printf 'title: A change\ncreated: 2026-10-15T09:00:00Z\nschema: default\nbase: %s\n' "$(git rev-parse HEAD)" >"$c/.grove.yaml"
	printf '# A change\n\n## Why\n\nBecause.\n\n## What changes\n\nA file.\n' >"$c/proposal.md"
	printf '# Tasks\n\n- [ ] T-01 Add the file\n' >"$c/tasks.md"
	git add -A && git commit -qm propose
	printf 'title: "One"\ncreated: 2026-11-14T02:00:00Z\nchanges: [261015-c0000]\nplanned: %s\n' "$(git rev-parse HEAD)" >grove/cycles/261114-k0000-one.yaml
	git add -A && git commit -qm plan
	if [ "${3:-}" = remote ]; then
		git clone -q --bare . "$d.git" && git remote add origin "$d.git" && git fetch -q origin && git branch -q -u "origin/$target"
	fi
	cd - >/dev/null
}

# Each scenario gets its own directory, repo and stand-in state.
scenario() {
	name=$1
	t=$top/$name
	mkdir -p "$t/state"
	export HERDR_STATE=$t/state TEND_STATE=$t/tend
	unset REVIEW IMPL_FAIL IMPL_SLEEP REPAIR_FAIL REV_FAIL BLOCK_AFTER_BUILD CREATE_LOST
}
run() { "$tend" "$t/repo" 261114-k0000 >"$t/out${1:-}" 2>"$t/err${1:-}"; } # [suffix for out and err]
reviews() { grep -c '^agent prompt rev-' "$HERDR_STATE/log"; }
prompts() { grep -c "^agent prompt impl-[^ ]* $1" "$HERDR_STATE/log"; } # text the prompt starts with
agent() { local a; a=$(basename "$HERDR_STATE"/agent-"$2"*); echo "$1" >"$HERDR_STATE/status-${a#agent-}"; } # status name-prefix
passed() { [ $? = 0 ] && [ "$(jq -r .conclusion "$t/out" 2>/dev/null)" = success ]; } # the last run's verdict
status() { "$tend" status "$t/repo" 261114-k0000; }
records() { (cd "$t/repo" && grove changes --json | jq '[.runs[]? | select(.cycle == "261114-k0000")] | length'); }
verdict() { ls "$TEND_STATE"/runs/*/verdict-*.json; }
worktree() { echo "$t"/state/wt-*; }
said() { grep -q "$1" "$t/err"; }

scenario remote-claim
project "$t/repo" main remote
run
check "claims with origin and reaches a verdict" test $? = 0
check "the verdict says success" test "$(jq -r .conclusion "$t/out" 2>/dev/null)" = success
check "with no Git excludes, the candidate stays clean" test -z "$(git -C "$(worktree)" status --porcelain 2>&1)"
check "the verdict is kept outside the candidate" test -f "$(verdict)"
check "the reviewer may write only there besides its checkout" grep -q -- "--add-dir $TEND_STATE/runs/" "$HERDR_STATE/log"

scenario remote-resume
project "$t/repo" main remote
IMPL_FAIL=stub_failure run
check "an implementer failure after the claim stops for a person" test $? = 3
id=$(cd "$t/repo" && grove changes --json | jq -r '.runs[0].id')
unset IMPL_FAIL
run
check "the rerun exits 0" test $? = 0
check "the rerun carries on the same attempt" said "carrying on attempt $id"
check "one record for the cycle" test "$(records)" = 1
check "an origin not listed for publication gets no push or pull request" test "$(grep -c '^gh ' "$HERDR_STATE/log")" = 0
check "and tend says why" said "is not listed in"
check "and its candidate branch is not pushed" test -z "$(git -C "$t/repo.git" branch --list 'grove/2*')"

scenario local-resume
project "$t/repo"
IMPL_FAIL=stub_failure run
check "without origin, a failure after the claim stops" test $? = 3
run
check "the rerun carries on and exits 0" test $? = 0
check "one record for the cycle" test "$(records)" = 1
check "without origin, nothing is published" test "$(grep -c '^gh ' "$HERDR_STATE/log")" = 0

head() { git -C "$(worktree)" rev-parse HEAD; }
branch() { git -C "$(worktree)" branch --show-current; }
pr() { echo "$HERDR_STATE/pr-$(branch | tr / _)"; }
creates() { grep -c '^gh pr create' "$HERDR_STATE/log"; }
allow() { mkdir -p "$TEND_STATE" && git -C "$t/repo" remote get-url origin >>"$TEND_STATE/publish"; } # let tend publish to its origin

scenario publish
project "$t/repo" main remote
allow
run
check "with origin, the run reaches a verdict" passed
check "the reviewed head is pushed to origin" test "$(git -C "$t/repo.git" rev-parse "$(branch)" 2>&1)" = "$(head)"
check "one pull request is opened, onto the target" test "$(grep -c '^gh pr create --base main ' "$HERDR_STATE/log")" = 1
check "titled after the cycle" grep -q '^gh pr create .*--title One ' "$HERDR_STATE/log"
check "its body names the reviewed head and the result" grep -q "Review of $(head | cut -c1-7): success" "$(pr)"
check "and what merging approves, from grove review" grep -q 'approves' "$(pr)"
check "the head carries the tend/review status" grep -qx "$(head) success" "$HERDR_STATE/statuses"
check "status points at the pull request" grep -q 'a person reviews https://example.test/pr/' <<<"$(status)"
run 2
check "a repeated run opens no second pull request" test "$(creates)" = 1
check "it updates the one there" grep -q '^gh pr edit https://example.test/pr/' "$HERDR_STATE/log"

scenario publish-lost
project "$t/repo" main remote
allow
CREATE_LOST=1 run
check "a pull request opened on GitHub but not acknowledged stops" test $? = 3
check "and says a rerun looks first" said "looks for one before opening another"
run
check "the rerun finishes" passed
check "without opening another" test "$(creates)" = 1
check "and sets the status the first run did not" grep -qx "$(head) success" "$HERDR_STATE/statuses"

scenario publish-new-head
project "$t/repo" main remote
allow
run
first=$(head)
(cd "$(worktree)" && echo more >>a.txt && git commit -qam "A later change")
REVIEW=decision run 2
check "a later commit is reviewed and published" test $? = 0
check "its failure goes on the new head" grep -qx "$(head) failure" "$HERDR_STATE/statuses"
check "the old head's pass is not the new head's" test "$(grep -c "^$(head) success" "$HERDR_STATE/statuses")" = 0
check "the body leads with the new head's failure" grep -q "Review of $(head | cut -c1-7): failure" "$(pr)"
check "says it needs a decision" grep -q "needs your decision" "$(pr)"
check "and keeps the earlier pass as history" grep -q "^- $(echo "$first" | cut -c1-7): success" "$(pr)"
check "still one pull request" test "$(creates)" = 1

scenario publish-diverged
project "$t/repo" main remote
allow
run
git clone -q "$t/repo.git" "$t/other" && (cd "$t/other" && git checkout -q "$(branch)" && echo theirs >>a.txt && git commit -qam theirs && git push -q origin HEAD)
(cd "$(worktree)" && echo ours >>a.txt && git commit -qam ours)
run 2
check "a branch someone else pushed to is not overwritten" test $? = 3
check "and tend says why" grep -q "could not push" "$t/err2"
check "their commit is still on origin" test "$(git -C "$t/repo.git" log -1 --format=%s "$(branch)")" = theirs

scenario cached-wrong-head
project "$t/repo"
run
v=$(verdict)
jq '.head = "0000000000000000000000000000000000000000"' "$v" >"$t/v" && mv "$t/v" "$v"
run
check "rerun exits 0" test $? = 0
check "a cached verdict naming another head is reviewed again" test "$(reviews)" = 2
check "the new verdict names the head" test "$(jq -r .head "$v")" = "$(git -C "$(worktree)" rev-parse HEAD)"

scenario cached-incomplete
project "$t/repo"
run
v=$(verdict)
jq 'del(.conclusion)' "$v" >"$t/v" && mv "$t/v" "$v"
run
check "a cached verdict without a conclusion is reviewed again" test "$(reviews)" = 2
check "the set-aside verdict is kept" test -n "$(ls "$v".rejected-* 2>/dev/null)"

scenario new-incomplete
project "$t/repo"
REVIEW=noconclusion run
check "a new verdict without a conclusion stops for a person" test $? = 3
check "and says so" said "without a valid verdict"

scenario target-branch
project "$t/repo" trunk
run
check "a project targeting trunk reaches a verdict" test $? = 0
check "the worktree is based on trunk" grep -q -- '--base trunk' "$HERDR_STATE/log"
check "the reviewer is shown the diff from trunk" grep -q 'git diff trunk...HEAD' "$HERDR_STATE/log"

scenario relative-state
project "$t/repo"
(cd "$t" && TEND_STATE=rel "$tend" "$t/repo" 261114-k0000 >"$t/out" 2>"$t/err")
check "a relative TEND_STATE reaches a verdict" test $? = 0
check "its files stay outside the candidate" test -z "$(git -C "$(worktree)" status --porcelain 2>&1)"

scenario root-moved
project "$t/repo"
run
(cd "$t/repo" && git mv grove agreed && sed 's/^root: .*/root: agreed/' grove.yaml >g && mv g grove.yaml && git commit -qam "Move the documents")
run
check "after the target moves its documents, the attempt carries on" test $? = 0
check "its finished candidate is not implemented again" test "$(grep -c '^agent prompt impl-' "$HERDR_STATE/log")" = 1

scenario legacy
project "$t/repo"
run
# Put the attempt back as the prototype left it: its files in .tend/ in the checkout.
run_dir=$(dirname "$(verdict)")
case $run_dir in "$TEND_STATE"/runs/*)
	touch "$run_dir/repaired" && mkdir "$(worktree)/.tend" && mv "$run_dir"/* "$(worktree)/.tend/" ;;
esac
run
check "an attempt with a legacy .tend/ carries on" test $? = 0
check "its verdict is reused, not discarded" test "$(reviews)" = 1
check "its consumed repair stays consumed, as asked" test -f "$run_dir/repair-asked"
check "and .tend/ leaves the candidate" test ! -e "$(worktree)/.tend"

scenario lock
project "$t/repo"
IMPL_SLEEP=30 "$tend" "$t/repo" 261114-k0000 >"$t/out1" 2>"$t/err1" & # not run: $! must be tend
first=$!
for _ in $(seq 100); do grep -q '^agent prompt impl-' "$HERDR_STATE/log" 2>/dev/null && break; sleep 0.1; done
run 2
check "a second tend while one runs is refused" test $? = 4
check "and names the holder" grep -q "pid $first" "$t/err2"
check "the implementer was prompted once" test "$(grep -c '^agent prompt impl-' "$HERDR_STATE/log")" = 1
kill -9 "$first"
wait "$first" 2>/dev/null
run 3
check "after the holder is killed, tend takes over" test $? = 0
check "carrying on its attempt" grep -q "carrying on attempt" "$t/err3"
kill "$(cat "$HERDR_STATE/sleeper")" 2>/dev/null

scenario repair
project "$t/repo"
REVIEW=repairable run
check "a failed actionable review is repaired once and passes" passed
check "the repair was asked once" test "$(prompts An)" = 1
check "status says done" grep -q "done: success" <<<"$(status)"

scenario repair-undelivered
project "$t/repo"
REVIEW=repairable REPAIR_FAIL=agent_prompt_stalled run
check "a stalled repair prompt stops for a person" test $? = 3
REVIEW=repairable run
check "the rerun neither skips the repair nor exits 0 with the failed verdict" test $? = 3
check "nor asks the repair again" test "$(prompts An)" = 1
check "it says the outcome is unknown and what to do" said "has no commit and impl-.* is not working on it"
check "status shows that" grep -q "needs-you: the repair asked" <<<"$(status)"
rm "$TEND_STATE"/runs/*/repair-asked
REVIEW=repairable run
check "with repair-asked deleted by a person, the repair is asked again and passes" passed
check "twice in all" test "$(prompts An)" = 2

scenario repair-blocked
project "$t/repo"
REVIEW=repairable BLOCK_AFTER_BUILD=1 run
check "an implementer blocked before the repair stops for a person" test $? = 3
check "and the repair is not used" test ! -e "$(echo "$TEND_STATE"/runs/*)/repair-asked"
agent idle impl
REVIEW=repairable run
check "once unblocked, the repair is asked and passes" passed

scenario survivor-implementing
project "$t/repo"
IMPL_FAIL=timeout run
check "an implementer prompt that timed out stops" test $? = 3
check "status shows the stop" grep -q "needs-you: impl-.*: timeout" <<<"$(status)"
agent working impl
run
check "the rerun waits for the implementer still working" test $? = 0
check "and does not prompt it again" test "$(prompts /grove-implement)" = 1
check "and says so" said "still working on an earlier prompt"

scenario survivor-reviewing
project "$t/repo"
REV_FAIL=timeout run
check "a reviewer prompt that timed out stops" test $? = 3
agent working rev
run
check "the rerun waits for the reviewer still working and passes" passed
check "it is prompted once" test "$(reviews)" = 1
check "and started once" test "$(grep -c '^agent start rev-' "$HERDR_STATE/log")" = 1

scenario survivor-repairing
project "$t/repo"
REVIEW=repairable REPAIR_FAIL=timeout run
check "a repair prompt that timed out stops" test $? = 3
agent working impl
REVIEW=repairable run
check "the rerun waits for the repair still working, reviews it and passes" passed
check "the repair was asked once" test "$(prompts An)" = 1

scenario status-interrupted
project "$t/repo"
IMPL_SLEEP=30 "$tend" "$t/repo" 261114-k0000 >"$t/out1" 2>"$t/err1" &
first=$!
for _ in $(seq 100); do grep -q '^agent prompt impl-' "$HERDR_STATE/log" 2>/dev/null && break; sleep 0.1; done
check "status names the running tend" grep -q "tend is on it: pid $first" <<<"$(status)"
check "and what it is waiting for" grep -q "working: impl-.* implements" <<<"$(status)"
kill -9 "$first"
wait "$first" 2>/dev/null
kill "$(cat "$HERDR_STATE/sleeper")" 2>/dev/null
sleep 0.5
check "after a kill, status says to run tend again" grep -q "tend stopped before this finished" <<<"$(status)"

[ "$fails" = 0 ] && echo "all passed" || { echo "$fails failed"; exit 1; }
