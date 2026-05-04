---
name: freeze-edit-scope
description: "[패턴 라이브러리] session 동안 Edit/Write를 single directory로 lock (Read/Grep/Glob은 열어 둠). debugging/scoped refactor 중 unrelated code edit 차단. 직접 invoke보다 orchestrator가 import해 사용. 트리거: '디렉토리 잠궈' / 'scope freeze' / '범위 제한' / '외부 Edit 차단'. 참조 위치: 4 핵심 작업 중, compose-safety-mode 컴포넌트, diagnose-bug 격리."
type: skill
---

# Freeze — 디렉토리 범위 편집 잠금


파일 변경을 세션의 나머지 기간 동안 한 디렉토리로 제한. 잠긴 디렉토리 밖 경로를 타깃팅하는 `Edit`이나 `Write`는 **PreToolUse hook에 의해 차단**된다 — 단순 경고가 아니다. Read-only 도구는 계속 동작.

## 이 스킬을 사용하는 경우

- **디버깅** — 10개 파일을 살펴보되 버그 있는 파일 하나만 수정하고 싶을 때. 버그 디렉토리를 freeze하고 나머지는 자유롭게 read.
- **범위 지정 refactor** — 이 작업에서는 `src/auth/`만 바뀌어야 할 때. Freeze하면 실수로 `src/billing/`에 drift하는 게 문자 그대로 불가능.
- **Blast radius 제한 코드 리뷰** — 전체 repo 리뷰하되, 건드리면 지정한 영역에서만.
- **에이전트와의 페어 프로그래밍** — 경계를 미리 설정해 에이전트가 관련 없는 모듈에서 깜짝 편집을 할 수 없게.

## 동작 방식

1. 스킬이 사용자에게 타겟 디렉토리를 묻는다.
2. 후행 슬래시 포함한 절대 경로로 resolve.
3. 그 경로를 상태 파일에 쓴다: `~/.claude/freeze-state.txt`.
4. `Edit|Write|NotebookEdit`의 `PreToolUse` hook이 상태 파일을 읽고 타겟 경로가 frozen prefix로 시작하지 않으면 도구를 차단.

상태 파일이 single source of truth. 삭제해 unfreeze.

## 경로 Prefix 매칭

### 후행 슬래시가 중요

후행 슬래시 없으면 `/Users/me/proj/src`는 다음을 **구별 못 한다**:

- `/Users/me/proj/src/file.ts` ✅ 내부, 허용돼야
- `/Users/me/proj/src-old/file.ts` ❌ 차단돼야, 그런데 prefix `/Users/me/proj/src`가 매칭!

후행 슬래시 있으면 `/Users/me/proj/src/`가 올바로 동작:

- `/Users/me/proj/src/file.ts` ✅ prefix로 시작
- `/Users/me/proj/src-old/file.ts` ✅ 올바르게 prefix로 시작하지 않음

작은 디테일이지만 큰 결과. 상태 파일 쓰기 전에 항상 후행 슬래시로 normalize.

## Read-only 도구는 면제

Freeze는 **변경** 도구에만 적용:

- `Edit`
- `Write`
- `NotebookEdit`

이들은 freeze와 무관하게 완전히 열려 있다:

- `Read`, `Grep`, `Glob` — 어디서든 자유롭게 read.
- `WebFetch`, `WebSearch` — 네트워크 read는 영향 없음.
- `Bash` — read-only 명령은 동작. **Bash는 loophole**: Bash invocation 안의 `sed -i`나 `cat > file`은 freeze를 우회. 파괴적 Bash도 차단하려면 `PreToolUse: Bash`에 별도의 "careful"-style hook을 층으로.

Freeze는 실수 편집 가드이지, 보안 경계가 아니다.

## 워크플로우

### Step 1: 타겟 디렉토리 묻기

free-text 입력(multiple choice 아님)으로 `AskUserQuestion` 사용:

```
Question: "Which directory should be the only writeable area?
          Files outside this path will be blocked from editing."
```

### Step 2: 절대 경로로 resolve

에이전트의 작업 디렉토리가 바뀌는 순간 상대 경로는 깨진다. 항상 절대 경로로 resolve:

```bash
TARGET=$(cd "<user_input>" 2>/dev/null && pwd)
if [ -z "$TARGET" ]; then
  echo "Directory not found: <user_input>"
  exit 1
fi
```

### Step 3: 후행 슬래시 append

```bash
[ "${TARGET%/}" = "$TARGET" ] && TARGET="$TARGET/"
```

이건 `TARGET이 /로 안 끝나면 / append`. 등가:

```bash
TARGET="${TARGET%/}/"   # 후행 / 제거 후 정확히 하나 append
```

### Step 4: 상태 파일 쓰기

```bash
mkdir -p "$HOME/.claude"
echo "$TARGET" > "$HOME/.claude/freeze-state.txt"
```

### Step 5: 사용자에게 confirm

다음처럼 사용자에게 알림:

> 편집이 잠김: `/Users/me/proj/src/`. Read 도구는 모든 곳에서 여전히 동작.
> 해제하려면 unfreeze 단계 실행 (`~/.claude/freeze-state.txt` 삭제).

## Generic Hook 통합

Hook은 stdin에서 도구 입력(Claude Code가 JSON envelope를 넘김)을 읽고, 파일 경로를 추출하고, stdout에 JSON decision을 emit.

### Hook 스크립트: `~/.claude/hooks/freeze-hook.sh`

```bash
#!/usr/bin/env bash
# Save as ~/.claude/hooks/freeze-hook.sh and chmod +x

set -euo pipefail

STATE_FILE="$HOME/.claude/freeze-state.txt"

# Freeze 비활성 → 모두 허용
if [ ! -f "$STATE_FILE" ]; then
  echo '{}'
  exit 0
fi

FROZEN_PATH=$(cat "$STATE_FILE")

# 도구 입력 envelope 읽기
input=$(cat)

# Edit/Write는 file_path; 일부 도구는 path. 둘 다 시도.
target=$(echo "$input" | jq -r '.tool_input.file_path // .tool_input.path // ""')

# 입력에 경로 없음 → 허용 (defensive: 검사 불가 도구를 차단하지 말 것)
if [ -z "$target" ]; then
  echo '{}'
  exit 0
fi

# 절대 경로로 resolve. 아직 존재하지 않는 파일은 realpath 실패
# (예: 새 파일 생성하는 Write) → raw target으로 fallback.
target_abs=$(realpath "$target" 2>/dev/null || echo "$target")

# Prefix 체크
if [[ "$target_abs" == "$FROZEN_PATH"* ]]; then
  echo '{}'
else
  # jq로 deny payload 안전하게 구성 (경로의 quoting 처리)
  jq -n \
    --arg reason "Freeze active — edits locked to $FROZEN_PATH (attempted: $target_abs)" \
    '{permissionDecision: "deny", reason: $reason}'
fi
```

실행 가능하게:

```bash
chmod +x ~/.claude/hooks/freeze-hook.sh
```

### `settings.json`에 wire up

`~/.claude/settings.json` (또는 프로젝트의 `.claude/settings.json`)에 추가:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|NotebookEdit",
        "hooks": [
          {
            "type": "command",
            "command": "$HOME/.claude/hooks/freeze-hook.sh"
          }
        ]
      }
    ]
  }
}
```

Matcher regex `Edit|Write|NotebookEdit`은 세 가지 내장 변경 도구를 커버. Harness가 추가 파일 변경 도구를 노출하면 matcher에 추가.

## Freeze 해제

Unfreeze하려면 상태 파일 삭제:

```bash
rm -f ~/.claude/freeze-state.txt
```

그게 전체 메커니즘. 자매 `unfreeze` 스킬로 wrap 가능:

1. `~/.claude/freeze-state.txt` 존재 확인.
2. 사용자에게 confirm 요청 (의도적이도록).
3. 파일 제거.
4. 확인: "Edits unlocked. All directories writeable again."

Unfreeze가 한 `rm` 명령이므로 전용 스킬은 선택 — 명령을 직접 실행해도 됨.

## 알아둘 가치 있는 실패 모드

- **Bash 우회**: `bash -c "echo foo > /etc/hosts"`는 차단 안 됨. Freeze는 `Edit`/`Write`/`NotebookEdit`만 본다. 필요하면 Bash 안전 hook과 페어링.
- **Symlink**: `realpath`가 resolve하므로 실제 타겟이 frozen 디렉토리 밖에 있는 symlink 편집은 차단됨. 보통 맞는 동작이지만 알아둘 가치.
- **새 파일**: `Write`가 새 파일을 만들 때 아직 없는 경로에 `realpath`가 실패. 스크립트는 raw `target`로 fallback, 에이전트가 절대 경로를 넘겼으면 괜찮음. 상대 경로를 넘겼으면 prefix 매칭이 오동작할 수 있음 — 스킬에서 절대 경로 권장.
- **Stale state**: 파일은 세션 간 유지된다. Unfreeze를 잊으면 다음 세션도 여전히 잠긴다. 문제가 되면 hook에 `~/.claude/freeze-state.txt` age 체크 추가.
- **여러 frozen 디렉토리**: 이 설계는 한 번에 정확히 하나의 frozen prefix 지원. 두 writeable 영역(예: `src/auth/`와 `tests/auth/`) 필요 시 상태 파일을 line당 한 경로로 변경하고, **any** line 매치 시 허용하도록 hook 업데이트.

## Hook 동작 확인

설치 후 빠른 smoke test:

```bash
# /tmp로 freeze 설정
echo "/tmp/" > ~/.claude/freeze-state.txt

# /tmp 밖 파일 편집 시도 — 거부돼야
# Claude Code에서 에이전트에게 프로젝트의 무언가를 Edit하라고 요청
# permission decision: deny가 freeze 메시지와 함께 보여야

# /tmp 내부 파일 편집 시도 — 허용돼야
# 그다음 unfreeze:
rm ~/.claude/freeze-state.txt
```

Deny 메시지가 안 나타나면 확인:

1. `~/.claude/hooks/freeze-hook.sh`이 실행 가능 (`ls -l`이 `x` bit 표시).
2. `settings.json` matcher가 실제로 당신 도구에 fire. 테스트 edit 실행 후 harness의 hook 실행 로그 확인.
3. `jq` 설치됨 (hook이 의존).
