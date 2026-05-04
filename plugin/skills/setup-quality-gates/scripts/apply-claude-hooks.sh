#!/usr/bin/env bash
# apply-claude-hooks.sh — buddy hook 설정을 ~/.claude/settings.json 또는
# 프로젝트 .claude/settings.json에 병합한다.
#
# 사용법:
#   ./apply-claude-hooks.sh [--scope user|project] [--dry-run]
#
# 의존: jq (brew install jq / apt install jq)

set -euo pipefail

SCOPE="user"
DRY_RUN=false
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE="$SCRIPT_DIR/../references/claude-hooks-template.json"

while [[ $# -gt 0 ]]; do
  case $1 in
    --scope)   SCOPE="$2"; shift 2 ;;
    --dry-run) DRY_RUN=true; shift ;;
    *) echo "unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ "$SCOPE" == "user" ]]; then
  TARGET="$HOME/.claude/settings.json"
else
  TARGET=".claude/settings.json"
  mkdir -p .claude
fi

if ! command -v jq &>/dev/null; then
  echo "[apply-claude-hooks] jq 가 없어. 먼저 설치해줘: brew install jq" >&2
  exit 1
fi

if [[ ! -f "$TEMPLATE" ]]; then
  echo "[apply-claude-hooks] 템플릿 파일을 못 찾겠어: $TEMPLATE" >&2
  exit 1
fi

# 기존 settings.json이 없으면 빈 객체로 시작
if [[ ! -f "$TARGET" ]]; then
  mkdir -p "$(dirname "$TARGET")"
  echo '{}' > "$TARGET"
fi

# 기존 hooks와 템플릿 hooks를 deep merge (배열은 concatenate)
MERGED=$(jq -s '
  def deep_merge:
    if (.[0] | type) == "object" and (.[1] | type) == "object" then
      reduce (.[1] | keys_unsorted[]) as $k (
        .[0];
        if (.[$k] | type) == "array" and (.[1][$k] | type) == "array" then
          .[$k] += .[1][$k]
        else
          .[$k] = ([ .[$k], .[1][$k] ] | deep_merge)
        end
      )
    else
      .[1]
    end;
  [.[0], .[1]] | deep_merge
' "$TARGET" "$TEMPLATE")

if [[ "$DRY_RUN" == "true" ]]; then
  echo "[apply-claude-hooks] dry-run — 병합 결과 미리보기:"
  echo "$MERGED" | jq .
  exit 0
fi

# 백업
cp "$TARGET" "${TARGET}.buddy.bak" 2>/dev/null || true

echo "$MERGED" > "$TARGET"
echo "[apply-claude-hooks] 완료 — $TARGET 에 buddy hook 병합됨"
echo "  백업: ${TARGET}.buddy.bak"
echo ""
echo "  등록된 hook:"
echo "$MERGED" | jq -r '.hooks // {} | to_entries[] | "  \(.key): \(.value | length)개"'
