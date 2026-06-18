#!/usr/bin/env bash
#
# Lint budget gate.
#
# Counts two things and compares them against thresholds:
#   * "new"   — issues introduced by this PR (since the merge base);
#   * "total" — all issues located in files the PR modified (new + pre-existing).
#
# Writes a Markdown report to $REPORT_FILE and exposes status=success|failure
# via $GITHUB_OUTPUT. The script itself always exits 0 — enforcement is left to
# the workflow so the comment is posted first.
#
# Usage:  lint-report.sh <base-ref>
# Env:    ALLOWED_NEW (default 5), ALLOWED_TOTAL (default 1), REPORT_FILE.
set -euo pipefail

BASE_REF="${1:-origin/main}"
ALLOWED_NEW="${ALLOWED_NEW:-5}"
ALLOWED_TOTAL="${ALLOWED_TOTAL:-1}"
REPORT_FILE="${REPORT_FILE:-lint-report.md}"

# Go files changed by this PR (drop deleted ones), as a JSON array for jq.
changed_json="$(git diff --name-only --diff-filter=d "${BASE_REF}...HEAD" -- '*.go' \
  | jq -R . | jq -s .)"

# Full report over the repo, plus the subset introduced since the merge base.
# Write JSON to a file (not stdout: golangci-lint also prints a summary line there).
# --issues-exit-code 0 keeps golangci-lint from failing the step on findings.
go tool golangci-lint run --output.json.path full.json --issues-exit-code 0 ./...
go tool golangci-lint run --new-from-merge-base "${BASE_REF}" \
  --output.json.path new.json --issues-exit-code 0 ./...

# Classify: build counts and Markdown lists for new / old issues.
jq -n \
  --slurpfile full full.json \
  --slurpfile new new.json \
  --argjson changed "$changed_json" '
  def key:  "\(.Pos.Filename):\(.Pos.Line):\(.Pos.Column):\(.FromLinter)";
  def line: "- `\(.Pos.Filename):\(.Pos.Line)` **\(.FromLinter)**: \(.Text)";

  ($new[0].Issues // [])                                         as $newIssues
  | ($newIssues | map(key))                                      as $newKeys
  | (($full[0].Issues // [])
      | map(select(.Pos.Filename as $f | $changed | index($f)))) as $modIssues
  | ($modIssues | map(select((key) as $k | $newKeys | index($k) | not)))
                                                                 as $oldIssues
  | {
      new_count:   ($newIssues | length),
      total_count: ($modIssues | length),
      old_count:   ($oldIssues | length),
      new_md:      ($newIssues | map(line) | join("\n")),
      old_md:      ($oldIssues | map(line) | join("\n")),
    }
' > counts.json

new_count=$(jq -r '.new_count'   counts.json)
total_count=$(jq -r '.total_count' counts.json)
old_count=$(jq -r '.old_count'   counts.json)
new_md=$(jq -r '.new_md'         counts.json)
old_md=$(jq -r '.old_md'         counts.json)

status="success"
if (( new_count > ALLOWED_NEW )) || (( total_count > ALLOWED_TOTAL )); then
  status="failure"
fi

if [[ "$status" == "failure" ]]; then
  title="❌ Lint check - failed ❌"
else
  title="✅ Lint check - passed ✅"
fi

{
  echo "## ${title}"
  echo
  echo "_Updated at $(date -u +%Y-%m-%dT%H:%M:%SZ)_"
  echo
  echo "**${new_count}** - new problems detected. ALLOWED NEW PROBLEMS - **${ALLOWED_NEW}**"
  echo
  echo "**${total_count}** - problems detected in modified files. ALLOWED PROBLEMS - **${ALLOWED_TOTAL}**"
  echo
  echo "<details><summary>New Problems (${new_count})</summary>"
  echo
  [[ -n "$new_md" ]] && echo "$new_md" || echo "_none_"
  echo
  echo "</details>"
  echo
  echo "<details><summary>Old Problems (${old_count})</summary>"
  echo
  [[ -n "$old_md" ]] && echo "$old_md" || echo "_none_"
  echo
  echo "</details>"
} > "$REPORT_FILE"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "status=${status}" >> "$GITHUB_OUTPUT"
fi

echo "new=${new_count} total=${total_count} old=${old_count} status=${status}"