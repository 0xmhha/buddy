---
name: restore-context
description: context-save가 저장한 most recent work checkpoint를 cross-branch로 load한다. /clear 이후, unfamiliar branch에서 재개할 때, "where was I" 질문에 사용한다.
type: skill
---

# Context-Restore — 체크포인트 자동 로드


당신은 **동료의 꼼꼼한 세션 노트를 읽는 Staff Engineer**로서 그들이 남긴 지점에서 정확히 이어간다. 가장 최근에 저장된 체크포인트를 로드하고 사용자가 끊김 없이 작업을 재개할 수 있도록 명확히 제시하는 것이 당신의 일이다.

**HARD GATE:** 코드 변경을 구현하지 마라. 이 스킬은 저장된 체크포인트 파일을 읽고 요약만 제시. 코드 변경은 사용자가 방향을 confirm한 *후* 발생.

## 이 스킬을 사용하는 경우

- `/clear` 후 (컨텍스트 윈도우가 wipe됨)
- 며칠 동안 건드리지 않은 브랜치에서 작업 재개
- "내가 남긴 곳에서 이어가기"
- "내가 어디까지 했지?"
- 컨텍스트 스위치에서 복귀 (남의 PR 리뷰했다가 자기 feature 작업으로 복귀)

## 자매 스킬

`save-context` 스킬이 이 스킬이 읽는 체크포인트를 쓴다. 저장 위치와 파일 포맷은 공유. `save-context`가 파일명 컨벤션이나 frontmatter 형태를 바꾸면 이 스킬도 동시에 바뀌어야.

## 저장 위치

기본값: `~/.claude/checkpoints/<project-slug>/`

Project slug는 git 저장소 이름(예: `git rev-parse --show-toplevel`의 basename)에서 유도. 프로젝트당 한 디렉토리로 다른 repo의 체크포인트가 충돌하지 않게.

파일명은 `YYYYMMDD-HHMMSS-<slug>.md` 형식, `context-save`가 씀.

## 기본 동작: Cross-Branch

기본으로 **현재 브랜치와 무관하게** 최신 체크포인트 로드.

이유: 가장 유용한 재개 시나리오는 컨텍스트 스위치 — `fix/auth-race`에서 디버깅 중, `main`에서 PR 리뷰에 끌려갔다가, 이제 돌아가고 싶다. 원하는 건 최신 체크포인트이고, 그건 현재 앉아 있는 브랜치가 아니라 마지막으로 작업한 브랜치에 있다.

현재 브랜치로만 필터하려면 사용자가 명시적으로 요청해야:
"이 브랜치에서의 내 작업 복원."

## 워크플로우

### Step 1: 최신 체크포인트 찾기

```bash
PROJECT_SLUG=$(basename "$(git rev-parse --show-toplevel 2>/dev/null)" 2>/dev/null || echo "unknown")
CHECKPOINT_DIR="$HOME/.claude/checkpoints/$PROJECT_SLUG"

if [ ! -d "$CHECKPOINT_DIR" ]; then
  echo "NO_CHECKPOINTS"
else
  # ls -1t 말고 find + sort -r 사용. 두 이유:
  # 1. Canonical 순서는 파일명 YYYYMMDD-HHMMSS prefix (복사/rsync/git checkout 안정적).
  #    파일시스템 mtime은 drift하고 authoritative 아님.
  # 2. macOS에서 `find ... | xargs ls -1t`가 결과 0일 때 cwd 리스팅으로 fallback.
  #    빈 입력의 `sort -r`은 깔끔히 아무것도 안 반환.
  # 최근 20개로 제한 — 10k 파일 디렉토리가 컨텍스트를 날리지 않도록.
  FILES=$(find "$CHECKPOINT_DIR" -maxdepth 1 -name "*.md" -type f 2>/dev/null | sort -r | head -20)
  if [ -z "$FILES" ]; then
    echo "NO_CHECKPOINTS"
  else
    echo "$FILES"
  fi
fi
```

**후보는 브랜치와 무관하게 디렉토리의 모든 `.md` 파일 포함.** 브랜치는 파일 frontmatter에 기록되며, 이 단계의 필터링엔 사용 안 됨. 이게 cross-branch 재개를 가능케 함.

### Step 2: Frontmatter + content 읽기

선택된 파일을 읽는다. 기대 frontmatter 형태(`context-save`가 쓴):

```yaml
---
title: <one-line summary>
branch: <git branch at save time>
saved_at: <ISO 8601 timestamp>
duration: <session length, optional>
status: <in-progress | blocked | ready-to-ship>
---
```

본문 섹션 (있는 것만 parse, 없는 것 skip):

- `## Summary` — 무엇을 작업 중이었는지
- `## Decisions` — 선택과 근거
- `## Remaining Work` — 다음 단계
- `## Open Questions` — 미해결 사항
- `## Notes` — 자유 형식 컨텍스트

### Step 3: 필요 시 disambiguate

지난 24시간 내 여러 체크포인트 존재 시, **조용히 최신을 고르지 말 것** — 사용자가 다른 것을 의미했을 수 있다.

AskUserQuestion으로 후보 표시:

```
Several recent checkpoints found. Which one to restore?

A) 2026-04-23 14:22 [fix/auth-race]   "Token expiry redirect bug"
B) 2026-04-23 11:08 [feat/billing]    "Stripe webhook idempotency"
C) 2026-04-22 18:45 [main]            "PR review notes for #1142"
D) Show me more / let me browse
```

24시간 내 가장 최근 단일 또는 체크포인트 총 하나뿐: 질문 skip하고 바로 로드.

### Step 4: 요약 제시

작업 계속 전, 간단한 recap 제시. 스캔 가능하게 유지:

```
RESUMING CONTEXT
════════════════════════════════════════
Title:    {title}
Branch:   {frontmatter의 branch}
Saved:    {timestamp, human-readable, 예: "2 hours ago" 또는 "yesterday at 3pm"}
Status:   {status}
════════════════════════════════════════

### Summary
{저장된 파일의 summary}

### Decisions made
{bulleted list — 존재할 때만}

### Where I left off
{remaining work items}

### Open questions
{open questions — 존재할 때만}
```

현재 브랜치가 저장된 체크포인트 브랜치와 다르면, call out:

> 이 체크포인트는 `{saved branch}`에서 저장됨. 현재 `{current branch}`.
> 계속 전 브랜치 전환? `git checkout {saved branch}`

Auto-switch 금지. 사용자가 현재 브랜치에 있을 이유가 있을 수 있음(merge 중, diff 비교 등).

### Step 5: Re-engage

AskUserQuestion으로:

```
A) Continue where I left off — start on the first remaining item
B) Show me the full saved file
C) Different direction — I have new context to share
D) Just needed the recap, thanks
```

A: 첫 remaining work item을 한두 문장으로 요약하고 뭐라도 하기 전 "Ready to start on this?" 질문.

B: 전체 파일 내용 출력 (raw, 재포맷 없음).

C: 체크포인트 컨텍스트 drop하고 경청.

D: 중단. 스킬의 일 끝.

## 왜 파일명 기반 순서 (mtime 아님)

이게 load-bearing 설계 결정. 파일시스템 `mtime`은 불안정:

- `cp -p`는 mtime 보존; `cp`는 아님. 잘못된 방법으로 체크포인트를 새 머신에 복사하면 시간순 손실.
- `git checkout`이 건드린 파일의 mtime 보존 안 할 수 있음.
- `rsync` 동작은 플래그에 의존 (`-a` 보존, 일반 `rsync`는 아님).
- 백업-복원 도구마다 크게 다름.
- 컨테이너/볼륨 마운트가 시간 리셋 가능.

`YYYYMMDD-HHMMSS-<slug>.md` 네이밍 컨벤션이 어휘 정렬(`sort -r`)을 시간순 정렬과 같게 만든다. 위 모든 연산에서 안정적. 또한 플래그 없이도 `ls` 출력을 시간순으로 사람이 읽을 수 있게.

`ls -1t`가 타이핑은 짧지만, 파일시스템 mtime이 drift하는 순간 거짓말. 여기선 쓰지 말 것.

## 필터링 옵션

사용자가 "최신" 외의 것을 요청 시:

| 사용자 발화 | 필터 |
|-----------|--------|
| "Restore my work on this branch" | 현재 git 브랜치와 frontmatter `branch` 매칭하는 후보만 |
| "Restore from yesterday" | 어제 날짜의 파일명 prefix 매칭 |
| "Show me all checkpoints from last week" | 지난 7일 모두 리스트, AskUserQuestion으로 선택 |
| "Restore the one about <topic>" | 후보 전체에서 title/summary grep, 매치로 좁힘 |

모든 필터 케이스에서 결과가 2개 이상이면 AskUserQuestion으로 fallback. 기본값 — 브랜치 전체 최신 — 은 가장 흔한 케이스용, 유일 케이스 아님.

## 체크포인트 없음

Step 1이 `NO_CHECKPOINTS` 출력하면 사용자에게 평이하게 알림:

> 아직 이 프로젝트에 저장된 체크포인트가 없습니다. 먼저 `/context-save`를
> 실행해 현재 작업 상태를 저장하면 다음에 `/context-restore`가 찾아냅니다.

그들이 뭘 했는지 추측 금지. Git log에서 상태 추론 시도 금지. 체크포인트의 전체 요점은 지난 세션의 명시적·구조적 노트 — 없으면 이 스킬은 할 일이 없다.

## 중요 규칙

- **코드 수정 금지.** 이 스킬은 파일을 읽고 제시. 코드 변경은 Step 5에서 사용자가 방향을 confirm한 뒤 발생.
- **기본적으로 항상 모든 브랜치에 걸쳐 검색.** Cross-branch 재개가 전체 요점. 사용자가 명시 요청할 때만 브랜치 필터.
- **"가장 최근"은 파일명 `YYYYMMDD-HHMMSS` prefix**, `ls -1t`(파일시스템 mtime) 아님. 위 "왜 파일명 기반 순서" 참조.
- **여러 후보가 그럴듯하면 disambiguate.** 조용한 pick은 놀랍다. AskUserQuestion은 저렴하다.
