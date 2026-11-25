turn() { # name text
	case $1 in
	impl-* | implement-*)
		case $2 in
		/grove-implement*)
			if [ -n "${UNFINISHED:-}" ]; then # a turn that ends with the cycle not built
				echo part >>a.txt && git commit -qam part
				return
			fi
			for d in grove/changes/*/; do
				[ "$d" = grove/changes/archive/ ] || grove archive "$(basename "$d" | cut -d- -f1-2)" >/dev/null 2>&1
			done
			eval "${BUILD_ALSO:-}" # more the build changes, as a test gives it
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
			e=$(grep -oE "End with the line '[^']*'" <<<"$2" | head -1) e=${e#*\'} e=${e%\'} # the trailer asked for
			case $2 in
			*"earlier head"*) printf 'Add the file, plainer\n\nThe project needed a plain file.\n\n%s\n' "$e" >"$m" ;;
			*) printf 'Add the file\n\nThe project needed a file.\n\n%s\n' "$e" >"$m" ;;
			esac ;;
		*"does not merge into it cleanly"*) # how the branch was left, then the target merged in, keeping ours
			{ git rev-parse HEAD; git rev-parse -q --verify MERGE_HEAD || true; git status --porcelain; } >"$s/at-conflict"
			git merge -q -X ours --no-edit origin/main >/dev/null 2>&1 ;;
		*) echo fix >>a.txt && { [ -z "${FIX_ALSO:-}" ] || { eval "$FIX_ALSO" && git add -A; }; } && git commit -qam repair ;;
		esac ;;
	prop-*)
		case $2 in
		/grove-propose*)
			[ -n "${DECLINE:-}" ] || { mkdir -p grove/changes/261115-p0000-a-proposal && echo "# A proposal" >grove/changes/261115-p0000-a-proposal/proposal.md; } ;;
		*)
			if [ -n "$(git status --porcelain)" ]; then
				git add -A && git commit -qm "$(printf 'Propose a thing\n\nBecause it is wanted.\n\nRefs #%s' "$(n=${1#prop-} && echo "${n%-*}")")"
				echo "1. (consequential) Which thing?" >"$(grep -oE '[^ ,]*/summary' <<<"$2" | head -1)"
			else
				echo "It names no outcome." >"$(grep -oE '[^ ,]*/declined' <<<"$2" | head -1)"
			fi ;;
		esac ;;
	rev-*)
		local v h
		v=$(grep -oE '[^ ]*/[a-z]+-[0-9a-f]{40}\.json' <<<"$2" | head -1)
		h=${v##*-} h=${h%.json}
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
	edit-*) # a project's own step that commits
		echo "edit: $2" >>a.txt && git commit -qam edit ;;
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
