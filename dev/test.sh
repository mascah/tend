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
# Herdr stand-in. The implementer archives every change and commits; the
# reviewer writes the verdict its prompt names, shaped by $REVIEW.
set -euo pipefail
s=$HERDR_STATE
printf '%s\n' "$*" >>"$s/log"
idle='{"result":{"agent":{"agent_status":"idle"}}}'
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
"agent get") [ -f "$s/agent-$3" ] ;;
"agent prompt")
	case $3 in
	impl-*)
		[ -z "${IMPL_FAIL:-}" ] || { echo '{"error":{"code":"stub_failure"}}'; exit 1; }
		for d in grove/changes/*/; do
			[ "$d" = grove/changes/archive/ ] || grove archive "$(basename "$d" | cut -d- -f1-2)" >/dev/null 2>&1
		done
		git add -A && git commit -qm build ;;
	rev-*)
		v=$(grep -o '\.tend/verdict-[0-9a-f]*\.json' <<<"$4" | head -1)
		h=${v#.tend/verdict-} h=${h%.json}
		case ${REVIEW:-success} in
		success) jq -n --arg h "$h" '{head: $h, conclusion: "success", actionable: false, summary: "ok"}' ;;
		noconclusion) jq -n --arg h "$h" '{head: $h, actionable: false, summary: "ok"}' ;;
		esac >"$v" ;;
	esac
	echo "$idle" ;;
*) echo '{"result":{}}' ;;
esac
EOS
chmod +x "$top/bin/herdr"
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
	# ponytail: tend still writes .tend/ into the candidate; slice 1 moves it out.
	printf '.tend/\n' >>.git/info/exclude
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
	export HERDR_STATE=$t/state
	unset REVIEW IMPL_FAIL
}
run() { "$tend" "$t/repo" 261114-k0000 >"$t/out" 2>"$t/err"; }
reviews() { grep -c '^agent prompt rev-' "$HERDR_STATE/log"; }
records() { (cd "$t/repo" && grove changes --json | jq '[.runs[]? | select(.cycle == "261114-k0000")] | length'); }
verdict() { ls "$t"/state/wt-*/.tend/verdict-*.json; }
said() { grep -q "$1" "$t/err"; }

scenario remote-claim
project "$t/repo" main remote
run
check "claims with origin and reaches a verdict" test $? = 0
check "the verdict says success" test "$(jq -r .conclusion "$t/out" 2>/dev/null)" = success

scenario remote-resume
project "$t/repo" main remote
IMPL_FAIL=1 run
check "an implementer failure after the claim stops for a person" test $? = 3
id=$(cd "$t/repo" && grove changes --json | jq -r '.runs[0].id')
unset IMPL_FAIL
run
check "the rerun exits 0" test $? = 0
check "the rerun carries on the same attempt" said "carrying on attempt $id"
check "one record for the cycle" test "$(records)" = 1

scenario local-resume
project "$t/repo"
IMPL_FAIL=1 run
check "without origin, a failure after the claim stops" test $? = 3
run
check "the rerun carries on and exits 0" test $? = 0
check "one record for the cycle" test "$(records)" = 1

scenario cached-wrong-head
project "$t/repo"
run
v=$(verdict)
jq '.head = "0000000000000000000000000000000000000000"' "$v" >"$t/v" && mv "$t/v" "$v"
run
check "rerun exits 0" test $? = 0
check "a cached verdict naming another head is reviewed again" test "$(reviews)" = 2
check "the new verdict names the head" test "$(jq -r .head "$v")" = "$(git -C "$(dirname "$v")/.." rev-parse HEAD)"

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

[ "$fails" = 0 ] && echo "all passed" || { echo "$fails failed"; exit 1; }
