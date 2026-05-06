---
name: design-claude-hooks
description: "Claude Code plugin/.claude scope의 PreToolUse, PostToolUse, Stop, SessionStart hook 표준 설계. matcher 패턴, JSON envelope 처리, decision schema (allow/ask/deny), idempotency, error handling, audit logging, hook composition. compose-safety-mode 확장. 트리거: 'hook 설계' / 'PreToolUse 추가' / 'PostToolUse formatter' / 'Stop audit' / 'SessionStart 설정' / 'hook 표준' / 'plugin hook'. 입력: 자동화 의도 (guard/format/audit/notify), 매칭 도구, decision 정책. 출력: hook scripts + settings.json wiring + audit log + composition map. 흐름: design-mcp-server/compose-safety-mode → design-claude-hooks → build-with-tdd."
type: skill
---

# Design Claude Hooks — Plugin Hook 표준 설계

## 1. 목적

Claude Code의 hook 시스템(`PreToolUse`, `PostToolUse`, `Stop`, `SessionStart` 등)을 buddy plugin scope에 표준화한다.

이 스킬은 다음 4개 영역을 처리:
1. **Hook 종류별 사용 시점** — PreToolUse / PostToolUse / Stop / SessionStart 각각의 적합한 use case
2. **JSON envelope 처리** — Claude Code가 stdin으로 넘기는 도구 입력 / stdout으로 받는 decision
3. **Composition** — 여러 hook 결합 (compose-safety-mode 확장)
4. **Audit / observability** — hook 동작 로그, false positive/negative 추적

## 2. 사용 시점 (When to invoke)

- buddy plugin에 자동화 hook 추가 결정 시
- destructive command guard, edit-scope freeze 같은 PreToolUse hook 표준화
- PostToolUse formatter / typecheck / test 자동화
- Stop hook으로 session-end audit (console.log 검사 등)
- SessionStart hook으로 onboarding / context inject
- 여러 hook 결합 시 race condition / order 문제 해결
- hook 에러 / silent fail 디버깅

## 3. 입력 (Inputs)

### 필수
- 자동화 의도 (guard / format / audit / notify / inject)
- 매칭 도구 (Bash / Edit / Write / NotebookEdit / glob pattern)
- decision 정책 (allow / ask / deny / 무관)
- scope (user-global / project / plugin)

### 선택
- 의존 도구 (jq, python3, gh CLI 등)
- audit log 경로
- composition group (다른 hook과 결합)
- timeout (default: 30sec)

### 입력 부족 시 forcing question
- "PreToolUse / PostToolUse 어느 거야? guard면 Pre, format이면 Post."
- "matcher가 정확히 어떤 도구야? Bash + 특정 명령? Edit + 특정 path?"
- "deny가 default야 allow가 default야? hook fail 시 fail-safe?"
- "settings.json은 user-global이야 project야 plugin이야?"

## 4. 핵심 원칙 (Principles)

1. **Hook은 fast (<100ms 권장)** — 모든 도구 호출이 hook 통과. 느리면 UX 저하.
2. **Hook은 idempotent** — 같은 input 재호출 시 동일 결과. side-effect 최소.
3. **Decision은 명시적** — `{}` (allow) / `{"permissionDecision":"ask"}` / `{"permissionDecision":"deny"}`. 묵시 reject 금지.
4. **Fail-safe default** — hook 자체 실패 시 도구 호출 차단? 허용? — 정책 명시.
5. **Audit log 강제** — 모든 hook 실행 기록 (timestamp, tool, input hash, decision, latency).
6. **Composition은 OR-deny** — 여러 hook 중 1개 deny = 전체 deny. silent override 금지.
7. **Matcher는 specific** — 너무 broad한 matcher (`.*`)는 모든 호출 통과. 의도 명확.
8. **Hook은 stateless** — 상태 필요 시 별도 파일 (예: `~/.claude/freeze-state.txt`). hook 자체는 stateless.

## 5. 4 Hook 종류별 설계

### Hook 1. PreToolUse
**용도**: 도구 호출 *전* 검증 / 차단 / 사용자 confirm 요청.

**대표 use case**:
- destructive command guard (rm -rf, DROP TABLE, force push)
- edit-scope freeze (특정 디렉토리 외 Edit 차단)
- secret commit 차단 (gitleaks)
- API key validation (Stripe key 형식 체크)

**Envelope** (stdin):
```json
{
  "tool_name": "Bash",
  "tool_input": {
    "command": "rm -rf /tmp/test"
  },
  "session_id": "...",
  "transcript_path": "..."
}
```

**Decision** (stdout):
```json
// allow
{}

// ask user
{
  "permissionDecision": "ask",
  "message": "[guard] rm -rf 실행 — production data 위험. 계속할까요?"
}

// deny outright
{
  "permissionDecision": "deny",
  "message": "[freeze] /etc 외부 파일 편집은 freeze 상태에서 차단됨"
}
```

### Hook 2. PostToolUse
**용도**: 도구 호출 *후* 자동 처리 (format / typecheck / test / log).

**대표 use case**:
- Edit/Write 후 Prettier auto-format
- Edit/Write 후 typecheck (.ts → tsc --noEmit)
- Bash 결과 console.log 검출 → warn
- PR 생성 후 GitHub Actions trigger

**Envelope**:
```json
{
  "tool_name": "Edit",
  "tool_input": { "file_path": "..." },
  "tool_response": { "success": true },
  "session_id": "..."
}
```

PostToolUse는 **decision 안 내림** (이미 실행됨). side-effect만 (format / log / notify).

### Hook 3. Stop
**용도**: 세션 종료 *전* 최종 검증.

**대표 use case**:
- 모든 modified file의 console.log 감사
- 미커밋 변경 사항 경고
- session 통계 저장

**Envelope**:
```json
{
  "session_id": "...",
  "transcript_path": "...",
  "modified_files": ["..."]
}
```

### Hook 4. SessionStart
**용도**: 세션 시작 *시* context 주입 / 상태 복원.

**대표 use case**:
- recent learning 자동 inject (persist-learning-jsonl)
- last checkpoint 복원 안내 (restore-context)
- onboarding 메시지 (신규 user)

## 5.5. Skill Hook Allowlist (자동 invoke 제외 명시)

buddy plugin 하의 일부 스킬은 **archive / 패턴 라이브러리** 성격이라 hook이 자동 invoke해서는 안 된다. allowlist로 명시 관리.

### Archive 3종 — 직접 invoke 절대 금지

```yaml
hook_skill_allowlist:
  archived_excluded:
    - route-intent           # ARCHIVE — command-routing 패턴이 흡수
    - route-multi-platform   # ARCHIVE — 동일
    - route-spec-to-code     # ARCHIVE — 동일
```

이 3개 스킬 description은 `[ARCHIVE / 참조 전용] LLM이 직접 invoke 금지` prefix를 가진다. hook이 keyword "라우팅"을 잡아 우발 invoke하는 것을 방지하려면:

1. SessionStart hook에서 archive skill을 LLM context에 노출하지 않음
2. PreToolUse hook이 Skill tool invoke 시 archive skill 이름 매칭 시 차단
3. plugin manifest에 명시 (예시):
   ```json
   {
     "hooks_skill_excluded": [
       "route-intent",
       "route-multi-platform",
       "route-spec-to-code"
     ]
   }
   ```

### 패턴 라이브러리 11개 — 직접 invoke 자제

description prefix `[패턴 라이브러리 / 다른 스킬이 참조]`를 가진 스킬:
- `persist-learning-jsonl`, `iterate-fix-verify`, `classify-qa-tiers`, `run-browser-qa`, `classify-review-risks`, `write-changelog`, `monitor-regressions`, `benchmark-llm-models`, `guard-destructive-commands`, `freeze-edit-scope`, `compose-safety-mode`, `detect-install-type`, `guide-setup-wizard`, `audit-live-devex`

이들은 orchestrator(autoplan, critique-plan 등)가 import해서 사용. 사용자가 직접 invoke해도 OK이지만 hook이 자동으로 호출하는 건 부적합.

권장: hook은 4 핵심(autoplan / validate-idea / review-scope / critique-plan) + 명시 GREEN 등급 스킬만 자동 trigger. 패턴 라이브러리는 orchestrator 본문에서 명시 호출 시만 활성.

## 6. 단계 (Phases)

### Phase 1. 의도 분류
hook이 무엇 해결?
- guard (PreToolUse + deny)
- ask (PreToolUse + ask)
- format (PostToolUse + side-effect)
- audit (Stop + log)
- inject (SessionStart + context)

### Phase 2. Matcher 설계
적합한 matcher pattern:
- 단일 도구: `"Bash"` / `"Edit"`
- 다중: `"Edit|Write|NotebookEdit"`
- regex: `"Bash"` (모든 Bash) vs file path 기반 (script 안에서)

너무 broad matcher (`.*`)는 모든 도구 통과 → 성능 저하. specific.

### Phase 3. Script 작성
표준 스켈레톤:
```bash
#!/usr/bin/env bash
set -euo pipefail

# stdin envelope 읽기
INPUT=$(cat)

# tool input 추출 (jq 권장)
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // ""')
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // ""')

# 검증 로직
if [[ "$COMMAND" =~ destructive_pattern ]]; then
  # ask 또는 deny
  jq -n --arg msg "[scope] reason" '{
    permissionDecision: "ask",
    message: $msg
  }'
  exit 0
fi

# default allow
echo '{}'
exit 0
```

### Phase 4. settings.json Wiring

user-global (`~/.claude/settings.json`):
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$HOME/.claude/hooks/guard.sh"
          }
        ]
      }
    ]
  }
}
```

project (`.claude/settings.json` in repo):
- repo 작업 시 자동 활성

plugin (`buddy/plugin/.claude-plugin/plugin.json`):
- plugin 설치 시 hook 자동 등록
- `${CLAUDE_PLUGIN_ROOT}` 환경변수 사용 (hardcode 금지)

### Phase 5. Composition (여러 hook 결합)
같은 matcher에 hook 여러 개:
- 순서대로 실행
- 1개라도 deny → 전체 deny (OR-deny)
- 1개라도 ask → ask (사용자 confirm)
- 모두 allow → allow

`compose-safety-mode` 스킬 패턴 활용:
```yaml
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - command: "guard-destructive.sh"
        - command: "secret-commit-block.sh"
    - matcher: "Edit"
      hooks:
        - command: "freeze-edit-scope.sh"
        - command: "lint-on-save.sh"
```

### Phase 6. Audit Log
모든 hook 실행 기록:
```bash
LOG="$HOME/.claude/hooks/audit.jsonl"
echo "{
  \"timestamp\": \"$(date -u +%FT%TZ)\",
  \"hook\": \"guard-destructive\",
  \"tool\": \"$TOOL_NAME\",
  \"input_hash\": \"$INPUT_HASH\",
  \"decision\": \"$DECISION\",
  \"latency_ms\": $LATENCY,
  \"reason\": \"$REASON\"
}" >> "$LOG"
```

JSONL append-only. `persist-learning-jsonl` 스킬 패턴.

### Phase 7. 테스트
```bash
# manual 테스트
echo '{"tool_name":"Bash","tool_input":{"command":"rm -rf /"}}' | ./guard.sh
# → expected: deny / ask

# Claude Code 통합 테스트
# Edit 작업 시 hook 실행되는지 확인
```

### Phase 8. 운영
- audit log 모니터링 (false positive ratio)
- latency 측정 (p99)
- hook 비활성화 정책 (긴급 escape hatch)

## 7. 출력 템플릿 (Output Format)

```yaml
hook_design:
  name: guard-destructive
  type: PreToolUse
  matcher: "Bash"
  scope: project  # user-global | project | plugin

intent:
  goal: "Prevent destructive bash commands without confirmation"
  patterns_blocked: [rm-rf, DROP-TABLE, force-push, kubectl-delete]
  decision_default: allow
  decision_on_match: ask

script:
  language: bash
  path: "$HOME/.claude/hooks/guard-destructive.sh"
  dependencies: [jq]
  timeout_sec: 5

settings_wiring:
  scope: project
  file: .claude/settings.json
  composition_with: [freeze-edit-scope, secret-commit-block]

composition:
  group: safety-mode
  order_in_group: 1
  fail_safe: deny  # hook script crash → deny
  combine_strategy: OR-deny

audit_log:
  path: ~/.claude/hooks/audit.jsonl
  format: jsonl
  retention: 90d
  fields: [timestamp, hook, tool, input_hash, decision, latency_ms, reason]

testing:
  manual:
    - input: "rm -rf /"
      expected_decision: ask
    - input: "ls -la"
      expected_decision: allow
  integration:
    - claude_code_session_test: yes
    - latency_p99_target_ms: 100

operations:
  monitoring:
    false_positive_rate_target: "<1%"
    latency_p99_target_ms: 100
  escape_hatch:
    method: "Comment out matcher in settings.json"
    audit: "audit.jsonl logs deactivation"
```

## 8. 자매 스킬 (Sibling Skills)

- 앞 단계: `design-mcp-server` (MCP entry point) 또는 `compose-safety-mode` — `Skill` tool로 invoke
- 페어: `guard-destructive-commands`, `freeze-edit-scope` (PreToolUse 구체 구현)
- 다음 단계: `build-with-tdd` (hook 자체 TDD)
- 후속: `setup-quality-gates` (CI 게이트와 hook 연결)
- 후속: `audit-security` (hook 보안 검증)

## 9. 참조 문서

상세 내용은 references/ 파일에서 필요 시 로드:

| 파일 | 내용 |
|------|------|
| `references/hook-events.md` | PreToolUse/PostToolUse/Stop/SessionStart stdin/stdout 완전 스키마 |
| `references/hook-patterns.md` | freeze-scope / guard-destructive / run-tests / save-restore 구현 예시 |
| `references/settings-schema.md` | ~/.claude/settings.json hook 필드 완전 명세 + composition 규칙 |

hook script 작성 시 `hook-events.md` → 입력 파싱, `hook-patterns.md` → 패턴 참조, `settings-schema.md` → 등록 방법 순서로 참조.

## 10. Anti-patterns

1. **Hook이 너무 느림 (>1초)** — 모든 도구 호출이 통과. UX 저하. <100ms 목표.
2. **Matcher가 너무 broad (`.*`)** — 모든 호출에 hook. 성능 + 의도 모호.
3. **Hook이 stateful** — race condition. state는 외부 파일 / KV.
4. **Silent fail (script crash → allow)** — 보안 위험. fail-safe deny.
5. **Audit log 없음** — false positive / negative 추적 불가.
6. **Composition order 무관심** — 어느 hook 먼저 실행? matters for performance.
7. **Hook을 main logic에 사용** — hook은 cross-cutting (guard / format / audit). business logic은 도구 자체.
8. **Hardcoded 경로** — plugin 설치 머신마다 다름. `$HOME` / `${CLAUDE_PLUGIN_ROOT}` 사용.
