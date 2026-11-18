#!/usr/bin/env bash
# Sync roadmap/STATUS.json -> GitHub labels + issues.
# Requires: gh auth with repo write scope.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STATUS="$ROOT/roadmap/STATUS.json"
REPO="${GITHUB_REPOSITORY:-maillunex-lang/BashCapitalOnline}"

if ! command -v gh >/dev/null; then
  echo "gh CLI required" >&2
  exit 1
fi
if ! command -v jq >/dev/null; then
  echo "jq required" >&2
  exit 1
fi

ensure_label() {
  local name="$1" color="$2" desc="$3"
  gh label create "$name" --repo "$REPO" --color "$color" --description "$desc" 2>/dev/null \
    || gh label edit "$name" --repo "$REPO" --color "$color" --description "$desc" 2>/dev/null \
    || true
}

ensure_label "roadmap:done" "0E8A16" "Roadmap item completed"
ensure_label "roadmap:in_progress" "FBCA04" "Roadmap item in progress"
ensure_label "roadmap:next" "1D76DB" "Roadmap next priority"
ensure_label "roadmap:backlog" "BFDADC" "Roadmap backlog"
ensure_label "roadmap:blocked" "D93F0B" "Roadmap blocked"

echo "Syncing items from $STATUS to $REPO ..."

jq -c '.items[]' "$STATUS" | while read -r item; do
  id=$(echo "$item" | jq -r .id)
  title=$(echo "$item" | jq -r .title)
  status=$(echo "$item" | jq -r .status)
  area=$(echo "$item" | jq -r .area)
  notes=$(echo "$item" | jq -r '.notes // empty')
  acceptance=$(echo "$item" | jq -r '.acceptance // [] | map("- [ ] " + .) | join("\n")')
  marker="<!-- roadmap-id: $id -->"
  label="roadmap:$status"
  body=$(cat <<EOF
$marker

**ID:** \`$id\`  
**Area:** \`$area\`  
**Status:** \`$status\`

$notes

## Acceptance
$acceptance

Source: \`roadmap/STATUS.json\` — keep in sync when closing/reopening.
EOF
)
  existing=$(gh issue list --repo "$REPO" --state all --search "$id in:body" --json number,title,body,state \
    --jq ".[] | select(.body | contains(\"roadmap-id: $id\")) | .number" | head -1)
  full_title="[$id] $title"
  if [[ -n "${existing:-}" ]]; then
    gh issue edit "$existing" --repo "$REPO" --title "$full_title" --body "$body" >/dev/null
    # refresh labels: remove other roadmap:* then add current
    for s in done in_progress next backlog blocked; do
      gh issue edit "$existing" --repo "$REPO" --remove-label "roadmap:$s" 2>/dev/null || true
    done
    gh issue edit "$existing" --repo "$REPO" --add-label "$label" 2>/dev/null || true
    if [[ "$status" == "done" ]]; then
      gh issue close "$existing" --repo "$REPO" 2>/dev/null || true
    else
      gh issue reopen "$existing" --repo "$REPO" 2>/dev/null || true
    fi
    echo "updated #$existing $id -> $status"
  else
    args=(issue create --repo "$REPO" --title "$full_title" --body "$body" --label "$label")
    if [[ "$status" == "done" ]]; then
      num=$(gh "${args[@]}" --jq .number 2>/dev/null || gh "${args[@]}")
      # create returns URL normally
      num=$(echo "$num" | grep -oE '[0-9]+$' || true)
      if [[ -n "${num:-}" ]]; then
        gh issue close "$num" --repo "$REPO" 2>/dev/null || true
      fi
    else
      gh "${args[@]}"
    fi
    echo "created $id -> $status"
  fi
done

echo "Done. View: https://github.com/$REPO/issues?q=label%3Aroadmap%3Anext"
