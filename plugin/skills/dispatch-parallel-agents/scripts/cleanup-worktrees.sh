#!/usr/bin/env bash
# cleanup-worktrees.sh — actor-track worktree를 병합 후 정리.
#
# 사용법:
#   ./cleanup-worktrees.sh <target-branch> <actor1> [actor2] ...
#
# 옵션:
#   --no-merge    worktree만 제거 (브랜치 병합 없이)
#   --force       미커밋 변경사항 있어도 강제 제거

set -euo pipefail

MERGE=true
FORCE=false

# 옵션 파싱
while [[ $# -gt 0 ]]; do
  case $1 in
    --no-merge) MERGE=false; shift ;;
    --force)    FORCE=true;  shift ;;
    *)          break ;;
  esac
done

if [[ $# -lt 2 ]]; then
  echo "사용법: $0 [--no-merge] [--force] <target-branch> <actor1> [actor2] ..." >&2
  exit 1
fi

TARGET_BRANCH="$1"
shift
ACTORS=("$@")

REPO_ROOT="$(git rev-parse --show-toplevel)"
REPO_NAME="$(basename "$REPO_ROOT")"
PARENT_DIR="$(dirname "$REPO_ROOT")"

for ACTOR in "${ACTORS[@]}"; do
  BRANCH="actor/$ACTOR"
  WORKTREE_PATH="$PARENT_DIR/${REPO_NAME}-actor-${ACTOR}"

  if [[ ! -d "$WORKTREE_PATH" ]]; then
    echo "[skip] $WORKTREE_PATH 없음"
    continue
  fi

  # 병합
  if [[ "$MERGE" == "true" ]]; then
    echo "[merge] $BRANCH → $TARGET_BRANCH"
    git checkout "$TARGET_BRANCH"
    if git merge --no-ff "$BRANCH" -m "merge(actor/$ACTOR): integrate actor track"; then
      echo "[merged] $BRANCH"
    else
      echo "[conflict] $BRANCH 병합 충돌 — 수동 해결 후 재시도" >&2
      git checkout "$TARGET_BRANCH"
      continue
    fi
  fi

  # worktree 제거
  if [[ "$FORCE" == "true" ]]; then
    git worktree remove --force "$WORKTREE_PATH"
  else
    git worktree remove "$WORKTREE_PATH"
  fi
  echo "[removed] $WORKTREE_PATH"

  # 브랜치 삭제 (병합 완료된 경우만)
  if [[ "$MERGE" == "true" ]]; then
    git branch -d "$BRANCH" 2>/dev/null && echo "[deleted branch] $BRANCH" || true
  fi
done

git worktree prune
echo ""
echo "완료. 남은 worktree:"
git worktree list
