# Claude Code Hook Events — 스키마 레퍼런스

Claude Code가 hook script에 stdin으로 전달하는 JSON envelope 완전 명세.

## PreToolUse

도구가 실행되기 *전*에 호출. 허용/차단/사용자 확인을 결정.

### stdin (입력)

```json
{
  "session_id": "abc123",
  "transcript_path": "/Users/user/.claude/projects/.../abc123.jsonl",
  "tool_name": "Bash",
  "tool_input": {
    "command": "rm -rf /tmp/test",
    "description": "Remove test directory"
  }
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `session_id` | string | 현재 세션 ID |
| `transcript_path` | string | 세션 트랜스크립트 JSONL 절대 경로 |
| `tool_name` | string | 도구 이름 (Bash, Edit, Write, Read, ...) |
| `tool_input` | object | 도구별 입력 파라미터 |

#### 도구별 `tool_input` 구조

**Bash**
```json
{ "command": "string", "description": "string (optional)" }
```

**Edit / MultiEdit**
```json
{ "file_path": "string", "old_string": "string", "new_string": "string" }
```

**Write**
```json
{ "file_path": "string", "content": "string" }
```

**Read**
```json
{ "file_path": "string", "limit": 100, "offset": 0 }
```

**Agent (subagent dispatch)**
```json
{ "description": "string", "prompt": "string", "subagent_type": "string" }
```

### stdout (출력 — decision)

```json
// 허용 (아무것도 반환하지 않거나 빈 객체)
{}

// 사용자에게 확인 요청
{
  "permissionDecision": "ask",
  "message": "[guard] rm -rf 실행 — 정말 할까요?"
}

// 차단
{
  "permissionDecision": "deny",
  "message": "[freeze] /etc 외부 파일 편집은 현재 scope에서 차단됨"
}
```

| 필드 | 값 | 설명 |
|------|-----|------|
| `permissionDecision` | `"ask"` | 사용자 확인 프롬프트 표시 |
| `permissionDecision` | `"deny"` | 도구 실행 차단 |
| (없음) | — | 허용 |
| `message` | string | 사용자에게 표시할 메시지 |

### Exit code 규칙

| exit code | 의미 |
|-----------|------|
| 0 | 정상 (stdout의 JSON을 decision으로 사용) |
| non-zero | hook 실패 → fail-safe 정책에 따라 allow 또는 deny |

---

## PostToolUse

도구 실행이 완료된 *후* 호출. decision 없음. side-effect (format/log/notify)만.

### stdin (입력)

```json
{
  "session_id": "abc123",
  "transcript_path": "...",
  "tool_name": "Edit",
  "tool_input": {
    "file_path": "src/index.ts",
    "old_string": "...",
    "new_string": "..."
  },
  "tool_response": {
    "success": true,
    "content": "File edited successfully"
  }
}
```

| 추가 필드 | 타입 | 설명 |
|----------|------|------|
| `tool_response` | object | 도구 실행 결과 |
| `tool_response.success` | boolean | 성공 여부 |
| `tool_response.content` | string | 결과 메시지 또는 출력 |

### stdout

PostToolUse는 decision을 내리지 않는다. stdout 출력은 Claude에게 전달될 수 있지만 실행을 차단하지는 않는다.

---

## Stop

세션이 종료되기 *직전* 호출. 최종 감사, 컨텍스트 저장 등에 사용.

### stdin (입력)

```json
{
  "session_id": "abc123",
  "transcript_path": "...",
  "stop_hook_active": true
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| `stop_hook_active` | boolean | Stop hook임을 나타내는 플래그 |

### stdout

Stop hook의 stdout 출력은 Claude에게 전달된다 (추가 지시로 해석될 수 있음).
세션 종료를 차단하려면 non-zero exit code + stdout에 이유 출력.

---

## SessionStart

새 세션이 시작될 때 호출. context 주입, 환경 확인 등에 사용.

### stdin (입력)

```json
{
  "session_id": "abc123",
  "transcript_path": "..."
}
```

### stdout

SessionStart hook의 stdout은 Claude의 초기 컨텍스트에 주입된다.
환경 정보, 이전 세션 요약, 온보딩 메시지 출력에 활용.

---

## Matcher 패턴

`settings.json`의 `matcher` 필드는 `tool_name`에 대한 정규식.

```json
"matcher": "Bash"           // Bash 도구만
"matcher": "Edit|Write"     // Edit 또는 Write
"matcher": "Edit|Write|MultiEdit"  // 파일 수정 계열 전체
"matcher": ".*"             // 모든 도구 (성능 주의)
```

매처 없으면 해당 hook 타입의 모든 도구에 적용.
