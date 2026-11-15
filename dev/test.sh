#!/usr/bin/env bash
# Regressions for tend, built from this checkout: real Git and Grove, a stand-in
# for Herdr whose agents act at once in tend's working directory. No live agent
# runs. TEND names another build to check instead.
#
#   dev/test.sh
set -uo pipefail
top=$(mktemp -d)
[ -n "${KEEP:-}" ] && echo "kept $top" || trap 'rm -rf "$top"' EXIT
tend=${TEND:-$top/tend}
[ -n "${TEND:-}" ] || (cd "$(dirname "$0")/.." && go build -o "$tend" .) || exit 1
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
		"Implement GitHub issue"*)
			echo "built" >>a.txt && git commit -qam "Build the issue"
			[ -n "${NO_NOTES:-}" ] || echo "Built it; ran the tests." >"$(grep -oE '[^ ,]*/implemented' <<<"$2" | head -1)" ;;
		*"requested changes at commit"*)
			echo plainer >>a.txt && git commit -qam "As asked"
			echo "Made it plainer, as asked." >"$(grep -oE '[^ ]*/reply-[0-9]+' <<<"$2")" ;;
		*squash-merged*) printf 'Add the file\n\nThe project needed a file.\n\nGrove: 261015-c0000\n' >"$(grep -oE '[^ ]*/message-[0-9a-f]{40}' <<<"$2")" ;;
		*) echo fix >>a.txt && git commit -qam repair ;;
		esac ;;
	prop-*)
		case $2 in
		/grove-propose*)
			[ -n "${DECLINE:-}" ] || { mkdir -p grove/changes/261115-p0000-a-proposal && echo "# A proposal" >grove/changes/261115-p0000-a-proposal/proposal.md; } ;;
		*)
			if [ -n "$(git status --porcelain)" ]; then
				git add -A && git commit -qm "$(printf 'Propose a thing\n\nBecause it is wanted.\n\nRefs #%s' "${1#prop-}")"
				echo "1. (consequential) Which thing?" >"$(grep -oE '[^ ,]*/summary' <<<"$2" | head -1)"
			else
				echo "It names no outcome." >"$(grep -oE '[^ ,]*/declined' <<<"$2" | head -1)"
			fi ;;
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
	impl-*) case $4 in /grove-implement* | "Implement GitHub issue"*) fail=${IMPL_FAIL:-} ;; *"requested changes at commit"*) fail=${REQUEST_FAIL:-} ;; *) fail=${REPAIR_FAIL:-} ;; esac ;;
	prop-*) case $4 in /grove-propose*) fail=${PROPOSE_FAIL:-} ;; *) fail=${FOLLOW_FAIL:-} ;; esac ;;
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
# GitHub stand-in. "auth token --user A" gives tok-A, and every write records the
# account whose token made it. One PR per head branch: pr-<branch> holds its body
# and pr-<branch>.by its author. Reviews go to reviews (a JSON array, read back
# through the caller's --jq), statuses to statuses as "<sha> <state> <account>".
# CREATE_LOST opens the PR and then fails, as when GitHub acted but tend never
# heard back. pr-<branch>.merged marks it merged; its head is origin's branch.
set -euo pipefail
s=$HERDR_STATE
printf 'gh %s\n' "$*" >>"$s/log"
by=${GH_TOKEN:-owner} by=${by#tok-}
arg() { local flag=$1; shift; while [ $# -gt 1 ]; do [ "$1" != "$flag" ] || { echo "$2"; return; }; shift; done; }
case "$1 $2" in
"auth token") echo "tok-$(arg --user "$@")" ;;
"pr list")
	# Every PR, then narrowed as the caller asks: by head branch in any state, by
	# author while open, or every open one.
	for w in "$s"/pr-*.where; do
		[ -e "$w" ] || continue
		k=${w%.where}
		[ -e "$k.merged" ] && st=MERGED || st=OPEN
		rd=$(jq -r --arg h "$(git -C "$(sed -n 2p "$w")" rev-parse "$(sed -n 1p "$w")" 2>/dev/null)" 'if any(.[]?; .state == "CHANGES_REQUESTED" and .commit_id == $h) then "CHANGES_REQUESTED" else "" end' "$s/reviews" 2>/dev/null || true)
		jq -n --arg b "$(sed -n 1p "$w")" --arg by "$(cat "$k.by")" --arg st "$st" --arg t "$(cat "$k.title" 2>/dev/null)" --arg u "https://example.test/pr/${k##*/pr-}" --arg rd "$rd" \
			'{headRefName: $b, author: $by, state: $st, title: $t, url: $u, number: 1, reviewDecision: $rd}'
	done | jq -s --arg h "$(arg --head "$@")" --arg a "$(arg --author "$@")" --arg st "$(arg --state "$@")" '
		if $h != "" then map(select(.headRefName == $h)) elif $a != "" then map(select(.author == $a and .state == "OPEN")) elif $st == "all" then . else map(select(.state == "OPEN")) end' |
		if q=$(arg --jq "$@") && [ -n "$q" ]; then jq -r "$q"; else cat; fi ;;
"repo view") case "$*" in *defaultBranchRef*) echo main ;; *) echo o/r ;; esac ;;
"issue list") jq '[.[] | select(.state == "OPEN" and (.labels | index("tend")))]' "$s/issues" | jq -r "$(arg --jq "$@")" ;;
"issue view") jq --argjson n "$3" '.[] | select(.number == $n) | {title, body, url, state, labels: (.labels | map({name: .}))}' "$s/issues" ;;
"issue comment")
	[ -f "$s/comments" ] || echo '[]' >"$s/comments"
	jq --argjson n "$3" --arg by "$by" --rawfile b /dev/stdin '. + [{issue: $n, body: $b, user: {login: $by}}]' "$s/comments" >"$s/c.new" && mv "$s/c.new" "$s/comments" ;;
"issue edit")
	[ -z "${EDIT_FAIL:-}" ] || [ -e "$s/edit-failed" ] || { touch "$s/edit-failed"; exit 1; }
	jq --argjson n "$3" --arg l "$(arg --remove-label "$@")" 'map(if .number == $n then .labels -= [$l] else . end)' "$s/issues" >"$s/i.new" && mv "$s/i.new" "$s/issues" ;;
"pr create")
	b=$(arg --head "$@")
	cat >"$s/pr-${b//\//_}"
	echo "$by" >"$s/pr-${b//\//_}.by"
	printf '%s\n%s\n' "$b" "$(git remote get-url origin)" >"$s/pr-${b//\//_}.where"
	arg --title "$@" >"$s/pr-${b//\//_}.title"
	[ -z "${CREATE_LOST:-}" ] || exit 1
	echo "https://example.test/pr/${b//\//_}" ;;
"pr edit") cat >"$s/pr-${3##*/}" ;;
"pr view")
	k=$s/pr-${3##*/}
	[ -f "$s/reviews" ] || echo '[]' >"$s/reviews"
	{ read -r b; read -r o; } <"$k.where"
	[ -e "$k.merged" ] && st=MERGED || st=OPEN
	jq --arg st "$st" --arg h "$(git -C "$o" rev-parse "$b")" '{state: $st, headRefOid: $h, reviews: map({state, commit: {oid: .commit_id}, author: {login: .user.login}})}' "$s/reviews" |
		jq -r "$(arg --jq "$@")" ;;
api*)
	[ -f "$s/reviews" ] || echo '[]' >"$s/reviews"
	case $2 in
	*/issues/*/comments)
		[ -f "$s/comments" ] || echo '[]' >"$s/comments"
		n=${2%/comments} n=${n##*/}
		jq --argjson n "$n" '[.[] | select(.issue == $n)]' "$s/comments" | jq -r "$(arg --jq "$@")" ;;
	*/reviews/*/comments)
		id=${2%/comments} id=${id##*/}
		jq --argjson id "$id" '[.[] | select(.id == $id)][0].comments // []' "$s/reviews" | jq -r "$(arg --jq "$@")" ;;
	*/reviews)
		if [ "${3:-}" = --input ]; then
			jq --arg by "$by" --slurpfile r /dev/stdin '. + [$r[0] + {id: (length + 1), state: "COMMENTED", user: {login: $by}}]' "$s/reviews" >"$s/reviews.new" && mv "$s/reviews.new" "$s/reviews"
		else
			jq -r "$(arg --jq "$@")" "$s/reviews"
		fi ;;
	*/statuses/*) echo "${2##*/} $(printf '%s\n' "$@" | sed -n 's/^state=//p') $by" >>"$s/statuses" ;;
	esac ;;
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
	unset NO_NOTES REVIEW IMPL_FAIL IMPL_SLEEP REPAIR_FAIL REV_FAIL BLOCK_AFTER_BUILD CREATE_LOST REQUEST_FAIL PROPOSE_FAIL FOLLOW_FAIL DECLINE EDIT_FAIL CAP
}
run() { "$tend" "$t/repo" 261114-k0000 >"$t/out${1:-}" 2>"$t/err${1:-}"; } # [suffix for out and err]
reviews() { grep -c '^agent prompt rev-' "$HERDR_STATE/log"; }
prompts() { grep -c "^agent prompt impl-[^ ]* $1" "$HERDR_STATE/log"; } # text the prompt starts with
agent() { local a; a=$(basename "$HERDR_STATE"/agent-"$2"*); echo "$1" >"$HERDR_STATE/status-${a#agent-}"; } # status name-prefix
passed() { [ $? = 0 ] && [ "$(jq -r .conclusion "$t/out" 2>/dev/null)" = success ]; } # the last run's verdict
status() { "$tend" status "$t/repo" 261114-k0000; }
records() { (cd "$t/repo" && grove changes --json | jq '[.runs[]? | select(.cycle == "261114-k0000")] | length'); }
verdict() { ls "$TEND_STATE"/runs/*/verdict-*.json; }
worktree() { echo "$t"/state/wt-tend-[0-9]*; }
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
allow() { mkdir -p "$TEND_STATE" && echo "$(git -C "$t/repo" remote get-url origin) the-bot" >>"$TEND_STATE/publish"; } # let tend publish to its origin, as the-bot

reviewed() { jq -r --arg h "$1" '[.[] | select(.commit_id == $h)] | length' "$HERDR_STATE/reviews"; } # how many reviews name that head

scenario publish
project "$t/repo" main remote
allow
run
check "with origin listed, the run reaches a verdict" passed
check "the reviewed head is pushed to origin" test "$(git -C "$t/repo.git" rev-parse "$(branch)" 2>&1)" = "$(head)"
check "one pull request is opened, onto the target" test "$(grep -c '^gh pr create --base main ' "$HERDR_STATE/log")" = 1
check "by the listed account, not the owner" test "$(cat "$(pr).by")" = the-bot
check "the implementer was asked for the commit message once" test "$(prompts Branch)" = 1
check "titled with its subject" grep -q '^gh pr create .*--title Add the file ' "$HERDR_STATE/log"
check "described by its body alone" test "$(cat "$(pr)")" = "$(printf 'The project needed a file.\n\nGrove: 261015-c0000')"
check "the implementer may write in the run folder" grep -q -- "--add-dir $TEND_STATE/runs/" <(grep '^agent start impl-' "$HERDR_STATE/log")
check "the head's review shows what merging approves" grep -q 'What merging approves' "$HERDR_STATE/reviews"
check "the head gets one review, a comment by the account" test "$(jq -r --arg h "$(head)" '[.[] | select(.commit_id == $h and .event == "COMMENT" and .user.login == "the-bot")] | length' "$HERDR_STATE/reviews")" = 1
check "saying it passed" grep -q 'review of .*: success' "$HERDR_STATE/reviews"
check "the head carries the tend/review status, set by the account" grep -qx "$(head) success the-bot" "$HERDR_STATE/statuses"
check "status points at the pull request" grep -q 'a person reviews https://example.test/pr/' <<<"$(status)"
echo "A person's description." >"$(pr)"
run 2
check "a repeated run opens no second pull request" test "$(creates)" = 1
check "leaves the description to people" test "$(cat "$(pr)")" = "A person's description."
check "and asks for no second commit message" test "$(prompts Branch)" = 1
check "and posts no second review of the head" test "$(reviewed "$(head)")" = 1

scenario publish-lost
project "$t/repo" main remote
allow
CREATE_LOST=1 run
check "a pull request opened on GitHub but not acknowledged stops" test $? = 3
check "and says a rerun looks first" said "looks for one before opening another"
run
check "the rerun finishes" passed
check "without opening another" test "$(creates)" = 1
check "and posts the review the first run did not" test "$(reviewed "$(head)")" = 1
check "the commit message was asked for once" test "$(prompts Branch)" = 1

scenario publish-repair
project "$t/repo" main remote
allow
REVIEW=repairable run
check "a run with a repair passes" passed
check "the failed head's review is posted on that commit" test "$(jq -r '[.[] | select(.body | test("failure"))] | length' "$HERDR_STATE/reviews")" = 1
check "and says tend repairs once" grep -q 'tend repairs once' "$HERDR_STATE/reviews"
check "as is the repaired head's" test "$(reviewed "$(head)")" = 1
check "each head has its own status" test "$(grep -c ' failure the-bot$' "$HERDR_STATE/statuses") $(grep -c "^$(head) success" "$HERDR_STATE/statuses")" = "1 1"

scenario publish-new-head
project "$t/repo" main remote
allow
run
first=$(head)
(cd "$(worktree)" && echo more >>a.txt && git commit -qam "A later change")
REVIEW=decision run 2
check "a later commit is reviewed and published" test $? = 0
check "its failure goes on the new head" grep -qx "$(head) failure the-bot" "$HERDR_STATE/statuses"
check "the old head's pass is not the new head's" test "$(grep -c "^$(head) success" "$HERDR_STATE/statuses")" = 0
check "its review says it needs a decision" grep -q 'needs a decision' "$HERDR_STATE/reviews"
check "the earlier head keeps its one review" test "$(reviewed "$first")" = 1
check "still one pull request" test "$(creates)" = 1

# A teammate requests changes on the PR's head, with a comment on a line.
request() {
	jq --arg h "$(head)" '. + [{id: (900 + length), state: "CHANGES_REQUESTED", commit_id: $h, user: {login: "teammate"}, body: "Say it plainly.", comments: [{path: "a.txt", line: 1, body: "Plainer."}]}]' "$HERDR_STATE/reviews" >"$t/r" && mv "$t/r" "$HERDR_STATE/reviews"
}
watch() { WATCH_ONCE=1 "$tend" watch >/dev/null 2>"$t/watch"; }

scenario request
project "$t/repo" main remote
allow
run
first=$(head)
request
run 2
check "a run after a teammate requests changes passes" passed
check "the implementer is asked once" test "$(prompts teammate)" = 1
check "with their review" grep -q '^agent prompt impl-.* Their review: Say it plainly\.' "$HERDR_STATE/log"
check "and their comment on a line" grep -q 'Their comments on lines: a.txt:1: Plainer\.' "$HERDR_STATE/log"
check "the new head is reviewed" test "$(reviews)" = 2 -a "$(head)" != "$first"
check "and published: its review carries the implementer's reply" test "$(jq -r --arg h "$(head)" '.[] | select(.commit_id == $h and .user.login == "the-bot") | .body' "$HERDR_STATE/reviews" | grep -c '^> Made it plainer, as asked\.')" = 1
check "its own status" grep -qx "$(head) success the-bot" "$HERDR_STATE/statuses"
check "still one pull request" test "$(creates)" = 1
run 3
check "the answered request is not asked again" test "$(prompts teammate)" = 1

scenario request-uncertain
project "$t/repo" main remote
allow
run
request
REQUEST_FAIL=agent_prompt_stalled run 2
check "a stalled request prompt stops" test $? = 3
run 3
check "the rerun stops, as nothing was committed" test $? = 3
check "and says so" grep -q "committed nothing and is not working on them" "$t/err3"
check "without asking again" test "$(prompts teammate)" = 1
rm "$TEND_STATE"/runs/*/request-*
run 4
check "with the mark deleted by a person, it is asked again and passes" passed
check "twice in all" test "$(prompts teammate)" = 2

scenario request-merged
project "$t/repo" main remote
allow
run
request
touch "$(pr).merged"
run 2
check "a request on a merged pull request is not acted on" test "$(prompts teammate)" = 0

scenario watch
project "$t/repo" main remote
allow
run
watch
check "watch runs nothing when no one requested changes" test "$(grep -c 'running tend' "$t/watch")" = 0
request
watch
check "watch runs tend when a teammate requests changes" grep -q 'changes requested on https://example.test/pr/' "$t/watch"
check "and the request is asked" test "$(prompts teammate)" = 1
request
printf 'x needs-you: something\nnext: a person\n' >"$(echo "$TEND_STATE"/runs/*)/status"
watch
check "watch leaves an attempt that needs a person" test "$(prompts teammate)" = 1

# Intake: issues in the stand-in's GitHub, and tend propose on the repo.
issues() { # number:labels... e.g. 1:tend 2:
	local i n l
	for i in "$@"; do
		n=${i%%:*} l=${i#*:}
		jq -n --argjson n "$n" --arg l "$l" '{number: $n, title: "Request \($n)", body: "Make thing \($n) better.", url: "https://example.test/issues/\($n)", state: "OPEN", labels: ($l | split(",") | map(select(. != "")))}'
	done | jq -s . >"$HERDR_STATE/issues"
}
propose() { "$tend" propose "$t/repo" >"$t/out${1:-}" 2>"$t/err${1:-}"; }
drafts() { grep -c '^agent prompt prop-[^ ]* /grove-propose' "$HERDR_STATE/log"; }
follows() { grep -c '^agent prompt prop-[^ ]* Without asking anyone' "$HERDR_STATE/log"; }
pr_of() { echo "$HERDR_STATE/pr-tend_issue-$1"; }
labelled() { jq -r --argjson n "$1" '.[] | select(.number == $n) | .labels | index("tend") != null' "$HERDR_STATE/issues"; }
comments() { jq --argjson n "$1" '[.[]? | select(.issue == $n)] | length' "$HERDR_STATE/comments" 2>/dev/null || echo 0; }

scenario propose
project "$t/repo" main remote
allow
issues 1:tend 2: 3:tend,bug
propose
check "tend propose exits 0" test $? = 0
check "each labelled issue is drafted once" test "$(drafts)" = 2
check "unattended, from the issue's words" grep -q '^agent prompt prop-1 /grove-propose --unattended Request 1. Make thing 1 better. (GitHub issue https://example.test/issues/1)' "$HERDR_STATE/log"
check "an unlabelled issue is left alone" test ! -e "$(pr_of 2)"
check "a pull request per draft, by the account" test "$(cat "$(pr_of 1).by") $(cat "$(pr_of 3).by")" = "the-bot the-bot"
check "from the issue's branch, onto the target" grep -q '^gh pr create --base main --head tend/issue-1 ' "$HERDR_STATE/log"
check "titled and described by the drafter's commit" grep -q -- '--title Propose a thing --body-file' "$HERDR_STATE/log"
check "described by its body, which refers to the issue" grep -qx 'Refs #1' "$(pr_of 1)"
check "the branch holds the change" test "$(git -C "$t/repo.git" ls-tree -r --name-only tend/issue-1 | grep -c '^grove/changes/261115-p0000')" = 1
check "a review lists the open questions" grep -q 'Which thing?' "$HERDR_STATE/reviews"
check "and says what merging with them open means" grep -q 'its implementation may then choose those answers' "$HERDR_STATE/reviews"
check "naming the change to revise" grep -q '/grove-propose 261115-p0000' "$HERDR_STATE/reviews"
propose 2
check "a second pass drafts nothing again" test "$(drafts)" = 2
check "and opens nothing again" test "$(creates)" = 2
check "nor reviews again" test "$(jq length "$HERDR_STATE/reviews")" = 2

scenario propose-declined
project "$t/repo" main remote
allow
issues 1:tend
DECLINE=1 propose
check "a request that cannot be drafted exits 0" test $? = 0
check "opens no pull request" test ! -e "$(pr_of 1)"
check "says why on the issue, as the account" test "$(jq -r '.[0].user.login' "$HERDR_STATE/comments")" = the-bot
check "in the drafter's words" grep -q 'It names no outcome.' "$HERDR_STATE/comments"
check "and takes the label off" test "$(labelled 1)" = false
DECLINE=1 propose 2
check "so a second pass leaves it" test "$(drafts) $(comments 1)" = "1 1"
issues 1:tend
propose 3
check "labelled again, it is drafted afresh" test "$(drafts)" = 2 -a -e "$(pr_of 1)"

scenario propose-decline-lost
project "$t/repo" main remote
allow
issues 1:tend
DECLINE=1 EDIT_FAIL=1 propose
check "a decline whose label removal fails stops" test $? = 3
DECLINE=1 propose 2
check "the rerun finishes it" test $? = 0 -a "$(labelled 1)" = false
check "without commenting twice or drafting again" test "$(comments 1) $(drafts)" = "1 1"

scenario propose-cap
project "$t/repo" main remote
allow
issues 1:tend 2:tend 3:tend
CAP=2 propose
check "the cap holds open proposals" test "$(drafts)" = 2 -a ! -e "$(pr_of 3)"
check "and says so" grep -q 'the cap; the rest wait' "$t/err"
touch "$(pr_of 1).merged"
CAP=2 propose 2
check "a merged proposal frees a place" test -e "$(pr_of 3)"

scenario propose-follow-lost
project "$t/repo" main remote
allow
issues 1:tend
FOLLOW_FAIL=agent_prompt_stalled propose
check "a stalled follow-up stops" test $? = 3
propose 2
check "the rerun stops too, not asking again" test $? = 3 -a "$(follows)" = 1
check "and says what to do" grep -q 'delete .*/followed' "$t/err2"
rm "$TEND_STATE"/intake/*/followed
propose 3
check "with the mark deleted, it is asked again and published" test $? = 0 -a "$(follows)" = 2 -a -e "$(pr_of 1)"
check "the draft was asked for once" test "$(drafts)" = 1

scenario propose-survivor
project "$t/repo" main remote
allow
issues 1:tend
PROPOSE_FAIL=timeout propose
check "a drafting prompt that timed out stops" test $? = 3
agent working prop
propose 2
check "the rerun waits for the drafter, then has it commit" test $? = 0 -a -e "$(pr_of 1)"
check "without asking for the draft again" test "$(drafts)" = 1

scenario propose-unlisted
project "$t/repo" main remote
issues 1:tend
propose
check "a repository not listed for publication gives no requests" test $? = 3 -a "$(drafts)" = 0
check "and tend says why" grep -q 'tend takes no requests from it' "$t/err"

scenario overview
project "$t/repo" main remote
allow
(cd "$t/repo" && c=grove/changes/261016-c0001-another && mkdir -p "$c" && printf 'title: Another\ncreated: 2026-10-16T09:00:00Z\nschema: default\nbase: %s\n' "$(git rev-parse HEAD)" >"$c/.grove.yaml" && printf '# Another\n\n## Why\n\nBecause.\n' >"$c/proposal.md" && printf '# Tasks\n\n- [ ] T-01 Do it\n' >"$c/tasks.md" && git add -A && git commit -qm another && git push -q origin main)
issues 1:tend 2:tend
CAP=1 propose
run
request
"$tend" status "$t/repo" >"$t/overview" 2>&1
check "the overview names an approved change without a cycle" grep -q '^  261016-c0001 Another' "$t/overview"
check "a cycle not done, with Grove's verdict" grep -q '^  261114-k0000 One: claimed by attempt' "$t/overview"
check "the attempt on this runner and its stage" grep -qE '^  [0-9]{6}-[a-z0-9]{5} \(261114-k0000\): done: success' "$t/overview"
check "the proposal to review" grep -A1 '^Proposals to review' "$t/overview" | grep -q 'Propose a thing'
check "the implementation to review, with the change requested" grep -A1 '^Implementations to review' "$t/overview" | grep -q 'Add the file .*changes requested'
check "and the request still waiting for intake" grep -A1 '^Requests waiting' "$t/overview" | grep -q '#2 Request 2'
check "but not the one already drafted" test "$(grep -c '#1 Request 1' "$t/overview")" = 0
touch "$(pr_of 1).merged"
"$tend" status "$t/repo" >"$t/overview2" 2>&1
check "nor once its proposal is merged" test "$(grep -c '#1 Request 1' "$t/overview2")" = 0
(cd "$t/repo" && mkdir -p grove/changes/archive && git mv grove/changes/261015-c0000-a-change grove/changes/archive/261015-c0000-a-change && git commit -qm landed && git push -q origin main)
"$tend" status "$t/repo" >"$t/overview3" 2>&1
check "an attempt whose cycle is done is left out" test "$(grep -c '(261114-k0000)' "$t/overview3")" = 0

scenario watch-intake
project "$t/repo" main remote
allow
issues 1:tend
WATCH_ONCE=1 "$tend" watch "$t/repo" >/dev/null 2>"$t/watch"
check "watch runs intake for a repository it is given" test -e "$(pr_of 1)"
issues 1:tend 2:tend
FOLLOW_FAIL=agent_prompt_stalled WATCH_ONCE=1 "$tend" watch "$t/repo" >/dev/null 2>&1
WATCH_ONCE=1 "$tend" watch "$t/repo" >/dev/null 2>&1
check "and leaves a request that needs a person" test "$(follows)" = 2

# A repository without Grove, with origin; a GitHub issue is the work item.
plain() {
	git init -q -b main "$t/repo" && (cd "$t/repo" && echo x >a.txt && git add -A && git commit -qm init &&
		git clone -q --bare . "$t/repo.git" && git remote add origin "$t/repo.git" && git fetch -q origin && git branch -q -u origin/main)
}
build() { "$tend" "$t/repo" issue-7 >"$t/out${1:-}" 2>"$t/err${1:-}"; }
bworktree() { git -C "$t/repo" worktree list --porcelain | awk '/^worktree /{w=substr($0,10)} $0 == "branch refs/heads/tend/build-7" {print w}'; }

scenario issue-build
plain
allow
issues 7:tend-build
build
check "an issue labelled tend-build is built, reviewed and published without Grove" test $? = 0 -a "$(jq -r .conclusion "$t/out" 2>/dev/null)" = success
check "the implementer is given the issue as its scope" grep -q '^agent prompt impl-issue-7-[0-9]* Implement GitHub issue #7 .* Request 7. Make thing 7 better.' "$HERDR_STATE/log"
check "the reviewer is given the issue and the implementer's notes" grep -q "^agent prompt rev-.*what it must do is GitHub issue #7, quoted here: Request 7.*notes are in $TEND_STATE/runs/issue-7-" "$HERDR_STATE/log"
check "on branch tend/build-7, pushed" test "$(git -C "$t/repo.git" rev-parse tend/build-7 2>&1)" = "$(git -C "$(bworktree)" rev-parse HEAD)"
check "the commit message is asked to close the issue" grep -q "^agent prompt impl-issue-7-.*End with the line 'Closes #7'" "$HERDR_STATE/log"
check "one pull request by the account" test "$(cat "$HERDR_STATE/pr-tend_build-7.by")" = the-bot
check "its review shows no grove review" test "$(grep -c 'What merging approves' "$HERDR_STATE/reviews")" = 0
check "its attempt is named by cksum of the repository's path, as before" test -d "$TEND_STATE/runs/issue-7-$(printf %s "$t/repo" | cksum | cut -d' ' -f1)"
check "status finds the attempt" grep -q '^attempt issue-7-.* of issue-7: tend/build-7' <<<"$("$tend" status "$t/repo" issue-7)"
check "the overview lists it to review" grep -A1 '^Implementations to review' <<<"$("$tend" status "$t/repo")" | grep -q 'Add the file'
build 2
check "a rerun neither builds nor reviews again" test "$(grep -c '^agent prompt impl-issue-7-[0-9]* Implement' "$HERDR_STATE/log") $(reviews)" = "1 1"

scenario issue-ineligible
plain
allow
issues 7:tend
build
check "an issue without tend-build stops, not approved" test $? = 3 -a "$(grep -c '^agent prompt' "$HERDR_STATE/log")" = 0
check "and says so" grep -q 'not labelled tend-build' "$t/err"
jq 'map(.state = "CLOSED")' "$HERDR_STATE/issues" >"$t/i" && mv "$t/i" "$HERDR_STATE/issues"
build 2
check "a closed issue is done: exit 0, nothing run" test $? = 0 -a "$(grep -c '^agent prompt' "$HERDR_STATE/log")" = 0

scenario issue-in-grove
project "$t/repo" main remote
allow
issues 7:tend-build
build
check "an issue in a repository with Grove is not built directly" test $? = 3 -a "$(grep -c '^agent prompt' "$HERDR_STATE/log")" = 0
check "and tend points to a proposal" grep -q 'label it tend for tend propose' "$t/err"

scenario issue-no-notes
plain
allow
issues 7:tend-build
NO_NOTES=1 build
check "a build without notes for the reviewer stops" test $? = 3
check "and says so" grep -q 'wrote no notes for the reviewer' "$t/err"

scenario issue-survivor
plain
allow
issues 7:tend-build
IMPL_FAIL=timeout build
check "an issue build whose prompt timed out stops" test $? = 3
agent working impl
build 2
check "the rerun waits for the implementer on the same branch and passes" test $? = 0 -a "$(jq -r .conclusion "$t/out2")" = success
check "without prompting it again" test "$(grep -c '^agent prompt impl-issue-7-[0-9]* Implement' "$HERDR_STATE/log")" = 1

scenario issue-repair
plain
allow
issues 7:tend-build
REVIEW=repairable build
check "an issue build is repaired once and passes" test $? = 0 -a "$(jq -r .conclusion "$t/out")" = success
check "the repair is told to update the notes" grep -q "^agent prompt impl-issue-7-.*update $TEND_STATE/runs/issue-7-[0-9]*/implemented where a claim changes" "$HERDR_STATE/log"

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
