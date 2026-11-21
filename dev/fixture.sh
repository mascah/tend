#!/bin/sh
# Builds a disposable repo to try tend on, in $1 (must not exist):
# Grove's ledger fixture, grove init, trial case (c)'s change, a cycle of one.
set -eu
# Needs a checkout of Grove for its ledger fixture: GROVE_CHECKOUT, or ../grove.
here=$(cd "$(dirname "$0")" && pwd)
top=$(cd "${GROVE_CHECKOUT:-$here/../../grove}" && pwd)
dir=$1
mkdir "$dir" && cp -R "$top/testdata/ledger/." "$dir" && cd "$dir"
git init -q -b main
g() { git -c user.name="Fixture Owner" -c user.email=fixture@example.com "$@"; }
g add -A && g commit -q -m "Ledger 0.3.0"
grove init > /dev/null
# No remote here, so the agent commits and stops; nobody opens a pull request.
sed -i '' '/^rules: {}$/d' grove.yaml
cat >> grove.yaml <<'EOY'
rules:
  skills:
    implement:
      - Commit after each task, naming its task ID.
      - When you are done, commit everything left uncommitted.
EOY
g add -A && g commit -q -m "Set up Grove"
base=$(git rev-parse HEAD)
c=grove/changes/261015-c0000-unknown-id-prints-an-error
mkdir -p $c/specs/record-editing
sed -n "/^  put \$c\/.grove.yaml/,/^  commit /p" "$top/dev/trials/c.sh" | awk -v c="$c" '
  /^  put / { f=$2; sub(/\$c/, c, f); next }
  /^EOF$/ { f=""; next }
  /^  commit / { next }
  f != "" { print > f }'
find $c -type f -exec sed -i '' "s/@BASE@/$base/g" {} +
g add -A && g commit -q -m "Propose an error for an unknown ID"
mkdir -p grove/cycles
cat > grove/cycles/261114-pr0t0-unknown-id.md <<EOC
---
title: "Editing an unknown ID prints an error"
created: 2026-11-14T02:00:00Z
changes: [261015-c0000]
planned: $(git rev-parse HEAD)
---
EOC
g add -A && g commit -q -m "Plan a cycle of one"
# Shared by every worktree: keep Claude Code's local settings out of git status.
printf '.claude/settings.local.json\n' >> .git/info/exclude
grove check
