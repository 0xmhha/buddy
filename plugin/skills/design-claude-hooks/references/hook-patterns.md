# Claude Code Hook 패턴 — 구현 예시

buddy plugin에서 자주 사용하는 4가지 hook 패턴의 완전한 구현 예시.

---

## 패턴 1: freeze-edit-scope (PreToolUse guard)

허가된 경로 외의 파일 편집을 차단. `build-feature` §5에서 actor-track 격리에 사용.

### `~/.claude/hooks/freeze-scope.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // ""')
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // ""')

# 편집 도구가 아니면 허용
if [[ "$TOOL" != "Edit" && "$TOOL" != "Write" && "$TOOL" != "MultiEdit" ]]; then
  echo '{}'
  exit 0
fi

# freeze 상태 파일 확인
FREEZE_FILE="${BUDDY_FREEZE_STATE:-$HOME/.claude/freeze-scope.txt}"
if [[ ! -f "$FREEZE_FILE" ]]; then
  echo '{}'
  exit 0
fi

ALLOWED_PREFIX=$(cat "$FREEZE_FILE")

# 허가된 경로 안에 있으면 허용
if [[ "$FILE" == "$ALLOWED_PREFIX"* ]]; then
  echo '{}'
  exit 0
fi

# 차단
jq -n --arg f "$FILE" --arg p "$ALLOWED_PREFIX" '{
  permissionDecision: "deny",
  message: ("[freeze] \($f) — 현재 scope(\($p))에서 벗어난 편집. freeze-edit-scope로 해제 후 시도해.")
}'
```

### freeze 상태 설정 (buddy hook-wrap)

```bash
# 특정 디렉토리로 scope 고정
echo "/path/to/feature-dir" > ~/.claude/freeze-scope.txt

# 해제
rm ~/.claude/freeze-scope.txt
```

### settings.json 등록

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [{ "type": "command", "command": "~/.claude/hooks/freeze-scope.sh" }]
      }
    ]
  }
}
```

---

## 패턴 2: guard-destructive-commands (PreToolUse ask)

위험한 bash 명령 실행 전 사용자 확인 요청.

### `~/.claude/hooks/guard-destructive.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // ""')

if [[ "$TOOL" != "Bash" ]]; then
  echo '{}'
  exit 0
fi

COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // ""')

# 위험 패턴 목록
PATTERNS=(
  'rm\s+-rf'
  'DROP\s+TABLE'
  'DROP\s+DATABASE'
  'git\s+push\s+.*--force'
  'git\s+reset\s+--hard'
  'kubectl\s+delete'
  'terraform\s+destroy'
  'chmod\s+-R\s+777'
  'truncate\s+'
  '>\s*/dev/sd'
)

for PATTERN in "${PATTERNS[@]}"; do
  if echo "$COMMAND" | grep -qiE "$PATTERN"; then
    jq -n --arg cmd "$COMMAND" --arg pat "$PATTERN" '{
      permissionDecision: "ask",
      message: ("[guard] 위험 명령 감지 (\($pat)): \($cmd | .[0:80])\n계속 실행할까요?")
    }'

    # audit log
    LOG="$HOME/.claude/hooks/audit.jsonl"
    echo "{\"ts\":\"$(date -u +%FT%TZ)\",\"hook\":\"guard-destructive\",\"pattern\":\"$PATTERN\",\"cmd\":\"$(echo "$COMMAND" | head -c 200)\"}" >> "$LOG"

    exit 0
  fi
done

echo '{}'
```

---

## 패턴 3: run-tests (PostToolUse side-effect)

파일 수정 후 관련 테스트를 자동 실행.

### `~/.claude/hooks/run-tests.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // ""')

if [[ "$TOOL" != "Edit" && "$TOOL" != "Write" && "$TOOL" != "MultiEdit" ]]; then
  exit 0
fi

FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // ""')

# 테스트 파일 자체이면 skip
if [[ "$FILE" =~ _test\.(go|ts|py|js)$ ]] || [[ "$FILE" =~ \.test\.(ts|js)$ ]]; then
  exit 0
fi

# 언어별 테스트 실행
if [[ "$FILE" =~ \.go$ ]]; then
  DIR=$(dirname "$FILE")
  go test "./$DIR/..." -count=1 -timeout=30s 2>&1 | tail -5
elif [[ "$FILE" =~ \.(ts|tsx|js|jsx)$ ]]; then
  npm test --silent -- --testPathPattern="$(basename "${FILE%.*}")" 2>&1 | tail -5
elif [[ "$FILE" =~ \.py$ ]]; then
  pytest "$(dirname "$FILE")" -q --tb=short 2>&1 | tail -5
fi

# PostToolUse는 exit code가 0이 아니어도 도구 결과에 영향 없음
exit 0
```

---

## 패턴 4: save-context / restore-context (Stop / SessionStart)

세션 간 컨텍스트 지속.

### Stop hook — `~/.claude/hooks/save-context.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // "unknown"')
TRANSCRIPT=$(echo "$INPUT" | jq -r '.transcript_path // ""')

CONTEXT_DIR="$HOME/.claude/contexts"
mkdir -p "$CONTEXT_DIR"
CONTEXT_FILE="$CONTEXT_DIR/$SESSION_ID.json"

# 트랜스크립트에서 마지막 assistant 메시지 추출
if [[ -f "$TRANSCRIPT" ]]; then
  LAST_MSG=$(tail -20 "$TRANSCRIPT" | jq -s '
    [.[] | select(.type == "assistant")] | last |
    .message.content[0].text // ""
  ' 2>/dev/null || echo '""')

  jq -n \
    --arg sid "$SESSION_ID" \
    --argjson msg "$LAST_MSG" \
    --arg ts "$(date -u +%FT%TZ)" \
    '{session_id: $sid, last_message: $msg, saved_at: $ts}' \
    > "$CONTEXT_FILE"
fi

# Stop hook stdout은 Claude에게 전달됨 — 저장 완료 알림
echo "컨텍스트 저장 완료: $CONTEXT_FILE"
```

### SessionStart hook — `~/.claude/hooks/restore-context.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

CONTEXT_DIR="$HOME/.claude/contexts"

# 가장 최근 컨텍스트 파일 찾기
LATEST=$(ls -t "$CONTEXT_DIR"/*.json 2>/dev/null | head -1)

if [[ -z "$LATEST" ]]; then
  exit 0
fi

# stdout이 Claude 초기 컨텍스트에 주입됨
SAVED_AT=$(jq -r '.saved_at' "$LATEST")
LAST_MSG=$(jq -r '.last_message' "$LATEST")

if [[ -n "$LAST_MSG" && "$LAST_MSG" != "null" ]]; then
  echo "## 이전 세션 컨텍스트 (저장: $SAVED_AT)"
  echo ""
  echo "$LAST_MSG" | head -20
  echo ""
  echo "---"
fi
```

---

## Audit Log 공통 포맷

모든 hook은 JSONL로 감사 로그를 기록.

```bash
LOG="$HOME/.claude/hooks/audit.jsonl"
echo "{
  \"ts\": \"$(date -u +%FT%TZ)\",
  \"hook\": \"$HOOK_NAME\",
  \"tool\": \"$TOOL_NAME\",
  \"decision\": \"$DECISION\",
  \"latency_ms\": $LATENCY,
  \"file\": \"$FILE\"
}" >> "$LOG"
```

로그 확인:
```bash
# 오늘 deny된 항목
jq 'select(.decision == "deny")' ~/.claude/hooks/audit.jsonl | head -20

# hook 별 실행 횟수
jq -r '.hook' ~/.claude/hooks/audit.jsonl | sort | uniq -c | sort -rn
```
