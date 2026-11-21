turn() { # name text
	case $1 in
	impl-*)
		case $2 in
		/grove-implement*)
			for d in grove/changes/*/; do
				[ "$d" = grove/changes/archive/ ] || grove archive "$(basename "$d" | cut -d- -f1-2)" >/dev/null 2>&1
			done
			git add -A && git commit -qm build
			case ${STRAY:-} in # work left outside the attempt, as memex's cycle did
			worktree) git worktree add -q "../stray-$1" -b "stray-$1" >/dev/null 2>&1 ;;
			branch) git branch "stray-$1" ;;
			esac
			[ -z "${BLOCK_AFTER_BUILD:-}" ] || block=$1 ;;
		"Implement GitHub issue"* | "Implement Jira issue"*)
			echo "built" >>a.txt && git commit -qam "Build the issue"
			[ -n "${NO_NOTES:-}" ] || echo "Built it; ran the tests." >"$(grep -oE '[^ ,]*/implemented' <<<"$2" | head -1)" ;;
		*"requested changes at commit"*)
			if [ -n "${REQUEST_DECLINE:-}" ]; then
				echo "Out of scope: the issue asks for no wording change." >"$(grep -oE '[^ ]*/reply-[0-9]+' <<<"$2")"
			else
				echo plainer >>a.txt && git commit -qam "As asked"
				echo "Made it plainer, as asked." >"$(grep -oE '[^ ]*/reply-[0-9]+' <<<"$2")"
			fi ;;
		*squash-merged*) # a new message for a later head, naming what was asked
			m=$(grep -oE '[^ ]*/message-[0-9a-f]{40}' <<<"$2" | head -1)
			case $2 in
			*"earlier head"*) printf 'Add the file, plainer\n\nThe project needed a plain file.\n\nGrove: 261015-c0000\n' >"$m" ;;
			*) printf 'Add the file\n\nThe project needed a file.\n\nGrove: 261015-c0000\n' >"$m" ;;
			esac ;;
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
		actionable) jq -n --arg h "$h" '{head: $h, conclusion: "failure", actionable: true, summary: "a.txt:1 lacks the fix."}' ;;
		repairable) # fails the first head it sees, passes the rest
			if [ -e "$s/failed-once" ]; then
				jq -n --arg h "$h" '{head: $h, conclusion: "success", actionable: false, summary: "ok"}'
			else
				touch "$s/failed-once"
				jq -n --arg h "$h" '{head: $h, conclusion: "failure", actionable: true, summary: "a.txt:1 lacks the fix."}'
			fi ;;
		esac >"$v" ;;
	check-*) # a project's own step: the verdict its prompt names, shaped by $CHECK; none writes nothing
		local v h
		v=$(grep -oE '[^ ]*check-[0-9a-f]{40}\.json' <<<"$2" | head -1)
		h=${v##*check-} h=${h%.json}
		case ${CHECK:-success} in
		success) jq -n --arg h "$h" '{head: $h, conclusion: "success", actionable: false, summary: "checked"}' >"$v" ;;
		failure) jq -n --arg h "$h" '{head: $h, conclusion: "failure", actionable: false, summary: "a.txt:1 is not checked."}' >"$v" ;;
		esac ;;
	esac
}
