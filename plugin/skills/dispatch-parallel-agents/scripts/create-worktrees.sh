#!/usr/bin/env bash
# create-worktrees.sh — actor-track별 git worktree를 일괄 생성.
#
# 사용법:
#   ./create-worktrees.sh <base-branch> <actor1> [actor2] ...
#
# 예시:
#   ./create-worktrees.sh main frontend backend integration
#
# 출력:
#   ../buddy-actor-frontend  (브랜치: actor/frontend)
#   ../buddy-actor-backend   (브랜치: actor/backend)
#   ../buddy-actor-integration (브랜치: actor/integration)

set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "사용법: $0 <base-branch> <actor1> [actor2] ..." >&2
  exit 1
fi

BASE_BRANCH="$1"
shift
ACTORS=("$@")

REPO_ROOT="$(git rev-parse --show-toplevel)"
REPO_NAME="$(basename "$REPO_ROOT")"
PARENT_DIR="$(dirname "$REPO_ROOT")"

# base branch가 존재하는지 확인
if ! git rev-parse --verify "$BASE_BRANCH" &>/dev/null; then
  echo "base branch '$BASE_BRANCH' 를 찾을 수 없어." >&2
  exit 1
fi

CREATED=()

for ACTOR in "${ACTORS[@]}"; do
  BRANCH="actor/$ACTOR"
  WORKTREE_PATH="$PARENT_DIR/${REPO_NAME}-actor-${ACTOR}"

  # 이미 존재하면 skip
  if [[ -d "$WORKTREE_PATH" ]]; then
    echo "[skip] $WORKTREE_PATH 이미 존재함"
    CREATED+=("$WORKTREE_PATH")
    continue
  fi

  # 브랜치가 이미 있으면 해당 브랜치로, 없으면 base에서 생성
  if git rev-parse --verify "$BRANCH" &>/dev/null; then
    git worktree add "$WORKTREE_PATH" "$BRANCH"
  else
    git worktree add -b "$BRANCH" "$WORKTREE_PATH" "$BASE_BRANCH"
  fi

  echo "[created] $WORKTREE_PATH (branch: $BRANCH)"
  CREATED+=("$WORKTREE_PATH")
done

echo ""
echo "생성된 worktree 목록:"
for WT in "${CREATED[@]}"; do
  echo "  $WT"
done

echo ""
echo "정리할 때: $(dirname "$0")/cleanup-worktrees.sh $BASE_BRANCH ${ACTORS[*]}"
